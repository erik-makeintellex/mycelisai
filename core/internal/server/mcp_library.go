package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/mcp"
)

type mcpLibraryRequest struct {
	Name              string               `json:"name"`
	Env               map[string]string    `json:"env,omitempty"`
	GovernanceContext mcpGovernanceContext `json:"governance_context,omitempty"`
}

type mcpPreparedLibraryRequest struct {
	Request    mcpLibraryRequest
	Entry      *mcp.LibraryEntry
	Inspection map[string]any
}

type mcpLibraryInstallResult struct {
	Server       mcp.ServerConfig
	Tools        []mcp.ToolDef
	SelfApproved bool
	AuditEventID string
}

// handleMCPLibrary returns the curated MCP server library organized by category.
// GET /api/v1/mcp/library
func (s *AdminServer) handleMCPLibrary(w http.ResponseWriter, r *http.Request) {
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}
	respondJSON(w, s.MCPLibrary.Categories)
}

// handleMCPLibraryInspect previews policy posture for a curated MCP library install.
// POST /api/v1/mcp/library/inspect
func (s *AdminServer) handleMCPLibraryInspect(w http.ResponseWriter, r *http.Request) {
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}

	prepared, ok := s.prepareMCPLibraryRequest(w, r)
	if !ok {
		return
	}
	respondJSON(w, prepared.Inspection)
}

// handleMCPLibraryInstall installs an MCP server from the curated library by name.
// POST /api/v1/mcp/library/install
func (s *AdminServer) handleMCPLibraryInstall(w http.ResponseWriter, r *http.Request) {
	prepared, result, ok := s.installFromMCPLibrary(w, r, "install")
	if !ok {
		return
	}
	respondJSON(w, map[string]any{
		"status":         "installed",
		"server":         redactMCPServerConfig(result.Server),
		"tools":          result.Tools,
		"governance":     prepared.Inspection["governance"],
		"self_approved":  result.SelfApproved,
		"audit_event_id": result.AuditEventID,
	})
}

// handleMCPLibraryApply runs the curated MCP inspect+install flow as a single API call.
// POST /api/v1/mcp/library/apply
func (s *AdminServer) handleMCPLibraryApply(w http.ResponseWriter, r *http.Request) {
	prepared, result, ok := s.installFromMCPLibrary(w, r, "apply")
	if !ok {
		return
	}
	respondJSON(w, map[string]any{
		"status":            "installed",
		"requires_approval": false, // nothing is ever left pending (MCPA D5)
		"self_approved":     result.SelfApproved,
		"audit_event_id":    result.AuditEventID,
		"server":            redactMCPServerConfig(result.Server),
		"tools":             result.Tools,
		"inspection":        prepared.Inspection,
		"governance":        prepared.Inspection["governance"],
	})
}

// installFromMCPLibrary is the one install/apply path (MCPA D1, D4, D5, D7):
// authority, subsystem checks, entry, approver for require_approval, declared
// env, fail-closed audit, then install and connect. It writes every refusal.
func (s *AdminServer) installFromMCPLibrary(w http.ResponseWriter, r *http.Request, logAction string) (mcpPreparedLibraryRequest, mcpLibraryInstallResult, bool) {
	identity, ok := requireMCPConfigWrite(w, r)
	if !ok {
		return mcpPreparedLibraryRequest{}, mcpLibraryInstallResult{}, false
	}
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return mcpPreparedLibraryRequest{}, mcpLibraryInstallResult{}, false
	}
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return mcpPreparedLibraryRequest{}, mcpLibraryInstallResult{}, false
	}
	prepared, ok := s.prepareMCPLibraryRequest(w, r)
	if !ok {
		return prepared, mcpLibraryInstallResult{}, false
	}
	selfApproved, ok := requireMCPApproverFor(w, r, identity, prepared.Inspection)
	if !ok || !requireDeclaredMCPEnv(w, r, prepared.Entry, prepared.Request.Env) {
		return prepared, mcpLibraryInstallResult{}, false
	}
	result, ok := s.installMCPLibraryEntry(w, r, prepared, selfApproved, logAction)
	return prepared, result, ok
}

