package cognitive

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// TokenBudgetExhaustedCode is the normalized availability/blocker code for an
// honest budget stop. A stop is never an approval request and never changes
// the A2b approver tiers.
const TokenBudgetExhaustedCode = "token_budget_exhausted"

// MinBudgetHeadroom is the smallest remaining allowance worth a provider
// call; below it the call is refused before it is made.
const MinBudgetHeadroom = 256

// Override and default levels, most specific first.
const (
	BudgetLevelAgent        = "agent"
	BudgetLevelTeam         = "team"
	BudgetLevelProfile      = "profile"
	BudgetLevelClass        = "class"
	BudgetLevelClassDefault = "class_default"
	BudgetLevelGlobal       = "global"
)

// BudgetSubject names what one inference is charged to.
type BudgetSubject struct {
	AgentID string
	TeamID  string
	Profile string
	Class   string
}

// ResolveBudgetClass returns an explicit valid budget_class, else derives it
// from the normalized data boundary: local_only -> local_large and
// leaves_org -> hosted_premium (hosted fails toward the tighter class).
func ResolveBudgetClass(cfg ProviderConfig) string {
	if class := strings.TrimSpace(cfg.BudgetClass); protocol.IsTokenBudgetClass(class) {
		return class
	}
	if normalizedDataBoundary(cfg.DataBoundary) == DataBoundaryLeavesOrg {
		return protocol.TokenBudgetClassHostedPremium
	}
	return protocol.TokenBudgetClassLocalLarge
}

// ResolveBudgetLimits resolves each field independently, most specific wins:
// agent -> team -> profile -> class override -> class default -> global. It
// also returns the level each field came from.
func ResolveBudgetLimits(spec protocol.TokenBudgetPolicySpec, subject BudgetSubject) (protocol.TokenBudgetLimits, map[string]string) {
	type level struct {
		name   string
		limits protocol.TokenBudgetLimits
		ok     bool
	}
	pick := func(entries map[string]protocol.TokenBudgetLimits, ref string) (protocol.TokenBudgetLimits, bool) {
		if ref == "" {
			return protocol.TokenBudgetLimits{}, false
		}
		limits, ok := entries[ref]
		return limits, ok
	}
	var levels []level
	add := func(name string, limits protocol.TokenBudgetLimits, ok bool) {
		levels = append(levels, level{name, limits, ok})
	}
	agent, ok := pick(spec.Overrides.Agent, subject.AgentID)
	add(BudgetLevelAgent, agent, ok)
	team, ok := pick(spec.Overrides.Team, subject.TeamID)
	add(BudgetLevelTeam, team, ok)
	profile, ok := pick(spec.Overrides.Profile, subject.Profile)
	add(BudgetLevelProfile, profile, ok)
	classOverride, ok := pick(spec.Overrides.Class, subject.Class)
	add(BudgetLevelClass, classOverride, ok)
	classDefault, ok := pick(spec.Classes, subject.Class)
	add(BudgetLevelClassDefault, classDefault, ok)
	add(BudgetLevelGlobal, spec.Global, true)

	var out protocol.TokenBudgetLimits
	sources := map[string]string{}
	fields := []struct {
		name string
		get  func(protocol.TokenBudgetLimits) int
		set  func(int)
	}{
		{"per_execution", func(l protocol.TokenBudgetLimits) int { return l.PerExecution }, func(v int) { out.PerExecution = v }},
		{"per_run", func(l protocol.TokenBudgetLimits) int { return l.PerRun }, func(v int) { out.PerRun = v }},
		{"per_team_day", func(l protocol.TokenBudgetLimits) int { return l.PerTeamDay }, func(v int) { out.PerTeamDay = v }},
		{"per_agent_day", func(l protocol.TokenBudgetLimits) int { return l.PerAgentDay }, func(v int) { out.PerAgentDay = v }},
		{"warn_pct", func(l protocol.TokenBudgetLimits) int { return l.WarnPct }, func(v int) { out.WarnPct = v }},
	}
	for _, field := range fields {
		for _, lvl := range levels {
			if value := field.get(lvl.limits); lvl.ok && value > 0 {
				field.set(value)
				sources[field.name] = lvl.name
				break
			}
		}
	}
	return out, sources
}
