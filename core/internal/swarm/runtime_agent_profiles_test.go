package swarm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/pkg/protocol"
)

var profileColumns = []string{
	"id", "name", "role", "system_prompt", "model", "tools", "inputs", "outputs",
	"verification_strategy", "verification_rubric", "validation_command", "profile_key",
	"description", "source", "locked", "capability_refs", "context_bindings", "usage_policy",
	"created_at", "updated_at",
}

func TestHydrateCreateTeamProfilesDoesNotResolveUnselectedCatalogueEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	args := map[string]any{
		"team_id": "minimal-team",
		"agents": []map[string]any{{
			"id": "coordinator", "role": "coordinator",
		}},
	}
	if err := registry.hydrateCreateTeamProfiles(context.Background(), args); err != nil {
		t.Fatalf("hydrate profiles without selection: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unselected catalogue entry was resolved: %v", err)
	}
}

func TestHydrateCreateTeamProfiles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	expectNoActiveProfile(mock, "default.researcher", protocol.ConfigDocumentScopeBuiltIn, "")
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WithArgs("default.researcher").
		WillReturnRows(sqlmock.NewRows(profileColumns).AddRow(
			"11111111-1111-1111-1111-111111111111", "Research Specialist", "researcher",
			"Research trusted sources.", "", []byte(`["web_search"]`), []byte(`[]`), []byte(`["source_summary"]`),
			"semantic", []byte(`["Cite evidence"]`), nil, "default.researcher",
			"Finds evidence", "built_in", true, []byte(`["web_search","read_text_file"]`),
			[]byte(`[{"kind":"public_web","access":"read"},{"kind":"approved_mount","access":"read"}]`),
			[]byte(`{"selection":"soma_or_manual","scope":"workspace"}`), now, now,
		))

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	member := map[string]any{
		"profile_ref": "default.researcher",
		"role":        "evidence-reviewer",
	}
	args := map[string]any{"agents": []map[string]any{member}}
	if err := registry.hydrateCreateTeamProfiles(context.Background(), args); err != nil {
		t.Fatalf("hydrate profiles: %v", err)
	}

	if member["role"] != "evidence-reviewer" {
		t.Fatalf("explicit role was overwritten: %#v", member["role"])
	}
	tools, ok := member["tools"].([]string)
	if !ok || len(tools) != 2 || tools[0] != "web_search" {
		t.Fatalf("capabilities not hydrated: %#v", member["tools"])
	}
	if member["context_bindings"] == nil || member["usage_policy"] == nil {
		t.Fatalf("context and usage policy must be hydrated: %#v", member)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHydrateCreateTeamProfilesRejectsUnknownProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectNoActiveProfile(mock, "missing.profile", protocol.ConfigDocumentScopeBuiltIn, "")
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WithArgs("missing.profile").
		WillReturnRows(sqlmock.NewRows(profileColumns))

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	err = registry.hydrateCreateTeamProfiles(context.Background(), map[string]any{"profile_ref": "missing.profile"})
	if err == nil {
		t.Fatal("expected unknown profile to block team creation")
	}
}

func TestHydrateCreateTeamProfilesRejectsProfileWithoutCatalogue(t *testing.T) {
	registry := &InternalToolRegistry{}
	err := registry.hydrateCreateTeamProfiles(context.Background(), map[string]any{"profile_ref": "custom.reviewer"})
	if err == nil {
		t.Fatal("expected profile reference without catalogue to fail closed")
	}
}

func TestHydrateCreateTeamProfilesPinsActiveWorkspaceRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindWorkerProfile,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: "delivery-reviewer", Name: "Delivery reviewer", Version: "2.0.0", OwnerID: "owner-1", Enabled: true,
			Scope:  protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-7"},
			Source: protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceSoma, Ref: "session-1"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel: protocol.ConfigDocumentRiskLow, ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: json.RawMessage(`{"role":"reviewer","system_prompt":"Review the retained delivery.","capability_refs":["artifact.review"],"outputs":["review_report"],"verification_strategy":"semantic","verification_rubric":["Findings cite evidence"]}`),
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		t.Fatal(err)
	}
	expectNoActiveProfile(mock, document.Metadata.ID, protocol.ConfigDocumentScopeOperator, "owner-user")
	mock.ExpectQuery("FROM config_document_activations").
		WithArgs("default", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
		WillReturnRows(activeProfileRows("22222222-2222-2222-2222-222222222222", document, digest))

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		OperatorID: "owner-user", WorkspaceID: document.Metadata.Scope.Ref,
	})
	args := map[string]any{
		"team_id": "review-team", "profile_ref": document.Metadata.ID,
		"profile_scope": map[string]any{"workspace_ref": document.Metadata.Scope.Ref},
	}
	if err := registry.hydrateCreateTeamProfiles(ctx, args); err != nil {
		t.Fatalf("hydrate profiles: %v", err)
	}
	manifest := buildRuntimeTeamManifest(args)
	if manifest == nil || len(manifest.Members) != 1 {
		t.Fatalf("manifest = %#v", manifest)
	}
	member := manifest.Members[0]
	if member.Profile == nil || member.Profile.RecordID != "22222222-2222-2222-2222-222222222222" || member.Profile.Digest != digest {
		t.Fatalf("profile lineage = %#v", member.Profile)
	}
	if member.Profile.Scope.Kind != protocol.ConfigDocumentScopeWorkspace || member.Role != "reviewer" {
		t.Fatalf("member = %#v", member)
	}
	if member.Verification == nil || member.Verification.Strategy != protocol.VerifySemantic {
		t.Fatalf("verification = %#v", member.Verification)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHydrateCreateTeamProfilesRejectsScopeOutsideConfirmedBoundary(t *testing.T) {
	registry := &InternalToolRegistry{}
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		OperatorID: "operator-1", WorkspaceID: "workspace-1", OrganizationID: "org-1",
	})
	err := registry.hydrateCreateTeamProfiles(ctx, map[string]any{
		"profile_scope": map[string]any{"workspace_ref": "workspace-2"},
	})
	if err == nil {
		t.Fatal("expected cross-workspace profile scope to fail closed")
	}
}

