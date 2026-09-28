package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/memory"
)

func memEditCall(t *testing.T, h http.Handler, who *RequestIdentity, id, body string) *httptest.ResponseRecorder {
	t.Helper()
	return doAuthenticatedRequestAs(t, h, http.MethodPatch, "/api/v1/memory/deployment-context/"+id, body, who)
}

func memEntryText(t *testing.T, db *sql.DB, id string) (title, content string, chunks int) {
	t.Helper()
	_ = db.QueryRow(`SELECT title, content FROM artifacts WHERE id::text = $1`, id).Scan(&title, &content)
	_ = db.QueryRow(`SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1`, id).Scan(&chunks)
	return title, content, chunks
}

func memAuditCount(t *testing.T, db *sql.DB, ids ...string) int {
	t.Helper()
	var n int
	_ = db.QueryRow(`SELECT count(*) FROM log_entries WHERE level = 'audit' AND context->>'action' = 'deployment_context_edit'
		AND context->>'artifact_id' = ANY($1::text[])`, "{"+strings.Join(ids, ",")+"}").Scan(&n)
	return n
}

func memTeamRecall(t *testing.T, db *sql.DB, query string) map[string]string {
	t.Helper()
	hits, _, err := memory.NewServiceWithDB(db).RecallGoverned(context.Background(), nil, query, memory.SemanticSearchOptions{Limit: 10, TeamID: "team-a"})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	out := map[string]string{}
	for _, hit := range hits {
		id, _ := hit.Metadata["artifact_id"].(string)
		title, _ := hit.Metadata["artifact_title"].(string)
		out[id] = title
	}
	return out
}

const memNewContent = "Team A delivery window moved to one to three, pastries arrive by van."

func TestMemEditRealDB_OwnerEditsAndRecallFollowsAtOnce(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	team := saveAs(t, h, adaOwner, adaTeamBody)
	if hits := memTeamRecall(t, db, "nine eleven"); hits[team] != "Ada team note" {
		t.Fatalf("entry must be recallable before the edit: %v", hits)
	}

	rr := memEditCall(t, h, adaOwner, team, `{"title":"Ada van schedule","content":"`+memNewContent+`","source_label":"driver call"}`)
	assertStatus(t, rr, http.StatusOK)
	var body struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	assertJSON(t, rr, &body)
	if !body.OK || body.Data["changed"] != true || body.Data["audit_event_id"] == "" || body.Data["embedding_status"] != "pending" || body.Data["title"] != "Ada van schedule" {
		t.Fatalf("edit must report the real change: %+v", body)
	}
	title, content, chunks := memEntryText(t, db, team)
	if title != "Ada van schedule" || content != memNewContent || chunks != 1 {
		t.Fatalf("entry not edited: title=%q content=%q chunks=%d", title, content, chunks)
	}
	if hits := memTeamRecall(t, db, "nine eleven"); hits[team] != "" {
		t.Fatalf("old text must no longer be recalled or cited: %v", hits)
	}
	if hits := memTeamRecall(t, db, "pastries van"); hits[team] != "Ada van schedule" {
		t.Fatalf("new text must be recalled and cited under the new title: %v", hits)
	}

	var audit string
	if err := db.QueryRow(`SELECT context::text FROM log_entries WHERE level = 'audit' AND context->>'action' = 'deployment_context_edit'
		AND context->>'artifact_id' = $1 AND context->>'result_status' = 'authorized'`, team).Scan(&audit); err != nil {
		t.Fatalf("the edit must be audited before it happens: %v", err)
	}
	sum := sha256.Sum256([]byte(memNewContent))
	for _, want := range []string{`"old_title": "Ada team note"`, `"new_title": "Ada van schedule"`, hex.EncodeToString(sum[:]), `"user": "ada"`} {
		if !strings.Contains(audit, want) {
			t.Fatalf("audit must carry %s: %s", want, audit)
		}
	}
	for _, leaked := range []string{"pastries", "nine to eleven"} {
		if strings.Contains(audit, leaked) {
			t.Fatalf("audit must never carry content, found %q", leaked)
		}
	}
	if n := memAuditCount(t, db, team); n != 2 {
		t.Fatalf("want authorized + edited audit records, got %d", n)
	}
}

