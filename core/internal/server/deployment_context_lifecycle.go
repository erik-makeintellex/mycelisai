package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/pkg/protocol"
)

// Governed memory lifecycle (M2): operators archive, restore, or permanently
// delete what they saved. Operator only; agents get no tool for this.
//
//	POST   /api/v1/memory/deployment-context/{id}/archive
//	POST   /api/v1/memory/deployment-context/{id}/restore
//	DELETE /api/v1/memory/deployment-context/{id}
//	PATCH  /api/v1/memory/deployment-context/{id}   (edit in place, MEM)
//
// Authority: the creator manages their own private or team entries; org-wide
// entries (company_knowledge, soma_operating_context, or visibility global)
// need a root admin with memory:write. The audit record is written first; if
// it cannot be written the change is refused with 503 and nothing changes.
const (
	codeMemoryEntryNotOwned = "memory_entry_not_owned"
	codeMemoryEntryNotFound = "memory_entry_not_found"
	codeMemoryChangeFailed  = "memory_change_failed"
	codeMemoryEntryArchived = "memory_entry_archived"
	codeMemoryEntryChanged  = "memory_entry_changed"
	scopeMemoryWrite        = "memory:write"
)

var memoryLifecycleCopy = map[string]roleBlockerText{
	codeMemoryEntryNotOwned: {User: blockerText{"Only the person who saved this can change it.",
		"Ask them to archive or delete it. Nothing was changed."}},
	codeMemoryEntryNotFound: {User: blockerText{"This saved item no longer exists.",
		"Refresh the list. It may already have been deleted."}},
	codeMemoryChangeFailed: {User: blockerText{"The change could not be saved, so nothing was changed.",
		"Try again in a moment."}},
	codeMemoryEntryArchived: {User: blockerText{"This saved item is archived, so it cannot be edited.",
		"Restore it first, then edit it. Nothing was changed."}},
	codeMemoryEntryChanged: {User: blockerText{"This saved item changed while you were editing it.",
		"Refresh the list and try again. Nothing was changed."}},
}

func (s *AdminServer) registerDeploymentContextLifecycleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/memory/deployment-context/{id}/archive", s.handleMemoryLifecycle("archive"))
	mux.HandleFunc("POST /api/v1/memory/deployment-context/{id}/restore", s.handleMemoryLifecycle("restore"))
	mux.HandleFunc("DELETE /api/v1/memory/deployment-context/{id}", s.handleMemoryLifecycle("delete"))
	mux.HandleFunc("PATCH /api/v1/memory/deployment-context/{id}", s.handleMemoryEdit)
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
	label := memoryOwnerLabel(identity)
	return label != "" && label == record.LoadedBy
}

