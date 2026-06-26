package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type groupTeamWorkStats struct {
	Total           int
	ActiveOrBlocked int
	OutputReady     int
	Archived        int
	LatestWorkAt    *time.Time
}

type archiveExpiredGroupsResult struct {
	ArchivedCount    int                  `json:"archived_count"`
	ArchivedGroupIDs []string             `json:"archived_group_ids"`
	Report           groupLifecycleReport `json:"report"`
}

// GET /api/v1/groups/lifecycle
func (s *AdminServer) HandleGroupLifecycleReport(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}
	report, err := s.buildGroupLifecycleReport(r.Context(), 0)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(report))
}

// POST /api/v1/groups/lifecycle/archive-expired
func (s *AdminServer) HandleArchiveExpiredGroups(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:write"); !ok {
		return
	}
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer,
		"groups.lifecycle",
		"Expired temporary groups archived",
		map[string]any{
			"operation": "groups.lifecycle.archive_expired",
		},
	)
	archivedIDs, err := s.archiveExpiredGroupsDB(r.Context(), auditEventID)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	report, err := s.buildGroupLifecycleReport(r.Context(), len(archivedIDs))
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(archiveExpiredGroupsResult{
		ArchivedCount:    len(archivedIDs),
		ArchivedGroupIDs: archivedIDs,
		Report:           report,
	}))
}

func (s *AdminServer) buildGroupLifecycleReport(ctx context.Context, archivedByLastCleanup int) (groupLifecycleReport, error) {
	groups, err := s.listGroupsDB(ctx)
	if err != nil {
		return groupLifecycleReport{}, err
	}
	now := time.Now().UTC()
	report := groupLifecycleReport{
		GeneratedAt: now,
		Summary: groupLifecycleSummary{
			TotalGroups:           len(groups),
			ArchivedByLastCleanup: archivedByLastCleanup,
		},
		Items: make([]groupLifecycleItem, 0, len(groups)),
	}

	for _, group := range groups {
		stats, statsErr := s.groupTeamWorkStats(ctx, group.TeamIDs)
		if statsErr != nil {
			return groupLifecycleReport{}, statsErr
		}
		outputCount := s.groupOutputCount(ctx, &group)
		item := buildGroupLifecycleItem(group, stats, outputCount, now)
		report.Items = append(report.Items, item)
		accumulateGroupLifecycleSummary(&report.Summary, item)
	}
	return report, nil
}

func (s *AdminServer) groupOutputCount(ctx context.Context, group *CollaborationGroup) int {
	outputs, err := s.listGroupOutputs(ctx, group, 100)
	if err != nil {
		return 0
	}
	return len(outputs)
}

func (s *AdminServer) groupTeamWorkStats(ctx context.Context, teamIDs []string) (groupTeamWorkStats, error) {
	var out groupTeamWorkStats
	for _, rawTeamID := range teamIDs {
		teamID := strings.TrimSpace(rawTeamID)
		if teamID == "" {
			continue
		}
		next, err := s.teamWorkStatsForTeam(ctx, teamID)
		if err != nil {
			return out, err
		}
		out.Total += next.Total
		out.ActiveOrBlocked += next.ActiveOrBlocked
		out.OutputReady += next.OutputReady
		out.Archived += next.Archived
		if next.LatestWorkAt != nil && (out.LatestWorkAt == nil || next.LatestWorkAt.After(*out.LatestWorkAt)) {
			latest := next.LatestWorkAt.UTC()
			out.LatestWorkAt = &latest
		}
	}
	return out, nil
}

func (s *AdminServer) teamWorkStatsForTeam(ctx context.Context, teamID string) (groupTeamWorkStats, error) {
	db := s.getDB()
	if db == nil {
		return groupTeamWorkStats{}, errors.New("database not available")
	}
	var stats groupTeamWorkStats
	var latest sql.NullTime
	err := db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE state IN ('new','briefed','queued','running','needs_operator','reviewing','degraded','paused'))::int,
			COUNT(*) FILTER (WHERE state='output_ready')::int,
			COUNT(*) FILTER (WHERE state='archived')::int,
			MAX(updated_at)
		FROM team_work_items
		WHERE tenant_id='default' AND team_id=$1`, teamID).
		Scan(&stats.Total, &stats.ActiveOrBlocked, &stats.OutputReady, &stats.Archived, &latest)
	if err != nil {
		return stats, err
	}
	if latest.Valid {
		ts := latest.Time.UTC()
		stats.LatestWorkAt = &ts
	}
	return stats, nil
}

func (s *AdminServer) archiveExpiredGroupsDB(ctx context.Context, updatedAuditEventID string) ([]string, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		UPDATE collaboration_groups
		SET status=$1,
		    updated_audit_event_id=$2,
		    updated_at=NOW()
		WHERE tenant_id='default'
		  AND status=$3
		  AND expiry IS NOT NULL
		  AND expiry <= NOW()
		RETURNING id::text`,
		groupStatusArchived,
		parseAuditUUID(updatedAuditEventID),
		groupStatusActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
