package server

import "time"

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
		AgeHours:                 groupAgeHours(group, now),
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

func groupAgeHours(group CollaborationGroup, now time.Time) int {
	if group.CreatedAt.IsZero() {
		return 0
	}
	ageHours := int(now.Sub(group.CreatedAt.UTC()).Hours())
	if ageHours < 0 {
		return 0
	}
	return ageHours
}

func isGroupStale(group CollaborationGroup, stats groupTeamWorkStats, now time.Time) bool {
	reference := group.UpdatedAt
	if stats.LatestWorkAt != nil && stats.LatestWorkAt.After(reference) {
		reference = *stats.LatestWorkAt
	}
	if reference.IsZero() {
		reference = group.CreatedAt
	}
	return !reference.IsZero() && now.Sub(reference.UTC()) >= staleGroupReviewAge
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
