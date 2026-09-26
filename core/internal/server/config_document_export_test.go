package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	exportRecordID     = "7d9f4c1e-2a3b-4c5d-8e9f-0a1b2c3d4e5f"
	exportSecretMarker = "EXPORTTESTSECRET"
	exportRoute        = "GET /api/v1/config-documents/{recordId}/export"
)

// seededExportDocument is inserted as a raw row, bypassing the validator. All
// secret-looking values are fake fixtures carrying exportSecretMarker.
func seededExportDocument() protocol.ConfigDocument {
	return protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindWorkerProfile,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: "seeded-profile", Name: "Seeded profile", Version: "1", OwnerID: "operator-1",
			Scope:      protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "primary"},
			Enabled:    true,
			Source:     protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceAPI, Ref: "seed"},
			SecretRefs: []string{"env:OPENAI_API_KEY"},
			Governance: protocol.ConfigDocumentGovernance{RiskLevel: protocol.ConfigDocumentRiskLow, ApprovalPosture: protocol.ApprovalPostureOptional},
		},
		Spec: json.RawMessage(`{"api_key":"sk-live-EXPORTTESTSECRET0001","api_key_ref":"env:OPENAI_API_KEY",` +
			`"connection":{"headers":{"Authorization":"Bearer EXPORTTESTSECRET0002xyz"}},` +
			`"args":["--verbose","ghp_EXPORTTESTSECRET0003abcdef"],` +
			`"dsn":"postgres://app:EXPORTTESTSECRET0010pw@db:5432/app","env_line":"export OPENAI_API_KEY=EXPORTTESTSECRET0011",` +
			`"private_key":"-----BEGIN PRIVATE KEY-----\nMIIEXPORTTESTSECRET0005\n-----END PRIVATE KEY-----\n",` +
			`"role":"reviewer"}`),
	}
}

func cleanExportDocument() protocol.ConfigDocument {
	document := seededExportDocument()
	document.Spec = json.RawMessage(`{"role":"reviewer","system_prompt":"Review the output.","capability_refs":["artifact.review"],` +
		`"outputs":["review_report"],"verification_strategy":"semantic","verification_rubric":["Every finding cites evidence"]}`)
	return document
}

var exportRowColumns = []string{
	"record_id", "tenant_id", "document_id", "api_version", "kind", "name", "version",
	"owner_id", "scope_kind", "scope_ref", "enabled", "source_kind", "source_ref",
	"secret_refs", "governance", "spec", "digest", "validation_state", "created_by", "created_at",
}

func expectExportRow(mock sqlmock.Sqlmock, document protocol.ConfigDocument, digest string) {
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	rows := sqlmock.NewRows(exportRowColumns).AddRow(
		exportRecordID, "default", document.Metadata.ID, document.APIVersion, string(document.Kind), document.Metadata.Name,
		document.Metadata.Version, document.Metadata.OwnerID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		document.Metadata.Enabled, string(document.Metadata.Source.Kind), document.Metadata.Source.Ref,
		string(secretRefs), string(governance), string(document.Spec), digest, "valid", "operator-1", time.Unix(1700000000, 0).UTC(),
	)
	mock.ExpectQuery(`FROM config_documents`).WithArgs("default", exportRecordID).WillReturnRows(rows)
}

func exportData(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var response protocol.APIResponse
	assertJSON(t, rr, &response)
	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", response.Data)
	}
	return data
}

