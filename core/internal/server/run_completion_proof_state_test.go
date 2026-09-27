package server

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

type completionPayloadMatch struct{ state, quality string }

func (m completionPayloadMatch) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		return false
	}
	var payload map[string]any
	if json.Unmarshal(raw, &payload) != nil {
		return false
	}
	return payload["execution_state"] == m.state && payload["proof_quality"] == m.quality
}

func TestMarkRunCompletedTxRecordsRealProofState(t *testing.T) {
	const runID = "22222222-2222-2222-2222-222222222222"
	cases := []struct {
		in             protocol.TrustProofQuality
		state, quality string
	}{
		{protocol.TrustProofQualityVerified, "verified", "verified"},
		{protocol.TrustProofQualityUnverified, "unverified", "unverified"},
		{protocol.TrustProofQualityFailed, "unverified", "failed"},
		{"", "unverified", "unverified"},
		{protocol.TrustProofQualityProposed, "unverified", "unverified"},
	}
	for _, tc := range cases {
		t.Run(string(tc.in)+"->"+tc.quality, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE mission_runs SET status").
				WithArgs(runs.StatusCompleted, runID, runs.StatusFailed).
				WillReturnResult(sqlmock.NewResult(0, 1))
			anyArg := sqlmock.AnyArg()
			mock.ExpectExec("INSERT INTO mission_events").
				WithArgs(anyArg, runID, anyArg, string(protocol.EventMissionCompleted), anyArg, anyArg, anyArg, completionPayloadMatch{tc.state, tc.quality}, anyArg).
				WillReturnResult(sqlmock.NewResult(0, 1))
			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			if err := newTestServer().markRunCompletedTx(tx, runID, "proof-1", tc.in); err != nil {
				t.Fatalf("mark: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("completion event must carry the real proof state: %v", err)
			}
		})
	}
}

func TestConfirmActionProofQualityFollowsDiskReadback(t *testing.T) {
	root := useTestWorkspace(t)
	content := "# Launch\n\nBeta opens Monday with governed teams, retained outputs, and audit trails."
	result := plannedToolExecutionResult{Name: "write_file", Arguments: map[string]any{"path": "notes/launch.md", "content": content}}
	if got := confirmActionProofQuality(nil, []plannedToolExecutionResult{result}); got != protocol.TrustProofQualityFailed {
		t.Fatalf("missing file quality = %q, want failed", got)
	}
	writeWorkspaceTestFile(t, root, "notes/launch.md", content)
	if got := confirmActionProofQuality(nil, []plannedToolExecutionResult{result}); got != protocol.TrustProofQualityVerified {
		t.Fatalf("real file quality = %q, want verified", got)
	}
}
