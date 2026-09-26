package cognitive

import (
	"fmt"
	"log"
	"strings"
)

// validateRootProvider fails Core startup closed when RootProvider is set but
// unusable, instead of letting an unknown or disabled root silently leave
// every unbound profile unresolved. An unset RootProvider is not validated:
// today's behavior (a profile with no explicit binding stays unbound) is
// preserved exactly.
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

// applyRootProviderDefaults binds every defaultExecutionProfiles entry that
// has no explicit Profiles[profile] binding to config.RootProvider. It never
// overwrites an explicit binding: one set in cognitive.yaml, the DB
// system_config overlay, or MYCELIS_PROFILE_<NAME>_PROVIDER always wins,
// because applyEnvOverrides and loadFromDB both run before this and already
// populated Profiles for every profile with an explicit source. A blank
// RootProvider is a no-op, so an operator who never sets root_provider /
// MYCELIS_ROOT_PROVIDER sees no behavior change at all.
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
		if strings.TrimSpace(config.Profiles[profile]) != "" {
			continue
		}
		config.Profiles[profile] = rootID
		bound = append(bound, profile)
	}
	if len(bound) > 0 {
		log.Printf("DEBUG: Applied root_provider %q to unbound profiles: %v", rootID, bound)
	}
}
