package swarm

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestTeam_ResponseDeliveryCarriesPendingWorkCorrelation(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:         "test-core",
		Name:       "Test Core",
		Type:       TeamTypeAction,
		Inputs:     []string{"swarm.team.test-core.internal.command"},
		Deliveries: []string{"swarm.team.test-core.signal.result"},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	internalTriggerCh := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		internalTriggerCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	resultCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.result", func(msg *nats.Msg) { resultCh <- msg }); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	const workID = "11111111-1111-1111-1111-111111111111"
	publishCorrelatedCommand(t, nc, workID, "run-9")
	select {
	case <-internalTriggerCh:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger")
	}

	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("The note is ready.")); err != nil {
		t.Fatalf("publish response: %v", err)
	}

	msg := waitForTeamSignal(t, resultCh, "correlated result")
	env, projected := decodeTeamSignalPayload(t, msg.Data)
	if env.Meta.RunID != "run-9" {
		t.Fatalf("run_id = %q, want run-9", env.Meta.RunID)
	}
	assertProjectedWorkOutput(t, projected, workID, "The note is ready.")
}

func TestTeam_StatusDeliveryCarriesOutputReadyCorrelation(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:         "test-core",
		Name:       "Test Core",
		Type:       TeamTypeAction,
		Inputs:     []string{"swarm.team.test-core.internal.command"},
		Deliveries: []string{"swarm.team.test-core.signal.status"},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	statusCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.status", func(msg *nats.Msg) { statusCh <- msg }); err != nil {
		t.Fatalf("subscribe status: %v", err)
	}
	internalTriggerCh := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		internalTriggerCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	nc.Flush()

	const workID = "11111111-1111-1111-1111-111111111111"
	publishCorrelatedCommand(t, nc, workID, "")
	select {
	case <-internalTriggerCh:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger")
	}
	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("Status-only note ready.")); err != nil {
		t.Fatalf("publish response: %v", err)
	}

	_, projected := decodeTeamSignalPayload(t, waitForTeamSignal(t, statusCh, "correlated status").Data)
	assertProjectedWorkOutput(t, projected, workID, "Status-only note ready.")
}

func TestTeam_ResponseDeliveryConsumesPendingCorrelationsFIFO(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:         "test-core",
		Name:       "Test Core",
		Type:       TeamTypeAction,
		Inputs:     []string{"swarm.team.test-core.internal.command"},
		Deliveries: []string{"swarm.team.test-core.signal.result"},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	internalTriggerCh := make(chan struct{}, 2)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		internalTriggerCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	resultCh := make(chan *nats.Msg, 2)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.result", func(msg *nats.Msg) { resultCh <- msg }); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	const firstWorkID = "11111111-1111-1111-1111-111111111111"
	const secondWorkID = "22222222-2222-2222-2222-222222222222"
	publishCorrelatedCommand(t, nc, firstWorkID, "")
	publishCorrelatedCommand(t, nc, secondWorkID, "")
	waitForInternalTriggers(t, internalTriggerCh, 2)

	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("first ready")); err != nil {
		t.Fatalf("publish first response: %v", err)
	}
	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("second ready")); err != nil {
		t.Fatalf("publish second response: %v", err)
	}

	_, first := decodeTeamSignalPayload(t, waitForTeamSignal(t, resultCh, "first correlated result").Data)
	_, second := decodeTeamSignalPayload(t, waitForTeamSignal(t, resultCh, "second correlated result").Data)
	assertProjectedWorkOutput(t, first, firstWorkID, "first ready")
	assertProjectedWorkOutput(t, second, secondWorkID, "second ready")
}

func TestTeam_ResponseDeliveryUsesExplicitCorrelationWithoutConsumingPending(t *testing.T) {
	team := NewTeam(&TeamManifest{ID: "test-core", Name: "Test Core"}, nil, nil, nil)
	const pendingWorkID = "11111111-1111-1111-1111-111111111111"
	const explicitWorkID = "22222222-2222-2222-2222-222222222222"
	team.rememberCommandCorrelation(teamCommandCorrelation{WorkItemID: pendingWorkID, TeamID: "test-core"})

	explicit := team.responseCommandCorrelation([]byte(`{"work_item_id":"` + explicitWorkID + `","team_id":"test-core","state":"running"}`))
	if explicit == nil || explicit.WorkItemID != explicitWorkID {
		t.Fatalf("explicit correlation = %+v, want %s", explicit, explicitWorkID)
	}
	pending := team.responseCommandCorrelation([]byte("plain follow-up"))
	if pending == nil || pending.WorkItemID != pendingWorkID {
		t.Fatalf("pending correlation = %+v, want %s", pending, pendingWorkID)
	}
}

func TestTeam_ResponseDeliveryConsumesMatchingExplicitCorrelation(t *testing.T) {
	team := NewTeam(&TeamManifest{ID: "test-core", Name: "Test Core"}, nil, nil, nil)
	const workID = "11111111-1111-1111-1111-111111111111"
	team.rememberCommandCorrelation(teamCommandCorrelation{WorkItemID: workID, TeamID: "test-core"})

	explicit := team.responseCommandCorrelation([]byte(`{"work_item_id":"` + workID + `","team_id":"test-core","state":"output_ready"}`))
	if explicit == nil || explicit.WorkItemID != workID {
		t.Fatalf("explicit correlation = %+v, want %s", explicit, workID)
	}
	if pending := team.consumeCommandCorrelation(); pending != nil {
		t.Fatalf("matching pending correlation was not consumed: %+v", pending)
	}
}

