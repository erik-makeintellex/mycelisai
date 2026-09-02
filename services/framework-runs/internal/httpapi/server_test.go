package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/framework-runs/internal/auth"
	"github.com/mycelis/framework-runs/internal/controller"
	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

const testToken = "0123456789abcdef0123456789abcdef"

type completingExecutor struct{}

func (completingExecutor) Apply(_ context.Context, command journal.Command) (protocol.ExecutorOutcome, error) {
	return protocol.ExecutorOutcome{Status: protocol.StatusCompleted, Result: &protocol.Result{
		Summary: "Candidate ready.", Outputs: []protocol.Output{}, Metadata: map[string]any{},
		FinishedAt: time.Date(2026, 9, 2, 12, 0, 1, 0, time.UTC),
	}}, nil
}

func TestEveryRouteRequiresScopedBearerAuth(t *testing.T) {
	server := newTestServer(t, nil, "runs:api")
	routes := []struct{ method, path string }{
		{"GET", "/health"}, {"GET", "/v1/capabilities"}, {"POST", "/v1/runs"},
		{"GET", "/v1/runs/run-1"}, {"GET", "/v1/runs/run-1/events"},
		{"POST", "/v1/runs/run-1/stop"}, {"POST", "/v1/runs/run-1/approvals/approval-1"},
	}
	for _, route := range routes {
		request := httptest.NewRequest(route.method, route.path, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d", route.method, route.path, response.Code)
		}
		assertExactErrorShape(t, response.Body.Bytes())
	}
	wrongScope := newTestServer(t, nil, "health:read")
	request := authorized("GET", "/health", "")
	response := httptest.NewRecorder()
	wrongScope.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("wrong scope = %d", response.Code)
	}
}

func TestHealthReadyButCapabilitiesAndCreateFailClosedWithoutExecutor(t *testing.T) {
	repository := journal.NewMemoryRepository()
	server := newServer(t, controller.New(repository, nil), "runs:api")
	response := do(server, authorized("GET", "/health", ""))
	if response.Code != http.StatusOK {
		t.Fatalf("health = %d: %s", response.Code, response.Body.String())
	}
	response = do(server, authorized("GET", "/v1/capabilities", ""))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"production_ready":false`) {
		t.Fatalf("capabilities = %d: %s", response.Code, response.Body.String())
	}
	response = do(server, authorized("POST", "/v1/runs", createJSON("run-unready")))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("create = %d: %s", response.Code, response.Body.String())
	}
	if _, err := repository.Get(context.Background(), "run-unready"); err != journal.ErrNotFound {
		t.Fatal("unready create was persisted")
	}
}

func TestStrictRoutesBodiesAndQueryControls(t *testing.T) {
	server := newTestServer(t, completingExecutor{}, "runs:api")
	cases := []*http.Request{
		authorized("POST", "/v1/runs?wait=true", createJSON("run-query")),
		authorized("POST", "/v1/runs", strings.TrimSuffix(createJSON("run-extra"), "}")+`,"unknown":true}`),
		authorized("GET", "/v1/runs/", ""),
		authorized("GET", "/v1//capabilities", ""),
	}
	for _, request := range cases {
		response := do(server, request)
		if response.Code < 400 || response.Code >= 500 {
			t.Errorf("%s = %d: %s", request.URL.String(), response.Code, response.Body.String())
		}
		assertExactErrorShape(t, response.Body.Bytes())
	}
	wrongMethod := do(server, authorized("GET", "/v1/runs", ""))
	if wrongMethod.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong method = %d", wrongMethod.Code)
	}
	assertExactErrorShape(t, wrongMethod.Body.Bytes())
}

func TestExactV1ErrorCodes(t *testing.T) {
	repository := journal.NewMemoryRepository()
	service := controller.New(repository, completingExecutor{})
	server := newServer(t, service, "runs:api")
	malformed := do(server, authorized("POST", "/v1/runs", `{}`))
	assertErrorCode(t, malformed, http.StatusUnprocessableEntity, "invalid_request")
	notFound := do(server, authorized("GET", "/v1/runs/missing", ""))
	assertErrorCode(t, notFound, http.StatusNotFound, "run_not_found")
	if response := do(server, authorized("POST", "/v1/runs", createJSON("run-conflict"))); response.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", response.Code, response.Body.String())
	}
	changed := strings.Replace(createJSON("run-conflict"), "Produce candidate", "Different candidate", 1)
	conflict := do(server, authorized("POST", "/v1/runs", changed))
	assertErrorCode(t, conflict, http.StatusConflict, "run_conflict")
	stop := `{"command_id":"stop-stale","expected_version":9,"actor_id":"core"}`
	version := do(server, authorized("POST", "/v1/runs/run-conflict/stop", stop))
	assertErrorCode(t, version, http.StatusConflict, "version_conflict")
	cursorRequest := authorized("GET", "/v1/runs/run-conflict/events", "")
	cursorRequest.Header.Set("Last-Event-ID", "+1")
	invalidCursor := do(server, cursorRequest)
	assertErrorCode(t, invalidCursor, http.StatusUnprocessableEntity, "invalid_cursor")
}

func TestSSEReplaysExactSequenceAfterCursor(t *testing.T) {
	repository := journal.NewMemoryRepository()
	service := controller.New(repository, completingExecutor{})
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	server := newServer(t, service, "runs:api")
	create := do(server, authorized("POST", "/v1/runs", createJSON("run-events")))
	if create.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", create.Code, create.Body.String())
	}
	now = now.Add(time.Second)
	if _, err := service.DispatchOnce(context.Background(), "worker"); err != nil {
		t.Fatal(err)
	}
	request := authorized("GET", "/v1/runs/run-events/events", "")
	request.Header.Set("Last-Event-ID", "1")
	response := do(server, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "id: 2\ndata: {") {
		t.Fatalf("SSE = %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "id: 1") {
		t.Fatal("SSE replay included cursor event")
	}
	request = authorized("GET", "/v1/runs/run-events/events", "")
	request.Header.Set("Last-Event-ID", "3")
	response = do(server, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"cursor_gap"`) {
		t.Fatalf("ahead cursor = %d: %s", response.Code, response.Body.String())
	}
}

