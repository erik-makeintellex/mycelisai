package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// S7b item 3: only the proposer (confirm_tokens.minted_by) or an approver
// holding approvals:decide may cancel a proposal; others get 403 and the
// proposal stays open.

const cancelProofID = "11111111-1111-1111-1111-111111111111"

func cancelBody() string { return `{"intent_proof_id":"` + cancelProofID + `"}` }

func expectCancelOwner(mock sqlmock.Sqlmock, mintedBy ...string) {
	rows := sqlmock.NewRows([]string{"minted_by"})
	for _, minter := range mintedBy {
		rows.AddRow(minter)
	}
	mock.ExpectQuery("SELECT COALESCE\\(t.minted_by, ''\\) FROM intent_proofs p").WithArgs(cancelProofID).WillReturnRows(rows)
}

func standardUser(id string) *RequestIdentity {
	return &RequestIdentity{UserID: id, Username: id, Role: "user", EffectiveRole: "member", PrincipalType: "user", Scopes: []string{"chat:write"}}
}

func TestCancelActionByNonOwnerIsForbiddenAndKeepsProposal(t *testing.T) {
	for name, identity := range map[string]*RequestIdentity{
		"other user":             standardUser("user-b"),
		"admin without approver": {UserID: "admin-2", Role: "admin", EffectiveRole: "admin", Scopes: []string{"governance:read"}},
	} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			expectCancelOwner(mock, "user-a")
			rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleCancelAction), "POST", "/api/v1/intent/cancel-action", cancelBody(), identity)
			assertStatus(t, rr, http.StatusForbidden)
			var resp protocol.APIResponse
			assertJSON(t, rr, &resp)
			if resp.OK || blockerCode(resp) != codeCancellerNotProposer {
				t.Fatalf("resp = %+v", resp)
			}
			// sqlmock fails any unexpected UPDATE, so the proposal was not cancelled.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet DB expectations: %v", err)
			}
		})
	}
}

func TestCancelActionWithoutIdentityIsUnauthorized(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	expectCancelOwner(mock, "user-a")
	rr := doRequest(t, http.HandlerFunc(s.HandleCancelAction), "POST", "/api/v1/intent/cancel-action", cancelBody())
	assertStatus(t, rr, http.StatusUnauthorized)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}

func TestCancelActionByProposerOrApproverSucceeds(t *testing.T) {
	for name, identity := range map[string]*RequestIdentity{
		"proposer": standardUser("user-a"),
		"approver": localAdminIdentityForTest(),
	} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			expectCancelOwner(mock, "user-a")
			mock.ExpectExec("UPDATE intent_proofs SET status = 'cancelled'").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(0, 1))
			rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleCancelAction), "POST", "/api/v1/intent/cancel-action", cancelBody(), identity)
			assertStatus(t, rr, http.StatusOK)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet DB expectations: %v", err)
			}
		})
	}
}

func TestCancelActionUnknownProofIsNotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	expectCancelOwner(mock)
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleCancelAction), "POST", "/api/v1/intent/cancel-action", cancelBody())
	assertStatus(t, rr, http.StatusNotFound)
}
