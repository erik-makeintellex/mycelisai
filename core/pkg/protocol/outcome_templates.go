package protocol

import (
	"fmt"
	"sort"
	"strings"
)

type BriefField string
type BriefResolutionClass string
type BriefValueSource string

const (
	BriefFieldTargetOutcome      BriefField = "target_outcome"
	BriefFieldAudience           BriefField = "audience"
	BriefFieldEssentialBehavior  BriefField = "essential_behavior"
	BriefFieldQualityBar         BriefField = "quality_bar"
	BriefFieldDeliveryForm       BriefField = "delivery_form"
	BriefFieldConstraints        BriefField = "constraints"
	BriefFieldAcceptanceEvidence BriefField = "acceptance_evidence"
)

const (
	BriefMustAsk  BriefResolutionClass = "must_ask"
	BriefCanInfer BriefResolutionClass = "can_infer"
	BriefCanDefer BriefResolutionClass = "can_defer"
)

const (
	BriefValueTemplateDefault BriefValueSource = "template_default"
	BriefValueOperator        BriefValueSource = "operator"
	BriefValuePolicy          BriefValueSource = "policy"
)

const (
	defaultOutcomeTemplateQuestionLimit = 3
	maximumOutcomeTemplateQuestionLimit = 4
)

var minimumBriefFieldOrder = []BriefField{
	BriefFieldTargetOutcome,
	BriefFieldAudience,
	BriefFieldEssentialBehavior,
	BriefFieldQualityBar,
	BriefFieldDeliveryForm,
	BriefFieldConstraints,
	BriefFieldAcceptanceEvidence,
}

// MinimumSufficientBrief is the generic outcome definition carried into
// approved work. Domain-specific validation belongs in declarative templates.
type MinimumSufficientBrief struct {
	TargetOutcome      string   `json:"target_outcome,omitempty"`
	Audience           string   `json:"audience,omitempty"`
	EssentialBehavior  []string `json:"essential_behavior,omitempty"`
	QualityBar         string   `json:"quality_bar,omitempty"`
	DeliveryForm       string   `json:"delivery_form,omitempty"`
	Constraints        []string `json:"constraints,omitempty"`
	AcceptanceEvidence []string `json:"acceptance_evidence,omitempty"`
}

// OutcomeTemplateFieldRule declares how one unresolved brief field is handled.
// Higher Impact values are asked first when clarification is bounded.
type OutcomeTemplateFieldRule struct {
	Field        BriefField           `json:"field"`
	UnresolvedAs BriefResolutionClass `json:"unresolved_as"`
	Question     string               `json:"question,omitempty"`
	Impact       int                  `json:"impact,omitempty"`
}

// OutcomeTemplate is reusable configuration that seeds a brief and an existing
// WorkIntent/output contract. It does not define an execution path.
type OutcomeTemplate struct {
	ID             string                     `json:"id"`
	Version        string                     `json:"version"`
	Digest         string                     `json:"digest"`
	IntentKind     string                     `json:"intent_kind,omitempty"`
	Defaults       MinimumSufficientBrief     `json:"defaults,omitempty"`
	FieldRules     []OutcomeTemplateFieldRule `json:"field_rules,omitempty"`
	QuestionLimit  int                        `json:"question_limit,omitempty"`
	OutputContract *WorkOutputContract        `json:"output_contract,omitempty"`
}

// OutcomeTemplateSnapshot prevents later template edits from redefining the
// configuration identity attached to approved or historical work.
type OutcomeTemplateSnapshot struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type OutcomeTemplateCompileInput struct {
	Template       OutcomeTemplate        `json:"template"`
	OperatorValues MinimumSufficientBrief `json:"operator_values,omitempty"`
	PolicyValues   MinimumSufficientBrief `json:"policy_values,omitempty"`
}

type UnresolvedBriefField struct {
	Field          BriefField           `json:"field"`
	Classification BriefResolutionClass `json:"classification"`
}

type BriefClarificationQuestion struct {
	Field  BriefField `json:"field"`
	Prompt string     `json:"prompt"`
}

type OutcomeTemplateCompileResult struct {
	Brief            MinimumSufficientBrief          `json:"brief"`
	ValueSources     map[BriefField]BriefValueSource `json:"value_sources,omitempty"`
	Unresolved       []UnresolvedBriefField          `json:"unresolved,omitempty"`
	Questions        []BriefClarificationQuestion    `json:"questions,omitempty"`
	Ready            bool                            `json:"ready"`
	TemplateSnapshot OutcomeTemplateSnapshot         `json:"template_snapshot"`
	WorkIntent       *WorkIntent                     `json:"work_intent,omitempty"`
}

