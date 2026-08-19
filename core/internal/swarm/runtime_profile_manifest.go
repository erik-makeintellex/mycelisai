package swarm

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func runtimeWorkerProfileSnapshot(raw any) *protocol.WorkerProfileSnapshot {
	switch snapshot := raw.(type) {
	case *protocol.WorkerProfileSnapshot:
		if snapshot == nil {
			return nil
		}
		copy := *snapshot
		return &copy
	case protocol.WorkerProfileSnapshot:
		copy := snapshot
		return &copy
	}
	source, ok := raw.(map[string]any)
	if !ok || strings.TrimSpace(stringValue(source["id"])) == "" {
		return nil
	}
	scopeSource, _ := source["scope"].(map[string]any)
	return &protocol.WorkerProfileSnapshot{
		ID: stringValue(source["id"]), Version: stringValue(source["version"]),
		Digest: stringValue(source["digest"]), RecordID: stringValue(source["record_id"]),
		TenantID: stringValue(source["tenant_id"]),
		Scope: protocol.ConfigDocumentScope{
			Kind: protocol.ConfigDocumentScopeKind(stringValue(scopeSource["kind"])), Ref: stringValue(scopeSource["ref"]),
		},
	}
}

func runtimeVerification(raw any) *protocol.Verification {
	switch verification := raw.(type) {
	case *protocol.Verification:
		if verification == nil {
			return nil
		}
		copy := *verification
		copy.Rubric = append([]string(nil), verification.Rubric...)
		return &copy
	case protocol.Verification:
		copy := verification
		copy.Rubric = append([]string(nil), verification.Rubric...)
		return &copy
	}
	source, ok := raw.(map[string]any)
	if !ok || strings.TrimSpace(stringValue(source["strategy"])) == "" {
		return nil
	}
	return &protocol.Verification{
		Strategy: protocol.VerifyStrategy(stringValue(source["strategy"])),
		Rubric:   stringSlice(source["rubric"]), ValidationCommand: stringValue(source["validation_command"]),
	}
}
