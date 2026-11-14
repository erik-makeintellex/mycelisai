package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildRunLifecycleStages(t *testing.T) {
	baseTime := time.Date(2026, 4, 13, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		events         []protocol.MissionEventEnvelope
		expectedStatus map[string]string
		expectedActor  string
	}{
		{
			name: "completed mission flow",
			events: []protocol.MissionEventEnvelope{
				{ID: "ev-1", RunID: "run-1", EventType: protocol.EventMissionStarted, SourceAgent: "soma", EmittedAt: baseTime},
				{ID: "ev-2", RunID: "run-1", EventType: protocol.EventTeamSpawned, SourceAgent: "soma", SourceTeam: "delivery-alpha", EmittedAt: baseTime.Add(1 * time.Second)},
				{ID: "ev-3", RunID: "run-1", EventType: protocol.EventToolInvoked, SourceAgent: "coder-1", Payload: map[string]any{"tool": "write_file"}, EmittedAt: baseTime.Add(2 * time.Second)},
				{ID: "ev-4", RunID: "run-1", EventType: protocol.EventToolCompleted, SourceAgent: "coder-1", Payload: map[string]any{"tool": "write_file"}, EmittedAt: baseTime.Add(3 * time.Second)},
				{ID: "ev-5", RunID: "run-1", EventType: protocol.EventArtifactCreated, SourceAgent: "coder-1", Payload: map[string]any{"package_slug": "wave8-pkg", "proof_id": "proof-1"}, EmittedAt: baseTime.Add(4 * time.Second)},
				{ID: "ev-6", RunID: "run-1", EventType: protocol.EventMissionCompleted, SourceAgent: "soma", EmittedAt: baseTime.Add(5 * time.Second)},
			},
			expectedStatus: map[string]string{
				"intent_received":       "completed",
				"proposal_generated":    "completed",
				"approval_required":     "completed",
				"approval_granted":      "completed",
				"mission_running":       "completed",
				"tool_action_execution": "completed",
				"package_generated":     "completed",
				"proof_generated":       "completed",
				"recovery_next_actions": "completed",
			},
			expectedActor: "soma",
		},
		{
			name: "approval denied and failed flow",
			events: []protocol.MissionEventEnvelope{
				{ID: "ev-1", RunID: "run-denied", EventType: protocol.EventMissionStarted, SourceAgent: "soma", EmittedAt: baseTime},
				{ID: "ev-2", RunID: "run-denied", EventType: protocol.EventType("approval.requested"), SourceAgent: "guard", Payload: map[string]any{"summary": "High-risk mutation requires approval"}, EmittedAt: baseTime.Add(1 * time.Second)},
				{ID: "ev-3", RunID: "run-denied", EventType: protocol.EventType("approval.denied"), SourceAgent: "operator", Payload: map[string]any{"reason": "Scope exceeds current sprint boundary"}, EmittedAt: baseTime.Add(2 * time.Second)},
			},
			expectedStatus: map[string]string{
				"intent_received":       "completed",
				"proposal_generated":    "completed",
				"approval_required":     "completed",
				"approval_granted":      "failed",
				"mission_running":       "pending",
				"tool_action_execution": "pending",
				"package_generated":     "pending",
				"proof_generated":       "failed",
				"recovery_next_actions": "failed",
			},
			expectedActor: "soma",
		},
		{
			name: "active tool invocation flow",
			events: []protocol.MissionEventEnvelope{
				{ID: "ev-1", RunID: "run-active", EventType: protocol.EventMissionStarted, SourceAgent: "soma", EmittedAt: baseTime},
				{ID: "ev-2", RunID: "run-active", EventType: protocol.EventTeamSpawned, SourceAgent: "soma", SourceTeam: "research-team", EmittedAt: baseTime.Add(1 * time.Second)},
				{ID: "ev-3", RunID: "run-active", EventType: protocol.EventToolInvoked, SourceAgent: "specialist-1", Payload: map[string]any{"tool": "consult_council"}, EmittedAt: baseTime.Add(2 * time.Second)},
			},
			expectedStatus: map[string]string{
				"intent_received":       "completed",
				"proposal_generated":    "completed",
				"approval_required":     "completed",
				"approval_granted":      "completed",
				"mission_running":       "completed",
				"tool_action_execution": "active",
				"package_generated":     "pending",
				"proof_generated":       "pending",
				"recovery_next_actions": "pending",
			},
			expectedActor: "soma",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stages := BuildRunLifecycleStages(tc.events)
			if len(stages) != 9 {
				t.Fatalf("expected 9 stages, got %d", len(stages))
			}
			if stages[0].Actor != tc.expectedActor {
				t.Errorf("expected stage[0] actor %q, got %q", tc.expectedActor, stages[0].Actor)
			}
			for _, st := range stages {
				want, ok := tc.expectedStatus[st.ID]
				if !ok {
					t.Errorf("unexpected stage ID %q", st.ID)
					continue
				}
				if st.Status != want {
					t.Errorf("stage %s (%s): expected status %q, got %q", st.ID, st.Label, want, st.Status)
				}
			}
		})
	}
}

