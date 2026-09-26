package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	wrItemsSQL    = "FROM team_work_items\n\t\tWHERE tenant_id='default' AND state NOT IN ('archived','output_ready')"
	wrOutcomesSQL = "FROM outcome_projects\n\t\tWHERE tenant_id='default' AND status <> 'archived' AND work_item_refs ?| $1::text[]"
	wrItemA       = "aaaaaaaa-0000-0000-0000-000000000001"
	wrItemB       = "aaaaaaaa-0000-0000-0000-000000000002"
	wrItemC       = "aaaaaaaa-0000-0000-0000-000000000003"
	wrProject1    = "bbbbbbbb-0000-0000-0000-000000000001"
	wrProject2    = "bbbbbbbb-0000-0000-0000-000000000002"
)

var wrNow = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

type wrItemRow struct {
	id, team, state, objective string
	needsOperator              bool
	age                        time.Duration
}

func wrItemRows(rows ...wrItemRow) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{
		"id", "team_id", "run_id", "intent_proof_id", "contract_id", "proof_id", "objective", "scope", "owner",
		"execution_shape", "execution_mode", "work_intent", "expected_outputs", "expected_proof", "capability_requirements",
		"governance_posture", "state", "last_event", "needs_operator", "degradation_state", "recovery_options",
		"output_refs", "proof_refs", "audit_refs", "created_at", "updated_at", "version",
	})
	for _, r := range rows {
		updated := wrNow.Add(-r.age)
		out.AddRow(r.id, r.team, "", "", "", "", r.objective, []byte(`[]`), "soma",
			"delegated_work", "team", []byte(`{"prompt":"SECRET-PROMPT-BODY"}`), []byte(`["report"]`), []byte(`[]`), []byte(`[]`),
			"propose", r.state, []byte(`{"headline":"LAST-EVENT-PAYLOAD"}`), r.needsOperator, "", []byte(`[]`),
			[]byte(`[{"output_id":"OUTPUT-REF-BODY"}]`), []byte(`["PROOF-REF-BODY"]`), []byte(`[]`), updated, updated, "v1")
	}
	return out
}

type wrProjectRow struct {
	id, outcome, title, status string
	refs                       []string
	recovery                   []string
}

func wrProjectRows(rows ...wrProjectRow) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{
		"id", "outcome_id", "title", "purpose", "execution_mode", "workspace_folder", "status", "run_id",
		"intent_proof_id", "contract_id", "proof_id", "work_item_refs", "output_refs", "proof_refs",
		"recovery_refs", "retention_policy", "created_at", "updated_at", "version",
	})
	for _, r := range rows {
		refs, _ := json.Marshal(r.refs)
		recovery, _ := json.Marshal(append([]string{}, r.recovery...))
		out.AddRow(r.id, r.outcome, r.title, "", "project", "", r.status, "", "", "", "",
			refs, []byte(`[]`), []byte(`[]`), recovery, "retained", wrNow, wrNow, "v1")
	}
	return out
}

func wrExpectItems(mock sqlmock.Sqlmock, limitArg int, rows *sqlmock.Rows) {
	mock.ExpectQuery(regexp.QuoteMeta(wrItemsSQL)).WithArgs(limitArg).WillReturnRows(rows)
}

func wrExpectOutcomes(mock sqlmock.Sqlmock, ids []string, rows *sqlmock.Rows) {
	mock.ExpectQuery(regexp.QuoteMeta(wrOutcomesSQL)).WithArgs(`{"` + strings.Join(ids, `","`) + `"}`).WillReturnRows(rows)
}

func wrGet(t *testing.T, s *AdminServer, path string, identity *RequestIdentity) (int, string) {
	t.Helper()
	mux := setupMux(t, "GET /api/v1/work/running", s.HandleWorkRunning)
	rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, path, "", identity)
	return rr.Code, rr.Body.String()
}