func TestTeamAgentResponsePayloadForTriggerCarriesDurableCorrelation(t *testing.T) {
	const workID = "11111111-1111-1111-1111-111111111111"
	trigger := []byte(`{"goal":"build the package","context":{"work_item_id":"` + workID + `","team_id":"delivery-team","run_id":"run-9"}}`)

	raw := teamAgentResponsePayloadForTrigger(ProcessResult{Text: "Package ready."}, trigger, "delivery-team")
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode response payload: %v", err)
	}
	if payload["work_item_id"] != workID || payload["team_id"] != "delivery-team" || payload["run_id"] != "run-9" {
		t.Fatalf("correlated response = %#v", payload)
	}
	if payload["state"] != string(protocol.TeamWorkStateOutputReady) {
		t.Fatalf("state = %v, want output_ready", payload["state"])
	}
}

func TestCorrelatedTeamResponsePayload_ProjectsBlockerAsDegraded(t *testing.T) {
	raw := []byte(`{"text":"Package draft written.","blocker":"Browser validation is incomplete.","artifacts":[{"type":"project_package"}]}`)
	correlation := &teamCommandCorrelation{
		WorkItemID: "11111111-1111-1111-1111-111111111111",
		TeamID:     "test-core",
		RunID:      "run-9",
	}

	projectedRaw := correlatedTeamResponsePayload(raw, correlation)
	var projected map[string]any
	if err := json.Unmarshal(projectedRaw, &projected); err != nil {
		t.Fatalf("decode projected response: %v", err)
	}
	if projected["state"] != string(protocol.TeamWorkStateDegraded) {
		t.Fatalf("state = %v, want degraded", projected["state"])
	}
	if projected["needs_operator"] != true {
		t.Fatalf("needs_operator = %v, want true", projected["needs_operator"])
	}
}

func TestTeamCommandCorrelationRejectsDuplicateIdempotencyKey(t *testing.T) {
	team := &Team{Manifest: &TeamManifest{ID: "test-core"}, seenCommandKeys: map[string]time.Time{}}
	correlation := teamCommandCorrelation{
		WorkItemID:     "11111111-1111-1111-1111-111111111111",
		TeamID:         "test-core",
		IdempotencyKey: "confirm-action:proof-1",
	}
	if !team.rememberCommandCorrelation(correlation) {
		t.Fatal("first command should be accepted")
	}
	if team.rememberCommandCorrelation(correlation) {
		t.Fatal("duplicate idempotency key should be rejected")
	}
}

func TestTeamCommandCorrelationSurvivesLongRunningWorker(t *testing.T) {
	team := &Team{Manifest: &TeamManifest{ID: "test-core"}, seenCommandKeys: map[string]time.Time{}}
	started := time.Now().UTC()
	correlation := teamCommandCorrelation{
		WorkItemID:     "11111111-1111-1111-1111-111111111111",
		TeamID:         "test-core",
		IdempotencyKey: "confirm-action:proof-long-running",
	}
	if !team.rememberCommandCorrelation(correlation) {
		t.Fatal("long-running command should be accepted")
	}
	if len(team.pendingCorrelations) != 1 {
		t.Fatalf("pending correlations = %d, want 1", len(team.pendingCorrelations))
	}
	if team.pendingCorrelations[0].ExpiresAt.Before(started.Add(15 * time.Minute)) {
		t.Fatalf("correlation expires before durable recovery deadline: %s", team.pendingCorrelations[0].ExpiresAt)
	}

	team.pruneExpiredCorrelationsLocked(started.Add(6 * time.Minute))
	if len(team.pendingCorrelations) != 1 {
		t.Fatal("six-minute worker response lost its durable correlation")
	}
	team.pruneExpiredCorrelationsLocked(started.Add(teamCommandCorrelationTTL + time.Second))
	if len(team.pendingCorrelations) != 0 {
		t.Fatal("expired worker correlation was not pruned")
	}
}

func publishCorrelatedCommand(t *testing.T, nc *nats.Conn, workID, runID string) {
	t.Helper()
	payload, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindWebAPI,
		"api.teams.work.ask",
		protocol.PayloadKindCommand,
		runID,
		"test-core",
		"soma-admin",
		[]byte(`{"goal":"write the note","context":{"work_item_id":"`+workID+`","team_id":"test-core"}}`),
	)
	if err != nil {
		t.Fatalf("wrap command payload: %v", err)
	}
	if err := nc.Publish("swarm.team.test-core.internal.command", payload); err != nil {
		t.Fatalf("publish command: %v", err)
	}
}

func waitForInternalTriggers(t *testing.T, ch <-chan struct{}, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		select {
		case <-ch:
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for internal trigger %d/%d", i+1, count)
		}
	}
}

func waitForTeamSignal(t *testing.T, ch <-chan *nats.Msg, label string) *nats.Msg {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %s", label)
	}
	return nil
}

func decodeTeamSignalPayload(t *testing.T, raw []byte) (protocol.SignalEnvelope, map[string]any) {
	t.Helper()
	var env protocol.SignalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode signal envelope: %v", err)
	}
	var projected map[string]any
	if err := json.Unmarshal(env.Payload, &projected); err != nil {
		t.Fatalf("decode projected payload: %v", err)
	}
	return env, projected
}

func assertProjectedWorkOutput(t *testing.T, projected map[string]any, workID, text string) {
	t.Helper()
	if projected["work_item_id"] != workID {
		t.Fatalf("work_item_id = %v", projected["work_item_id"])
	}
	if projected["state"] != "output_ready" {
		t.Fatalf("state = %v", projected["state"])
	}
	if projected["text"] != text {
		t.Fatalf("text = %v", projected["text"])
	}
}