func TestConfigDocumentExportNeverReturnsSeededSecrets(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		opt, mock := withDB(t)
		mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
		expectExportRow(mock, seededExportDocument(), "sha256:"+strings.Repeat("a", 64))
		rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export?format="+format, "")
		assertStatus(t, rr, http.StatusOK)
		if strings.Contains(rr.Body.String(), exportSecretMarker) {
			t.Fatalf("%s export body leaked a seeded secret: %s", format, rr.Body.String())
		}
		data := exportData(t, rr)
		if _, present := data["stored_digest"]; present {
			t.Fatalf("%s redacted export must omit stored_digest: %#v", format, data)
		}
		if data["redaction_applied"] != true || data["content_matches_stored_digest"] != false || data["notice"] == nil {
			t.Fatalf("%s redaction posture = %#v", format, data)
		}
		wantPaths := []any{"spec.api_key", "spec.args[1]", "spec.connection.headers.Authorization", "spec.dsn", "spec.env_line", "spec.private_key"}
		if !reflect.DeepEqual(data["redacted_paths"], wantPaths) {
			t.Fatalf("%s redacted_paths = %#v, want %#v", format, data["redacted_paths"], wantPaths)
		}
		content, _ := data["content"].(string)
		if !strings.Contains(content, "env:OPENAI_API_KEY") || !strings.Contains(content, protocol.ConfigDocumentRedactedValue) {
			t.Fatalf("%s content lost refs or markers:\n%s", format, content)
		}
		if data["record_id"] != exportRecordID || data["document_id"] != "seeded-profile" || data["format"] != format {
			t.Fatalf("%s identity fields = %#v", format, data)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations (export must be a single read): %v", err)
		}
	}
}

func TestConfigDocumentExportCleanRevisionMatchesStoredDigestAndIsDeterministic(t *testing.T) {
	document := cleanExportDocument()
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		t.Fatalf("digest clean document: %v", err)
	}
	for _, format := range []string{"yaml", "json"} {
		var hashes []any
		for i := 0; i < 2; i++ {
			opt, mock := withDB(t)
			mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
			expectExportRow(mock, document, digest)
			rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export?format="+format, "")
			assertStatus(t, rr, http.StatusOK)
			data := exportData(t, rr)
			if data["redaction_applied"] != false || data["content_matches_stored_digest"] != true || data["stored_digest"] != digest {
				t.Fatalf("%s clean export posture = %#v", format, data)
			}
			hashes = append(hashes, data["export_sha256"])
		}
		if hashes[0] != hashes[1] || hashes[0] == "" {
			t.Fatalf("%s export_sha256 not deterministic: %#v", format, hashes)
		}
	}
}

func TestConfigDocumentExportDefaultsToYAML(t *testing.T) {
	opt, mock := withDB(t)
	mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
	expectExportRow(mock, cleanExportDocument(), "not-a-digest EXPORTTESTSECRET")
	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusOK)
	if strings.Contains(rr.Body.String(), exportSecretMarker) {
		t.Fatalf("malformed stored digest was echoed: %s", rr.Body.String())
	}
	data := exportData(t, rr)
	if _, present := data["stored_digest"]; present || data["content_matches_stored_digest"] != false {
		t.Fatalf("malformed stored digest must be omitted: %#v", data)
	}
	if data["format"] != "yaml" || !strings.HasPrefix(data["content"].(string), "apiVersion: mycelis.ai/v1\n") {
		t.Fatalf("default export = %#v", data)
	}
}

func TestConfigDocumentExportDeniesWithoutRootAdminScope(t *testing.T) {
	standard := &RequestIdentity{UserID: "u-1", Username: "user", Role: "user", Scopes: []string{"soma:work", "runs:read", "outputs:read"}}
	adminMissingScope := &RequestIdentity{UserID: "u-2", Username: "admin2", Role: "admin", Scopes: []string{"config_documents:write", "groups:read"}}
	nonAdminWithScope := &RequestIdentity{UserID: "u-3", Username: "viewer", Role: "user", Scopes: []string{"config_documents:read"}}
	cases := []struct {
		name     string
		identity *RequestIdentity
		want     int
	}{
		{"anonymous", nil, http.StatusUnauthorized},
		{"standard web user", standard, http.StatusForbidden},
		{"admin missing config_documents:read", adminMissingScope, http.StatusForbidden},
		{"non-admin holding the scope", nonAdminWithScope, http.StatusForbidden},
	}
	for _, tc := range cases {
		opt, mock := withDB(t)
		mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
		path := "/api/v1/config-documents/" + exportRecordID + "/export"
		var rr *httptest.ResponseRecorder
		if tc.identity == nil {
			rr = doRequest(t, mux, http.MethodGet, path, "")
		} else {
			rr = doAuthenticatedRequestAs(t, mux, http.MethodGet, path, "", tc.identity)
		}
		assertStatus(t, rr, tc.want)
		assertNoExportDocumentFields(t, tc.name, rr.Body.String())
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: denied request touched the store: %v", tc.name, err)
		}
	}
}

