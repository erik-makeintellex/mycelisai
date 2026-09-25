package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/pkg/protocol"
)

// workModePostures is the one-sentence approval posture per work mode.
var workModePostures = map[string]string{
	protocol.WorkModeReadOnly:            "Read only: nothing is changed.",
	protocol.WorkModeProposeOnly:         "Every action is a proposal until you approve it.",
	protocol.WorkModeExecuteWithApproval: "Actions run after you approve them.",
	protocol.WorkModeExecuteBounded:      "Bounded actions run; high-risk actions need approval.",
	protocol.WorkModeResumeContinuity:    "Continues retained work within its approved scope.",
}

// collectGroups lists collaboration groups. Groups follow the same access
// rule as /api/v1/groups (root admin with groups:read); other callers get a
// `restricted` source note instead of group rows.
func (c *deploymentCollector) collectGroups() {
	if c.identity == nil || c.identity.Role != "admin" || !hasScope(c.identity, "groups:read") {
		c.addSource(deploymentSourceGroups, protocol.DeploymentSourceRestricted, "groups require root admin with groups:read", 0)
		return
	}
	if c.s.getDB() == nil {
		c.addSource(deploymentSourceGroups, protocol.DeploymentSourceUnavailable, "database not configured", 0)
		return
	}
	groups, err := c.s.listGroupsDB(c.ctx)
	if err != nil {
		log.Printf("[deployments] groups query: %v", err)
		c.addSource(deploymentSourceGroups, protocol.DeploymentSourceUnavailable, "groups query failed", 0)
		return
	}
	for i := range groups {
		c.records = append(c.records, c.groupRecord(&groups[i]))
	}
	c.addSource(deploymentSourceGroups, protocol.DeploymentSourceOK, "", len(groups))
}

func (c *deploymentCollector) groupRecord(g *CollaborationGroup) protocol.DeploymentRecord {
	members := []protocol.DeploymentMember{}
	if profile := strings.TrimSpace(g.CoordinatorProfile); profile != "" {
		coordinator := protocol.DeploymentMember{ID: "coordinator", Role: "coordinator", Profile: profile, Tools: []protocol.DeploymentTool{}}
		c.resolveMemberProvider(&coordinator, profile, "")
		members = append(members, coordinator)
	}
	for _, teamID := range g.TeamIDs {
		members = append(members, protocol.DeploymentMember{ID: teamID, Role: "team", Tools: []protocol.DeploymentTool{}})
	}
	for _, userID := range g.MemberUserIDs {
		members = append(members, protocol.DeploymentMember{ID: userID, Role: "person", Tools: []protocol.DeploymentTool{}})
	}
	rec := protocol.DeploymentRecord{
		ID:                g.ID,
		Kind:              protocol.DeploymentKindGroup,
		Name:              g.Name,
		Purpose:           g.GoalStatement,
		WhatRuns:          fmt.Sprintf("Group with %d team(s) and %d person(s)", len(g.TeamIDs), len(g.MemberUserIDs)),
		Scope:             protocol.DeploymentScope{Kind: protocol.DeploymentScopeGlobal, ID: deploymentFirstNonEmpty(g.TenantID, "default")},
		Origin:            "operator",
		LifecycleState:    c.groupLifecycle(g),
		Since:             deploymentTime(g.CreatedAt),
		ExpiresAt:         g.Expiry,
		LastActivity:      deploymentTime(g.UpdatedAt),
		Members:           members,
		Tools:             deploymentTools(g.AllowedCapabilities, "capability"),
		WorkMode:          g.WorkMode,
		ApprovalPolicyRef: g.ApprovalPolicyRef,
		ApprovalPosture:   workModePostures[g.WorkMode],
		ConfigRef:         deploymentConfigRef("group", g.ID),
	}
	return rec
}

func (c *deploymentCollector) groupLifecycle(g *CollaborationGroup) string {
	switch g.Status {
	case groupStatusPaused:
		return protocol.DeploymentStatePaused
	case groupStatusArchived:
		return protocol.DeploymentStateArchived
	}
	if g.Expiry != nil && g.Expiry.Before(c.now) {
		return protocol.DeploymentStateExpired
	}
	return protocol.DeploymentStateRunning
}

