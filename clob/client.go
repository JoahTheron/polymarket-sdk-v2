package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/bububa/polymarket-client/internal/polyauth"
	"github.com/bububa/polymarket-client/relayer"
)

var errMissingIdentifier = errors.New("polymarket: missing identifier on output value")

type Client struct {
	host          string
	geoblock      string
	rpcURL        string
	httpClient    *http.Client
	auth          Auth
	userAgent     string
	relayerClient RelayerSubmitter
}

// RelayerSubmitter is the relayer capability used by CLOB CTF helpers.
type RelayerSubmitter interface {
	SubmitTransaction(context.Context, relayer.SubmitTransactionRequest, *relayer.SubmitTransactionResponse) error
}

// Option customizes a CLOB client created by NewClient.
type Option func(*Client)

// WithHTTPClient sets the HTTP client used for REST requests.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithCredentials sets Polymarket API credentials for L2-authenticated requests.
func WithCredentials(creds Credentials) Option {
	return func(c *Client) { c.auth.Credentials = &creds }
}

// WithSigner sets the signer used for L1 and L2 authenticated requests.
func WithSigner(signer *polyauth.Signer) Option {
	return func(c *Client) { c.auth.Signer = signer }
}

// WithChainID sets the EVM chain ID used for auth signatures and CTF transactions.
func WithChainID(chainID int64) Option {
	return func(c *Client) { c.auth.ChainID = chainID }
}

// WithServerTime makes authenticated requests use the CLOB server timestamp.
func WithServerTime(enabled bool) Option {
	return func(c *Client) { c.auth.UseServerTime = enabled }
}

// WithGeoblockHost sets the host used by geoblock-related requests.
func WithGeoblockHost(host string) Option {
	return func(c *Client) { c.geoblock = strings.TrimRight(host, "/") }
}

// WithRPCURL sets the Polygon JSON-RPC endpoint used by on-chain CTF helpers.
func WithRPCURL(rpcURL string) Option {
	return func(c *Client) { c.rpcURL = rpcURL }
}

// WithRelayerSubmitter sets the relayer used by SubmitRelayerTransaction helpers.
func WithRelayerSubmitter(submitter RelayerSubmitter) Option {
	return func(c *Client) { c.relayerClient = submitter }
}

// WithRelayerClient sets the Polymarket Relayer API client used by CTF helpers.
func WithRelayerClient(client *relayer.Client) Option {
	return WithRelayerSubmitter(client)
}

