package server

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/mycelis/core/internal/artifacts"
)

func (s *AdminServer) buildRetainedPackageContinuityFromState(ctx context.Context, home OrganizationHomePayload, requestContext string) (TeamLeadGuidanceResponse, bool) {
	if s == nil {
		return TeamLeadGuidanceResponse{}, false
	}

	pkg, err := s.latestRetainedPackage(ctx)
	if err != nil {
		log.Printf("team lead continuity enrichment skipped: %v", err)
		return TeamLeadGuidanceResponse{}, false
	}
	if pkg == nil {
		return TeamLeadGuidanceResponse{}, false
	}

	organizationName := safeOrganizationName(home.Name)
	teamLeadLabel := safeTeamLeadLabel(home.TeamLeadLabel)
	titles := retainedOutputTitles(pkg.Outputs, 3)
	nextOwner := retainedNextOwner(pkg.Group, pkg.Outputs)
	coordinator := retainedCoordinatorLabel(pkg.Group)
	groupState := retainedGroupStateLabel(pkg.Group)
	resumeGoal := strings.TrimSpace(requestContext)
	if resumeGoal == "" {
		resumeGoal = pkg.Group.GoalStatement
	}
	if strings.TrimSpace(resumeGoal) == "" {
		resumeGoal = fmt.Sprintf("Resume the retained package for %s after a reboot or reload.", organizationName)
	}

	return TeamLeadGuidanceResponse{
		Action:       TeamLeadGuidedActionResumeRetainedPackage,
		RequestLabel: "Resume retained package continuity",
		Headline:     fmt.Sprintf("Resume %s through %s", pkg.Group.Name, organizationName),
		Summary:      fmt.Sprintf("%s can resume %s through the %s retained package. %d retained output%s are already reviewable, and %s owns the next handoff after reload or reboot.", teamLeadLabel, organizationName, pkg.Group.Name, len(pkg.Outputs), pluralSuffix(len(pkg.Outputs)), nextOwner),
		PrioritySteps: []string{
			fmt.Sprintf("Reopen %s and review retained outputs like %s.", pkg.Group.Name, humanJoin(titles)),
			fmt.Sprintf("Treat %s as completed work already captured in the %s package.", humanJoin(titles), groupState),
			fmt.Sprintf("Hand the next step to %s after %s reviews the retained package summary.", nextOwner, coordinator),
		},
		SuggestedFollowUps: []string{
			fmt.Sprintf("Open %s for retained output review", pkg.Group.Name),
			"Plan next steps for this organization",
			"Review my organization setup",
		},
		ExecutionContract: &TeamLeadExecutionContract{
			ExecutionMode:     TeamLeadExecutionModeContinuityResume,
			OwnerLabel:        coordinator,
			Summary:           fmt.Sprintf("Resume %s from the %s retained package, confirm the finished outputs, and keep the next owner explicit instead of rebuilding finished work.", organizationName, pkg.Group.Name),
			ContinuityLabel:   pkg.Group.Name,
			ContinuitySummary: fmt.Sprintf("%s is %s and already retains %s.", pkg.Group.Name, groupState, humanJoin(titles)),
			ResumeCheckpoint:  fmt.Sprintf("Open %s and continue with %s after reviewing %s.", pkg.Group.Name, nextOwner, titles[0]),
			TargetOutputs:     titles,
			Workstreams:       buildContinuityExecutionWorkstreamsFromPackage(pkg.Group, pkg.Outputs, nextOwner),
			WorkflowGroup: &TeamLeadWorkflowGroupDraft{
				GroupID:                pkg.Group.ID,
				Name:                   pkg.Group.Name,
				GoalStatement:          resumeGoal,
				WorkMode:               "resume_continuity",
				CoordinatorProfile:     pkg.Group.CoordinatorProfile,
				AllowedCapabilities:    append([]string(nil), pkg.Group.AllowedCapabilities...),
				RecommendedMemberLimit: maxInt(len(pkg.Group.TeamIDs), 1),
				ExpiryHours:            retainedExpiryHours(pkg.Group),
				Summary:                fmt.Sprintf("Reopen the %s retained package, review %s, and continue with %s as the next owner.", pkg.Group.Name, humanJoin(titles), nextOwner),
			},
		},
	}, true
}

type retainedPackage struct {
	Group   CollaborationGroup
	Outputs []artifacts.Artifact
}

func (s *AdminServer) latestRetainedPackage(ctx context.Context) (*retainedPackage, error) {
	groups, err := s.listGroupsDB(ctx)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
	})
	for _, group := range groups {
		outputs, err := s.listGroupOutputs(ctx, &group, 8)
		if err != nil {
			return nil, err
		}
		if len(outputs) == 0 {
			continue
		}
		return &retainedPackage{
			Group:   group,
			Outputs: outputs,
		}, nil
	}
	return nil, nil
}

func retainedOutputTitles(outputs []artifacts.Artifact, limit int) []string {
	if limit <= 0 {
		limit = 3
	}
	titles := make([]string, 0, limit)
	seen := map[string]struct{}{}
	for _, output := range outputs {
		title := strings.TrimSpace(output.Title)
		if title == "" {
			title = strings.TrimSpace(output.AgentID)
		}
		if title == "" {
			continue
		}
		if _, ok := seen[title]; ok {
			continue
		}
		seen[title] = struct{}{}
		titles = append(titles, title)
		if len(titles) >= limit {
			break
		}
	}
	if len(titles) == 0 {
		return []string{"Retained package continuity summary"}
	}
	return titles
}

func retainedNextOwner(group CollaborationGroup, outputs []artifacts.Artifact) string {
	for _, output := range outputs {
		if label := humanizeOwnerLabel(output.AgentID); label != "" {
			return label
		}
	}
	return retainedCoordinatorLabel(group)
}

func retainedCoordinatorLabel(group CollaborationGroup) string {
	if label := humanizeOwnerLabel(group.CoordinatorProfile); label != "" {
		return label
	}
	return "Team Lead continuity"
}

func retainedGroupStateLabel(group CollaborationGroup) string {
	switch strings.TrimSpace(group.Status) {
	case groupStatusArchived:
		return "archived"
	case groupStatusPaused:
		return "paused"
	case groupStatusActive:
		return "active"
	default:
		return "retained"
	}
}

func retainedExpiryHours(group CollaborationGroup) int {
	if group.Expiry == nil {
		return 0
	}
	until := time.Until(group.Expiry.UTC())
	if until <= 0 {
		return 0
	}
	hours := int(until.Hours())
	if hours < 1 {
		return 1
	}
	return hours
}

func humanizeOwnerLabel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "_", " ")
	trimmed = strings.ReplaceAll(trimmed, "-", " ")
	parts := strings.Fields(trimmed)
	for index, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		parts[index] = string(runes)
	}
	return strings.Join(parts, " ")
}
