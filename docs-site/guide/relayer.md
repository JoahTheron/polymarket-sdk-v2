# Relayer Integration

The Relayer API submits pre-signed EIP-712 transactions to Polygon on the user's behalf.

## Direct Relayer Usage

```go
import "github.com/bububa/polymarket-client/relayer"

relayerClient := relayer.New(relayer.Config{
    Credentials: &relayer.Credentials{
        APIKey:  "your-relayer-api-key",
        Address: "0xYourWalletAddress",
    },
})

// Submit a transaction
var resp relayer.SubmitTransactionResponse
err := relayerClient.SubmitTransaction(ctx, relayer.SubmitTransactionRequest{
    Type:     "order",
    Payload:  signedPayload,
    Signature: sig,
}, &resp)
```

## Through CLOB Client

The CLOB client can delegate CTF transactions to a configured Relayer:

```go
clobClient := clob.NewClient("",
    clob.WithRelayerClient(relayerClient),
)

// This internally calls relayerClient.SubmitTransaction
var resp relayer.SubmitTransactionResponse
err := clobClient.SubmitRelayerTransaction(ctx, req, &resp)
```

## Querying Transactions

```go
// By ID
tx := relayer.Transaction{TransactionID: "tx-id"}
err := relayerClient.GetTransaction(ctx, &tx)

// Recent transactions (requires auth)
txs, err := relayerClient.GetRecentTransactions(ctx)

// Check Safe wallet deployment
deployed := relayer.SafeDeployedResponse{Address: "0x..."}
err = relayerClient.IsSafeDeployed(ctx, &deployed)
```
