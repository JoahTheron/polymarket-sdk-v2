# bridge Package

Go client for the Polymarket Bridge API.

**Default Host:** `https://bridge-api.polymarket.com`  
**Auth:** None (all endpoints are public)

## Creating a Client

```go
client := bridge.New(bridge.Config{})
```

## Endpoints

| Method | Description |
|---|---|
| `GetBridges` | List all supported bridges and assets |
| `GetConfiguration` | Get user's bridge configuration |
| `RequestDepositAddresses` | Create deposit addresses for a user |
| `RequestQuote` | Get a bridge quote |
| `RequestWithdrawAddresses` | Create withdrawal addresses |
| `GetDepositStatus` | Check deposit transaction status |

## Usage Example

```go
// Get supported assets
assets, err := client.GetBridges(ctx)

// Request a bridge quote
quote, err := client.RequestQuote(ctx, bridge.QuoteRequest{
    FromAmountBaseUnit: "1000000",
    FromChainID:        1,      // Ethereum
    FromTokenAddress:   "0x...",
    ToChainID:          137,    // Polygon
    ToTokenAddress:     "0x...",
    RecipientAddress:   "0x...",
})
```
