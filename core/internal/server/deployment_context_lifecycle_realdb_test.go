package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func memoryUser(id, name, role string, scopes ...string) *RequestIdentity {
	return &RequestIdentity{UserID: id, Username: name, Role: role, PrincipalType: "local_user", AuthSource: "test", Scopes: scopes}
}

var (
	adaOwner       = memoryUser("user-ada", "ada", "operator", "memory:read")
	bobOther       = memoryUser("user-bob", "bob", "operator", "memory:read")
	adminNoMemory  = memoryUser("user-root", "root", "admin", "groups:read")
	adminWithWrite = memoryUser("user-root", "root", "admin", "memory:write")
)

func lifecycleMux(s *AdminServer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/memory/deployment-context", s.HandleDeploymentContext)
	s.registerDeploymentContextLifecycleRoutes(mux)
	return mux
}

func saveAs(t *testing.T, h http.Handler, who *RequestIdentity, body string) string {
	t.Helper()
	rr := doAuthenticatedRequestAs(t, h, http.MethodPost, "/api/v1/memory/deployment-context", body, who)
	assertStatus(t, rr, http.StatusCreated)
	var saved map[string]any
	assertJSON(t, rr, &saved)
	id, _ := saved["artifact_id"].(string)
	if id == "" {
		t.Fatalf("save returned no id: %+v", saved)
	}
	return id
}

func lifecycleCall(t *testing.T, h http.Handler, who *RequestIdentity, action, id string) *httptest.ResponseRecorder {
	t.Helper()
	path := "/api/v1/memory/deployment-context/" + id
	method := http.MethodDelete
	if action != "delete" {
		method, path = http.MethodPost, path+"/"+action
	}
	return doAuthenticatedRequestAs(t, h, method, path, "", who)
}

func lifecycleBlockerCode(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]any
	assertJSON(t, rr, &body)
	data, _ := body["data"].(map[string]any)
	code, _ := data["code"].(string)
	if body["ok"] != false || body["error"] == "" {
		t.Fatalf("expected a normalized blocker: %+v", body)
	}
	return code
}

func lifecycleState(t *testing.T, db *sql.DB, id string) (state string, chunks int) {
	t.Helper()
	_ = db.QueryRow(`SELECT COALESCE(metadata->>'lifecycle_state', 'active') FROM artifacts WHERE id::text = $1`, id).Scan(&state)
	_ = db.QueryRow(`SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1 AND COALESCE(metadata->>'lifecycle_state', 'active') = 'active'`, id).Scan(&chunks)
	return state, chunks
}

const (
	adaPrivateBody = `{"title":"Ada private note","knowledge_class":"customer_context","visibility":"private","content":"Ada supplier reminder: call the flour mill on Friday."}`
	adaTeamBody    = `{"title":"Ada team note","knowledge_class":"customer_context","visibility":"team","team_id":"team-a","content":"Team A delivery window is nine to eleven."}`
	adaOrgBody     = `{"title":"Ada org note","knowledge_class":"company_knowledge","visibility":"global","content":"The shop closes on public holidays."}`
)

