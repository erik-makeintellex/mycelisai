package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

const readbackReport = "# Weekly report\n\nThroughput rose 12 percent; two incidents closed with root causes recorded."

func readbackTeamItem() protocol.TeamWorkItem {
	return protocol.TeamWorkItem{
		TeamID: "report-team", WorkItemID: "11111111-1111-1111-1111-111111111111",
		RunID: "22222222-2222-2222-2222-222222222222", IntentProofID: "33333333-3333-3333-3333-333333333333",
		ContractID: "44444444-4444-4444-4444-444444444444", Objective: readbackTestRequest,
		State: protocol.TeamWorkStateOutputReady,
	}
}

func readbackRef(id, storage string) protocol.TeamOutputRef {
	return protocol.TeamOutputRef{OutputID: id, Kind: "document", Label: id, StorageRef: storage}
}

func TestReadbackTeamOutputRefsRejectsMissingUnreadableAndDigestMismatch(t *testing.T) {
	root := useTestWorkspace(t)
	writeWorkspaceTestFile(t, root, "groups/report-team/report.md", readbackReport)
	item := readbackTeamItem()
	good := readbackRef("report", "groups/report-team/report.md")

	if rb := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{good}, nil); rb.failureCode() != "" {
		t.Fatalf("real ref must read back: %+v", rb.Failed)
	} else if got := rb.ByRef[teamOutputRefKey(good)]; got.Checksum != diskChecksum(t, root, "groups/report-team/report.md") {
		t.Fatalf("checksum must describe disk: %+v", got)
	}
	claimed := map[string]string{teamOutputRefKey(good): "sha256:" + sha256Hex(readbackReport+"x")}
	if code := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{good}, claimed).failureCode(); code != "output_digest_mismatch" {
		t.Fatalf("mismatched claimed digest code = %q", code)
	}
	claimed[teamOutputRefKey(good)] = sha256Hex(readbackReport)
	if code := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{good}, claimed).failureCode(); code != "" {
		t.Fatalf("matching claimed digest code = %q", code)
	}
	missing := readbackRef("ghost", "groups/report-team/ghost.md")
	if code := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{good, missing}, nil).failureCode(); code != "output_missing" {
		t.Fatalf("one missing ref must fail the result, code = %q", code)
	}
	writeWorkspaceTestFile(t, root, "groups/report-team/empty-dir/.keep", "")
	if code := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{readbackRef("dir", "groups/report-team/empty-dir")}, nil).failureCode(); code != "output_empty" {
		t.Fatalf("folder without content code = %q", code)
	}
	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		locked := filepath.Join(root, "groups", "report-team", "locked.md")
		writeWorkspaceTestFile(t, root, "groups/report-team/locked.md", readbackReport)
		if err := os.Chmod(locked, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })
		if code := readbackTeamOutputRefs(item, []protocol.TeamOutputRef{readbackRef("locked", "groups/report-team/locked.md")}, nil).failureCode(); code != "output_unreadable" {
			t.Fatalf("unreadable ref code = %q", code)
		}
	}
}

func TestTeamWorkSignalProjection_ClaimedRefWithoutFileDegrades(t *testing.T) {
	useTestWorkspace(t)
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, now)
	recovery := []string{"Core could not read back the claimed output. Ask Soma to have the team regenerate it and return a readable retained file."}
	expectProjectedTeamWorkUpdateWithRecovery(mock, workID, protocol.TeamWorkStateDegraded, true, "output_missing", recovery)
	expectProjectedInteraction(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp: now, SourceKind: protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger", PayloadKind: protocol.PayloadKindResult,
			TeamID: "research-team", AgentID: "builder",
		},
		Payload: json.RawMessage(`{"context":{"work_item_id":"` + workID + `"},"state":"output_ready",
			"outputs":[{"output_id":"claimed-report","kind":"document","storage_ref":"groups/research-team/claimed.md"}]}`),
	})
	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectTeamProofInsert(mock sqlmock.Sqlmock, status protocol.ProofArtifactStatus, strength protocol.TrustEvidenceStrength, quality protocol.TrustProofQuality) {
	anyArg := sqlmock.AnyArg()
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WithArgs(anyArg, anyArg, anyArg, anyArg, "team_signal_result", string(status), anyArg, anyArg, string(strength), string(quality), anyArg, anyArg, anyArg, anyArg, anyArg, anyArg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("55555555-5555-5555-5555-555555555555"))
}

func TestTeamResultProofWithoutPlanIsUnverifiedNeverVerified(t *testing.T) {
	root := useTestWorkspace(t)
	writeWorkspaceTestFile(t, root, "groups/report-team/report.md", readbackReport)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	item := readbackTeamItem()
	refs := []protocol.TeamOutputRef{readbackRef("report", "groups/report-team/report.md")}
	readback := readbackTeamOutputRefs(item, refs, nil)
	expectTeamProofInsert(mock, protocol.ProofArtifactStatusSuccess, protocol.TrustEvidenceStrengthRetainedOutput, protocol.TrustProofQualityUnverified)

	p := &teamWorkSignalProjection{server: newTestServer()}
	if _, err := p.recordAsyncCompletionProof(t.Context(), db, item, protocol.PayloadKindResult, refs, false, readback); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("proof must be unverified without a validation plan: %v", err)
	}
}

func TestTeamResultProofWithMissingRefRecordsFailedProof(t *testing.T) {
	useTestWorkspace(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	item := readbackTeamItem()
	refs := []protocol.TeamOutputRef{readbackRef("ghost", "groups/report-team/ghost.md")}
	readback := readbackTeamOutputRefs(item, refs, nil)
	item.State = protocol.TeamWorkStateDegraded
	expectTeamProofInsert(mock, protocol.ProofArtifactStatusDegraded, protocol.TrustEvidenceStrengthDegraded, protocol.TrustProofQualityFailed)

	p := &teamWorkSignalProjection{server: newTestServer()}
	if _, err := p.recordAsyncCompletionProof(t.Context(), db, item, protocol.PayloadKindResult, refs, false, readback); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("missing ref must record a failed proof: %v", err)
	}
	// Without a readback verdict a degraded item records nothing.
	if id, err := p.recordAsyncCompletionProof(t.Context(), db, item, protocol.PayloadKindResult, refs, false, nil); err != nil || id != "" {
		t.Fatalf("degraded item without readback must not record proof: id=%q err=%v", id, err)
	}
}
