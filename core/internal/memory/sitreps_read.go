package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// SitRep reads (MEM-LANES-2). A SitRep summarizes one team's activity, so
// the HTTP route decides which teams a caller may read (proven membership,
// or every team for a scoped root admin) and these reads never widen that:
// an empty team list reads nothing.

const sitRepColumns = `SELECT id, team_id, timestamp, time_window_start, time_window_end, summary, key_events, strategies, status FROM sitreps`

// ListSitReps retrieves recent SitReps for one team.
func (s *Service) ListSitReps(ctx context.Context, teamID string, limit int) ([]map[string]any, error) {
	return s.ListSitRepsForTeams(ctx, []string{teamID}, limit)
}

// ListSitRepsForTeams retrieves recent SitReps of the given teams, newest
// first. No teams returns none, never every team.
func (s *Service) ListSitRepsForTeams(ctx context.Context, teamIDs []string, limit int) ([]map[string]any, error) {
	ids := nonEmpty(teamIDs)
	if len(ids) == 0 {
		return []map[string]any{}, nil
	}
	return s.querySitReps(ctx, sitRepColumns+` WHERE team_id::text = ANY($1::text[]) ORDER BY timestamp DESC LIMIT $2`, textArrayLiteral(ids), sitRepLimit(limit))
}

// ListAllSitReps retrieves recent SitReps of every team (the scoped
// root-admin read).
func (s *Service) ListAllSitReps(ctx context.Context, limit int) ([]map[string]any, error) {
	return s.querySitReps(ctx, sitRepColumns+` ORDER BY timestamp DESC LIMIT $1`, sitRepLimit(limit))
}

// SitRepTeamIDs lists the teams that have SitReps.
func (s *Service) SitRepTeamIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT team_id::text FROM sitreps WHERE team_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, strings.TrimSpace(id))
	}
	return ids, rows.Err()
}

func sitRepLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	return limit
}

func (s *Service) querySitReps(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []map[string]any{}
	for rows.Next() {
		var id, tID, summary, status string
		var strategies sql.NullString
		var ts, winStart, winEnd time.Time
		var keyEventsJSON []byte

		if err := rows.Scan(&id, &tID, &ts, &winStart, &winEnd, &summary, &keyEventsJSON, &strategies, &status); err != nil {
			return nil, err
		}
		entry := map[string]any{
			"id":                id,
			"team_id":           tID,
			"timestamp":         ts,
			"time_window_start": winStart,
			"time_window_end":   winEnd,
			"summary":           summary,
			"strategies":        strategies.String,
			"status":            status,
		}
		var keyEvents []string
		if json.Unmarshal(keyEventsJSON, &keyEvents) == nil {
			entry["key_events"] = keyEvents
		}
		results = append(results, entry)
	}
	return results, rows.Err()
}
