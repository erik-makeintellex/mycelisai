package configdocuments

import (
	"encoding/json"
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

// CompileOutcomeTemplateDocument adapts a validated configuration document to
// the existing WorkIntent compiler. It does not grant execution authority.
func CompileOutcomeTemplateDocument(
	document protocol.ConfigDocument,
	operatorValues protocol.MinimumSufficientBrief,
	policyValues protocol.MinimumSufficientBrief,
) (protocol.OutcomeTemplateCompileResult, error) {
	if issues := protocol.ValidateConfigDocument(document); len(issues) > 0 {
		return protocol.OutcomeTemplateCompileResult{}, fmt.Errorf("invalid config document: %s", issues[0].Code)
	}
	if document.Kind != protocol.ConfigDocumentKindOutcomeTemplate {
		return protocol.OutcomeTemplateCompileResult{}, fmt.Errorf("config document kind %q is not an outcome template", document.Kind)
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return protocol.OutcomeTemplateCompileResult{}, err
	}
	var template protocol.OutcomeTemplate
	if err := json.Unmarshal(document.Spec, &template); err != nil {
		return protocol.OutcomeTemplateCompileResult{}, fmt.Errorf("decode outcome template spec: %w", err)
	}
	template.ID = document.Metadata.ID
	template.Version = document.Metadata.Version
	template.Digest = digest
	return protocol.CompileOutcomeTemplate(protocol.OutcomeTemplateCompileInput{
		Template: template, OperatorValues: operatorValues, PolicyValues: policyValues,
	})
}
