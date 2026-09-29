package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/pkg/protocol"
)

// MEM-LANES-3: the /api/v1/inception/recipes routes returned and changed
// every recipe for any caller. Reads follow the owner-lane rule the tool path
// uses; changes need the owner (private/team) or root admin with
// memory:write (org-wide), audit first; unknown and unreadable ids answer
// alike. Real PostgreSQL (MYCELIS_MEMORY_TEST_DSN).

func memlanes3InceptionMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/inception/recipes", s.HandleListInceptionRecipes)
	mux.HandleFunc("GET /api/v1/inception/recipes/search", s.HandleSearchInceptionRecipes)
	mux.HandleFunc("GET /api/v1/inception/recipes/{id}", s.HandleGetInceptionRecipe)
	mux.HandleFunc("POST /api/v1/inception/recipes", s.HandleCreateInceptionRecipe)
	mux.HandleFunc("PATCH /api/v1/inception/recipes/{id}/quality", s.HandleUpdateRecipeQuality)
	return mux
}

// memlanes3SeedRecipe writes a recipe the way the tool path does: the row
// plus its owner-stamped vector row.
func memlanes3SeedRecipe(t *testing.T, db *sql.DB, title, owner, visibility string) string {
	t.Helper()
	id, err := inception.NewStore(db).CreateRecipe(context.Background(), inception.Recipe{Category: "memlanes3", Title: title, IntentPattern: title + " pattern"})
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}
	meta := map[string]any{"type": "inception_recipe", "source": "inception_recipe", "recipe_id": id, "tenant_id": "default", "visibility": visibility, "owner_user_id": owner}
	raw, _ := json.Marshal(meta)
	if _, err := db.Exec(`INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, NULL, $2)`, "[inception:memlanes3] "+title, raw); err != nil {
		t.Fatalf("seed recipe owner row: %v", err)
	}
	return id
}

