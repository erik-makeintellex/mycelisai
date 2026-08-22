package configdocuments

import (
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

// CompileWorkerProfileDocument validates and adapts a WorkerProfile envelope
// without storing, activating, or projecting it into runtime state.
func CompileWorkerProfileDocument(document protocol.ConfigDocument) (protocol.WorkerProfileCompileResult, error) {
	issues := protocol.ValidateConfigDocument(document)
	if len(issues) != 0 {
		return protocol.WorkerProfileCompileResult{}, &ValidationError{Issues: issues}
	}
	if document.Kind != protocol.ConfigDocumentKindWorkerProfile {
		return protocol.WorkerProfileCompileResult{}, fmt.Errorf("config document kind %q is not a WorkerProfile", document.Kind)
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return protocol.WorkerProfileCompileResult{}, fmt.Errorf("compile worker profile digest: %w", err)
	}
	profile, err := protocol.DecodeWorkerProfileSpec(document.Spec)
	if err != nil {
		return protocol.WorkerProfileCompileResult{}, err
	}
	return protocol.WorkerProfileCompileResult{
		Profile: profile,
		Snapshot: protocol.WorkerProfileSnapshot{
			ID: document.Metadata.ID, Version: document.Metadata.Version, Digest: digest,
		},
		Ready: true,
	}, nil
}
