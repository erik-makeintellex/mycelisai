package server

import (
	"database/sql"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// MEM-LIST (HIGH): GET /api/v1/memory/deployment-context returned every
// entry's id, title and preview to any signed-in user. A viewer now lists
// only what they may read: their own private and team entries, entries of a
// team they are a proven member of, and org-wide entries. Real PostgreSQL
// (MYCELIS_MEMORY_TEST_DSN).

type memlistWorld struct {
	h                      http.Handler
	db                     *sql.DB
	ada, bob, carol, admin *RequestIdentity
	adaPrivate, adaTeam    string
	adaOrg, bobPrivate     string
	bobTeamB, bobGlobal    string
	adaArchived            string
	unknownID              string
	teamA                  string
}

func memlistSeed(t *testing.T) *memlistWorld {
	t.Helper()
	db := openMemoryServerTestDB(t)
	run := uuid.NewString()[:8]
	teamA, teamB := "memlist-a-"+run, "memlist-b-"+run
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id IN ($1, $2)`, teamA, teamB)
	})
	// carol is a proven member of team A (owner group membership), nothing else.
	carolIDs := b1rcSeedTenantTeam(t, db, "default", "memlist-c-"+run, teamA)
	w := &memlistWorld{
		h: lifecycleMux(memoryTestServer(db, chatOnlyBrain())), db: db, teamA: teamA,
		ada:       memoryUser(uuid.NewString(), "ada-"+run, "operator", "memory:read"),
		bob:       memoryUser(uuid.NewString(), "bob-"+run, "operator", "memory:read"),
		carol:     memoryUser(carolIDs.user, "u-memlist-c-"+run, "operator", "memory:read"),
		admin:     memoryUser(uuid.NewString(), "root-"+run, "admin", "memory:write"),
		unknownID: uuid.NewString(),
	}
	w.adaPrivate = saveAs(t, w.h, w.ada, `{"title":"memlist ada private","visibility":"private","content":"Ada private flour mill reminder."}`)
	w.adaTeam = saveAs(t, w.h, w.ada, `{"title":"memlist ada team","visibility":"team","team_id":"`+teamA+`","content":"Team A delivery window."}`)
	w.adaOrg = saveAs(t, w.h, w.ada, `{"title":"memlist ada org","knowledge_class":"company_knowledge","visibility":"global","content":"Closed on public holidays."}`)
	w.bobPrivate = saveAs(t, w.h, w.bob, `{"title":"memlist bob private","knowledge_class":"user_private_context","visibility":"private","content":"Bob private budget notes."}`)
	w.bobTeamB = saveAs(t, w.h, w.bob, `{"title":"memlist bob team b","visibility":"team","team_id":"`+teamB+`","content":"Team B rota."}`)
	w.bobGlobal = saveAs(t, w.h, w.bob, `{"title":"memlist bob global","visibility":"global","content":"Customers park at the back."}`)
	w.adaArchived = saveAs(t, w.h, w.ada, `{"title":"memlist ada archived","visibility":"private","content":"Ada old supplier list."}`)
	assertStatus(t, lifecycleCall(t, w.h, w.ada, "archive", w.adaArchived), http.StatusOK)
	return w
}

// memlistIDs lists the entry ids a viewer sees (nil viewer = no identity).
func memlistIDs(t *testing.T, h http.Handler, who *RequestIdentity, query string) []string {
	t.Helper()
	path := "/api/v1/memory/deployment-context" + query
	var rr = doRequest(t, h, http.MethodGet, path, "")
	if who != nil {
		rr = doAuthenticatedRequestAs(t, h, http.MethodGet, path, "", who)
	}
	assertStatus(t, rr, http.StatusOK)
	var body struct {
		Entries []struct {
			ArtifactID string `json:"artifact_id"`
		} `json:"entries"`
		Count int `json:"count"`
	}
	assertJSON(t, rr, &body)
	ids := make([]string, 0, len(body.Entries))
	for _, e := range body.Entries {
		ids = append(ids, e.ArtifactID)
	}
	if body.Count != len(ids) {
		t.Fatalf("count %d does not match %d entries", body.Count, len(ids))
	}
	sort.Strings(ids)
	return ids
}

func memlistWant(ids ...string) []string {
	sort.Strings(ids)
	return ids
}

func TestMemListRealDB_EachViewerSeesOnlyReadableEntries(t *testing.T) {
	w := memlistSeed(t)
	for _, tc := range []struct {
		name string
		who  *RequestIdentity
		want []string
	}{
		{"ada: own private and team, org-wide", w.ada, memlistWant(w.adaPrivate, w.adaTeam, w.adaOrg, w.bobGlobal)},
		{"bob: own private and team, org-wide", w.bob, memlistWant(w.bobPrivate, w.bobTeamB, w.adaOrg, w.bobGlobal)},
		{"carol: team A member, org-wide", w.carol, memlistWant(w.adaTeam, w.adaOrg, w.bobGlobal)},
		{"root admin: org-wide only (admins manage no private or team entry)", w.admin, memlistWant(w.adaOrg, w.bobGlobal)},
		{"no identity: org-wide only", nil, memlistWant(w.adaOrg, w.bobGlobal)},
	} {
		if got := memlistIDs(t, w.h, tc.who, ""); strings.Join(got, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestMemListRealDB_ArchivedFollowTheSameFilter(t *testing.T) {
	w := memlistSeed(t)
	if got := memlistIDs(t, w.h, w.ada, "?include_archived=true"); !memlistHas(got, w.adaArchived) || memlistHas(got, w.bobPrivate) {
		t.Fatalf("ada with archived: %v", got)
	}
	for _, who := range []*RequestIdentity{w.bob, w.carol, w.admin, nil} {
		got := memlistIDs(t, w.h, who, "?include_archived=true")
		if memlistHas(got, w.adaArchived) || memlistHas(got, w.adaPrivate) {
			t.Fatalf("archived private entry leaked to a non-owner: %v", got)
		}
	}
}

func memlistHas(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestMemListRealDB_LimitCountsOnlyReadableRows(t *testing.T) {
	db := openMemoryServerTestDB(t)
	h := lifecycleMux(memoryTestServer(db, chatOnlyBrain()))
	ada := memoryUser(uuid.NewString(), "ada-limit", "operator")
	bob := memoryUser(uuid.NewString(), "bob-limit", "operator")
	first := saveAs(t, h, ada, `{"title":"memlist limit ada 1","visibility":"private","content":"one"}`)
	second := saveAs(t, h, ada, `{"title":"memlist limit ada 2","visibility":"private","content":"two"}`)
	for i := 0; i < 4; i++ { // newer rows ada may not read
		saveAs(t, h, bob, `{"title":"memlist limit bob","visibility":"private","content":"bob"}`)
	}
	if got := memlistIDs(t, h, ada, "?limit=2"); strings.Join(got, ",") != strings.Join(memlistWant(first, second), ",") {
		t.Fatalf("limit=2 must return ada's two readable rows, got %v", got)
	}
}

// memlistCall runs one entry route: PATCH, archive, restore or delete.
func memlistCall(t *testing.T, h http.Handler, who *RequestIdentity, action, id string) (int, string) {
	t.Helper()
	if action == "patch" {
		rr := memEditCall(t, h, who, id, `{"title":"memlist probe"}`)
		return rr.Code, rr.Body.String()
	}
	rr := lifecycleCall(t, h, who, action, id)
	return rr.Code, rr.Body.String()
}

func TestMemListRealDB_UnreadableIsIndistinguishableFromUnknown(t *testing.T) {
	w := memlistSeed(t)
	cases := []struct {
		who *RequestIdentity
		id  string
	}{
		{w.bob, w.adaPrivate}, {w.bob, w.adaTeam}, {w.bob, w.adaArchived},
		{w.carol, w.adaPrivate}, {w.carol, w.bobTeamB},
		{w.admin, w.adaPrivate}, {w.admin, w.bobPrivate}, {w.admin, w.adaTeam},
	}
	for _, action := range []string{"patch", "archive", "restore", "delete"} {
		for _, tc := range cases {
			wantCode, wantBody := memlistCall(t, w.h, tc.who, action, w.unknownID)
			code, body := memlistCall(t, w.h, tc.who, action, tc.id)
			if wantCode != http.StatusNotFound || code != wantCode || body != wantBody {
				t.Errorf("%s by %s on an unreadable entry: %d %s, unknown id: %d %s", action, tc.who.Username, code, body, wantCode, wantBody)
			}
		}
	}
	var audits int
	_ = w.db.QueryRow(`SELECT count(*) FROM log_entries WHERE level = 'audit' AND context->>'artifact_id' IN ($1, $2, $3, $4)`,
		w.adaPrivate, w.adaTeam, w.bobTeamB, w.bobPrivate).Scan(&audits)
	if audits != 0 {
		t.Fatalf("refused calls must not be audited, found %d", audits)
	}
	if got := memlistIDs(t, w.h, w.ada, "?include_archived=true"); !memlistHas(got, w.adaPrivate) || !memlistHas(got, w.adaTeam) || !memlistHas(got, w.adaArchived) {
		t.Fatalf("refused calls changed ada's entries: %v", got)
	}
}

func TestMemListRealDB_ReadableButNotManagedKeeps403(t *testing.T) {
	w := memlistSeed(t)
	for _, action := range []string{"patch", "archive", "restore", "delete"} {
		code, body := memlistCall(t, w.h, w.carol, action, w.adaTeam)
		if code != http.StatusForbidden || !strings.Contains(body, `"code":"`+codeMemoryEntryNotOwned+`"`) {
			t.Errorf("%s by a team member on the team entry: %d %s", action, code, body)
		}
		code, body = memlistCall(t, w.h, w.bob, action, w.adaOrg)
		if code != http.StatusForbidden || !strings.Contains(body, `"code":"`+codeAdminRequired+`"`) {
			t.Errorf("%s by bob on an org-wide entry: %d %s", action, code, body)
		}
	}
	// The owner still manages their team entry.
	assertStatus(t, memEditCall(t, w.h, w.ada, w.adaTeam, `{"title":"memlist ada team edited"}`), http.StatusOK)
}
