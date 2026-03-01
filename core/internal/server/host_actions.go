package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mycelis/core/internal/hostcmd"
	"github.com/mycelis/core/pkg/protocol"
)

type invokeHostActionRequest struct {
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	TimeoutMS int      `json:"timeout_ms,omitempty"`
}

// HandleHostStatus exposes local host command status and environment profile.
// GET /api/v1/host/status
func (s *AdminServer) HandleHostStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "host:read"); !ok {
		return
	}
	if r.Method != http.MethodGet {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	host, _ := os.Hostname()
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if strings.TrimSpace(workspace) == "" {
		workspace = "./workspace"
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"os":                  runtime.GOOS,
		"arch":                runtime.GOARCH,
		"hostname":            host,
		"workspace":           workspace,
		"local_command_v0":    true,
		"allowed_commands":    hostcmd.AllowedCommands(),
		"allowlist_env_field": "MYCELIS_LOCAL_COMMAND_ALLOWLIST",
	}))
}

// HandleHostActions lists supported host actuation actions.
// GET /api/v1/host/actions
func (s *AdminServer) HandleHostActions(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "host:read"); !ok {
		return
	}
	if r.Method != http.MethodGet {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actions := []map[string]any{
		{
			"action_id":         "local-command",
			"display_name":      "Local Command (Allowlist)",
			"description":       "Execute one allowlisted local command without shell interpolation.",
			"risk_level":        "medium",
			"requires_approval": false,
			"arguments_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command":    map[string]any{"type": "string"},
					"args":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"timeout_ms": map[string]any{"type": "integer"},
				},
				"required": []string{"command"},
			},
			"allowed_commands": hostcmd.AllowedCommands(),
		},
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(actions))
}

// HandleInvokeHostAction executes one supported host action.
// POST /api/v1/host/actions/{id}/invoke
func (s *AdminServer) HandleInvokeHostAction(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "host:invoke"); !ok {
		return
	}
	if r.Method != http.MethodPost {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actionID := strings.TrimSpace(r.PathValue("id"))
	if actionID == "" {
		respondAPIError(w, "Missing action ID", http.StatusBadRequest)
		return
	}

	if actionID != "local-command" {
		respondAPIError(w, "Unknown host action: "+actionID, http.StatusNotFound)
		return
	}

	var req invokeHostActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Command) == "" {
		respondAPIError(w, "command is required", http.StatusBadRequest)
		return
	}

	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	result, err := hostcmd.Execute(r.Context(), req.Command, req.Args, timeout)
	if err != nil {
		switch {
		case errors.Is(err, hostcmd.ErrCommandNotAllowed):
			respondAPIError(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, hostcmd.ErrInvalidArguments):
			respondAPIError(w, err.Error(), http.StatusBadRequest)
		default:
			respondAPIError(w, "Failed to invoke local command: "+err.Error(), http.StatusBadGateway)
		}
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"action_id": "local-command",
		"result":    result,
	}))
}
