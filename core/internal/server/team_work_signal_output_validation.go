package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func deliverableResultMissingOutputs(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) bool {
	return deliverableResultOutputIssue(item, payloadKind, outputRefs) != ""
}

func deliverableResultOutputIssue(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) string {
	if payloadKind != protocol.PayloadKindResult ||
		(item.ExecutionShape != protocol.TeamExecutionShapeDeliverable && item.ExecutionShape != protocol.TeamExecutionShapeDelegatedWork) ||
		len(item.ExpectedOutputs) == 0 {
		return ""
	}
	if len(outputRefs) == 0 {
		return "missing_retained_output"
	}
	if !teamWorkExpectsProjectPackage(item) {
		return ""
	}
	teamRoot := "groups/" + strings.Trim(strings.TrimSpace(item.TeamID), "/") + "/"
	for _, ref := range outputRefs {
		storageRef := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.StorageRef), "\\", "/"), "/")
		if strings.EqualFold(strings.TrimSpace(ref.Kind), "project_package") &&
			strings.TrimSpace(ref.Entrypoint) != "" &&
			strings.HasPrefix(storageRef+"/", teamRoot) {
			return ""
		}
	}
	return "invalid_deliverable_shape"
}

func teamWorkExpectsProjectPackage(item protocol.TeamWorkItem) bool {
	if item.WorkIntent != nil && item.WorkIntent.OutputContract != nil &&
		strings.EqualFold(strings.TrimSpace(item.WorkIntent.OutputContract.Shape), "app_package") {
		return true
	}
	for _, expected := range item.ExpectedOutputs {
		lower := strings.ToLower(expected)
		if strings.Contains(lower, "project package") || strings.Contains(lower, "application package") ||
			strings.Contains(lower, "browser game") || strings.Contains(lower, "playable app") {
			return true
		}
	}
	return false
}
