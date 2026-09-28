package deploymentcontext

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/memory"
)

// MEM-LIST: the SQL read rule (memory.GovernedReadClause) must match the Go
// ownership rule the entry routes use, including rows saved before owner ids
// were recorded (loaded_by label only) and owner ids taking precedence.

func memlistIngest(t *testing.T, svc *Service, req IngestRequest) string {
	t.Helper()
	res, err := svc.Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("ingest %q: %v", req.Title, err)
	}
	return res.ArtifactID
}

func memlistListed(t *testing.T, svc *Service, reader memory.GovernedReader) string {
	t.Helper()
	entries, err := svc.ListEntries(context.Background(), 50, false, reader)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ArtifactID)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func memlistJoin(ids ...string) string {
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func TestMemListRealDB_SQLReadRuleMatchesOwnership(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	svc := serviceWith(db, nil)
	owned := func(id string) map[string]any { return map[string]any{"owner_user_id": id} }
	legacy := memlistIngest(t, svc, IngestRequest{Title: "legacy", Content: "legacy private row", Visibility: "private", UserLabel: "ada"})
	adaOwned := memlistIngest(t, svc, IngestRequest{Title: "owned", Content: "owned private row", Visibility: "private", UserLabel: "ada", ExtraMetadata: owned("u-ada")})
	spoofed := memlistIngest(t, svc, IngestRequest{Title: "spoof", Content: "owned by someone else with ada's label", Visibility: "private", UserLabel: "ada", ExtraMetadata: owned("u-other")})
	team := memlistIngest(t, svc, IngestRequest{Title: "team", Content: "team row", Visibility: "team", TeamID: "memlist-t", UserLabel: "zed", ExtraMetadata: owned("u-zed")})
	teamNoID := memlistIngest(t, svc, IngestRequest{Title: "team without id", Content: "team row without a team", Visibility: "team", UserLabel: "zed", ExtraMetadata: owned("u-zed")})
	org := memlistIngest(t, svc, IngestRequest{Title: "org", Content: "company row", KnowledgeClass: KnowledgeClassCompanyKnowledge, Visibility: "team", TeamID: "memlist-t", UserLabel: "zed"})
	global := memlistIngest(t, svc, IngestRequest{Title: "global", Content: "global row", Visibility: "global", UserLabel: "zed"})

	for _, tc := range []struct {
		name   string
		reader memory.GovernedReader
		want   string
	}{
		{"no identity", memory.GovernedReader{}, memlistJoin(org, global)},
		{"ada by id and label", memory.GovernedReader{UserID: "u-ada", Label: "ada"}, memlistJoin(legacy, adaOwned, org, global)},
		{"label alone never matches an owner id", memory.GovernedReader{UserID: "u-mallory", Label: "ada"}, memlistJoin(legacy, org, global)},
		{"blank user id matches no blank owner", memory.GovernedReader{UserID: "   "}, memlistJoin(org, global)},
		{"proven team member", memory.GovernedReader{UserID: "u-carol", Label: "carol", TeamKeys: []string{memory.GovernedTeamKey("default", "memlist-t")}}, memlistJoin(team, org, global)},
		{"membership in another tenant proves nothing", memory.GovernedReader{UserID: "u-carol", TeamKeys: []string{memory.GovernedTeamKey("t2", "memlist-t")}}, memlistJoin(org, global)},
		{"owner reads a team entry without membership", memory.GovernedReader{UserID: "u-zed", Label: "zed"}, memlistJoin(team, teamNoID, org, global)},
	} {
		if got := memlistListed(t, svc, tc.reader); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
	_ = spoofed
	refs, err := svc.TeamRefs(context.Background())
	if err != nil || len(refs) != 1 || refs[0] != (TeamRef{TenantID: "default", TeamID: "memlist-t"}) {
		t.Fatalf("team refs must name only teams owning team entries: %+v err=%v", refs, err)
	}
}
