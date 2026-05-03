package artifacts

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type artifactScanner interface {
	Scan(dest ...any) error
}

// scanArtifacts scans multiple rows into Artifact slices.
func scanArtifacts(rows *sql.Rows) ([]Artifact, error) {
	var result []Artifact
	for rows.Next() {
		a, err := scanArtifactRowValues(rows)
		if err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		result = append(result, *a)
	}
	return result, rows.Err()
}

func scanArtifactRow(row *sql.Row) (*Artifact, error) {
	return scanArtifactRowValues(row)
}

func scanArtifactRowValues(scanner artifactScanner) (*Artifact, error) {
	a := &Artifact{}
	var (
		missionID, teamID      *uuid.UUID
		traceID, content, path sql.NullString
		fileSize               sql.NullInt64
		trustScore             sql.NullFloat64
		metadataJSON           []byte
	)

	if err := scanner.Scan(
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
