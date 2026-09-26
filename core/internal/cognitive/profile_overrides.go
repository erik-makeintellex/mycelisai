package cognitive

import (
	"os"
	"regexp"
	"strings"
)

// Override origins, reported per profile as override_origin by the
// cognitive status API when source=override.
const (
	// ProfileOriginEnv is a non-empty MYCELIS_PROFILE_<NAME>_PROVIDER.
	ProfileOriginEnv = "env"
	// ProfileOriginDB is a system_config role.<name> row.
	ProfileOriginDB = "db"
	// ProfileOriginRuntime is an in-memory override (mission-profile
	// activation or a config built in code); it does not survive a restart.
	ProfileOriginRuntime = "runtime"
)

// Validation codes for a profile override request, in addition to the
// Execution* resolution codes.
const (
	OverrideInvalidProfile   = "invalid_profile"
	OverrideUnknownProfile   = "unknown_profile"
	OverrideEnvPinned        = "profile_env_pinned"
	OverrideBoundaryMismatch = "fallback_boundary_mismatch"
)

var profileNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// ProfileOverrideError is a rejected profile override request.
type ProfileOverrideError struct {
	Profile string
	Code    string
	Message string
}

func (e *ProfileOverrideError) Error() string { return e.Message }

// ProfileOverrideState is one profile's current binding and override
// provenance, read under the router lock.
type ProfileOverrideState struct {
	ProviderID   string
	Source       string
	Origin       string
	DBRowPresent bool
}

// ProfileOverrideClear is the outcome of ClearProfileOverride.
type ProfileOverrideClear struct {
	Changed                 bool
	Effective               ProfileBinding
	RemainingOverrideOrigin string
}

// IsExecutionProfile reports whether name is one of the seven execution
// profiles that root_provider binds.
func IsExecutionProfile(name string) bool {
	for _, profile := range defaultExecutionProfiles {
		if profile == name {
			return true
		}
	}
	return false
}

// ValidProfileName reports whether name is a well-formed profile name.
func ValidProfileName(name string) bool { return profileNamePattern.MatchString(name) }

func recordProfileOverride(config *BrainConfig, profile, providerID, origin string) {
	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}
	config.Profiles[profile] = providerID
	markProfileSource(config, profile, ProfileSourceOverride)
	setProfileOrigin(config, profile, origin)
	if origin == ProfileOriginDB {
		if config.dbProfileRows == nil {
			config.dbProfileRows = make(map[string]string)
		}
		config.dbProfileRows[profile] = providerID
	}
}

func setProfileOrigin(config *BrainConfig, profile, origin string) {
	if config.ProfileOverrideOrigins == nil {
		config.ProfileOverrideOrigins = make(map[string]string)
	}
	config.ProfileOverrideOrigins[profile] = origin
}

// recordEnvProfileOverrides tags every profile pinned by a non-empty
// MYCELIS_PROFILE_<NAME>_PROVIDER with ProfileOriginEnv. It runs right after
// applyEnvOverrides and mirrors its key resolution; it never changes a
// binding. A DB row under the same profile stays recorded in dbProfileRows.
func recordEnvProfileOverrides(config *BrainConfig) {
	for _, env := range os.Environ() {
		key, value, ok := strings.Cut(env, "=")
		if !ok || strings.TrimSpace(value) == "" {
			continue
		}
		rawField, ok := strings.CutPrefix(key, "MYCELIS_PROFILE_")
		if !ok {
			continue
		}
		rawProfile, ok := strings.CutSuffix(rawField, "_PROVIDER")
		if !ok || rawProfile == "" {
			continue
		}
		setProfileOrigin(config, resolveProfileKey(rawProfile, config.Profiles), ProfileOriginEnv)
	}
}

