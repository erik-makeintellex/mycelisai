package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestHandleDelegateTask_PublishesToInternalCommand(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	done := make(chan []byte, 1)
	if _, err := nc.Subscribe("swarm.team.alpha.internal.command", func(msg *nats.Msg) {
		done <- msg.Data
	}); err != nil {
		t.Fatalf("subscribe command: %v", err)
	}
	nc.Flush()

	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		RunID:         "run-123",
		TeamID:        "alpha",
		AgentID:       "soma-admin",
		SourceKind:    protocol.SourceKindSystem,
		SourceChannel: "swarm.team.alpha.internal.trigger",
		PayloadKind:   protocol.PayloadKindCommand,
	})
	out, err := reg.handleDelegateTask(ctx, map[string]any{
		"team_id": "alpha",
		"task":    "inspect gate state",
	})
	if err != nil {
		t.Fatalf("delegate_task error: %v", err)
	}
	if !strings.Contains(out, "alpha") {
		t.Fatalf("unexpected delegate output: %s", out)
	}

	select {
	case got := <-done:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(got, &env); err != nil {
			t.Fatalf("decode signal envelope: %v", err)
		}
		if env.Meta.SourceKind != protocol.SourceKindInternalTool {
			t.Fatalf("source_kind = %q, want %q", env.Meta.SourceKind, protocol.SourceKindInternalTool)
		}
		if env.Meta.SourceChannel != "internal_tool.delegate_task" {
			t.Fatalf("source_channel = %q", env.Meta.SourceChannel)
		}
		if env.Meta.PayloadKind != protocol.PayloadKindCommand {
			t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, protocol.PayloadKindCommand)
		}
		if env.Meta.TeamID != "alpha" {
			t.Fatalf("team_id = %q, want alpha", env.Meta.TeamID)
		}
		if env.Meta.RunID != "run-123" {
			t.Fatalf("run_id = %q, want run-123", env.Meta.RunID)
		}
		if env.Meta.AgentID != "soma-admin" {
			t.Fatalf("agent_id = %q, want soma-admin", env.Meta.AgentID)
		}
		if env.Text != "inspect gate state" {
			t.Fatalf("command text = %q", env.Text)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for delegated task")
	}
}

func TestHandlePublishSignal_WrapsCanonicalStatusSubject(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	done := make(chan []byte, 1)
	if _, err := nc.Subscribe("swarm.team.alpha.signal.status", func(msg *nats.Msg) {
		done <- msg.Data
	}); err != nil {
		t.Fatalf("subscribe status: %v", err)
	}
	nc.Flush()

	out, err := reg.handlePublishSignal(context.Background(), map[string]any{
		"subject": "swarm.team.alpha.signal.status",
		"message": `{"phase":"sync","state":"running"}`,
	})
	if err != nil {
		t.Fatalf("publish_signal error: %v", err)
	}
	if !strings.Contains(out, "swarm.team.alpha.signal.status") {
		t.Fatalf("unexpected publish output: %s", out)
	}

	select {
	case got := <-done:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(got, &env); err != nil {
			t.Fatalf("decode signal envelope: %v", err)
		}
		if env.Meta.SourceKind != protocol.SourceKindInternalTool {
			t.Fatalf("source_kind = %q, want %q", env.Meta.SourceKind, protocol.SourceKindInternalTool)
		}
		if env.Meta.SourceChannel != "internal_tool.publish_signal" {
			t.Fatalf("source_channel = %q", env.Meta.SourceChannel)
		}
		if env.Meta.PayloadKind != protocol.PayloadKindStatus {
			t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, protocol.PayloadKindStatus)
		}
		if env.Meta.TeamID != "alpha" {
			t.Fatalf("team_id = %q, want alpha", env.Meta.TeamID)
		}
		if string(env.Payload) != `{"phase":"sync","state":"running"}` {
			t.Fatalf("payload = %s", string(env.Payload))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for wrapped status signal")
	}
}

