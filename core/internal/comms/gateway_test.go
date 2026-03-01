package comms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGateway_Send_WebhookProvider(t *testing.T) {
	var gotBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	g := NewGateway()
	g.Register(newWebhookProvider("slack", "chat", "Slack", ts.URL, nil))

	res, err := g.Send(context.Background(), SendRequest{
		Provider:  "slack",
		Recipient: "#ops",
		Message:   "health check",
		Metadata:  map[string]any{"source": "test"},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if res.Status != "sent" {
		t.Fatalf("status = %q", res.Status)
	}
	if gotBody["recipient"] != "#ops" {
		t.Fatalf("recipient = %v", gotBody["recipient"])
	}
	if gotBody["message"] != "health check" {
		t.Fatalf("message = %v", gotBody["message"])
	}
}

func TestGateway_Send_UnconfiguredProvider(t *testing.T) {
	g := NewGateway()
	g.Register(newWebhookProvider("slack", "chat", "Slack", "", nil))

	_, err := g.Send(context.Background(), SendRequest{
		Provider: "slack",
		Message:  "ping",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestGateway_ListProviders(t *testing.T) {
	g := NewGateway()
	g.Register(newWebhookProvider("slack", "chat", "Slack", "", nil))
	g.Register(newWebhookProvider("webhook", "automation", "Generic", "http://example", nil))

	providers := g.ListProviders()
	if len(providers) != 2 {
		t.Fatalf("providers len = %d", len(providers))
	}
	if providers[0].Name != "slack" {
		t.Fatalf("providers sorted order unexpected: %#v", providers)
	}
}
