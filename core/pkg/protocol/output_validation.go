package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type OutputValidationKind string
type OutputValidationCheck string
type OutputValidationActionKind string
type OutputValidationObservationKind string

const (
	OutputValidationInteractiveBrowser OutputValidationKind = "interactive_browser"

	OutputValidationCheckLoad           OutputValidationCheck = "load"
	OutputValidationCheckNoPageErrors   OutputValidationCheck = "no_page_errors"
	OutputValidationCheckNoFailedAssets OutputValidationCheck = "no_failed_local_assets"

	OutputValidationActionClick    OutputValidationActionKind = "click"
	OutputValidationActionKeyPress OutputValidationActionKind = "key_press"
	OutputValidationActionKeyHold  OutputValidationActionKind = "key_hold"
	OutputValidationActionFill     OutputValidationActionKind = "fill"
	OutputValidationActionPointer  OutputValidationActionKind = "pointer"

	OutputValidationObserveVisualChange   OutputValidationObservationKind = "visual_change"
	OutputValidationObserveTextChange     OutputValidationObservationKind = "text_change"
	OutputValidationObserveValueChange    OutputValidationObservationKind = "value_change"
	OutputValidationObserveElementVisible OutputValidationObservationKind = "element_visible"
	OutputValidationObserveURLChange      OutputValidationObservationKind = "url_change"
)

// OutputValidationPlan is approved with WorkIntent and cannot be redefined by
// the output that it validates.
type OutputValidationPlan struct {
	Kind     OutputValidationKind    `json:"kind"`
	Required bool                    `json:"required"`
	Checks   []OutputValidationCheck `json:"checks,omitempty"`
	Probe    *OutputValidationProbe  `json:"probe,omitempty"`
}

type OutputValidationProbe struct {
	Action  OutputValidationAction      `json:"action"`
	Observe OutputValidationObservation `json:"observe"`
}

type OutputValidationAction struct {
	Kind       OutputValidationActionKind `json:"kind"`
	Target     string                     `json:"target,omitempty"`
	Key        string                     `json:"key,omitempty"`
	Value      string                     `json:"value,omitempty"`
	DurationMS int                        `json:"duration_ms,omitempty"`
}

type OutputValidationObservation struct {
	Kind   OutputValidationObservationKind `json:"kind"`
	Target string                          `json:"target,omitempty"`
}

// NormalizeOutputValidationPlan returns a deep, normalized copy suitable for
// carrying from conversation through an approved WorkIntent.
func NormalizeOutputValidationPlan(raw *OutputValidationPlan) *OutputValidationPlan {
	if raw == nil {
		return nil
	}
	plan := *raw
	plan.Kind = OutputValidationKind(strings.ToLower(strings.TrimSpace(string(plan.Kind))))
	plan.Checks = normalizeValidationChecks(plan.Checks)
	if raw.Probe != nil {
		probe := *raw.Probe
		probe.Action.Kind = OutputValidationActionKind(strings.ToLower(strings.TrimSpace(string(probe.Action.Kind))))
		probe.Action.Target = strings.TrimSpace(probe.Action.Target)
		probe.Action.Key = strings.TrimSpace(probe.Action.Key)
		probe.Action.Value = strings.TrimSpace(probe.Action.Value)
		probe.Observe.Kind = OutputValidationObservationKind(strings.ToLower(strings.TrimSpace(string(probe.Observe.Kind))))
		probe.Observe.Target = strings.TrimSpace(probe.Observe.Target)
		plan.Probe = &probe
	}
	return &plan
}

// DecodeOutputValidationPlan restores a plan from typed or JSON-decoded
// contract data and returns only a normalized, runnable plan.
func DecodeOutputValidationPlan(raw any) (*OutputValidationPlan, error) {
	var plan *OutputValidationPlan
	switch value := raw.(type) {
	case *OutputValidationPlan:
		plan = NormalizeOutputValidationPlan(value)
	case OutputValidationPlan:
		plan = NormalizeOutputValidationPlan(&value)
	case nil:
		return nil, errors.New("output validation plan is required")
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode output validation plan: %w", err)
		}
		var decoded OutputValidationPlan
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, fmt.Errorf("decode output validation plan: %w", err)
		}
		plan = NormalizeOutputValidationPlan(&decoded)
	}
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return plan, nil
}

// Validate rejects required interactive output plans that an adapter cannot run.
func (p *OutputValidationPlan) Validate() error {
	if p == nil {
		return errors.New("output validation plan is required")
	}
	if p.Kind != OutputValidationInteractiveBrowser {
		return fmt.Errorf("unsupported output validation kind %q", p.Kind)
	}
	if !p.Required {
		return errors.New("interactive browser validation must be required")
	}
	if p.Probe == nil {
		return errors.New("interactive browser validation requires a probe")
	}
	if err := p.Probe.Action.validate(); err != nil {
		return err
	}
	if err := p.Probe.Observe.validate(); err != nil {
		return err
	}
	if len(p.Checks) == 0 {
		return errors.New("interactive browser validation requires checks")
	}
	for _, check := range p.Checks {
		if !supportedValidationCheck(check) {
			return fmt.Errorf("unsupported output validation check %q", check)
		}
	}
	return nil
}

func normalizeValidationChecks(raw []OutputValidationCheck) []OutputValidationCheck {
	seen := make(map[OutputValidationCheck]struct{}, len(raw))
	checks := make([]OutputValidationCheck, 0, len(raw))
	for _, value := range raw {
		check := OutputValidationCheck(strings.ToLower(strings.TrimSpace(string(value))))
		if check == "" {
			continue
		}
		if _, ok := seen[check]; ok {
			continue
		}
		seen[check] = struct{}{}
		checks = append(checks, check)
	}
	return checks
}

func supportedValidationCheck(check OutputValidationCheck) bool {
	switch check {
	case OutputValidationCheckLoad, OutputValidationCheckNoPageErrors, OutputValidationCheckNoFailedAssets:
		return true
	default:
		return false
	}
}

func (a OutputValidationAction) validate() error {
	target := strings.TrimSpace(a.Target)
	key := strings.TrimSpace(a.Key)
	switch a.Kind {
	case OutputValidationActionClick, OutputValidationActionPointer:
		if target == "" {
			return fmt.Errorf("%s action requires a target", a.Kind)
		}
	case OutputValidationActionKeyPress:
		if key == "" {
			return errors.New("key_press action requires a key")
		}
	case OutputValidationActionKeyHold:
		if key == "" || a.DurationMS <= 0 {
			return errors.New("key_hold action requires a key and positive duration_ms")
		}
	case OutputValidationActionFill:
		if target == "" {
			return errors.New("fill action requires a target")
		}
	default:
		return fmt.Errorf("unsupported output validation action %q", a.Kind)
	}
	return nil
}

func (o OutputValidationObservation) validate() error {
	switch o.Kind {
	case OutputValidationObserveURLChange:
		return nil
	case OutputValidationObserveVisualChange, OutputValidationObserveTextChange,
		OutputValidationObserveValueChange, OutputValidationObserveElementVisible:
		if strings.TrimSpace(o.Target) == "" {
			return fmt.Errorf("%s observation requires a target", o.Kind)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output validation observation %q", o.Kind)
	}
}
