package configdocuments

import (
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

// CompileDocument dispatches family-specific compilation while preserving one
// ConfigDocument envelope and one validation path.
func CompileDocument(
	document protocol.ConfigDocument,
	operatorValues protocol.MinimumSufficientBrief,
	policyValues protocol.MinimumSufficientBrief,
) (any, error) {
	switch document.Kind {
	case protocol.ConfigDocumentKindOutcomeTemplate:
		return CompileOutcomeTemplateDocument(document, operatorValues, policyValues)
	case protocol.ConfigDocumentKindWorkerProfile:
		return CompileWorkerProfileDocument(document)
	case protocol.ConfigDocumentKindCodeContextSource:
		return CompileCodeContextSourceDocument(document)
	default:
		return nil, fmt.Errorf("unsupported config document kind %q", document.Kind)
	}
}
