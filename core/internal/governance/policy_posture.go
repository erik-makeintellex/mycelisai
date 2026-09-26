package governance

import (
	"fmt"
	"regexp"
	"strings"
)

// PostureTargetPrefix marks a policy group that applies to work shaped by an
// Outcome Template posture (posture:<outcome-template-id>). Posture groups are
// stricter-only: they can require approval, never allow or deny. They cannot
// match NATS team:/agent: targets, so Evaluate is unaffected by them.
const PostureTargetPrefix = "posture:"

// ValidatePolicyConfig is the single check shared by startup load and the
// admin PUT. Every action (default and rule) must be exactly ALLOW, DENY, or
// REQUIRE_APPROVAL. A policy with no restricting rule and an ALLOW default
// (including an empty file) is an allow-everything policy and is rejected, so
// such a file starts Core degraded. Posture groups also keep their contract.
func ValidatePolicyConfig(cfg *PolicyConfig) error {
	if cfg == nil {
		return fmt.Errorf("policy config is required")
	}
	if !isKnownAction(cfg.Defaults.DefaultAction) {
		return fmt.Errorf("defaults.default_action must be %s, %s, or %s, got %q", ActionAllow, ActionDeny, ActionRequireApproval, cfg.Defaults.DefaultAction)
	}
	restricting := false
	for index, group := range cfg.Groups {
		for ruleIndex, rule := range group.Rules {
			if !isKnownAction(rule.Action) {
				return fmt.Errorf("policy group %d (%q) rule %d: action must be %s, %s, or %s, got %q", index, group.Name, ruleIndex, ActionAllow, ActionDeny, ActionRequireApproval, rule.Action)
			}
			if !isPostureGroup(group) && rule.Action != ActionAllow {
				restricting = true
			}
		}
	}
	if cfg.Defaults.DefaultAction == ActionAllow && !restricting {
		return fmt.Errorf("an ALLOW default needs at least one DENY or REQUIRE_APPROVAL rule; an empty or allow-only policy is not accepted")
	}
	for index, group := range cfg.Groups {
		if !isPostureGroup(group) {
			continue
		}
		label := fmt.Sprintf("policy group %d (%q)", index, group.Name)
		for _, target := range group.Targets {
			id, ok := strings.CutPrefix(target, PostureTargetPrefix)
			if !ok || strings.TrimSpace(id) == "" || id != strings.TrimSpace(id) {
				return fmt.Errorf("%s: posture groups may only target posture:<outcome-template-id>, got %q", label, target)
			}
		}
		if len(group.Rules) == 0 {
			return fmt.Errorf("%s: posture groups need at least one rule", label)
		}
		for ruleIndex, rule := range group.Rules {
			if rule.Action != ActionRequireApproval {
				return fmt.Errorf("%s rule %d: posture rules may only use %s, got %q", label, ruleIndex, ActionRequireApproval, rule.Action)
			}
			if strings.TrimSpace(rule.Condition) != "" {
				return fmt.Errorf("%s rule %d: posture rules must not carry a condition", label, ruleIndex)
			}
			if _, err := compilePostureIntent(rule.Intent); err != nil {
				return fmt.Errorf("%s rule %d: %w", label, ruleIndex, err)
			}
		}
	}
	return nil
}

func isPostureGroup(group PolicyGroup) bool {
	for _, target := range group.Targets {
		if strings.HasPrefix(target, PostureTargetPrefix) {
			return true
		}
	}
	return false
}

func compilePostureIntent(intent string) (*regexp.Regexp, error) {
	if !strings.HasPrefix(intent, "^") || !strings.HasSuffix(intent, "$") {
		return nil, fmt.Errorf("posture intent %q must be an anchored regex (^...$)", intent)
	}
	pattern, err := regexp.Compile(intent)
	if err != nil {
		return nil, fmt.Errorf("posture intent %q: %w", intent, err)
	}
	return pattern, nil
}

// PostureRequiresApproval reports whether any posture group for postureID
// matches a planned tool name or capability id. With nothing planned, rules
// are evaluated against the empty string, so ^.*$ still requires approval.
// A rule that no longer compiles fails closed.
func (e *Engine) PostureRequiresApproval(postureID string, tools, capabilityIDs []string) (bool, string) {
	postureID = strings.TrimSpace(postureID)
	if postureID == "" {
		return false, ""
	}
	if e == nil || e.Config == nil {
		return true, ""
	}
	candidates := append(append([]string(nil), tools...), capabilityIDs...)
	if len(candidates) == 0 {
		candidates = []string{""}
	}
	target := PostureTargetPrefix + postureID
	for _, group := range e.Config.Groups {
		if !isPostureGroup(group) || !containsString(group.Targets, target) {
			continue
		}
		for _, rule := range group.Rules {
			pattern, err := compilePostureIntent(rule.Intent)
			if err != nil {
				return true, group.Name
			}
			for _, candidate := range candidates {
				if pattern.MatchString(candidate) {
					return true, group.Name
				}
			}
		}
	}
	return false, ""
}

// PostureRequiresApproval evaluates a snapshot of the live policy; a nil or
// degraded guard requires approval for any posture-scoped work.
func (g *Guard) PostureRequiresApproval(postureID string, tools, capabilityIDs []string) (bool, string) {
	return (&Engine{Config: g.policySnapshot()}).PostureRequiresApproval(postureID, tools, capabilityIDs)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func isKnownAction(action string) bool {
	return action == ActionAllow || action == ActionDeny || action == ActionRequireApproval
}
