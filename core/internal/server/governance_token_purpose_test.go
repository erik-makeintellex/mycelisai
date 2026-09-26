package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/invocation"
	"github.com/mycelis/core/pkg/protocol"
)

const purposeJoinQuery = "SELECT t.intent_proof_id, t.consumed, t.expires_at, p.resolved_intent, p.scope_validation\\s+FROM confirm_tokens t JOIN intent_proofs p"

func expectPurposeLookup(t *testing.T, mock sqlmock.Sqlmock, resolvedIntent string, scope protocol.ScopeValidation) {
	t.Helper()
	raw, err := json.Marshal(scope)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(purposeJoinQuery).
		WithArgs(uuid.MustParse(approverTestToken)).
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "resolved_intent", "scope_validation"}).
			AddRow(approverTestProof, false, time.Now().Add(time.Hour), resolvedIntent, raw))
}

func blueprintScope() protocol.ScopeValidation {
	return *buildScopeFromBlueprint(&protocol.MissionBlueprint{Teams: []protocol.BlueprintTeam{{Name: "t"}}})
}

func groupScope() protocol.ScopeValidation {
	return protocol.ScopeValidation{Tools: []string{"write_file"}, AffectedResources: []string{"collaboration_groups", "team_channels"}, RiskLevel: "high"}
}

func TestConfirmTokenPurposes(t *testing.T) {
	posture := approverTestScope(postureRaisedApproval())
	bp, grp := blueprintScope(), groupScope()
	cases := []struct {
		name    string
		purpose confirmTokenPurpose
		intent  string
		scope   protocol.ScopeValidation
		want    bool
	}{
		{"commit accepts blueprint token", blueprintCommitPurpose, "Build a scraper", bp, true},
		{"commit rejects posture chat token", blueprintCommitPurpose, chatActionResolvedIntent, posture, false},
		{"commit rejects chat token", blueprintCommitPurpose, chatActionResolvedIntent, approverTestScope(nil), false},
		{"commit rejects blueprint-shaped posture scope", blueprintCommitPurpose, "Build", func() protocol.ScopeValidation { s := bp; s.Approval = postureRaisedApproval(); return s }(), false},
		{"commit rejects group token", blueprintCommitPurpose, "groups.create", grp, false},
		{"commit rejects invocation token", blueprintCommitPurpose, invocation.CapabilityID, bp, false},
		{"group accepts its own op", groupMutationPurpose("create", ""), "groups.create", grp, true},
		{"group rejects other group", groupMutationPurpose("update", "g-1"), "groups.update.g-2", grp, false},
		{"group rejects posture chat token", groupMutationPurpose("create", ""), chatActionResolvedIntent, posture, false},
		{"group rejects blueprint token named like a group op", groupMutationPurpose("create", ""), "groups.create", bp, false},
	}
	for _, tc := range cases {
		scope := tc.scope
		if got := tc.purpose(tc.intent, &scope); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

// A posture-raised chat token presented to intent/commit is rejected without
// being consumed (no UPDATE), and still confirms through confirm-action for an
// approver.
func TestIntentCommitRejectsPostureTokenWithoutConsuming(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	scope := approverTestScope(postureRaisedApproval())
	expectPurposeLookup(t, mock, chatActionResolvedIntent, scope)
	body := `{"intent":"anything","confirm_token":"` + approverTestToken + `","teams":[{"name":"t","role":"r","agents":[]}]}`
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body, standardUserIdentity())
	assertStatus(t, rr, http.StatusForbidden)
	if !strings.Contains(rr.Body.String(), codeTokenWrongPurpose) {
		t.Fatalf("expected %s, got %s", codeTokenWrongPurpose, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("wrong-purpose token must not be consumed or activate anything: %v", err)
	}

	var audits []string
	expectConfirmSuccess(t, mock, scope, &audits)
	assertStatus(t, confirmAs(t, s, adminWithScopes(scopeApprovalsDecide)), http.StatusOK)
}

func TestConsumeConfirmTokenForConcurrentSingleWinner(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(false)
	for i := 0; i < 2; i++ {
		expectPurposeLookup(t, mock, "groups.create", groupScope())
	}
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 0))
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = s.consumeConfirmTokenFor(approverTestToken, groupMutationPurpose("create", ""))
		}(i)
	}
	wg.Wait()
	wins, lost := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, errTokenAlreadyUsed):
			lost++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if wins != 1 || lost != 1 {
		t.Fatalf("expected exactly one winner, got wins=%d lost=%d", wins, lost)
	}
}

// Two confirms of one token: the one whose UPDATE affects no row gets 409
// token_already_used and executes nothing.
func TestConfirmActionLosingConcurrentConfirmGets409(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = \\$1").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}).AddRow(approverTestProof, false, time.Now().Add(time.Hour)))
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	rr := confirmAs(t, s, adminWithScopes("*"))
	assertStatus(t, rr, http.StatusConflict)
	if !strings.Contains(rr.Body.String(), codeTokenAlreadyUsed) {
		t.Fatalf("expected %s, got %s", codeTokenAlreadyUsed, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("losing confirm must not load scope, create a run, or execute: %v", err)
	}

	// An already-consumed token is also 409.
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = \\$1").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}).AddRow(approverTestProof, true, time.Now().Add(time.Hour)))
	mock.ExpectRollback()
	assertStatus(t, confirmAs(t, s, adminWithScopes("*")), http.StatusConflict)
}