// memoryOwnerLabel is the saved-by label a legacy row (no owner id) matches.
func memoryOwnerLabel(identity *RequestIdentity) string {
	if label := strings.TrimSpace(identity.Username); label != "" {
		return label
	}
	return strings.TrimSpace(identity.UserID)
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

// authorizeMemoryChange looks up the entry named in the path and applies the
// lifecycle authority rule. It writes the refusal and returns ok=false when
// the caller may not change the entry. Archive, restore, delete and edit share it.
// An entry the caller may not read answers exactly like an unknown id (404,
// same body): no existence oracle. Readable but not manageable stays 403.
func (s *AdminServer) authorizeMemoryChange(w http.ResponseWriter, r *http.Request, identity *RequestIdentity) (*deploymentcontext.Service, *deploymentcontext.EntryRecord, bool) {
	svc := s.deploymentContextService()
	record, err := svc.Lookup(r.Context(), strings.TrimSpace(r.PathValue("id")))
	readable := false
	if err == nil {
		if readable, err = s.canReadMemoryEntry(r, identity, record); err == nil && !readable {
			err = deploymentcontext.ErrEntryNotFound
		}
	}
	switch {
	case errors.Is(err, deploymentcontext.ErrEntryNotFound):
		respondBlockerText(w, r, http.StatusNotFound, codeMemoryEntryNotFound, memoryLifecycleCopy[codeMemoryEntryNotFound], "", nil)
		return nil, nil, false
	case err != nil:
		respondAPIError(w, "The memory store is unavailable. Nothing was changed.", http.StatusServiceUnavailable)
		return nil, nil, false
	}
	if code, scope := memoryLifecycleDenial(identity, record); code == codeAdminRequired {
		respondBlocker(w, r, http.StatusForbidden, code, "Missing required scope: "+scope, map[string]string{"required_scope": scope})
		return nil, nil, false
	} else if code != "" {
		respondBlockerText(w, r, http.StatusForbidden, code, memoryLifecycleCopy[code], "", nil)
		return nil, nil, false
	}
	return svc, record, true
}

func (s *AdminServer) handleMemoryLifecycle(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			respondAPIError(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		svc, record, ok := s.authorizeMemoryChange(w, r, identity)
		if !ok {
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

// memoryEditPatch is the PATCH body. Only these fields exist; knowledge_class
// and visibility are rejected as unknown (moving classes is promotion).
type memoryEditPatch struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	SourceLabel *string `json:"source_label"`
}

func decodeMemoryEdit(r *http.Request) (deploymentcontext.EditRequest, error) {
	var patch memoryEditPatch
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patch); err != nil {
		return deploymentcontext.EditRequest{}, fmt.Errorf("invalid edit body: %v", err)
	}
	if dec.More() {
		return deploymentcontext.EditRequest{}, errors.New("invalid edit body: one JSON object expected")
	}
	req := deploymentcontext.EditRequest{Title: patch.Title, Content: patch.Content, SourceLabel: patch.SourceLabel}
	return req, req.Validate()
}

// handleMemoryEdit edits a saved entry in place: same authority rule as
// archive/delete, audit first and fail closed, archived entries answer 409.
func (s *AdminServer) handleMemoryEdit(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	req, err := decodeMemoryEdit(r)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	svc, record, ok := s.authorizeMemoryChange(w, r, identity)
	if !ok {
		return
	}
	if record.LifecycleState == deploymentcontext.LifecycleArchived {
		respondBlockerText(w, r, http.StatusConflict, codeMemoryEntryArchived, memoryLifecycleCopy[codeMemoryEntryArchived], "", nil)
		return
	}
	req.ArtifactID, req.Actor, req.Authorized = record.ArtifactID, auditUserLabelFromRequest(r), record
	// Audit first. No audit record, no change.
	auditID, err := s.auditMemoryEdit(r, "authorized", record, req, nil)
	if err != nil || auditID == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, fmt.Sprintf("audit unavailable: %v", err), nil)
		return
	}
	result, err := svc.Edit(r.Context(), req)
	if err != nil {
		_, _ = s.auditMemoryEdit(r, "failed", record, req, nil)
		switch {
		case errors.Is(err, deploymentcontext.ErrEntryNotFound):
			respondBlockerText(w, r, http.StatusNotFound, codeMemoryEntryNotFound, memoryLifecycleCopy[codeMemoryEntryNotFound], "", nil)
		case errors.Is(err, deploymentcontext.ErrEntryArchived):
			respondBlockerText(w, r, http.StatusConflict, codeMemoryEntryArchived, memoryLifecycleCopy[codeMemoryEntryArchived], "", nil)
		case errors.Is(err, deploymentcontext.ErrEntryChanged):
			respondBlockerText(w, r, http.StatusConflict, codeMemoryEntryChanged, memoryLifecycleCopy[codeMemoryEntryChanged], "", nil)
		default:
			log.Printf("memory edit %s failed: %v", record.ArtifactID, err)
			respondBlockerText(w, r, http.StatusInternalServerError, codeMemoryChangeFailed, memoryLifecycleCopy[codeMemoryChangeFailed], err.Error(), nil)
		}
		return
	}
	status := "edited"
	if !result.Changed {
		status = "unchanged"
	}
	if _, err := s.auditMemoryEdit(r, status, record, req, result); err != nil {
		log.Printf("memory edit %s: completion audit failed after the authorized record: %v", record.ArtifactID, err)
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(struct {
		*deploymentcontext.EditResult
		AuditEventID string `json:"audit_event_id"`
	}{result, auditID}))
}

// auditMemoryEdit records an edit step: old and new title, the patched
// fields, and for new content only its SHA-256 and length, never the text.
func (s *AdminServer) auditMemoryEdit(r *http.Request, result string, record *deploymentcontext.EntryRecord, req deploymentcontext.EditRequest, done *deploymentcontext.EditResult) (string, error) {
	newTitle, fields := record.Title, []string{}
	if req.Title != nil {
		newTitle, fields = strings.TrimSpace(*req.Title), append(fields, "title")
	}
	ctx := map[string]any{
		"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "deployment_context_edit",
		"result_status": result, "artifact_id": record.ArtifactID, "knowledge_class": record.KnowledgeClass,
		"visibility": record.Visibility, "old_title": record.Title, "new_title": newTitle,
		"old_content_length": record.ContentLength, "old_chunk_count": record.ChunkCount,
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		sum := sha256.Sum256([]byte(content))
		ctx["content_sha256"], ctx["content_length"] = hex.EncodeToString(sum[:]), utf8.RuneCountInString(content)
		fields = append(fields, "content")
	}
	if req.SourceLabel != nil {
		fields = append(fields, "source_label")
	}
	ctx["fields"] = fields
	if done != nil {
		ctx["chunk_count"], ctx["chunks_removed"], ctx["embedding_status"] = done.ChunkCount, done.ChunksRemoved, done.EmbeddingStatus
	}
	message := fmt.Sprintf("Saved memory edit: %s (%s)", record.ArtifactID, result)
	return s.createAuditEvent(protocol.TemplateChatToProposal, "deployment-context-lifecycle", message, attachActorIdentity(ctx, r))
}
