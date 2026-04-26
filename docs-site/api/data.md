# data Package

Go client for the Polymarket Data API.

**Default Host:** `https://data-api.polymarket.com`  
**Auth:** None (all endpoints are public)

## Creating a Client

```go
client := data.New(data.Config{})
// or with custom host:
client := data.New(data.Config{Host: "https://data-api.polymarket.com"})
```

## Endpoints

| Method | Description |
|---|---|
| `GetHealth` | API health check (`GET /`) |
| `GetPositions` | Current user positions |
| `GetMarketPositions` | Positions grouped by outcome token |
| `GetClosedPositions` | Closed positions for a user |
| `GetValue` | Portfolio value (USDC + tokens) |
| `GetTrades` | User/market trade history |
| `GetActivity` | User activity events |
| `GetHolders` | Top token holders per market |
| `GetTraded` | Markets traded count |
| `GetOpenInterest` | Market-level open interest |
| `GetLiveVolume` | Live trading volume |
| `GetLeaderboard` | Trader leaderboard rankings |
| `GetBuilderLeaderboard` | Builder attribution rankings |
| `GetBuilderVolume` | Builder volume time series |
| `DownloadAccountingSnapshot` | Download ZIP accounting snapshot |

## Usage Example

```go
// Get user positions
positions, err := client.GetPositions(ctx, data.PositionParams{
    User: "0xYourWalletAddress",
    Limit: 100,
})

// Get open interest for markets
interest, err := client.GetOpenInterest(ctx, []string{"0xcondition1", "0xcondition2"})

// Get trader leaderboard
entries, err := client.GetLeaderboard(ctx, data.LeaderboardParams{
    Limit:      50,
    TimePeriod: "allTime",
    OrderBy:    "volume",
})
```
