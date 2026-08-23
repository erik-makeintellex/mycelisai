package configdocuments

import (
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

func CompileCodeContextSourceDocument(document protocol.ConfigDocument) (protocol.CodeContextSourceCompileResult, error) {
	issues := protocol.ValidateConfigDocument(document)
	if len(issues) != 0 {
		return protocol.CodeContextSourceCompileResult{}, &ValidationError{Issues: issues}
	}
	if document.Kind != protocol.ConfigDocumentKindCodeContextSource {
		return protocol.CodeContextSourceCompileResult{}, fmt.Errorf("config document kind %q is not a CodeContextSource", document.Kind)
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return protocol.CodeContextSourceCompileResult{}, fmt.Errorf("compile code context source digest: %w", err)
	}
	source, err := protocol.DecodeCodeContextSourceSpec(document.Spec)
	if err != nil {
		return protocol.CodeContextSourceCompileResult{}, err
	}
	if source.ExtractionVersion == "" {
		source.ExtractionVersion = protocol.CodeContextExtractionVersionV1
	}
	return protocol.CodeContextSourceCompileResult{
		Source: source,
		Scope:  document.Metadata.Scope,
		Digest: digest,
	}, nil
}
