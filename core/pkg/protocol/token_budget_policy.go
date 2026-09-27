package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

// ConfigDocumentKindTokenBudgetPolicy holds token budgets per agent execution.
// The built-in defaults document is seeded from
// core/config/documents/templates/token-budget-defaults.yaml; operator
// overrides live in one operator-scope document written only by
// /api/v1/cognitive/budgets/overrides (root admin + cognitive:write).
const ConfigDocumentKindTokenBudgetPolicy ConfigDocumentKind = "TokenBudgetPolicy"

const (
	TokenBudgetDefaultsDocumentID  = "token-budget-defaults"
	TokenBudgetOverridesDocumentID = "token-budget-overrides"

	TokenBudgetClassLocalLarge     = "local_large"
	TokenBudgetClassLocalSmall     = "local_small"
	TokenBudgetClassHostedStandard = "hosted_standard"
	TokenBudgetClassHostedPremium  = "hosted_premium"

	TokenBudgetScopeExecution = "execution"
	TokenBudgetScopeRun       = "run"
	TokenBudgetScopeTeamDay   = "team_day"
	TokenBudgetScopeAgentDay  = "agent_day"

	TokenBudgetPeriodExecution    = "execution"
	TokenBudgetPeriodRun          = "run"
	TokenBudgetPeriodUTCDay       = "utc_day"
	TokenBudgetPeriodSinceRestart = "since_restart"

	// TokenBudgetMinLimit and TokenBudgetMaxLimit bound every configured
	// limit. There is no "unlimited" value.
	TokenBudgetMinLimit = 1024
	TokenBudgetMaxLimit = 5000000
)

// TokenBudgetClasses lists the model classes in display order.
var TokenBudgetClasses = []string{TokenBudgetClassLocalLarge, TokenBudgetClassLocalSmall, TokenBudgetClassHostedStandard, TokenBudgetClassHostedPremium}

var tokenBudgetRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

// TokenBudgetLimits is one level of limits. Zero means "not set at this
// level" in overrides; defaults (global and classes) set every field.
type TokenBudgetLimits struct {
	PerExecution int `json:"per_execution,omitempty"`
	PerRun       int `json:"per_run,omitempty"`
	PerTeamDay   int `json:"per_team_day,omitempty"`
	PerAgentDay  int `json:"per_agent_day,omitempty"`
	WarnPct      int `json:"warn_pct,omitempty"`
}

type TokenBudgetOverrides struct {
	Agent   map[string]TokenBudgetLimits `json:"agent,omitempty"`
	Team    map[string]TokenBudgetLimits `json:"team,omitempty"`
	Profile map[string]TokenBudgetLimits `json:"profile,omitempty"`
	Class   map[string]TokenBudgetLimits `json:"class,omitempty"`
}

type TokenBudgetPolicySpec struct {
	Global    TokenBudgetLimits            `json:"global"`
	Classes   map[string]TokenBudgetLimits `json:"classes,omitempty"`
	Overrides TokenBudgetOverrides         `json:"overrides,omitempty"`
}

// DefaultTokenBudgetPolicySpec is the owner-approved D3 table. The shipped
// seed file must equal it (pinned by a configdocuments test).
func DefaultTokenBudgetPolicySpec() TokenBudgetPolicySpec {
	limits := func(execution, run, teamDay, agentDay int) TokenBudgetLimits {
		return TokenBudgetLimits{PerExecution: execution, PerRun: run, PerTeamDay: teamDay, PerAgentDay: agentDay, WarnPct: 80}
	}
	return TokenBudgetPolicySpec{
		Global: limits(32000, 128000, 500000, 250000),
		Classes: map[string]TokenBudgetLimits{
			TokenBudgetClassLocalLarge:     limits(64000, 256000, 2000000, 1000000),
			TokenBudgetClassLocalSmall:     limits(24000, 96000, 1000000, 500000),
			TokenBudgetClassHostedStandard: limits(32000, 128000, 300000, 150000),
			TokenBudgetClassHostedPremium:  limits(24000, 96000, 150000, 75000),
		},
	}
}

func IsTokenBudgetClass(class string) bool {
	for _, known := range TokenBudgetClasses {
		if class == known {
			return true
		}
	}
	return false
}

// DecodeTokenBudgetPolicySpec strictly decodes a spec; unknown fields fail.
func DecodeTokenBudgetPolicySpec(raw json.RawMessage) (TokenBudgetPolicySpec, error) {
	var spec TokenBudgetPolicySpec
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&spec); err != nil {
		return TokenBudgetPolicySpec{}, fmt.Errorf("decode token budget policy spec: %w", err)
	}
	return spec, nil
}

