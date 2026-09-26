package server

import (
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// sideEffectArg matches one sqlmock argument and runs fn as a side effect
// exactly when the driver checks that argument, which happens synchronously
// inside ExecContext — before the surrounding tx commits. It gives a
// deterministic way to land a mutation "between commit and apply" without
// timing-based races.
type sideEffectArg struct {
	want string
	fn   func()
}

func (a sideEffectArg) Match(v driver.Value) bool {
	a.fn()
	s, ok := v.(string)
	return ok && s == a.want
}

// TestCognitiveProfileOverride_ConcurrentPutAndDeleteSerialize is R1/F3: a
// PUT and a DELETE against the same profile, fired concurrently, must not
// interleave their DB commit and runtime apply. Whichever one's apply runs
// last must fully determine the final state — never a mix of one request's
// DB row with the other's in-memory binding. Run with -race.
func TestCognitiveProfileOverride_ConcurrentPutAndDeleteSerialize(t *testing.T) {
	s, mock, mux := profileOverrideServer(t)
	mock.MatchExpectationsInOrder(false)
	s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB)

	// Two independent audit writes, order not guaranteed.
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(2, 1))

	// PUT's transaction.
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO system_config").WithArgs("role.coder", "vllm").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// DELETE's transaction.
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM system_config").WithArgs("coder").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	var wg sync.WaitGroup
	responses := make(chan *httptest.ResponseRecorder, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		responses <- doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/profiles", `{"profiles":{"coder":"vllm"}}`)
	}()
	go func() {
		defer wg.Done()
		responses <- doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", "")
	}()
	wg.Wait()
	close(responses)
	for rr := range responses {
		assertStatus(t, rr, http.StatusOK)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}

	state := s.Cognitive.ProfileOverrideState("coder")
	putLastApplied := state.Source == cognitive.ProfileSourceOverride && state.DBRowPresent && state.Origin == cognitive.ProfileOriginDB && state.ProviderID == "vllm"
	deleteLastApplied := state.Source == cognitive.ProfileSourceRoot && !state.DBRowPresent && state.Origin == "" && state.ProviderID == "vllm"
	if !putLastApplied && !deleteLastApplied {
		t.Fatalf("memory state does not match either request's DB outcome (torn/interleaved apply): %+v", state)
	}
}

// TestCognitiveProfileOverride_PutApplyFailureAfterCommitReturns500 is R1/F3:
// when a PUT's role.* rows commit but the runtime apply then fails (the
// provider was disabled between the pre-commit validation and the
// post-commit apply), the row is durable, so the response must never be the
// 400/409 used for an outright rejection. It must be 500 with
// override_committed_not_applied and a hint naming the DELETE reset route,
// and the audit row and the system_config row must both be present.
func TestCognitiveProfileOverride_PutApplyFailureAfterCommitReturns500(t *testing.T) {
	s, mock, mux := profileOverrideServer(t)
	expectAudit(mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO system_config").WithArgs("role.coder", sideEffectArg{
		want: "ollama",
		fn: func() {
			// Simulate a concurrent provider-disable landing between the
			// pre-commit validation and the post-commit runtime apply.
			provider, _ := s.Cognitive.ProviderSnapshot("ollama")
			provider.Enabled = false
			s.Cognitive.StoreProviderConfig("ollama", provider)
		},
	}).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rr := doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/profiles", `{"profiles":{"coder":"ollama"}}`)
	assertStatus(t, rr, http.StatusInternalServerError)

	var committed profileOverrideCommitted
	decodeOverrideData(t, rr, &committed)
	if committed.Code != overrideCommittedNotApplied {
		t.Fatalf("code = %q, want %q", committed.Code, overrideCommittedNotApplied)
	}
	if !strings.Contains(committed.RecommendedAction, "DELETE /api/v1/cognitive/profiles/{profile}/override") {
		t.Fatalf("recommended_action = %q, missing the DELETE reset hint", committed.RecommendedAction)
	}
	if len(committed.Profiles) != 1 || committed.Profiles[0] != "coder" {
		t.Fatalf("profiles = %v", committed.Profiles)
	}
	// The row and the audit event both landed: the commit happened before
	// the apply failed, so the row is durable even though the request
	// reports 500.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v (row/audit must still be present)", err)
	}
}
