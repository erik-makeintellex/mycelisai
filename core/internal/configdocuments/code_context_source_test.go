package configdocuments

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestCompileCodeContextSourceDocumentUsesEnvelopeLineage(t *testing.T) {
	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindCodeContextSource,
		Metadata: protocol.ConfigDocumentMetadata{
			ID:      "code-context-source",
			Name:    "Core source",
			Version: "1.0.0",
			OwnerID: "operator-1",
			Scope:   protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"},
			Enabled: true,
			Source:  protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceFile, Ref: "code-context/core.yaml"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel:       protocol.ConfigDocumentRiskMedium,
				ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: json.RawMessage(`{"source_id":"core-source","source_type":"repository","root_path":"core"}`),
	}

	compiled, err := CompileCodeContextSourceDocument(document)
	if err != nil {
		t.Fatalf("CompileCodeContextSourceDocument() error = %v", err)
	}
	if compiled.Source.SourceID != "core-source" || compiled.Scope.Ref != "workspace-1" || compiled.Digest == "" {
		t.Fatalf("compiled = %+v", compiled)
	}
}

func TestCompileDocumentDispatchesCodeContextSource(t *testing.T) {
	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindCodeContextSource,
		Metadata: protocol.ConfigDocumentMetadata{
			ID:      "code-context-source",
			Name:    "Core source",
			Version: "1.0.0",
			OwnerID: "operator-1",
			Scope:   protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"},
			Enabled: true,
			Source:  protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceFile, Ref: "code-context/core.yaml"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel:       protocol.ConfigDocumentRiskMedium,
				ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: json.RawMessage(`{"source_id":"core-source","source_type":"repository","root_path":"core"}`),
	}

	compiled, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{})
	if err != nil {
		t.Fatalf("CompileDocument() error = %v", err)
	}
	if _, ok := compiled.(protocol.CodeContextSourceCompileResult); !ok {
		t.Fatalf("compiled type = %T", compiled)
	}
}
