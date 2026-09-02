package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/mycelis/framework-runs/internal/auth"
	"github.com/mycelis/framework-runs/internal/controller"
	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

const maxRequestBytes = 1 << 20

type Server struct {
	service *controller.Service
	handler http.Handler
}

func New(service *controller.Service, authenticator *auth.Authenticator) *Server {
	server := &Server{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /v1/capabilities", server.capabilities)
	mux.HandleFunc("POST /v1/runs", server.create)
	mux.HandleFunc("GET /v1/runs/{run_id}", server.get)
	mux.HandleFunc("GET /v1/runs/{run_id}/events", server.events)
	mux.HandleFunc("POST /v1/runs/{run_id}/stop", server.stop)
	mux.HandleFunc("POST /v1/runs/{run_id}/approvals/{approval_id}", server.approve)
	server.handler = securityHeaders(authenticator.Middleware("runs:api",
		rejectNonCanonicalPath(rejectQuery(requireCanonicalRoute(mux)))))
	return server
}

func (server *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	server.handler.ServeHTTP(response, request)
}

func (server *Server) health(response http.ResponseWriter, request *http.Request) {
	if err := server.service.Health(request.Context()); err != nil {
		writeError(response, http.StatusServiceUnavailable, "unready", "Framework Runs service is not ready.", true)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{
		"healthy": true, "message": "framework Runs service ready",
		"backend": "framework_runs", "protocol": "runs_api",
	})
}

func (server *Server) capabilities(response http.ResponseWriter, request *http.Request) {
	if err := server.service.Health(request.Context()); err != nil {
		writeError(response, http.StatusServiceUnavailable, "unready", "Framework Runs service is not ready.", true)
		return
	}
	writeJSON(response, http.StatusOK, server.service.Capabilities())
}

func (server *Server) create(response http.ResponseWriter, request *http.Request) {
	var payload protocol.CreateRequest
	if err := decodeStrict(response, request, &payload); err != nil {
		writeError(response, http.StatusUnprocessableEntity, "invalid_request", "Request body does not match the Runs API contract.", false)
		return
	}
	run, _, err := server.service.Create(request.Context(), payload)
	if err != nil {
		writeCreateError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, run)
}

func (server *Server) get(response http.ResponseWriter, request *http.Request) {
	runID := request.PathValue("run_id")
	if protocol.ValidateExternalID("run_id", runID) != nil {
		writeError(response, http.StatusNotFound, "run_not_found", "Run was not found.", false)
		return
	}
	run, err := server.service.Get(request.Context(), runID)
	if err != nil {
		writeServiceError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, run)
}

func (server *Server) stop(response http.ResponseWriter, request *http.Request) {
	runID := request.PathValue("run_id")
	if protocol.ValidateExternalID("run_id", runID) != nil {
		writeError(response, http.StatusNotFound, "run_not_found", "Run was not found.", false)
		return
	}
	var payload protocol.StopRequest
	if err := decodeStrict(response, request, &payload); err != nil {
		writeError(response, http.StatusUnprocessableEntity, "invalid_request", "Request body does not match the Runs API contract.", false)
		return
	}
	principal, _ := auth.PrincipalFromContext(request.Context())
	receipt, replay, err := server.service.Stop(request.Context(), runID, payload, principal.ID)
	if err != nil {
		writeControlError(response, err)
		return
	}
	status := http.StatusAccepted
	if replay {
		status = http.StatusOK
	}
	writeJSON(response, status, receipt)
}

func (server *Server) approve(response http.ResponseWriter, request *http.Request) {
	runID := request.PathValue("run_id")
	approvalID := request.PathValue("approval_id")
	if protocol.ValidateExternalID("run_id", runID) != nil || protocol.ValidateExternalID("approval_id", approvalID) != nil {
		writeError(response, http.StatusNotFound, "run_not_found", "Run or approval was not found.", false)
		return
	}
	var payload protocol.ApprovalDecisionRequest
	if err := decodeStrict(response, request, &payload); err != nil {
		writeError(response, http.StatusUnprocessableEntity, "invalid_request", "Request body does not match the Runs API contract.", false)
		return
	}
	if payload.ApprovalID != approvalID {
		writeError(response, http.StatusConflict, "approval_mismatch", "Approval id does not match route.", false)
		return
	}
	principal, _ := auth.PrincipalFromContext(request.Context())
	receipt, replay, err := server.service.Decide(request.Context(), runID, approvalID, payload, principal.ID)
	if err != nil {
		writeControlError(response, err)
		return
	}
	status := http.StatusAccepted
	if replay {
		status = http.StatusOK
	}
	writeJSON(response, status, receipt)
}

func decodeStrict(response http.ResponseWriter, request *http.Request, target any) error {
	mediaType, parameters, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || (parameters["charset"] != "" && !strings.EqualFold(parameters["charset"], "utf-8")) {
		return errors.New("content type must be application/json")
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("request body must contain exactly one JSON value")
	}
	return nil
}

func rejectNonCanonicalPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		cleaned := path.Clean(request.URL.Path)
		if request.URL.Path != cleaned || (request.URL.Path != "/" && strings.HasSuffix(request.URL.Path, "/")) {
			writeError(response, http.StatusNotFound, "http_error", "Route was not found.", false)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func rejectQuery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.RawQuery != "" {
			writeError(response, http.StatusUnprocessableEntity, "invalid_request", "Query controls are not supported.", false)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(response, request)
	})
}

