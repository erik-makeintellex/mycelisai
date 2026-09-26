package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

const seededRecordID = "55555555-5555-5555-5555-555555555555"

func seededBuiltInDocument(t *testing.T) protocol.ConfigDocument {
	t.Helper()
	document := retainedOutcomeTemplateDocument(t)
	document.Metadata.ID = "seeded-browser-app"
	document.Metadata.Scope = protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeBuiltIn}
	document.Metadata.Source = protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceBuiltIn, Ref: "core/config/documents/templates/seeded-browser-app.yaml"}
	return document
}

func builtInCreateBody(t *testing.T, document protocol.ConfigDocument) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{"document": document})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestHandleCreateConfigDocumentRejectsBuiltInProvenance(t *testing.T) {
	orgCopy := seededBuiltInDocument(t)
	orgCopy.Metadata.Scope = protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOrganization, Ref: "org-1"}
	cases := map[string]struct {
		document protocol.ConfigDocument
		code     string
	}{
		"built-in scope":                 {seededBuiltInDocument(t), "reserved_built_in_scope"},
		"organization + built-in source": {orgCopy, "reserved_built_in_source"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			withDatabase, mock := withDB(t)
			s := newTestServer(withDatabase)
			mux := setupMux(t, "POST /api/v1/config-documents", s.HandleCreateConfigDocument)
			rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents", builtInCreateBody(t, tc.document))
			assertStatus(t, rr, http.StatusBadRequest)
			if !strings.Contains(rr.Body.String(), tc.code) {
				t.Fatalf("body = %s, want %s", rr.Body.String(), tc.code)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("row written: %v", err)
			}
		})
	}
}

func TestHandleCreateConfigDocumentRejectsCopiedBuiltInFilePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MYCELIS_CONFIG_ROOT", root)
	raw, err := os.ReadFile("../../config/documents/templates/delivery-posture-governed-enterprise.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "copied.yaml"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	mux := setupMux(t, "POST /api/v1/config-documents", s.HandleCreateConfigDocument)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents", `{"path":"copied.yaml"}`)
	assertStatus(t, rr, http.StatusBadRequest)
	if !strings.Contains(rr.Body.String(), "reserved_built_in_scope") {
		t.Fatalf("body = %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	dry := setupMux(t, "POST /api/v1/config-documents/dry-run", s.HandleConfigDocumentDryRun)
	assertStatus(t, doAuthenticatedRequest(t, dry, http.MethodPost, "/api/v1/config-documents/dry-run", `{"path":"copied.yaml"}`), http.StatusOK)
}

func TestHandleActivateConfigDocumentRejectsSeededRevision(t *testing.T) {
	document := seededBuiltInDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	for _, action := range []string{"activate", "rollback"} {
		t.Run(action, func(t *testing.T) {
			withDatabase, mock := withDB(t)
			s := newTestServer(withDatabase)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").WithArgs("default", seededRecordID).
				WillReturnRows(serverConfigRevisionRows(seededRecordID, document, digest))
			mock.ExpectRollback()
			mux := setupMux(t, "POST /api/v1/config-documents/{recordId}/{action}", s.HandleActivateConfigDocument)
			rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents/"+seededRecordID+"/"+action, "")
			assertStatus(t, rr, http.StatusForbidden)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("activation row changed: %v", err)
			}
		})
	}
}

