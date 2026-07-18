package inputs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestNormalizeNATSMessagePreservesSignalMetadata(t *testing.T) {
	raw, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindSensor,
		"warehouse.temperature",
		protocol.PayloadKindTelemetry,
		"run-1",
		"ops-team",
		"sensor-agent",
		[]byte(`{"channel_key":"temperature","value":72}`),
	)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	event := NormalizeNATSMessage(Source{
		ID: "warehouse-sensor", AdapterKind: AdapterSensor, TenantID: "default",
	}, "swarm.global.input.warehouse-sensor", raw)
	if event.SourceKind != string(protocol.SourceKindSensor) || event.PayloadKind != string(protocol.PayloadKindTelemetry) {
		t.Fatalf("event metadata = %+v", event)
	}
	if event.ChannelKey != "temperature" || event.RunID != "run-1" || event.TeamID != "ops-team" || event.AgentID != "sensor-agent" {
		t.Fatalf("event routing metadata = %+v", event)
	}
	if event.SourceTimestamp == nil || event.PayloadHash == "" {
		t.Fatalf("event proof metadata = %+v", event)
	}
}

func TestRecordNATSMessageIgnoresUnregisteredSubject(t *testing.T) {
	svc := NewService()
	_, _, matched, err := svc.RecordNATSMessage(context.Background(), "swarm.global.input.unknown", []byte(`{"ok":true}`))
	if err != nil {
		t.Fatalf("RecordNATSMessage: %v", err)
	}
	if matched {
		t.Fatalf("unknown subject should be ignored")
	}
}

func TestRecordBusMessageUsesRegisteredSourceAndHeaders(t *testing.T) {
	svc := NewService()
	_, err := svc.Add(context.Background(), SourceInput{
		ID:                    "media-events",
		Name:                  "Media Events",
		AdapterKind:           AdapterWebhook,
		AllowedIngressSubject: "swarm.global.input.media-events",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	headers := map[string][]string{
		HeaderChannelKey:    {"render-status"},
		HeaderRunID:         {"run-42"},
		HeaderSourceKind:    {string(protocol.SourceKindAutomationTrigger)},
		HeaderPayloadKind:   {string(protocol.PayloadKindStatus)},
		HeaderSourceChannel: {"render.worker"},
	}
	source, event, matched, err := svc.RecordBusMessage(
		context.Background(),
		"swarm.global.input.media-events",
		[]byte(`{"channel_key":"payload-channel","state":"rendering"}`),
		headers,
	)
	if err != nil {
		t.Fatalf("RecordBusMessage: %v", err)
	}
	if !matched || source.ID != "media-events" {
		t.Fatalf("matched source = %v %+v", matched, source)
	}
	if event.ChannelKey != "render-status" || event.RunID != "run-42" {
		t.Fatalf("event routing metadata = %+v", event)
	}
	if event.SourceKind != string(protocol.SourceKindAutomationTrigger) ||
		event.PayloadKind != string(protocol.PayloadKindStatus) ||
		event.SourceChannel != "render.worker" {
		t.Fatalf("event source metadata = %+v", event)
	}
}

func TestStoreRecordEventUpdatesLatestForAppendLatest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store := NewStore(db)
	now := time.Date(2026, 7, 18, 10, 30, 0, 0, time.UTC)
	eventID := "11111111-1111-1111-1111-111111111111"
	payload := json.RawMessage(`{"channel_key":"metrics","value":42}`)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO input_source_events").
		WithArgs(
			"api-source", "metrics", string(payload), "hash-1", sqlmock.AnyArg(),
			"run-1", "team-1", "agent-1", string(protocol.SourceKindWebAPI),
			"swarm.global.input.api-source", string(protocol.PayloadKindEvent), "default",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"event_id", "source_id", "channel_key", "payload", "payload_hash",
			"source_timestamp", "received_at", "run_id", "team_id", "agent_id",
			"source_kind", "source_channel", "payload_kind", "tenant_id",
		}).AddRow(
			eventID, "api-source", "metrics", payload, "hash-1", now,
			now, "run-1", "team-1", "agent-1", string(protocol.SourceKindWebAPI),
			"swarm.global.input.api-source", string(protocol.PayloadKindEvent), "default",
		))
	mock.ExpectExec("INSERT INTO input_source_latest").
		WithArgs("api-source", "metrics", eventID, string(payload), now, sqlmock.AnyArg(), "default").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	recorded, err := store.RecordEvent(context.Background(), Source{
		ID: "api-source", Name: "API Source", BufferMode: BufferAppendLatest, TenantID: "default",
	}, IngestEvent{
		SourceID: "api-source", ChannelKey: "metrics", Payload: payload, PayloadHash: "hash-1",
		SourceTimestamp: &now, RunID: "run-1", TeamID: "team-1", AgentID: "agent-1",
		SourceKind: string(protocol.SourceKindWebAPI), SourceChannel: "swarm.global.input.api-source",
		PayloadKind: string(protocol.PayloadKindEvent), TenantID: "default",
	})
	if err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if recorded.EventID != eventID || recorded.ChannelKey != "metrics" || recorded.PayloadHash != "hash-1" {
		t.Fatalf("recorded = %+v", recorded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStoreRecordEventUpsertsWindowForWindowedRollup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store := NewStore(db)
	now := time.Date(2026, 7, 18, 10, 31, 0, 0, time.UTC)
	payload := json.RawMessage(`{"channel_key":"frames","count":12}`)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO input_source_events").
		WillReturnRows(sqlmock.NewRows([]string{
			"event_id", "source_id", "channel_key", "payload", "payload_hash",
			"source_timestamp", "received_at", "run_id", "team_id", "agent_id",
			"source_kind", "source_channel", "payload_kind", "tenant_id",
		}).AddRow(
			"22222222-2222-2222-2222-222222222222", "camera", "frames", payload,
			"hash-2", now, now, "", "", "", string(protocol.SourceKindIoT),
			"swarm.global.input.camera", string(protocol.PayloadKindTelemetry), "default",
		))
	mock.ExpectExec("INSERT INTO input_source_windows").
		WithArgs("camera", "frames", "20260718T1031Z", sqlmock.AnyArg(), string(payload), sqlmock.AnyArg(), now, "default").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err = store.RecordEvent(context.Background(), Source{
		ID: "camera", Name: "Camera", BufferMode: BufferWindowedRollup, TenantID: "default",
	}, IngestEvent{
		SourceID: "camera", ChannelKey: "frames", Payload: payload, PayloadHash: "hash-2",
		SourceTimestamp: &now, SourceKind: string(protocol.SourceKindIoT),
		SourceChannel: "swarm.global.input.camera", PayloadKind: string(protocol.PayloadKindTelemetry),
		TenantID: "default",
	})
	if err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
