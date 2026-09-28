package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// MCPS: authority for the operator-direct MCP call route
// POST /api/v1/mcp/servers/{id}/tools/{tool}/call (contract
// MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md, D1-D8). Core enforces it; the BFF,
// MCP annotations, the model and the request body never lower a class or
// grant a call.

const (
	mcpDirectCallAuditSource = "mcp.direct_call"
	mcpDirectReadScope       = "outputs:read"
	mcpServerStatusConnected = "connected"
)

// directMCPReadTools is the one code-owned direct-call read allowlist (D3),
// keyed by exact server name and exact tool name. Every other MCP tool is
// classified by capabilityRiskForTool, which rates every mcp:* name high.
var directMCPReadTools = map[string]map[string]bool{
	"filesystem": {
		"list_directory": true, "list_directory_with_sizes": true, "directory_tree": true,
		"read_text_file": true, "read_file": true, "read_media_file": true, "read_multiple_files": true,
		"get_file_info": true, "search_files": true, "list_allowed_directories": true,
	},
}

// directMCPToolRisk is "low" only for an allowlisted read; anything else is
// the fail-closed classifier's answer ("high" for every mcp:* name).
func directMCPToolRisk(serverName, toolName string, args map[string]any) string {
	if directMCPReadTools[serverName][toolName] {
		return "low"
	}
	return capabilityRiskForTool("mcp:"+serverName+"/"+toolName, args)
}

// mcpDirectCall is an authorized direct call: the resolved server name and,
// for a high-risk call, the audit record written before it runs.
type mcpDirectCall struct {
	ServerName   string
	Risk         string
	AuditEventID string
	Args         map[string]any // confined arguments the pool call receives (MCPL)
}

type mcpDirectCallKey struct{}

// authorizeMCPToolCall runs before any argument normalization or pool call.
// It writes the response itself and returns false on every refusal.
func (s *AdminServer) authorizeMCPToolCall(w http.ResponseWriter, r *http.Request, serverID uuid.UUID, toolName string, args map[string]any) (*http.Request, mcpDirectCall, bool) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return r, mcpDirectCall{}, false
	}
	serverName, ok := s.resolveDirectMCPTool(w, r, serverID, toolName)
	if !ok {
		return r, mcpDirectCall{}, false
	}
	call := mcpDirectCall{ServerName: serverName, Risk: directMCPToolRisk(serverName, toolName, args)}
	if call.Risk == "low" {
		if !hasScope(identity, mcpDirectReadScope) {
			respondBlocker(w, r, http.StatusForbidden, codeMCPCallForbidden, "Missing required scope: "+mcpDirectReadScope, nil)
			return r, mcpDirectCall{}, false
		}
		if call.Args, ok = s.confineMCPCallArguments(w, r, serverID, call, toolName, args); !ok {
			return r, mcpDirectCall{}, false
		}
		return withMCPDirectCall(r, call), call, true
	}
	if !isApprover(identity) {
		s.auditMCPDirectCall(r, "mcp_tool_call_refused", serverID, call, toolName, args, nil)
		var extra map[string]string
		if identity.Role == "admin" {
			// UX1: only an admin viewer is told which permission is missing.
			extra = map[string]string{"required_scope": scopeApprovalsDecide}
		}
		respondBlockerText(w, r, http.StatusForbidden, codeAdminRequired, mcpDirectCallNeedsApprover, "Missing required scope: "+scopeApprovalsDecide, extra)
		return r, mcpDirectCall{}, false
	}
	if call.Args, ok = s.confineMCPCallArguments(w, r, serverID, call, toolName, args); !ok {
		return r, mcpDirectCall{}, false
	}
	auditID, err := s.auditMCPDirectCall(r, "mcp_tool_called", serverID, call, toolName, args, nil)
	if err != nil || strings.TrimSpace(auditID) == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "direct MCP call audit event could not be recorded", nil)
		return r, mcpDirectCall{}, false
	}
	call.AuditEventID = auditID
	return withMCPDirectCall(r, call), call, true
}