// collectPlannedDepartments lists departments of persisted organizations.
// Owner decision 2: starters show departments as planned; nothing runs until
// the operator approves a launch.
func (c *deploymentCollector) collectPlannedDepartments() {
	store := c.s.organizationStore()
	summaries, err := store.ListContext(c.ctx)
	if err != nil {
		log.Printf("[deployments] organizations list: %v", err)
		c.addSource(deploymentSourceOrganizations, protocol.DeploymentSourceUnavailable, "organization store unavailable", 0)
		return
	}
	status, detail := protocol.DeploymentSourceOK, ""
	if store.Backend() == "memory" {
		detail = "organizations are held in memory (not durable)"
	}
	defaults := c.templateDeliveryDefaults(summaries)
	count, failed := 0, 0
	for _, summary := range summaries {
		home, ok, getErr := store.GetContext(c.ctx, summary.ID)
		if getErr != nil || !ok {
			failed++
			continue
		}
		for _, dept := range home.Departments {
			c.records = append(c.records, c.plannedDepartmentRecord(home, dept, defaults[home.TemplateID]))
			count++
		}
	}
	if failed > 0 {
		status = protocol.DeploymentSourceDegraded
		detail = fmt.Sprintf("%d organization(s) could not be read", failed)
	}
	c.addSource(deploymentSourceOrganizations, status, detail, count)
}

// templateDeliveryDefaults loads starter delivery defaults once per request
// for the templates the listed organizations were created from.
func (c *deploymentCollector) templateDeliveryDefaults(summaries []OrganizationSummary) map[string]*bootstrap.DeliveryDefaults {
	out := map[string]*bootstrap.DeliveryDefaults{}
	needed := false
	for _, s := range summaries {
		if s.TemplateID != "" {
			needed = true
			break
		}
	}
	if !needed {
		return out
	}
	catalog, err := c.s.loadOrganizationStarterCatalog()
	if err != nil {
		return out
	}
	for _, entry := range catalog {
		if entry.DeliveryDefaults != nil {
			out[entry.ID] = entry.DeliveryDefaults
		}
	}
	return out
}

func (c *deploymentCollector) plannedDepartmentRecord(home OrganizationHomePayload, dept OrganizationDepartmentSummary, defaults *bootstrap.DeliveryDefaults) protocol.DeploymentRecord {
	members := make([]protocol.DeploymentMember, 0, len(dept.AgentTypeProfiles))
	for _, profile := range dept.AgentTypeProfiles {
		members = append(members, protocol.DeploymentMember{
			ID:             profile.ID,
			Role:           deploymentFirstNonEmpty(profile.Name, profile.ID),
			Profile:        deploymentFirstNonEmpty(profile.AIEngineEffectiveProfileID, dept.AIEngineEffectiveProfileID, home.AIEngineProfileID),
			Model:          profile.OutputModelEffectiveID,
			ProviderOrigin: protocol.ProviderOriginPlanned,
			ProviderStatus: "planned",
			EngineSummary:  deploymentFirstNonEmpty(profile.AIEngineEffectiveSummary, dept.AIEngineEffectiveSummary),
			Status:         "offline",
			Tools:          []protocol.DeploymentTool{},
		})
	}
	origin := "operator"
	if home.TemplateID != "" {
		origin = "template:" + home.TemplateID
	}
	specialists := dept.SpecialistCount
	if specialists == 0 {
		specialists = len(members)
	}
	rec := protocol.DeploymentRecord{
		ID:             home.ID + ":" + dept.ID,
		Kind:           protocol.DeploymentKindPlannedDepartment,
		Name:           dept.Name,
		Purpose:        home.Purpose,
		WhatRuns:       fmt.Sprintf("Planned: %d specialist(s), launched when you approve", specialists),
		Scope:          protocol.DeploymentScope{Kind: protocol.DeploymentScopeOrganization, ID: home.ID, Name: home.Name},
		Origin:         origin,
		LifecycleState: protocol.DeploymentStatePlanned,
		Members:        members,
		Tools:          []protocol.DeploymentTool{},
		ConfigRef:      deploymentConfigRef("organization", home.ID),
	}
	if defaults != nil {
		rec.WorkMode = defaults.WorkMode
		rec.ApprovalPolicyRef = defaults.ApprovalPolicyRef
		rec.ApprovalPosture = deploymentFirstNonEmpty(defaults.ApprovalPosture, workModePostures[defaults.WorkMode])
	}
	return rec
}
