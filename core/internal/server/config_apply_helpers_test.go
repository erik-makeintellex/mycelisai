package server

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	applyPolicyRef  = "fixture-approve-external"
	applyProofID    = "33333333-3333-3333-3333-333333333333"
	applyToken      = "44444444-4444-4444-4444-444444444444"
	applyGroupID    = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	applyCustomID   = "custom-acme-studio"
	applyFixtureSrc = "fixture-minimal-starter"
)

// newApplyServer wires a sqlmock DB, the fixture starter catalog, and a
// policy that resolves the fixture approval policy.
func newApplyServer(t *testing.T) (*AdminServer, sqlmock.Sqlmock) {
	t.Helper()
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, withGuard(fixtureStarterPolicy(applyPolicyRef)))
	s.TemplateBundlesPath = writeStarterCatalog(t, applyPolicyRef)
	return s, mock
}

func applyMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/config/templates/{id}/clone", s.HandleCloneTemplate)
	mux.HandleFunc("POST /api/v1/config/{kind}/{id}/apply", s.HandleApplyConfig)
	mux.HandleFunc("GET /api/v1/config/{kind}/{id}", s.HandleGetConfigView)
	return mux
}

type applyEnvelope struct {
	OK    bool           `json:"ok"`
	Error string         `json:"error"`
	Data  map[string]any `json:"data"`
}

// postApply sends body (any JSON-encodable value) and decodes the envelope.
func postApply(t *testing.T, s *AdminServer, path string, body any, identity *RequestIdentity, want int) applyEnvelope {
	t.Helper()
	if identity == nil {
		identity = localAdminIdentityForTest()
	}
	raw, _ := json.Marshal(body)
	rr := doAuthenticatedRequestAs(t, applyMux(s), "POST", path, string(raw), identity)
	if rr.Code != want {
		t.Fatalf("POST %s status = %d, want %d: %s", path, rr.Code, want, rr.Body.String())
	}
	var env applyEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return env
}

// fixtureStarterDoc returns the fixture starter re-identified as id, as YAML.
func fixtureStarterDoc(t *testing.T, id string) string {
	t.Helper()
	data, err := os.ReadFile("../../config/templates/starters/testdata/minimal-starter.yaml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	body := strings.Replace(string(data), "id: fixture-minimal-starter", "id: "+id, 1)
	return strings.Replace(body, "fixture-approve-external", applyPolicyRef, 1)
}

// fixtureStarterJSON is fixtureStarterDoc as canonical JSON bytes.
func fixtureStarterJSON(t *testing.T, id string) []byte {
	t.Helper()
	doc, err := parseConfigDocument(fixtureStarterDoc(t, id), "yaml")
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	raw, _, err := canonicalConfigJSON(doc)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func operatorTemplateRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "source_template_id", "document", "document_sha256", "created_by", "created_at", "updated_at"})
}

func expectOperatorTemplate(mock sqlmock.Sqlmock, id string, doc []byte) {
	rows := operatorTemplateRows()
	if doc != nil {
		now := time.Now()
		rows.AddRow(id, "Acme Studio", applyFixtureSrc, doc, "sha", "test-user-001", now, now)
	}
	mock.ExpectQuery("FROM operator_templates WHERE id").WillReturnRows(rows)
}

func auditArgs(source string) []driver.Value {
	return []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", source,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}
}

// expectPreviewMint expects the dry-run audit, intent proof, and token.
func expectPreviewMint(mock sqlmock.Sqlmock, auditSource string) {
	mock.ExpectExec("INSERT INTO log_entries").WithArgs(auditArgs(auditSource)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO intent_proofs").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO confirm_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
}

// expectConsume expects the in-transaction token consume and intent read.
// consumed/expired tokens stop before the intent read.
func expectConsume(mock sqlmock.Sqlmock, intent string, consumed bool, expiresAt time.Time) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}).AddRow(applyProofID, consumed, expiresAt))
	if consumed || time.Now().After(expiresAt) {
		mock.ExpectRollback()
		return
	}
	mock.ExpectExec("UPDATE confirm_tokens SET consumed").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT resolved_intent FROM intent_proofs").
		WillReturnRows(sqlmock.NewRows([]string{"resolved_intent"}).AddRow(intent))
}

// expectCommitTrail expects audit, the mutation, the governance run, the
// mutation mission event, run completion, proof confirmation, and commit.
func expectCommitTrail(mock sqlmock.Sqlmock, eventType, mutation string) {
	mock.ExpectExec("INSERT INTO log_entries").WithArgs(auditArgs(eventType)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(mutation).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO mission_runs").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", eventType, "info", "admin", "governance",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE mission_runs SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", "mission.completed", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed'").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

// previewIntent reads the resolved intent a 202 preview minted.
func previewIntent(t *testing.T, env applyEnvelope) string {
	t.Helper()
	proof, ok := env.Data["intent_proof"].(map[string]any)
	if !ok || env.Data["requires_approval"] != true || env.Data["confirm_token"] == nil {
		t.Fatalf("expected approval payload, got %+v", env.Data)
	}
	intent, _ := proof["resolved_intent"].(string)
	if intent == "" {
		t.Fatalf("preview proof has no resolved_intent: %+v", proof)
	}
	return intent
}

func mockMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
