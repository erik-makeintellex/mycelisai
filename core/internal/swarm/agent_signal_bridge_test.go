package swarm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestAgentPublishToolBusSignal_StatusChannelForMCP(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	agent := &Agent{
		Manifest: protocol.AgentManifest{ID: "agent-alpha"},
		TeamID:   "alpha",
		nc:       nc,
		runID:    "run-123",
	}

	statusCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.alpha.signal.status", func(msg *nats.Msg) {
		statusCh <- msg
	}); err != nil {
		t.Fatalf("subscribe status: %v", err)
	}
	nc.Flush()

	agent.publishToolBusSignal(protocol.PayloadKindStatus, protocol.SourceKindMCP, map[string]any{
		"state": "invoked",
		"tool":  "read_file",
	})

	select {
	case msg := <-statusCh:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			t.Fatalf("decode envelope: %v", err)
		}
		if env.Meta.SourceKind != protocol.SourceKindMCP {
			t.Fatalf("source_kind = %q, want %q", env.Meta.SourceKind, protocol.SourceKindMCP)
		}
		if env.Meta.SourceChannel != "swarm.team.alpha.internal.trigger" {
			t.Fatalf("source_channel = %q, want swarm.team.alpha.internal.trigger", env.Meta.SourceChannel)
		}
		if env.Meta.PayloadKind != protocol.PayloadKindStatus {
			t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, protocol.PayloadKindStatus)
		}
		if env.Meta.RunID != "run-123" {
			t.Fatalf("run_id = %q, want run-123", env.Meta.RunID)
		}
		if env.Meta.TeamID != "alpha" {
			t.Fatalf("team_id = %q, want alpha", env.Meta.TeamID)
		}
		if env.Meta.AgentID != "agent-alpha" {
			t.Fatalf("agent_id = %q, want agent-alpha", env.Meta.AgentID)
		}
		if len(env.Payload) == 0 {
			t.Fatal("expected structured payload")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for status signal")
	}
}

func TestAgentPublishToolBusSignal_ResultChannelForMCP(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	agent := &Agent{
		Manifest: protocol.AgentManifest{ID: "agent-beta"},
		TeamID:   "beta",
		nc:       nc,
		runID:    "run-987",
	}

	resultCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.beta.signal.result", func(msg *nats.Msg) {
		resultCh <- msg
	}); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	agent.publishToolBusSignal(protocol.PayloadKindResult, protocol.SourceKindMCP, map[string]any{
		"state":          "completed",
		"tool":           "web_search",
		"result_preview": "ok",
	})

	select {
	case msg := <-resultCh:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			t.Fatalf("decode envelope: %v", err)
		}
		if env.Meta.SourceKind != protocol.SourceKindMCP {
			t.Fatalf("source_kind = %q, want %q", env.Meta.SourceKind, protocol.SourceKindMCP)
		}
		if env.Meta.PayloadKind != protocol.PayloadKindResult {
			t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, protocol.PayloadKindResult)
		}
		if env.Meta.RunID != "run-987" {
			t.Fatalf("run_id = %q, want run-987", env.Meta.RunID)
		}
		if env.Meta.TeamID != "beta" {
			t.Fatalf("team_id = %q, want beta", env.Meta.TeamID)
		}
		if env.Meta.AgentID != "agent-beta" {
			t.Fatalf("agent_id = %q, want agent-beta", env.Meta.AgentID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for result signal")
	}
}

func TestAgentPublishToolBusSignal_PersistsLatestCheckpoint(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	memSvc := memory.NewServiceWithDB(db)
	reg := NewInternalToolRegistry(InternalToolDeps{Mem: memSvc})

	channelKey := "signal.latest.swarm.team.alpha.signal.status"
	mock.ExpectExec("DELETE FROM temp_memory_channels").
		WithArgs("default", channelKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", channelKey, "agent-alpha", sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("checkpoint-agent-1"))

	agent := &Agent{
		Manifest:      protocol.AgentManifest{ID: "agent-alpha"},
		TeamID:        "alpha",
		nc:            nc,
		runID:         "run-123",
		ctx:           context.Background(),
		internalTools: reg,
	}

	agent.publishToolBusSignal(protocol.PayloadKindStatus, protocol.SourceKindMCP, map[string]any{
		"state": "invoked",
		"tool":  "filesystem/read_file",
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