// CompileOutcomeTemplate resolves a template into the existing WorkIntent
// contract. Operator values replace template defaults; policy values replace
// both. Only must_ask fields prevent compilation of a ready WorkIntent.
func CompileOutcomeTemplate(input OutcomeTemplateCompileInput) (OutcomeTemplateCompileResult, error) {
	template, rules, err := normalizeOutcomeTemplate(input.Template)
	if err != nil {
		return OutcomeTemplateCompileResult{}, err
	}

	brief, sources := resolveMinimumSufficientBrief(
		template.Defaults,
		input.OperatorValues,
		input.PolicyValues,
	)
	result := OutcomeTemplateCompileResult{
		Brief:        brief,
		ValueSources: sources,
		TemplateSnapshot: OutcomeTemplateSnapshot{
			ID: template.ID, Version: template.Version, Digest: template.Digest,
		},
	}

	questionCandidates := make([]briefQuestionCandidate, 0, len(minimumBriefFieldOrder))
	for index, field := range minimumBriefFieldOrder {
		if minimumBriefFieldResolved(brief, field) {
			continue
		}
		rule := rules[field]
		result.Unresolved = append(result.Unresolved, UnresolvedBriefField{
			Field: field, Classification: rule.UnresolvedAs,
		})
		if rule.UnresolvedAs == BriefMustAsk {
			questionCandidates = append(questionCandidates, briefQuestionCandidate{
				field: field, prompt: rule.Question, impact: rule.Impact, fieldOrder: index,
			})
		}
	}

	sort.SliceStable(questionCandidates, func(i, j int) bool {
		if questionCandidates[i].impact == questionCandidates[j].impact {
			return questionCandidates[i].fieldOrder < questionCandidates[j].fieldOrder
		}
		return questionCandidates[i].impact > questionCandidates[j].impact
	})
	limit := template.QuestionLimit
	if limit == 0 {
		limit = defaultOutcomeTemplateQuestionLimit
	}
	if limit > maximumOutcomeTemplateQuestionLimit {
		limit = maximumOutcomeTemplateQuestionLimit
	}
	if len(questionCandidates) > limit {
		questionCandidates = questionCandidates[:limit]
	}
	for _, candidate := range questionCandidates {
		result.Questions = append(result.Questions, BriefClarificationQuestion{
			Field: candidate.field, Prompt: candidate.prompt,
		})
	}

	result.Ready = len(questionCandidates) == 0 && !hasMustAsk(result.Unresolved)
	if result.Ready {
		result.WorkIntent = compileBriefWorkIntent(template, brief, result.TemplateSnapshot)
	}
	return result, nil
}

type briefQuestionCandidate struct {
	field      BriefField
	prompt     string
	impact     int
	fieldOrder int
}

func normalizeOutcomeTemplate(raw OutcomeTemplate) (OutcomeTemplate, map[BriefField]OutcomeTemplateFieldRule, error) {
	template := raw
	template.ID = strings.TrimSpace(raw.ID)
	template.Version = strings.TrimSpace(raw.Version)
	template.Digest = strings.TrimSpace(raw.Digest)
	template.IntentKind = strings.TrimSpace(raw.IntentKind)
	template.Defaults = normalizeMinimumSufficientBrief(raw.Defaults)
	if template.ID == "" || template.Version == "" || template.Digest == "" {
		return OutcomeTemplate{}, nil, fmt.Errorf("outcome template id, version, and digest are required")
	}
	if template.QuestionLimit < 0 {
		return OutcomeTemplate{}, nil, fmt.Errorf("outcome template question limit cannot be negative")
	}

	rules := defaultOutcomeTemplateFieldRules()
	seen := make(map[BriefField]struct{}, len(raw.FieldRules))
	for _, rawRule := range raw.FieldRules {
		field := BriefField(strings.ToLower(strings.TrimSpace(string(rawRule.Field))))
		if !validMinimumBriefField(field) {
			return OutcomeTemplate{}, nil, fmt.Errorf("unsupported minimum brief field %q", rawRule.Field)
		}
		if _, duplicate := seen[field]; duplicate {
			return OutcomeTemplate{}, nil, fmt.Errorf("duplicate minimum brief field rule %q", field)
		}
		seen[field] = struct{}{}

		classification := BriefResolutionClass(strings.ToLower(strings.TrimSpace(string(rawRule.UnresolvedAs))))
		if !validBriefResolutionClass(classification) {
			return OutcomeTemplate{}, nil, fmt.Errorf("unsupported unresolved classification %q for %s", rawRule.UnresolvedAs, field)
		}
		rule := rules[field]
		rule.UnresolvedAs = classification
		if question := strings.TrimSpace(rawRule.Question); question != "" {
			rule.Question = question
		}
		if rawRule.Impact != 0 {
			rule.Impact = rawRule.Impact
		}
		rules[field] = rule
	}
	template.FieldRules = cloneOutcomeTemplateFieldRules(raw.FieldRules)
	template.OutputContract = cloneWorkOutputContract(raw.OutputContract)
	return template, rules, nil
}

func defaultOutcomeTemplateFieldRules() map[BriefField]OutcomeTemplateFieldRule {
	return map[BriefField]OutcomeTemplateFieldRule{
		BriefFieldTargetOutcome: {
			Field: BriefFieldTargetOutcome, UnresolvedAs: BriefMustAsk, Impact: 100,
			Question: "What outcome should this work produce?",
		},
		BriefFieldAudience: {
			Field: BriefFieldAudience, UnresolvedAs: BriefCanInfer, Impact: 50,
			Question: "Who is the intended audience?",
		},
		BriefFieldEssentialBehavior: {
			Field: BriefFieldEssentialBehavior, UnresolvedAs: BriefCanInfer, Impact: 70,
			Question: "Which behavior is essential to the outcome?",
		},
		BriefFieldQualityBar: {
			Field: BriefFieldQualityBar, UnresolvedAs: BriefCanInfer, Impact: 60,
			Question: "What quality bar should the result meet?",
		},
		BriefFieldDeliveryForm: {
			Field: BriefFieldDeliveryForm, UnresolvedAs: BriefMustAsk, Impact: 80,
			Question: "What delivery form should the outcome take?",
		},
		BriefFieldConstraints: {
			Field: BriefFieldConstraints, UnresolvedAs: BriefCanDefer, Impact: 40,
			Question: "Which constraints must shape the work?",
		},
		BriefFieldAcceptanceEvidence: {
			Field: BriefFieldAcceptanceEvidence, UnresolvedAs: BriefMustAsk, Impact: 90,
			Question: "What evidence will show that the outcome is accepted?",
		},
	}
}
