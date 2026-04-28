# polymarket-sdk-v2

> Fork of [`bububa/polymarket-client`](https://github.com/bububa/polymarket-client) at v1.0.4 with a critical production-CLOB-v2 fix. Submitted upstream as a PR; this fork is the version that actually places orders against `clob.polymarket.com` today.

Go SDK for [Polymarket](https://polymarket.com) — the decentralized prediction market platform on Polygon.

## Why this fork exists

Verified live against production CLOB v2 on 2026-04-28: every order
submitted through upstream `bububa/polymarket-client v1.0.4` was
rejected by the API with a generic `400 {"error":"Invalid order payload"}`.

After comparing the wire body byte-for-byte against the official
[`py-clob-client-v2`](https://github.com/Polymarket/py-clob-client-v2),
two compounded issues with the order **salt** were isolated:

| | Upstream (v1.0.4 — rejected) | This fork (accepted) |
|---|---|---|
| **Salt value** | full uint256 random (~78 decimal digits) | bounded to `random ∈ [0, ms_timestamp)` (≤ ~1.78e12, 13 digits) — mirrors the official Python generator |
| **Salt JSON serialization** | `String` type → JSON **string** (e.g. `"123..."`) | `Int64` type → JSON **number** (e.g. `123...`) |

Polymarket parses `salt` as a JSON Number that must fit in safe-integer range (≤ 2^53 − 1). uint256 overflows; quoted strings are rejected outright. Both fixes are required.

End-to-end verification: `$1 GTC BUY → 200 OK, status:"live"`, cancel succeeds.

The full fix is two commits: [`dadd379`](https://github.com/JoahTheron/polymarket-sdk-v2/commit/dadd379) (the bug fix itself) and [`0f66254`](https://github.com/JoahTheron/polymarket-sdk-v2/commit/0f66254) (module rename so polyscalper consumers see "our" SDK in their `go.mod`).

## Features

- **Complete CLOB v2 coverage** — market data, order management, positions, RFQ (request-for-quote), rewards
- **WebSocket support** — live order book and order update streams
- **Three-tier auth** — public (no auth), L1 (EIP-712 signatures), L2 (API key + passphrase + wallet signature)
- **All Polymarket APIs** — CLOB, Relayer, Data, Gamma, Bridge
- **Zero live dependencies** — all tests use `httptest.NewServer`, run entirely offline
- **One external dependency** — `github.com/ethereum/go-ethereum` only
- **Production-verified order placement** — only fork where v1.0.4-era PostOrder actually succeeds against `clob.polymarket.com`

## Installation

```bash
go get github.com/JoahTheron/polymarket-sdk-v2
```

Requires **Go 1.23+** (CI uses `>=1.23.0`; `go.mod` declares 1.22).

## Quick Start

### Read-Only (No Auth Required)

```go
package main

import (
    "context"
    "fmt"

    "github.com/JoahTheron/polymarket-sdk-v2/clob"
)

func main() {
    client := clob.NewClient("") // defaults to CLOB v2 host

    // Fetch market data
    marketInfo := clob.ClobMarketInfo{ConditionID: "0xabc123"}
    if err := client.GetClobMarketInfo(context.Background(), &marketInfo); err != nil {
        panic(err)
    }
    fmt.Printf("Market: %s (negRisk=%v)\n", marketInfo.ConditionID, marketInfo.NegRisk)

    // Get order book
    book, err := client.GetOrderBook(context.Background(), "token-id-here")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Best bid: %v, Best ask: %v\n", book.Bids[0].Price, book.Asks[0].Price)
}
```

### Trading (L2 Authentication Required)

```go
package main

import (
    "context"

    "github.com/ethereum/go-ethereum/crypto"

    "github.com/JoahTheron/polymarket-sdk-v2/clob"
    "github.com/JoahTheron/polymarket-sdk-v2/internal/polyauth"
)

func main() {
    // Load your private key (never hardcode in production)
    privateKey, _ := crypto.HexToECDSA("your-private-key-hex")

    client := clob.NewClient("",
        clob.WithCredentials(clob.Credentials{
            Key:        "your-api-key",
            Secret:     "your-api-secret",
            Passphrase: "your-api-passphrase",
        }),
        clob.WithSigner(polyauth.NewSigner(privateKey)),
        clob.WithChainID(clob.PolygonChainID), // 137
    )

    // Place an order
    order, err := client.PostOrder(context.Background(), clob.PostOrderRequest{
        TokenID: "token-id",
        Price:   clob.Float64{Value: 0.50},
        Size:    clob.Float64{Value: 10.0},
        Side:    clob.SideBuy,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Order placed: %s\n", order.Success)
}
```

### Using Other APIs

```go
// Data API — positions, trades, activity (no auth)
import "github.com/JoahTheron/polymarket-sdk-v2/data"

dataClient := data.New(data.Config{})
positions, _ := dataClient.GetPositions(ctx, data.PositionParams{User: "0x..."})

// Gamma API — events, markets search, tags (no auth)
import "github.com/JoahTheron/polymarket-sdk-v2/gamma"

gammaClient := gamma.New(gamma.Config{})
markets, _ := gammaClient.GetMarkets(ctx, gamma.MarketFilterParams{/* ... */})

// Relayer API — submit signed transactions (L1 auth via API key)
import "github.com/JoahTheron/polymarket-sdk-v2/relayer"

relayerClient := relayer.New(relayer.Config{
    Credentials: &relayer.Credentials{
        APIKey:  "...",
        Address: "0x...",
    },
})
```

## Package Overview

| Package | Purpose | Default Host | Auth Required |
|---|---|---|---|
| [`clob`](#clob-package) | CLOB v2 — orders, markets, positions, RFQ | `https://clob.polymarket.com` | Depends on endpoint |
| [`clob/ws`](#clobws-package) | WebSocket live order book & updates | `wss://ws-orderbook.clob.polymarket.com` | L2 |
| [`relayer`](#relayer-package) | Submit signed on-chain transactions | `https://relayer-v2.polymarket.com` | L1 |
| [`data`](#data-package) | Positions, trades, activity, leaderboard | `https://data-api.polymarket.com` | None |
| [`gamma`](#gamma-package) | Market search, events, tags, profiles | `https://gamma-api.polymarket.com` | None |
| [`bridge`](#bridge-package) | Bridge API | `https://bridge-api.polymarket.com` | None |
| [`shared`](#shared-package) | Shared scalar types (`String`, `Int`, `Float64`, `Time`) | — | — |

## Authentication

Polymarket uses three authentication levels:

| Level | Description | How It Works | Endpoints |
|---|---|---|---|
| **AuthNone (0)** | Public access | No headers | Market data, orderbook, prices |
| **AuthL1 (1)** | Wallet-signed | EIP-712 signature of timestamp + nonce | `CreateAPIKey`, `DeriveAPIKey` |
| **AuthL2 (2)** | Full trading | API key + HMAC-secret + wallet signature | Orders, trades, positions, RFQ |

L2 auth requires BOTH a `polyauth.Signer` (from your private key) AND `Credentials` (API key, secret, passphrase).

### Creating API Keys

```go
client := clob.NewClient("",
    clob.WithSigner(polyauth.NewSigner(privateKey)),
    clob.WithChainID(clob.PolygonChainID),
)

// Create new API key (L1 — wallet-signed)
creds, err := client.CreateAPIKey(ctx, nonce)
// Use returned credentials for L2 requests
```

## CLOB Package

All CLOB v2 endpoints:

### Market Data (No Auth)

| Method | Endpoint | Description |
|---|---|---|
| `GetOk` | `/ok` | Health check |
| `GetVersion` | `/version` | API version |
| `GetServerTime` | `/time` | Server timestamp |
| `GetMarkets` | `/markets` | Paginated markets |
| `GetClobMarketInfo` | `/clob-markets/:id` | Single market details |
| `GetOrderBook` | `/book` | Order book for token |
| `GetMidpoint` | `/midpoint` | Midpoint price |
| `GetPrice` | `/price` | Last price by side |
| `GetSpread` | `/spread` | Bid-ask spread |
| `GetLastTradePrice` | `/last-trade-price` | Most recent trade |
| `GetTickSize` | `/tick-size` | Minimum price increment |

### Orders & Trading (AuthL2)

| Method | Endpoint | Description |
|---|---|---|
| `PostOrder` | `/order` | Submit single order |
| `PostOrders` | `/orders` | Submit batch orders (supports `postOnly`, `deferExec`) |
| `CancelOrder` | `/order` | Cancel by order ID |
| `CancelOrders` | `/orders` | Cancel multiple orders |
| `CancelAll` | `/cancel-all` | Cancel all user orders |
| `CancelMarketOrders` | `/cancel-market-orders` | Cancel by market |
| `GetOrder` | `/data/order/:id` | Get order by ID |
| `GetOpenOrders` | `/data/orders` | List open orders |
| `GetTrades` | `/data/trades` | List user trades |

### RFQ (Request for Quote) (AuthL2)

| Method | Endpoint | Description |
|---|---|---|
| `CreateRFQRequest` | `/rfq/request` | Create RFQ |
| `GetRFQRequests` | `/rfq/data/requests` | List RFQs |
| `CreateRFQQuote` | `/rfq/quote` | Create RFQ quote |
| `AcceptRFQRequest` | `/rfq/request/accept` | Accept RFQ |
| `ApproveRFQQuote` | `/rfq/quote/approve` | Approve quote |

### Rewards (AuthL2 + Public)

| Method | Auth | Description |
|---|---|---|
| `GetEarningsForUserForDay` | L2 | User rewards for a date |
| `GetCurrentRewards` | None | Active reward campaigns |
| `GetRewardsForMarket` | None | Rewards for a market |
| `GetBuilderFeeRate` | None | Builder fee configuration |

### WebSocket (`clob/ws`)

```go
import "github.com/JoahTheron/polymarket-sdk-v2/clob/ws"

wsClient, err := ws.New(ws.Config{
    Host: "", // defaults to production
    // Optional: auth for order notifications
    Signer:      polyauth.NewSigner(privateKey),
    Credentials: &ws.Credentials{/* ... */},
    ChainID:     137,
})
defer wsClient.Close()

// Subscribe to order book
err = wsClient.SubscribeOrderBook("token-id")
// Subscribe to order updates (requires auth)
err = wsClient.SubscribeOrders()

// Read updates
for update := range wsClient.Channel {
    fmt.Printf("Update: %+v\n", update)
}
```

## Development

```bash
# Build
go build -v ./...

# Run tests (all offline — httptest.NewServer)
go test -v ./...

# Tidy dependencies
go mod tidy
```

### Test Files

| File | Coverage |
|---|---|
| `clob/auth_test.go` | Auth header generation |
| `clob/client_test.go` | CLOB v2 endpoints, flexible JSON parsing |
| `clob/ctf_test.go` | CTF relayer transaction submission |
| `relayer/client_test.go` | Relayer documented endpoints |
| `shared/flex_test.go` | flexible JSON scalar serialization |

## License

[MIT](LICENSE)
