package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// deploymentsCoreStartedAt approximates when boot-loaded standing teams began.
var deploymentsCoreStartedAt = time.Now().UTC()

// missionRuntimeMissionID returns the mission id when a Soma team id has the
// "<mission uuid>.<team>" shape produced by ConvertBlueprintToManifests.
func missionRuntimeMissionID(teamID string) (string, bool) {
	prefix, _, found := strings.Cut(teamID, ".")
	if !found {
		return "", false
	}
	if _, err := uuid.Parse(prefix); err != nil {
		return "", false
	}
	return prefix, true
}

// collectStandingTeams lists boot-loaded Soma teams (not mission teams).
func (c *deploymentCollector) collectStandingTeams() {
	if !c.runtimeOK {
		c.addSource(deploymentSourceStanding, protocol.DeploymentSourceUnavailable, "Soma runtime is not running", 0)
		return
	}
	count := 0
	for _, tm := range c.runtime {
		if tm == nil {
			continue
		}
		if _, isMission := missionRuntimeMissionID(tm.ID); isMission {
			continue
		}
		members := c.teamMembers(tm)
		rec := protocol.DeploymentRecord{
			ID:             tm.ID,
			Kind:           protocol.DeploymentKindStandingTeam,
			Name:           deploymentFirstNonEmpty(tm.Name, tm.ID),
			Purpose:        tm.Description,
			WhatRuns:       agentCountSentence(len(members), "in a standing team"),
			Scope:          protocol.DeploymentScope{Kind: protocol.DeploymentScopeGlobal, ID: "core"},
			Origin:         "bundle",
			RuntimeID:      tm.ID,
			LifecycleState: teamLifecycle(members),
			Since:          deploymentTime(deploymentsCoreStartedAt),
			Members:        members,
			Tools:          unionMemberTools(members),
			InputChannels:  tm.Inputs,
			OutputChannels: tm.Deliveries,
			ConfigRef:      deploymentConfigRef("team", tm.ID),
		}
		if hb := latestMemberHeartbeat(members); hb != nil {
			rec.LastActivity = hb.LastHeartbeatAt
		}
		c.records = append(c.records, rec)
		count++
	}
	c.addSource(deploymentSourceStanding, protocol.DeploymentSourceOK, "", count)
}

const deploymentMissionTeamsSQL = `
	SELECT t.id::text, t.name, COALESCE(t.role, ''), t.mission_id::text,
	       COALESCE(m.name, ''), COALESCE(m.directive, ''), COALESCE(m.status, 'draft'),
	       t.created_at, m.activated_at
	FROM teams t
	JOIN missions m ON m.id = t.mission_id
	ORDER BY t.created_at DESC
	LIMIT 500`

const deploymentMissionMembersSQL = `
	SELECT sm.team_id::text, sm.manifest
	FROM service_manifests sm
	JOIN teams t ON t.id = sm.team_id
	WHERE t.mission_id IS NOT NULL
	ORDER BY sm.name`

type deploymentMissionRow struct {
	teamID, name, role, missionID, missionName, directive, missionStatus string
	createdAt, activatedAt                                               sql.NullTime
}

// collectMissionTeams lists DB teams linked to missions, keyed by teams.id.
// Runtime state is matched within the mission (never across missions by name).
func (c *deploymentCollector) collectMissionTeams() {
	db := c.s.getDB()
	if db == nil {
		detail := "database not configured"
		if hidden := c.runtimeMissionTeamCount(); hidden > 0 {
			detail = fmt.Sprintf("database not configured; %d running mission team(s) hidden until it is reachable", hidden)
		}
		c.addSource(deploymentSourceMission, protocol.DeploymentSourceUnavailable, detail, 0)
		return
	}
	rows, err := db.QueryContext(c.ctx, deploymentMissionTeamsSQL)
	if err != nil {
		log.Printf("[deployments] mission teams query: %v", err)
		c.addSource(deploymentSourceMission, protocol.DeploymentSourceUnavailable, "mission teams query failed", 0)
		return
	}
	var list []deploymentMissionRow
	for rows.Next() {
		var row deploymentMissionRow
		if err := rows.Scan(&row.teamID, &row.name, &row.role, &row.missionID, &row.missionName,
			&row.directive, &row.missionStatus, &row.createdAt, &row.activatedAt); err != nil {
			rows.Close()
			log.Printf("[deployments] mission teams scan: %v", err)
			c.addSource(deploymentSourceMission, protocol.DeploymentSourceUnavailable, "mission teams scan failed", 0)
			return
		}
		list = append(list, row)
	}
	rows.Close()

	members, membersErr := c.loadMissionMembers()
	status, detail := protocol.DeploymentSourceOK, ""
	if membersErr != nil {
		log.Printf("[deployments] mission members query: %v", membersErr)
		status, detail = protocol.DeploymentSourceDegraded, "mission team members unavailable"
	}
	for _, row := range list {
		c.records = append(c.records, c.missionTeamRecord(row, members[row.teamID]))
	}
	c.addSource(deploymentSourceMission, status, detail, len(list))
}

