package bootstrap

import (
	"fmt"
	"strings"
)

// retiredBridgeBundleID is the V8 migration bridge bundle, retired by BOOT-B
// (owner decision 2026-09-29) with no alias: a deployment pinned to it must
// fail closed at startup and choose a replacement explicitly.
const retiredBridgeBundleID = "v8-migration-standing-team-bridge"

func rejectRetiredBundleID(id string) error {
	if !strings.EqualFold(strings.TrimSpace(id), retiredBridgeBundleID) {
		return nil
	}
	return fmt.Errorf("bootstrap template bundle %q was retired with no alias; set MYCELIS_BOOTSTRAP_TEMPLATE_ID=%s (admin-core and council-core, the default when unset) or %s (adds prime-architect, prime-development and agui-design-architect)",
		retiredBridgeBundleID, DefaultStartupBundleID, OptionalDevSwarmBundleID)
}

// resolveBundleExtends merges each extending bundle's base teams ahead of its
// own, so selecting an optional bundle boots the base teams plus the optional
// ones while the base bundle itself stays unchanged. Only one level is
// allowed; a missing base, a chain, or a team ID defined twice fails closed.
func resolveBundleExtends(bundles []*TemplateBundle) error {
	byID := make(map[string]*TemplateBundle, len(bundles))
	for _, bundle := range bundles {
		byID[bundle.ID] = bundle
	}
	for _, bundle := range bundles {
		baseID := strings.TrimSpace(bundle.Extends)
		if baseID == "" {
			continue
		}
		base, ok := byID[baseID]
		if !ok || base == bundle {
			return fmt.Errorf("template bundle %q extends %q, which is not a separate mounted bundle", bundle.ID, baseID)
		}
		if strings.TrimSpace(base.Extends) != "" {
			return fmt.Errorf("template bundle %q extends %q, which itself extends %q; only one level is supported", bundle.ID, baseID, base.Extends)
		}
		merged, err := cloneTeamManifests(base.Teams)
		if err != nil {
			return fmt.Errorf("template bundle %q: clone base %q teams: %w", bundle.ID, baseID, err)
		}
		seen := make(map[string]struct{}, len(merged)+len(bundle.Teams))
		for _, team := range merged {
			seen[team.ID] = struct{}{}
		}
		for _, team := range bundle.Teams {
			if _, exists := seen[team.ID]; exists {
				return fmt.Errorf("template bundle %q redefines team %q from base %q", bundle.ID, team.ID, baseID)
			}
			seen[team.ID] = struct{}{}
		}
		bundle.Teams = append(merged, bundle.Teams...)
		if bundle.ProviderPolicy.IsEmpty() {
			bundle.ProviderPolicy = base.ProviderPolicy.Clone()
		}
		if strings.TrimSpace(bundle.Kernel.Mode) == "" {
			bundle.Kernel = base.Kernel
		}
		if strings.TrimSpace(bundle.Council.Mode) == "" {
			bundle.Council = base.Council
		}
	}
	return nil
}
