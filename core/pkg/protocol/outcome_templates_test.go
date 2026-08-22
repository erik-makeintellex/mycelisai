package protocol

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompileOutcomeTemplateValuePrecedence(t *testing.T) {
	tests := []struct {
		name     string
		operator MinimumSufficientBrief
		policy   MinimumSufficientBrief
		want     MinimumSufficientBrief
		sources  map[BriefField]BriefValueSource
	}{
		{
			name:    "template defaults",
			want:    completeMinimumBrief("template"),
			sources: allBriefSources(BriefValueTemplateDefault),
		},
		{
			name:     "operator replaces template",
			operator: completeMinimumBrief("operator"),
			want:     completeMinimumBrief("operator"),
			sources:  allBriefSources(BriefValueOperator),
		},
		{
			name:     "policy replaces operator and template",
			operator: completeMinimumBrief("operator"),
			policy:   completeMinimumBrief("policy"),
			want:     completeMinimumBrief("policy"),
			sources:  allBriefSources(BriefValuePolicy),
		},
		{
			name: "layers resolve independently",
			operator: MinimumSufficientBrief{
				TargetOutcome: "operator outcome",
				DeliveryForm:  "operator form",
			},
			policy: MinimumSufficientBrief{
				Audience:    "policy audience",
				Constraints: []string{"policy constraint"},
			},
			want: MinimumSufficientBrief{
				TargetOutcome:      "operator outcome",
				Audience:           "policy audience",
				EssentialBehavior:  []string{"template behavior"},
				QualityBar:         "template quality",
				DeliveryForm:       "operator form",
				Constraints:        []string{"policy constraint"},
				AcceptanceEvidence: []string{"template evidence"},
			},
			sources: map[BriefField]BriefValueSource{
				BriefFieldTargetOutcome:      BriefValueOperator,
				BriefFieldAudience:           BriefValuePolicy,
				BriefFieldEssentialBehavior:  BriefValueTemplateDefault,
				BriefFieldQualityBar:         BriefValueTemplateDefault,
				BriefFieldDeliveryForm:       BriefValueOperator,
				BriefFieldConstraints:        BriefValuePolicy,
				BriefFieldAcceptanceEvidence: BriefValueTemplateDefault,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := completeOutcomeTemplate()
			result, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{
				Template: template, OperatorValues: tt.operator, PolicyValues: tt.policy,
			})
			if err != nil {
				t.Fatalf("CompileOutcomeTemplate() error = %v", err)
			}
			if !reflect.DeepEqual(result.Brief, tt.want) {
				t.Fatalf("brief = %#v, want %#v", result.Brief, tt.want)
			}
			if !reflect.DeepEqual(result.ValueSources, tt.sources) {
				t.Fatalf("sources = %#v, want %#v", result.ValueSources, tt.sources)
			}
			if !result.Ready || result.WorkIntent == nil {
				t.Fatalf("complete brief should compile a ready WorkIntent: %#v", result)
			}
		})
	}
}

func TestCompileOutcomeTemplateQuestionBoundsAndImpact(t *testing.T) {
	tests := []struct {
		name           string
		template       OutcomeTemplate
		operator       MinimumSufficientBrief
		wantFields     []BriefField
		wantUnresolved int
	}{
		{
			name:           "defaults to three high impact questions",
			template:       templateIdentity(),
			wantFields:     []BriefField{BriefFieldTargetOutcome, BriefFieldAcceptanceEvidence, BriefFieldDeliveryForm},
			wantUnresolved: 7,
		},
		{
			name: "absolute maximum is four",
			template: OutcomeTemplate{
				ID: "bounded", Version: "1", Digest: "sha256:bounded", QuestionLimit: 99,
				Defaults: MinimumSufficientBrief{AcceptanceEvidence: []string{"accepted"}},
				FieldRules: []OutcomeTemplateFieldRule{
					{Field: BriefFieldAudience, UnresolvedAs: BriefMustAsk, Question: "Audience?", Impact: 250},
					{Field: BriefFieldEssentialBehavior, UnresolvedAs: BriefMustAsk, Question: "Behavior?", Impact: 240},
					{Field: BriefFieldQualityBar, UnresolvedAs: BriefMustAsk, Question: "Quality?", Impact: 230},
					{Field: BriefFieldConstraints, UnresolvedAs: BriefMustAsk, Question: "Constraints?", Impact: 220},
				},
			},
			wantFields:     []BriefField{BriefFieldAudience, BriefFieldEssentialBehavior, BriefFieldQualityBar, BriefFieldConstraints},
			wantUnresolved: 6,
		},
		{
			name: "template may request a smaller turn",
			template: OutcomeTemplate{
				ID: "small", Version: "1", Digest: "sha256:small", QuestionLimit: 2,
			},
			wantFields:     []BriefField{BriefFieldTargetOutcome, BriefFieldAcceptanceEvidence},
			wantUnresolved: 7,
		},
		{
			name:     "resolved must ask fields are not questioned",
			template: templateIdentity(),
			operator: MinimumSufficientBrief{
				TargetOutcome: "Outcome", AcceptanceEvidence: []string{"Evidence"},
			},
			wantFields:     []BriefField{BriefFieldDeliveryForm},
			wantUnresolved: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{
				Template: tt.template, OperatorValues: tt.operator,
			})
			if err != nil {
				t.Fatalf("CompileOutcomeTemplate() error = %v", err)
			}
			if len(result.Questions) > maximumOutcomeTemplateQuestionLimit {
				t.Fatalf("questions exceeded absolute maximum: %#v", result.Questions)
			}
			gotFields := make([]BriefField, 0, len(result.Questions))
			for _, question := range result.Questions {
				gotFields = append(gotFields, question.Field)
				if strings.TrimSpace(question.Prompt) == "" {
					t.Fatalf("question for %s has no prompt", question.Field)
				}
				if unresolvedClass(result.Unresolved, question.Field) != BriefMustAsk {
					t.Fatalf("question field %s is not classified must_ask", question.Field)
				}
			}
			if !reflect.DeepEqual(gotFields, tt.wantFields) {
				t.Fatalf("question fields = %#v, want %#v", gotFields, tt.wantFields)
			}
			if len(result.Unresolved) != tt.wantUnresolved {
				t.Fatalf("unresolved count = %d, want %d: %#v", len(result.Unresolved), tt.wantUnresolved, result.Unresolved)
			}
			if result.Ready || result.WorkIntent != nil {
				t.Fatalf("must_ask fields must prevent a ready WorkIntent: %#v", result)
			}
		})
	}
}

