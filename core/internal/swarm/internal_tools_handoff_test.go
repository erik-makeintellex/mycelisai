package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type recordingHandoffRecorder struct {
	records []HandoffRequest
	reads   []HandoffReadRequest
	err     error
}

func (r *recordingHandoffRecorder) RecordHandoff(_ context.Context, req HandoffRequest) (HandoffReceipt, error) {
	r.records = append(r.records, req)
	if r.err != nil {
		return HandoffReceipt{}, r.err
	}
	return HandoffReceipt{HandoffID: "h-1", WorkItemID: "w-1", Status: "queued"}, nil
}

func (r *recordingHandoffRecorder) ReadHandoffInput(_ context.Context, req HandoffReadRequest) (HandoffInput, error) {
	r.reads = append(r.reads, req)
	if r.err != nil {
		return HandoffInput{}, r.err
	}
	return HandoffInput{HandoffID: req.HandoffID, ArtifactID: req.ArtifactID, Title: "Fact sheet", Content: "facts"}, nil
}

const handoffArtifactID = "11111111-1111-4111-8111-111111111111"

func handOffArgs() map[string]any {
	return map[string]any{
		"target_team_id": "marketing", "artifact_ids": []any{handoffArtifactID},
		"note": "Fact sheet for the weekend promo.", "expected_action": "use",
		// Forged identity fields are ignored; identity comes from the invocation.
		"source_team": "admin-core", "role": "admin", "run_id": "run-forged",
	}
}

func requireHandoffCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || HandoffBlockCode(err) != code {
		t.Fatalf("error = %v, want blocker %s", err, code)
	}
}

func TestHandOffUsesAuthoritativeLeadIdentity(t *testing.T) {
	recorder := &recordingHandoffRecorder{}
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(recorder)

	out, err := registry.handleHandOff(agentInvocation("research-lead", "research", "lead", "run-7"), handOffArgs())
	if err != nil {
		t.Fatalf("hand_off: %v", err)
	}
	var receipt HandoffReceipt
	if json.Unmarshal([]byte(out), &receipt) != nil || receipt.Status != "queued" || strings.Contains(out, "delivered") {
		t.Fatalf("hand_off output = %s, want queued receipt", out)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("recorder calls = %d", len(recorder.records))
	}
	got := recorder.records[0]
	if got.SourceTeamID != "research" || got.SourceAgentID != "research-lead" || got.RunID != "run-7" ||
		got.TargetTeamID != "marketing" || got.ExpectedAction != "use" || len(got.ArtifactIDs) != 1 {
		t.Fatalf("recorded request = %#v", got)
	}
}

func TestHandOffDeniesSpecialist(t *testing.T) {
	recorder := &recordingHandoffRecorder{}
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(recorder)
	_, err := registry.handleHandOff(agentInvocation("research-analyst", "research", "researcher", "run-7"), handOffArgs())
	requireHandoffCode(t, err, "handoff_not_team_lead")
	_, err = registry.handleReadHandoffInput(agentInvocation("marketing-writer", "marketing", "writer", "run-7"), map[string]any{"handoff_id": "h-1", "artifact_id": handoffArtifactID})
	requireHandoffCode(t, err, "handoff_not_team_lead")
	if len(recorder.records)+len(recorder.reads) != 0 {
		t.Fatal("denied calls reached the recorder")
	}
}

func TestHandOffOutsideRunNeedsApproval(t *testing.T) {
	recorder := &recordingHandoffRecorder{}
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(recorder)
	_, err := registry.handleHandOff(agentInvocation("research-lead", "research", "lead", ""), handOffArgs())
	requireHandoffCode(t, err, "handoff_needs_approval")
	if len(recorder.records) != 0 {
		t.Fatal("out-of-run handoff reached the recorder")
	}
}

func TestHandOffRejectsInvalidRequests(t *testing.T) {
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(&recordingHandoffRecorder{})
	ctx := agentInvocation("research-lead", "research", "lead", "run-7")
	for name, mutate := range map[string]func(map[string]any){
		"self target":    func(a map[string]any) { a["target_team_id"] = "research" },
		"no artifacts":   func(a map[string]any) { a["artifact_ids"] = []any{} },
		"bad artifact":   func(a map[string]any) { a["artifact_ids"] = []any{"not-a-uuid"} },
		"long note":      func(a map[string]any) { a["note"] = strings.Repeat("x", 1001) },
		"empty note":     func(a map[string]any) { a["note"] = " " },
		"unknown action": func(a map[string]any) { a["expected_action"] = "deploy" },
	} {
		args := handOffArgs()
		mutate(args)
		if _, err := registry.handleHandOff(ctx, args); HandoffBlockCode(err) != "handoff_invalid_request" {
			t.Fatalf("%s: error = %v", name, err)
		}
	}
}

