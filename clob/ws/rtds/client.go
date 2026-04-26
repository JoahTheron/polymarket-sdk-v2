package rtds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"go.uber.org/atomic"
)

const (
	// DefaultHost is the production Polymarket RTDS WebSocket URL.
	DefaultHost = "wss://rtds.polymarket.com"
)

// Client is a reconnecting WebSocket client for Polymarket RTDS topics.
type Client struct {
	url      string
	dialOpts *websocket.DialOptions

	creds  *atomic.Pointer[Credentials]
	conn   *atomic.Pointer[websocket.Conn]
	closed *atomic.Bool
	ctx    context.Context
	cancel context.CancelFunc
	msgs   chan *Message
	errs   chan error

	autoReconnect bool
	reconnecting  *atomic.Bool

	subsMu sync.RWMutex
	subs   []Subscription
}

// Config configures an RTDS client.
type Config struct {
	// URL is the RTDS WebSocket URL.
	URL string
	// Header is sent during the WebSocket handshake.
	Header http.Header
	// Credentials are used when subscribing to authenticated topics.
	Credentials *Credentials
}

// New creates an RTDS client.
func New(config Config) *Client {
	url := config.URL
	if url == "" {
		url = DefaultHost
	}
	ctx, cancel := context.WithCancel(context.Background())
	clt := &Client{
		url:           url,
		creds:         atomic.NewPointer[Credentials](config.Credentials),
		ctx:           ctx,
		cancel:        cancel,
		msgs:          make(chan *Message, 1024),
		errs:          make(chan error, 64),
		autoReconnect: true,

		conn:         atomic.NewPointer[websocket.Conn](nil),
		reconnecting: atomic.NewBool(false),
		closed:       atomic.NewBool(false),
	}
	if len(config.Header) > 0 {
		clt.dialOpts = &websocket.DialOptions{
			HTTPHeader: config.Header.Clone(),
		}
	}
	return clt
}

// NewClient creates an RTDS client for url.
func NewClient(url string) *Client {
	return New(Config{URL: url})
}

// WithCredentials sets credentials for authenticated topic subscriptions.
func (c *Client) WithCredentials(creds *Credentials) *Client {
	c.creds.Store(creds)
	return c
}

// WithAutoReconnect enables or disables automatic reconnect after read failures.
func (c *Client) WithAutoReconnect(enabled bool) *Client {
	c.autoReconnect = enabled
	return c
}

// Connect opens the RTDS WebSocket and starts read loop.
func (c *Client) Connect(ctx context.Context) error {
	return c.connect(ctx)
}

func (c *Client) connect(ctx context.Context) error {
	if c.closed.Load() {
		return errors.New("polymarket: RTDS client is closed")
	}

	conn, _, err := websocket.Dial(ctx, c.url, c.dialOpts)
	if err != nil {
		return fmt.Errorf("polymarket: RTDS dial: %w", err)
	}

	if oldConn := c.conn.Swap(conn); oldConn != nil {
		_ = oldConn.CloseNow()
	}

	go c.readLoop(ctx, conn)
	c.replaySubscriptions(ctx)
	return nil
}

// Close closes the active RTDS connection and stops background loops.
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	c.cancel()
	if conn := c.conn.Load(); conn != nil {
		_ = conn.CloseNow()
	}
	return nil
}

// IsConnected reports whether a WebSocket connection is currently attached.
func (c *Client) IsConnected() bool {
	return !c.closed.Load() && c.conn.Load() != nil
}

// Messages returns decoded RTDS messages.
func (c *Client) Messages() <-chan *Message { return c.msgs }

// Errors returns asynchronous RTDS connection and decode errors.
func (c *Client) Errors() <-chan error { return c.errs }

// Subscribe sends and records a topic subscription.
func (c *Client) Subscribe(ctx context.Context, sub Subscription) error {
	c.subsMu.Lock()
	c.subs = append(c.subs, sub)
	c.subsMu.Unlock()
	return c.sendJSON(ctx, SubscriptionRequest{Action: ActionSubscribe, Subscriptions: []Subscription{sub}})
}

