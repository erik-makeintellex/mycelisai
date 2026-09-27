package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestWorkRunning_ScopeDeniedDefaultDeny(t *testing.T) {
	cases := map[string]*RequestIdentity{
		"standard web user":          {UserID: "u-std", Role: "user", PrincipalType: "web_user", Scopes: []string{"soma:work", "runs:read", "outputs:read"}},
		"admin without groups:read":  {UserID: "u-adm", Role: "admin", Scopes: []string{"groups:write", "runs:read"}},
		"non-admin with groups:read": {UserID: "u-op", Role: "operator", Scopes: []string{"groups:read"}},
		"admin with empty scopes":    {UserID: "u-empty", Role: "admin"},
	}
	for name, identity := range cases {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			code, body := wrGet(t, s, "/api/v1/work/running", identity)
			if code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s", code, body)
			}
			if denialLeaksData(body) || strings.Contains(body, "work_item_id") {
				t.Fatalf("denial leaked data: %s", body)
			}
			wrMet(t, mock) // denied before any read
		})
	}
}

func TestWorkRunning_AnonymousUnauthorized(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "GET /api/v1/work/running", s.HandleWorkRunning)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/work/running", "")
	assertStatus(t, rr, http.StatusUnauthorized)
	if denialLeaksData(rr.Body.String()) {
		t.Fatalf("anonymous denial leaked data: %s", rr.Body.String())
	}
	wrMet(t, mock)
}

func TestWorkRunning_AdminWithGroupsReadAllowed(t *testing.T) {
	for _, scopes := range [][]string{{"groups:read"}, {"groups:*"}} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		wrExpectItems(mock, 51, wrItemRows())
		identity := &RequestIdentity{UserID: "u-adm", Role: "admin", Scopes: scopes}
		if code, body := wrGet(t, s, "/api/v1/work/running", identity); code != http.StatusOK {
			t.Fatalf("scopes=%v status = %d body=%s", scopes, code, body)
		}
		wrMet(t, mock)
	}
}

func TestWorkRunning_TruthfulStatesNeverShowTerminalWork(t *testing.T) {
	want := map[string]protocol.OutcomeHealthState{
		"new": protocol.OutcomeHealthWaiting, "briefed": protocol.OutcomeHealthWaiting,
		"queued": protocol.OutcomeHealthWaiting, "paused": protocol.OutcomeHealthWaiting,
		"running": protocol.OutcomeHealthRunning, "reviewing": protocol.OutcomeHealthRunning,
		"needs_operator": protocol.OutcomeHealthBlocked, "degraded": protocol.OutcomeHealthDegraded,
	}
	rows := []wrItemRow{}
	visible := []string{}
	idFor := map[string]string{}
	n := 0
	add := func(state string, needs bool, show bool) string {
		n++
		id := fmt.Sprintf("cccccccc-0000-0000-0000-%012d", n)
		rows = append(rows, wrItemRow{id: id, team: "t", state: state, objective: state, needsOperator: needs, age: time.Duration(n) * time.Second})
		if show {
			visible = append(visible, id)
		}
		return id
	}
	for state := range want {
		idFor[state] = add(state, false, true)
	}
	flagged := add("running", true, true)
	// Terminal rows the SQL filter should already drop; the Go guard must too.
	add("archived", false, false)
	add("output_ready", false, false)
	add("output_ready", true, false)

	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	wrExpectItems(mock, 51, wrItemRows(rows...))
	wrExpectOutcomes(mock, visible, wrProjectRows())
	code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
	if code != http.StatusOK {
		t.Fatalf("status = %d body=%s", code, body)
	}
	wrMet(t, mock)
	got := map[string]protocol.WorkRunningItem{}
	for _, group := range wrDecode(t, body).Unassigned {
		for _, item := range group.Items {
			got[item.WorkItemID] = item
			if item.Details.State == protocol.TeamWorkStateArchived || item.Details.State == protocol.TeamWorkStateOutputReady ||
				item.OutcomeHealth == protocol.OutcomeHealthCompleted || item.OutcomeHealth == protocol.OutcomeHealthArchived {
				t.Fatalf("terminal work shown as current: %+v", item)
			}
		}
	}
	if len(got) != len(visible) {
		t.Fatalf("visible items = %d, want %d", len(got), len(visible))
	}
	for state, health := range want {
		item := got[idFor[state]]
		if item.OutcomeHealth != health || item.OutcomeHealth != protocol.OutcomeHealthForTeamWork(protocol.TeamWorkItem{State: protocol.TeamWorkState(state)}) {
			t.Fatalf("state %s health = %q, want %q", state, item.OutcomeHealth, health)
		}
	}
	if got[flagged].OutcomeHealth != protocol.OutcomeHealthBlocked || !got[flagged].NeedsOperator {
		t.Fatalf("needs_operator flag must read Blocked: %+v", got[flagged])
	}
}

