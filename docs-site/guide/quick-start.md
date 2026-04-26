# Quick Start

## 1. Query Public Market Data

No authentication needed.

```go
package main

import (
    "context"
    "fmt"
    "github.com/bububa/polymarket-client/clob"
)

func main() {
    client := clob.NewClient("") // uses CLOB v2 host by default

    // Get a specific market
    info := clob.ClobMarketInfo{ConditionID: "0x1a2b3c"}
    if err := client.GetClobMarketInfo(context.Background(), &info); err != nil {
        panic(err)
    }
    fmt.Printf("Market: %s\n", info.ConditionID)
    fmt.Printf("NegRisk: %v\n", info.NegRisk)

    // Get order book
    book, err := client.GetOrderBook(context.Background(), "token-id")
    if err != nil {
        panic(err)
    }
    if len(book.Bids) > 0 {
        fmt.Printf("Best bid: %v\n", book.Bids[0].Price)
    }
    if len(book.Asks) > 0 {
        fmt.Printf("Best ask: %v\n", book.Asks[0].Price)
    }
}
```

## 2. Look Up Positions

```go
import "github.com/bububa/polymarket-client/data"

dataClient := data.New(data.Config{})
positions, err := dataClient.GetPositions(ctx, data.PositionParams{
    User: "0xYourWalletAddress",
})
```

## 3. Search Markets

```go
import "github.com/bububa/polymarket-client/gamma"

gammaClient := gamma.New(gamma.Config{})
var results gamma.SearchResults
err := gammaClient.Search(ctx, "election 2024", &results)

// Get a specific market by ID (pre-allocation)
market := gamma.Market{ID: "123"}
err = gammaClient.GetMarket(ctx, &market)

// Or by slug (leave ID empty)
market = gamma.Market{Slug: "election-2024"}
err = gammaClient.GetMarket(ctx, &market)
```

## 4. Place an Order (Auth Required)

See [L2 Auth](/guide/auth-l2) for the full setup. Once configured:

```go
resp, err := client.PostOrder(ctx, clob.PostOrderRequest{
    TokenID: "0xtoken...",
    Price:   clob.Float64{Value: 0.65},
    Size:    clob.Float64{Value: 50.0},
    Side:    clob.SideBuy,
})
```

## Next Steps

- Understand the [Authentication system](/guide/auth-levels)
- Learn about [custom types](/guide/types) (String, Int, Float64, Time)
- Set up [WebSocket streams](/guide/websocket)
