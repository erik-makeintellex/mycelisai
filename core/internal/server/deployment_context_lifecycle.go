package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/pkg/protocol"
)

// Governed memory lifecycle (M2): operators archive, restore, or permanently
// delete what they saved. Operator only; agents get no tool for this.
//
//	POST   /api/v1/memory/deployment-context/{id}/archive
//	POST   /api/v1/memory/deployment-context/{id}/restore
//	DELETE /api/v1/memory/deployment-context/{id}
//
// Authority: the creator manages their own private or team entries; org-wide
// entries (company_knowledge, soma_operating_context, or visibility global)
// need a root admin with memory:write. The audit record is written first; if
// it cannot be written the change is refused with 503 and nothing changes.
const (
	codeMemoryEntryNotOwned = "memory_entry_not_owned"
	codeMemoryEntryNotFound = "memory_entry_not_found"
	codeMemoryChangeFailed  = "memory_change_failed"
	scopeMemoryWrite        = "memory:write"
)

var memoryLifecycleCopy = map[string]roleBlockerText{
	codeMemoryEntryNotOwned: {User: blockerText{"Only the person who saved this can change it.",
		"Ask them to archive or delete it. Nothing was changed."}},
	codeMemoryEntryNotFound: {User: blockerText{"This saved item no longer exists.",
		"Refresh the list. It may already have been deleted."}},
	codeMemoryChangeFailed: {User: blockerText{"The change could not be saved, so nothing was changed.",
		"Try again in a moment."}},
}

func (s *AdminServer) registerDeploymentContextLifecycleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/memory/deployment-context/{id}/archive", s.handleMemoryLifecycle("archive"))
	mux.HandleFunc("POST /api/v1/memory/deployment-context/{id}/restore", s.handleMemoryLifecycle("restore"))
	mux.HandleFunc("DELETE /api/v1/memory/deployment-context/{id}", s.handleMemoryLifecycle("delete"))
}

// isOrgWideMemory reports whether an entry needs admin authority to change.
func isOrgWideMemory(record *deploymentcontext.EntryRecord) bool {
	switch record.KnowledgeClass {
	case deploymentcontext.KnowledgeClassCompanyKnowledge, deploymentcontext.KnowledgeClassSomaOperating:
		return true
	}
	return record.Visibility != "private" && record.Visibility != "team"
}

// isMemoryOwner matches the saving user by user id; rows saved before owner
// ids were recorded fall back to the saved user label.
func isMemoryOwner(identity *RequestIdentity, record *deploymentcontext.EntryRecord) bool {
	if identity == nil {
		return false
	}
	if record.OwnerUserID != "" {
		return strings.TrimSpace(identity.UserID) != "" && identity.UserID == record.OwnerUserID
	}
	label := strings.TrimSpace(identity.Username)
	if label == "" {
		label = strings.TrimSpace(identity.UserID)
	}
	return label != "" && label == record.LoadedBy
}

// memoryLifecycleDenial returns "" when identity may change the entry, or the
// blocker code (and the missing scope for admin_required).
func memoryLifecycleDenial(identity *RequestIdentity, record *deploymentcontext.EntryRecord) (string, string) {
	if isOrgWideMemory(record) {
		if identity != nil && identity.Role == "admin" && hasScope(identity, scopeMemoryWrite) {
			return "", ""
		}
		return codeAdminRequired, scopeMemoryWrite
	}
	if isMemoryOwner(identity, record) {
		return "", ""
	}
	return codeMemoryEntryNotOwned, ""
}

// canManageMemoryEntry tells the list which entries the viewer may change.
func canManageMemoryEntry(identity *RequestIdentity, entry *deploymentcontext.Entry) bool {
	code, _ := memoryLifecycleDenial(identity, &deploymentcontext.EntryRecord{KnowledgeClass: entry.KnowledgeClass,
		Visibility: entry.Visibility, LoadedBy: entry.LoadedBy, OwnerUserID: entry.OwnerUserID, TeamID: entry.TeamID})
	return code == ""
}

