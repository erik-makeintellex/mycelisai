package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/exchange"
)

// handleListArtifacts returns artifacts filtered by query params.
// GET /api/v1/artifacts?mission_id=&team_id=&agent_id=&limit=
func (s *AdminServer) handleListArtifacts(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var result []artifacts.Artifact
	var err error

	missionIDStr := r.URL.Query().Get("mission_id")
	teamIDStr := r.URL.Query().Get("team_id")
	agentIDStr := r.URL.Query().Get("agent_id")

	switch {
	case missionIDStr != "":
		id, parseErr := uuid.Parse(missionIDStr)
		if parseErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid mission_id: %s"}`, missionIDStr), http.StatusBadRequest)
			return
		}
		result, err = s.Artifacts.ListByMission(ctx, id, limit)
	case teamIDStr != "":
		id, parseErr := uuid.Parse(teamIDStr)
		if parseErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid team_id: %s"}`, teamIDStr), http.StatusBadRequest)
			return
		}
		result, err = s.Artifacts.ListByTeam(ctx, id, limit)
	case agentIDStr != "":
		result, err = s.Artifacts.ListByAgent(ctx, agentIDStr, limit)
	default:
		result, err = s.Artifacts.ListRecent(ctx, limit)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"list failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []artifacts.Artifact{}
	}

	respondJSON(w, result)
}

// handleSaveArtifactToFolder persists a cached image artifact into the workspace.
// POST /api/v1/artifacts/{id}/save
func (s *AdminServer) handleSaveArtifactToFolder(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	var body struct {
		Folder   string `json:"folder,omitempty"`
		Filename string `json:"filename,omitempty"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	workspaceRoot := os.Getenv("MYCELIS_WORKSPACE")
	if workspaceRoot == "" {
		workspaceRoot = "./workspace"
	}
	path, err := s.Artifacts.SaveImageToWorkspace(r.Context(), id, workspaceRoot, body.Folder, body.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]string{
		"id":        id.String(),
		"file_path": path,
	})
}

// handleGetArtifact returns a single artifact by ID.
// GET /api/v1/artifacts/{id}
func (s *AdminServer) handleGetArtifact(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	artifact, err := s.Artifacts.Get(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	respondJSON(w, artifact)
}

// handleDownloadArtifact returns binary or file-backed artifact content as a download.
// GET /api/v1/artifacts/{id}/download
func (s *AdminServer) handleDownloadArtifact(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	artifact, err := s.Artifacts.Get(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	filename := downloadFilename(artifact)
	if resolved, ok := resolveArtifactFilePath(artifact, s.Artifacts.DataDir); ok {
		if artifact.ContentType != "" {
			w.Header().Set("Content-Type", artifact.ContentType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		http.ServeFile(w, r, resolved)
		return
	}

	body, ok, err := artifactDownloadBody(artifact)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, `{"error":"artifact has no downloadable content"}`, http.StatusNotFound)
		return
	}

	if artifact.ContentType != "" {
		w.Header().Set("Content-Type", artifact.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// handleStoreArtifact persists a new artifact.
// POST /api/v1/artifacts
func (s *AdminServer) handleStoreArtifact(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var input artifacts.Artifact
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if input.AgentID == "" {
		http.Error(w, `{"error":"agent_id is required"}`, http.StatusBadRequest)
		return
	}
	if input.ArtifactType == "" {
		http.Error(w, `{"error":"artifact_type is required"}`, http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}
	if input.Status == "" {
		input.Status = "pending"
	}

	stored, err := s.Artifacts.Store(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"store failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if s.Exchange != nil {
		_, _ = s.Exchange.PublishArtifact(r.Context(), exchange.ArtifactNormalizationInput{
			ArtifactID:   stored.ID,
			ArtifactType: string(stored.ArtifactType),
			Title:        stored.Title,
			AgentID:      stored.AgentID,
			Status:       stored.Status,
			TargetRole:   "soma",
			Tags:         []string{"artifact", string(stored.ArtifactType)},
		})
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, stored)
}

// handleUpdateArtifactStatus updates the governance status of an artifact.
// PUT /api/v1/artifacts/{id}/status
func (s *AdminServer) handleUpdateArtifactStatus(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if body.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}

	if err := s.Artifacts.UpdateStatus(r.Context(), id, body.Status); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": body.Status})
}

func artifactDownloadBody(artifact *artifacts.Artifact) ([]byte, bool, error) {
	if artifact == nil || strings.TrimSpace(artifact.Content) == "" {
		return nil, false, nil
	}

	content := strings.TrimSpace(artifact.Content)
	if artifact.ArtifactType == artifacts.TypeImage || strings.HasPrefix(strings.ToLower(strings.TrimSpace(artifact.ContentType)), "image/") {
		raw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, false, fmt.Errorf("decode artifact image content: %w", err)
		}
		return raw, true, nil
	}

	return []byte(content), true, nil
}

func resolveArtifactFilePath(artifact *artifacts.Artifact, dataDir string) (string, bool) {
	if artifact == nil || strings.TrimSpace(artifact.FilePath) == "" {
		return "", false
	}

	candidates := make([]string, 0, 2)
	rawPath := strings.TrimSpace(artifact.FilePath)
	if filepath.IsAbs(rawPath) {
		candidates = append(candidates, filepath.Clean(rawPath))
	} else {
		workspaceRoot := os.Getenv("MYCELIS_WORKSPACE")
		if workspaceRoot == "" {
			workspaceRoot = "./workspace"
		}
		candidates = append(candidates, filepath.Clean(filepath.Join(workspaceRoot, filepath.FromSlash(rawPath))))
		if strings.TrimSpace(dataDir) != "" {
			candidates = append(candidates, filepath.Clean(filepath.Join(dataDir, filepath.FromSlash(rawPath))))
		}
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func downloadFilename(artifact *artifacts.Artifact) string {
	if artifact == nil {
		return "artifact"
	}
	if strings.TrimSpace(artifact.FilePath) != "" {
		name := filepath.Base(filepath.FromSlash(strings.TrimSpace(artifact.FilePath)))
		if name != "." && name != string(filepath.Separator) && name != "" {
			return name
		}
	}
	title := strings.TrimSpace(artifact.Title)
	if title == "" {
		return "artifact"
	}
	return title
}
