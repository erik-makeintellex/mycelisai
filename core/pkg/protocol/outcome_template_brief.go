package protocol

import "strings"

func resolveMinimumSufficientBrief(defaults, operator, policy MinimumSufficientBrief) (MinimumSufficientBrief, map[BriefField]BriefValueSource) {
	resolved := MinimumSufficientBrief{}
	sources := make(map[BriefField]BriefValueSource, len(minimumBriefFieldOrder))
	applyMinimumBriefValues(&resolved, sources, normalizeMinimumSufficientBrief(defaults), BriefValueTemplateDefault)
	applyMinimumBriefValues(&resolved, sources, normalizeMinimumSufficientBrief(operator), BriefValueOperator)
	applyMinimumBriefValues(&resolved, sources, normalizeMinimumSufficientBrief(policy), BriefValuePolicy)
	return resolved, sources
}

func applyMinimumBriefValues(target *MinimumSufficientBrief, sources map[BriefField]BriefValueSource, values MinimumSufficientBrief, source BriefValueSource) {
	if values.TargetOutcome != "" {
		target.TargetOutcome = values.TargetOutcome
		sources[BriefFieldTargetOutcome] = source
	}
	if values.Audience != "" {
		target.Audience = values.Audience
		sources[BriefFieldAudience] = source
	}
	if len(values.EssentialBehavior) > 0 {
		target.EssentialBehavior = append([]string(nil), values.EssentialBehavior...)
		sources[BriefFieldEssentialBehavior] = source
	}
	if values.QualityBar != "" {
		target.QualityBar = values.QualityBar
		sources[BriefFieldQualityBar] = source
	}
	if values.DeliveryForm != "" {
		target.DeliveryForm = values.DeliveryForm
		sources[BriefFieldDeliveryForm] = source
	}
	if len(values.Constraints) > 0 {
		target.Constraints = append([]string(nil), values.Constraints...)
		sources[BriefFieldConstraints] = source
	}
	if len(values.AcceptanceEvidence) > 0 {
		target.AcceptanceEvidence = append([]string(nil), values.AcceptanceEvidence...)
		sources[BriefFieldAcceptanceEvidence] = source
	}
}

func normalizeMinimumSufficientBrief(raw MinimumSufficientBrief) MinimumSufficientBrief {
	return MinimumSufficientBrief{
		TargetOutcome:      strings.TrimSpace(raw.TargetOutcome),
		Audience:           strings.TrimSpace(raw.Audience),
		EssentialBehavior:  dedupeStrings(compactStrings(raw.EssentialBehavior)),
		QualityBar:         strings.TrimSpace(raw.QualityBar),
		DeliveryForm:       strings.TrimSpace(raw.DeliveryForm),
		Constraints:        dedupeStrings(compactStrings(raw.Constraints)),
		AcceptanceEvidence: dedupeStrings(compactStrings(raw.AcceptanceEvidence)),
	}
}

func minimumBriefFieldResolved(brief MinimumSufficientBrief, field BriefField) bool {
	switch field {
	case BriefFieldTargetOutcome:
		return brief.TargetOutcome != ""
	case BriefFieldAudience:
		return brief.Audience != ""
	case BriefFieldEssentialBehavior:
		return len(brief.EssentialBehavior) > 0
	case BriefFieldQualityBar:
		return brief.QualityBar != ""
	case BriefFieldDeliveryForm:
		return brief.DeliveryForm != ""
	case BriefFieldConstraints:
		return len(brief.Constraints) > 0
	case BriefFieldAcceptanceEvidence:
		return len(brief.AcceptanceEvidence) > 0
	default:
		return false
	}
}

func compileBriefWorkIntent(template OutcomeTemplate, brief MinimumSufficientBrief, snapshot OutcomeTemplateSnapshot) *WorkIntent {
	output := cloneWorkOutputContract(template.OutputContract)
	if output == nil {
		output = &WorkOutputContract{}
	}
	if output.Shape == "" {
		output.Shape = brief.DeliveryForm
	}
	if output.PrimaryDeliverable == "" {
		output.PrimaryDeliverable = brief.DeliveryForm
	}
	output.Validation = append(output.Validation, brief.AcceptanceEvidence...)

	intentKind := template.IntentKind
	if intentKind == "" {
		intentKind = "one_shot"
	}
	intent := NormalizeWorkIntent(&WorkIntent{
		Kind: intentKind, Objective: brief.TargetOutcome, OutputContract: output,
	})
	briefCopy := cloneMinimumSufficientBrief(brief)
	snapshotCopy := snapshot
	intent.MinimumSufficientBrief = &briefCopy
	intent.OutcomeTemplateSnapshot = &snapshotCopy
	return intent
}

func cloneMinimumSufficientBrief(brief MinimumSufficientBrief) MinimumSufficientBrief {
	copy := brief
	copy.EssentialBehavior = append([]string(nil), brief.EssentialBehavior...)
	copy.Constraints = append([]string(nil), brief.Constraints...)
	copy.AcceptanceEvidence = append([]string(nil), brief.AcceptanceEvidence...)
	return copy
}

func cloneWorkOutputContract(raw *WorkOutputContract) *WorkOutputContract {
	if raw == nil {
		return nil
	}
	copy := *raw
	copy.Validation = append([]string(nil), raw.Validation...)
	copy.OutputValidation = NormalizeOutputValidationPlan(raw.OutputValidation)
	return &copy
}

func cloneOutcomeTemplateFieldRules(raw []OutcomeTemplateFieldRule) []OutcomeTemplateFieldRule {
	if raw == nil {
		return nil
	}
	return append([]OutcomeTemplateFieldRule(nil), raw...)
}

func hasMustAsk(unresolved []UnresolvedBriefField) bool {
	for _, field := range unresolved {
		if field.Classification == BriefMustAsk {
			return true
		}
	}
	return false
}

func validMinimumBriefField(field BriefField) bool {
	for _, candidate := range minimumBriefFieldOrder {
		if field == candidate {
			return true
		}
	}
	return false
}

func validBriefResolutionClass(classification BriefResolutionClass) bool {
	switch classification {
	case BriefMustAsk, BriefCanInfer, BriefCanDefer:
		return true
	default:
		return false
	}
}
