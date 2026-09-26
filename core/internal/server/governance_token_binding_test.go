package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// otherUserIdentity is a second standard user (not the fixture minter).
func otherUserIdentity() *RequestIdentity {
	id := standardUserIdentity()
	id.UserID, id.Username = "u-other", "other@example.com"
	return id
}

func requestAs(identity *RequestIdentity) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if identity == nil {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity, identity))
}

func TestConfirmerMayConfirmOwn(t *testing.T) {
	for name, c := range map[string]struct {
		identity *RequestIdentity
		mintedBy string
		want     bool
	}{
		"minter":                {standardUserIdentity(), "u-std", true},
		"other user":            {otherUserIdentity(), "u-std", false},
		"anonymous":             {nil, "u-std", false},
		"legacy empty minter":   {standardUserIdentity(), "", false},
		"approver any minter":   {adminWithScopes(scopeApprovalsDecide), "u-std", true},
		"admin without decide":  {adminWithScopes(scopeGovernanceRead), "u-std", false},
		"username not a userid": {standardUserIdentity(), "std@example.com", false},
	} {
		if got := confirmerMayConfirmOwn(requestAs(c.identity), c.mintedBy); got != c.want {
			t.Errorf("%s: got %v want %v", name, got, c.want)
		}
	}
}

// Q3: a self-confirmable chat token minted by u-std is refused for another
// user (403 confirmer_not_proposer, rollback keeps the token), then the
// proposer and an approver can confirm it.
func TestConfirmActionBindsSelfConfirmToProposer(t *testing.T) {
	for name, approval := range map[string]*protocol.ApprovalPolicy{
		"auto":        nil,
		"medium-risk": {ApprovalRequired: true, ApprovalReason: "capability_risk", CapabilityRisk: "medium"},
	} {
		t.Run(name, func(t *testing.T) {
			workspace := t.TempDir()
			t.Setenv("MYCELIS_WORKSPACE", workspace)
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			scope := approverTestScope(approval)
			for _, intruder := range []*RequestIdentity{otherUserIdentity(), adminWithScopes(scopeGovernanceWrite)} {
				expectTokenAndScope(t, mock, scope)
				mock.ExpectRollback()
				rr := confirmAs(t, s, intruder)
				assertStatus(t, rr, http.StatusForbidden)
				if !strings.Contains(rr.Body.String(), codeConfirmerNotProposer) {
					t.Fatalf("expected %s, got %s", codeConfirmerNotProposer, rr.Body.String())
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("non-proposer confirm must roll back before any effect: %v", err)
				}
				if _, err := os.Stat(filepath.Join(workspace, "output", "confirmed.txt")); err == nil {
					t.Fatal("non-proposer confirm executed the planned tool")
				}
			}
			var audits []string
			expectConfirmSuccess(t, mock, scope, &audits)
			assertStatus(t, confirmAs(t, s, standardUserIdentity()), http.StatusOK)
			var adminAudits []string
			expectConfirmSuccess(t, mock, scope, &adminAudits)
			assertStatus(t, confirmAs(t, s, adminWithScopes(scopeApprovalsDecide)), http.StatusOK)
		})
	}
}

// Q3: a mission_blueprint token is refused at intent/commit for a principal
// that did not mint it; nothing is consumed.
func TestIntentCommitBindsBlueprintTokenToProposer(t *testing.T) {
	for name, identity := range map[string]*RequestIdentity{"other-user": otherUserIdentity(), "anonymous": nil} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			expectBoundLookup(t, mock, "Build a scraper", blueprintScope(), tokenPurposeMissionBlueprint, "u-std")
			body := `{"intent":"Build a scraper","confirm_token":"` + approverTestToken + `","teams":[{"name":"t","role":"r","agents":[]}]}`
			var rr *httptest.ResponseRecorder
			if identity == nil {
				rr = doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
			} else {
				rr = doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body, identity)
			}
			assertStatus(t, rr, http.StatusForbidden)
			if !strings.Contains(rr.Body.String(), codeConfirmerNotProposer) {
				t.Fatalf("expected %s, got %s", codeConfirmerNotProposer, rr.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("refused commit must not consume the token: %v", err)
			}
		})
	}
}

