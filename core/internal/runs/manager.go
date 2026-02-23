// Package runs provides the V7 mission_run lifecycle manager.
// A mission is a definition; a run is a single execution instance.
// run_id is mandatory on all V7 events — every mission activation creates a run.
package runs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// RunStatus classifies the lifecycle state of a mission run.
type RunStatus = string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusCompleted RunStatus = "completed"
	StatusFailed    RunStatus = "failed"
)

// MissionRun is a single execution instance of a mission.
// One mission definition → many runs (each activation = new run).
type MissionRun struct {
	ID          string                 `json:"id"`
	MissionID   string                 `json:"mission_id"`
	TenantID    string                 `json:"tenant_id"`
	Status      RunStatus              `json:"status"`
	RunDepth    int                    `json:"run_depth"`
	ParentRunID string                 `json:"parent_run_id,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Manager creates and manages mission_run records.
// Implements protocol.RunsManager (interface defined in pkg/protocol/events.go).
type Manager struct {
	db *sql.DB
}

// NewManager creates a Manager backed by the shared DB.
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// CreateRun inserts a new mission_run record with status=running and returns the run ID.
// Implements protocol.RunsManager.CreateRun.
func (m *Manager) CreateRun(ctx context.Context, missionID string) (string, error) {
	if m.db == nil {
		return "", fmt.Errorf("runs: database not available")
	}
	if missionID == "" {
		return "", fmt.Errorf("runs: mission_id is required")
	}

	id := uuid.New().String()
	_, err := m.db.ExecContext(ctx, `
		INSERT INTO mission_runs (id, mission_id, tenant_id, status, run_depth, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, missionID, "default", StatusRunning, 0, time.Now())
	if err != nil {
		return "", fmt.Errorf("runs: insert failed: %w", err)
	}

	log.Printf("[runs] created run %s for mission %s", id, missionID)
	return id, nil
}

// CreateChildRun creates a run that is a child of a trigger chain.
// Used by Team B (Trigger Engine) — not by Team A directly.
func (m *Manager) CreateChildRun(ctx context.Context, missionID, parentRunID string, depth int) (string, error) {
	if m.db == nil {
		return "", fmt.Errorf("runs: database not available")
	}

	id := uuid.New().String()
	_, err := m.db.ExecContext(ctx, `
		INSERT INTO mission_runs (id, mission_id, tenant_id, status, run_depth, parent_run_id, started_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,'')::uuid, $7)
	`, id, missionID, "default", StatusRunning, depth, parentRunID, time.Now())
	if err != nil {
		return "", fmt.Errorf("runs: insert child failed: %w", err)
	}

	log.Printf("[runs] created child run %s (parent=%s, depth=%d)", id, parentRunID, depth)
	return id, nil
}

// UpdateRunStatus sets the run status. For terminal statuses (completed, failed),
// completed_at is also set.
// Implements protocol.RunsManager.UpdateRunStatus.
func (m *Manager) UpdateRunStatus(ctx context.Context, runID string, status string) error {
	if m.db == nil {
		return fmt.Errorf("runs: database not available")
	}

	var err error
	switch status {
	case StatusCompleted, StatusFailed:
		_, err = m.db.ExecContext(ctx, `
			UPDATE mission_runs SET status = $1, completed_at = NOW() WHERE id = $2
		`, status, runID)
	default:
		_, err = m.db.ExecContext(ctx, `
			UPDATE mission_runs SET status = $1 WHERE id = $2
		`, status, runID)
	}
	if err != nil {
		return fmt.Errorf("runs: update status failed: %w", err)
	}
	return nil
}

// GetRun retrieves a run by ID.
func (m *Manager) GetRun(ctx context.Context, runID string) (*MissionRun, error) {
	if m.db == nil {
		return nil, fmt.Errorf("runs: database not available")
	}

	var run MissionRun
	var parentRunID sql.NullString
	var completedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, `
		SELECT id, mission_id, tenant_id, status, run_depth,
		       COALESCE(parent_run_id::text, ''), started_at, completed_at
		FROM mission_runs WHERE id = $1
	`, runID).Scan(
		&run.ID, &run.MissionID, &run.TenantID, &run.Status, &run.RunDepth,
		&parentRunID, &run.StartedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("runs: not found: %s", runID)
	} else if err != nil {
		return nil, fmt.Errorf("runs: query failed: %w", err)
	}

	if parentRunID.Valid {
		run.ParentRunID = parentRunID.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		run.CompletedAt = &t
	}
	return &run, nil
}

// ListRunsForMission returns recent runs for a mission, newest first.
// Used by GET /api/v1/runs/{id}/chain to build the causal chain view.
func (m *Manager) ListRunsForMission(ctx context.Context, missionID string, limit int) ([]MissionRun, error) {
	if m.db == nil {
		return nil, fmt.Errorf("runs: database not available")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := m.db.QueryContext(ctx, `
		SELECT id, mission_id, tenant_id, status, run_depth,
		       COALESCE(parent_run_id::text, ''), started_at, completed_at
		FROM mission_runs
		WHERE mission_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, missionID, limit)
	if err != nil {
		return nil, fmt.Errorf("runs: list query failed: %w", err)
	}
	defer rows.Close()

	var runs []MissionRun
	for rows.Next() {
		var run MissionRun
		var parentRunID sql.NullString
		var completedAt sql.NullTime
		if err := rows.Scan(
			&run.ID, &run.MissionID, &run.TenantID, &run.Status, &run.RunDepth,
			&parentRunID, &run.StartedAt, &completedAt,
		); err != nil {
			log.Printf("[runs] scan error: %v", err)
			continue
		}
		if parentRunID.Valid {
			run.ParentRunID = parentRunID.String
		}
		if completedAt.Valid {
			t := completedAt.Time
			run.CompletedAt = &t
		}
		runs = append(runs, run)
	}

	if runs == nil {
		runs = []MissionRun{}
	}
	return runs, nil
}
