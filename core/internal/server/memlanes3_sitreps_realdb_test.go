package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// MEM-LANES-3 item 3: a member of the same team id in another tenant read the
// default tenant's SitReps. Membership must be proven in the SitRep's tenant
// (Core's Archivist writes SitReps in tenant "default").
func TestMemLanes3SitrepsRealDB_OtherTenantMemberOfSameTeamID(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	run := uuid.NewString()[:8]
	team := uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM sitreps WHERE summary LIKE 'MemLanes3 sitrep%'`)
		_, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = $1`, team)
		_, _ = db.Exec(`DELETE FROM teams WHERE id::text = $1`, team)
	})
	home := b1rcSeedTenantTeam(t, db, "default", "ml3-home-"+run, team)
	other := b1rcSeedTenantTeam(t, db, "memlanes3-t2", "ml3-t2-"+run, team)
	if _, err := db.Exec(`INSERT INTO teams (id, name) VALUES ($1, 'memlanes3')`, team); err != nil {
		t.Fatalf("seed team: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sitreps (team_id, time_window_start, time_window_end, summary) VALUES ($1, NOW(), NOW(), 'MemLanes3 sitrep 9502.')`, team); err != nil {
		t.Fatalf("seed sitrep: %v", err)
	}
	h := http.HandlerFunc(s.HandleListSitReps)
	get := func(who *RequestIdentity, q string) (int, string) {
		rr := doAuthenticatedRequestAs(t, h, http.MethodGet, "/api/v1/memory/sitreps"+q, "", who)
		return rr.Code, rr.Body.String()
	}
	homeUser := memoryUser(home.user, "u-ml3-home-"+run, "operator")
	otherUser := memoryUser(other.user, "u-ml3-t2-"+run, "operator")
	if code, body := get(homeUser, "?team_id="+team); code != http.StatusOK || !strings.Contains(body, "9502") {
		t.Errorf("the default-tenant member must read the team's SitRep: %d %s", code, body)
	}
	if code, body := get(otherUser, "?team_id="+team); code != http.StatusForbidden || strings.Contains(body, "9502") {
		t.Errorf("another tenant's member of the same team id: %d %s", code, body)
	}
	if code, body := get(otherUser, ""); code != http.StatusOK || strings.Contains(body, "9502") {
		t.Errorf("another tenant's member listing all: %d %s", code, body)
	}
}
