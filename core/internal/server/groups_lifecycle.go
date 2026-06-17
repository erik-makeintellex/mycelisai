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

const staleGroupReviewAge = 72 * time.Hour

type groupLifecycleSummary struct {
	TotalGroups              int `json:"total_groups"`
	ActiveGroups             int `json:"active_groups"`
	ExpiredActiveGroups      int `json:"expired_active_groups"`
	StandingNoExpiryGroups   int `json:"standing_no_expiry_groups"`
	StaleStandingGroups      int `json:"stale_standing_groups"`
	ReviewNeededGroups       int `json:"review_needed_groups"`
	OutputReadyIdleGroups    int `json:"output_ready_idle_groups"`
	TeamWorkNeedingAttention int `json:"team_work_needing_attention"`
	ArchivedByLastCleanup    int `json:"archived_by_last_cleanup,omitempty"`
}

type groupLifecycleReport struct {
	GeneratedAt time.Time             `json:"generated_at"`
	Summary     groupLifecycleSummary `json:"summary"`
	Items       []groupLifecycleItem  `json:"items"`
}

type groupLifecycleItem struct {
	GroupID                  string     `json:"group_id"`
	Name                     string     `json:"name"`
	Status                   string     `json:"status"`
	WorkMode                 string     `json:"work_mode"`
	Kind                     string     `json:"kind"`
	Recommendation           string     `json:"recommendation"`
	Reason                   string     `json:"reason"`
	Expiry                   *time.Time `json:"expiry,omitempty"`
	Expired                  bool       `json:"expired"`
	AgeHours                 int        `json:"age_hours"`
	TeamCount                int        `json:"team_count"`
	OutputCount              int        `json:"output_count"`
	TeamWorkCount            int        `json:"team_work_count"`
	ActiveOrBlockedWorkCount int        `json:"active_or_blocked_work_count"`
	OutputReadyWorkCount     int        `json:"output_ready_work_count"`
	ArchivedWorkCount        int        `json:"archived_work_count"`
	LatestWorkAt             *time.Time `json:"latest_work_at,omitempty"`
}

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

func buildGroupLifecycleItem(group CollaborationGroup, stats groupTeamWorkStats, outputCount int, now time.Time) groupLifecycleItem {
	expired := group.Expiry != nil && !group.Expiry.After(now)
	kind := "standing"
	if group.Status == groupStatusArchived {
		kind = "archived"
	} else if expired {
		kind = "expired"
	} else if group.Expiry != nil {
		kind = "temporary"
	}

	recommendation, reason := groupLifecycleRecommendation(group, stats, outputCount, expired, now)
	ageHours := 0
	if !group.CreatedAt.IsZero() {
		ageHours = int(now.Sub(group.CreatedAt.UTC()).Hours())
		if ageHours < 0 {
			ageHours = 0
		}
	}
	return groupLifecycleItem{
		GroupID:                  group.ID,
		Name:                     group.Name,
		Status:                   group.Status,
		WorkMode:                 group.WorkMode,
		Kind:                     kind,
		Recommendation:           recommendation,
		Reason:                   reason,
		Expiry:                   group.Expiry,
		Expired:                  expired,
		AgeHours:                 ageHours,
		TeamCount:                len(group.TeamIDs),
		OutputCount:              outputCount,
		TeamWorkCount:            stats.Total,
		ActiveOrBlockedWorkCount: stats.ActiveOrBlocked,
		OutputReadyWorkCount:     stats.OutputReady,
		ArchivedWorkCount:        stats.Archived,
		LatestWorkAt:             stats.LatestWorkAt,
	}
}

func groupLifecycleRecommendation(group CollaborationGroup, stats groupTeamWorkStats, outputCount int, expired bool, now time.Time) (string, string) {
	if group.Status == groupStatusArchived {
		return "retained", "Archived group record; retained outputs remain reviewable."
	}
	if expired {
		return "archive_expired", "Temporary group expiry has passed."
	}
	if stats.ActiveOrBlocked > 0 {
		return "review_work", "Linked team work still needs output, recovery, or operator review."
	}
	if outputCount > 0 || stats.OutputReady > 0 {
		return "archive_completed", "Outputs are retained and no linked team work is active."
	}
	if group.Expiry == nil && group.Status == groupStatusActive && isGroupStale(group, stats, now) {
		return "review_standing", "Standing group has no expiry and has not produced retained output recently."
	}
	if group.Expiry != nil {
		return "keep_active", "Temporary group is still inside its expiry window."
	}
	return "keep_active", "Standing group is retained for ongoing work."
}

func isGroupStale(group CollaborationGroup, stats groupTeamWorkStats, now time.Time) bool {
	reference := group.UpdatedAt
	if stats.LatestWorkAt != nil && stats.LatestWorkAt.After(reference) {
		reference = *stats.LatestWorkAt
	}
	if reference.IsZero() {
		reference = group.CreatedAt
	}
	if reference.IsZero() {
		return false
	}
	return now.Sub(reference.UTC()) >= staleGroupReviewAge
}

func accumulateGroupLifecycleSummary(summary *groupLifecycleSummary, item groupLifecycleItem) {
	if item.Status == groupStatusActive {
		summary.ActiveGroups++
	}
	if item.Expiry == nil && item.Status != groupStatusArchived {
		summary.StandingNoExpiryGroups++
	}
	switch item.Recommendation {
	case "archive_expired":
		summary.ExpiredActiveGroups++
		summary.ReviewNeededGroups++
	case "review_work":
		summary.ReviewNeededGroups++
	case "review_standing":
		summary.StaleStandingGroups++
		summary.ReviewNeededGroups++
	case "archive_completed":
		summary.OutputReadyIdleGroups++
		summary.ReviewNeededGroups++
	}
	summary.TeamWorkNeedingAttention += item.ActiveOrBlockedWorkCount
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
