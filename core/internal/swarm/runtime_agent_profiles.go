package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/pkg/protocol"
)

const runtimeProfileResolutionTimeout = 5 * time.Second

// hydrateCreateTeamProfiles is the profile-to-runtime boundary. Catalogue
// entries remain inert until create_team selects them; this function resolves
// only those selected refs and pins their immutable lineage into the manifest
// before Soma starts the team and its provider or NATS lifecycle.
func (r *InternalToolRegistry) hydrateCreateTeamProfiles(ctx context.Context, args map[string]any) error {
	scope, err := runtimeProfileScope(ctx, args["profile_scope"])
	if err != nil {
		return err
	}
	delete(args, "profile_snapshot")
	if manifest, ok := args["manifest"].(map[string]any); ok {
		delete(manifest, "profile_snapshot")
		if err := r.hydrateProfileContainer(ctx, manifest, scope); err != nil {
			return err
		}
	}
	return r.hydrateProfileContainer(ctx, args, scope)
}

func (r *InternalToolRegistry) hydrateProfileContainer(ctx context.Context, args map[string]any, scope catalogue.ProfileResolutionScope) error {
	switch rawAgents := args["agents"].(type) {
	case []any:
		for _, raw := range rawAgents {
			member, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if err := r.hydrateRuntimeProfile(ctx, member, scope); err != nil {
				return err
			}
		}
	case []map[string]any:
		for _, member := range rawAgents {
			if err := r.hydrateRuntimeProfile(ctx, member, scope); err != nil {
				return err
			}
		}
	}
	return r.hydrateRuntimeProfile(ctx, args, scope)
}

func (r *InternalToolRegistry) hydrateRuntimeProfile(ctx context.Context, target map[string]any, scope catalogue.ProfileResolutionScope) error {
	delete(target, "profile_snapshot")
	ref := strings.TrimSpace(stringValue(target["profile_ref"]))
	if ref == "" {
		return nil
	}
	if r.catalogue == nil {
		return fmt.Errorf("create_team profile %s: profile catalogue is unavailable", ref)
	}
	resolveCtx, cancel := context.WithTimeout(ctx, runtimeProfileResolutionTimeout)
	defer cancel()
	profile, err := r.catalogue.ResolveActive(resolveCtx, ref, scope)
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
	setProfileDefault(target, "verification", profileVerification(profile))
	if profile.ProfileSnapshot != nil {
		snapshot := *profile.ProfileSnapshot
		target["profile_snapshot"] = &snapshot
	}
	return nil
}

func runtimeProfileScope(ctx context.Context, raw any) (catalogue.ProfileResolutionScope, error) {
	scope := catalogue.ProfileResolutionScope{TenantID: "default"}
	if meta, ok := ToolInvocationContextFromContext(ctx); ok {
		scope.OperatorRef = firstNonEmptyString(meta.OperatorID, meta.UserLabel)
		scope.WorkspaceRef = strings.TrimSpace(meta.WorkspaceID)
		scope.OrganizationRef = strings.TrimSpace(meta.OrganizationID)
	}
	requested, _ := raw.(map[string]any)
	for _, check := range []struct{ key, trusted string }{
		{"operator_ref", scope.OperatorRef}, {"workspace_ref", scope.WorkspaceRef}, {"organization_ref", scope.OrganizationRef},
	} {
		value := strings.TrimSpace(stringValue(requested[check.key]))
		if value != "" && value != check.trusted {
			return scope, fmt.Errorf("create_team profile scope %s is outside the confirmed request boundary", check.key)
		}
	}
	return scope, nil
}

func profileTools(profile *catalogue.AgentTemplate) []string {
	if len(profile.CapabilityRefs) > 0 {
		return append([]string(nil), profile.CapabilityRefs...)
	}
	return append([]string(nil), profile.Tools...)
}

func profileVerification(profile *catalogue.AgentTemplate) *protocol.Verification {
	if strings.TrimSpace(profile.VerificationStrategy) == "" {
		return nil
	}
	return &protocol.Verification{
		Strategy:          protocol.VerifyStrategy(profile.VerificationStrategy),
		Rubric:            append([]string(nil), profile.VerificationRubric...),
		ValidationCommand: profile.ValidationCommand,
	}
}

func setProfileDefault(target map[string]any, key string, value any) {
	if _, exists := target[key]; exists {
		return
	}
	target[key] = value
}
