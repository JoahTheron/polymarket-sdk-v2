// Package gamma provides a read-only client for the Polymarket Gamma market-data API.
//
// # Overview
//
// The Gamma API exposes market and event metadata, full-text search, tag
// relationships, sports data, public profiles, and user comments. All endpoints
// are public (no auth required).
//
// # Usage
//
//	client := gamma.New(gamma.Config{})  // defaults to gamma-api.polymarket.com
//
//	// Search markets, events, profiles
//	var results gamma.SearchResults
//	client.Search(ctx, "election", &results)
//
//	// Get events and markets
//	events, _ := client.GetEvents(ctx, gamma.EventFilterParams{Active: truePtr})
//	markets, _ := client.GetMarkets(ctx, gamma.MarketFilterParams{Slug: "..."})
//
//	// Tags and related content
//	tags, _ := client.GetTags(ctx)
//	related, _ := client.GetRelatedTags(ctx, 12, gamma.RelatedTagParams{})
//
//	// Public profiles and comments
//	profile := gamma.PublicProfile{Address: "0x..."}
//	client.GetPublicProfile(ctx, &profile)
//	comments, _ := client.GetComments(ctx, gamma.CommentFilterParams{})
//
// Default host: https://gamma-api.polymarket.com
package gamma
