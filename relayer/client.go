package relayer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bububa/polymarket-client/internal/polyauth"
	"github.com/bububa/polymarket-client/internal/polyhttp"
)

var errMissingIdentifier = errors.New("polymarket: missing identifier on output value")

// DefaultHost is the production Polymarket Relayer API host.
const DefaultHost = "https://relayer-v2.polymarket.com"

// Client is a Polymarket Relayer API client.
type Client struct {
	host       string
	httpClient *http.Client
	userAgent  string
	creds      *Credentials
	builder    *BuilderCredentials
}

// Config configures a Relayer API client.
type Config struct {
	// Host is the Relayer API host.
	Host string
	// HTTPClient is the client used for HTTP requests.
	HTTPClient *http.Client
	// UserAgent sets the User-Agent header.
	UserAgent string
	// Credentials configures Relayer API-key authentication.
	Credentials *Credentials
	// BuilderCredentials configures builder-key authentication.
	BuilderCredentials *BuilderCredentials
}

// New creates a Relayer API client.
func New(config Config) *Client {
	if config.Host == "" {
		config.Host = DefaultHost
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if config.UserAgent == "" {
		config.UserAgent = "polymarket-client-go/relayer"
	}
	return &Client{
		host:       strings.TrimRight(config.Host, "/"),
		httpClient: config.HTTPClient,
		userAgent:  config.UserAgent,
		creds:      config.Credentials,
		builder:    config.BuilderCredentials,
	}
}

// Host returns the configured Relayer API host.
func (c *Client) Host() string { return c.host }

// SubmitTransaction submits a signed transaction to the relayer and writes the response into out.
func (c *Client) SubmitTransaction(ctx context.Context, req SubmitTransactionRequest, out *SubmitTransactionResponse) error {
	return c.do(ctx, http.MethodPost, "/submit", nil, req, out)
}

// GetTransaction writes a relayer transaction by out.TransactionID into out.
func (c *Client) GetTransaction(ctx context.Context, out *Transaction) error {
	if out == nil || out.TransactionID == "" {
		return errMissingIdentifier
	}
	var raw json.RawMessage
	q := url.Values{"id": []string{out.TransactionID}}
	if err := c.do(ctx, http.MethodGet, "/transaction", q, nil, &raw); err != nil {
		return err
	}
	payload := bytes.TrimSpace(raw)
	if bytes.HasPrefix(payload, []byte("[")) {
		var rows []Transaction
		if err := json.Unmarshal(payload, &rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("relayer: transaction %q not found", out.TransactionID)
		}
		*out = rows[0]
		return nil
	}
	return json.Unmarshal(payload, out)
}

// GetRecentTransactions returns recent transactions owned by the authenticated address.
func (c *Client) GetRecentTransactions(ctx context.Context, _ ...string) ([]Transaction, error) {
	var out []Transaction
	return out, c.do(ctx, http.MethodGet, "/transactions", nil, nil, &out)
}

// GetNonce writes the current relayer nonce for out.Address into out.
func (c *Client) GetNonce(ctx context.Context, out *NonceResponse, nonceType ...NonceType) error {
	if out == nil || out.Address == "" {
		return errMissingIdentifier
	}
	q := url.Values{"address": []string{out.Address}}
	if len(nonceType) > 0 && nonceType[0] != "" {
		q.Set("type", string(nonceType[0]))
	}
	return c.do(ctx, http.MethodGet, "/nonce", q, nil, out)
}

// GetRelayPayload writes the relayer address and nonce for out.Address into out.
func (c *Client) GetRelayPayload(ctx context.Context, out *NonceResponse, nonceType NonceType) error {
	if out == nil || out.Address == "" {
		return errMissingIdentifier
	}
	q := url.Values{"address": []string{out.Address}}
	if nonceType != "" {
		q.Set("type", string(nonceType))
	}
	return c.do(ctx, http.MethodGet, "/relay-payload", q, nil, out)
}

// GetRelayerNonce writes the relayer address and nonce for out.Address into out.
func (c *Client) GetRelayerNonce(ctx context.Context, out *NonceResponse, nonceType ...NonceType) error {
	var typ NonceType
	if len(nonceType) > 0 {
		typ = nonceType[0]
	}
	return c.GetRelayPayload(ctx, out, typ)
}

// IsSafeDeployed writes whether out.Address is deployed into out.
func (c *Client) IsSafeDeployed(ctx context.Context, out *SafeDeployedResponse) error {
	if out == nil || out.Address == "" {
		return errMissingIdentifier
	}
	q := url.Values{"address": []string{out.Address}}
	return c.do(ctx, http.MethodGet, "/deployed", q, nil, out)
}

// GetAPIKeys returns all relayer API keys available to the authenticated caller.
func (c *Client) GetAPIKeys(ctx context.Context) ([]APIKey, error) {
	var out []APIKey
	return out, c.do(ctx, http.MethodGet, "/relayer/api/keys", nil, nil, &out)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	payload, err := json.Marshal(body)
	if body == nil {
		payload = nil
	} else if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	fullURL := c.host + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}
	var reader io.Reader
	if len(payload) > 0 {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.creds != nil {
		req.Header.Set("RELAYER_API_KEY", c.creds.APIKey)
		req.Header.Set("RELAYER_API_KEY_ADDRESS", c.creds.Address)
	}
	if c.builder != nil {
		ts := time.Now().Unix()
		sig, err := polyauth.HMACSignature(c.builder.Secret, ts, method, path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("POLY_BUILDER_API_KEY", c.builder.APIKey)
		req.Header.Set("POLY_BUILDER_TIMESTAMP", strconv.FormatInt(ts, 10))
		req.Header.Set("POLY_BUILDER_PASSPHRASE", c.builder.Passphrase)
		req.Header.Set("POLY_BUILDER_SIGNATURE", sig)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return &polyhttp.APIError{StatusCode: resp.StatusCode, Message: string(data), Body: data, RequestBody: payload}
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