func writeServiceError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, controller.ErrExecutorUnavailable), errors.Is(err, journal.ErrCapacity):
		writeError(response, http.StatusServiceUnavailable, "unready", "Framework Runs service cannot accept work.", true)
	case errors.Is(err, journal.ErrNotFound):
		writeError(response, http.StatusNotFound, "run_not_found", "Run was not found.", false)
	case errors.Is(err, journal.ErrCursorGap):
		writeError(response, http.StatusConflict, "cursor_gap", "The requested event cursor is unavailable.", true)
	case errors.Is(err, journal.ErrConflict), errors.Is(err, journal.ErrLeaseLost):
		writeError(response, http.StatusConflict, "command_conflict", "The request conflicts with durable run state.", true)
	case strings.HasPrefix(err.Error(), "invalid "):
		writeError(response, http.StatusUnprocessableEntity, "invalid_request", "Request body does not match the Runs API contract.", false)
	default:
		writeError(response, http.StatusServiceUnavailable, "unready", "Framework Runs service is temporarily unavailable.", true)
	}
}

func writeCreateError(response http.ResponseWriter, err error) {
	if errors.Is(err, journal.ErrRunConflict) {
		writeError(response, http.StatusConflict, "run_conflict", "Run id already exists with different content.", false)
		return
	}
	writeServiceError(response, err)
}

func writeControlError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, controller.ErrExecutorUnavailable):
		writeError(response, http.StatusServiceUnavailable, "unready", "Framework Runs service cannot apply controls.", true)
	case errors.Is(err, journal.ErrVersionConflict):
		writeError(response, http.StatusConflict, "version_conflict", "Run version does not match expected_version.", true)
	case errors.Is(err, journal.ErrCommandConflict), errors.Is(err, journal.ErrLeaseLost):
		writeError(response, http.StatusConflict, "command_conflict", "Command conflicts with durable run state.", true)
	case errors.Is(err, journal.ErrInvalidRunState):
		writeError(response, http.StatusConflict, "invalid_run_state", "Run state does not allow this control.", false)
	case errors.Is(err, journal.ErrApprovalNotFound):
		writeError(response, http.StatusNotFound, "approval_not_found", "Approval was not found.", false)
	default:
		writeServiceError(response, err)
	}
}

func requireCanonicalRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		expected, known := canonicalMethod(request.URL.Path)
		if !known {
			writeError(response, http.StatusNotFound, "http_error", "Route was not found.", false)
			return
		}
		if request.Method != expected {
			writeError(response, http.StatusMethodNotAllowed, "http_error", "Method is not allowed.", false)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func canonicalMethod(route string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(route, "/"), "/")
	if len(parts) == 1 && parts[0] == "health" {
		return http.MethodGet, true
	}
	if len(parts) == 2 && parts[0] == "v1" && parts[1] == "capabilities" {
		return http.MethodGet, true
	}
	if len(parts) == 2 && parts[0] == "v1" && parts[1] == "runs" {
		return http.MethodPost, true
	}
	if len(parts) >= 3 && parts[0] == "v1" && parts[1] == "runs" && parts[2] != "" {
		switch {
		case len(parts) == 3:
			return http.MethodGet, true
		case len(parts) == 4 && parts[3] == "events":
			return http.MethodGet, true
		case len(parts) == 4 && parts[3] == "stop":
			return http.MethodPost, true
		case len(parts) == 5 && parts[3] == "approvals" && parts[4] != "":
			return http.MethodPost, true
		}
	}
	return "", false
}

func writeError(response http.ResponseWriter, status int, code, message string, recoverable bool) {
	writeJSON(response, status, protocol.ErrorEnvelope{Error: protocol.Error{
		Code: code, Message: message, Recoverable: recoverable,
	}})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func decimalCursor(header http.Header) (uint64, error) {
	values := header.Values("Last-Event-ID")
	if len(values) == 0 {
		return 0, nil
	}
	if len(values) != 1 || values[0] == "" || strings.TrimSpace(values[0]) != values[0] {
		return 0, errors.New("invalid cursor")
	}
	for _, digit := range values[0] {
		if digit < '0' || digit > '9' {
			return 0, errors.New("invalid cursor")
		}
	}
	return strconv.ParseUint(values[0], 10, 64)
}

func terminal(status protocol.Status) bool {
	return status == protocol.StatusCompleted || status == protocol.StatusFailed || status == protocol.StatusCancelled
}
