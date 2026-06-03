package server

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkSignalProjection_ResultPersistsOutputRefs(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdateWithOutputs(mock, workID, protocol.TeamWorkStateOutputReady, false, "", outputRefsMatch{
		TeamID:     "research-team",
		WorkItemID: workID,
		Kind:       "project_package",
		Label:      "Playable prototype",
		StorageRef: "groups/research-team/generated/prototype",
		Entrypoint: "index.html",
		ProofRef:   "proof-1",
	})
	expectProjectedInteraction(mock, "research-team", workID, "output_ready", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{
			"work_item_id":"` + workID + `",
			"summary":"Prototype ready",
			"outputs":[{
				"id":"prototype-package",
				"kind":"project_package",
				"title":"Playable prototype",
				"folder":"groups/research-team/generated/prototype",
				"entrypoint":"index.html",
				"validation":"local smoke passed"
			}]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_ResultPersistsNormalizedOutputRefs(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "media-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "media-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdateWithOutputs(mock, workID, protocol.TeamWorkStateOutputReady, false, "", outputRefsMatch{
		TeamID:     "media-team",
		WorkItemID: workID,
		Kind:       "media",
		Label:      "Comic page",
		StorageRef: "groups/media-team/media/comic-page.png",
	})
	expectProjectedInteraction(mock, "media-team", workID, "output_ready", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.media-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "media-team",
			AgentID:       "artist",
		},
		Payload: json.RawMessage(`{
			"work_item_id":"` + workID + `",
			"summary":"Image ready",
			"output_refs":[{
				"output_id":"comic-page-output",
				"kind":"media",
				"label":"Comic page",
				"storage_ref":"groups/media-team/media/comic-page.png"
			}]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.media-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestOutputRefFromMapNormalizesViewerURLToWorkspaceStorageRef(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:     "first-demo-game-team",
		WorkItemID: "work-1",
		RunID:      "run-1",
		ProofID:    "proof-1",
	}
	env := protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			TeamID: "first-demo-game-team",
			RunID:  "run-1",
		},
	}

	ref, ok := outputRefFromMap(item, env, map[string]any{
		"id":          "playable-package",
		"kind":        "project_package",
		"title":       "Playable package",
		"storage_ref": "/api/v1/workspace/files/view?path=groups%2Ffirst-demo-game-team%2Fgenerated%2Ffirst-game%2Findex.html",
		"entrypoint":  "groups/first-demo-game-team/generated/first-game/index.html",
	})
	if !ok {
		t.Fatal("outputRefFromMap returned no ref")
	}
	if ref.StorageRef != "groups/first-demo-game-team/generated/first-game" {
		t.Fatalf("storage_ref = %q, want workspace folder", ref.StorageRef)
	}
	if ref.Entrypoint != "index.html" {
		t.Fatalf("entrypoint = %q, want relative package entrypoint", ref.Entrypoint)
	}
}

func TestOutputRefFromMapDerivesFilePathFromViewerHrefForMedia(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:     "media-team",
		WorkItemID: "work-1",
	}
	env := protocol.SignalEnvelope{Meta: protocol.SignalMeta{TeamID: "media-team"}}

	ref, ok := outputRefFromMap(item, env, map[string]any{
		"id":     "comic-page",
		"kind":   "media",
		"title":  "Comic page",
		"folder": "groups/media-team/media",
		"href":   "/api/v1/workspace/files/view?path=groups%2Fmedia-team%2Fmedia%2Fcomic-page.png",
	})
	if !ok {
		t.Fatal("outputRefFromMap returned no ref")
	}
	if ref.StorageRef != "groups/media-team/media/comic-page.png" {
		t.Fatalf("storage_ref = %q, want decoded workspace file path", ref.StorageRef)
	}
}

func expectProjectedTeamWorkUpdateWithOutputs(mock sqlmock.Sqlmock, workID string, state protocol.TeamWorkState, needsOperator bool, degradation string, outputs sqlmock.Argument) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			workID, string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(), outputs, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

type outputRefsMatch struct {
	TeamID     string
	WorkItemID string
	Kind       string
	Label      string
	StorageRef string
	Entrypoint string
	ProofRef   string
}

func (m outputRefsMatch) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		text, ok := value.(string)
		if !ok {
			return false
		}
		raw = []byte(text)
	}
	var refs []protocol.TeamOutputRef
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) == 0 {
		return false
	}
	for _, ref := range refs {
		if m.TeamID != "" && ref.TeamID != m.TeamID {
			continue
		}
		if m.WorkItemID != "" && ref.WorkItemID != m.WorkItemID {
			continue
		}
		if m.Kind != "" && ref.Kind != m.Kind {
			continue
		}
		if m.Label != "" && ref.Label != m.Label {
			continue
		}
		if m.StorageRef != "" && ref.StorageRef != m.StorageRef {
			continue
		}
		if m.Entrypoint != "" && ref.Entrypoint != m.Entrypoint {
			continue
		}
		if m.ProofRef != "" && ref.ProofRef != m.ProofRef {
			continue
		}
		return true
	}
	return false
}
