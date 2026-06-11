package server

import (
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func expectConfirmedCreateTeamLifecycle(mock sqlmock.Sqlmock, teamID string, now time.Time) {
	mock.ExpectBegin()
	expectTeamWorkItemInsert(mock, teamID, protocol.TeamExecutionShapeCreateTeam, protocol.TeamWorkStateNew, now)
	expectTeamStatusEventInsert(mock, teamID, protocol.TeamWorkStateNew, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateNew, sqlmock.AnyArg())
	expectTeamInteractionInsert(mock, teamID, "create_team", now)
	mock.ExpectCommit()
}

func expectConfirmedDeliverableLifecycle(mock sqlmock.Sqlmock, teamID, path string, now time.Time) {
	mock.ExpectBegin()
	expectTeamWorkItemInsert(mock, teamID, protocol.TeamExecutionShapeDeliverable, protocol.TeamWorkStateOutputReady, now)
	expectTeamStatusEventInsert(mock, teamID, protocol.TeamWorkStateQueued, now)
	expectTeamStatusEventInsert(mock, teamID, protocol.TeamWorkStateRunning, now)
	expectTeamStatusEventInsert(mock, teamID, protocol.TeamWorkStateOutputReady, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateOutputReady, jsonContainsArg(path))
	expectTeamInteractionInsert(mock, teamID, "output_ready", now)
	mock.ExpectCommit()
}
