package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func tierBlueprint(agents int, tools ...string) *protocol.MissionBlueprint {
	team := protocol.BlueprintTeam{Name: "alpha", Role: "build"}
	for i := 0; i < agents; i++ {
		team.Agents = append(team.Agents, protocol.AgentManifest{ID: fmt.Sprintf("a%d", i), Role: "worker", Tools: tools})
	}
	return &protocol.MissionBlueprint{MissionID: "m-1", Intent: "Build it", Teams: []protocol.BlueprintTeam{team}}
}

// expectCommitToken mocks the commit token lookup with the scope stored at
// negotiate time and the digest of bound.
func expectCommitToken(t *testing.T, mock sqlmock.Sqlmock, stored *protocol.ScopeValidation, bound *protocol.MissionBlueprint, mintedBy string) {
	t.Helper()
	raw, _ := json.Marshal(stored)
	mock.ExpectQuery(purposeJoinQuery).WillReturnRows(
		sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "resolved_intent", "scope_validation", "purpose", "binding_digest", "minted_by"}).
			AddRow(approverTestProof, false, time.Now().Add(time.Hour), bound.Intent, raw, tokenPurposeMissionBlueprint, blueprintDigest(bound), mintedBy))
}

func expectCommitSuccess(mock sqlmock.Sqlmock, bp *protocol.MissionBlueprint, audits *[]string) {
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO missions").WillReturnResult(sqlmock.NewResult(0, 1))
	for _, team := range bp.Teams {
		mock.ExpectExec("INSERT INTO teams").WillReturnResult(sqlmock.NewResult(0, 1))
		for range team.Agents {
			mock.ExpectExec("INSERT INTO service_manifests").WillReturnResult(sqlmock.NewResult(0, 1))
		}
	}
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed'").WillReturnResult(sqlmock.NewResult(0, 1))
	a := sqlmock.AnyArg()
	mock.ExpectExec("INSERT INTO log_entries").WithArgs(a, a, a, a, a, a, a, auditContextCapture{audits}).WillReturnResult(sqlmock.NewResult(0, 1))
}

func commitAs(t *testing.T, s *AdminServer, bp *protocol.MissionBlueprint, identity *RequestIdentity) *httptest.ResponseRecorder {
	t.Helper()
	fields, _ := json.Marshal(bp)
	var body map[string]any
	_ = json.Unmarshal(fields, &body)
	body["confirm_token"] = approverTestToken
	encoded, _ := json.Marshal(body)
	return doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", string(encoded), identity)
}

func TestBlueprintScopeClassifiesServerSide(t *testing.T) {
	for name, c := range map[string]struct {
		bp   *protocol.MissionBlueprint
		tier int
	}{
		"small low-risk":  {tierBlueprint(2, "read_file"), approverTierAuto},
		"eleven agents":   {tierBlueprint(11), approverTierApprover},
		"broadcast tool":  {tierBlueprint(1, "broadcast"), approverTierApprover},
		"mcp tool":        {tierBlueprint(1, "mcp:github/create_issue"), approverTierApprover},
		"medium write":    {tierBlueprint(1, "write_file"), approverTierAuto},
		"publish signal":  {tierBlueprint(4, "publish_signal"), approverTierApprover},
		"six agents only": {tierBlueprint(6), approverTierAuto},
	} {
		if got, _ := approverTier(buildScopeFromBlueprint(c.bp)); got != c.tier {
			t.Errorf("%s: tier %d, want %d", name, got, c.tier)
		}
	}
}

// A2b C1: tier-2 blueprints need an approver at commit; the token is kept.
func TestIntentCommitTier2BlueprintNeedsApprover(t *testing.T) {
	for name, bp := range map[string]*protocol.MissionBlueprint{
		"eleven agents":  tierBlueprint(11),
		"broadcast tool": tierBlueprint(1, "broadcast"),
		"mcp tool":       tierBlueprint(1, "mcp:filesystem/write_file"),
	} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			expectCommitToken(t, mock, buildScopeFromBlueprint(bp), bp, "u-std")
			rr := commitAs(t, s, bp, standardUserIdentity())
			assertStatus(t, rr, http.StatusForbidden)
			if !strings.Contains(rr.Body.String(), "approver_required") {
				t.Fatalf("expected approver_required, got %s", rr.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("refused commit must not consume the token: %v", err)
			}
		})
	}
}

// A client cannot lower the tier: even if the stored scope were low, the
// decoded body is reclassified server-side and the higher tier wins.
func TestIntentCommitReclassifiesBodyOverStoredScope(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	bp := tierBlueprint(12)
	expectCommitToken(t, mock, buildScopeFromBlueprint(tierBlueprint(1)), bp, "u-std")
	assertStatus(t, commitAs(t, s, bp, standardUserIdentity()), http.StatusForbidden)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIntentCommitApproverAndProposerPaths(t *testing.T) {
	admin := adminWithScopes(scopeApprovalsDecide)
	for name, c := range map[string]struct {
		bp       *protocol.MissionBlueprint
		minter   string
		identity *RequestIdentity
		want     []string
	}{
		"admin approves another's tier 2": {tierBlueprint(11), "u-std", admin, []string{`"approval_authority":"approvals:decide"`, `"approval_tier":2`, `"self_approved":false`}},
		"admin self-approves tier 2":      {tierBlueprint(1, "broadcast"), admin.UserID, admin, []string{`"approval_authority":"approvals:decide"`, `"approval_tier":2`, `"self_approved":true`}},
		"standard user low-risk":          {tierBlueprint(2, "read_file"), "u-std", standardUserIdentity(), []string{`"approval_authority":"proposer"`, `"approval_tier":0`, `"self_approved":true`}},
	} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			var audits []string
			expectCommitToken(t, mock, buildScopeFromBlueprint(c.bp), c.bp, c.minter)
			expectCommitSuccess(mock, c.bp, &audits)
			assertStatus(t, commitAs(t, s, c.bp, c.identity), http.StatusOK)
			joined := strings.Join(audits, "\n")
			for _, want := range c.want {
				if !strings.Contains(joined, want) {
					t.Fatalf("commit audit must record %s: %v", want, audits)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
