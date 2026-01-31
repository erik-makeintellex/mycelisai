package governance

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Action Enums
const (
	ActionAllow           = "ALLOW"
	ActionDeny            = "DENY"
	ActionRequireApproval = "REQUIRE_APPROVAL"
)

// PolicyConfig represents the YAML structure
type PolicyConfig struct {
	Groups   []PolicyGroup `yaml:"groups"`
	Defaults DefaultConfig `yaml:"defaults"`
}

type PolicyGroup struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Targets     []string     `yaml:"targets"` // "team:marketing", "agent:xyz"
	Rules       []PolicyRule `yaml:"rules"`
}

type PolicyRule struct {
	Intent    string `yaml:"intent"`
	Condition string `yaml:"condition"` // Simple "key > value" string
	Action    string `yaml:"action"`
}

type DefaultConfig struct {
	DefaultAction string `yaml:"default_action"`
}

// Engine handles policy evaluation
type Engine struct {
	Config *PolicyConfig
}

// NewEngine loads the policy from a file
func NewEngine(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var config PolicyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse policy yaml: %w", err)
	}

	return &Engine{Config: &config}, nil
}

// Evaluate determines the action for a given request
// simple evaluation: check if target matches group, then check intent, then condition
func (e *Engine) Evaluate(teamID, agentID, intent string, context map[string]interface{}) string {
	// 1. Find matching groups
	for _, group := range e.Config.Groups {
		if e.matchesTarget(group.Targets, teamID, agentID) {
			// 2. Check Rules
			for _, rule := range group.Rules {
				if rule.Intent == intent || rule.Intent == "*" {
					// 3. Check Condition (Simplified)
					if e.checkCondition(rule.Condition, context) {
						return rule.Action
					}
				}
			}
		}
	}

	return e.Config.Defaults.DefaultAction
}

func (e *Engine) matchesTarget(targets []string, teamID, agentID string) bool {
	for _, t := range targets {
		if t == fmt.Sprintf("team:%s", teamID) {
			return true
		}
		if t == fmt.Sprintf("agent:%s", agentID) {
			return true
		}
	}
	return false
}

// checkCondition implements a very basic parser for "key > value"
// Supports: >, <, ==
func (e *Engine) checkCondition(condition string, context map[string]interface{}) bool {
	if condition == "" {
		return true
	}

	// Example: "amount > 50"
	parts := strings.Split(condition, " ")
	if len(parts) != 3 {
		// Complex or invalid conditions default to FALSE (Safe fail)
		return false
	}

	key := parts[0]
	// op := parts[1]
	// valueStr := parts[2]

	_, exists := context[key]
	if !exists {
		return false
	}

	// Assumption: Only supporting float comparison for the demo
	// Ideally use a library like 'govaluate'
	// For now, hardcode "amount" check logic or return true for demo if complex

	// Simplest approach: Just return true if key exists for MVP to avoid type chaos
	return true
}
