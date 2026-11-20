package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// deploymentAgentMember builds one agent member. The provider and model come
// from the cognitive router's execution resolution (the same path the agent
// uses at inference time: profile = manifest.Model or "chat", explicit
// provider = manifest.Provider), not from the declared `model` field.
func (c *deploymentCollector) deploymentAgentMember(agent protocol.AgentManifest) protocol.DeploymentMember {
	member := protocol.DeploymentMember{
		ID:     agent.ID,
		Role:   agent.Role,
		Status: "offline",
		Tools:  deploymentTools(agent.Tools, "internal"),
	}
	profile := strings.TrimSpace(agent.Model)
	if profile == "" {
		profile = "chat"
	}
	member.Profile = profile
	c.resolveMemberProvider(&member, profile, strings.TrimSpace(agent.Provider))
	if hb := c.heartbeats[agent.ID]; hb != nil {
		member.Status = deploymentAgentStatus(hb.Status)
		member.LastHeartbeatAt = deploymentTime(hb.LastHeartbeat)
	}
	return member
}

// resolveMemberProvider fills provider, model, origin, and status. Only ids
// are copied from the router; endpoints and keys are never read here.
func (c *deploymentCollector) resolveMemberProvider(member *protocol.DeploymentMember, profile, explicitProvider string) {
	router := c.s.Cognitive
	if router == nil || router.Config == nil {
		member.Provider = explicitProvider
		member.ProviderOrigin = protocol.ProviderOriginDeclared
		member.ProviderStatus = "unknown"
		return
	}
	avail := router.ExecutionAvailability(profile, explicitProvider)
	member.Provider = avail.ProviderID
	member.Model = avail.ModelID
	switch {
	case avail.FallbackApplied:
		member.ProviderOrigin = protocol.ProviderOriginFallback
	case explicitProvider != "":
		member.ProviderOrigin = protocol.ProviderOriginManifest
	default:
		member.ProviderOrigin = protocol.ProviderOriginProfile
	}
	if avail.Available {
		member.ProviderStatus = "available"
	} else {
		member.ProviderStatus = avail.Code
		if member.ProviderStatus == "" {
			member.ProviderStatus = "unavailable"
		}
	}
}

func deploymentAgentStatus(status state.AgentStatus) string {
	switch status {
	case state.StatusIdle:
		return "idle"
	case state.StatusBusy:
		return "busy"
	case state.StatusError:
		return "error"
	default:
		return "offline"
	}
}

// deploymentTools tags each tool name with its source. Names prefixed with
// "mcp:" are MCP tools; everything else takes defaultSource.
func deploymentTools(names []string, defaultSource string) []protocol.DeploymentTool {
	out := make([]protocol.DeploymentTool, 0, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		source := defaultSource
		if strings.HasPrefix(name, "mcp:") {
			source = "mcp"
		}
		out = append(out, protocol.DeploymentTool{Name: name, Source: source})
	}
	return out
}

// unionMemberTools merges member tools into one de-duplicated record list.
func unionMemberTools(members []protocol.DeploymentMember) []protocol.DeploymentTool {
	seen := map[string]struct{}{}
	out := []protocol.DeploymentTool{}
	for _, m := range members {
		for _, tool := range m.Tools {
			if _, ok := seen[tool.Name]; ok {
				continue
			}
			seen[tool.Name] = struct{}{}
			out = append(out, tool)
		}
	}
	return out
}

// teamLifecycle derives a runtime team state from member heartbeats: any
// errored member -> error, any busy member -> running, otherwise idle.
func teamLifecycle(members []protocol.DeploymentMember) string {
	busy := false
	for _, m := range members {
		switch m.Status {
		case "error":
			return protocol.DeploymentStateError
		case "busy":
			busy = true
		}
	}
	if busy {
		return protocol.DeploymentStateRunning
	}
	return protocol.DeploymentStateIdle
}

func (c *deploymentCollector) teamMembers(tm *swarm.TeamManifest) []protocol.DeploymentMember {
	members := make([]protocol.DeploymentMember, 0, len(tm.Members)+1)
	hasReviewer := false
	for _, agent := range tm.Members {
		if hasIdealReviewerToken(agent.Role) || hasIdealReviewerToken(agent.ID) {
			hasReviewer = true
		}
		members = append(members, c.deploymentAgentMember(agent))
	}
	if !hasReviewer && len(tm.Members) > 0 {
		members = append(members, c.deploymentAgentMember(protocol.AgentManifest{
			ID:   "council-sentry",
			Role: "delivery_reviewer",
		}))
	}
	return members
}

func hasIdealReviewerToken(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, token := range []string{"review", "verifier", "sentry", "qa", "audit", "validator"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func latestMemberHeartbeat(members []protocol.DeploymentMember) *protocol.DeploymentMember {
	var latest *protocol.DeploymentMember
	for i := range members {
		if members[i].LastHeartbeatAt == nil {
			continue
		}
		if latest == nil || members[i].LastHeartbeatAt.After(*latest.LastHeartbeatAt) {
			latest = &members[i]
		}
	}
	return latest
}

func agentCountSentence(n int, doing string) string {
	noun := "agents"
	if n == 1 {
		noun = "agent"
	}
	return fmt.Sprintf("%d %s %s", n, noun, doing)
}