func TestMemLanes3InceptionRealDB_RecipeRoutesFollowTheOwner(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	s.Inception = inception.NewStore(db)
	clean := func() {
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE content LIKE '%MemLanes3 wombat%'`)
		_, _ = db.Exec(`DELETE FROM inception_recipes WHERE title LIKE 'MemLanes3 wombat%'`)
	}
	clean()
	t.Cleanup(clean)
	mux := memlanes3InceptionMux(s)
	ada := memoryUser(uuid.NewString(), "ada-ml3", "operator", "memory:read")
	bob := memoryUser(uuid.NewString(), "bob-ml3", "operator", "memory:read")
	root := memoryUser(uuid.NewString(), "root-ml3", "admin", "memory:write")
	adaID := memlanes3SeedRecipe(t, db, "MemLanes3 wombat ada private 9601", ada.UserID, "private")
	orgID := memlanes3SeedRecipe(t, db, "MemLanes3 wombat org 9602", root.UserID, "global")

	call := func(who *RequestIdentity, method, path, body string) *httptest.ResponseRecorder {
		if who == nil {
			return doRequest(t, mux, method, path, body)
		}
		return doAuthenticatedRequestAs(t, mux, method, path, body, who)
	}
	refused := func(label string, rr *httptest.ResponseRecorder, status int) {
		t.Helper()
		var env protocol.APIResponse
		if rr.Code != status || json.Unmarshal(rr.Body.Bytes(), &env) != nil || env.OK || env.Error == "" {
			t.Errorf("%s: want a %d blocker, got %d %s", label, status, rr.Code, rr.Body.String())
		}
		if strings.Contains(rr.Body.String(), "9601") {
			t.Errorf("%s leaked A's recipe", label)
		}
	}
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/api/v1/inception/recipes", ""}, {"GET", "/api/v1/inception/recipes/search?q=wombat", ""},
		{"GET", "/api/v1/inception/recipes/" + adaID, ""}, {"POST", "/api/v1/inception/recipes", `{"category":"x","title":"MemLanes3 wombat anon","intent_pattern":"x"}`},
		{"PATCH", "/api/v1/inception/recipes/" + adaID + "/quality", `{"score":0.9}`},
	} {
		refused("no identity "+c.method+" "+c.path, call(nil, c.method, c.path, c.body), http.StatusUnauthorized)
	}

	for _, path := range []string{"/api/v1/inception/recipes?category=memlanes3", "/api/v1/inception/recipes/search?q=wombat"} {
		body := call(bob, "GET", path, "").Body.String()
		if strings.Contains(body, "9601") || !strings.Contains(body, "9602") {
			t.Errorf("B %s: want only the org-wide recipe: %s", path, body)
		}
		if body := call(ada, "GET", path, "").Body.String(); !strings.Contains(body, "9601") || !strings.Contains(body, "9602") {
			t.Errorf("A %s: want own and org-wide recipes: %s", path, body)
		}
	}
	unknown := call(bob, "GET", "/api/v1/inception/recipes/"+uuid.NewString(), "")
	hidden := call(bob, "GET", "/api/v1/inception/recipes/"+adaID, "")
	refused("B get A's recipe", hidden, http.StatusNotFound)
	if hidden.Body.String() != unknown.Body.String() {
		t.Errorf("an unreadable recipe must answer exactly like an unknown id:\n%s\n%s", hidden.Body.String(), unknown.Body.String())
	}
	if rr := call(ada, "GET", "/api/v1/inception/recipes/"+adaID, ""); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "9601") {
		t.Errorf("A get own recipe: %d %s", rr.Code, rr.Body.String())
	}

	quality := func(id string) float64 {
		var q float64
		_ = db.QueryRow(`SELECT quality_score FROM inception_recipes WHERE id = $1`, id).Scan(&q)
		return q
	}
	unknownPatch := call(bob, "PATCH", "/api/v1/inception/recipes/"+uuid.NewString()+"/quality", `{"score":0.1}`)
	hiddenPatch := call(bob, "PATCH", "/api/v1/inception/recipes/"+adaID+"/quality", `{"score":0.1}`)
	refused("B modify A's recipe", hiddenPatch, http.StatusNotFound)
	if hiddenPatch.Body.String() != unknownPatch.Body.String() {
		t.Errorf("a change to an unreadable recipe must answer like an unknown id")
	}
	refused("B modify an org-wide recipe", call(bob, "PATCH", "/api/v1/inception/recipes/"+orgID+"/quality", `{"score":0.1}`), http.StatusForbidden)
	if quality(adaID) != 0 || quality(orgID) != 0 {
		t.Fatalf("a refused change moved a quality score: ada=%v org=%v", quality(adaID), quality(orgID))
	}
	if rr := call(ada, "PATCH", "/api/v1/inception/recipes/"+adaID+"/quality", `{"score":0.9}`); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "audit_event_id") || quality(adaID) != 0.9 {
		t.Errorf("A rates own recipe: %d %s (quality %v)", rr.Code, rr.Body.String(), quality(adaID))
	}
	if rr := call(root, "PATCH", "/api/v1/inception/recipes/"+orgID+"/quality", `{"score":0.8}`); rr.Code != http.StatusOK || quality(orgID) != 0.8 {
		t.Errorf("root admin with memory:write rates an org-wide recipe: %d %s", rr.Code, rr.Body.String())
	}

	// A new recipe belongs to its creator; org-wide needs root admin with memory:write.
	refused("B creates an org-wide recipe", call(bob, "POST", "/api/v1/inception/recipes", `{"category":"memlanes3","title":"MemLanes3 wombat bob org 9604","intent_pattern":"x","visibility":"global"}`), http.StatusForbidden)
	if rr := call(bob, "POST", "/api/v1/inception/recipes", `{"category":"memlanes3","title":"MemLanes3 wombat bob 9603","intent_pattern":"x"}`); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "audit_event_id") {
		t.Fatalf("B creates own recipe: %d %s", rr.Code, rr.Body.String())
	}
	if body := call(bob, "GET", "/api/v1/inception/recipes/search?q=wombat", "").Body.String(); !strings.Contains(body, "9603") {
		t.Errorf("B reads own new recipe: %s", body)
	}
	if body := call(ada, "GET", "/api/v1/inception/recipes/search?q=wombat", "").Body.String(); strings.Contains(body, "9603") || strings.Contains(body, "9604") {
		t.Errorf("A reads B's recipe: %s", body)
	}
}

// A team recipe is readable by a proven team member, but only its owner may
// change it (M2 manage rule); a team share on create needs membership.
func TestMemLanes3InceptionRealDB_TeamRecipeReadableNotManageableByMember(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	s.Inception = inception.NewStore(db)
	run := uuid.NewString()[:8]
	team := "memlanes3-team-" + run
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE content LIKE '%MemLanes3 numbat%'`)
		_, _ = db.Exec(`DELETE FROM inception_recipes WHERE title LIKE 'MemLanes3 numbat%'`)
		_, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = $1`, team)
	})
	ids := b1rcSeedTenantTeam(t, db, "default", "ml3-inc-"+run, team)
	var cyID string
	if err := db.QueryRow(`INSERT INTO users (id, username, account_id) VALUES (gen_random_uuid(), $1, $2) RETURNING id`, "u-ml3-cy-"+run, ids.account).Scan(&cyID); err != nil {
		t.Fatalf("seed cy: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO org_memberships (account_id, user_id, group_id, role_id) VALUES ($1,$2,$3,$4)`, ids.account, cyID, ids.group, ids.role); err != nil {
		t.Fatalf("seed cy membership: %v", err)
	}
	ada := memoryUser(ids.user, "u-ml3-inc-"+run, "operator", "memory:read")
	cy := memoryUser(cyID, "u-ml3-cy-"+run, "operator", "memory:read")
	bob := memoryUser(uuid.NewString(), "bob-ml3-"+run, "operator", "memory:read")
	mux := memlanes3InceptionMux(s)
	mux.Handle("/api/v1/memory/", lifecycleMux(s))
	// Team keys are proven for teams with a governed team entry (memoryReader).
	saveAs(t, mux, ada, `{"title":"MemLanes3 numbat team note","visibility":"team","team_id":"`+team+`","content":"MemLanes3 numbat team note."}`)

	rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/inception/recipes", `{"category":"memlanes3","title":"MemLanes3 numbat team recipe 9701","intent_pattern":"x","visibility":"team","team_id":"`+team+`"}`, ada)
	var created struct {
		Data map[string]string `json:"data"`
	}
	if rr.Code != http.StatusOK || json.Unmarshal(rr.Body.Bytes(), &created) != nil || created.Data["id"] == "" {
		t.Fatalf("A creates a team recipe: %d %s", rr.Code, rr.Body.String())
	}
	id := created.Data["id"]
	if rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/inception/recipes", `{"category":"memlanes3","title":"MemLanes3 numbat bob team 9702","intent_pattern":"x","visibility":"team","team_id":"`+team+`"}`, bob); rr.Code != http.StatusForbidden {
		t.Errorf("a non-member shares with the team: %d %s", rr.Code, rr.Body.String())
	}
	if rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/inception/recipes/"+id, "", cy); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "9701") {
		t.Errorf("team member C reads the team recipe: %d %s", rr.Code, rr.Body.String())
	}
	if rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/inception/recipes/"+id, "", bob); rr.Code != http.StatusNotFound {
		t.Errorf("non-member B reads the team recipe: %d %s", rr.Code, rr.Body.String())
	}
	if rr := doAuthenticatedRequestAs(t, mux, "PATCH", "/api/v1/inception/recipes/"+id+"/quality", `{"score":0.2}`, cy); rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "memory_entry_not_owned") {
		t.Errorf("team member C changes A's recipe: %d %s", rr.Code, rr.Body.String())
	}
	var q float64
	_ = db.QueryRow(`SELECT quality_score FROM inception_recipes WHERE id = $1`, id).Scan(&q)
	if q != 0 {
		t.Errorf("a refused change moved the quality score to %v", q)
	}
}