func TestGenerateConfirmTokenRecordsMint(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	if _, err := s.generateConfirmToken(approverTestProof, protocol.TemplateChatToProposal, confirmTokenMint{Purpose: tokenPurposeChatAction}); err != errTokenMintUnbound {
		t.Fatalf("mint without a principal must be refused, got %v", err)
	}
	if _, err := s.generateConfirmToken(approverTestProof, protocol.TemplateChatToProposal, confirmTokenMint{MintedBy: "u-std"}); err != errTokenMintUnbound {
		t.Fatalf("mint without a purpose must be refused, got %v", err)
	}
	a := sqlmock.AnyArg()
	mock.ExpectExec("INSERT INTO confirm_tokens \\(token, intent_proof_id, template_id, expires_at, purpose, binding_digest, minted_by\\)").
		WithArgs(a, approverTestProof, string(protocol.TemplateChatToProposal), a, tokenPurposeMissionBlueprint, "abc", "u-std").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mint := confirmTokenMint{Purpose: tokenPurposeMissionBlueprint, BindingDigest: "abc", MintedBy: "u-std"}
	if tok, err := s.generateConfirmToken(approverTestProof, protocol.TemplateChatToProposal, mint); err != nil || tok == nil {
		t.Fatalf("mint failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Every Go INSERT INTO confirm_tokens must record purpose, binding digest and
// minting principal (A2b item 5).
func TestEveryConfirmTokenInsertRecordsMint(t *testing.T) {
	insert := regexp.MustCompile(`(?is)INSERT\s+INTO\s+confirm_tokens\s*\(([^)]*)\)`)
	found := 0
	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range insert.FindAllStringSubmatch(string(raw), -1) {
			found++
			cols := strings.ReplaceAll(m[1], " ", "")
			for _, col := range []string{"purpose", "binding_digest", "minted_by"} {
				if !strings.Contains(cols, col) {
					t.Errorf("%s: INSERT INTO confirm_tokens misses %s: %s", path, col, m[1])
				}
			}
		}
		return nil
	})
	if err != nil || found < 2 {
		t.Fatalf("expected the server and invocation mint sites, found %d (err %v)", found, err)
	}
}

// A2b item 7: confirm-action accepts only chat_action tokens. Blueprint and
// group tokens get 400 token_wrong_purpose; legacy NULL tokens get 409
// token_purpose_unknown. Nothing is consumed, no scope is loaded, no run made.
func TestConfirmActionRejectsNonChatTokensWithoutConsuming(t *testing.T) {
	for purpose, want := range map[string]struct {
		status int
		code   string
	}{
		tokenPurposeMissionBlueprint: {http.StatusBadRequest, codeTokenWrongPurpose},
		tokenPurposeGroupMutation:    {http.StatusBadRequest, codeTokenWrongPurpose},
		"":                           {http.StatusConflict, codeTokenPurposeUnknown},
	} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		mock.ExpectBegin()
		mock.ExpectQuery(confirmTokenTxQuery).WillReturnRows(
			sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "purpose", "binding_digest", "minted_by"}).
				AddRow(approverTestProof, false, time.Now().Add(time.Hour), purpose, "", approverTestMinter))
		mock.ExpectRollback()
		rr := confirmAs(t, s, standardUserIdentity())
		assertStatus(t, rr, want.status)
		if !strings.Contains(rr.Body.String(), want.code) {
			t.Fatalf("purpose %q: expected %s, got %s", purpose, want.code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("purpose %q: token must not be consumed and no run created: %v", purpose, err)
		}
	}
}

// A2b item 5: each purpose is refused at the wrong route (not consumed) and
// accepted at its own; NULL purpose is refused everywhere with 409.
func TestDurablePurposeRoutesTokens(t *testing.T) {
	bp, grp := blueprintScope(), groupScope()
	commit := func(s *AdminServer) (string, error) {
		proofID, _, err := s.consumeProposerTokenFor(requestAs(standardUserIdentity()), approverTestToken, blueprintCommitPurpose, nil, nil)
		return proofID, err
	}
	group := func(s *AdminServer) (string, error) {
		return s.consumeConfirmTokenFor(approverTestToken, groupMutationPurpose("create", ""))
	}
	for name, c := range map[string]struct {
		route   func(*AdminServer) (string, error)
		intent  string
		scope   protocol.ScopeValidation
		purpose string
		want    error
	}{
		"group token at commit":    {commit, "groups.create", grp, tokenPurposeGroupMutation, errTokenWrongPurpose},
		"chat token at commit":     {commit, chatActionResolvedIntent, approverTestScope(nil), tokenPurposeChatAction, errTokenWrongPurpose},
		"blueprint token at group": {group, "Build a scraper", bp, tokenPurposeMissionBlueprint, errTokenWrongPurpose},
		"NULL purpose at commit":   {commit, "Build a scraper", bp, "", errTokenPurposeUnknown},
		"NULL purpose at group":    {group, "groups.create", grp, "", errTokenPurposeUnknown},
		// A blueprint-shaped proof cannot be relabeled by the column alone.
		"column lies at commit": {commit, chatActionResolvedIntent, bp, tokenPurposeMissionBlueprint, errTokenWrongPurpose},
		"blueprint at commit":   {commit, "Build a scraper", bp, tokenPurposeMissionBlueprint, nil},
		"group at group":        {group, "groups.create", grp, tokenPurposeGroupMutation, nil},
	} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		expectBoundLookup(t, mock, c.intent, c.scope, c.purpose, approverTestMinter)
		if c.want == nil {
			mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 1))
		}
		if _, err := c.route(s); !errors.Is(err, c.want) && !(c.want == nil && err == nil) {
			t.Fatalf("%s: got %v, want %v", name, err, c.want)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	rr := httptest.NewRecorder()
	respondConfirmTokenError(rr, errTokenPurposeUnknown, http.StatusBadRequest)
	if rr.Code != http.StatusConflict || !strings.Contains(rr.Body.String(), "propose this again") {
		t.Fatalf("NULL purpose must be 409 with re-propose guidance: %d %s", rr.Code, rr.Body.String())
	}
}
