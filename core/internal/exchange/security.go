package exchange

import (
	"context"
	"slices"
	"strings"
)

type actorContextKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, normalizeActor(actor))
}

func ActorFromContext(ctx context.Context) Actor {
	actor, _ := ctx.Value(actorContextKey{}).(Actor)
	return normalizeActor(actor)
}

func normalizeActor(actor Actor) Actor {
	actor.Role = normalizeRole(actor.Role)
	actor.Team = strings.TrimSpace(strings.ToLower(actor.Team))
	if actor.Scopes == nil {
		actor.Scopes = []string{}
	}
	return actor
}

func normalizeRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	switch {
	case role == "":
		return ""
	case strings.Contains(role, ":"):
		return strings.SplitN(role, ":", 2)[0]
	default:
		return role
	}
}

func (a Actor) IsAdmin() bool {
	if a.Role == "admin" {
		return true
	}
	return slices.Contains(a.Scopes, "*")
}

func (a Actor) HasRole(roles []string) bool {
	if a.IsAdmin() {
		return true
	}
	for _, role := range roles {
		if normalizeRole(role) == a.Role {
			return true
		}
	}
	return false
}

func defaultActorFromCreatedBy(createdBy string) Actor {
	return normalizeActor(Actor{Role: createdBy})
}

func defaultActorRoleForAgent(agent string) string {
	label := strings.TrimSpace(strings.ToLower(agent))
	switch {
	case label == "":
		return ""
	case strings.Contains(label, "soma"):
		return "soma"
	case strings.Contains(label, "lead"):
		return "team_lead"
	case strings.Contains(label, "review"):
		return "review"
	case strings.Contains(label, "automation"):
		return "automation"
	case strings.Contains(label, "mcp"):
		return "mcp"
	default:
		return "specialist"
	}
}

func uniqueRoles(values ...[]string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, group := range values {
		for _, raw := range group {
			role := normalizeRole(raw)
			if role == "" {
				continue
			}
			if _, ok := seen[role]; ok {
				continue
			}
			seen[role] = struct{}{}
			out = append(out, role)
		}
	}
	return out
}

func channelReaders(channel *Channel) []string {
	roles := []string{}
	for _, participant := range channel.Participants {
		if participant.CanRead {
			roles = append(roles, participant.Role)
		}
	}
	return uniqueRoles(roles)
}

func channelWriters(channel *Channel) []string {
	roles := []string{}
	for _, participant := range channel.Participants {
		if participant.CanWrite {
			roles = append(roles, participant.Role)
		}
	}
	return uniqueRoles(roles)
}

func canReadChannel(actor Actor, channel *Channel) bool {
	if actor.IsAdmin() {
		return true
	}
	if channel == nil {
		return false
	}
	switch channel.SensitivityClass {
	case "admin_only":
		return false
	case "org_visible":
		return actor.Role != ""
	default:
		return actor.HasRole(uniqueRoles(channelReaders(channel), channel.Reviewers))
	}
}

func canWriteChannel(actor Actor, channel *Channel) bool {
	if actor.IsAdmin() {
		return true
	}
	if channel == nil {
		return false
	}
	return actor.HasRole(channelWriters(channel))
}

func canReviewChannel(actor Actor, channel *Channel) bool {
	if actor.IsAdmin() {
		return true
	}
	if channel == nil {
		return false
	}
	return actor.HasRole(channel.Reviewers)
}

func canAccessThread(actor Actor, channel *Channel, thread *Thread) bool {
	if actor.IsAdmin() {
		return true
	}
	if thread == nil || !canReadChannel(actor, channel) {
		return false
	}
	if actor.HasRole(thread.Participants) {
		return true
	}
	return actor.HasRole(uniqueRoles(thread.AllowedReviewers, channel.Reviewers))
}

func canPublishToThread(actor Actor, channel *Channel, thread *Thread) bool {
	if actor.IsAdmin() {
		return true
	}
	if thread == nil || !canWriteChannel(actor, channel) {
		return false
	}
	if actor.HasRole(thread.Participants) {
		return true
	}
	return actor.HasRole(uniqueRoles(thread.AllowedReviewers, thread.EscalationRights, channel.Reviewers))
}

func canReadItem(actor Actor, channel *Channel, item *ExchangeItem) bool {
	if actor.IsAdmin() {
		return true
	}
	if item == nil || !canReadChannel(actor, channel) {
		return false
	}
	switch item.SensitivityClass {
	case "admin_only":
		return false
	case "org_visible":
		return actor.Role != ""
	}
	if actor.HasRole(item.AllowedConsumers) {
		return true
	}
	if item.TargetRole != "" && normalizeRole(item.TargetRole) == actor.Role {
		return true
	}
	if item.SourceRole != "" && normalizeRole(item.SourceRole) == actor.Role {
		return true
	}
	return false
}

func canUseCapability(actor Actor, capability CapabilityDefinition) bool {
	if actor.IsAdmin() {
		return true
	}
	return actor.HasRole(capability.DefaultAllowedRoles)
}

func defaultCapabilityForSchema(schemaID string) string {
	schema, ok := SchemaByID(schemaID)
	if !ok || len(schema.RequiredCapabilities) == 0 {
		return ""
	}
	return schema.RequiredCapabilities[0]
}

func classifyTrustBoundary(sourceRole string, capability CapabilityDefinition) string {
	role := normalizeRole(sourceRole)
	switch {
	case role == "mcp":
		return "bounded_external"
	case role == "api", role == "provider":
		return "bounded_external"
	case role == "remote_node", capability.Source == "node":
		return "restricted"
	case capability.Source == "mcp" || capability.Source == "api":
		return "bounded_external"
	case capability.Source == "node":
		return "restricted"
	case capability.Source == "unavailable":
		return "unavailable"
	default:
		return "trusted_internal"
	}
}

func defaultSensitivityForChannel(channel *Channel) string {
	if channel == nil || channel.SensitivityClass == "" {
		return "role_scoped"
	}
	return channel.SensitivityClass
}

func defaultAllowedConsumers(channel *Channel, targetRole string, capability CapabilityDefinition) []string {
	roles := [][]string{
		channelReaders(channel),
		channel.Reviewers,
		capability.DefaultAllowedRoles,
	}
	if normalizeRole(targetRole) != "" {
		roles = append(roles, []string{normalizeRole(targetRole)})
	}
	return uniqueRoles(roles...)
}
