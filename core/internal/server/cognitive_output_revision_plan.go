package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const revisionAcceptanceCriterion = "requested revision changes its declared visible revision state"

func buildTrustedOutputRevisionPlan(ctx *chatContinuationContext, request string) ([]protocol.PlannedToolCall, bool) {
	if ctx == nil || !ctx.OwnershipValidated || ctx.TeamID == "" || ctx.RevisionTarget == "" ||
		(ctx.Intent != "update" && ctx.Intent != "fork") {
		return nil, false
	}
	contract := inheritedRevisionContentContract(ctx, request)
	digestLabel := strings.TrimPrefix(ctx.SourceDigest, "sha256:")
	if len(digestLabel) > 10 {
		digestLabel = digestLabel[:10]
	}
	handoff := fmt.Sprintf("groups/%s/planning/RESEARCH_COUNCIL_HANDOFF_REVISION_%s.md", ctx.TeamID, digestLabel)
	brief := fmt.Sprintf("groups/%s/planning/TEAM_EVOCATION.md", ctx.TeamID)
	title := firstNonEmptyString(ctx.Title, "Delivered package") + " Revision 2"
	entrypoint := strings.TrimRight(ctx.RevisionTarget, "/") + "/index.html"
	revisionRequest := fmt.Sprintf("%s Preserve source output %s at digest %s. Retain the revision at %s with entrypoint %s and package title %s.",
		request, ctx.Reference, ctx.SourceDigest, ctx.RevisionTarget, entrypoint, title)
	calls := buildTeamEvocationDeliveryPlan(revisionRequest, ctx.TeamID, brief, contract)
	if len(calls) != 2 {
		return nil, false
	}
	calls[0].Arguments["path"] = handoff
	calls[0].Arguments["source_output"] = continuationLineageMap(ctx)
	calls[1].Arguments["research_handoff"] = handoff
	ask := mapArgument(calls[1].Arguments["ask"])
	context := mapArgument(ask["context"])
	resultContract := mapArgument(context["result_contract"])
	resultContract["package_folder"] = ctx.RevisionTarget
	resultContract["package_entrypoint"] = entrypoint
	resultContract["package_title"] = title
	resultContract["source_lineage"] = continuationLineageMap(ctx)
	context["research_council_handoff"] = handoff
	context["source_lineage"] = continuationLineageMap(ctx)
	ask["context"] = context
	calls[1].Arguments["ask"] = ask
	calls[1].Arguments["source_lineage"] = continuationLineageMap(ctx)
	return calls, true
}

func inheritedRevisionContentContract(ctx *chatContinuationContext, request string) map[string]any {
	revision := contentContractForTeamRequest(request)
	criteria := confirmedActionStringSlice(revision["acceptance_criteria"])
	outputs := confirmedActionStringSlice(revision["expected_outputs"])
	proof := confirmedActionStringSlice(revision["proof_required"])
	if source := ctx.SourceWorkIntent; source != nil && source.OutputContract != nil {
		criteria = mergeExpectedOutputs(source.OutputContract.AcceptanceCriteria, criteria)
		outputs = mergeExpectedOutputs([]string{source.OutputContract.PrimaryDeliverable}, outputs)
		proof = mergeExpectedOutputs(source.OutputContract.Validation, proof)
		revision["output_validation"] = protocol.NormalizeOutputValidationPlan(source.OutputContract.OutputValidation)
	}
	criteria = mergeExpectedOutputs(criteria, []string{revisionAcceptanceCriterion})
	proof = mergeExpectedOutputs(proof, []string{
		`revision validation hook: [data-mycelis-validation-action="revision"] changes [data-mycelis-revision-state]`,
		"requested revision is described in retained proof and usage notes: " + strings.TrimSpace(request),
	})
	revision["acceptance_criteria"] = criteria
	revision["expected_outputs"] = outputs
	revision["proof_required"] = proof
	return revision
}

func continuationLineageMap(ctx *chatContinuationContext) map[string]any {
	return map[string]any{
		"team_id": ctx.TeamID, "run_id": ctx.RunID, "work_item_id": ctx.WorkItemID,
		"output_id": ctx.OutputID, "source_digest": ctx.SourceDigest,
		"source_version": ctx.SourceVersion, "source_reference": ctx.Reference,
		"revision_target": ctx.RevisionTarget,
	}
}

func inheritRevisionWorkIntent(current *protocol.WorkIntent, ctx *chatContinuationContext, request string) *protocol.WorkIntent {
	if ctx == nil || !ctx.OwnershipValidated {
		return current
	}
	intent := protocol.NormalizeWorkIntent(current)
	if intent == nil {
		intent = &protocol.WorkIntent{Kind: "project"}
	}
	contract := inheritedRevisionContentContract(ctx, request)
	if intent.OutputContract == nil {
		intent.OutputContract = &protocol.WorkOutputContract{}
	}
	intent.OutputContract.AcceptanceCriteria = confirmedActionStringSlice(contract["acceptance_criteria"])
	intent.OutputContract.SemanticValidationRequired = len(intent.OutputContract.AcceptanceCriteria) > 0
	if ctx.SourceWorkIntent != nil && ctx.SourceWorkIntent.OutputContract != nil {
		source := ctx.SourceWorkIntent.OutputContract
		intent.OutputContract.Shape = source.Shape
		intent.OutputContract.Retention = source.Retention
		intent.OutputContract.OutputValidation = protocol.NormalizeOutputValidationPlan(source.OutputValidation)
	}
	intent.OutputContract.PrimaryDeliverable = ctx.RevisionTarget
	return protocol.NormalizeWorkIntent(intent)
}