func (s *AdminServer) handleMemoryLifecycle(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			respondAPIError(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		svc := s.deploymentContextService()
		record, err := svc.Lookup(r.Context(), strings.TrimSpace(r.PathValue("id")))
		switch {
		case errors.Is(err, deploymentcontext.ErrEntryNotFound):
			respondBlockerText(w, r, http.StatusNotFound, codeMemoryEntryNotFound, memoryLifecycleCopy[codeMemoryEntryNotFound], "", nil)
			return
		case err != nil:
			respondAPIError(w, "The memory store is unavailable. Nothing was changed.", http.StatusServiceUnavailable)
			return
		}
		if code, scope := memoryLifecycleDenial(identity, record); code == codeAdminRequired {
			respondBlocker(w, r, http.StatusForbidden, code, "Missing required scope: "+scope, map[string]string{"required_scope": scope})
			return
		} else if code != "" {
			respondBlockerText(w, r, http.StatusForbidden, code, memoryLifecycleCopy[code], "", nil)
			return
		}
		target := map[string]string{"archive": deploymentcontext.LifecycleArchived, "restore": deploymentcontext.LifecycleActive}[action]
		if action != "delete" && record.LifecycleState == target {
			respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"artifact_id": record.ArtifactID, "lifecycle_state": target, "changed": false}))
			return
		}
		// Audit first. No audit record, no change.
		auditID, err := s.auditMemoryLifecycle(r, action, "authorized", record, nil)
		if err != nil || auditID == "" {
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, fmt.Sprintf("audit unavailable: %v", err), nil)
			return
		}
		data := map[string]any{"artifact_id": record.ArtifactID, "changed": true, "audit_event_id": auditID}
		actor := auditUserLabelFromRequest(r)
		switch action {
		case "archive":
			err = svc.Archive(r.Context(), record.ArtifactID, actor)
		case "restore":
			err = svc.Restore(r.Context(), record.ArtifactID, actor)
		default:
			var result deploymentcontext.DeleteResult
			result, err = svc.Delete(r.Context(), record.ArtifactID)
			data["chunks_removed"] = result.ChunksRemoved
			target = "deleted"
		}
		if err != nil {
			log.Printf("memory %s %s failed: %v", action, record.ArtifactID, err)
			_, _ = s.auditMemoryLifecycle(r, action, "failed", record, nil)
			respondBlockerText(w, r, http.StatusInternalServerError, codeMemoryChangeFailed, memoryLifecycleCopy[codeMemoryChangeFailed], err.Error(), nil)
			return
		}
		if _, err := s.auditMemoryLifecycle(r, action, target, record, data); err != nil {
			log.Printf("memory %s %s: completion audit failed after the authorized record: %v", action, record.ArtifactID, err)
		}
		data["lifecycle_state"] = target
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(data))
	}
}

// auditMemoryLifecycle records a lifecycle step. Delete records never carry
// the title or content: they are the tombstone.
func (s *AdminServer) auditMemoryLifecycle(r *http.Request, action, result string, record *deploymentcontext.EntryRecord, extra map[string]any) (string, error) {
	ctx := map[string]any{
		"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "deployment_context_" + action,
		"result_status": result, "artifact_id": record.ArtifactID, "knowledge_class": record.KnowledgeClass,
		"visibility": record.Visibility, "chunk_count": record.ChunkCount, "content_length": record.ContentLength,
	}
	message := fmt.Sprintf("Saved memory %s: %s (%s)", action, record.ArtifactID, result)
	if action != "delete" {
		ctx["title"] = record.Title
	}
	if removed, ok := extra["chunks_removed"]; ok {
		ctx["chunks_removed"] = removed
	}
	return s.createAuditEvent(protocol.TemplateChatToProposal, "deployment-context-lifecycle", message, attachActorIdentity(ctx, r))
}
