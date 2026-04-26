package rtds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestClientSubscribesAndDecodesMessage(t *testing.T) {
	gotSub := make(chan SubscriptionRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := context.Background()
		var sub SubscriptionRequest
		if err := wsjson.Read(ctx, conn, &sub); err != nil {
			t.Errorf("read subscription: %v", err)
			return
		}
		gotSub <- sub

		if err := wsjson.Write(ctx, conn, Message{
			Topic:     "crypto_prices",
			Type:      "update",
			Timestamp: 1700000000,
			Payload:   []byte(`{"symbol":"BTCUSDT","timestamp":1700000000,"value":"65000"}`),
		}); err != nil {
			t.Errorf("write message: %v", err)
		}

		// Keep connection alive until test completes
		<-ctx.Done()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient(url).WithAutoReconnect(false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.SubscribeCryptoPrices(ctx, []string{"BTCUSDT"}); err != nil {
		t.Fatal(err)
	}

	select {
	case sub := <-gotSub:
		if sub.Action != ActionSubscribe || len(sub.Subscriptions) != 1 || sub.Subscriptions[0].Topic != "crypto_prices" {
			t.Fatalf("unexpected subscription: %#v", sub)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for subscription")
	}

	select {
	case msg := <-client.Messages():
		var price CryptoPrice
		if err := msg.AsCryptoPrice(&price); err != nil {
			t.Fatal(err)
		}
		if price.Symbol != "BTCUSDT" || price.Value != "65000" {
			t.Fatalf("unexpected price: %#v", price)
		}
	case err := <-client.Errors():
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}