// resolveDirectMCPTool resolves before authorizing (D2): the server must be
// registered and connected, and toolName must equal a discovered tool name
// exactly (case-sensitive, untrimmed).
func (s *AdminServer) resolveDirectMCPTool(w http.ResponseWriter, r *http.Request, serverID uuid.UUID, toolName string) (string, bool) {
	notFound := func() (string, bool) {
		respondBlocker(w, r, http.StatusNotFound, codeMCPToolNotFound, "", nil)
		return "", false
	}
	unavailable := func(detail string) (string, bool) {
		respondBlockerText(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, mcpServerOfflineCopy, detail, nil)
		return "", false
	}
	if serverID == swarm.InternalServerID {
		return notFound()
	}
	if s.MCP == nil || s.MCP.DB == nil || s.MCPPool == nil {
		return unavailable("MCP registry or connection pool is not initialized")
	}
	ctx := r.Context()
	server, err := s.MCP.Get(ctx, serverID)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound()
	}
	if err != nil || server == nil {
		log.Printf("MCPS: resolve MCP server %s: %v", serverID, err)
		return unavailable("MCP server registry could not be read")
	}
	if server.Status != mcpServerStatusConnected {
		return unavailable("MCP server status is " + server.Status)
	}
	tools, err := s.MCP.ListTools(ctx, serverID)
	if err != nil {
		log.Printf("MCPS: list tools for MCP server %s: %v", serverID, err)
		return unavailable("MCP tool cache could not be read")
	}
	for _, tool := range tools {
		if tool.Name == toolName {
			return server.Name, true
		}
	}
	return notFound()
}

// auditMCPDirectCall writes the D6 audit record: argument keys only, never
// values; extra adds fixed, value-free fields. The caller decides whether a
// failure blocks (called) or not (refused and failed, best-effort).
func (s *AdminServer) auditMCPDirectCall(r *http.Request, action string, serverID uuid.UUID, call mcpDirectCall, toolName string, args map[string]any, extra map[string]any) (string, error) {
	record := map[string]any{
		"actor": "operator", "user": auditUserLabelFromRequest(r), "action": action,
		"server_id": serverID.String(), "server_name": call.ServerName, "tool": toolName, "risk": call.Risk,
		"tier": approverTierApprover, "authority": scopeApprovalsDecide,
		"self_approved": action != "mcp_tool_call_refused", "argument_keys": sortedMCPArgumentKeys(args),
	}
	for key, value := range extra {
		record[key] = value
	}
	message := "Direct MCP tool call " + call.ServerName + "/" + toolName
	switch action {
	case "mcp_tool_call_refused":
		message += " refused"
	case "mcp_tool_call_failed":
		message += " failed"
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, mcpDirectCallAuditSource, message, attachActorIdentity(record, r))
	if err != nil && action != "mcp_tool_called" {
		log.Printf("MCPS: refusal audit not recorded (best-effort): %v", err)
	}
	return auditID, err
}

func sortedMCPArgumentKeys(args map[string]any) []string {
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func withMCPDirectCall(r *http.Request, call mcpDirectCall) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), mcpDirectCallKey{}, call))
}

// attributeMCPToolCallExchange adds the human requester and the audit record
// to the retained Exchange item (D6/D7). The item keeps the MCP publisher
// role: operators are not writers of the MCP output channels, so the human
// is recorded in the retained result instead of as the publishing actor.
func attributeMCPToolCallExchange(r *http.Request, input exchange.MCPNormalizationInput) exchange.MCPNormalizationInput {
	if input.Result == nil {
		input.Result = map[string]any{}
	}
	if identity := IdentityFromContext(r.Context()); identity != nil {
		input.Result["requested_by"] = map[string]any{
			"user_id": auditActorIDFromRequest(r), "role": identity.Role, "auth_source": identity.AuthSource,
		}
	}
	if call, ok := r.Context().Value(mcpDirectCallKey{}).(mcpDirectCall); ok {
		input.Result["risk"] = call.Risk
		if call.AuditEventID != "" {
			input.Result["audit_event_id"] = call.AuditEventID
		}
	}
	return input
}

// MCPL: Core-side filesystem path confinement. The upstream filesystem MCP
// server also confines paths; Core does not rely on it.

// errMCPWorkspaceRoot marks a workspace root Core cannot resolve at all.
var errMCPWorkspaceRoot = errors.New("the workspace root could not be resolved")

// mcpPathError names the refused argument key, never its value.
type mcpPathError struct{ Key, Reason string }

func (e *mcpPathError) Error() string { return "argument " + strconv.Quote(e.Key) + " " + e.Reason }

// mcpWorkspaceRoot is the configured root, absolute (abs) and with the
// symlinks of its existing prefix resolved (real).
type mcpWorkspaceRoot struct{ abs, real string }

func resolveMCPWorkspaceRoot() (mcpWorkspaceRoot, error) {
	abs, err := filepath.Abs(strings.TrimSpace(mcp.ResolveFilesystemWorkspaceRoot()))
	if err != nil {
		return mcpWorkspaceRoot{}, errMCPWorkspaceRoot
	}
	real, err := resolveExistingMCPPath(abs)
	if err != nil {
		return mcpWorkspaceRoot{}, errMCPWorkspaceRoot
	}
	return mcpWorkspaceRoot{abs: abs, real: real}, nil
}

