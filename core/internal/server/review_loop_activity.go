package server

import "strings"

func safeActivityName(loopName string) string {
	name := strings.TrimSpace(loopName)
	if name == "" {
		return "Organization review"
	}
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "department"):
		return "Department check"
	case strings.Contains(lower, "agent type"):
		return "Specialist review"
	default:
		return name
	}
}

func recentLoopActivity(results []ReviewLoopResult, limit int) []LoopActivityItem {
	if len(results) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 5
	}
	if len(results) > limit {
		results = results[:limit]
	}

	items := make([]LoopActivityItem, 0, len(results))
	for _, result := range results {
		items = append(items, LoopActivityItem{
			ID:        result.ID,
			Name:      safeActivityName(result.LoopName),
			LastRunAt: result.ReviewedAt,
			Status:    result.ActivityStatus,
			Summary:   result.ActivitySummary,
		})
	}
	return items
}
