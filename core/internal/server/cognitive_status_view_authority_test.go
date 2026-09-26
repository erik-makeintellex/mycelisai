package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

// S6e F2: only root admin with cognitive:read or cognitive:write sees
// endpoints, models, config snapshots and override detail on
// GET /cognitive/status, /cognitive/config and /brains.

// secretEndpoint looks like a credentialed URL; it must never reach a
// caller without the full view.
const secretEndpoint = "http://svc-user:sk-live-S6E-SECRET@127.0.0.1:9/v1"

// withCognitiveAdmin attaches the local admin ("*") identity to req.
func withCognitiveAdmin(req *http.Request) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, localAdminIdentityForTest()))
}

var narrowViewCallers = []struct {
	name     string
	identity *RequestIdentity
}{
	{"anonymous", nil},
	{"standard user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"soma:work", "runs:read", "outputs:read"}}},
	{"non-admin with cognitive:read", &RequestIdentity{UserID: "u-2", Role: "user", Scopes: []string{"cognitive:read", "*"}}},
	{"admin without cognitive scope", &RequestIdentity{UserID: "u-3", Role: "admin", Scopes: []string{"config_documents:write"}}},
}

var fullViewCallers = []struct {
	name     string
	identity *RequestIdentity
}{
	{"admin cognitive:read", &RequestIdentity{UserID: "r-1", Role: "admin", Scopes: []string{"cognitive:read"}}},
	{"admin cognitive:write", &RequestIdentity{UserID: "r-2", Role: "admin", Scopes: []string{"cognitive:write"}}},
	{"admin wildcard", localAdminIdentityForTest()},
}

func newCognitiveViewFixture(t *testing.T) *routingFixture {
	t.Helper()
	f := newRoutingFixture(t)
	vllm, _ := f.s.Cognitive.ProviderSnapshot("vllm")
	vllm.Endpoint = secretEndpoint
	f.s.Cognitive.StoreProviderConfig("vllm", vllm)
	f.mux.HandleFunc("GET /api/v1/cognitive/status", f.s.HandleCognitiveStatus)
	f.mux.HandleFunc("GET /api/v1/cognitive/config", f.s.HandleCognitiveConfig)
	return f
}

func viewRequest(t *testing.T, f *routingFixture, path string, identity *RequestIdentity) *httptest.ResponseRecorder {
	t.Helper()
	if identity == nil {
		return doRequest(t, f.mux, http.MethodGet, path, "")
	}
	return doAuthenticatedRequestAs(t, f.mux, http.MethodGet, path, "", identity)
}

func sortedKeys(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func assertNoSecretEndpoint(t *testing.T, body string) {
	t.Helper()
	for _, leak := range []string{"S6E-SECRET", "127.0.0.1", "endpoint", "db_row_present", "override_origin", "model_id", "root_provider"} {
		if strings.Contains(body, leak) {
			t.Fatalf("narrow view leaked %q: %s", leak, body)
		}
	}
}

func TestCognitiveStatusView_NarrowForCallersWithoutFullView(t *testing.T) {
	for _, caller := range narrowViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/cognitive/status", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			assertNoSecretEndpoint(t, rr.Body.String())
			var resp map[string]any
			assertJSON(t, rr, &resp)
			if got := sortedKeys(resp); got != "media,overlay_error,profile_route_health,profiles,text" {
				t.Fatalf("top-level keys = %s", got)
			}
			text := resp["text"].(map[string]any)
			if text["status"] != "online" {
				t.Fatalf("text = %v", text)
			}
			for _, absent := range []string{"endpoint", "model", "provider_id"} {
				if _, ok := text[absent]; ok {
					t.Fatalf("text.%s present: %v", absent, text)
				}
			}
			if got := sortedKeys(resp["media"].(map[string]any)); got != "status" {
				t.Fatalf("media keys = %s", got)
			}
			profiles := resp["profiles"].(map[string]any)
			chat, ok := profiles["chat"].(map[string]any)
			if !ok || chat["available"] != true || chat["code"] == nil {
				t.Fatalf("chat summary = %v", profiles["chat"])
			}
			for name, p := range profiles {
				if got := sortedKeys(p.(map[string]any)); got != "available,code,reachable" {
					t.Fatalf("profile %s keys = %s", name, got)
				}
			}
		})
	}
}

func TestCognitiveStatusView_FullForScopedRootAdmin(t *testing.T) {
	for _, caller := range fullViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/cognitive/status", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			var resp map[string]any
			assertJSON(t, rr, &resp)
			text := resp["text"].(map[string]any)
			if text["endpoint"] != secretEndpoint || text["provider_id"] != "vllm" || text["model"] != "m-vllm" {
				t.Fatalf("admin text view incomplete: %v", text)
			}
			if resp["root_provider"] != "vllm" || resp["root_provider_model"] != "m-vllm" {
				t.Fatalf("admin root view incomplete: %v", resp)
			}
			chat := resp["profiles"].(map[string]any)["chat"].(map[string]any)
			if chat["provider_id"] != "vllm" || chat["source"] == nil {
				t.Fatalf("admin profile view incomplete: %v", chat)
			}
		})
	}
}

func TestCognitiveConfigView_RequiresScopedRootAdmin(t *testing.T) {
	for _, caller := range narrowViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/cognitive/config", caller.identity)
			want := http.StatusForbidden
			if caller.identity == nil {
				want = http.StatusUnauthorized
			}
			assertStatus(t, rr, want)
			assertNoSecretEndpoint(t, rr.Body.String())
			if strings.Contains(rr.Body.String(), "providers") {
				t.Fatalf("config leaked: %s", rr.Body.String())
			}
		})
	}
	for _, caller := range fullViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/cognitive/config", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			body := rr.Body.String()
			var resp map[string]any
			assertJSON(t, rr, &resp)
			if !strings.Contains(body, "S6E-SECRET") || resp["providers"] == nil {
				t.Fatalf("admin config incomplete: %s", body)
			}
		})
	}
}

func TestBrainListView_NarrowForCallersWithoutFullView(t *testing.T) {
	for _, caller := range narrowViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/brains", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			assertNoSecretEndpoint(t, rr.Body.String())
			var resp struct {
				Data []map[string]any `json:"data"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || len(resp.Data) != 5 {
				t.Fatalf("decode: %v body=%s", err, rr.Body.String())
			}
			for _, entry := range resp.Data {
				if got := sortedKeys(entry); got != "data_boundary,enabled,id,location,status,type" {
					t.Fatalf("brain %v keys = %s", entry["id"], got)
				}
			}
		})
	}
	for _, caller := range fullViewCallers {
		t.Run(caller.name, func(t *testing.T) {
			f := newCognitiveViewFixture(t)
			rr := viewRequest(t, f, "/api/v1/brains", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			if body := rr.Body.String(); !strings.Contains(body, "S6E-SECRET") || !strings.Contains(body, `"model_id":"m-vllm"`) {
				t.Fatalf("admin brains view incomplete: %s", body)
			}
		})
	}
}
