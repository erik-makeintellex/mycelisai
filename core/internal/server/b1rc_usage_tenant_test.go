package server

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// B1R-C condition 2 (MED): usage reads are tenant-keyed. A non-admin reads in
// the tenant its membership was proven in; a scoped root admin reads tenant
// "default". QA showed a tenant-t2 member reading tenant default's usage for
// the same team_id (200) and both tenants sharing one counter.

func b1rcCharge(t *testing.T, router *cognitive.Router, correlation cognitive.InferenceCorrelation) {
	t.Helper()
	ctx := cognitive.WithExecutionMeter(context.Background(), cognitive.NewExecutionMeter(cognitive.ExecutionKindAgentTurn, correlation))
	if _, err := router.InferWithContract(ctx, cognitive.InferRequest{Profile: "chat", Prompt: "p", Correlation: correlation}); err != nil {
		t.Fatal(err)
	}
}

func b1rcReadTeam(t *testing.T, mux *http.ServeMux, team string, identity *RequestIdentity) (int, map[string]any) {
	t.Helper()
	rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id="+team, "", identity)
	if rr.Code != http.StatusOK {
		return rr.Code, nil
	}
	return rr.Code, tokenBudgetData(t, rr)
}

func TestB1rcTokenBudgetUsageReadsInTheProvenTenant(t *testing.T) {
	user := &RequestIdentity{UserID: b1rbMemberID, Role: "user"}
	for _, tc := range []struct {
		name, proven string
		want         float64
	}{{"default member", "default", 1124}, {"t2 member", "t2", 2248}, {"member of a tenant with no usage", "t9", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			dbOpt, mock := withDirectDB(t)
			mock.ExpectQuery(b1rbMemberQuery).WithArgs("secret-team", b1rbMemberID).
				WillReturnRows(sqlmock.NewRows([]string{"tenant"}).AddRow(tc.proven))
			s, mux := b1rbUsageServer(t, dbOpt) // tenant default: 1124 on secret-team
			t2 := cognitive.InferenceCorrelation{TenantID: "t2", RunID: "r-1", TeamID: "secret-team", AgentID: "x"}
			b1rcCharge(t, s.Cognitive, t2)
			b1rcCharge(t, s.Cognitive, t2)
			code, data := b1rcReadTeam(t, mux, "secret-team", user)
			if code != http.StatusOK || data["used"] != tc.want {
				t.Fatalf("proven %q read = %d %v, want used %v", tc.proven, code, data, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("scoped root admin reads tenant default", func(t *testing.T) {
		s, mux := b1rbUsageServer(t)
		b1rcCharge(t, s.Cognitive, cognitive.InferenceCorrelation{TenantID: "t2", TeamID: "secret-team"})
		admin := &RequestIdentity{UserID: "a-1", Role: "admin", Scopes: []string{"cognitive:read"}}
		if code, data := b1rcReadTeam(t, mux, "secret-team", admin); code != http.StatusOK || data["used"] != float64(1124) {
			t.Fatalf("admin read = %d %v, want tenant default 1124", code, data)
		}
	})
}

type b1rcIDs struct{ user, account, group, role string }

func b1rcSeedTenantTeam(t *testing.T, db *sql.DB, tenant, slug, team string) b1rcIDs {
	t.Helper()
	var ids b1rcIDs
	scan := func(into *string, query string, args ...any) {
		if err := db.QueryRow(query, args...).Scan(into); err != nil {
			t.Fatalf("seed %s: %v", slug, err)
		}
	}
	scan(&ids.account, `INSERT INTO accounts (tenant_id, slug, name) VALUES ($1,$2,$2) RETURNING id`, tenant, slug)
	scan(&ids.group, `INSERT INTO groups (account_id, key, name) VALUES ($1,'g','g') RETURNING id`, ids.account)
	scan(&ids.role, `INSERT INTO roles (account_id, key, name) VALUES ($1,'member','member') RETURNING id`, ids.account)
	scan(&ids.user, `INSERT INTO users (id, username, account_id) VALUES (gen_random_uuid(), $1, $2) RETURNING id`, "u-"+slug, ids.account)
	var membership string
	scan(&membership, `INSERT INTO org_memberships (account_id, user_id, group_id, role_id) VALUES ($1,$2,$3,$4) RETURNING id`,
		ids.account, ids.user, ids.group, ids.role)
	if _, err := db.Exec(`INSERT INTO runtime_team_manifests (tenant_id, team_id, manifest_digest, manifest, owner_account_id, owner_group_id, ownership_provisioned_by, ownership_provisioned_at)
		VALUES ($1,$2,'d',jsonb_build_object('id',$2::text),$3,$4,$5,NOW())`, tenant, team, ids.account, ids.group, ids.user); err != nil {
		t.Fatalf("seed team %s: %v", slug, err)
	}
	return ids
}

// Real PostgreSQL (MYCELIS_TOKEN_BUDGET_TEST_DSN): the same team_id in tenants
// default and t2. Each member reads only its own tenant's usage, the two
// tenants never share a counter, and ledger rows carry the tenant.
func TestB1rcTokenBudgetUsageRealDBTenantIsolation(t *testing.T) {
	db := openTokenBudgetRealDB(t)
	if _, err := db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = 'b1rc-shared'`); err != nil {
		t.Fatalf("clean: %v", err)
	}
	run := uuid.NewString()[:8]
	a := b1rcSeedTenantTeam(t, db, "default", "b1rc-a-"+run, "b1rc-shared")
	c := b1rcSeedTenantTeam(t, db, "b1rc-t2", "b1rc-c-"+run, "b1rc-shared")
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"b1rc-shared": {PerTeamDay: 2400}}
	router := tokenBudgetRouter()
	router.Adapters["local"] = &b1rbCountingProvider{}
	router.Budgets = cognitive.NewBudgetGovernor(spec, cognitive.NewPostgresBudgetLedger(db))
	s := newTestServer(func(s *AdminServer) { s.Cognitive = router; s.DB = db })
	mux := tokenBudgetMux(s)
	def := cognitive.InferenceCorrelation{RunID: "b1rc-run", TeamID: "b1rc-shared", AgentID: "b1rc-agent"}
	t2 := def
	t2.TenantID = "b1rc-t2"

	// Tenant default spends its team-day cap: 1124 + 1124, then a stop.
	b1rcCharge(t, router, def)
	b1rcCharge(t, router, def)
	ctx := cognitive.WithExecutionMeter(context.Background(), cognitive.NewExecutionMeter(cognitive.ExecutionKindAgentTurn, def))
	if _, err := router.InferWithContract(ctx, cognitive.InferRequest{Profile: "chat", Prompt: "p", Correlation: def}); cognitive.AsTokenBudgetExhausted(err) == nil {
		t.Fatalf("tenant default not stopped: %v", err)
	}
	// Tenant t2's same-named team is not stopped by it.
	b1rcCharge(t, router, t2)

	for _, tc := range []struct {
		name string
		user string
		want float64
	}{{"default member", a.user, 2248}, {"t2 member", c.user, 1124}} {
		if code, data := b1rcReadTeam(t, mux, "b1rc-shared", &RequestIdentity{UserID: tc.user, Role: "user"}); code != http.StatusOK || data["used"] != tc.want {
			t.Fatalf("%s read = %d %v, want used %v", tc.name, code, data, tc.want)
		}
	}
	var tenants int
	if err := db.QueryRow(`SELECT count(DISTINCT tenant_id) FROM token_usage_ledger WHERE team_id='b1rc-shared' AND tenant_id IN ('default','b1rc-t2') AND outcome='charged'`).Scan(&tenants); err != nil || tenants != 2 {
		t.Fatalf("ledger tenants = %d err=%v", tenants, err)
	}
	restarted := cognitive.NewBudgetGovernor(spec, cognitive.NewPostgresBudgetLedger(db))
	limits := protocol.TokenBudgetLimits{PerTeamDay: 2400}
	if u := restarted.Usage(context.Background(), "b1rc-t2", protocol.TokenBudgetScopeTeamDay, "b1rc-shared", limits); u.Used != 1124 || u.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("restarted t2 usage = %+v", u)
	}
}