// Unsubscribe sends and removes a topic subscription.
func (c *Client) Unsubscribe(ctx context.Context, sub Subscription) error {
	c.subsMu.Lock()
	c.removeSubscriptionLocked(sub)
	c.subsMu.Unlock()
	return c.sendJSON(ctx, SubscriptionRequest{Action: ActionUnsubscribe, Subscriptions: []Subscription{sub}})
}

// SubscribeCryptoPrices subscribes to Binance crypto price updates.
func (c *Client) SubscribeCryptoPrices(ctx context.Context, symbols []string) error {
	var filters any
	if len(symbols) > 0 {
		filters = symbols
	}
	return c.Subscribe(ctx, Subscription{Topic: "crypto_prices", Type: "update", Filters: filters})
}

// SubscribeChainlinkPrices subscribes to Chainlink crypto price updates.
func (c *Client) SubscribeChainlinkPrices(ctx context.Context, symbol string) error {
	var filters any
	if symbol != "" {
		filters = map[string]string{"symbol": symbol}
	}
	return c.Subscribe(ctx, Subscription{Topic: "crypto_prices_chainlink", Type: "*", Filters: filters})
}

// SubscribeComments subscribes to comment events.
func (c *Client) SubscribeComments(ctx context.Context, commentType CommentType, creds *Credentials) error {
	if creds == nil {
		creds = c.creds.Load()
	}
	eventType := string(commentType)
	if eventType == "" {
		eventType = "*"
	}
	return c.Subscribe(ctx, Subscription{Topic: "comments", Type: eventType, CLOBAuth: creds})
}

func (c *Client) replaySubscriptions(ctx context.Context) {
	c.subsMu.RLock()
	subs := append([]Subscription(nil), c.subs...)
	c.subsMu.RUnlock()
	if len(subs) == 0 {
		return
	}
	if err := c.sendJSON(ctx, SubscriptionRequest{Action: ActionSubscribe, Subscriptions: subs}); err != nil {
		c.sendErr(fmt.Errorf("polymarket: replay RTDS subscriptions: %w", err))
	}
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if c.ctx.Err() != nil || websocket.CloseStatus(err) != -1 {
				return
			}
			c.sendErr(fmt.Errorf("polymarket: RTDS read: %w", err))
			c.scheduleReconnect(conn)
			return
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendErr(fmt.Errorf("polymarket: decode RTDS message: %w", err))
			continue
		}
		select {
		case c.msgs <- &msg:
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) sendJSON(ctx context.Context, v any) error {
	conn := c.conn.Load()
	if conn == nil {
		return errors.New("polymarket: RTDS websocket is not connected")
	}
	return wsjson.Write(ctx, conn, v)
}

func (c *Client) scheduleReconnect(conn *websocket.Conn) {
	if c.closed.Load() || !c.autoReconnect {
		return
	}
	if !c.reconnecting.CompareAndSwap(false, true) {
		return
	}
	if !c.conn.CompareAndSwap(conn, nil) {
		_ = conn.CloseNow()
		c.reconnecting.CompareAndSwap(true, false)
		return
	}
	go func() {
		defer func() {
			c.reconnecting.CompareAndSwap(true, false)
		}()
		backoff := time.Second
		for {
			if c.closed.Load() {
				return
			}
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(backoff):
			}
			ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
			err := c.connect(ctx)
			cancel()
			if err == nil {
				return
			}
			c.sendErr(err)
			if backoff < time.Minute {
				backoff *= 2
			}
		}
	}()
}

func (c *Client) removeSubscriptionLocked(target Subscription) {
	for idx := len(c.subs) - 1; idx >= 0; idx-- {
		sub := c.subs[idx]
		if sub.Topic == target.Topic && sub.Type == target.Type {
			c.subs = append(c.subs[:idx], c.subs[idx+1:]...)
			return
		}
	}
}

func (c *Client) sendErr(err error) {
	select {
	case c.errs <- err:
	default:
	}
}