// Audit first: with no audit store, a recipe write is refused and nothing
// changes.
func TestMemLanes3InceptionRealDB_NoAuditNoChange(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	s.Inception = inception.NewStore(db)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE content LIKE '%MemLanes3 dunnart%'`)
		_, _ = db.Exec(`DELETE FROM inception_recipes WHERE title LIKE 'MemLanes3 dunnart%'`)
	})
	ada := memoryUser(uuid.NewString(), "ada-ml3-audit", "operator", "memory:read")
	id := memlanes3SeedRecipe(t, db, "MemLanes3 dunnart ada 9801", ada.UserID, "private")
	s.DB = nil // the audit store is unavailable; the recipe and memory stores still work
	mux := memlanes3InceptionMux(s)
	if rr := doAuthenticatedRequestAs(t, mux, "PATCH", "/api/v1/inception/recipes/"+id+"/quality", `{"score":0.7}`, ada); rr.Code != http.StatusServiceUnavailable {
		t.Errorf("quality change without audit: %d %s", rr.Code, rr.Body.String())
	}
	if rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/inception/recipes", `{"category":"memlanes3","title":"MemLanes3 dunnart new 9802","intent_pattern":"x"}`, ada); rr.Code != http.StatusServiceUnavailable {
		t.Errorf("create without audit: %d %s", rr.Code, rr.Body.String())
	}
	var q float64
	var created int
	_ = db.QueryRow(`SELECT quality_score FROM inception_recipes WHERE id = $1`, id).Scan(&q)
	_ = db.QueryRow(`SELECT count(*) FROM inception_recipes WHERE title LIKE 'MemLanes3 dunnart new%'`).Scan(&created)
	if q != 0 || created != 0 {
		t.Errorf("an unaudited write changed state: quality=%v created=%d", q, created)
	}
}