func TestHydrateCreateTeamProfilesStripsUntrustedSnapshotAndPreservesEmptyOverride(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectNoActiveProfile(mock, "default.researcher", protocol.ConfigDocumentScopeBuiltIn, "")
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WithArgs("default.researcher").
		WillReturnRows(sqlmock.NewRows(profileColumns).AddRow(
			"11111111-1111-1111-1111-111111111111", "Research Specialist", "researcher",
			"Research trusted sources.", "", []byte(`["web_search"]`), []byte(`[]`), []byte(`["source_summary"]`),
			"semantic", []byte(`["Cite evidence"]`), nil, "default.researcher",
			"Finds evidence", "built_in", true, []byte(`["web_search"]`), []byte(`[]`),
			[]byte(`{"selection":"soma_or_manual","scope":"workspace"}`), time.Now(), time.Now(),
		))

	args := map[string]any{
		"team_id": "research-team", "profile_ref": "default.researcher",
		"tools":            []any{},
		"profile_snapshot": map[string]any{"id": "forged", "digest": "sha256:forged"},
	}
	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	if err := registry.hydrateCreateTeamProfiles(context.Background(), args); err != nil {
		t.Fatalf("hydrate profiles: %v", err)
	}
	if _, exists := args["profile_snapshot"]; exists {
		t.Fatalf("untrusted profile snapshot survived hydration: %#v", args["profile_snapshot"])
	}
	if tools, ok := args["tools"].([]any); !ok || len(tools) != 0 {
		t.Fatalf("explicit empty tools override was replaced: %#v", args["tools"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeManifestKeepsResolvedProfileSnapshot(t *testing.T) {
	args := map[string]any{
		"team_id":     "review-team",
		"profile_ref": "delivery-reviewer",
		"profile_snapshot": map[string]any{
			"id": "delivery-reviewer", "version": "2.0.0", "digest": "sha256:original",
			"record_id": "22222222-2222-2222-2222-222222222222", "tenant_id": "default",
			"scope": map[string]any{"kind": "workspace", "ref": "workspace-7"},
		},
	}
	manifest := buildRuntimeTeamManifest(args)
	if manifest == nil || manifest.Members[0].Profile == nil {
		t.Fatalf("manifest = %#v", manifest)
	}
	snapshot := args["profile_snapshot"].(map[string]any)
	snapshot["version"] = "3.0.0"
	snapshot["digest"] = "sha256:replacement"
	if manifest.Members[0].Profile.Version != "2.0.0" || manifest.Members[0].Profile.Digest != "sha256:original" {
		t.Fatalf("manifest profile changed after source activation data changed: %#v", manifest.Members[0].Profile)
	}
}

func expectNoActiveProfile(mock sqlmock.Sqlmock, profileID string, kind protocol.ConfigDocumentScopeKind, ref string) {
	mock.ExpectQuery("FROM config_document_activations").
		WithArgs("default", string(protocol.ConfigDocumentKindWorkerProfile), profileID, string(kind), ref).
		WillReturnRows(sqlmock.NewRows(activeProfileColumns))
}

var activeProfileColumns = []string{
	"record_id", "tenant_id", "document_id", "api_version", "kind", "name", "version", "owner_id",
	"scope_kind", "scope_ref", "enabled", "source_kind", "source_ref", "secret_refs", "governance", "spec",
	"digest", "validation_state", "created_by", "created_at",
}

func activeProfileRows(recordID string, document protocol.ConfigDocument, digest string) *sqlmock.Rows {
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	return sqlmock.NewRows(activeProfileColumns).AddRow(
		recordID, "default", document.Metadata.ID, document.APIVersion, string(document.Kind), document.Metadata.Name,
		document.Metadata.Version, document.Metadata.OwnerID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		document.Metadata.Enabled, string(document.Metadata.Source.Kind), document.Metadata.Source.Ref, secretRefs, governance,
		document.Spec, digest, "valid", "owner-user", time.Now().UTC(),
	)
}
