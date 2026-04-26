# relayer Package

Go client for the Polymarket Relayer API.

**Default Host:** `https://relayer-v2.polymarket.com`  
**Auth:** Relayer API key or builder key

## Creating a Client

```go
client := relayer.New(relayer.Config{
    Credentials: &relayer.Credentials{
        APIKey:  "your-api-key",
        Address: "0xYourAddress",
    },
})
```

## Endpoints

| Method | Description |
|---|---|
| `SubmitTransaction` | Submit signed transaction to the relayer |
| `GetTransaction` | Get transaction by ID |
| `GetRecentTransactions` | List recent transactions |
| `GetNonce` | Get next relayer nonce |
| `GetRelayPayload` | Get payload for relaying |
| `IsSafeDeployed` | Check if Safe wallet is deployed |
| `GetAPIKeys` | List relayer API keys |

## Authentication

Two auth modes are supported:

| Mode | Config | Headers |
|---|---|---|
| Relayer API key | `Credentials{APIKey, Address}` | `RELAYER_API_KEY`, `RELAYER_API_KEY_ADDRESS` |
| Builder API key | `BuilderCredentials{APIKey, Builder}` | `POLY_BUILDER_API_KEY`, `POLY_BUILDER`, `POLY_BUILDER_SIGNATURE_TYPES` |

## Pre-Allocation Pattern

All single-entity getters accept an output pointer and return `error`:

```go
var tx relayer.Transaction
tx.TransactionID = "tx-id"
err := client.GetTransaction(ctx, &tx)
```
