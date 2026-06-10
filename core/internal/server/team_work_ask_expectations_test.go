package server

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func expectTeamWorkAskInsert(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState, needsOperator bool, degradation string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_work_items").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Soma",
			string(protocol.TeamExecutionShapeDelegatedWork), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.ApprovalPostureAutoAllowed), string(state), needsOperator,
			degradation, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func expectTeamWorkAskStatus(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(state), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			teamWorkAskSourceChannel, string(protocol.PayloadKindStatus), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectTeamWorkAskUpdate(mock sqlmock.Sqlmock, state protocol.TeamWorkState, needsOperator bool, degradation string) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			sqlmock.AnyArg(), string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectTeamWorkAskUpdateWithRetainedTextRefs(mock sqlmock.Sqlmock, state protocol.TeamWorkState, needsOperator bool, degradation string) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			sqlmock.AnyArg(), string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(),
			outputRefsMatch{TeamID: "qa-team", Kind: "text_reply", Label: "Team text reply"},
			stringListContainsMatch{Prefix: "team_status_event:"},
			stringListContainsMatch{Prefix: "team_interaction:"},
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectTeamWorkAskInteraction(mock sqlmock.Sqlmock, teamID, verb, payloadKind string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			teamWorkAskSourceChannel, sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
			payloadKind, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectTeamWorkAskInteractionFailure(mock sqlmock.Sqlmock, teamID, verb, payloadKind string, err error) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			teamWorkAskSourceChannel, sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
			payloadKind, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnError(err)
}

type stringListContainsMatch struct {
	Prefix string
}

func (m stringListContainsMatch) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		text, ok := value.(string)
		if !ok {
			return false
		}
		raw = []byte(text)
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err != nil {
		return false
	}
	for _, item := range items {
		if strings.HasPrefix(item, m.Prefix) {
			return true
		}
	}
	return false
}
