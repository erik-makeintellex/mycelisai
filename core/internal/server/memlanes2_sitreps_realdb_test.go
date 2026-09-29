package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// MEM-LANES-2 C3: GET /api/v1/memory/sitreps returned any team's sitreps to
// any caller. It needs an identity and returns only the sitreps of teams the
// caller is a proven member of, or every team's for a root admin with
// groups:read. Real PostgreSQL (MYCELIS_MEMORY_TEST_DSN).
func TestMemLanes2SitrepsRealDB_SitrepsAnyUserAnyTeam(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	run := uuid.NewString()[:8]
	teamA, teamB := uuid.NewString(), uuid.NewString()
	clean := func() {
		_, _ = db.Exec(`DELETE FROM sitreps WHERE summary LIKE 'MemLanes2 sitrep%'`)
		_, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = ANY($1)`, "{"+teamA+","+teamB+"}")
		_, _ = db.Exec(`DELETE FROM teams WHERE id::text = ANY($1)`, "{"+teamA+","+teamB+"}")
	}
	t.Cleanup(clean)
	ids := b1rcSeedTenantTeam(t, db, "default", "ml2-sit-"+run, teamA) // ada is a proven member of team A only
	for team, summary := range map[string]string{teamA: "MemLanes2 sitrep alpha 7501.", teamB: "MemLanes2 sitrep bravo 7502."} {
		if _, err := db.Exec(`INSERT INTO teams (id, name) VALUES ($1, $2)`, team, "memlanes2-"+run); err != nil {
			t.Fatalf("seed team: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO sitreps (team_id, time_window_start, time_window_end, summary) VALUES ($1, NOW() - interval '1 hour', NOW(), $2)`, team, summary); err != nil {
			t.Fatalf("seed sitrep: %v", err)
		}
	}
	ada := memoryUser(ids.user, "u-ml2-sit-"+run, "operator", "runs:read")
	bob := memoryUser(uuid.NewString(), "bob-ml2-"+run, "operator", "runs:read")
	root := memoryUser(uuid.NewString(), "root-ml2-"+run, "admin", "groups:read")
	rootNoRead := memoryUser(uuid.NewString(), "root2-ml2-"+run, "admin", "memory:write")
	h := http.HandlerFunc(s.HandleListSitReps)

	for _, tc := range []struct {
		name       string
		who        *RequestIdentity
		query      string
		status     int
		want, deny []string
	}{
		{"no identity", nil, "", http.StatusUnauthorized, nil, []string{"7501", "7502"}},
		{"no identity, team A", nil, "?team_id=" + teamA, http.StatusUnauthorized, nil, []string{"7501", "7502"}},
		{"bob, team A", bob, "?team_id=" + teamA, http.StatusForbidden, nil, []string{"7501", "7502"}},
		{"bob, all", bob, "", http.StatusOK, nil, []string{"7501", "7502"}},
		{"ada, all", ada, "", http.StatusOK, []string{"7501"}, []string{"7502"}},
		{"ada, team A", ada, "?team_id=" + teamA, http.StatusOK, []string{"7501"}, []string{"7502"}},
		{"ada, team B", ada, "?team_id=" + teamB, http.StatusForbidden, nil, []string{"7501", "7502"}},
		{"root admin with groups:read, all", root, "", http.StatusOK, []string{"7501", "7502"}, nil},
		{"root admin with groups:read, team B", root, "?team_id=" + teamB, http.StatusOK, []string{"7502"}, []string{"7501"}},
		{"admin without groups:read, all", rootNoRead, "", http.StatusOK, nil, []string{"7501", "7502"}},
	} {
		path := "/api/v1/memory/sitreps" + tc.query
		var rr *httptest.ResponseRecorder
		if tc.who == nil {
			rr = doRequest(t, h, http.MethodGet, path, "")
		} else {
			rr = doAuthenticatedRequestAs(t, h, http.MethodGet, path, "", tc.who)
		}
		body := rr.Body.String()
		if rr.Code != tc.status {
			t.Errorf("%s: status %d, want %d: %s", tc.name, rr.Code, tc.status, body)
		}
		if tc.status != http.StatusOK {
			var envelope protocol.APIResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil || envelope.OK || envelope.Error == "" {
				t.Errorf("%s: a refusal must use the blocker envelope: %s", tc.name, body)
			}
		}
		for _, marker := range tc.want {
			if !strings.Contains(body, marker) {
				t.Errorf("%s: missing sitrep %s", tc.name, marker)
			}
		}
		for _, marker := range tc.deny {
			if strings.Contains(body, marker) {
				t.Errorf("%s: leaked sitrep %s", tc.name, marker)
			}
		}
	}
}