func wrDecode(t *testing.T, body string) protocol.WorkRunningView {
	t.Helper()
	var env struct {
		OK   bool                     `json:"ok"`
		Data protocol.WorkRunningView `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil || !env.OK {
		t.Fatalf("decode running view: ok=%v err=%v body=%s", env.OK, err, body)
	}
	return env.Data
}

func wrMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations (exact query count, no writes): %v", err)
	}
}

func TestWorkRunning_GroupsByOutcomeAndTeamWithOneItemsQuery(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	long := strings.Repeat("é", 200)
	wrExpectItems(mock, 51, wrItemRows(
		wrItemRow{id: wrItemA, team: "team-alpha", state: "running", objective: "Draft the brief", age: time.Minute},
		wrItemRow{id: wrItemB, team: "team-beta", state: "queued", objective: long, age: 2 * time.Minute},
		wrItemRow{id: wrItemC, team: "team-gamma", state: "degraded", objective: "Rebuild index", age: 3 * time.Minute},
	))
	wrExpectOutcomes(mock, []string{wrItemA, wrItemB, wrItemC}, wrProjectRows(
		wrProjectRow{id: wrProject1, outcome: "outcome-1", title: "Launch plan", status: "active", refs: []string{wrItemA, wrItemC}},
	))

	code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
	if code != http.StatusOK {
		t.Fatalf("status = %d body=%s", code, body)
	}
	wrMet(t, mock)
	view := wrDecode(t, body)
	if view.Truncated || view.GeneratedAt.IsZero() {
		t.Fatalf("truncated=%v generated_at=%v", view.Truncated, view.GeneratedAt)
	}
	if want := (protocol.WorkRunningSummary{Running: 1, Waiting: 1, Degraded: 1}); view.Summary != want {
		t.Fatalf("summary = %+v, want %+v", view.Summary, want)
	}
	if len(view.Outcomes) != 1 || len(view.Outcomes[0].Items) != 2 {
		t.Fatalf("outcomes = %+v", view.Outcomes)
	}
	outcome := view.Outcomes[0]
	if outcome.ProjectID != wrProject1 || outcome.OutcomeID != "outcome-1" || outcome.Title != "Launch plan" {
		t.Fatalf("outcome identity = %+v", outcome)
	}
	if outcome.OutcomeHealth != protocol.OutcomeHealthDegraded {
		t.Fatalf("aggregate health = %q, want degraded", outcome.OutcomeHealth)
	}
	for _, item := range outcome.Items {
		if item.ProjectID != wrProject1 {
			t.Fatalf("item %s project_id = %q", item.WorkItemID, item.ProjectID)
		}
	}
	if len(view.Unassigned) != 1 || view.Unassigned[0].TeamID != "team-beta" || len(view.Unassigned[0].Items) != 1 {
		t.Fatalf("unassigned = %+v", view.Unassigned)
	}
	beta := view.Unassigned[0].Items[0]
	if beta.ProjectID != "" || len([]rune(beta.Title)) != protocol.WorkRunningTitleMaxRunes {
		t.Fatalf("unassigned item project=%q title runes=%d", beta.ProjectID, len([]rune(beta.Title)))
	}
	if beta.Details.State != protocol.TeamWorkStateQueued || beta.Details.ExecutionShape != protocol.TeamExecutionShapeDelegatedWork {
		t.Fatalf("details = %+v", beta.Details)
	}
	for _, leak := range []string{"SECRET-PROMPT-BODY", "LAST-EVENT-PAYLOAD", "OUTPUT-REF-BODY", "PROOF-REF-BODY", "work_intent", "last_event", "output_refs", "proof_refs", "objective"} {
		if strings.Contains(body, leak) {
			t.Fatalf("response discloses %q: %s", leak, body)
		}
	}
}

func TestWorkRunning_LimitValidation(t *testing.T) {
	for _, raw := range []string{"0", "-1", "201", "abc", "", "1.5"} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		code, body := wrGet(t, s, "/api/v1/work/running?limit="+raw, localAdminIdentityForTest())
		if code != http.StatusBadRequest {
			t.Fatalf("limit=%q status = %d body=%s", raw, code, body)
		}
		wrMet(t, mock)
	}
	for limit, arg := range map[string]int{"1": 2, "200": 201} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		wrExpectItems(mock, arg, wrItemRows())
		if code, body := wrGet(t, s, "/api/v1/work/running?limit="+limit, localAdminIdentityForTest()); code != http.StatusOK {
			t.Fatalf("limit=%s status = %d body=%s", limit, code, body)
		}
		wrMet(t, mock)
	}
}

func TestWorkRunning_TruncatedAtLimitPlusOne(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	wrExpectItems(mock, 3, wrItemRows(
		wrItemRow{id: wrItemA, team: "t1", state: "running", objective: "a", age: time.Minute},
		wrItemRow{id: wrItemB, team: "t1", state: "running", objective: "b", age: 2 * time.Minute},
		wrItemRow{id: wrItemC, team: "t1", state: "running", objective: "c", age: 3 * time.Minute},
	))
	wrExpectOutcomes(mock, []string{wrItemA, wrItemB}, wrProjectRows())
	code, body := wrGet(t, s, "/api/v1/work/running?limit=2", localAdminIdentityForTest())
	if code != http.StatusOK {
		t.Fatalf("status = %d body=%s", code, body)
	}
	wrMet(t, mock)
	view := wrDecode(t, body)
	if !view.Truncated || view.Summary.Running != 2 || len(view.Unassigned[0].Items) != 2 {
		t.Fatalf("truncated=%v summary=%+v unassigned=%+v", view.Truncated, view.Summary, view.Unassigned)
	}
}

func TestWorkRunning_EmptyIsExplicitAndSkipsOutcomeQuery(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	wrExpectItems(mock, 51, wrItemRows())
	code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
	if code != http.StatusOK {
		t.Fatalf("status = %d body=%s", code, body)
	}
	wrMet(t, mock)
	if !strings.Contains(body, `"outcomes":[]`) || !strings.Contains(body, `"unassigned":[]`) {
		t.Fatalf("empty view must use explicit empty arrays: %s", body)
	}
}

func TestWorkRunning_ArchivedOutcomeNeverGroupsRunningWork(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	wrExpectItems(mock, 51, wrItemRows(
		wrItemRow{id: wrItemA, team: "team-alpha", state: "running", objective: "only archived owner", age: time.Minute},
		wrItemRow{id: wrItemB, team: "team-beta", state: "queued", objective: "archived and live owner", age: 2 * time.Minute},
	))
	// The SQL predicate (asserted by wrOutcomesSQL) excludes archived projects;
	// the rows below simulate filter drift so the Go guard is proven as well.
	wrExpectOutcomes(mock, []string{wrItemA, wrItemB}, wrProjectRows(
		wrProjectRow{id: wrProject2, outcome: "o-archived", title: "Archived", status: "archived", refs: []string{wrItemA, wrItemB}},
		wrProjectRow{id: wrProject1, outcome: "o-live", title: "Live", status: "active", refs: []string{wrItemB}},
	))
	code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
	if code != http.StatusOK {
		t.Fatalf("status = %d body=%s", code, body)
	}
	wrMet(t, mock)
	view := wrDecode(t, body)
	if strings.Contains(body, wrProject2) || strings.Contains(body, "o-archived") {
		t.Fatalf("archived OutcomeProject surfaced as a group: %s", body)
	}
	if len(view.Outcomes) != 1 || view.Outcomes[0].ProjectID != wrProject1 || len(view.Outcomes[0].Items) != 1 || view.Outcomes[0].Items[0].WorkItemID != wrItemB {
		t.Fatalf("live project must own item B only: %+v", view.Outcomes)
	}
	if len(view.Unassigned) != 1 || view.Unassigned[0].TeamID != "team-alpha" || view.Unassigned[0].Items[0].WorkItemID != wrItemA || view.Unassigned[0].Items[0].ProjectID != "" {
		t.Fatalf("item whose only project is archived must be ungrouped: %+v", view.Unassigned)
	}
}