func TestMemEditRealDB_AuthorityMatchesLifecycleRule(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	org := saveAs(t, h, adaOwner, adaOrgBody)
	patch := `{"title":"hijacked","content":"attacker text"}`

	// Not readable, so answered like an unknown id (MEM-LIST).
	for _, who := range []*RequestIdentity{bobOther, adminWithWrite} {
		rr := memEditCall(t, h, who, private, patch)
		assertStatus(t, rr, http.StatusNotFound)
		if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryNotFound {
			t.Fatalf("%s editing a private entry: code %q", who.Username, code)
		}
	}
	for _, who := range []*RequestIdentity{adaOwner, bobOther, adminNoMemory} {
		rr := memEditCall(t, h, who, org, patch)
		assertStatus(t, rr, http.StatusForbidden)
		if code := lifecycleBlockerCode(t, rr); code != codeAdminRequired {
			t.Fatalf("%s editing an org-wide entry: code %q", who.Username, code)
		}
	}
	rr := memEditCall(t, h, adaOwner, "00000000-0000-0000-0000-000000000000", patch)
	assertStatus(t, rr, http.StatusNotFound)
	if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryNotFound {
		t.Fatalf("unknown entry code %q", code)
	}
	assertStatus(t, doRequest(t, h, http.MethodPatch, "/api/v1/memory/deployment-context/"+private, patch), http.StatusUnauthorized)
	for _, id := range []string{private, org} {
		if title, content, _ := memEntryText(t, db, id); title == "hijacked" || content == "attacker text" {
			t.Fatalf("a denied edit changed %s", id)
		}
	}
	if n := memAuditCount(t, db, private, org); n != 0 {
		t.Fatalf("denied edits must not record a change, found %d", n)
	}
	assertStatus(t, memEditCall(t, h, adminWithWrite, org, `{"content":"Closed on public holidays and the first Monday."}`), http.StatusOK)
	if _, content, _ := memEntryText(t, db, org); content != "Closed on public holidays and the first Monday." {
		t.Fatalf("admin with memory:write must edit org-wide entries: %q", content)
	}
}

func TestMemEditRealDB_ArchivedConflictsDeletedIsGone(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	assertStatus(t, lifecycleCall(t, h, adaOwner, "archive", private), http.StatusOK)
	rr := memEditCall(t, h, adaOwner, private, `{"title":"edited while archived"}`)
	assertStatus(t, rr, http.StatusConflict)
	if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryArchived {
		t.Fatalf("archived entry code %q", code)
	}
	if title, _, _ := memEntryText(t, db, private); title != "Ada private note" {
		t.Fatalf("archived entry was edited: %q", title)
	}
	assertStatus(t, lifecycleCall(t, h, adaOwner, "delete", private), http.StatusOK)
	rr = memEditCall(t, h, adaOwner, private, `{"title":"edited after delete"}`)
	assertStatus(t, rr, http.StatusNotFound)
	if code := lifecycleBlockerCode(t, rr); code != codeMemoryEntryNotFound {
		t.Fatalf("deleted entry code %q", code)
	}
	if n := memAuditCount(t, db, private); n != 0 {
		t.Fatalf("refused edits must not be audited as changes, found %d", n)
	}
}

func TestMemEditRealDB_RejectsBadPatchesBeforeAnyLookup(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	for name, body := range map[string]string{
		"unknown field":    `{"title":"x","knowledge_class":"company_knowledge"}`,
		"visibility":       `{"visibility":"global"}`,
		"empty patch":      `{}`,
		"null only":        `{"title":null}`,
		"blank title":      `{"title":"   "}`,
		"blank content":    `{"content":""}`,
		"not json":         `title=x`,
		"array":            `[{"title":"x"}]`,
		"trailing garbage": `{"title":"x"} {"title":"y"}`,
	} {
		for _, id := range []string{private, "00000000-0000-0000-0000-000000000000"} {
			rr := memEditCall(t, h, adaOwner, id, body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("%s on %s: want 400, got %d %s", name, id, rr.Code, rr.Body.String())
			}
		}
	}
	if title, content, _ := memEntryText(t, db, private); title != "Ada private note" || !strings.Contains(content, "flour mill") {
		t.Fatalf("a rejected patch changed the entry: %q", title)
	}
	if n := memAuditCount(t, db, private); n != 0 {
		t.Fatalf("rejected patches must not be audited, found %d", n)
	}
}

func TestMemEditRealDB_AuditUnavailableChangesNothing(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	private := saveAs(t, h, adaOwner, adaPrivateBody)
	noAudit := memoryTestServer(db, chatOnlyBrain())
	noAudit.DB = nil // the activity log is unavailable; the store still works
	rr := memEditCall(t, lifecycleMux(noAudit), adaOwner, private, `{"title":"unaudited","content":"unaudited text"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if code := lifecycleBlockerCode(t, rr); code != codeServiceUnavailable {
		t.Fatalf("edit without audit: code %q", code)
	}
	if title, content, chunks := memEntryText(t, db, private); title != "Ada private note" || strings.Contains(content, "unaudited") || chunks != 1 {
		t.Fatalf("audit failure changed the entry: title=%q chunks=%d", title, chunks)
	}
}

func TestMemEditRealDB_ConcurrentEditAndDeleteThroughHandlers(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	for round := 0; round < 8; round++ {
		id := saveAs(t, h, adaOwner, adaPrivateBody)
		var wg sync.WaitGroup
		var edit, del *httptest.ResponseRecorder
		wg.Add(2)
		go func() {
			defer wg.Done()
			edit = memEditCall(t, h, adaOwner, id, `{"content":"`+strings.Repeat("Race text for the mill. ", 120)+`"}`)
		}()
		go func() { defer wg.Done(); del = lifecycleCall(t, h, adaOwner, "delete", id) }()
		wg.Wait()
		if del.Code != http.StatusOK || (edit.Code != http.StatusOK && edit.Code != http.StatusNotFound) {
			t.Fatalf("round %d: delete=%d edit=%d %s", round, del.Code, edit.Code, edit.Body.String())
		}
		if title, _, chunks := memEntryText(t, db, id); title != "" || chunks != 0 {
			t.Fatalf("round %d: rows survived: title=%q chunks=%d", round, title, chunks)
		}
	}
}