func TestHandOffBlockedDuringPlanningAndWithoutRecorder(t *testing.T) {
	planning := WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "research-lead", TeamID: "research", AgentRole: "lead", RunID: "run-7", PlanningOnly: true})
	registry := &InternalToolRegistry{}
	_, err := registry.handleHandOff(agentInvocation("research-lead", "research", "lead", "run-7"), handOffArgs())
	requireHandoffCode(t, err, "handoff_unavailable")
	registry.SetHandoffRecorder(&recordingHandoffRecorder{})
	_, err = registry.handleHandOff(planning, handOffArgs())
	requireHandoffCode(t, err, "handoff_needs_approval")
	if !blocksProposalPlanningTool("hand_off") || !blocksProposalPlanningTool("read_handoff_input") {
		t.Fatal("handoff tools must never execute during proposal planning")
	}
}

func TestReadHandoffInputScopesReaderToInvocationTeam(t *testing.T) {
	recorder := &recordingHandoffRecorder{}
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(recorder)
	out, err := registry.handleReadHandoffInput(agentInvocation("marketing-lead", "marketing", "lead", "run-7"), map[string]any{
		"handoff_id": "h-1", "artifact_id": handoffArtifactID, "team_id": "research",
	})
	if err != nil || !strings.Contains(out, "facts") {
		t.Fatalf("read_handoff_input = %q, %v", out, err)
	}
	if got := recorder.reads[0]; got.ReaderTeamID != "marketing" || got.ReaderAgentID != "marketing-lead" {
		t.Fatalf("reader identity = %#v", got)
	}
	recorder.err = &HandoffBlockedError{Code: "handoff_input_not_in_scope", Detail: "not yours"}
	_, err = registry.handleReadHandoffInput(agentInvocation("sales-lead", "sales", "lead", "run-7"), map[string]any{"handoff_id": "h-1", "artifact_id": handoffArtifactID})
	requireHandoffCode(t, err, "handoff_input_not_in_scope")
}

// The dispatched command must carry the correlation the receiving team uses
// for its acceptance and result signals, or projection drops them.
func TestHandoffCommandCorrelatesReceiverSignals(t *testing.T) {
	raw, err := HandoffCommandEnvelope(HandoffCommand{
		HandoffID: "h-1", WorkItemID: "22222222-2222-4222-8222-222222222222", TargetTeamID: "marketing",
		SourceTeamID: "research", RunID: "33333333-3333-4333-8333-333333333333", Note: "Use the fact sheet.",
		ExpectedAction: "use", IdempotencyKey: "team-handoff:h-1",
		Inputs: []HandoffInputRef{{ArtifactID: handoffArtifactID, Title: "Fact sheet"}},
	})
	if err != nil {
		t.Fatalf("envelope: %v", err)
	}
	correlation := extractTeamCommandCorrelation("marketing", raw, nil)
	if correlation == nil || correlation.WorkItemID != "22222222-2222-4222-8222-222222222222" ||
		correlation.TeamID != "marketing" || correlation.RunID != "33333333-3333-4333-8333-333333333333" ||
		correlation.commandKey() != "team-handoff:h-1" {
		t.Fatalf("receiver correlation = %#v", correlation)
	}
	if !strings.Contains(string(raw), "read_handoff_input") || !strings.Contains(string(raw), handoffArtifactID) {
		t.Fatalf("command does not tell the receiver how to read its inputs: %s", raw)
	}
}

// S7: an agent that did not declare read_handoff_input is refused before
// the executor is reached.
func TestUndeclaredHandoffToolIsRefused(t *testing.T) {
	agent, exec, _, emitter := denialTestAgent(`{"tool_call":{"name":"read_handoff_input","arguments":{"handoff_id":"h-1","artifact_id":"` + handoffArtifactID + `"}}}`)
	agent.processMessageStructuredWithPosture("Read the handoff.", nil, false)
	if exec.findCalls != 0 || exec.callCalls != 0 {
		t.Fatalf("undeclared handoff tool reached the executor: find=%d call=%d", exec.findCalls, exec.callCalls)
	}
	if payload := emitter.denied(t); payload["tool"] != "read_handoff_input" {
		t.Fatalf("tool.denied payload = %#v", payload)
	}
}

func TestHandoffToolsAreRegisteredAndRiskClassified(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	for _, name := range []string{"hand_off", "read_handoff_input"} {
		tool := registry.Get(name)
		if tool == nil || tool.Manifest == nil {
			t.Fatalf("%s registered=%v with manifest=%v", name, tool != nil, tool != nil && tool.Manifest != nil)
		}
	}
}

func TestHandOffDeniesCoderSpecialist(t *testing.T) {
	recorder := &recordingHandoffRecorder{}
	registry := &InternalToolRegistry{}
	registry.SetHandoffRecorder(recorder)
	_, err := registry.handleHandOff(agentInvocation("research-coder", "research", "coder", "run-7"), handOffArgs())
	requireHandoffCode(t, err, "handoff_not_team_lead")
	if actor := exchangeActorForInvocation(agentInvocation("research-coder", "research", "coder", "run-7"), nil); actor.Role != "specialist" {
		t.Fatalf("coder exchange actor = %s, want specialist", actor.Role)
	}
	if len(recorder.records) != 0 {
		t.Fatal("coder specialist reached the recorder")
	}
}
