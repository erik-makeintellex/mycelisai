package server

// Context Snapshot API — persists point-in-time conversation state so it can be
// restored when switching between mission profiles.
//
// Endpoints:
//   POST /api/v1/context/snapshot         — create a new snapshot
//   GET  /api/v1/context/snapshots        — list the 20 most recent snapshots (no messages)
//   GET  /api/v1/context/snapshots/{id}   — fetch full snapshot including messages

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// ContextSnapshot is returned by the list endpoint (messages omitted for size).
type ContextSnapshot struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	SourceProfile string    `json:"source_profile,omitempty"`
	TenantID      string    `json:"tenant_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// ContextSnapshotFull includes the full payload (messages, run_state, role_providers).
type ContextSnapshotFull struct {
	ContextSnapshot
	Messages      json.RawMessage `json:"messages"`
	RunState      json.RawMessage `json:"run_state"`
	RoleProviders json.RawMessage `json:"role_providers"`
}

// POST /api/v1/context/snapshot
func (s *AdminServer) HandleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		respondAPIError(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Name          string          `json:"name"`
		Description   string          `json:"description,omitempty"`
		Messages      json.RawMessage `json:"messages"`
		RunState      json.RawMessage `json:"run_state"`
		RoleProviders json.RawMessage `json:"role_providers"`
		SourceProfile string          `json:"source_profile,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "Snapshot " + time.Now().Format("2006-01-02 15:04")
	}
	if len(req.Messages) == 0 {
		req.Messages = json.RawMessage("[]")
	}
	if len(req.RunState) == 0 {
		req.RunState = json.RawMessage("{}")
	}
	if len(req.RoleProviders) == 0 {
		req.RoleProviders = json.RawMessage("{}")
	}

	var id string
	var createdAt time.Time
	err := s.DB.QueryRowContext(r.Context(), `
		INSERT INTO context_snapshots
		    (name, description, messages, run_state, role_providers, source_profile, tenant_id)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), 'default')
		RETURNING id, created_at`,
		req.Name, nullStrPtr(req.Description),
		[]byte(req.Messages), []byte(req.RunState), []byte(req.RoleProviders),
		req.SourceProfile,
	).Scan(&id, &createdAt)
	if err != nil {
		log.Printf("HandleCreateSnapshot insert error: %v", err)
		respondAPIError(w, "Failed to save snapshot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(ContextSnapshot{
		ID:            id,
		Name:          req.Name,
		Description:   req.Description,
		SourceProfile: req.SourceProfile,
		TenantID:      "default",
		CreatedAt:     createdAt,
	}))
}

// GET /api/v1/context/snapshots — list 20 most recent, without messages payload.
func (s *AdminServer) HandleListSnapshots(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess([]ContextSnapshot{}))
		return
	}

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, name, COALESCE(description,''), COALESCE(source_profile,''), tenant_id, created_at
		FROM context_snapshots
		WHERE tenant_id = 'default'
		ORDER BY created_at DESC
		LIMIT 20`)
	if err != nil {
		log.Printf("HandleListSnapshots query error: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	snapshots := make([]ContextSnapshot, 0)
	for rows.Next() {
		var snap ContextSnapshot
		if err := rows.Scan(&snap.ID, &snap.Name, &snap.Description, &snap.SourceProfile, &snap.TenantID, &snap.CreatedAt); err != nil {
			continue
		}
		snapshots = append(snapshots, snap)
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(snapshots))
}

// GET /api/v1/context/snapshots/{id} — full snapshot including messages.
func (s *AdminServer) HandleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing snapshot ID", http.StatusBadRequest)
		return
	}
	if s.DB == nil {
		respondAPIError(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	var snap ContextSnapshotFull
	var desc, srcProfile sql.NullString
	err := s.DB.QueryRowContext(r.Context(), `
		SELECT id, name, description, messages, run_state, role_providers,
		       source_profile, tenant_id, created_at
		FROM context_snapshots
		WHERE id = $1 AND tenant_id = 'default'`, id).
		Scan(&snap.ID, &snap.Name, &desc,
			&snap.Messages, &snap.RunState, &snap.RoleProviders,
			&srcProfile, &snap.TenantID, &snap.CreatedAt)
	if err == sql.ErrNoRows {
		respondAPIError(w, "Snapshot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("HandleGetSnapshot query error: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	snap.Description = desc.String
	snap.SourceProfile = srcProfile.String

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(snap))
}

// nullStrPtr converts an empty string to nil for nullable DB columns.
func nullStrPtr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
