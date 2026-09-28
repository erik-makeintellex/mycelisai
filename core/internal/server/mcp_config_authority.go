package server

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
)

// MCP configuration authority (contract MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md).
// Installing, applying or deleting an MCP server and writing a toolset change
// what Core launches and what agents may call, so every write needs a root
// admin with mcp_config:write (D1/D2), only declared env keys (D4), an
// approver for require_approval entries (D5), and a fail-closed audit record
// before any side effect (D7). Reads stay authentication-only (D3).

// scopeMCPConfigWrite gates MCP configuration writes. It is deliberately not
// "mcp:write", which parses as an agent tool grant (mcp.ParseToolRef).
const scopeMCPConfigWrite = "mcp_config:write"

// mcpConfigAuditSource is the log_entries source of every D7 record.
const mcpConfigAuditSource = "mcp-config"

// mcpEnvEmptyKeyLabel stands in for an empty env key in rejected_env_keys.
const mcpEnvEmptyKeyLabel = "(empty)"

// mcpErrorTextCap mirrors the MCP tool-failure cap (mcp.maxToolErrorBytes).
const mcpErrorTextCap = 2048

// Fixed failure messages (D8): raw errors go only to a redacted log.
const (
	mcpInstallFailedMessage = "MCP server install failed. Check the Core log for the redacted reason."
	mcpDeleteFailedMessage  = "MCP server delete failed. Check the Core log for the redacted reason."
)

// requireMCPConfigWrite admits only a root admin holding mcp_config:write
// (web admins, the API key and break-glass hold "*"). No identity is 401 and
// everyone else is 403 admin_required. Call it before any other check.
func requireMCPConfigWrite(w http.ResponseWriter, r *http.Request) (*RequestIdentity, bool) {
	return requireRootAdminScope(w, r, scopeMCPConfigWrite)
}

// requireMCPApproverFor enforces D5: a require_approval entry installs only
// for an approver, recorded as self-approved tier 2. It returns selfApproved.
func requireMCPApproverFor(w http.ResponseWriter, r *http.Request, identity *RequestIdentity, inspection map[string]any) (bool, bool) {
	if decision, _ := inspection["decision"].(string); decision != "require_approval" {
		return false, true
	}
	if !isApprover(identity) {
		respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "Missing required scope: "+scopeApprovalsDecide,
			map[string]string{"required_scope": scopeApprovalsDecide})
		return false, false
	}
	return true, true
}

// undeclaredMCPEnvKeys returns the sorted request env keys that the entry does
// not declare. Matching is exact and case-sensitive; an empty key is never
// declared.
func undeclaredMCPEnvKeys(entry *mcp.LibraryEntry, env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	declared := map[string]struct{}{}
	if entry != nil {
		for _, key := range entry.DeclaredEnvKeys() {
			declared[key] = struct{}{}
		}
	}
	var rejected []string
	for key := range env {
		if key == "" {
			rejected = append(rejected, mcpEnvEmptyKeyLabel)
			continue
		}
		if _, ok := declared[key]; !ok {
			rejected = append(rejected, key)
		}
	}
	sort.Strings(rejected)
	return rejected
}

// requireDeclaredMCPEnv writes 400 mcp_env_rejected and returns false when
// the request env holds any undeclared key. The response names keys only,
// never values.
func requireDeclaredMCPEnv(w http.ResponseWriter, r *http.Request, entry *mcp.LibraryEntry, env map[string]string) bool {
	rejected := undeclaredMCPEnvKeys(entry, env)
	if len(rejected) == 0 {
		return true
	}
	allowed := []string{}
	if entry != nil {
		allowed = entry.DeclaredEnvKeys()
	}
	sort.Strings(allowed)
	respondBlocker(w, r, http.StatusBadRequest, codeMCPEnvRejected, "", map[string]string{
		"rejected_env_keys": strings.Join(rejected, ", "),
		"allowed_env_keys":  strings.Join(allowed, ", "),
	})
	return false
}

