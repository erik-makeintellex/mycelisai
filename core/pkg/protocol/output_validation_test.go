package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOutputValidationPlanValidateInteractiveBrowserProbe(t *testing.T) {
	plan := &OutputValidationPlan{
		Kind:     OutputValidationInteractiveBrowser,
		Required: true,
		Checks:   []OutputValidationCheck{OutputValidationCheckLoad, OutputValidationCheckNoPageErrors},
		Probe: &OutputValidationProbe{
			Action: OutputValidationAction{
				Kind: OutputValidationActionKeyHold, Key: "ArrowRight", DurationMS: 500,
			},
			Observe: OutputValidationObservation{
				Kind: OutputValidationObserveVisualChange, Target: "canvas",
			},
		},
	}

	if err := plan.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestOutputValidationPlanRejectsIncompleteProbe(t *testing.T) {
	tests := []struct {
		name string
		plan *OutputValidationPlan
		want string
	}{
		{name: "missing probe", plan: &OutputValidationPlan{
			Kind: OutputValidationInteractiveBrowser, Required: true,
		}, want: "requires a probe"},
		{name: "click target", plan: &OutputValidationPlan{
			Kind: OutputValidationInteractiveBrowser, Required: true,
			Probe: &OutputValidationProbe{
				Action:  OutputValidationAction{Kind: OutputValidationActionClick},
				Observe: OutputValidationObservation{Kind: OutputValidationObserveURLChange},
			},
		}, want: "requires a target"},
		{name: "key hold duration", plan: &OutputValidationPlan{
			Kind: OutputValidationInteractiveBrowser, Required: true,
			Probe: &OutputValidationProbe{
				Action:  OutputValidationAction{Kind: OutputValidationActionKeyHold, Key: "ArrowRight"},
				Observe: OutputValidationObservation{Kind: OutputValidationObserveURLChange},
			},
		}, want: "positive duration_ms"},
		{name: "observation target", plan: &OutputValidationPlan{
			Kind: OutputValidationInteractiveBrowser, Required: true,
			Probe: &OutputValidationProbe{
				Action:  OutputValidationAction{Kind: OutputValidationActionKeyPress, Key: "Enter"},
				Observe: OutputValidationObservation{Kind: OutputValidationObserveTextChange},
			},
		}, want: "requires a target"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestWorkOutputContractMarshalsOwnedValidationPlan(t *testing.T) {
	contract := WorkOutputContract{
		Shape: "app_package",
		OutputValidation: &OutputValidationPlan{
			Kind: OutputValidationInteractiveBrowser, Required: true,
			Checks: []OutputValidationCheck{OutputValidationCheckLoad},
			Probe: &OutputValidationProbe{
				Action:  OutputValidationAction{Kind: OutputValidationActionClick, Target: "button"},
				Observe: OutputValidationObservation{Kind: OutputValidationObserveURLChange},
			},
		},
	}

	payload, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"output_validation", "interactive_browser", "click", "url_change"} {
		if !strings.Contains(string(payload), want) {
			t.Fatalf("payload = %s, missing %q", payload, want)
		}
	}
}

func TestDecodeOutputValidationPlanRestoresJSONContractMap(t *testing.T) {
	raw := map[string]any{
		"kind": " INTERACTIVE_BROWSER ", "required": true,
		"checks": []any{" LOAD ", "load", "NO_PAGE_ERRORS"},
		"probe": map[string]any{
			"action":  map[string]any{"kind": " CLICK ", "target": " button "},
			"observe": map[string]any{"kind": " URL_CHANGE "},
		},
	}

	plan, err := DecodeOutputValidationPlan(raw)
	if err != nil {
		t.Fatalf("DecodeOutputValidationPlan() error = %v", err)
	}
	if plan.Kind != OutputValidationInteractiveBrowser || len(plan.Checks) != 2 {
		t.Fatalf("decoded plan = %#v, want normalized and deduplicated", plan)
	}
	if plan.Probe.Action.Target != "button" || plan.Probe.Observe.Kind != OutputValidationObserveURLChange {
		t.Fatalf("decoded probe = %#v, want normalized fields", plan.Probe)
	}
}
