package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) deploymentContextService() *deploymentcontext.Service {
	return deploymentcontext.NewService(s.Artifacts, s.Mem, s.Cognitive)
}

// HandleDeploymentContext manages governed deployment knowledge that becomes
// durable context for Soma and downstream teams: keyword-searchable on save,
// semantically searchable once an embedding engine embeds it.
// This store is intentionally separate from Soma's ordinary remembered facts.
// GET  /api/v1/memory/deployment-context
// POST /api/v1/memory/deployment-context
func (s *AdminServer) HandleDeploymentContext(w http.ResponseWriter, r *http.Request) {
	svc := s.deploymentContextService()

	switch r.Method {
	case http.MethodGet:
		if svc == nil || svc.Artifacts == nil || svc.Artifacts.DB == nil {
			respondError(w, "deployment context store unavailable", http.StatusServiceUnavailable)
			return
		}

		limit := 20
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		entries, err := svc.List(r.Context(), limit)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		semantic := "unavailable"
		if svc.Cognitive.EmbeddingAvailable(r.Context()) {
			semantic = "available"
		}
		respondJSON(w, map[string]any{
			"entries":         entries,
			"count":           len(entries),
			"semantic_search": semantic,
		})
	case http.MethodPost:
		if err := svc.Ready(); err != nil {
			respondError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		var req struct {
			KnowledgeClass    string   `json:"knowledge_class,omitempty"`
			Title             string   `json:"title"`
			Content           string   `json:"content"`
			ContentType       string   `json:"content_type,omitempty"`
			SourceLabel       string   `json:"source_label,omitempty"`
			SourceKind        string   `json:"source_kind,omitempty"`
			Visibility        string   `json:"visibility,omitempty"`
			SensitivityClass  string   `json:"sensitivity_class,omitempty"`
			TrustClass        string   `json:"trust_class,omitempty"`
			Tags              []string `json:"tags,omitempty"`
			AgentID           string   `json:"agent_id,omitempty"`
			TeamID            string   `json:"team_id,omitempty"`
			SomaContextKind   string   `json:"soma_context_kind,omitempty"`
			OutputSpecificity string   `json:"output_specificity,omitempty"`
			ContentDomain     string   `json:"content_domain,omitempty"`
			TargetGoalSets    []string `json:"target_goal_sets,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		result, err := svc.Ingest(r.Context(), deploymentcontext.IngestRequest{
			KnowledgeClass:    req.KnowledgeClass,
			Title:             req.Title,
			Content:           req.Content,
			ContentType:       req.ContentType,
			SourceLabel:       req.SourceLabel,
			SourceKind:        req.SourceKind,
			Visibility:        req.Visibility,
			SensitivityClass:  req.SensitivityClass,
			TrustClass:        req.TrustClass,
			Tags:              req.Tags,
			AgentID:           req.AgentID,
			TeamID:            req.TeamID,
			UserLabel:         auditUserLabelFromRequest(r),
			SomaContextKind:   req.SomaContextKind,
			OutputSpecificity: req.OutputSpecificity,
			ContentDomain:     req.ContentDomain,
			TargetGoalSets:    req.TargetGoalSets,
		})
		var saveErr *deploymentcontext.SaveError
		if errors.As(err, &saveErr) {
			log.Printf("deployment context save failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "Saving failed, so nothing was stored. Try again.", "code": saveErr.Code})
			return
		}
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		respondJSON(w, result)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleDeploymentContextBackfill embeds pending governed chunks when an
// embedding engine is available. Root admin with memory:write only; audited.
// POST /api/v1/memory/deployment-context/backfill
func (s *AdminServer) HandleDeploymentContextBackfill(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "memory:write"); !ok {
		return
	}
	svc := s.deploymentContextService()
	if err := svc.Ready(); err != nil {
		respondAPIError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	result, err := svc.BackfillEmbeddings(r.Context(), limit)
	if err != nil {
		log.Printf("deployment context backfill failed: %v", err)
		respondAPIError(w, "Backfill failed; pending entries stay keyword-searchable.", http.StatusInternalServerError)
		return
	}
	auditID, _ := s.createAuditEvent(protocol.TemplateChatToProposal, "deployment-context-backfill",
		fmt.Sprintf("Deployment context backfill: %d embedded, %d remaining (%s)", result.Embedded, result.Remaining, result.Status),
		attachActorIdentity(map[string]any{
			"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "deployment_context_backfill",
			"result_status": result.Status, "embedded": result.Embedded, "remaining": result.Remaining,
		}, r))
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"embedded": result.Embedded, "remaining": result.Remaining, "status": result.Status, "audit_event_id": auditID,
	}))
}
