package bootstrap

import (
	"fmt"
	"strings"
)

const (
	TemplateLoopTypeReview   = "review"
	TemplateLoopTypeLearning = "learning"

	TemplateLoopOwnerTeam      = "team"
	TemplateLoopOwnerAgentType = "agent_type"
)

// TemplateLoopProfile declares a bundle-defined Review or Learning Loop
// profile that is validated against the bundle's teams/members and seeded
// into organizations created from the bundle.
type TemplateLoopProfile struct {
	ID              string   `yaml:"id" json:"id"`
	Name            string   `yaml:"name" json:"name"`
	Type            string   `yaml:"type" json:"type"`
	Description     string   `yaml:"description,omitempty" json:"description,omitempty"`
	OwnerType       string   `yaml:"owner_type" json:"owner_type"`
	OwnerID         string   `yaml:"owner_id" json:"owner_id"`
	IntervalSeconds int      `yaml:"interval_seconds,omitempty" json:"interval_seconds,omitempty"`
	EventTriggers   []string `yaml:"event_triggers,omitempty" json:"event_triggers,omitempty"`
}

func (b *TemplateBundle) validateLoops() error {
	if len(b.Loops) == 0 {
		return nil
	}

	teamIDs := make(map[string]struct{}, len(b.Teams))
	memberIDs := make(map[string]struct{})
	for _, team := range b.Teams {
		if team == nil {
			continue
		}
		teamIDs[team.ID] = struct{}{}
		for _, member := range team.Members {
			memberIDs[member.ID] = struct{}{}
		}
	}

	seen := make(map[string]struct{}, len(b.Loops))
	for idx := range b.Loops {
		loop := &b.Loops[idx]
		loop.ID = strings.TrimSpace(loop.ID)
		loop.Name = strings.TrimSpace(loop.Name)
		loop.Type = strings.TrimSpace(loop.Type)
		loop.Description = strings.TrimSpace(loop.Description)
		loop.OwnerType = strings.TrimSpace(loop.OwnerType)
		loop.OwnerID = strings.TrimSpace(loop.OwnerID)

		if loop.ID == "" {
			return fmt.Errorf("loops[%d]: missing id", idx)
		}
		if _, exists := seen[loop.ID]; exists {
			return fmt.Errorf("loops[%d]: duplicate loop id %q", idx, loop.ID)
		}
		seen[loop.ID] = struct{}{}

		if loop.Name == "" {
			return fmt.Errorf("loops[%d] (%s): missing name", idx, loop.ID)
		}
		if loop.Type != TemplateLoopTypeReview && loop.Type != TemplateLoopTypeLearning {
			return fmt.Errorf("loops[%d] (%s): type must be %q or %q, got %q", idx, loop.ID, TemplateLoopTypeReview, TemplateLoopTypeLearning, loop.Type)
		}
		if loop.OwnerType != TemplateLoopOwnerTeam && loop.OwnerType != TemplateLoopOwnerAgentType {
			return fmt.Errorf("loops[%d] (%s): owner_type must be %q or %q, got %q", idx, loop.ID, TemplateLoopOwnerTeam, TemplateLoopOwnerAgentType, loop.OwnerType)
		}
		if loop.OwnerID == "" {
			return fmt.Errorf("loops[%d] (%s): missing owner_id", idx, loop.ID)
		}
		if loop.OwnerType == TemplateLoopOwnerTeam {
			if _, ok := teamIDs[loop.OwnerID]; !ok {
				return fmt.Errorf("loops[%d] (%s): owner_id %q not found in bundle teams", idx, loop.ID, loop.OwnerID)
			}
		} else {
			if _, ok := memberIDs[loop.OwnerID]; !ok {
				return fmt.Errorf("loops[%d] (%s): owner_id %q not found in bundle team members", idx, loop.ID, loop.OwnerID)
			}
		}
		if loop.IntervalSeconds < 0 {
			return fmt.Errorf("loops[%d] (%s): interval_seconds must be >= 0, got %d", idx, loop.ID, loop.IntervalSeconds)
		}

		if len(loop.EventTriggers) > 0 {
			triggers := make([]string, 0, len(loop.EventTriggers))
			for tIdx, trigger := range loop.EventTriggers {
				trimmed := strings.TrimSpace(trigger)
				if trimmed == "" {
					return fmt.Errorf("loops[%d] (%s): event_triggers[%d] must not be empty", idx, loop.ID, tIdx)
				}
				triggers = append(triggers, trimmed)
			}
			loop.EventTriggers = triggers
		}
	}

	return nil
}
