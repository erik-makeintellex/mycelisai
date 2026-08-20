package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestApplyThreadWorkerProfileSelectionPinsActiveProfileReference(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	document := retainedWorkerProfileDocument(t, "alpha")
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectQuery("SELECT .*FROM config_document_activations").
		WithArgs(
			"default", string(document.Kind), document.Metadata.ID,
			string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		).
		WillReturnRows(serverConfigRevisionRows("66666666-6666-6666-6666-666666666666", document, digest))

	planned, err := s.applyThreadWorkerProfileSelection(
		t.Context(), "", retainedWorkerProfileThread(document, "Create a team using this Worker Profile."),
		"Create a team using this Worker Profile.", "workspace-1", "", "operator-1", []protocol.PlannedToolCall{{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "qa-profile-team", "profile_ref": "default.builder",
				"profile_snapshot": map[string]any{"digest": "sha256:forged"},
			},
		}},
	)
	if err != nil {
		t.Fatalf("apply Worker Profile selection: %v", err)
	}
	if got := firstNonEmptyString(planned[0].Arguments["profile_ref"]); got != document.Metadata.ID {
		t.Fatalf("profile_ref = %q, want %q", got, document.Metadata.ID)
	}
	if _, exists := planned[0].Arguments["profile_snapshot"]; exists {
		t.Fatalf("untrusted profile_snapshot remained: %#v", planned[0].Arguments)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyThreadWorkerProfileSelectionIgnoresOrdinaryTeamRequest(t *testing.T) {
	s := &AdminServer{}
	planned := []protocol.PlannedToolCall{{
		Name: "create_team", Arguments: map[string]any{"profile_ref": "default.builder"},
	}}
	resolved, err := s.applyThreadWorkerProfileSelection(
		t.Context(), "", nil, "Create a delivery team.", "", "", "", planned,
	)
	if err != nil || firstNonEmptyString(resolved[0].Arguments["profile_ref"]) != "default.builder" {
		t.Fatalf("ordinary team request changed: (%#v, %v)", resolved, err)
	}
}