func TestDeploymentContextLifecycleRealDB_OwnerArchivesRestoresDeletes(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	team := saveAs(t, h, adaOwner, adaTeamBody)

	assertStatus(t, lifecycleCall(t, h, adaOwner, "archive", private), http.StatusOK)
	if state, chunks := lifecycleState(t, db, private); state != "archived" || chunks != 0 {
		t.Fatalf("archive: state=%s activeChunks=%d", state, chunks)
	}
	list := doAuthenticatedRequestAs(t, h, http.MethodGet, "/api/v1/memory/deployment-context", "", adaOwner)
	if strings.Contains(list.Body.String(), private) {
		t.Fatal("default list must hide archived entries")
	}
	withArchived := doAuthenticatedRequestAs(t, h, http.MethodGet, "/api/v1/memory/deployment-context?include_archived=true", "", adaOwner)
	if !strings.Contains(withArchived.Body.String(), private) || !strings.Contains(withArchived.Body.String(), `"can_manage":true`) {
		t.Fatalf("show archived must list the entry as manageable: %s", withArchived.Body.String())
	}
	assertStatus(t, lifecycleCall(t, h, adaOwner, "restore", private), http.StatusOK)
	if state, chunks := lifecycleState(t, db, private); state != "active" || chunks != 1 {
		t.Fatalf("restore: state=%s activeChunks=%d", state, chunks)
	}
	assertStatus(t, lifecycleCall(t, h, adaOwner, "archive", team), http.StatusOK)

	assertStatus(t, lifecycleCall(t, h, adaOwner, "delete", private), http.StatusOK)
	var left int
	_ = db.QueryRow(`SELECT (SELECT count(*) FROM artifacts WHERE id::text = $1) + (SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1)`, private).Scan(&left)
	if left != 0 {
		t.Fatalf("delete left %d rows", left)
	}
	var tombstone string
	if err := db.QueryRow(`SELECT context::text FROM log_entries WHERE level = 'audit' AND context->>'action' = 'deployment_context_delete'
		AND context->>'artifact_id' = $1 AND context->>'result_status' = 'deleted'`, private).Scan(&tombstone); err != nil {
		t.Fatalf("delete must leave a tombstone audit record: %v", err)
	}
	for _, leaked := range []string{"flour mill", "Ada private note", "supplier"} {
		if strings.Contains(tombstone, leaked) {
			t.Fatalf("tombstone must not keep content, found %q in %s", leaked, tombstone)
		}
	}
	var audits int
	_ = db.QueryRow(`SELECT count(*) FROM log_entries WHERE level = 'audit' AND context->>'artifact_id' IN ($1, $2)`, private, team).Scan(&audits)
	if audits < 4 {
		t.Fatalf("every lifecycle change must be audited, found %d", audits)
	}
	assertStatus(t, lifecycleCall(t, h, adaOwner, "restore", private), http.StatusNotFound)
}

func TestDeploymentContextLifecycleRealDB_AuthorityNegatives(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	org := saveAs(t, h, adaOwner, adaOrgBody)

	for _, action := range []string{"archive", "restore", "delete"} {
		// Not readable, so answered like an unknown id (MEM-LIST).
		for _, who := range []*RequestIdentity{bobOther, adminWithWrite} {
			rr := lifecycleCall(t, h, who, action, private)
			assertStatus(t, rr, http.StatusNotFound)
			if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryNotFound {
				t.Fatalf("%s by %s on a private entry: code %q", action, who.Username, code)
			}
		}
		for _, who := range []*RequestIdentity{adaOwner, bobOther, adminNoMemory} {
			rr := lifecycleCall(t, h, who, action, org)
			assertStatus(t, rr, http.StatusForbidden)
			if code := lifecycleBlockerCode(t, rr); code != codeAdminRequired {
				t.Fatalf("%s by %s on an org-wide entry: code %q", action, who.Username, code)
			}
		}
	}
	assertStatus(t, doRequest(t, lifecycleMux(memoryTestServer(db, chatOnlyBrain())), http.MethodPost, "/api/v1/memory/deployment-context/"+private+"/archive", ""), http.StatusUnauthorized)
	for _, id := range []string{private, org} {
		if state, chunks := lifecycleState(t, db, id); state != "active" || chunks != 1 {
			t.Fatalf("denied calls changed %s: state=%s chunks=%d", id, state, chunks)
		}
	}
	var audits int
	_ = db.QueryRow(`SELECT count(*) FROM log_entries WHERE level = 'audit' AND context->>'artifact_id' IN ($1, $2)`, private, org).Scan(&audits)
	if audits != 0 {
		t.Fatalf("denied calls must not record a change, found %d audits", audits)
	}

	assertStatus(t, lifecycleCall(t, h, adminWithWrite, "archive", org), http.StatusOK)
	assertStatus(t, lifecycleCall(t, h, adminWithWrite, "delete", org), http.StatusOK)
	rr := lifecycleCall(t, h, adaOwner, "archive", "00000000-0000-0000-0000-000000000000")
	assertStatus(t, rr, http.StatusNotFound)
	if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryNotFound {
		t.Fatalf("unknown entry code %q", code)
	}
}

func TestDeploymentContextLifecycleRealDB_AuditUnavailableChangesNothing(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)

	noAudit := memoryTestServer(db, chatOnlyBrain())
	noAudit.DB = nil // the activity log is unavailable; the store still works
	blocked := lifecycleMux(noAudit)
	for _, action := range []string{"archive", "delete"} {
		rr := lifecycleCall(t, blocked, adaOwner, action, private)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if code := lifecycleBlockerCode(t, rr); code != codeServiceUnavailable {
			t.Fatalf("%s without audit: code %q", action, code)
		}
	}
	if state, chunks := lifecycleState(t, db, private); state != "active" || chunks != 1 {
		t.Fatalf("audit failure changed the entry: state=%s chunks=%d", state, chunks)
	}
}
