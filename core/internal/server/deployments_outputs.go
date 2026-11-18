package server

import (
	"database/sql"
	"log"

	"github.com/mycelis/core/pkg/protocol"
)

// One grouped query covers every deployment: artifacts linked to a team id
// (mission teams, group teams) or, with no team id, to an agent id (standing
// team members). Proof envelopes are not linked to artifacts yet, so
// proof_status stays `not_recorded`.
const deploymentOutputsSQL = `
	SELECT COALESCE(team_id::text, ''), COALESCE(agent_id, ''), COALESCE(status, ''),
	       COUNT(*), MAX(created_at)
	FROM artifacts
	GROUP BY 1, 2, 3`

const deploymentRecentArtifactsSQL = `
	SELECT id::text, COALESCE(team_id::text, ''), COALESCE(agent_id, ''),
	       COALESCE(title, ''), COALESCE(file_path, ''), COALESCE(content_type, ''),
	       COALESCE(file_size_bytes, 0), COALESCE(status, ''), created_at
	FROM artifacts
	ORDER BY created_at DESC
	LIMIT 100`

type deploymentOutputBuckets struct {
	byTeam        map[string]*protocol.DeploymentOutputsSummary
	byAgent       map[string]*protocol.DeploymentOutputsSummary
	recentByTeam  map[string][]protocol.DeploymentOutputItem
	recentByAgent map[string][]protocol.DeploymentOutputItem
}

func newDeploymentOutputsSummary() *protocol.DeploymentOutputsSummary {
	return &protocol.DeploymentOutputsSummary{ByStatus: map[string]int{}, ProofStatus: "not_recorded"}
}

func mergeOutputs(dst, src *protocol.DeploymentOutputsSummary) {
	if src == nil {
		return
	}
	dst.Retained += src.Retained
	for status, n := range src.ByStatus {
		dst.ByStatus[status] += n
	}
	dst.LastOutputAt = laterDeploymentTime(dst.LastOutputAt, src.LastOutputAt)
}

// attachOutputs adds outputs_summary to team and group records. Planned
// departments have none. A failure marks the outputs source degraded and
// leaves outputs_summary unset (unknown) rather than zero.
func (c *deploymentCollector) attachOutputs() {
	needed := false
	for _, rec := range c.records {
		if rec.Kind != protocol.DeploymentKindPlannedDepartment {
			needed = true
			break
		}
	}
	if !needed {
		return
	}
	db := c.s.getDB()
	if db == nil {
		c.addSource(deploymentSourceOutputs, protocol.DeploymentSourceUnavailable, "database not configured; output counts unknown", 0)
		return
	}
	buckets, err := c.loadOutputBuckets(db)
	if err != nil {
		log.Printf("[deployments] outputs query: %v", err)
		c.addSource(deploymentSourceOutputs, protocol.DeploymentSourceDegraded, "output counts unavailable", 0)
		return
	}

	standing := map[string]*protocol.DeploymentOutputsSummary{}
	total := 0
	for i := range c.records {
		rec := &c.records[i]
		summary := newDeploymentOutputsSummary()
		switch rec.Kind {
		case protocol.DeploymentKindStandingTeam:
			mergeOutputs(summary, buckets.byTeam[rec.ID])
			mergeOutputs(summary, buckets.byAgent[rec.ID])
			recent := append([]protocol.DeploymentOutputItem(nil), buckets.recentByTeam[rec.ID]...)
			recent = append(recent, buckets.recentByAgent[rec.ID]...)
			for _, m := range rec.Members {
				mergeOutputs(summary, buckets.byAgent[m.ID])
				recent = append(recent, buckets.recentByAgent[m.ID]...)
			}
			if len(recent) > 6 {
				recent = recent[:6]
			}
			for j := range recent {
				recent[j].ViewURL = "/resources?team=" + rec.ID
			}
			rec.RecentOutputs = recent
			standing[rec.ID] = summary
		case protocol.DeploymentKindMissionTeam:
			mergeOutputs(summary, buckets.byTeam[rec.ID])
			rec.RecentOutputs = buckets.recentByTeam[rec.ID]
		default:
			continue
		}
		rec.OutputsSummary = summary
		rec.LastActivity = laterDeploymentTime(rec.LastActivity, summary.LastOutputAt)
		total += summary.Retained
	}
	for i := range c.records {
		rec := &c.records[i]
		if rec.Kind != protocol.DeploymentKindGroup {
			continue
		}
		summary := newDeploymentOutputsSummary()
		for _, m := range rec.Members {
			if m.Role != "team" {
				continue
			}
			if b, ok := buckets.byTeam[m.ID]; ok {
				mergeOutputs(summary, b)
			} else {
				mergeOutputs(summary, standing[m.ID])
			}
		}
		rec.OutputsSummary = summary
		rec.LastActivity = laterDeploymentTime(rec.LastActivity, summary.LastOutputAt)
	}
	c.addSource(deploymentSourceOutputs, protocol.DeploymentSourceOK, "", total)
}

func (c *deploymentCollector) loadOutputBuckets(db *sql.DB) (*deploymentOutputBuckets, error) {
	rows, err := db.QueryContext(c.ctx, deploymentOutputsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	b := &deploymentOutputBuckets{
		byTeam:        map[string]*protocol.DeploymentOutputsSummary{},
		byAgent:       map[string]*protocol.DeploymentOutputsSummary{},
		recentByTeam:  map[string][]protocol.DeploymentOutputItem{},
		recentByAgent: map[string][]protocol.DeploymentOutputItem{},
	}
	for rows.Next() {
		var teamID, agentID, status string
		var count int
		var last sql.NullTime
		if err := rows.Scan(&teamID, &agentID, &status, &count, &last); err != nil {
			return nil, err
		}
		target := b.byTeam
		key := teamID
		if teamID == "" {
			target, key = b.byAgent, agentID
		}
		if key == "" {
			continue
		}
		summary, ok := target[key]
		if !ok {
			summary = newDeploymentOutputsSummary()
			target[key] = summary
		}
		summary.Retained += count
		if status == "" {
			status = "unknown"
		}
		summary.ByStatus[status] += count
		if last.Valid {
			summary.LastOutputAt = laterDeploymentTime(summary.LastOutputAt, deploymentTime(last.Time))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if rows2, err := db.QueryContext(c.ctx, deploymentRecentArtifactsSQL); err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var id, teamID, agentID, title, filePath, contentType, status string
			var sizeBytes int64
			var createdAt sql.NullTime
			if err := rows2.Scan(&id, &teamID, &agentID, &title, &filePath, &contentType, &sizeBytes, &status, &createdAt); err != nil {
				break
			}
			item := protocol.DeploymentOutputItem{
				ID: id, Title: deploymentFirstNonEmpty(title, filePath, id), FilePath: filePath,
				ContentType: contentType, Bytes: sizeBytes, Status: status,
				DownloadURL: "/api/v1/artifacts/" + id + "/download", ViewURL: "/resources?artifact=" + id,
			}
			if createdAt.Valid {
				item.CreatedAt = deploymentTime(createdAt.Time)
			}
			if teamID != "" {
				if len(b.recentByTeam[teamID]) < 6 {
					b.recentByTeam[teamID] = append(b.recentByTeam[teamID], item)
				}
			} else if agentID != "" {
				if len(b.recentByAgent[agentID]) < 6 {
					b.recentByAgent[agentID] = append(b.recentByAgent[agentID], item)
				}
			}
		}
	}
	return b, nil
}