func TestHandlePublishSignal_PrivateReferenceAndCheckpoint(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	memSvc := memory.NewServiceWithDB(db)
	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc, Mem: memSvc})

	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())

	channelKey := "team.alpha.private.files"
	mock.ExpectExec("DELETE FROM temp_memory_channels").
		WithArgs("default", channelKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", channelKey, "soma-admin", sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("checkpoint-1"))

	done := make(chan []byte, 1)
	if _, err := nc.Subscribe("swarm.team.alpha.signal.result", func(msg *nats.Msg) {
		done <- msg.Data
	}); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		RunID:   "run-901",
		TeamID:  "alpha",
		AgentID: "soma-admin",
	})
	out, err := reg.handlePublishSignal(ctx, map[string]any{
		"subject":      "swarm.team.alpha.signal.result",
		"message":      "write completed for private draft",
		"privacy_mode": "reference",
		"channel_key":  channelKey,
		"file_path":    "drafts/plan.md",
	})
	if err != nil {
		t.Fatalf("publish_signal private reference error: %v", err)
	}
	if !strings.Contains(out, "Latest channel checkpoint") {
		t.Fatalf("expected checkpoint acknowledgement, got: %s", out)
	}

	select {
	case got := <-done:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(got, &env); err != nil {
			t.Fatalf("decode signal envelope: %v", err)
		}
		if env.Meta.PayloadKind != protocol.PayloadKindResult {
			t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, protocol.PayloadKindResult)
		}

		var payload map[string]any
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("decode private payload: %v", err)
		}
		if payload["mode"] != "private_reference" {
			t.Fatalf("mode = %v, want private_reference", payload["mode"])
		}
		if payload["channel_key"] != channelKey {
			t.Fatalf("channel_key = %v, want %s", payload["channel_key"], channelKey)
		}
		fileRef, ok := payload["file_ref"].(map[string]any)
		if !ok {
			t.Fatalf("expected file_ref map, got %T", payload["file_ref"])
		}
		if fileRef["path"] != "drafts/plan.md" {
			t.Fatalf("file_ref.path = %v, want drafts/plan.md", fileRef["path"])
		}
		if _, exposed := payload["message"]; exposed {
			t.Fatalf("private payload should not expose full message directly: %+v", payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for private reference signal")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}

func TestHandleReadSignals_LatestOnlyReturnsCheckpoint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	memSvc := memory.NewServiceWithDB(db)
	reg := NewInternalToolRegistry(InternalToolDeps{Mem: memSvc})

	channelKey := "team.alpha.private.files"
	now := time.Now().UTC()

	mock.ExpectQuery("SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata").
		WithArgs("default", channelKey, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "channel_key", "owner_agent_id", "content", "metadata", "expires_at", "created_at", "updated_at",
		}).AddRow(
			"checkpoint-9",
			"default",
			channelKey,
			"soma-admin",
			`{"message":"write completed for private draft","file_ref":{"kind":"workspace_file","path":"drafts/plan.md"}}`,
			[]byte(`{"subject":"swarm.team.alpha.signal.result","privacy_mode":"reference"}`),
			nil,
			now,
			now,
		))

	out, err := reg.handleReadSignals(context.Background(), map[string]any{
		"latest_only": true,
		"channel_key": channelKey,
	})
	if err != nil {
		t.Fatalf("read_signals latest_only error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("decode latest checkpoint output: %v", err)
	}
	if parsed["mode"] != "latest_checkpoint" {
		t.Fatalf("mode = %v, want latest_checkpoint", parsed["mode"])
	}
	if parsed["channel_key"] != channelKey {
		t.Fatalf("channel_key = %v, want %s", parsed["channel_key"], channelKey)
	}
	latest, ok := parsed["latest"].(map[string]any)
	if !ok {
		t.Fatalf("expected latest map, got %T", parsed["latest"])
	}
	content, ok := latest["content"].(map[string]any)
	if !ok {
		t.Fatalf("expected latest.content map, got %T", latest["content"])
	}
	if content["message"] != "write completed for private draft" {
		t.Fatalf("latest.content.message = %v", content["message"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
