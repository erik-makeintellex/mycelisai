package server

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

type testTeamWorkPublisher struct {
	connected  bool
	publishErr error
	flushErr   error
}

func (p testTeamWorkPublisher) IsConnected() bool            { return p.connected }
func (p testTeamWorkPublisher) Publish(string, []byte) error { return p.publishErr }
func (p testTeamWorkPublisher) Flush() error                 { return p.flushErr }

func TestPublishTeamWorkAskSurfacesDispatchFailures(t *testing.T) {
	tests := []struct {
		name      string
		publisher teamWorkPublisher
		wantState string
	}{
		{name: "offline", publisher: testTeamWorkPublisher{}, wantState: "nats_offline"},
		{name: "publish", publisher: testTeamWorkPublisher{connected: true, publishErr: errors.New("publish failed")}, wantState: "team_ask_publish_failed"},
		{name: "flush", publisher: testTeamWorkPublisher{connected: true, flushErr: errors.New("flush failed")}, wantState: "team_ask_publish_unflushed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, err := publishTeamWorkAsk(test.publisher, "swarm.team.test.internal.command", []byte("work"))
			if err == nil {
				t.Fatal("publishTeamWorkAsk error = nil, want failure")
			}
			if state != test.wantState {
				t.Fatalf("state = %q, want %q", state, test.wantState)
			}
		})
	}
}

