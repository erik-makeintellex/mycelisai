package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/deploymentcontext"
)

func (s *AdminServer) deploymentContextService() *deploymentcontext.Service {
	return deploymentcontext.NewService(s.Artifacts, s.Mem, s.Cognitive)
}

// HandleDeploymentContext manages governed deployment knowledge that should
// become durable pgvector-backed context for Soma and downstream teams.
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
		respondJSON(w, map[string]any{
			"entries": entries,
			"count":   len(entries),
		})
	case http.MethodPost:
		if err := svc.Ready(); err != nil {
			respondError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		var req struct {
			KnowledgeClass   string   `json:"knowledge_class,omitempty"`
			Title            string   `json:"title"`
			Content          string   `json:"content"`
			ContentType      string   `json:"content_type,omitempty"`
			SourceLabel      string   `json:"source_label,omitempty"`
			SourceKind       string   `json:"source_kind,omitempty"`
			Visibility       string   `json:"visibility,omitempty"`
			SensitivityClass string   `json:"sensitivity_class,omitempty"`
			TrustClass       string   `json:"trust_class,omitempty"`
			Tags             []string `json:"tags,omitempty"`
			AgentID          string   `json:"agent_id,omitempty"`
			TeamID           string   `json:"team_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		result, err := svc.Ingest(r.Context(), deploymentcontext.IngestRequest{
			KnowledgeClass:   req.KnowledgeClass,
			Title:            req.Title,
			Content:          req.Content,
			ContentType:      req.ContentType,
			SourceLabel:      req.SourceLabel,
			SourceKind:       req.SourceKind,
			Visibility:       req.Visibility,
			SensitivityClass: req.SensitivityClass,
			TrustClass:       req.TrustClass,
			Tags:             req.Tags,
			AgentID:          req.AgentID,
			TeamID:           req.TeamID,
			UserLabel:        auditUserLabelFromRequest(r),
		})
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