func TestHandleGetRunStages(t *testing.T) {
	evOpt, mock := withEventsStore(t)
	s := newTestServer(evOpt)

	runID := "11112222-3333-4444-5555-666677778888"
	now := time.Now().UTC()

	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "tenant_id", "event_type", "severity",
			"source_agent", "source_team", "payload", "audit_event_id", "emitted_at",
		}).
			AddRow("ev-1", runID, "default", "mission.started", "info", "soma", "core", `{}`, "", now).
			AddRow("ev-2", runID, "default", "mission.completed", "info", "soma", "core", `{}`, "", now))

	mux := http.NewServeMux()
	s.RegisterRunRoutes(mux)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/"+runID+"/stages", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", resp["data"])
	}
	if data["run_id"] != runID {
		t.Errorf("expected run_id=%s, got %v", runID, data["run_id"])
	}
	stages, ok := data["stages"].([]any)
	if !ok || len(stages) != 9 {
		t.Fatalf("expected 9 stages in response, got %v", data["stages"])
	}
}

func TestAssembleTeamSignalDeliveryPackage(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		teamID      string
		runID       string
		title       string
		items       []artifacts.Artifact
		wantKind    string
		wantPosture string
	}{
		{
			name:   "multi-agent code implementation bundle verified",
			teamID: "team-code",
			runID:  "run-code-1",
			title:  "Backend Service Slice",
			items: []artifacts.Artifact{
				{ID: uuid.New(), AgentID: "coder-alpha", ArtifactType: artifacts.TypeCode, Title: "handler.go", ContentType: "text/x-go", FilePath: "core/internal/server/handler.go", Content: "package server", Status: "approved", CreatedAt: now},
				{ID: uuid.New(), AgentID: "verifier-beta", ArtifactType: artifacts.TypeDocument, Title: "proof.md", ContentType: "text/markdown", FilePath: "workspace/proof.md", Content: "# Verified", Status: "completed", CreatedAt: now},
			},
			wantKind:    TeamPackageImplementationBundle,
			wantPosture: VerificationPostureVerified,
		},
		{
			name:   "multi-agent creative bundle with pending review",
			teamID: "team-creative",
			runID:  "run-creative-1",
			title:  "Launch Campaign Assets",
			items: []artifacts.Artifact{
				{ID: uuid.New(), AgentID: "designer-1", ArtifactType: artifacts.TypeImage, Title: "Hero Banner", ContentType: "image/png", FilePath: "campaigns/hero.png", Status: "approved", CreatedAt: now},
				{ID: uuid.New(), AgentID: "copywriter-1", ArtifactType: artifacts.TypeDocument, Title: "Launch Copy", ContentType: "text/markdown", FilePath: "campaigns/copy.md", Status: "pending", CreatedAt: now},
			},
			wantKind:    TeamPackageCreativeBundle,
			wantPosture: VerificationPosturePendingReview,
		},
		{
			name:   "media bundle with rejected artifact marks failed posture",
			teamID: "team-media",
			runID:  "run-media-1",
			title:  "Visual Pack",
			items: []artifacts.Artifact{
				{ID: uuid.New(), AgentID: "media-agent", ArtifactType: artifacts.TypeImage, Title: "Draft Render", ContentType: "image/webp", Status: "rejected", CreatedAt: now},
			},
			wantKind:    TeamPackageMediaBundle,
			wantPosture: VerificationPostureFailed,
		},
		{
			name:   "document bundle verified",
			teamID: "team-docs",
			runID:  "run-docs-1",
			title:  "Architecture Brief",
			items: []artifacts.Artifact{
				{ID: uuid.New(), AgentID: "writer-1", ArtifactType: artifacts.TypeDocument, Title: "brief.md", ContentType: "text/markdown", Status: "approved", CreatedAt: now},
			},
			wantKind:    TeamPackageDocumentBundle,
			wantPosture: VerificationPostureVerified,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := AssembleTeamSignalDeliveryPackage(tc.teamID, tc.runID, tc.title, tc.items)
			if summary == nil || summary.DeliveryPackage == nil {
				t.Fatalf("expected non-nil ExecutionSummary and DeliveryPackage")
			}
			if summary.DeliveryPackage.PackageKind != tc.wantKind {
				t.Errorf("expected package kind %q, got %q", tc.wantKind, summary.DeliveryPackage.PackageKind)
			}
			if summary.Status != tc.wantPosture || summary.DeliveryPackage.Status != tc.wantPosture {
				t.Errorf("expected posture %q, got summary=%q pkg=%q", tc.wantPosture, summary.Status, summary.DeliveryPackage.Status)
			}
			if len(summary.OutputRefs) != len(tc.items) {
				t.Errorf("expected %d output refs, got %d", len(tc.items), len(summary.OutputRefs))
			}
		})
	}
}
