package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// RunLifecycleStage represents one stage in the canonical V8.3 9-stage operator progression.
type RunLifecycleStage struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Status    string `json:"status"` // "completed" | "active" | "pending" | "failed"
	Timestamp string `json:"timestamp,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Actor     string `json:"actor,omitempty"`
}

// BuildRunLifecycleStages projects a chronological slice of MissionEventEnvelope records
// into the canonical V8.3 9-stage operator timeline.
func BuildRunLifecycleStages(events []protocol.MissionEventEnvelope) []RunLifecycleStage {
	stages := []RunLifecycleStage{
		{ID: "intent_received", Label: "Intent Received", Status: "pending"},
		{ID: "proposal_generated", Label: "Proposal Generated", Status: "pending"},
		{ID: "approval_required", Label: "Approval Gate", Status: "pending"},
		{ID: "approval_granted", Label: "Approval Granted", Status: "pending"},
		{ID: "mission_running", Label: "Mission Running", Status: "pending"},
		{ID: "tool_action_execution", Label: "Tool / Action Execution", Status: "pending"},
		{ID: "package_generated", Label: "Package Generated", Status: "pending"},
		{ID: "proof_generated", Label: "Proof Verified", Status: "pending"},
		{ID: "recovery_next_actions", Label: "Recovery & Next Actions", Status: "pending"},
	}
	if len(events) == 0 {
		return stages
	}

	var (
		intentEvt, proposalEvt, approvalReq, approvalGrant, approvalDeny *protocol.MissionEventEnvelope
		runningEvt, toolInvoked, toolDone, toolFailed                    *protocol.MissionEventEnvelope
		artifactEvt, proofEvt, completedEvt, failedEvt                   *protocol.MissionEventEnvelope
		invokedCount, doneCount                                          int
	)

	for i := range events {
		ev := &events[i]
		if intentEvt == nil {
			intentEvt = ev
		}
		switch strings.TrimSpace(string(ev.EventType)) {
		case string(protocol.EventMissionStarted), string(protocol.EventTriggerFired):
			intentEvt = ev
			if runningEvt == nil {
				runningEvt = ev
			}
			if _, hasProposal := ev.Payload["proposal"]; hasProposal && proposalEvt == nil {
				proposalEvt = ev
			}
		case string(protocol.EventTeamSpawned), string(protocol.EventTemplateCloned), "proposal.generated", "mission.proposal":
			if proposalEvt == nil {
				proposalEvt = ev
			}
			runningEvt = ev
		case "approval.requested":
			approvalReq = ev
			if proposalEvt == nil {
				proposalEvt = ev
			}
		case "approval.granted", string(protocol.EventConfigApplied):
			approvalGrant = ev
		case "approval.denied":
			approvalDeny = ev
		case string(protocol.EventAgentStarted), "agent.heartbeat":
			runningEvt = ev
		case string(protocol.EventToolInvoked):
			invokedCount++
			toolInvoked = ev
			if proposalEvt == nil {
				proposalEvt = ev
			}
			if runningEvt == nil {
				runningEvt = ev
			}
		case string(protocol.EventToolCompleted):
			doneCount++
			toolDone = ev
		case string(protocol.EventToolFailed):
			toolFailed = ev
		case string(protocol.EventArtifactCreated):
			artifactEvt = ev
			if hasProofPayload(ev.Payload) {
				proofEvt = ev
			}
		case string(protocol.EventMemoryStored):
			if proofEvt == nil {
				proofEvt = ev
			}
		case string(protocol.EventMissionCompleted):
			completedEvt = ev
			if proofEvt == nil {
				proofEvt = ev
			}
		case string(protocol.EventMissionFailed), string(protocol.EventMissionCancelled):
			failedEvt = ev
		}
	}

	applyStageEvent(&stages[0], intentEvt, "completed", "Mission intent accepted for execution.")

	switch {
	case proposalEvt != nil:
		applyStageEvent(&stages[1], proposalEvt, "completed", stageSummaryFromEvent(proposalEvt, "Execution plan and team scope prepared."))
	case approvalReq != nil || approvalGrant != nil || approvalDeny != nil || completedEvt != nil || artifactEvt != nil:
		applyStageEvent(&stages[1], intentEvt, "completed", "Execution plan resolved from mission intent.")
	case intentEvt != nil:
		applyStageEvent(&stages[1], intentEvt, "completed", "Direct mission intent scope established.")
	}

	switch {
	case approvalDeny != nil:
		applyStageEvent(&stages[2], firstStageEvent(approvalReq, approvalDeny), "completed", "Governance approval gate evaluated.")
	case approvalReq != nil && approvalGrant == nil && completedEvt == nil && toolInvoked == nil && artifactEvt == nil:
		applyStageEvent(&stages[2], approvalReq, "active", stageSummaryFromEvent(approvalReq, "Awaiting operator approval decision."))
	case approvalReq != nil:
		applyStageEvent(&stages[2], approvalReq, "completed", "Operator approval gate satisfied.")
	case approvalGrant != nil || toolInvoked != nil || toolDone != nil || artifactEvt != nil || completedEvt != nil || runningEvt != nil:
		applyStageEvent(&stages[2], firstStageEvent(approvalGrant, runningEvt, intentEvt), "completed", "Governed policy gate cleared.")
	}

	switch {
	case approvalDeny != nil:
		applyStageEvent(&stages[3], approvalDeny, "failed", stageSummaryFromEvent(approvalDeny, "Approval denied by operator."))
	case approvalGrant != nil:
		applyStageEvent(&stages[3], approvalGrant, "completed", stageSummaryFromEvent(approvalGrant, "Operator approval confirmed."))
	case stages[2].Status == "completed" && (toolInvoked != nil || toolDone != nil || artifactEvt != nil || completedEvt != nil || runningEvt != nil):
		applyStageEvent(&stages[3], firstStageEvent(runningEvt, intentEvt), "completed", "Authorized under active governance profile.")
	}

	switch {
	case approvalDeny != nil:
		stages[4].Status = "pending"
	case failedEvt != nil && toolFailed == nil:
		applyStageEvent(&stages[4], failedEvt, "failed", stageSummaryFromEvent(failedEvt, "Mission run failed during execution."))
	case completedEvt != nil || artifactEvt != nil || toolInvoked != nil || toolDone != nil || toolFailed != nil:
		applyStageEvent(&stages[4], firstStageEvent(runningEvt, intentEvt), "completed", "Mission runtime active and dispatched.")
	case runningEvt != nil && stages[2].Status != "active":
		applyStageEvent(&stages[4], runningEvt, "active", "Mission runtime executing.")
	}

	switch {
	case toolFailed != nil:
		applyStageEvent(&stages[5], toolFailed, "failed", stageSummaryFromEvent(toolFailed, "Tool execution encountered an error."))
	case completedEvt != nil && (toolDone != nil || toolInvoked != nil || artifactEvt != nil):
		ev := firstStageEvent(toolDone, toolInvoked, artifactEvt, completedEvt)
		applyStageEvent(&stages[5], ev, "completed", stageSummaryFromEvent(ev, "All planned tool actions completed."))
	case toolInvoked != nil && (invokedCount > doneCount || (completedEvt == nil && artifactEvt == nil)):
		applyStageEvent(&stages[5], firstStageEvent(toolDone, toolInvoked), "active", stageSummaryFromEvent(toolInvoked, "Executing governed tool invocation."))
	case toolDone != nil || artifactEvt != nil:
		ev := firstStageEvent(toolDone, artifactEvt)
		applyStageEvent(&stages[5], ev, "completed", stageSummaryFromEvent(ev, "Tool execution completed."))
	}

	switch {
	case artifactEvt != nil:
		applyStageEvent(&stages[6], artifactEvt, "completed", stageSummaryFromEvent(artifactEvt, "Deliverable package artifact recorded."))
	case completedEvt != nil:
		applyStageEvent(&stages[6], completedEvt, "completed", "Run output package assembled.")
	}

	switch {
	case completedEvt != nil || proofEvt != nil:
		ev := firstStageEvent(proofEvt, completedEvt)
		applyStageEvent(&stages[7], ev, "completed", stageSummaryFromEvent(ev, "Execution proof verified against audit spine."))
	case failedEvt != nil || toolFailed != nil || approvalDeny != nil:
		applyStageEvent(&stages[7], firstStageEvent(failedEvt, toolFailed, approvalDeny), "failed", "Proof verification halted due to run failure.")
	}

	switch {
	case approvalDeny != nil || toolFailed != nil || failedEvt != nil:
		ev := firstStageEvent(approvalDeny, toolFailed, failedEvt)
		applyStageEvent(&stages[8], ev, "failed", fmt.Sprintf("Recovery required: %s", stageSummaryFromEvent(ev, "Inspect failed step and retry or adjust scope.")))
	case completedEvt != nil:
		applyStageEvent(&stages[8], completedEvt, "completed", "Next actions ready: inspect delivery package and promote verified outputs.")
	}

	return stages
}

func applyStageEvent(stage *RunLifecycleStage, ev *protocol.MissionEventEnvelope, status, summary string) {
	stage.Status = status
	stage.Summary = summary
	if ev == nil {
		return
	}
	if !ev.EmittedAt.IsZero() {
		stage.Timestamp = ev.EmittedAt.UTC().Format(time.RFC3339)
	}
	stage.Actor = firstNonEmptyString(ev.SourceAgent, ev.SourceTeam)
}

func firstStageEvent(candidates ...*protocol.MissionEventEnvelope) *protocol.MissionEventEnvelope {
	for _, ev := range candidates {
		if ev != nil {
			return ev
		}
	}
	return nil
}

func hasProofPayload(payload map[string]any) bool {
	return len(payload) > 0 && firstNonEmptyString(payload["proof_id"], payload["proof_path"], payload["proof_ref"]) != ""
}

func stageSummaryFromEvent(ev *protocol.MissionEventEnvelope, fallback string) string {
	if ev == nil || len(ev.Payload) == 0 {
		return fallback
	}
	if s := firstNonEmptyString(ev.Payload["summary"], ev.Payload["message"], ev.Payload["reason"], ev.Payload["error"]); s != "" {
		return s
	}
	if pkg := firstNonEmptyString(ev.Payload["package_slug"], ev.Payload["package_kind"]); pkg != "" {
		return "Package: " + pkg
	}
	if tool := firstNonEmptyString(ev.Payload["tool"], ev.Payload["tool_name"]); tool != "" {
		return "Tool: " + tool
	}
	return fallback
}

// handleGetRunStages returns the 9-stage V8.3 lifecycle projection for a run.
func (s *AdminServer) handleGetRunStages(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		respondAPIError(w, "run_id is required", http.StatusBadRequest)
		return
	}
	if s.Events == nil {
		respondAPIError(w, "event store not initialized", http.StatusServiceUnavailable)
		return
	}
	events, err := s.Events.GetRunTimeline(r.Context(), runID)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"run_id":       runID,
		"stages":       BuildRunLifecycleStages(events),
		"events_count": len(events),
	}))
}