func (s *AdminServer) prepareMCPLibraryRequest(w http.ResponseWriter, r *http.Request) (mcpPreparedLibraryRequest, bool) {
	var req mcpLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return mcpPreparedLibraryRequest{}, false
	}
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return mcpPreparedLibraryRequest{}, false
	}

	entry := s.MCPLibrary.FindByName(req.Name)
	if entry == nil {
		http.Error(w, fmt.Sprintf(`{"error":"server %q not found in library"}`, req.Name), http.StatusNotFound)
		return mcpPreparedLibraryRequest{}, false
	}

	inspectCtx := normalizeMCPGovernanceContext(r, req.GovernanceContext)
	return mcpPreparedLibraryRequest{
		Request:    req,
		Entry:      entry,
		Inspection: buildMCPLibraryInspectionReport(entry, inspectCtx),
	}, true
}

// mcpInstallAuditRecord is the D7 mcp_server_installed record: labels and env
// keys only, never env values.
func mcpInstallAuditRecord(prepared mcpPreparedLibraryRequest, selfApproved, replacedExisting bool, logAction string) map[string]any {
	record := map[string]any{
		"server_name": prepared.Entry.Name, "transport": prepared.Entry.Transport, "route": logAction,
		"deployment_boundary": mcpLibraryDeploymentBoundary(prepared.Entry),
		"credential_boundary": mcpLibraryCredentialBoundary(prepared.Entry),
		"env_keys":            sortedMCPEnvKeys(prepared.Request.Env),
		"replaced_existing":   replacedExisting,
		"decision":            prepared.Inspection["decision"],
		"self_approved":       selfApproved,
		"tier":                approverTierAuto,
	}
	if selfApproved {
		record["tier"], record["authority"] = approverTierApprover, scopeApprovalsDecide
	}
	if ctx, ok := prepared.Inspection["governance_context"].(mcpGovernanceContext); ok {
		record["owner_user_id"], record["actor_role"] = ctx.OwnerUserID, ctx.ActorRole
	}
	return record
}

func (s *AdminServer) installMCPLibraryEntry(w http.ResponseWriter, r *http.Request, prepared mcpPreparedLibraryRequest, selfApproved bool, logAction string) (mcpLibraryInstallResult, bool) {
	ctx := r.Context()
	name := prepared.Entry.Name
	existing, err := s.MCP.FindServerByName(ctx, name)
	if err != nil {
		log.Printf("MCPA: %s %s: replace lookup: %s", logAction, name, redactedMCPErrorText(err))
		respondAPIError(w, mcpInstallFailedMessage, http.StatusInternalServerError)
		return mcpLibraryInstallResult{}, false
	}
	record := mcpInstallAuditRecord(prepared, selfApproved, existing != nil, logAction)
	auditID, ok := s.auditMCPConfigChange(w, r, "mcp_server_installed", name, record)
	if !ok {
		return mcpLibraryInstallResult{}, false
	}
	failed := func(stage string, err error) (mcpLibraryInstallResult, bool) {
		s.recordMCPConfigFailure(r, "mcp_server_install_failed", name, auditID, stage, err, record)
		respondAPIError(w, mcpInstallFailedMessage, http.StatusInternalServerError)
		return mcpLibraryInstallResult{}, false
	}

	runtimeCfg, err := mcp.ApplyRuntimeDefaults(prepared.Entry.ToServerConfig(prepared.Request.Env))
	if err != nil {
		return failed("runtime_defaults", err)
	}
	installed, err := s.MCP.Install(ctx, runtimeCfg)
	if err != nil {
		return failed("install", err)
	}
	if err := s.MCPPool.Connect(ctx, *installed); err != nil {
		// The row is registered with status "error" but nothing launched, so
		// the answer is an honest blocker, never "installed".
		s.recordMCPConfigFailure(r, "mcp_server_install_failed", name, auditID, "connect", err, record)
		respondBlocker(w, r, http.StatusBadGateway, codeMCPConnectFailed, redactedMCPErrorText(err), map[string]string{
			"server_id": installed.ID.String(), "server_name": installed.Name, "audit_event_id": auditID,
		})
		return mcpLibraryInstallResult{}, false
	}

	tools, err := s.MCP.ListTools(ctx, installed.ID)
	if err != nil {
		log.Printf("MCPA: %s %s: list tools: %s", logAction, name, redactedMCPErrorText(err))
		tools = []mcp.ToolDef{}
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}
	return mcpLibraryInstallResult{Server: *installed, Tools: tools, SelfApproved: selfApproved, AuditEventID: auditID}, true
}