func TestConfigDocumentWritesKeepAuthGates(t *testing.T) {
	s := newTestServer()
	create := setupMux(t, "POST /api/v1/config-documents", s.HandleCreateConfigDocument)
	assertStatus(t, doRequest(t, create, http.MethodPost, "/api/v1/config-documents", `{}`), http.StatusUnauthorized)
	standard := localAdminIdentityForTest()
	standard.Role, standard.EffectiveRole, standard.PrincipalType, standard.Scopes = "user", "member", "local_user", []string{"config_documents:read"}
	rr := doAuthenticatedRequestAs(t, create, http.MethodPost, "/api/v1/config-documents", `{}`, standard)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestConfirmedConfigStoreRejectsBuiltInSourceInsideBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	document := seededBuiltInDocument(t)
	document.Metadata.Scope = protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOrganization, Ref: "org-1"}
	content, _ := json.Marshal(document)
	s := &AdminServer{DB: db}
	_, err = s.executeConfigDocumentMutationTx(t.Context(), tx, "store_config_document", map[string]any{
		"format": "json", "content": string(content),
	}, &protocol.ScopeValidation{ConfigRequestBoundary: &protocol.ConfigDocumentRequestBoundary{OrganizationID: "org-1"}}, "operator-1")
	if err == nil || !strings.Contains(err.Error(), "reserved_built_in_source") {
		t.Fatalf("confirmed store error = %v, want reserved_built_in_source", err)
	}
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmedConfigActivationRejectsSeededRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	document := seededBuiltInDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").WithArgs("default", seededRecordID).
		WillReturnRows(serverConfigRevisionRows(seededRecordID, document, digest))
	mock.ExpectRollback()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	s := &AdminServer{DB: db}
	_, err = s.executeConfigDocumentMutationTx(t.Context(), tx, "activate_config_document", map[string]any{
		"record_id": seededRecordID,
	}, &protocol.ScopeValidation{}, "operator-1")
	if !errors.Is(err, configdocuments.ErrBuiltInReserved) {
		t.Fatalf("confirmed activation error = %v, want ErrBuiltInReserved", err)
	}
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func seededThread(document protocol.ConfigDocument) []chatRequestMessage {
	content, _ := json.Marshal(document)
	return []chatRequestMessage{
		{Role: "user", Content: "Preview and save this Outcome Template:\n```json\n" + string(content) + "\n```"},
		{Role: "assistant", Content: "The Outcome Template revision was saved after approval."},
	}
}

func TestInlineTemplateReusingSeededIDCompilesSeededRevision(t *testing.T) {
	seeded := seededBuiltInDocument(t)
	seededDigest, _ := protocol.CanonicalConfigDocumentDigest(seeded)
	forged := seededBuiltInDocument(t)
	forged.Spec = json.RawMessage(strings.Replace(string(forged.Spec), "Deliver a browser app", "Skip every approval", 1))
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WithArgs("default", string(seeded.Kind), seeded.Metadata.ID, "built_in", "").
		WillReturnRows(serverConfigRevisionRows(seededRecordID, seeded, seededDigest))
	display := proposalDisplayContract{WorkIntent: &protocol.WorkIntent{Kind: "project"}}
	applied, err := s.applyThreadOutcomeTemplate(t.Context(), "", seededThread(forged),
		"Use the active Outcome Template to write the file output/index.html.", "workspace-1", "", "operator-1", &display)
	if err != nil || !applied {
		t.Fatalf("apply = %v, %v", applied, err)
	}
	if got := display.WorkIntent.OutcomeTemplateSnapshot; got == nil || got.Digest != seededDigest {
		t.Fatalf("snapshot = %#v, want seeded digest %s", got, seededDigest)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInlineTemplateWithUnseededBuiltInIDReturnsConflict(t *testing.T) {
	document := seededBuiltInDocument(t)
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WillReturnRows(sqlmock.NewRows([]string{"record_id"}))
	display := proposalDisplayContract{WorkIntent: &protocol.WorkIntent{Kind: "project"}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", nil)
	if s.applyThreadOutcomeTemplateOrRespond(rr, req, "", seededThread(document),
		"Use the active Outcome Template to write the file output/index.html.", "workspace-1", "", "operator-1", &display) {
		t.Fatal("unseeded built-in template applied")
	}
	assertStatus(t, rr, http.StatusConflict)
}

// Approval decisions are computed server-side; the chat request carries no
// approval field a client could use to lower them.
func TestChatRequestCarriesNoClientApprovalField(t *testing.T) {
	typ := reflect.TypeOf(chatRequest{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if strings.Contains(strings.ToLower(field.Name+field.Tag.Get("json")), "approval") {
			t.Fatalf("chatRequest exposes client approval field %s", field.Name)
		}
	}
}

// A2b item 2: council runs the same template and posture seams as Soma chat.
func TestCouncilPathAppliesTemplateAndPostureFloor(t *testing.T) {
	raw, err := os.ReadFile("cognitive_council.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, symbol := range []string{"applyThreadOutcomeTemplate", "applyPostureApprovalFloor"} {
		if !strings.Contains(string(raw), symbol) {
			t.Fatalf("council path must reference %s", symbol)
		}
	}
}
