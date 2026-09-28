package server

import (
	"net/http"
	"strings"
	"testing"
)

// MEM-LIST read surface: GET /api/v1/memory/search returns governed chunk
// text, so it applies the same read filter as the list. Scope parameters in
// the query (team_id, agent_id, visibility) narrow a search; they never widen
// what the caller may read.

func memlistSearch(t *testing.T, w *memlistWorld, who *RequestIdentity, query string) string {
	t.Helper()
	path := "/api/v1/memory/search?limit=50&q=" + query
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/memory/search", memoryTestServer(w.db, chatOnlyBrain()).HandleMemorySearch)
	rr := doRequest(t, mux, http.MethodGet, path, "")
	if who != nil {
		rr = doAuthenticatedRequestAs(t, mux, http.MethodGet, path, "", who)
	}
	assertStatus(t, rr, http.StatusOK)
	return rr.Body.String()
}

func TestMemListRealDB_SearchAppliesTheReadFilter(t *testing.T) {
	w := memlistSeed(t)
	const ada, team, bob = "flour+mill", "delivery+window", "budget+notes"
	for _, tc := range []struct {
		name, query, scope string
		who                *RequestIdentity
		leaked             string
	}{
		{"bob searching ada's private text", ada, "", w.bob, w.adaPrivate},
		{"bob naming ada's agent scope", ada, "&agent_id=soma", w.bob, w.adaPrivate},
		{"bob naming ada's team", team, "&team_id=" + w.teamA, w.bob, w.adaTeam},
		{"carol searching bob's private text", bob, "", w.carol, w.bobPrivate},
		{"admin naming the admin agent scope", bob, "&agent_id=admin", w.admin, w.bobPrivate},
		{"no identity", ada, "", nil, w.adaPrivate},
	} {
		if body := memlistSearch(t, w, tc.who, tc.query+tc.scope); strings.Contains(body, tc.leaked) {
			t.Errorf("%s: search returned an unreadable entry: %s", tc.name, body)
		}
	}
	for _, tc := range []struct {
		name, query string
		who         *RequestIdentity
		want        string
	}{
		{"ada finds her private entry", ada, w.ada, w.adaPrivate},
		{"carol finds her team's entry", team, w.carol, w.adaTeam},
		{"bob finds his private entry", bob, w.bob, w.bobPrivate},
		{"everyone finds org-wide entries", "public+holidays", w.bob, w.adaOrg},
	} {
		if body := memlistSearch(t, w, tc.who, tc.query); !strings.Contains(body, tc.want) {
			t.Errorf("%s: readable entry missing: %s", tc.name, body)
		}
	}
}
