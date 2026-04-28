package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JoahTheron/polymarket-sdk-v2/internal/polyhttp"
)

const DefaultHost = "https://data-api.polymarket.com"

type Client struct {
	host string
	http *polyhttp.Client
}

// Config configures a Data API client.
type Config struct {
	Host       string
	HTTPClient *http.Client
	UserAgent  string
}

// New creates a Data API client.
func New(config Config) *Client {
	if config.Host == "" {
		config.Host = DefaultHost
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if config.UserAgent == "" {
		config.UserAgent = "polymarket-client-go/data"
	}
	return &Client{
		host: config.Host,
		http: &polyhttp.Client{BaseURL: config.Host, HTTPClient: config.HTTPClient, UserAgent: config.UserAgent},
	}
}

// Host returns the configured Data API host.
func (c *Client) Host() string { return c.host }

// GetHealth writes the Data API health response into out.
func (c *Client) GetHealth(ctx context.Context, out *Health) error {
	return c.http.GetJSON(ctx, "/", nil, polyhttp.AuthNone, out)
}

// GetPositions returns current positions for a user.
func (c *Client) GetPositions(ctx context.Context, params PositionParams) ([]Position, error) {
	var out []Position
	return out, c.http.GetJSON(ctx, "/positions", params.values(), polyhttp.AuthNone, &out)
}

// GetMarketPositions returns positions grouped by outcome token for a market.
func (c *Client) GetMarketPositions(ctx context.Context, params MarketPositionsParams) ([]MarketPositions, error) {
	var out []MarketPositions
	return out, c.http.GetJSON(ctx, "/v1/market-positions", params.values(), polyhttp.AuthNone, &out)
}

// GetClosedPositions returns closed positions for a user.
func (c *Client) GetClosedPositions(ctx context.Context, params ClosedPositionParams) ([]ClosedPosition, error) {
	var out []ClosedPosition
	return out, c.http.GetJSON(ctx, "/closed-positions", params.values(), polyhttp.AuthNone, &out)
}

// GetValue returns the total value of a user's positions.
func (c *Client) GetValue(ctx context.Context, user string, markets []string) ([]Value, error) {
	var out []Value
	q := url.Values{"user": []string{user}}
	setCommaList(q, "market", markets)
	return out, c.http.GetJSON(ctx, "/value", q, polyhttp.AuthNone, &out)
}

// GetTrades returns user or market trades.
func (c *Client) GetTrades(ctx context.Context, params TradeParams) ([]Trade, error) {
	var out []Trade
	return out, c.http.GetJSON(ctx, "/trades", params.values(), polyhttp.AuthNone, &out)
}

// GetActivity returns user activity.
func (c *Client) GetActivity(ctx context.Context, params ActivityParams) ([]Activity, error) {
	var out []Activity
	return out, c.http.GetJSON(ctx, "/activity", params.values(), polyhttp.AuthNone, &out)
}

// GetHolders returns top holders for markets.
func (c *Client) GetHolders(ctx context.Context, params HoldersParams) ([]Holder, error) {
	var out []Holder
	return out, c.http.GetJSON(ctx, "/holders", params.values(), polyhttp.AuthNone, &out)
}

// GetTraded writes the total markets a user has traded into out.
func (c *Client) GetTraded(ctx context.Context, user string, out *Traded) error {
	return c.http.GetJSON(ctx, "/traded", url.Values{"user": []string{user}}, polyhttp.AuthNone, out)
}

// GetOpenInterest returns open interest for markets.
func (c *Client) GetOpenInterest(ctx context.Context, markets []string) ([]OpenInterest, error) {
	var out []OpenInterest
	q := url.Values{}
	setCommaList(q, "market", markets)
	return out, c.http.GetJSON(ctx, "/oi", q, polyhttp.AuthNone, &out)
}

// GetLiveVolume returns live volume for an event or markets.
func (c *Client) GetLiveVolume(ctx context.Context, params LiveVolumeParams) ([]LiveVolume, error) {
	var out []LiveVolume
	return out, c.http.GetJSON(ctx, "/live-volume", params.values(), polyhttp.AuthNone, &out)
}

// GetLeaderboard returns trader leaderboard rankings.
func (c *Client) GetLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, error) {
	var out []LeaderboardEntry
	return out, c.http.GetJSON(ctx, "/v1/leaderboard", params.values(), polyhttp.AuthNone, &out)
}

// GetBuilderLeaderboard returns aggregated builder leaderboard rows.
func (c *Client) GetBuilderLeaderboard(ctx context.Context, params BuilderLeaderboardParams) ([]BuilderLeaderboardEntry, error) {
	var out []BuilderLeaderboardEntry
	return out, c.http.GetJSON(ctx, "/v1/builders/leaderboard", params.values(), polyhttp.AuthNone, &out)
}

// GetBuilderVolume returns daily builder volume time-series rows.
func (c *Client) GetBuilderVolume(ctx context.Context, params BuilderVolumeParams) ([]BuilderVolume, error) {
	var out []BuilderVolume
	return out, c.http.GetJSON(ctx, "/v1/builders/volume", params.values(), polyhttp.AuthNone, &out)
}

// DownloadAccountingSnapshot downloads the accounting snapshot ZIP for a user.
func (c *Client) DownloadAccountingSnapshot(ctx context.Context, user string) ([]byte, error) {
	q := url.Values{"user": []string{user}}
	fullURL := c.http.BaseURL + "/v1/accounting/snapshot?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/zip")
	req.Header.Set("User-Agent", c.http.UserAgent)
	resp, err := c.http.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, &polyhttp.APIError{StatusCode: resp.StatusCode, Message: string(bytes.TrimSpace(payload)), Body: payload}
	}
	return payload, nil
}

func setInt(q url.Values, key string, val int) {
	if val > 0 {
		q.Set(key, strconv.Itoa(val))
	}
}

func setInt64(q url.Values, key string, val int64) {
	if val > 0 {
		q.Set(key, strconv.FormatInt(val, 10))
	}
}

func setString(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func setBool(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func setCommaList[T ~string](q url.Values, key string, vals []T) {
	if len(vals) == 0 {
		return
	}
	parts := make([]string, len(vals))
	for i, val := range vals {
		parts[i] = string(val)
	}
	q.Set(key, strings.Join(parts, ","))
}

func setIntList(q url.Values, key string, vals []int) {
	if len(vals) == 0 {
		return
	}
	parts := make([]string, 0, len(vals))
	for _, val := range vals {
		if val > 0 {
			parts = append(parts, strconv.Itoa(val))
		}
	}
	if len(parts) > 0 {
		q.Set(key, strings.Join(parts, ","))
	}
}
