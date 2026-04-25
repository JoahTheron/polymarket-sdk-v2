package relayer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRelayerDocumentedEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transaction":
			if r.URL.Query().Get("id") != "tx-1" {
				t.Fatalf("transaction query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"transactionID":"tx-1","data":"0x01","value":"","signature":"0xsig","type":"SAFE","owner":"0xowner"}]`))
		case "/transactions":
			_, _ = w.Write([]byte(`[]`))
		case "/nonce":
			if r.URL.Query().Get("address") != "0xabc" || r.URL.Query().Get("type") != "SAFE" {
				t.Fatalf("nonce query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"nonce":"31"}`))
		case "/relay-payload":
			if r.URL.Query().Get("address") != "0xabc" || r.URL.Query().Get("type") != "PROXY" {
				t.Fatalf("relay-payload query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"address":"0xrelayer","nonce":"32"}`))
		case "/deployed":
			if r.URL.Query().Get("address") != "0xsafe" {
				t.Fatalf("deployed query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"deployed":true}`))
		case "/relayer/api/keys":
			_, _ = w.Write([]byte(`[{"apiKey":"key-1","address":"0xabc","createdAt":"2026-02-24T18:20:11.237485Z","updatedAt":"2026-02-24T18:20:11.237485Z"}]`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	client := New(Config{Host: srv.URL})
	tx := Transaction{TransactionID: "tx-1"}
	if err := client.GetTransaction(context.Background(), &tx); err != nil {
		t.Fatal(err)
	}
	if tx.TransactionID != "tx-1" || tx.Data != "0x01" || tx.Type != "SAFE" {
		t.Fatalf("unexpected transaction: %+v", tx)
	}
	if _, err := client.GetRecentTransactions(context.Background()); err != nil {
		t.Fatal(err)
	}
	nonce := NonceResponse{Address: "0xabc"}
	if err := client.GetNonce(context.Background(), &nonce, NonceTypeSafe); err != nil {
		t.Fatal(err)
	}
	payload := NonceResponse{Address: "0xabc"}
	if err := client.GetRelayPayload(context.Background(), &payload, NonceTypeProxy); err != nil {
		t.Fatal(err)
	}
	deployed := SafeDeployedResponse{Address: "0xsafe"}
	if err := client.IsSafeDeployed(context.Background(), &deployed); err != nil {
		t.Fatal(err)
	}
	keys, err := client.GetAPIKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0].Key != "key-1" || keys[0].UpdatedAt.IsZero() {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}
