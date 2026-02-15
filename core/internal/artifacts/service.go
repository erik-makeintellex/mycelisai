package artifacts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ArtifactType classifies the output category.
type ArtifactType string

const (
	TypeCode     ArtifactType = "code"
	TypeDocument ArtifactType = "document"
	TypeImage    ArtifactType = "image"
	TypeAudio    ArtifactType = "audio"
	TypeData     ArtifactType = "data"
	TypeFile     ArtifactType = "file"
	TypeChart    ArtifactType = "chart"
)

// Artifact represents a structured agent output stored in the ledger.
type Artifact struct {
	ID            uuid.UUID       `json:"id"`
	MissionID     *uuid.UUID      `json:"mission_id,omitempty"`
	TeamID        *uuid.UUID      `json:"team_id,omitempty"`
	AgentID       string          `json:"agent_id"`
	TraceID       string          `json:"trace_id,omitempty"`
	ArtifactType  ArtifactType    `json:"artifact_type"`
	Title         string          `json:"title"`
	ContentType   string          `json:"content_type"` // MIME
	Content       string          `json:"content,omitempty"`
	FilePath      string          `json:"file_path,omitempty"`
	FileSizeBytes int64           `json:"file_size_bytes,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	TrustScore    *float64        `json:"trust_score,omitempty"`
	Status        string          `json:"status"` // pending, approved, rejected, archived
	CreatedAt     time.Time       `json:"created_at"`
}

// Service manages artifact persistence and retrieval.
type Service struct {
	DB      *sql.DB
	DataDir string // base path for file-referenced artifacts (e.g., /data/artifacts)
}

// NewService creates an artifacts service.
func NewService(db *sql.DB, dataDir string) *Service {
	return &Service{DB: db, DataDir: dataDir}
}

// Store persists a new artifact and returns it with generated ID.
func (s *Service) Store(ctx context.Context, a Artifact) (*Artifact, error) {
	if a.Metadata == nil {
		a.Metadata = json.RawMessage(`{}`)
	}

	var missionID, teamID *uuid.UUID
	if a.MissionID != nil && *a.MissionID != uuid.Nil {
		missionID = a.MissionID
	}
	if a.TeamID != nil && *a.TeamID != uuid.Nil {
		teamID = a.TeamID
	}

	var traceID, content, filePath sql.NullString
	var fileSize sql.NullInt64
	var trustScore sql.NullFloat64

	if a.TraceID != "" {
		traceID = sql.NullString{String: a.TraceID, Valid: true}
	}
	if a.Content != "" {
		content = sql.NullString{String: a.Content, Valid: true}
	}
	if a.FilePath != "" {
		filePath = sql.NullString{String: a.FilePath, Valid: true}
	}
	if a.FileSizeBytes > 0 {
		fileSize = sql.NullInt64{Int64: a.FileSizeBytes, Valid: true}
	}
	if a.TrustScore != nil {
		trustScore = sql.NullFloat64{Float64: *a.TrustScore, Valid: true}
	}

	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO artifacts (mission_id, team_id, agent_id, trace_id, artifact_type,
		                       title, content_type, content, file_path, file_size_bytes,
		                       metadata, trust_score, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`, missionID, teamID, a.AgentID, traceID, a.ArtifactType,
		a.Title, a.ContentType, content, filePath, fileSize,
		a.Metadata, trustScore, a.Status)

	if err := row.Scan(&a.ID, &a.CreatedAt); err != nil {
		return nil, fmt.Errorf("store artifact: %w", err)
	}
	return &a, nil
}

