package cognitive

import (
	"fmt"
	"log"
	"strings"
)

// Profile binding sources, reported per profile by the cognitive status API.
const (
	// ProfileSourceDefault is a binding shipped in cognitive.yaml profiles.
	ProfileSourceDefault = "default"
	// ProfileSourceRoot is a binding applied from root_provider.
	ProfileSourceRoot = "root"
	// ProfileSourceOverride is an operator override: a non-empty
	// MYCELIS_PROFILE_<NAME>_PROVIDER, a DB system_config role.<name> row,
	// or a runtime profile update.
	ProfileSourceOverride = "override"
	// ProfileSourceFallback is an explicit same-boundary ProfileFallbacks
	// rebind applied at startup.
	ProfileSourceFallback = "fallback"
	// ProfileSourceUnbound is reported for an execution profile with no
	// binding at all.
	ProfileSourceUnbound = "unbound"
)

// ProfileBinding is one profile's effective provider and where it came from.
type ProfileBinding struct {
	ProviderID string `json:"provider_id,omitempty"`
	ModelID    string `json:"model_id,omitempty"`
	Source     string `json:"source"`
}

// validateRootProvider fails Core startup closed when RootProvider is set but
// unusable, instead of letting an unknown or disabled root silently leave
// profiles on their shipped defaults. An unset RootProvider is not validated.
func validateRootProvider(config *BrainConfig) error {
	if config == nil {
		return nil
	}
	rootID := strings.TrimSpace(config.RootProvider)
	if rootID == "" {
		return nil
	}
	provider, ok := config.Providers[rootID]
	if !ok {
		return fmt.Errorf(
			"root_provider %q does not match any configured provider; configure the provider first or fix MYCELIS_ROOT_PROVIDER/root_provider",
			rootID,
		)
	}
	if !provider.Enabled {
		return fmt.Errorf(
			"root_provider %q is configured but disabled; enable it (MYCELIS_PROVIDER_%s_ENABLED=true) or point root_provider/MYCELIS_ROOT_PROVIDER at an enabled provider",
			rootID, strings.ToUpper(strings.ReplaceAll(rootID, "-", "_")),
		)
	}
	return nil
}

// markLoadedProfilesAsDefault tags every binding loaded from cognitive.yaml
// as a shipped default and snapshots it for persistence. It must run before
// the DB overlay and env overrides, which tag their own bindings as
// overrides.
func markLoadedProfilesAsDefault(config *BrainConfig) {
	if config == nil {
		return
	}
	config.yamlProfileDefaults = make(map[string]string, len(config.Profiles))
	for profile, providerID := range config.Profiles {
		if strings.TrimSpace(providerID) == "" {
			continue
		}
		config.yamlProfileDefaults[profile] = providerID
		markProfileSource(config, profile, ProfileSourceDefault)
	}
}

func markProfileSource(config *BrainConfig, profile, source string) {
	if config.ProfileSources == nil {
		config.ProfileSources = make(map[string]string)
	}
	config.ProfileSources[profile] = source
}

// SetProfileOverride binds profile to providerID as a runtime operator
// override on a config that is still being built. It is not safe for
// concurrent use and does not validate: a running Router must use
// Router.SetProfileOverride, which validates and holds the router lock.
func (c *BrainConfig) SetProfileOverride(profile, providerID string) {
	if c == nil {
		return
	}
	recordProfileOverride(c, profile, providerID, ProfileOriginRuntime)
}

// ProfileSource reports where profile's current binding came from. A bound
// profile with no recorded source (a config built in code) is a default.
func (c *BrainConfig) ProfileSource(profile string) string {
	if c == nil || strings.TrimSpace(c.Profiles[profile]) == "" {
		return ProfileSourceUnbound
	}
	if source := c.ProfileSources[profile]; source != "" {
		return source
	}
	return ProfileSourceDefault
}

// EffectiveProfileBindings reports every execution profile plus any other
// bound profile with its effective provider, model, and source.
func (c *BrainConfig) EffectiveProfileBindings() map[string]ProfileBinding {
	out := make(map[string]ProfileBinding)
	if c == nil {
		return out
	}
	add := func(profile string) {
		providerID := strings.TrimSpace(c.Profiles[profile])
		out[profile] = ProfileBinding{
			ProviderID: providerID,
			ModelID:    strings.TrimSpace(c.Providers[providerID].ModelID),
			Source:     c.ProfileSource(profile),
		}
	}
	for _, profile := range defaultExecutionProfiles {
		add(profile)
	}
	for profile := range c.Profiles {
		add(profile)
	}
	return out
}

// applyRootProviderDefaults binds every defaultExecutionProfiles entry to
// config.RootProvider unless an operator override pins it. Root replaces the
// shipped cognitive.yaml defaults; it never replaces an override from
// MYCELIS_PROFILE_<NAME>_PROVIDER or the DB overlay, which applyEnvOverrides
// and loadFromDB tag before this runs. A blank RootProvider is a no-op.
func applyRootProviderDefaults(config *BrainConfig) {
	if config == nil {
		return
	}
	rootID := strings.TrimSpace(config.RootProvider)
	if rootID == "" {
		return
	}
	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}

	var bound []string
	for _, profile := range defaultExecutionProfiles {
		if config.ProfileSource(profile) == ProfileSourceOverride {
			continue
		}
		// The same recompute a governed override reset uses, so a reset
		// and a restart produce the same binding.
		recomputeProfileBinding(config, profile)
		bound = append(bound, profile)
	}
	if len(bound) > 0 {
		log.Printf("INFO: Applied root_provider %q to profiles without an operator override: %v", rootID, bound)
	}
}

// persistableProfiles returns the profile map to write to cognitive.yaml.
// Only shipped defaults are written: a root-, override-, or fallback-derived
// binding is written back as its cognitive.yaml default (or omitted), so
// operator overrides live only in system_config role.* and env, and never
// return from YAML as defaults after a restart.
func (c *BrainConfig) persistableProfiles() map[string]string {
	if c == nil || c.Profiles == nil {
		return nil
	}
	out := make(map[string]string, len(c.Profiles))
	for profile, providerID := range c.Profiles {
		switch c.ProfileSources[profile] {
		case ProfileSourceRoot, ProfileSourceOverride, ProfileSourceFallback:
			if def, ok := c.yamlProfileDefaults[profile]; ok {
				out[profile] = def
			}
			continue
		}
		out[profile] = providerID
	}
	return out
}
