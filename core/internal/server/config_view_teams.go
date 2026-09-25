package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

const configMissionTeamSQL = `
	SELECT t.id::text, t.name, COALESCE(t.role, ''), t.mission_id::text,
	       COALESCE(m.name, ''), COALESCE(m.status, 'draft')
	FROM teams t
	JOIN missions m ON m.id = t.mission_id
	WHERE t.id = $1`

const configMissionMembersSQL = `
	SELECT manifest FROM service_manifests WHERE team_id = $1 ORDER BY name`

// teamConfigDocument resolves a team by runtime id first (standing teams and
// live mission teams by their runtime id), then by teams.id for mission teams.
func (s *AdminServer) teamConfigDocument(r *http.Request, id string) (*configDocument, error) {
	c := s.newDeploymentCollector(r)
	for _, tm := range c.runtime {
		if tm != nil && tm.ID == id {
			meta := map[string]any{"id": tm.ID, "name": tm.Name, "runtime_id": tm.ID}
			origin := "bundle"
			if missionID, ok := missionRuntimeMissionID(tm.ID); ok {
				meta["mission_id"] = missionID
				origin = "mission:" + missionID
			}
			meta["origin"] = origin
			return &configDocument{doc: c.teamManifestDoc(tm, meta, "runtime:soma"), source: configSourceRuntime}, nil
		}
	}
	if _, err := uuid.Parse(id); err != nil {
		if !c.runtimeOK {
			return nil, configUnavailable("Soma runtime is not running; standing team config is unavailable")
		}
		return nil, configNotFound("team", id)
	}
	return s.missionTeamConfigDocument(c, id)
}

// missionTeamConfigDocument reads a mission team keyed by teams.id. When the
// mission team is live, its runtime manifest is shown instead of the stored
// member rows.
func (s *AdminServer) missionTeamConfigDocument(c *deploymentCollector, id string) (*configDocument, error) {
	db := s.getDB()
	if db == nil {
		return nil, configUnavailable("database not configured; mission team config is unavailable")
	}
	var teamID, name, role, missionID, missionName, missionStatus string
	err := db.QueryRowContext(c.ctx, configMissionTeamSQL, id).
		Scan(&teamID, &name, &role, &missionID, &missionName, &missionStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, configNotFound("team", id)
	}
	if err != nil {
		log.Printf("[config-view] mission team query: %v", err)
		return nil, configUnavailable("mission team query failed")
	}
	meta := map[string]any{
		"id": teamID, "name": name, "mission_id": missionID, "mission_name": missionName,
		"mission_status": missionStatus, "origin": "mission:" + missionID,
	}
	if runtime := c.findMissionRuntime(missionID, name); runtime != nil {
		meta["runtime_id"] = runtime.ID
		return &configDocument{doc: c.teamManifestDoc(runtime, meta, "runtime:soma"), source: configSourceRuntime}, nil
	}
	agents, err := loadConfigMissionMembers(c, id)
	if err != nil {
		log.Printf("[config-view] mission members query: %v", err)
		return nil, configUnavailable("mission team members query failed")
	}
	manifest := &swarm.TeamManifest{ID: teamID, Name: name, Description: role, Members: agents}
	return &configDocument{doc: c.teamManifestDoc(manifest, meta, "database:service_manifests"), source: configSourceDatabase}, nil
}

func loadConfigMissionMembers(c *deploymentCollector, teamID string) ([]protocol.AgentManifest, error) {
	rows, err := c.s.getDB().QueryContext(c.ctx, configMissionMembersSQL, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []protocol.AgentManifest
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		// Decoding into AgentManifest keeps declared fields only; any extra
		// keys stored in the row never reach the document.
		var agent protocol.AgentManifest
		if err := json.Unmarshal(raw, &agent); err != nil {
			continue
		}
		out = append(out, agent)
	}
	return out, rows.Err()
}

// teamManifestDoc projects a TeamManifest with the resolved provider per
// member. Only provider and model ids are included; provider blocks
// (endpoints, keys) are never read into the document.
func (c *deploymentCollector) teamManifestDoc(tm *swarm.TeamManifest, meta map[string]any, specOrigin string) map[string]any {
	members := make([]any, 0, len(tm.Members))
	for _, agent := range tm.Members {
		resolved := c.deploymentAgentMember(agent)
		member := map[string]any{
			"id":             agent.ID,
			"role":           agent.Role,
			"system_prompt":  agent.SystemPrompt,
			"model":          agent.Model,
			"provider":       agent.Provider,
			"inputs":         stringsOrEmpty(agent.Inputs),
			"outputs":        stringsOrEmpty(agent.Outputs),
			"tools":          stringsOrEmpty(agent.Tools),
			"max_iterations": agent.EffectiveMaxIterations(),
			"resolved": map[string]any{
				"profile":         resolved.Profile,
				"provider":        resolved.Provider,
				"model":           resolved.Model,
				"provider_origin": resolved.ProviderOrigin,
				"provider_status": resolved.ProviderStatus,
			},
		}
		if agent.Verification != nil {
			member["verification"] = agent.Verification
		}
		members = append(members, member)
	}
	spec := map[string]any{
		"id":          tm.ID,
		"name":        tm.Name,
		"type":        string(tm.Type),
		"description": tm.Description,
		"provider":    tm.Provider,
		"members":     members,
		"inputs":      stringsOrEmpty(tm.Inputs),
		"deliveries":  stringsOrEmpty(tm.Deliveries),
	}
	if len(tm.AskRouting) > 0 {
		spec["ask_routing"] = tm.AskRouting
	}
	if tm.Schedule != nil {
		spec["schedule"] = tm.Schedule
	}
	doc := configDoc("TeamManifest", meta, spec)
	doc["x_origin"] = map[string]any{"spec": specOrigin, "members.resolved": "cognitive:router"}
	return doc
}

func stringsOrEmpty(values []string) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, v)
	}
	return out
}
