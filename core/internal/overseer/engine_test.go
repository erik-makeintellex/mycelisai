package overseer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// startTestNATS spins up an embedded NATS server for isolated testing.
func startTestNATS(t *testing.T) (*natsserver.Server, *nats.Conn) {
	t.Helper()
	opts := &natsserver.Options{Port: -1}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		t.Fatalf("nats connect: %v", err)
	}
	return srv, nc
}

// TestStateReconciliation verifies that the DAG HALTS until a valid
// CTSEnvelope with matching state arrives, then ADVANCES.
func TestStateReconciliation(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	engine := NewEngine(nc)
	if err := engine.Start(); err != nil {
		t.Fatalf("engine start: %v", err)
	}
	defer engine.Shutdown()

	taskID := "test-task-001"
	desiredState := "completed"

	// Simulate agent: after 2 seconds, publish a valid CTS envelope resolving the task
	go func() {
		time.Sleep(2 * time.Second)

		payload, _ := json.Marshal(map[string]string{"state": desiredState})
		env := protocol.CTSEnvelope{
			Meta: protocol.CTSMeta{
				SourceNode: "test-agent",
				Timestamp:  time.Now(),
				TraceID:    taskID,
			},
			SignalType: protocol.SignalTaskComplete,
			Payload:    payload,
		}
		data, _ := json.Marshal(env)
		nc.Publish("swarm.team.test.telemetry", data)
		nc.Flush()
	}()

	// Issue task with 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskPayload, _ := json.Marshal(map[string]string{"command": "execute"})

	start := time.Now()
	err := engine.IssueTask(ctx, taskID, desiredState, taskPayload)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("IssueTask failed: %v", err)
	}

	// Assert: DAG only advanced AFTER the ~2s delay (when envelope arrived)
	if elapsed < 1900*time.Millisecond {
		t.Errorf("DAG advanced too early: %v (expected >= ~2s)", elapsed)
	}
	if elapsed > 4*time.Second {
		t.Errorf("DAG took too long: %v (expected ~2s)", elapsed)
	}

	t.Logf("State reconciliation completed in %v", elapsed)
}

// TestStateReconciliation_Timeout verifies that the DAG does NOT advance
// if no valid CTSEnvelope arrives within the timeout.
func TestStateReconciliation_Timeout(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	engine := NewEngine(nc)
	if err := engine.Start(); err != nil {
		t.Fatalf("engine start: %v", err)
	}
	defer engine.Shutdown()

	// Issue task with 500ms timeout — no envelope will arrive
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	taskPayload, _ := json.Marshal(map[string]string{"command": "execute"})

	err := engine.IssueTask(ctx, "timeout-task", "completed", taskPayload)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	t.Logf("Timeout correctly enforced: %v", err)
}

// TestStateReconciliation_StateMismatch verifies that a CTS envelope with
// the wrong state does NOT advance the DAG.
func TestStateReconciliation_StateMismatch(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	engine := NewEngine(nc)
	if err := engine.Start(); err != nil {
		t.Fatalf("engine start: %v", err)
	}
	defer engine.Shutdown()

	taskID := "mismatch-task"

	// Publish envelope with wrong state after 200ms
	go func() {
		time.Sleep(200 * time.Millisecond)
		payload, _ := json.Marshal(map[string]string{"state": "running"}) // wrong state
		env := protocol.CTSEnvelope{
			Meta: protocol.CTSMeta{
				SourceNode: "test-agent",
				Timestamp:  time.Now(),
				TraceID:    taskID,
			},
			SignalType: protocol.SignalTelemetry,
			Payload:    payload,
		}
		data, _ := json.Marshal(env)
		nc.Publish("swarm.team.test.telemetry", data)
		nc.Flush()
	}()

	// Should timeout because state never matches
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	taskPayload, _ := json.Marshal(map[string]string{"command": "execute"})
	err := engine.IssueTask(ctx, taskID, "completed", taskPayload)
	if err == nil {
		t.Fatal("expected timeout error due to state mismatch, got nil")
	}

	t.Logf("State mismatch correctly prevented advancement: %v", err)
}

// TestCTSEnvelope_Validation covers the CTS schema enforcement.
func TestCTSEnvelope_Validation(t *testing.T) {
	tests := []struct {
		name    string
		env     protocol.CTSEnvelope
		wantErr bool
	}{
		{
			name: "valid envelope",
			env: protocol.CTSEnvelope{
				Meta:       protocol.CTSMeta{SourceNode: "agent-1", Timestamp: time.Now()},
				SignalType: protocol.SignalTelemetry,
				Payload:    json.RawMessage(`{"key": "value"}`),
			},
			wantErr: false,
		},
		{
			name: "missing source_node",
			env: protocol.CTSEnvelope{
				Meta:       protocol.CTSMeta{Timestamp: time.Now()},
				SignalType: protocol.SignalTelemetry,
				Payload:    json.RawMessage(`{"key": "value"}`),
			},
			wantErr: true,
		},
		{
			name: "missing timestamp",
			env: protocol.CTSEnvelope{
				Meta:       protocol.CTSMeta{SourceNode: "agent-1"},
				SignalType: protocol.SignalTelemetry,
				Payload:    json.RawMessage(`{"key": "value"}`),
			},
			wantErr: true,
		},
		{
			name: "missing signal_type",
			env: protocol.CTSEnvelope{
				Meta:    protocol.CTSMeta{SourceNode: "agent-1", Timestamp: time.Now()},
				Payload: json.RawMessage(`{"key": "value"}`),
			},
			wantErr: true,
		},
		{
			name: "missing payload",
			env: protocol.CTSEnvelope{
				Meta:       protocol.CTSMeta{SourceNode: "agent-1", Timestamp: time.Now()},
				SignalType: protocol.SignalTelemetry,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCTSMiddleware_RejectsInvalid verifies the NATS-level middleware
// rejects malformed and schema-violating messages.
func TestCTSMiddleware_RejectsInvalid(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "garbage bytes",
			data:    []byte("not json at all"),
			wantErr: true,
		},
		{
			name:    "empty json object",
			data:    []byte(`{}`),
			wantErr: true,
		},
		{
			name: "valid CTS",
			data: func() []byte {
				env := protocol.CTSEnvelope{
					Meta:       protocol.CTSMeta{SourceNode: "x", Timestamp: time.Now()},
					SignalType: protocol.SignalTelemetry,
					Payload:    json.RawMessage(`{"v":1}`),
				}
				b, _ := json.Marshal(env)
				return b
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.ValidateTelemetryMessage(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTelemetryMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