// confine returns the absolute path the upstream server receives, after
// checking it lexically against the root and, following symlinks, against
// the root's real path. Non-strings and dangling symlinks are refused.
func (ws mcpWorkspaceRoot) confine(key string, raw any) (string, error) {
	text, ok := raw.(string)
	if !ok {
		return "", &mcpPathError{Key: key, Reason: "is not a path string"}
	}
	candidate := ws.lexical(text)
	if !mcpPathWithin(ws.abs, candidate) {
		return "", &mcpPathError{Key: key, Reason: "resolves outside the workspace root"}
	}
	real, err := resolveExistingMCPPath(candidate)
	if err != nil {
		return "", &mcpPathError{Key: key, Reason: "could not be resolved inside the workspace root"}
	}
	if !mcpPathWithin(ws.real, real) {
		return "", &mcpPathError{Key: key, Reason: "follows a symlink outside the workspace root"}
	}
	return candidate, nil
}

// resolveExistingMCPPath evaluates symlinks on the longest existing prefix
// of the absolute, cleaned path p and appends the missing remainder. An
// entry that exists but does not resolve (a dangling symlink) is an error,
// because a write through it would land wherever it points.
func resolveExistingMCPPath(p string) (string, error) {
	current, rest := p, []string{}
	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			return filepath.Join(append([]string{real}, rest...)...), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if _, lerr := os.Lstat(current); lerr == nil {
			return "", fmt.Errorf("dangling symlink")
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		rest = append([]string{filepath.Base(current)}, rest...)
		current = parent
	}
}

func mcpPathWithin(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// confineMCPCallArguments refuses a call whose filesystem path arguments
// leave the workspace: 403 mcp_path_outside_workspace (503 when the root
// itself cannot be resolved), zero pool calls, and for a high-risk call a
// best-effort keys-only refusal audit written instead of the "called" one.
func (s *AdminServer) confineMCPCallArguments(w http.ResponseWriter, r *http.Request, serverID uuid.UUID, call mcpDirectCall, toolName string, args map[string]any) (map[string]any, bool) {
	confined, err := normalizeMCPToolCallArgumentsForServer(call.ServerName, args)
	if err == nil {
		return confined, true
	}
	if errors.Is(err, errMCPWorkspaceRoot) {
		respondBlockerText(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, mcpServerOfflineCopy, "the MYCELIS_WORKSPACE root could not be resolved", nil)
		return nil, false
	}
	if call.Risk != "low" {
		s.auditMCPDirectCall(r, "mcp_tool_call_refused", serverID, call, toolName, args, map[string]any{"refusal_reason": codeMCPPathOutsideWorkspace})
	}
	respondBlocker(w, r, http.StatusForbidden, codeMCPPathOutsideWorkspace, err.Error(), nil)
	return nil, false
}

// recordMCPToolCallFailure answers a pool call that failed after
// authorization (for example a tool the cache lists but the live server no
// longer offers). It records the outcome: a best-effort
// mcp_tool_call_failed audit linked to the "called" audit, and a failed
// Exchange item. The 502 text is redacted and capped (MCPA).
func (s *AdminServer) recordMCPToolCallFailure(w http.ResponseWriter, r *http.Request, serverID uuid.UUID, call mcpDirectCall, toolName string, args map[string]any, callErr error) {
	message := "tool call failed: " + redactedMCPErrorText(callErr)
	data := map[string]any{}
	failureAuditID := ""
	if call.AuditEventID != "" {
		data["audit_event_id"] = call.AuditEventID
		id, err := s.auditMCPDirectCall(r, "mcp_tool_call_failed", serverID, call, toolName, args,
			map[string]any{"audit_event_id": call.AuditEventID, "failed_stage": "call"})
		if err == nil && id != "" {
			failureAuditID = id
			data["failure_audit_event_id"] = id
		}
	}
	if s.Exchange != nil {
		input := attributeMCPToolCallExchange(r, mcpToolCallExchangeInput(serverID, call.ServerName, toolName, message, true, args, map[string]any{"is_error": true, "error": message}))
		if failureAuditID != "" {
			input.Result["failure_audit_event_id"] = failureAuditID
		}
		item, err := s.Exchange.PublishMCPResult(r.Context(), input)
		if err != nil {
			log.Printf("MCPL: failed MCP call %s/%s not retained: %s", call.ServerName, toolName, redactedMCPErrorText(err))
		}
		if item != nil {
			data["exchange_item_id"] = item.ID.String()
		}
	}
	response := protocol.APIResponse{OK: false, Error: message}
	if len(data) > 0 {
		response.Data = data
	}
	respondAPIJSON(w, http.StatusBadGateway, response)
}
