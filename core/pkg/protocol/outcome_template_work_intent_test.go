package protocol

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompileOutcomeTemplateMapsResolvedBriefToWorkIntent(t *testing.T) {
	template := completeOutcomeTemplate()
	template.ID = " template-id "
	template.Version = " v4 "
	template.Digest = " sha256:fixed "
	template.IntentKind = " Project "
	template.OutputContract = &WorkOutputContract{
		Retention:  " User_Deliverable ",
		Validation: []string{"template validation", "template validation"},
	}
	operator := completeMinimumBrief("operator")
	operator.EssentialBehavior = []string{" behavior one ", "", "behavior one", "behavior two"}
	operator.Constraints = []string{" constraint ", "constraint"}
	operator.AcceptanceEvidence = []string{" proof one ", "", "proof two"}

	result, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{
		Template: template, OperatorValues: operator,
	})
	if err != nil {
		t.Fatalf("CompileOutcomeTemplate() error = %v", err)
	}
	wantBrief := operator
	wantBrief.EssentialBehavior = []string{"behavior one", "behavior two"}
	wantBrief.Constraints = []string{"constraint"}
	wantBrief.AcceptanceEvidence = []string{"proof one", "proof two"}
	if !reflect.DeepEqual(result.Brief, wantBrief) {
		t.Fatalf("normalized brief = %#v, want %#v", result.Brief, wantBrief)
	}
	intent := result.WorkIntent
	if intent == nil {
		t.Fatal("expected WorkIntent")
	}
	if intent.Kind != "project" || intent.Objective != operator.TargetOutcome {
		t.Fatalf("WorkIntent identity = %#v", intent)
	}
	if intent.OutputContract == nil || intent.OutputContract.Shape != "operator form" {
		t.Fatalf("output shape = %#v", intent.OutputContract)
	}
	if intent.OutputContract.PrimaryDeliverable != "operator form" || intent.OutputContract.Retention != "user_deliverable" {
		t.Fatalf("output contract mapping = %#v", intent.OutputContract)
	}
	wantValidation := []string{"template validation", "proof one", "proof two"}
	if !reflect.DeepEqual(intent.OutputContract.Validation, wantValidation) {
		t.Fatalf("output validation = %#v, want %#v", intent.OutputContract.Validation, wantValidation)
	}
	if intent.MinimumSufficientBrief == nil || !reflect.DeepEqual(*intent.MinimumSufficientBrief, wantBrief) {
		t.Fatalf("WorkIntent brief = %#v, want %#v", intent.MinimumSufficientBrief, wantBrief)
	}
	wantSnapshot := OutcomeTemplateSnapshot{ID: "template-id", Version: "v4", Digest: "sha256:fixed"}
	if intent.OutcomeTemplateSnapshot == nil || *intent.OutcomeTemplateSnapshot != wantSnapshot {
		t.Fatalf("WorkIntent snapshot = %#v, want %#v", intent.OutcomeTemplateSnapshot, wantSnapshot)
	}
	if result.TemplateSnapshot != wantSnapshot {
		t.Fatalf("result snapshot = %#v, want %#v", result.TemplateSnapshot, wantSnapshot)
	}
}

func TestCompileOutcomeTemplateRejectsInvalidDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		template OutcomeTemplate
		want     string
	}{
		{name: "missing identity", template: OutcomeTemplate{}, want: "id, version, and digest"},
		{
			name:     "negative question limit",
			template: OutcomeTemplate{ID: "id", Version: "1", Digest: "digest", QuestionLimit: -1},
			want:     "cannot be negative",
		},
		{
			name: "unknown field",
			template: OutcomeTemplate{ID: "id", Version: "1", Digest: "digest", FieldRules: []OutcomeTemplateFieldRule{
				{Field: "unknown", UnresolvedAs: BriefMustAsk},
			}},
			want: "unsupported minimum brief field",
		},
		{
			name: "unknown classification",
			template: OutcomeTemplate{ID: "id", Version: "1", Digest: "digest", FieldRules: []OutcomeTemplateFieldRule{
				{Field: BriefFieldAudience, UnresolvedAs: "later"},
			}},
			want: "unsupported unresolved classification",
		},
		{
			name: "duplicate field rule",
			template: OutcomeTemplate{ID: "id", Version: "1", Digest: "digest", FieldRules: []OutcomeTemplateFieldRule{
				{Field: BriefFieldAudience, UnresolvedAs: BriefCanInfer},
				{Field: BriefFieldAudience, UnresolvedAs: BriefCanDefer},
			}},
			want: "duplicate minimum brief field rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{Template: tt.template})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestCompileOutcomeTemplateCopiesApprovedState(t *testing.T) {
	template := completeOutcomeTemplate()
	template.OutputContract = &WorkOutputContract{Validation: []string{"template validation"}}
	operator := completeMinimumBrief("operator")

	result, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{
		Template: template, OperatorValues: operator,
	})
	if err != nil {
		t.Fatalf("CompileOutcomeTemplate() error = %v", err)
	}
	template.ID = "changed"
	template.Version = "changed"
	template.Digest = "changed"
	template.OutputContract.Validation[0] = "changed"
	operator.EssentialBehavior[0] = "changed"
	operator.Constraints[0] = "changed"
	operator.AcceptanceEvidence[0] = "changed"

	if result.TemplateSnapshot != (OutcomeTemplateSnapshot{ID: "template-id", Version: "1.2.3", Digest: "sha256:template"}) {
		t.Fatalf("snapshot changed with source template: %#v", result.TemplateSnapshot)
	}
	if result.WorkIntent.OutcomeTemplateSnapshot.ID != "template-id" {
		t.Fatalf("WorkIntent snapshot changed with source template: %#v", result.WorkIntent.OutcomeTemplateSnapshot)
	}
	if result.Brief.EssentialBehavior[0] != "operator behavior" || result.Brief.Constraints[0] != "operator constraint" {
		t.Fatalf("brief changed with source values: %#v", result.Brief)
	}
	if result.WorkIntent.OutputContract.Validation[0] != "template validation" {
		t.Fatalf("output contract changed with source template: %#v", result.WorkIntent.OutputContract)
	}

	result.Brief.EssentialBehavior[0] = "result changed"
	if result.WorkIntent.MinimumSufficientBrief.EssentialBehavior[0] != "operator behavior" {
		t.Fatalf("WorkIntent brief aliases compile result: %#v", result.WorkIntent.MinimumSufficientBrief)
	}
}