func TestReconcileOneOverdueTeamWorkProjectsRecoverableState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	teamID := "delivery-team"
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id.*recovery_deadline_at <= NOW").
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id"}).AddRow(workID, teamID))
	mock.ExpectQuery("SELECT id::text, team_id.*FOR UPDATE").
		WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, "", "", "", "", "Build a retained package", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["project package"]`), []byte(`["runtime proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateRunning), []byte(`null`), false, "",
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
	expectRecoveryStatusEvent(mock, teamID, workID, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "team_work_recovery_deadline_exceeded")
	expectRecoveryInteraction(mock, teamID, workID, now)
	mock.ExpectCommit()

	reconciled, err := s.reconcileOneOverdueTeamWork(t.Context())
	if err != nil {
		t.Fatalf("reconcile overdue work: %v", err)
	}
	if !reconciled {
		t.Fatal("reconciled = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestReconcileOneOverdueReviewingTeamWorkUsesValidationRecovery(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	teamID := "delivery-team"
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id.*recovery_deadline_at <= NOW").
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id"}).AddRow(workID, teamID))
	mock.ExpectQuery("SELECT id::text, team_id.*FOR UPDATE").
		WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, "", "", "", "", "Build a retained package", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["project package"]`), []byte(`["runtime proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateReviewing), []byte(`null`), false, "",
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
	expectValidationRecoveryStatusEvent(mock, teamID, workID, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "runtime_validation_deadline_exceeded")
	expectRecoveryInteraction(mock, teamID, workID, now)
	mock.ExpectCommit()

	reconciled, err := s.reconcileOneOverdueTeamWork(t.Context())
	if err != nil {
		t.Fatalf("reconcile overdue reviewing work: %v", err)
	}
	if !reconciled {
		t.Fatal("reconciled = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestProjectOverdueExternalMutationRequiresVerification(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: protocol.NormalizeWorkIntent(&protocol.WorkIntent{
		SideEffect: &protocol.WorkSideEffectContract{
			EffectKind:      protocol.WorkEffectExternalMutation,
			RetrySafety:     protocol.WorkRetryUnsafe,
			SideEffectState: protocol.WorkSideEffectAccepted,
		},
	})}
	projectOverdueRecovery(&item)
	if item.DegradationState != "external_mutation_outcome_unknown" {
		t.Fatalf("degradation = %q", item.DegradationState)
	}
	if item.WorkIntent.SideEffect.SideEffectState != protocol.WorkSideEffectUnknown {
		t.Fatalf("side-effect state = %q", item.WorkIntent.SideEffect.SideEffectState)
	}
	if len(item.RecoveryOptions) != 2 {
		t.Fatalf("recovery options = %#v, want verify and archive only", item.RecoveryOptions)
	}
	for _, option := range item.RecoveryOptions {
		if strings.Contains(strings.ToLower(option), "retry with") {
			t.Fatalf("unsafe recovery offered retry: %q", option)
		}
	}
	event := overdueTeamWorkStatusEvent(item)
	if event.Headline != "External change needs verification" || event.BlockedBy[0] != "external_mutation_outcome_unknown" {
		t.Fatalf("event = %#v", event)
	}
}

func TestProjectOverdueExternalMutationRequiresVerificationBeforeRetry(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: protocol.NormalizeWorkIntent(&protocol.WorkIntent{
		SideEffect: &protocol.WorkSideEffectContract{
			EffectKind:      protocol.WorkEffectExternalMutation,
			IdempotencyKey:  "invoice-2026-08-11",
			RetrySafety:     protocol.WorkRetrySafe,
			SideEffectState: protocol.WorkSideEffectAccepted,
		},
	})}
	projectOverdueRecovery(&item)
	if len(item.RecoveryOptions) != 2 || !strings.Contains(item.RecoveryOptions[0], "verify the external system") {
		t.Fatalf("recovery options = %#v", item.RecoveryOptions)
	}
	for _, option := range item.RecoveryOptions {
		if strings.Contains(option, "invoice-2026-08-11") {
			t.Fatalf("unverified recovery exposed retry key: %#v", item.RecoveryOptions)
		}
	}
}

func TestProjectOverdueReviewingUsesValidationRecovery(t *testing.T) {
	item := protocol.TeamWorkItem{State: protocol.TeamWorkStateReviewing}
	projectOverdueRecovery(&item)
	if item.DegradationState != "runtime_validation_deadline_exceeded" {
		t.Fatalf("degradation = %q", item.DegradationState)
	}
	if len(item.RecoveryOptions) != 3 || !strings.Contains(strings.ToLower(item.RecoveryOptions[0]), "validation") {
		t.Fatalf("recovery options = %#v", item.RecoveryOptions)
	}
	event := overdueTeamWorkStatusEvent(item)
	if event.Headline != "Output validation needs recovery" || event.BlockedBy[0] != "runtime_validation_deadline_exceeded" {
		t.Fatalf("event = %#v", event)
	}
}

func TestUpdateTeamWorkItemReviewingSetsShortValidationDeadline(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	workID := "11111111-1111-1111-1111-111111111111"
	item := protocol.TeamWorkItem{WorkItemID: workID, TeamID: "app-team", State: protocol.TeamWorkStateReviewing}
	event := protocol.TeamStatusEvent{TeamID: "app-team", WorkItemID: workID, State: protocol.TeamWorkStateReviewing}

	mock.ExpectExec(`UPDATE team_work_items[\s\S]*WHEN \$2='reviewing'[\s\S]*INTERVAL '2 minutes'`).
		WithArgs(
			workID, string(protocol.TeamWorkStateReviewing), sqlmock.AnyArg(), false, "",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.updateTeamWorkItemLastEventExec(t.Context(), s.getDB(), &item, event); err != nil {
		t.Fatalf("update reviewing item: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestReconcileOneOverdueTeamWorkReturnsCleanlyWhenQueueEmpty(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id.*recovery_deadline_at <= NOW").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectCommit()

	reconciled, err := s.reconcileOneOverdueTeamWork(t.Context())
	if err != nil {
		t.Fatalf("reconcile empty queue: %v", err)
	}
	if reconciled {
		t.Fatal("reconciled = true, want false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMarkRunDegradedWhenSettledWaitsForActiveSibling(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	runID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectBegin()
	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(`COUNT\(\*\) FILTER.*FROM team_work_items`).
		WithArgs(runID).
		WillReturnRows(sqlmock.NewRows([]string{"unresolved", "degraded"}).AddRow(1, 1))
	if err := s.markRunDegradedWhenSettledTx(t.Context(), tx, protocol.TeamWorkItem{RunID: runID}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMarkRunDegradedWhenSettledTerminatesMultiItemRun(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	runID := "22222222-2222-2222-2222-222222222222"
	item := protocol.TeamWorkItem{RunID: runID, IntentProofID: "33333333-3333-3333-3333-333333333333", TeamID: "delivery-team"}
	mock.ExpectBegin()
	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(`COUNT\(\*\) FILTER.*FROM team_work_items`).
		WithArgs(runID).
		WillReturnRows(sqlmock.NewRows([]string{"unresolved", "degraded"}).AddRow(0, 1))
	mock.ExpectExec(`UPDATE mission_runs.*status=\$1`).
		WithArgs(runs.StatusDegraded, runID, runs.StatusCompleted, runs.StatusFailed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), runID, string(protocol.EventMissionDegraded), string(protocol.SeverityWarn), item.TeamID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.markRunDegradedWhenSettledTx(t.Context(), tx, item); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectRecoveryStatusEvent(mock sqlmock.Sqlmock, teamID, workID string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.TeamWorkStateDegraded), "Team work needs recovery", sqlmock.AnyArg(), "operator_attention",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindSystem),
			"team-work.recovery-reconciler", string(protocol.PayloadKindError), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectValidationRecoveryStatusEvent(mock sqlmock.Sqlmock, teamID, workID string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.TeamWorkStateDegraded), "Output validation needs recovery", sqlmock.AnyArg(), "operator_attention",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindSystem),
			"team-work.recovery-reconciler", string(protocol.PayloadKindError), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectRecoveryInteraction(mock sqlmock.Sqlmock, teamID, workID string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.SourceKindSystem), "team-work.recovery-reconciler", "Soma", "degraded", sqlmock.AnyArg(),
			string(protocol.PayloadKindError), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}