func TestConfigDocumentExportRejectsBadFormatRecordAndKind(t *testing.T) {
	cases := []struct {
		name, path string
		want       int
	}{
		{"unknown format", "/api/v1/config-documents/" + exportRecordID + "/export?format=xml", http.StatusBadRequest},
		{"uppercase format", "/api/v1/config-documents/" + exportRecordID + "/export?format=YAML", http.StatusBadRequest},
		{"empty format", "/api/v1/config-documents/" + exportRecordID + "/export?format=", http.StatusBadRequest},
		{"repeated format", "/api/v1/config-documents/" + exportRecordID + "/export?format=json&format=yaml", http.StatusBadRequest},
		{"path traversal record", "/api/v1/config-documents/..%2F..%2Fetc%2Fpasswd/export", http.StatusBadRequest},
		{"non-uuid record", "/api/v1/config-documents/not-a-record/export", http.StatusBadRequest},
		{"urn uuid record", "/api/v1/config-documents/urn:uuid:" + exportRecordID + "/export", http.StatusBadRequest},
	}
	for _, tc := range cases {
		opt, mock := withDB(t)
		mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
		rr := doAuthenticatedRequest(t, mux, http.MethodGet, tc.path, "")
		assertStatus(t, rr, tc.want)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: rejected request touched the store: %v", tc.name, err)
		}
	}

	opt, mock := withDB(t)
	mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
	unsupported := seededExportDocument()
	unsupported.Kind = "TemplateBundle"
	expectExportRow(mock, unsupported, "sha256:"+strings.Repeat("b", 64))
	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusBadRequest)
	assertNoExportDocumentFields(t, "unsupported kind", rr.Body.String())
}

func TestConfigDocumentExportErrorPathsCarryNoDocumentFragments(t *testing.T) {
	opt, mock := withDB(t)
	mux := setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
	mock.ExpectQuery(`FROM config_documents`).WillReturnError(errors.New("pq: row sk-live-EXPORTTESTSECRET0001 broke"))
	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusInternalServerError)
	assertNoExportDocumentFields(t, "internal error", rr.Body.String())

	opt, mock = withDB(t)
	mux = setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
	mock.ExpectQuery(`FROM config_documents`).WillReturnRows(sqlmock.NewRows(exportRowColumns))
	rr = doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusNotFound)

	opt, mock = withDB(t)
	mux = setupMux(t, exportRoute, newTestServer(opt).HandleExportConfigDocument)
	broken := seededExportDocument()
	broken.Spec = json.RawMessage(`{"api_key":"sk-EXPORTTESTSECRET0001","api_key":"dup"}`)
	expectExportRow(mock, broken, "sha256:"+strings.Repeat("c", 64))
	rr = doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusInternalServerError)
	assertNoExportDocumentFields(t, "render error", rr.Body.String())

	mux = setupMux(t, exportRoute, newTestServer().HandleExportConfigDocument)
	rr = doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/config-documents/"+exportRecordID+"/export", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func assertNoExportDocumentFields(t *testing.T, name, body string) {
	t.Helper()
	for _, fragment := range []string{exportSecretMarker, "seeded-profile", "content", "redacted_paths", "spec", "TemplateBundle"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("%s: body carries document fragment %q: %s", name, fragment, body)
		}
	}
}
