// Package clob provides a Go client for the Polymarket CLOB v2 API.
//
// # Overview
//
// The CLOB (Central Limit Order Book) v2 API is the primary trading interface
// for Polymarket. It supports market data queries, order management, position
// tracking, and RFQ (Request for Quote) workflows.
//
// # Authentication Levels
//
// Endpoints require one of three authentication levels:
//
//	AuthNone (0) — Public endpoints (market data, orderbook, prices)
//	AuthL1   (1) — EIP-712 wallet signature (CreateAPIKey, DeriveAPIKey)
//	AuthL2   (2) — API key + HMAC secret + wallet signature (orders, trades, positions)
//
// # Creating a Client
//
// Read-only (public data):
//
//	client := clob.NewClient("")
//	client := clob.NewClient(clob.V2Host)
//
// With full trading access:
//
//	client := clob.NewClient("",
//	    clob.WithCredentials(clob.Credentials{
//	        Key:        "your-api-key",
//	        Secret:     "your-api-secret",
//	        Passphrase: "your-passphrase",
//	    }),
//	    clob.WithSigner(polyauth.NewSigner(privateKey)),
//	    clob.WithChainID(clob.PolygonChainID),
//	)
//
// # Market Data (No Auth Required)
//
//	market := clob.ClobMarketInfo{ConditionID: "0xabc123"}
//	client.GetClobMarketInfo(ctx, &market)
//	book := clob.OrderBookSummary{AssetID: "token-id"}
//	client.GetOrderBook(ctx, &book)
//	var mid clob.MidpointResponse
//	client.GetMidpoint(ctx, "token-id", &mid)
//	var price clob.PriceResponse
//	client.GetPrice(ctx, "token-id", clob.Buy, &price)
//	var tick clob.TickSizeResponse
//	client.GetTickSize(ctx, "token-id", &tick)
//
// # Orders & Trading (AuthL2 Required)
//
//	client.PostOrder(ctx, clob.PostOrderRequest{...})
//	client.PostOrders(ctx, []clob.PostOrderRequest{...}, false, false)
//	client.CancelOrder(ctx, "order-id")
//	client.GetOpenOrders(ctx, clob.OpenOrderParams{Market: "0x..."})
//	client.GetTrades(ctx, clob.TradeParams{...})
//
// # RFQ (Request for Quote)
//
//	client.CreateRFQRequest(ctx, clob.CreateRFQRequest{...})
//	client.CreateRFQQuote(ctx, clob.CreateRFQQuoteRequest{...})
//	client.AcceptRFQRequest(ctx, "request-id")
//
// # Rewards & Builder APIs
//
//	client.GetEarningsForUserForDay(ctx, date, sigType, "")
//	client.GetCurrentRewards(ctx, "")
//	client.GetBuilderFeeRate(ctx, "builder-code")
//
// See README.md for the full endpoint reference.
package clob
