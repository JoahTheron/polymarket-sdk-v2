# clob/ws Package

WebSocket client for live order book and order notification streams with automatic reconnection.

**Default Host:** `wss://ws-subscriptions-clob.polymarket.com`  
**Auth:** None (market channels) / L2 (user channel)

## Creating a Client

```go
import "github.com/bububa/polymarket-client/clob/ws"

// Read-only — market-channel only
client := ws.New(
    ws.WithHost(ws.DefaultHost),
)

// With authentication — enables user-channel subscriptions
client := ws.New(
    ws.WithCredentials(&clob.Credentials{
        Key:        "your-api-key",
        Secret:     "your-api-secret",
        Passphrase: "your-passphrase",
    }),
)
```

## Options

| Option | Description |
|---|---|
| `WithHost(host)` | Set the WebSocket host (default: `wss://ws-subscriptions-clob.polymarket.com`) |
| `WithDialOptions(opts)` | Pass custom `*websocket.DialOptions` (headers, subprotocols, etc.) |
| `WithCredentials(creds)` | Set L2 API credentials for user-channel subscriptions |
| `WithAutoReconnect(bool)` | Enable/disable automatic reconnect on read error (default: `true`) |
| `WithOnConnected(fn)` | Callback fired once on first successful connection |
| `WithOnReconnected(fn)` | Callback fired on each successful reconnect |
| `WithOnDisconnected(fn)` | Callback fired when connection drops with no reconnect pending |

## Connecting

Three channel URLs are available:

```go
ctx := context.Background()

client.ConnectMarket(ctx)   // wss://.../ws/market  – order book, prices, etc.
client.ConnectUser(ctx)     // wss://.../ws/user    – order fills, trade events
client.ConnectSports(ctx)   // wss://.../ws         – public sports feed
```

## Market Subscriptions (no auth)

| Method | Description |
|---|---|
| `SubscribeOrderBook(ctx, assetIDs)` | Order book snapshots and deltas |
| `SubscribeLastTradePrice(ctx, assetIDs)` | Last-trade price updates |
| `SubscribePrices(ctx, assetIDs)` | Price change events |
| `SubscribeTickSizeChange(ctx, assetIDs)` | Tick size change notifications |
| `SubscribeMidpoints(ctx, assetIDs)` | Midpoint updates |
| `SubscribeBestBidAsk(ctx, assetIDs)` | Top-of-book bid/ask updates |
| `SubscribeNewMarkets(ctx, assetIDs)` | New market listing events |
| `SubscribeMarketResolutions(ctx, assetIDs)` | Market resolution events |

Each subscribe method has a matching `Unsubscribe...` variant.

## User Subscriptions (requires credentials)

| Method | Description |
|---|---|
| `SubscribeUserEvents(ctx, markets)` | All user order and trade events |
| `SubscribeOrders(ctx, markets)` | Order status updates (alias for `SubscribeUserEvents`) |
| `SubscribeTrades(ctx, markets)` | Trade execution confirmations |

## Reading Events

```go
for event := range client.Events() {
    switch e := event.(type) {
    case *ws.BookEvent:
        // Order book snapshot: e.Bids, e.Asks
    case *ws.PriceEvent:
        // Price update: e.AssetID, e.Price
    case *ws.OrderEvent:
        // Order fill or status change
    }
}
```

## Connection Lifecycle Callbacks

```go
client := ws.New(
    ws.WithOnConnected(func() {
        log.Println("first connection established")
    }),
    ws.WithOnReconnected(func() {
        log.Println("reconnected after disconnect")
    }),
    ws.WithOnDisconnected(func() {
        log.Println("connection lost, no more reconnects")
    }),
)
```

## Reconnection

The client auto-reconnects with exponential backoff (1 s → 2 s → 4 s → ... → 60 s cap).  
Subscriptions are automatically replayed on each successful reconnect.

## Closing

```go
client.Close() // idempotent — safe to call multiple times
```