// auditMCPConfigChange writes the D7 record before a configuration write. An
// unavailable audit store (nil DB, insert error or empty id) is 503
// service_unavailable and the caller must stop with zero side effects.
// record must hold keys and labels only, never env values.
func (s *AdminServer) auditMCPConfigChange(w http.ResponseWriter, r *http.Request, action, subject string, record map[string]any) (string, bool) {
	auditID, err := s.writeMCPConfigAudit(r, action, subject, record)
	if err != nil || strings.TrimSpace(auditID) == "" {
		if err != nil {
			log.Printf("MCPA: %s audit not recorded: %s", action, redactedMCPErrorText(err))
		}
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "MCP configuration audit event could not be recorded", nil)
		return "", false
	}
	return auditID, true
}

// recordMCPConfigFailure is the best-effort "..._failed" record for a write
// that failed after its audit; the raw error is logged redacted, never stored.
func (s *AdminServer) recordMCPConfigFailure(r *http.Request, action, subject, auditID, stage string, err error, record map[string]any) {
	log.Printf("MCPA: %s %s failed at %s: %s", action, subject, stage, redactedMCPErrorText(err))
	failed := map[string]any{"audit_event_id": auditID, "failed_stage": stage}
	for key, value := range record {
		failed[key] = value
	}
	if _, auditErr := s.writeMCPConfigAudit(r, action, subject, failed); auditErr != nil {
		log.Printf("MCPA: %s failure record not written (best-effort): %s", action, redactedMCPErrorText(auditErr))
	}
}

func (s *AdminServer) writeMCPConfigAudit(r *http.Request, action, subject string, record map[string]any) (string, error) {
	ctx := map[string]any{"actor": "operator", "user": auditUserLabelFromRequest(r), "action": action, "authority": scopeMCPConfigWrite}
	for key, value := range record {
		ctx[key] = value
	}
	return s.createAuditEvent(protocol.TemplateChatToProposal, mcpConfigAuditSource, "MCP configuration "+action+": "+subject, attachActorIdentity(ctx, r))
}

// redactedMCPErrorText is err through mcp.RedactToolText, capped like MCP
// tool-failure text. It is safe for logs and response bodies.
func redactedMCPErrorText(err error) string {
	if err == nil {
		return ""
	}
	text := mcp.RedactToolText(err.Error())
	if len(text) <= mcpErrorTextCap {
		return text
	}
	cut := mcpErrorTextCap
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return strings.TrimRight(text[:cut], " \n") + " ... [truncated]"
}

// deleteMCPServer is D9: resolve (404 mcp_server_not_found, nothing
// disconnected), audit, then Disconnect and Delete. It resolves before the
// audit so the record names the server (D7).
func (s *AdminServer) deleteMCPServer(w http.ResponseWriter, r *http.Request, serverID uuid.UUID) {
	ctx := r.Context()
	server, err := s.MCP.Get(ctx, serverID)
	if errors.Is(err, sql.ErrNoRows) {
		respondBlocker(w, r, http.StatusNotFound, codeMCPServerNotFound, "", nil)
		return
	}
	if err != nil || server == nil {
		log.Printf("MCPA: resolve MCP server %s for delete: %s", serverID, redactedMCPErrorText(err))
		respondAPIError(w, mcpDeleteFailedMessage, http.StatusInternalServerError)
		return
	}
	record := map[string]any{"server_id": serverID.String(), "server_name": server.Name}
	auditID, ok := s.auditMCPConfigChange(w, r, "mcp_server_deleted", server.Name, record)
	if !ok {
		return
	}
	if err := s.MCPPool.Disconnect(serverID); err != nil {
		log.Printf("MCPA: delete %s: disconnect (best-effort): %s", server.Name, redactedMCPErrorText(err))
	}
	if err := s.MCP.Delete(ctx, serverID); err != nil {
		s.recordMCPConfigFailure(r, "mcp_server_delete_failed", server.Name, auditID, "delete", err, record)
		respondAPIError(w, mcpDeleteFailedMessage, http.StatusInternalServerError)
		return
	}
	respondJSON(w, map[string]string{"status": "deleted", "audit_event_id": auditID})
}
