package swarm

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/catalogue"
)

func (r *InternalToolRegistry) hydrateCreateTeamProfiles(ctx context.Context, args map[string]any) error {
	if r.catalogue == nil {
		return nil
	}
	switch rawAgents := args["agents"].(type) {
	case []any:
		for _, raw := range rawAgents {
			member, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if err := r.hydrateRuntimeProfile(ctx, member); err != nil {
				return err
			}
		}
	case []map[string]any:
		for _, member := range rawAgents {
			if err := r.hydrateRuntimeProfile(ctx, member); err != nil {
				return err
			}
		}
	}
	return r.hydrateRuntimeProfile(ctx, args)
}

func (r *InternalToolRegistry) hydrateRuntimeProfile(ctx context.Context, target map[string]any) error {
	ref := strings.TrimSpace(stringValue(target["profile_ref"]))
	if ref == "" {
		return nil
	}
	profile, err := r.catalogue.Resolve(ctx, ref)
	if err != nil {
		return fmt.Errorf("create_team profile %s: %w", ref, err)
	}
	setProfileDefault(target, "role", profile.Role)
	setProfileDefault(target, "system_prompt", profile.SystemPrompt)
	setProfileDefault(target, "model", profile.Model)
	setProfileDefault(target, "tools", profileTools(profile))
	setProfileDefault(target, "inputs", profile.Inputs)
	setProfileDefault(target, "outputs", profile.Outputs)
	setProfileDefault(target, "context_bindings", profile.ContextBindings)
	setProfileDefault(target, "usage_policy", profile.UsagePolicy)
	return nil
}

func profileTools(profile *catalogue.AgentTemplate) []string {
	if len(profile.CapabilityRefs) > 0 {
		return append([]string(nil), profile.CapabilityRefs...)
	}
	return append([]string(nil), profile.Tools...)
}

func setProfileDefault(target map[string]any, key string, value any) {
	if current, exists := target[key]; exists && !profileValueEmpty(current) {
		return
	}
	target[key] = value
}

func profileValueEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	}
	return false
}