func TestCompileOutcomeTemplateClassifiesNonBlockingFields(t *testing.T) {
	tests := []struct {
		name        string
		rules       []OutcomeTemplateFieldRule
		wantClasses map[BriefField]BriefResolutionClass
	}{
		{
			name: "canonical defaults",
			wantClasses: map[BriefField]BriefResolutionClass{
				BriefFieldAudience:          BriefCanInfer,
				BriefFieldEssentialBehavior: BriefCanInfer,
				BriefFieldQualityBar:        BriefCanInfer,
				BriefFieldConstraints:       BriefCanDefer,
			},
		},
		{
			name: "template overrides classification",
			rules: []OutcomeTemplateFieldRule{
				{Field: BriefFieldAudience, UnresolvedAs: BriefCanDefer},
				{Field: BriefFieldEssentialBehavior, UnresolvedAs: BriefCanDefer},
				{Field: BriefFieldQualityBar, UnresolvedAs: BriefCanInfer},
				{Field: BriefFieldConstraints, UnresolvedAs: BriefCanInfer},
			},
			wantClasses: map[BriefField]BriefResolutionClass{
				BriefFieldAudience:          BriefCanDefer,
				BriefFieldEssentialBehavior: BriefCanDefer,
				BriefFieldQualityBar:        BriefCanInfer,
				BriefFieldConstraints:       BriefCanInfer,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := templateIdentity()
			template.Defaults = MinimumSufficientBrief{
				TargetOutcome: "Resolved outcome", DeliveryForm: "Document",
				AcceptanceEvidence: []string{"Reviewed"},
			}
			template.FieldRules = tt.rules
			result, err := CompileOutcomeTemplate(OutcomeTemplateCompileInput{Template: template})
			if err != nil {
				t.Fatalf("CompileOutcomeTemplate() error = %v", err)
			}
			if !result.Ready || result.WorkIntent == nil || len(result.Questions) != 0 {
				t.Fatalf("only can_infer/can_defer fields should be ready: %#v", result)
			}
			for field, want := range tt.wantClasses {
				if got := unresolvedClass(result.Unresolved, field); got != want {
					t.Fatalf("classification for %s = %q, want %q", field, got, want)
				}
			}
		})
	}
}

func templateIdentity() OutcomeTemplate {
	return OutcomeTemplate{ID: "template-id", Version: "1.2.3", Digest: "sha256:template"}
}

func completeOutcomeTemplate() OutcomeTemplate {
	template := templateIdentity()
	template.Defaults = completeMinimumBrief("template")
	return template
}

func completeMinimumBrief(prefix string) MinimumSufficientBrief {
	return MinimumSufficientBrief{
		TargetOutcome:      prefix + " outcome",
		Audience:           prefix + " audience",
		EssentialBehavior:  []string{prefix + " behavior"},
		QualityBar:         prefix + " quality",
		DeliveryForm:       prefix + " form",
		Constraints:        []string{prefix + " constraint"},
		AcceptanceEvidence: []string{prefix + " evidence"},
	}
}

func allBriefSources(source BriefValueSource) map[BriefField]BriefValueSource {
	result := make(map[BriefField]BriefValueSource, len(minimumBriefFieldOrder))
	for _, field := range minimumBriefFieldOrder {
		result[field] = source
	}
	return result
}

func unresolvedClass(fields []UnresolvedBriefField, target BriefField) BriefResolutionClass {
	for _, field := range fields {
		if field.Field == target {
			return field.Classification
		}
	}
	return ""
}