// ListByMission returns artifacts for a specific mission.
func (s *Service) ListByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,
		       title, content_type, content, file_path, file_size_bytes,
		       metadata, trust_score, status, created_at
		FROM artifacts
		WHERE mission_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, missionID, limit)
	if err != nil {
		return nil, fmt.Errorf("list by mission: %w", err)
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

// ListByTeam returns artifacts for a specific team.
func (s *Service) ListByTeam(ctx context.Context, teamID uuid.UUID, limit int) ([]Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,
		       title, content_type, content, file_path, file_size_bytes,
		       metadata, trust_score, status, created_at
		FROM artifacts
		WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("list by team: %w", err)
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

// ListByAgent returns artifacts for a specific agent.
func (s *Service) ListByAgent(ctx context.Context, agentID string, limit int) ([]Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,
		       title, content_type, content, file_path, file_size_bytes,
		       metadata, trust_score, status, created_at
		FROM artifacts
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("list by agent: %w", err)
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

// ListRecent returns the most recent artifacts across all missions.
func (s *Service) ListRecent(ctx context.Context, limit int) ([]Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,
		       title, content_type, content, file_path, file_size_bytes,
		       metadata, trust_score, status, created_at
		FROM artifacts
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent: %w", err)
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

// Get retrieves a single artifact by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Artifact, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,
		       title, content_type, content, file_path, file_size_bytes,
		       metadata, trust_score, status, created_at
		FROM artifacts
		WHERE id = $1
	`, id)

	a, err := scanArtifactRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("artifact %s not found", id)
		}
		return nil, fmt.Errorf("get artifact: %w", err)
	}
	return a, nil
}

// UpdateStatus changes the governance status of an artifact.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result, err := s.DB.ExecContext(ctx, `UPDATE artifacts SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("artifact %s not found", id)
	}
	return nil
}

// scanArtifacts scans multiple rows into Artifact slices.
func scanArtifacts(rows *sql.Rows) ([]Artifact, error) {
	var result []Artifact
	for rows.Next() {
		a, err := scanArtifactFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *a)
	}
	return result, rows.Err()
}

func scanArtifactFromRows(rows *sql.Rows) (*Artifact, error) {
	a := &Artifact{}
	var (
		missionID, teamID      *uuid.UUID
		traceID, content, path sql.NullString
		fileSize               sql.NullInt64
		trustScore             sql.NullFloat64
		metadataJSON           []byte
	)

	if err := rows.Scan(
		&a.ID, &missionID, &teamID, &a.AgentID, &traceID, &a.ArtifactType,
		&a.Title, &a.ContentType, &content, &path, &fileSize,
		&metadataJSON, &trustScore, &a.Status, &a.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan artifact: %w", err)
	}

	a.MissionID = missionID
	a.TeamID = teamID
	a.TraceID = traceID.String
	a.Content = content.String
	a.FilePath = path.String
	a.FileSizeBytes = fileSize.Int64
	if trustScore.Valid {
		a.TrustScore = &trustScore.Float64
	}
	a.Metadata = metadataJSON
	if a.Metadata == nil {
		a.Metadata = json.RawMessage(`{}`)
	}

	return a, nil
}

func scanArtifactRow(row *sql.Row) (*Artifact, error) {
	a := &Artifact{}
	var (
		missionID, teamID      *uuid.UUID
		traceID, content, path sql.NullString
		fileSize               sql.NullInt64
		trustScore             sql.NullFloat64
		metadataJSON           []byte
	)

	if err := row.Scan(
		&a.ID, &missionID, &teamID, &a.AgentID, &traceID, &a.ArtifactType,
		&a.Title, &a.ContentType, &content, &path, &fileSize,
		&metadataJSON, &trustScore, &a.Status, &a.CreatedAt,
	); err != nil {
		return nil, err
	}

	a.MissionID = missionID
	a.TeamID = teamID
	a.TraceID = traceID.String
	a.Content = content.String
	a.FilePath = path.String
	a.FileSizeBytes = fileSize.Int64
	if trustScore.Valid {
		a.TrustScore = &trustScore.Float64
	}
	a.Metadata = metadataJSON
	if a.Metadata == nil {
		a.Metadata = json.RawMessage(`{}`)
	}

	return a, nil
}
