package overseer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// TaskState tracks the desired vs actual state for a single DAG node.
// The Resolved channel is closed when ActualState == DesiredState.
type TaskState struct {
	TaskID       string
	DesiredState string
	ActualState  string
	Resolved     chan struct{}
}

// GovernanceCallback is invoked when an envelope's TrustScore falls below
// the AutoExecuteThreshold. The callback should route the envelope to
// Zone D (Deliverables Tray) for human review.
type GovernanceCallback func(env *protocol.CTSEnvelope)

// Engine is the Overseer's cybernetic reconciliation loop.
// It enforces Zero-Trust Actuation: no DAG node advances until
// a valid CTSEnvelope confirms ActualState == DesiredState.
// Phase 5.2: The Governance Valve adds trust-based routing —
// high-trust envelopes auto-execute, low-trust halts for Zone D.
type Engine struct {
	nc     *nats.Conn
	tasks  map[string]*TaskState
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Trust Economy — Governance Valve
	autoExecuteThreshold float64
	thresholdMu          sync.RWMutex
	governanceCallback   GovernanceCallback
}

// NewEngine creates a new Overseer instance bound to a NATS connection.
// Default AutoExecuteThreshold is 0.7 — envelopes with TrustScore >= 0.7
// bypass human approval.
func NewEngine(nc *nats.Conn) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		nc:                   nc,
		tasks:                make(map[string]*TaskState),
		ctx:                  ctx,
		cancel:               cancel,
		autoExecuteThreshold: 0.7,
	}
}

// SetGovernanceCallback registers the function called when a low-trust
// envelope is halted. Typically this broadcasts to SSE for Zone D display.
func (e *Engine) SetGovernanceCallback(cb GovernanceCallback) {
	e.governanceCallback = cb
}

// SetAutoExecuteThreshold updates the trust threshold (0.0–1.0).
// Envelopes with TrustScore >= threshold bypass human approval.
func (e *Engine) SetAutoExecuteThreshold(t float64) {
	e.thresholdMu.Lock()
	defer e.thresholdMu.Unlock()
	e.autoExecuteThreshold = t
}

// GetAutoExecuteThreshold returns the current trust threshold.
func (e *Engine) GetAutoExecuteThreshold() float64 {
	e.thresholdMu.RLock()
	defer e.thresholdMu.RUnlock()
	return e.autoExecuteThreshold
}

// Start brings the Overseer online. It subscribes to:
//   - swarm.mission.task      (task commands issued by the DAG)
//   - swarm.team.*.telemetry  (CTS envelopes from agents/sensors)
func (e *Engine) Start() error {
	_, err := e.nc.Subscribe(protocol.TopicMissionTask, e.handleTaskCommand)
	if err != nil {
		return fmt.Errorf("overseer: subscribe mission task: %w", err)
	}

	_, err = e.nc.Subscribe(protocol.TopicTeamTelemetryWild, e.handleTelemetry)
	if err != nil {
		return fmt.Errorf("overseer: subscribe telemetry: %w", err)
	}

	log.Println("Overseer Engine Online. Zero-Trust Actuation Loop active.")
	return nil
}

// IssueTask publishes a task command onto swarm.mission.task and HALTS
// the DAG until a CTSEnvelope with a matching TraceID confirms that
// ActualState == desiredState, or the context expires.
func (e *Engine) IssueTask(ctx context.Context, taskID, desiredState string, payload json.RawMessage) error {
	// 1. Register pending reconciliation
	ts := &TaskState{
		TaskID:       taskID,
		DesiredState: desiredState,
		Resolved:     make(chan struct{}),
	}

	e.mu.Lock()
	e.tasks[taskID] = ts
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.tasks, taskID)
		e.mu.Unlock()
	}()

	// 2. Publish task command as a CTS envelope
	cmd := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "overseer",
			Timestamp:  time.Now(),
			TraceID:    taskID,
		},
		SignalType: protocol.SignalTelemetry,
		Payload:    payload,
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("overseer: marshal task command: %w", err)
	}

	if err := e.nc.Publish(protocol.TopicMissionTask, data); err != nil {
		return fmt.Errorf("overseer: publish task: %w", err)
	}
	e.nc.Flush()

	log.Printf("Overseer: Task [%s] issued. DAG HALTED. Awaiting state: %s", taskID, desiredState)

	// 3. HALT — block until reconciliation or timeout
	select {
	case <-ts.Resolved:
		log.Printf("Overseer: Task [%s] reconciled. DAG ADVANCING.", taskID)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("overseer: task [%s] timed out waiting for state %q", taskID, desiredState)
	}
}

// handleTaskCommand logs task commands for observability.
// External agents/runners subscribe to this topic to pick up work.
func (e *Engine) handleTaskCommand(msg *nats.Msg) {
	log.Printf("Overseer: Task command received on %s (%d bytes)", msg.Subject, len(msg.Data))
}

// handleTelemetry validates incoming CTS envelopes and reconciles state.
// Non-conforming messages are REJECTED (Zero-Trust).
// Phase 5.2: The Governance Valve intercepts low-trust envelopes before
// state reconciliation, routing them to Zone D for human review.
func (e *Engine) handleTelemetry(msg *nats.Msg) {
	// 1. Enforce CTS schema — reject non-conforming
	env, err := protocol.ValidateTelemetryMessage(msg.Data)
	if err != nil {
		log.Printf("Overseer: REJECTED telemetry on %s: %v", msg.Subject, err)
		return
	}

	// 2. GOVERNANCE VALVE — Trust threshold check
	// If the envelope carries an explicit TrustScore below the threshold,
	// halt the DAG and route to Zone D for human approval.
	// Unscored envelopes (TrustScore == 0) pass through to preserve
	// backward compatibility with pre-5.2 agents.
	threshold := e.GetAutoExecuteThreshold()
	if env.HasTrustScore() && env.TrustScore < threshold {
		log.Printf("Overseer: GOVERNANCE HALT — TrustScore %.2f < threshold %.2f for %s",
			env.TrustScore, threshold, env.Meta.SourceNode)
		if e.governanceCallback != nil {
			e.governanceCallback(env)
		}
		return // Do NOT advance DAG — await human approval
	}

	// 3. Check if this envelope resolves a pending task
	traceID := env.Meta.TraceID
	if traceID == "" {
		return
	}

	e.mu.RLock()
	ts, exists := e.tasks[traceID]
	e.mu.RUnlock()

	if !exists {
		return
	}

	// 4. Extract actual state from payload
	var stateReport struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(env.Payload, &stateReport); err != nil {
		log.Printf("Overseer: Failed to parse state from payload: %v", err)
		return
	}

	// 5. Reconcile: ActualState == DesiredState?
	ts.ActualState = stateReport.State
	if ts.ActualState == ts.DesiredState {
		close(ts.Resolved)
	} else {
		log.Printf("Overseer: Task [%s] state mismatch: actual=%q desired=%q — still halted",
			ts.TaskID, ts.ActualState, ts.DesiredState)
	}
}

// PendingTasks returns a snapshot of task IDs awaiting reconciliation.
func (e *Engine) PendingTasks() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ids := make([]string, 0, len(e.tasks))
	for id := range e.tasks {
		ids = append(ids, id)
	}
	return ids
}

// Shutdown stops the Overseer loop.
func (e *Engine) Shutdown() {
	e.cancel()
}
