package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const applyGroupPath = "/api/v1/config/groups/" + applyGroupID + "/apply"

func expectGroupRow(mock sqlmock.Sqlmock, found bool) {
	rows := sqlmock.NewRows(collaborationGroupColumns())
	if found {
		now := time.Now().UTC()
		rows.AddRow(applyGroupID, "default", "ops-alpha", "coordinate reliability runs", "propose_only",
			[]byte(`["runs.read"]`), []byte(`["u1"]`), []byte(`["admin-core"]`),
			"ops-profile", "policy.ops", "active", "test-user-001", nil, "", "", now, now)
	}
	mock.ExpectQuery("FROM collaboration_groups").WillReturnRows(rows)
}

func groupDocument(goal string, extraSpec string) string {
	return "apiVersion: mycelis/v1alpha1\nkind: CollaborationGroup\nmetadata:\n  id: " + applyGroupID +
		"\n  status: active\nspec:\n  name: ops-alpha\n  goal_statement: " + goal +
		"\n  work_mode: propose_only\n  approval_policy_ref: policy.ops\n  coordinator_profile: ops-profile\n" +
		"  allowed_capabilities: [runs.read]\n  team_ids: [admin-core]\n  member_user_ids: [u1]\n" + extraSpec
}

func TestApplyGroup_PreviewThenConfirm(t *testing.T) {
	s, mock := newApplyServer(t)
	doc := groupDocument("coordinate weekly reliability reviews", "")
	expectGroupRow(mock, true)
	expectPreviewMint(mock, auditSourceConfigApplyPreview)
	env := postApply(t, s, applyGroupPath, map[string]any{"document": doc}, nil, http.StatusAccepted)
	intent := previewIntent(t, env)
	diff := env.Data["preview"].(map[string]any)["diff"].(map[string]any)
	if changed := diff["changed"].([]any); len(changed) != 1 || changed[0] != "goal_statement" {
		t.Fatalf("diff = %+v", diff)
	}

	expectGroupRow(mock, true)
	expectConsume(mock, intent, false, time.Now().Add(time.Minute))
	expectCommitTrail(mock, "config.applied", "UPDATE collaboration_groups")
	expectGroupRow(mock, true) // reload after commit
	env = postApply(t, s, applyGroupPath, map[string]any{"document": doc, "confirm_token": applyToken}, nil, http.StatusOK)
	if env.Data["applied"] != true || env.Data["kind"] != configKindGroup {
		t.Fatalf("apply result = %+v", env.Data)
	}
	mockMet(t, mock)
}

func TestApplyGroup_ValidationAndNotFound(t *testing.T) {
	s, mock := newApplyServer(t)
	cases := map[string]string{
		"unknown spec field": groupDocument("goal", "  surprise_field: x\n"),
		"bad work mode":      strings.Replace(groupDocument("goal", ""), "work_mode: propose_only", "work_mode: wild", 1),
		"wrong kind":         strings.Replace(groupDocument("goal", ""), "kind: CollaborationGroup", "kind: TeamManifest", 1),
		"metadata id":        strings.Replace(groupDocument("goal", ""), "  id: "+applyGroupID, "  id: other", 1),
	}
	for name, doc := range cases {
		expectGroupRow(mock, true)
		mock.ExpectExec("INSERT INTO log_entries").WithArgs(auditArgs(auditSourceConfigApplyPreview)...).WillReturnResult(sqlmock.NewResult(1, 1))
		env := postApply(t, s, applyGroupPath, map[string]any{"document": doc}, nil, http.StatusAccepted)
		preview := env.Data["preview"].(map[string]any)
		if env.Data["blocked"] != true || preview["valid"] != false || len(preview["validation_errors"].([]any)) == 0 {
			t.Fatalf("%s: preview = %+v", name, env.Data)
		}
	}
	expectGroupRow(mock, false)
	postApply(t, s, applyGroupPath, map[string]any{"document": groupDocument("goal", "")}, nil, http.StatusNotFound)
	postApply(t, s, "/api/v1/config/groups/not-a-uuid/apply", map[string]any{"document": groupDocument("goal", "")}, nil, http.StatusNotFound)
	mockMet(t, mock)
}