// NewClient creates a CLOB client for host.
func NewClient(host string, opts ...Option) *Client {
	if host == "" {
		host = V2Host
	}
	c := &Client{
		host:       strings.TrimRight(host, "/"),
		geoblock:   "https://polymarket.com",
		rpcURL:     "https://polygon-rpc.com",
		httpClient: http.DefaultClient,
		userAgent:  "polymarket-client-go/clob",
		auth:       Auth{ChainID: PolygonChainID},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Host() string { return c.host }

// SubmitRelayerTransaction submits a pre-signed transaction through the configured relayer.
func (c *Client) SubmitRelayerTransaction(ctx context.Context, req relayer.SubmitTransactionRequest, out *relayer.SubmitTransactionResponse) error {
	if c.relayerClient == nil {
		return errors.New("polymarket: relayer client is not configured")
	}
	return c.relayerClient.SubmitTransaction(ctx, req, out)
}

type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("polymarket API error: status %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("polymarket API error: status %d", e.StatusCode)
}

func (c *Client) GetOk(ctx context.Context) (string, error) {
	var out string
	err := c.do(ctx, http.MethodGet, "/ok", nil, nil, 0, &out)
	return out, err
}

func (c *Client) GetVersion(ctx context.Context) (int, error) {
	var out int
	err := c.do(ctx, http.MethodGet, "/version", nil, nil, 0, &out)
	return out, err
}

func (c *Client) GetServerTime(ctx context.Context) (int64, error) {
	var raw any
	if err := c.do(ctx, http.MethodGet, "/time", nil, nil, 0, &raw); err != nil {
		return 0, err
	}
	return scalarInt64(raw)
}

func (c *Client) GetSamplingSimplifiedMarkets(ctx context.Context, cursor string, out *Page[SimplifiedMarket]) error {
	return c.getPage(ctx, "/sampling-simplified-markets", cursor, out)
}

func (c *Client) GetSamplingMarkets(ctx context.Context, cursor string, out *Page[Market]) error {
	return c.getPage(ctx, "/sampling-markets", cursor, out)
}

func (c *Client) GetSimplifiedMarkets(ctx context.Context, cursor string, out *Page[SimplifiedMarket]) error {
	return c.getPage(ctx, "/simplified-markets", cursor, out)
}

func (c *Client) GetMarkets(ctx context.Context, cursor string, out *Page[Market]) error {
	return c.getPage(ctx, "/markets", cursor, out)
}

func (c *Client) GetMarket(ctx context.Context, out *Market) error {
	if out == nil || out.ConditionID == "" {
		return errMissingIdentifier
	}
	return c.do(ctx, http.MethodGet, "/markets/"+url.PathEscape(out.ConditionID), nil, nil, 0, out)
}

func (c *Client) GetMarketByToken(ctx context.Context, out *MarketByToken) error {
	if out == nil || (out.PrimaryTokenID == "" && out.SecondaryTokenID == "") {
		return errMissingIdentifier
	}
	tokenID := out.PrimaryTokenID.String()
	if tokenID == "" {
		tokenID = out.SecondaryTokenID.String()
	}
	return c.do(ctx, http.MethodGet, "/markets-by-token/"+url.PathEscape(tokenID), nil, nil, 0, out)
}

func (c *Client) GetClobMarketInfo(ctx context.Context, out *ClobMarketInfo) error {
	if out == nil || out.ConditionID == "" {
		return errMissingIdentifier
	}
	return c.do(ctx, http.MethodGet, "/clob-markets/"+url.PathEscape(out.ConditionID), nil, nil, 0, out)
}

func (c *Client) GetOrderBook(ctx context.Context, out *OrderBookSummary) error {
	if out == nil || out.AssetID == "" {
		return errMissingIdentifier
	}
	q := url.Values{"token_id": []string{out.AssetID.String()}}
	return c.do(ctx, http.MethodGet, "/book", q, nil, 0, out)
}

func (c *Client) GetOrderBooks(ctx context.Context, books []BookParams) ([]OrderBookSummary, error) {
	var out []OrderBookSummary
	return out, c.do(ctx, http.MethodPost, "/books", nil, books, 0, &out)
}

func (c *Client) GetMidpoint(ctx context.Context, tokenID string, out *MidpointResponse) error {
	return c.do(ctx, http.MethodGet, "/midpoint", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetMidpoints(ctx context.Context, books []BookParams) (map[string]Float64, error) {
	var out map[string]Float64
	return out, c.do(ctx, http.MethodPost, "/midpoints", nil, books, 0, &out)
}

func (c *Client) GetPrice(ctx context.Context, tokenID string, side Side, out *PriceResponse) error {
	return c.do(ctx, http.MethodGet, "/price", url.Values{"token_id": []string{tokenID}, "side": []string{string(side)}}, nil, 0, out)
}

func (c *Client) GetPrices(ctx context.Context, books []BookParams) (map[string]map[Side]Float64, error) {
	var out map[string]map[Side]Float64
	return out, c.do(ctx, http.MethodPost, "/prices", nil, books, 0, &out)
}

func (c *Client) GetSpread(ctx context.Context, tokenID string, out *SpreadResponse) error {
	return c.do(ctx, http.MethodGet, "/spread", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetSpreads(ctx context.Context, books []BookParams) (map[string]Float64, error) {
	var out map[string]Float64
	return out, c.do(ctx, http.MethodPost, "/spreads", nil, books, 0, &out)
}

func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string, out *LastTradePriceResponse) error {
	return c.do(ctx, http.MethodGet, "/last-trade-price", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetLastTradesPrices(ctx context.Context, books []BookParams) ([]LastTradesPricesResponse, error) {
	var out []LastTradesPricesResponse
	return out, c.do(ctx, http.MethodPost, "/last-trades-prices", nil, books, 0, &out)
}

func (c *Client) GetTickSize(ctx context.Context, tokenID string, out *TickSizeResponse) error {
	return c.do(ctx, http.MethodGet, "/tick-size", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetTickSizeByTokenID(ctx context.Context, tokenID string, out *TickSizeResponse) error {
	return c.do(ctx, http.MethodGet, "/tick-size/"+url.PathEscape(tokenID), nil, nil, 0, out)
}

func (c *Client) GetNegRisk(ctx context.Context, tokenID string, out *NegRiskResponse) error {
	return c.do(ctx, http.MethodGet, "/neg-risk", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetFeeRate(ctx context.Context, tokenID string, out *FeeRateResponse) error {
	return c.do(ctx, http.MethodGet, "/fee-rate", url.Values{"token_id": []string{tokenID}}, nil, 0, out)
}

func (c *Client) GetFeeRateByTokenID(ctx context.Context, tokenID string, out *FeeRateResponse) error {
	return c.do(ctx, http.MethodGet, "/fee-rate/"+url.PathEscape(tokenID), nil, nil, 0, out)
}

func (c *Client) GetPricesHistory(ctx context.Context, params PriceHistoryParams, out *PriceHistoryResponse) error {
	return c.do(ctx, http.MethodGet, "/prices-history", values(params), nil, 0, out)
}

func (c *Client) GetBatchPricesHistory(ctx context.Context, params BatchPriceHistoryParams, out *BatchPriceHistoryResponse) error {
	return c.do(ctx, http.MethodPost, "/batch-prices-history", nil, params, 0, out)
}

func (c *Client) GetCurrentRebates(ctx context.Context, params RebateParams) ([]Rebate, error) {
	var out []Rebate
	return out, c.do(ctx, http.MethodGet, "/rebates/current", values(params), nil, 0, &out)
}

func (c *Client) CreateAPIKey(ctx context.Context, nonce int64, out *Credentials) error {
	return c.do(ctx, http.MethodPost, "/auth/api-key", nil, nil, 1, out, nonce)
}

func (c *Client) DeriveAPIKey(ctx context.Context, nonce int64, out *Credentials) error {
	return c.do(ctx, http.MethodGet, "/auth/derive-api-key", nil, nil, 1, out, nonce)
}

func (c *Client) GetAPIKeys(ctx context.Context) ([]Credentials, error) {
	var out apiKeysResponse
	return out.APIKeys, c.do(ctx, http.MethodGet, "/auth/api-keys", nil, nil, 2, &out)
}

func (c *Client) DeleteAPIKey(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/auth/api-key", nil, nil, 2, nil)
}

func (c *Client) GetClosedOnlyMode(ctx context.Context, out *BanStatus) error {
	return c.do(ctx, http.MethodGet, "/auth/ban-status/closed-only", nil, nil, 2, out)
}

func (c *Client) GetOrder(ctx context.Context, out *OpenOrder) error {
	if out == nil || out.ID == "" {
		return errMissingIdentifier
	}
	return c.do(ctx, http.MethodGet, "/data/order/"+url.PathEscape(out.ID), nil, nil, 2, out)
}

func (c *Client) GetOpenOrders(ctx context.Context, params OpenOrderParams) ([]OpenOrder, error) {
	var out []OpenOrder
	return out, c.do(ctx, http.MethodGet, "/data/orders", values(params), nil, 2, &out)
}

func (c *Client) GetPreMigrationOrders(ctx context.Context, params OpenOrderParams) ([]OpenOrder, error) {
	var out []OpenOrder
	return out, c.do(ctx, http.MethodGet, "/data/pre-migration-orders", values(params), nil, 2, &out)
}

func (c *Client) GetTrades(ctx context.Context, params TradeParams) ([]Trade, error) {
	var out []Trade
	return out, c.do(ctx, http.MethodGet, "/data/trades", values(params), nil, 2, &out)
}

func (c *Client) PostOrder(ctx context.Context, req PostOrderRequest, out *PostOrderResponse) error {
	return c.do(ctx, http.MethodPost, "/order", nil, req, 2, out)
}

func (c *Client) PostOrders(ctx context.Context, reqs []PostOrderRequest, postOnly, deferExec bool) ([]PostOrderResponse, error) {
	q := url.Values{}
	if postOnly {
		q.Set("post_only", "true")
	}
	if deferExec {
		q.Set("defer_exec", "true")
	}
	var out []PostOrderResponse
	return out, c.do(ctx, http.MethodPost, "/orders", q, reqs, 2, &out)
}

func (c *Client) CancelOrder(ctx context.Context, orderID string, out *CancelOrdersResponse) error {
	return c.do(ctx, http.MethodDelete, "/order", nil, map[string]string{"orderID": orderID}, 2, out)
}

func (c *Client) CancelOrders(ctx context.Context, orderIDs []string, out *CancelOrdersResponse) error {
	return c.do(ctx, http.MethodDelete, "/orders", nil, orderIDs, 2, out)
}

func (c *Client) CancelAll(ctx context.Context, out *CancelOrdersResponse) error {
	return c.do(ctx, http.MethodDelete, "/cancel-all", nil, nil, 2, out)
}

func (c *Client) CancelMarketOrders(ctx context.Context, params OrderMarketCancelParams, out *CancelOrdersResponse) error {
	return c.do(ctx, http.MethodDelete, "/cancel-market-orders", nil, params, 2, out)
}

func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	var out []Notification
	return out, c.do(ctx, http.MethodGet, "/notifications", nil, nil, 2, &out)
}

func (c *Client) DropNotifications(ctx context.Context, params DropNotificationParams) error {
	return c.do(ctx, http.MethodDelete, "/notifications", values(params), nil, 2, nil)
}

func (c *Client) GetBalanceAllowance(ctx context.Context, params BalanceAllowanceParams, out *BalanceAllowanceResponse) error {
	return c.do(ctx, http.MethodGet, "/balance-allowance", values(params), nil, 2, out)
}

func (c *Client) UpdateBalanceAllowance(ctx context.Context, params BalanceAllowanceParams, out *BalanceAllowanceResponse) error {
	return c.do(ctx, http.MethodPost, "/balance-allowance/update", nil, params, 2, out)
}

func (c *Client) IsOrderScoring(ctx context.Context, orderID string, out *OrderScoring) error {
	return c.do(ctx, http.MethodGet, "/order-scoring", url.Values{"order_id": []string{orderID}}, nil, 2, out)
}

func (c *Client) AreOrdersScoring(ctx context.Context, orderIDs []string) (map[string]bool, error) {
	var out map[string]bool
	return out, c.do(ctx, http.MethodPost, "/orders-scoring", nil, map[string][]string{"orderIds": orderIDs}, 2, &out)
}

func (c *Client) PostHeartbeat(ctx context.Context, heartbeatID string, out *HeartbeatResponse) error {
	body := map[string]string{}
	if heartbeatID != "" {
		body["heartbeat_id"] = heartbeatID
	}
	return c.do(ctx, http.MethodPost, "/v1/heartbeats", nil, body, 2, out)
}

func (c *Client) GetEarningsForUserForDay(ctx context.Context, date string, signatureType SignatureType, cursor string, out *Page[UserEarning]) error {
	q := url.Values{"date": []string{date}, "signature_type": []string{strconv.Itoa(int(signatureType))}}
	if cursor != "" {
		q.Set("next_cursor", cursor)
	}
	return c.do(ctx, http.MethodGet, "/rewards/user", q, nil, 2, out)
}

func (c *Client) GetTotalEarningsForUserForDay(ctx context.Context, date string, signatureType SignatureType, out *UserEarning) error {
	q := url.Values{"date": []string{date}, "signature_type": []string{strconv.Itoa(int(signatureType))}}
	return c.do(ctx, http.MethodGet, "/rewards/user/total", q, nil, 2, out)
}

func (c *Client) GetRewardPercentages(ctx context.Context, signatureType SignatureType) (map[string]Float64, error) {
	var out map[string]Float64
	q := url.Values{"signature_type": []string{strconv.Itoa(int(signatureType))}}
	return out, c.do(ctx, http.MethodGet, "/rewards/user/percentages", q, nil, 2, &out)
}

func (c *Client) GetUserEarningsAndMarketsConfig(ctx context.Context, params EarningsParams, signatureType SignatureType, out *Page[UserRewardsEarning]) error {
	q := values(params)
	q.Set("signature_type", strconv.Itoa(int(signatureType)))
	return c.do(ctx, http.MethodGet, "/rewards/user/markets", q, nil, 2, out)
}

func (c *Client) GetCurrentRewards(ctx context.Context, cursor string, out *Page[CurrentReward]) error {
	return c.getPage(ctx, "/rewards/markets/current", cursor, out)
}

func (c *Client) GetRewardsForMarket(ctx context.Context, conditionID, cursor string, out *Page[MarketReward]) error {
	q := url.Values{}
	if cursor != "" {
		q.Set("next_cursor", cursor)
	}
	return c.do(ctx, http.MethodGet, "/rewards/markets/"+url.PathEscape(conditionID), q, nil, 0, out)
}

func (c *Client) CreateBuilderAPIKey(ctx context.Context, out *Credentials) error {
	return c.do(ctx, http.MethodPost, "/auth/builder-api-key", nil, nil, 2, out)
}

func (c *Client) GetBuilderAPIKeys(ctx context.Context) ([]BuilderAPIKey, error) {
	var out []BuilderAPIKey
	return out, c.do(ctx, http.MethodGet, "/auth/builder-api-key", nil, nil, 2, &out)
}

func (c *Client) RevokeBuilderAPIKey(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/auth/builder-api-key", nil, nil, 2, nil)
}

func (c *Client) GetBuilderTrades(ctx context.Context, params BuilderTradeParams, out *Page[BuilderTrade]) error {
	return c.do(ctx, http.MethodGet, "/builder/trades", values(params), nil, 2, out)
}

func (c *Client) GetBuilderFeeRate(ctx context.Context, builderCode string, out *BuilderFeeRate) error {
	return c.do(ctx, http.MethodGet, "/fees/builder-fees/"+url.PathEscape(builderCode), nil, nil, 0, out)
}

func (c *Client) CreateReadonlyAPIKey(ctx context.Context, out *ReadonlyAPIKey) error {
	return c.do(ctx, http.MethodPost, "/auth/readonly-api-key", nil, nil, 2, out)
}

func (c *Client) GetReadonlyAPIKeys(ctx context.Context) ([]ReadonlyAPIKey, error) {
	var out []ReadonlyAPIKey
	return out, c.do(ctx, http.MethodGet, "/auth/readonly-api-keys", nil, nil, 2, &out)
}

func (c *Client) DeleteReadonlyAPIKey(ctx context.Context, key string) error {
	return c.do(ctx, http.MethodDelete, "/auth/readonly-api-key", nil, map[string]string{"key": key}, 2, nil)
}

func (c *Client) GetMarketTradesEvents(ctx context.Context, conditionID string) ([]Trade, error) {
	var out []Trade
	return out, c.do(ctx, http.MethodGet, "/markets/live-activity/"+url.PathEscape(conditionID), nil, nil, 0, &out)
}

func (c *Client) CreateRFQRequest(ctx context.Context, req CreateRFQRequest) (map[string]any, error) {
	var out map[string]any
	return out, c.do(ctx, http.MethodPost, "/rfq/request", nil, req, 2, &out)
}

func (c *Client) CancelRFQRequest(ctx context.Context, requestID string) error {
	return c.do(ctx, http.MethodDelete, "/rfq/request", nil, CancelRFQRequest{RequestID: requestID}, 2, nil)
}

func (c *Client) GetRFQRequests(ctx context.Context, params RFQListParams, out *Page[RfqRequest]) error {
	return c.do(ctx, http.MethodGet, "/rfq/data/requests", values(params), nil, 2, out)
}

func (c *Client) CreateRFQQuote(ctx context.Context, req CreateRFQQuoteRequest) (map[string]any, error) {
	var out map[string]any
	return out, c.do(ctx, http.MethodPost, "/rfq/quote", nil, req, 2, &out)
}

func (c *Client) CancelRFQQuote(ctx context.Context, quoteID string) error {
	return c.do(ctx, http.MethodDelete, "/rfq/quote", nil, CancelRFQQuoteRequest{QuoteID: quoteID}, 2, nil)
}

func (c *Client) GetRFQRequesterQuotes(ctx context.Context, params RFQListParams, out *Page[RfqQuote]) error {
	return c.do(ctx, http.MethodGet, "/rfq/data/requester/quotes", values(params), nil, 2, out)
}

func (c *Client) GetRFQQuoterQuotes(ctx context.Context, params RFQListParams, out *Page[RfqQuote]) error {
	return c.do(ctx, http.MethodGet, "/rfq/data/quoter/quotes", values(params), nil, 2, out)
}

func (c *Client) GetRFQBestQuote(ctx context.Context, params RFQListParams, out *RfqQuote) error {
	return c.do(ctx, http.MethodGet, "/rfq/data/best-quote", values(params), nil, 2, out)
}

func (c *Client) AcceptRFQRequest(ctx context.Context, requestID string) error {
	return c.do(ctx, http.MethodPost, "/rfq/request/accept", nil, map[string]string{"requestId": requestID}, 2, nil)
}

func (c *Client) ApproveRFQQuote(ctx context.Context, quoteID string) (map[string]any, error) {
	var out map[string]any
	return out, c.do(ctx, http.MethodPost, "/rfq/quote/approve", nil, map[string]string{"quoteId": quoteID}, 2, &out)
}

func (c *Client) GetRFQConfig(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	return out, c.do(ctx, http.MethodGet, "/rfq/config", nil, nil, 0, &out)
}

func (c *Client) getPage(ctx context.Context, path, cursor string, out any) error {
	q := url.Values{}
	if cursor != "" {
		q.Set("next_cursor", cursor)
	}
	return c.do(ctx, http.MethodGet, path, q, nil, 0, out)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, authLevel int, out any, nonceOpt ...int64) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	fullURL := c.host + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}
	var r io.Reader
	if len(bodyBytes) > 0 {
		r = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if len(bodyBytes) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := c.addAuthHeaders(ctx, req, method, path, bodyBytes, authLevel, nonceOpt...); err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Message: string(data), Body: data}
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if p, ok := out.(*string); ok {
		*p = strings.Trim(string(data), `"`)
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *Client) addAuthHeaders(ctx context.Context, req *http.Request, method, path string, body []byte, level int, nonceOpt ...int64) error {
	if level == 0 {
		return nil
	}
	if c.auth.Signer == nil {
		return errors.New("polymarket: signer is required for authenticated request")
	}
	ts := nowUnix()
	if c.auth.UseServerTime {
		serverTS, err := c.GetServerTime(ctx)
		if err != nil {
			return err
		}
		ts = serverTS
	}
	if level == 1 {
		nonce := int64(0)
		if len(nonceOpt) > 0 {
			nonce = nonceOpt[0]
		}
		headers, err := polyauth.L1Headers(c.auth.Signer, c.auth.ChainID, ts, nonce)
		if err != nil {
			return err
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		return nil
	}
	if c.auth.Credentials == nil {
		return errors.New("polymarket: api credentials are required for level 2 authenticated request")
	}
	secret, err := polyauth.DecodeAPISecret(c.auth.Credentials.Secret)
	if err != nil {
		return err
	}
	headers, err := polyauth.L2Headers(c.auth.Signer, c.auth.Credentials.Key, secret, c.auth.Credentials.Passphrase, ts, method, path, body)
	if err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return nil
}

func values(v any) url.Values {
	q := url.Values{}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return q
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return q
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		name := sf.Tag.Get("url")
		if name == "-" {
			continue
		}
		if name == "" {
			name = sf.Name
		}
		parts := strings.Split(name, ",")
		key := parts[0]
		omitempty := len(parts) > 1 && parts[1] == "omitempty"
		fv := rv.Field(i)
		if omitempty && fv.IsZero() {
			continue
		}
		switch fv.Kind() {
		case reflect.Slice:
			for j := 0; j < fv.Len(); j++ {
				q.Add(key, fmt.Sprint(fv.Index(j).Interface()))
			}
		default:
			q.Set(key, fmt.Sprint(fv.Interface()))
		}
	}
	return q
}

func scalarInt64(v any) (int64, error) {
	switch x := v.(type) {
	case float64:
		return int64(x), nil
	case string:
		return strconv.ParseInt(x, 10, 64)
	case map[string]any:
		for _, k := range []string{"time", "timestamp"} {
			if y, ok := x[k]; ok {
				return scalarInt64(y)
			}
		}
	}
	return 0, fmt.Errorf("polymarket: cannot parse server time from %T", v)
}