// ValidateTokenBudgetPolicySpec returns every structural issue of a spec.
func ValidateTokenBudgetPolicySpec(raw json.RawMessage) []ConfigDocumentValidationIssue {
	spec, err := DecodeTokenBudgetPolicySpec(raw)
	if err != nil {
		return []ConfigDocumentValidationIssue{{Code: "spec.invalid_token_budget", Field: "spec", Message: err.Error()}}
	}
	return spec.Issues()
}

// Issues validates bounds, classes, refs and the per_execution <= per_run <=
// per_team_day ordering of every complete level (global and classes).
func (spec TokenBudgetPolicySpec) Issues() []ConfigDocumentValidationIssue {
	var issues []ConfigDocumentValidationIssue
	add := func(field, message string) {
		issues = append(issues, ConfigDocumentValidationIssue{Code: "spec.invalid_token_budget", Field: field, Message: message})
	}
	checkComplete := func(field string, limits TokenBudgetLimits) {
		for _, value := range []int{limits.PerExecution, limits.PerRun, limits.PerTeamDay, limits.PerAgentDay, limits.WarnPct} {
			if value == 0 {
				add(field, "every default limit and warn_pct must be set")
				return
			}
		}
		issues = append(issues, limitIssues(field, limits)...)
		if msg := TokenBudgetOrderingIssue(limits); msg != "" {
			add(field, msg)
		}
	}
	checkComplete("spec.global", spec.Global)
	for _, class := range sortedBudgetKeys(spec.Classes) {
		if !IsTokenBudgetClass(class) {
			add("spec.classes."+class, "unknown budget class")
			continue
		}
		checkComplete("spec.classes."+class, spec.Classes[class])
	}
	levels := []struct {
		name    string
		entries map[string]TokenBudgetLimits
	}{{"agent", spec.Overrides.Agent}, {"team", spec.Overrides.Team}, {"profile", spec.Overrides.Profile}, {"class", spec.Overrides.Class}}
	for _, level := range levels {
		for _, ref := range sortedBudgetKeys(level.entries) {
			field := "spec.overrides." + level.name + "." + ref
			if !tokenBudgetRefPattern.MatchString(ref) || (level.name == "class" && !IsTokenBudgetClass(ref)) {
				add(field, "override ref is not a valid "+level.name)
				continue
			}
			if level.entries[ref] == (TokenBudgetLimits{}) {
				add(field, "an override must set at least one field")
				continue
			}
			issues = append(issues, limitIssues(field, level.entries[ref])...)
		}
	}
	return issues
}

// ValidTokenBudgetRef reports whether ref is a safe override reference.
func ValidTokenBudgetRef(ref string) bool { return tokenBudgetRefPattern.MatchString(ref) }

// TokenBudgetOrderingIssue enforces per_execution <= per_run <= per_team_day.
func TokenBudgetOrderingIssue(limits TokenBudgetLimits) string {
	if limits.PerExecution > limits.PerRun || limits.PerRun > limits.PerTeamDay {
		return "limits must satisfy per_execution <= per_run <= per_team_day"
	}
	return ""
}

func limitIssues(field string, limits TokenBudgetLimits) []ConfigDocumentValidationIssue {
	var issues []ConfigDocumentValidationIssue
	named := []struct {
		name  string
		value int
	}{{"per_execution", limits.PerExecution}, {"per_run", limits.PerRun}, {"per_team_day", limits.PerTeamDay}, {"per_agent_day", limits.PerAgentDay}}
	for _, item := range named {
		if item.value != 0 && (item.value < TokenBudgetMinLimit || item.value > TokenBudgetMaxLimit) {
			issues = append(issues, ConfigDocumentValidationIssue{Code: "spec.invalid_token_budget", Field: field + "." + item.name,
				Message: fmt.Sprintf("%s must be an integer in [%d, %d]", item.name, TokenBudgetMinLimit, TokenBudgetMaxLimit)})
		}
	}
	if limits.WarnPct != 0 && (limits.WarnPct < 1 || limits.WarnPct > 99) {
		issues = append(issues, ConfigDocumentValidationIssue{Code: "spec.invalid_token_budget", Field: field + ".warn_pct", Message: "warn_pct must be in [1, 99]"})
	}
	return issues
}

func sortedBudgetKeys(entries map[string]TokenBudgetLimits) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
