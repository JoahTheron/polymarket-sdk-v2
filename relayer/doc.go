// Package relayer provides a client for the Polymarket Relayer API.
//
// # Overview
//
// The Relayer API submits EIP-712 signed transactions to the Polygon blockchain
// on behalf of users. It handles nonce management and transaction ordering.
//
// # Authentication
//
// Uses RELAYER_API_KEY and RELAYER_API_KEY_ADDRESS headers, or POLY_BUILDER_*
// headers for builder-authenticated requests.
//
// # Usage
//
//	client := relayer.New(relayer.Config{
//	    Credentials: &relayer.Credentials{
//	        APIKey:  "your-relayer-api-key",
//	        Address: "0xYourAddress",
//	    },
//	})
//
//	// Submit a signed transaction
//	var resp relayer.SubmitTransactionResponse
//	err := client.SubmitTransaction(ctx, relayer.SubmitTransactionRequest{...}, &resp)
//
//	// Check transaction status
//	tx := relayer.Transaction{TransactionID: "tx-id"}
//	err = client.GetTransaction(ctx, &tx)
//
//	// Get recent transactions
//	txs, err := client.GetRecentTransactions(ctx)
package relayer