func (c *deploymentCollector) loadMissionMembers() (map[string][]protocol.AgentManifest, error) {
	out := map[string][]protocol.AgentManifest{}
	rows, err := c.s.getDB().QueryContext(c.ctx, deploymentMissionMembersSQL)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var teamID string
		var raw []byte
		if err := rows.Scan(&teamID, &raw); err != nil {
			return out, err
		}
		var agent protocol.AgentManifest
		if err := json.Unmarshal(raw, &agent); err != nil {
			continue
		}
		out[teamID] = append(out[teamID], agent)
	}
	return out, rows.Err()
}

func (c *deploymentCollector) missionTeamRecord(row deploymentMissionRow, agents []protocol.AgentManifest) protocol.DeploymentRecord {
	runtime := c.findMissionRuntime(row.missionID, row.name)
	var members []protocol.DeploymentMember
	if runtime != nil && len(runtime.Members) > 0 {
		members = c.teamMembers(runtime)
	} else {
		members = make([]protocol.DeploymentMember, 0, len(agents))
		for _, agent := range agents {
			members = append(members, c.deploymentAgentMember(agent))
		}
	}
	rec := protocol.DeploymentRecord{
		ID:             row.teamID,
		Kind:           protocol.DeploymentKindMissionTeam,
		Name:           deploymentFirstNonEmpty(row.name, row.teamID),
		Purpose:        deploymentFirstNonEmpty(row.role, row.directive),
		WhatRuns:       agentCountSentence(len(members), "on a mission"),
		Scope:          protocol.DeploymentScope{Kind: protocol.DeploymentScopeMission, ID: row.missionID, Name: deploymentFirstNonEmpty(row.missionName, row.directive)},
		Origin:         "mission:" + row.missionID,
		LifecycleState: missionLifecycle(row.missionStatus, runtime != nil, members),
		Members:        members,
		Tools:          unionMemberTools(members),
		ConfigRef:      deploymentConfigRef("team", row.teamID),
	}
	if runtime != nil {
		rec.RuntimeID = runtime.ID
		rec.InputChannels = runtime.Inputs
		rec.OutputChannels = runtime.Deliveries
	}
	if row.createdAt.Valid {
		rec.Since = deploymentTime(row.createdAt.Time)
	}
	if row.activatedAt.Valid {
		rec.LastActivity = deploymentTime(row.activatedAt.Time)
	}
	if hb := latestMemberHeartbeat(members); hb != nil {
		rec.LastActivity = laterDeploymentTime(rec.LastActivity, hb.LastHeartbeatAt)
	}
	return rec
}

// findMissionRuntime finds the live Soma team for a DB mission team: the id
// must carry the mission prefix and the manifest name must match.
func (c *deploymentCollector) findMissionRuntime(missionID, name string) *swarm.TeamManifest {
	for _, tm := range c.runtime {
		if tm == nil {
			continue
		}
		if mid, ok := missionRuntimeMissionID(tm.ID); ok && mid == missionID && tm.Name == name {
			return tm
		}
	}
	return nil
}

func (c *deploymentCollector) runtimeMissionTeamCount() int {
	n := 0
	for _, tm := range c.runtime {
		if tm == nil {
			continue
		}
		if _, ok := missionRuntimeMissionID(tm.ID); ok {
			n++
		}
	}
	return n
}

// missionLifecycle maps a mission status plus runtime presence to a state.
func missionLifecycle(missionStatus string, live bool, members []protocol.DeploymentMember) string {
	if live {
		return teamLifecycle(members)
	}
	switch strings.ToLower(strings.TrimSpace(missionStatus)) {
	case "active":
		return protocol.DeploymentStateIdle
	case "draft", "pending", "planned":
		return protocol.DeploymentStatePlanned
	case "paused":
		return protocol.DeploymentStatePaused
	case "failed", "error":
		return protocol.DeploymentStateError
	default:
		return protocol.DeploymentStateArchived
	}
}

func deploymentFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