func TestWorkRunning_DuplicateRefsResolveDeterministically(t *testing.T) {
	run := func(s *AdminServer, mock sqlmock.Sqlmock) protocol.WorkRunningView {
		wrExpectItems(mock, 51, wrItemRows(wrItemRow{id: wrItemA, team: "t1", state: "running", objective: "a"}))
		// Query order is updated_at DESC, id ASC: project 2 is the most recent.
		wrExpectOutcomes(mock, []string{wrItemA}, wrProjectRows(
			wrProjectRow{id: wrProject2, outcome: "o-2", title: "Newer", status: "output_ready", refs: []string{wrItemA, wrItemA}},
			wrProjectRow{id: wrProject1, outcome: "o-1", title: "Older", status: "active", refs: []string{wrItemA}},
		))
		code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
		if code != http.StatusOK {
			t.Fatalf("status = %d body=%s", code, body)
		}
		return wrDecode(t, body)
	}
	// Restart-safe: a fresh server (no carried state) over the same durable rows
	// yields the same projection. Replay-safe: a repeated read on one server is
	// identical and issues only the two SELECTs (sqlmock rejects any write).
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	first, replay := run(s, mock), run(s, mock)
	wrMet(t, mock)
	dbOpt2, mock2 := withDB(t)
	restarted := run(newTestServer(dbOpt2), mock2)
	wrMet(t, mock2)
	for _, view := range []protocol.WorkRunningView{first, replay, restarted} {
		if len(view.Outcomes) != 1 || view.Outcomes[0].ProjectID != wrProject2 || len(view.Outcomes[0].Items) != 1 || len(view.Unassigned) != 0 {
			t.Fatalf("item must appear once under the most recent project: %+v", view)
		}
		// Project output_ready (Completed) must not outrank a running item.
		if view.Outcomes[0].OutcomeHealth != protocol.OutcomeHealthRunning || view.Summary.Running != 1 {
			t.Fatalf("aggregate = %q summary=%+v", view.Outcomes[0].OutcomeHealth, view.Summary)
		}
		view.GeneratedAt = first.GeneratedAt
		if fmt.Sprintf("%+v", view) != fmt.Sprintf("%+v", first) {
			t.Fatalf("projection drifted across replay/restart:\n%+v\n%+v", view, first)
		}
	}
}

func TestWorkRunning_PerSourceDegradationFailsClosed(t *testing.T) {
	items := func() *sqlmock.Rows {
		return wrItemRows(wrItemRow{id: wrItemA, team: "t1", state: "running", objective: "LEAKY-TITLE"})
	}
	cases := []struct {
		name, source string
		setup        func(sqlmock.Sqlmock)
	}{
		{"team work query error", "team_work_items", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(regexp.QuoteMeta(wrItemsSQL)).WillReturnError(errors.New("connection reset"))
		}},
		{"team work row error mid-stream", "team_work_items", func(m sqlmock.Sqlmock) {
			wrExpectItems(m, 51, wrItemRows(
				wrItemRow{id: wrItemA, team: "t1", state: "running", objective: "LEAKY-TITLE"},
				wrItemRow{id: wrItemB, team: "t1", state: "running", objective: "b"},
			).RowError(1, errors.New("stream broken")))
		}},
		{"outcome link query error", "outcome_projects", func(m sqlmock.Sqlmock) {
			wrExpectItems(m, 51, items())
			m.ExpectQuery(regexp.QuoteMeta(wrOutcomesSQL)).WillReturnError(errors.New("relation busy"))
		}},
		{"outcome link scan error", "outcome_projects", func(m sqlmock.Sqlmock) {
			wrExpectItems(m, 51, items())
			m.ExpectQuery(regexp.QuoteMeta(wrOutcomesSQL)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("x"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			tc.setup(mock)
			code, body := wrGet(t, s, "/api/v1/work/running", localAdminIdentityForTest())
			wrAssertUnavailable(t, code, body, tc.source)
			wrMet(t, mock)
		})
	}
	t.Run("no database", func(t *testing.T) {
		code, body := wrGet(t, newTestServer(), "/api/v1/work/running", localAdminIdentityForTest())
		wrAssertUnavailable(t, code, body, "database")
	})
}

func wrAssertUnavailable(t *testing.T, code int, body, source string) {
	t.Helper()
	if code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s", code, body)
	}
	if !strings.Contains(body, `"source":"`+source+`"`) || !strings.Contains(body, `"recovery_hint"`) {
		t.Fatalf("503 must name source %q with a recovery hint: %s", source, body)
	}
	for _, leak := range []string{"LEAKY-TITLE", "work_item_id", `"items"`, `"summary"`, "connection reset", "relation busy"} {
		if strings.Contains(body, leak) {
			t.Fatalf("503 body leaked %q (partial list or raw error): %s", leak, body)
		}
	}
}