// recomputeProfileBinding rebinds profile from the non-DB precedence chain:
// an env override, then root_provider (execution profiles only), then the
// cognitive.yaml default, then unbound. NewRouter and ClearProfileOverride
// both use it, so a reset equals what a restart would produce.
func recomputeProfileBinding(config *BrainConfig, profile string) {
	if config.ProfileOverrideOrigins[profile] == ProfileOriginEnv {
		return
	}
	delete(config.ProfileOverrideOrigins, profile)
	if rootID := strings.TrimSpace(config.RootProvider); rootID != "" && IsExecutionProfile(profile) {
		config.Profiles[profile] = rootID
		markProfileSource(config, profile, ProfileSourceRoot)
		return
	}
	if def, ok := config.yamlProfileDefaults[profile]; ok {
		config.Profiles[profile] = def
		markProfileSource(config, profile, ProfileSourceDefault)
		return
	}
	delete(config.Profiles, profile)
	delete(config.ProfileSources, profile)
}

// KnownProfile reports whether profile is a well-formed execution profile
// or a profile currently present in the routing table.
func (r *Router) KnownProfile(profile string) bool {
	if r == nil || r.Config == nil || !ValidProfileName(profile) {
		return false
	}
	if IsExecutionProfile(profile) {
		return true
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.Config.Profiles[profile]
	return ok
}

// ProfileOverrideState returns profile's binding and override provenance.
func (r *Router) ProfileOverrideState(profile string) ProfileOverrideState {
	if r == nil || r.Config == nil {
		return ProfileOverrideState{Source: ProfileSourceUnbound}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.profileOverrideStateLocked(profile)
}

func (r *Router) profileOverrideStateLocked(profile string) ProfileOverrideState {
	_, dbRow := r.Config.dbProfileRows[profile]
	return ProfileOverrideState{
		ProviderID:   strings.TrimSpace(r.Config.Profiles[profile]),
		Source:       r.Config.ProfileSource(profile),
		Origin:       r.Config.ProfileOverrideOrigins[profile],
		DBRowPresent: dbRow,
	}
}

// ValidateProfileOverrides checks every profile->provider pair without
// changing anything. It returns the first rejection.
func (r *Router) ValidateProfileOverrides(overrides map[string]string) error {
	if r == nil || r.Config == nil {
		return &ProfileOverrideError{Code: ExecutionRouterUnavailable, Message: "cognitive router is offline"}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for profile, providerID := range overrides {
		if err := r.validateProfileOverrideLocked(profile, providerID); err != nil {
			return err
		}
	}
	return nil
}

// SetProfileOverrides validates and pins every pair under one write lock,
// all or nothing. origin is ProfileOriginDB after a committed role.* write,
// or ProfileOriginRuntime for an in-memory activation.
func (r *Router) SetProfileOverrides(overrides map[string]string, origin string) error {
	if r == nil || r.Config == nil {
		return &ProfileOverrideError{Code: ExecutionRouterUnavailable, Message: "cognitive router is offline"}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for profile, providerID := range overrides {
		if err := r.validateProfileOverrideLocked(profile, providerID); err != nil {
			return err
		}
	}
	for profile, providerID := range overrides {
		recordProfileOverride(r.Config, profile, strings.TrimSpace(providerID), origin)
	}
	return nil
}

// SetProfileOverride validates and pins one profile under the router lock.
func (r *Router) SetProfileOverride(profile, providerID, origin string) error {
	return r.SetProfileOverrides(map[string]string{profile: providerID}, origin)
}

// ClearProfileOverride drops profile's DB row mirror and any DB or runtime
// override, then recomputes the binding exactly as NewRouter would (env,
// root, YAML default, unbound, then the explicit same-boundary fallback). An
// env override is never cleared here; it is reported as remaining.
func (r *Router) ClearProfileOverride(profile string) ProfileOverrideClear {
	var out ProfileOverrideClear
	if r == nil || r.Config == nil {
		return out
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg := r.Config
	if _, ok := cfg.dbProfileRows[profile]; ok {
		delete(cfg.dbProfileRows, profile)
		out.Changed = true
	}
	origin := cfg.ProfileOverrideOrigins[profile]
	if origin != ProfileOriginEnv && (origin != "" || cfg.ProfileSources[profile] == ProfileSourceOverride) {
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]string)
		}
		recomputeProfileBinding(cfg, profile)
		if IsExecutionProfile(profile) {
			r.applyExplicitFallbackLocked(profile)
		}
		out.Changed = true
	}
	if cfg.ProfileOverrideOrigins[profile] == ProfileOriginEnv {
		out.RemainingOverrideOrigin = ProfileOriginEnv
	}
	providerID := strings.TrimSpace(cfg.Profiles[profile])
	out.Effective = ProfileBinding{
		ProviderID: providerID,
		ModelID:    strings.TrimSpace(cfg.Providers[providerID].ModelID),
		Source:     cfg.ProfileSource(profile),
	}
	return out
}

func (r *Router) validateProfileOverrideLocked(profile, providerID string) error {
	reject := func(code, message string) error {
		return &ProfileOverrideError{Profile: profile, Code: code, Message: message}
	}
	if !ValidProfileName(profile) {
		return reject(OverrideInvalidProfile, "profile name must match ^[a-z][a-z0-9_]{0,31}$")
	}
	if _, bound := r.Config.Profiles[profile]; !bound && !IsExecutionProfile(profile) {
		return reject(OverrideUnknownProfile, "profile "+profile+" is not a known profile")
	}
	if r.Config.ProfileOverrideOrigins[profile] == ProfileOriginEnv {
		return reject(OverrideEnvPinned, "profile "+profile+" is pinned by MYCELIS_PROFILE_"+strings.ToUpper(profile)+"_PROVIDER; change it in .env.compose and recreate Core")
	}
	providerID = strings.TrimSpace(providerID)
	provider, ok := r.Config.Providers[providerID]
	switch {
	case providerID == "" || !ok:
		return reject(ExecutionProviderMissing, "provider "+providerID+" is not configured")
	case !provider.Enabled:
		return reject(ExecutionProviderDisabled, "provider "+providerID+" is disabled")
	case strings.TrimSpace(provider.ModelID) == "":
		return reject(ExecutionModelMissing, "provider "+providerID+" has no model configured")
	case r.Adapters == nil || r.Adapters[providerID] == nil:
		return reject(ExecutionProviderOffline, "provider "+providerID+" is not initialized at runtime")
	}
	boundary := normalizedDataBoundary(provider.DataBoundary)
	for _, candidateID := range r.Config.ProfileFallbacks[profile] {
		candidate, ok := r.Config.Providers[strings.TrimSpace(candidateID)]
		if ok && normalizedDataBoundary(candidate.DataBoundary) != boundary {
			return reject(OverrideBoundaryMismatch, "provider "+providerID+" has data_boundary "+boundary+" but profile_fallbacks for "+profile+" list "+strings.TrimSpace(candidateID)+" on another boundary")
		}
	}
	return nil
}

// ProviderSnapshot returns a copy of one provider's config under the lock.
func (r *Router) ProviderSnapshot(providerID string) (ProviderConfig, bool) {
	if r == nil || r.Config == nil {
		return ProviderConfig{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.Config.Providers[providerID]
	return provider, ok
}

// StoreProviderConfig replaces one provider's config under the write lock.
func (r *Router) StoreProviderConfig(providerID string, provider ProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Config.Providers == nil {
		r.Config.Providers = make(map[string]ProviderConfig)
	}
	r.Config.Providers[providerID] = provider
}

// ConfigSnapshot returns a copy of the routing config, safe to serialize
// while other requests change profiles.
func (r *Router) ConfigSnapshot() *BrainConfig {
	if r == nil || r.Config == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := r.Config
	return &BrainConfig{
		Providers:              copyMap(c.Providers),
		Profiles:               copyMap(c.Profiles),
		ProfileFallbacks:       c.ProfileFallbacks,
		Media:                  c.Media,
		RootProvider:           c.RootProvider,
		ProfileSources:         copyMap(c.ProfileSources),
		ProfileOverrideOrigins: copyMap(c.ProfileOverrideOrigins),
		OverlayError:           c.OverlayError,
	}
}

func copyMap[V any](in map[string]V) map[string]V {
	if in == nil {
		return nil
	}
	out := make(map[string]V, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