func newTestServer(t *testing.T, worker interface {
	Apply(context.Context, journal.Command) (protocol.ExecutorOutcome, error)
}, scope string) *Server {
	return newServer(t, controller.New(journal.NewMemoryRepository(), worker), scope)
}

func newServer(t *testing.T, service *controller.Service, scope string) *Server {
	t.Helper()
	credential, err := auth.NewCredential("core", testToken, scope)
	if err != nil {
		t.Fatal(err)
	}
	authenticator, err := auth.New(credential)
	if err != nil {
		t.Fatal(err)
	}
	return New(service, authenticator)
}

func authorized(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+testToken)
	if body != "" {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	return request
}

func do(server *Server, request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func assertExactErrorShape(t *testing.T, raw []byte) {
	t.Helper()
	var decoded map[string]any
	if json.Unmarshal(raw, &decoded) != nil || len(decoded) != 1 || decoded["error"] == nil {
		t.Fatalf("invalid error envelope: %s", raw)
	}
	errorBody, ok := decoded["error"].(map[string]any)
	if !ok || len(errorBody) != 3 {
		t.Fatalf("invalid error body: %s", raw)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status || !strings.Contains(response.Body.String(), `"code":"`+code+`"`) {
		t.Fatalf("error = %d: %s, want %d/%s", response.Code, response.Body.String(), status, code)
	}
	assertExactErrorShape(t, response.Body.Bytes())
}

func createJSON(runID string) string {
	return `{"run_id":"` + runID + `","intent":"Produce candidate","correlation":{` +
		`"run_id":"` + runID + `","intent_proof_id":"proof-1","execution_contract_id":"contract-1",` +
		`"work_item_id":"work-1","idempotency_key":"idem:` + runID + `","source_kind":"web_api",` +
		`"source_channel":"api.intent","payload_kind":"command","graph_revision":"graph-1"}}`
}
