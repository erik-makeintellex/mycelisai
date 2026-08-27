package server

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func TestResolveOutputContinuationOwnershipBindsDurableLineage(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "moonlit", "generated", "package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workIntent := []byte(`{"kind":"project","output_contract":{"shape":"app_package","acceptance_criteria":["Win works"],"semantic_validation_required":true}}`)
	outputs := []byte(`[{"output_id":"package-v1","team_id":"moonlit","work_item_id":"22222222-2222-4222-8222-222222222222","run_id":"11111111-1111-4111-8111-111111111111","kind":"project_package","label":"Moonlit Keep First Playable","storage_ref":"groups/moonlit/generated/package","entrypoint":"index.html","proof_ref":"proof-v1"}]`)
	mock.ExpectQuery("(?s)FROM team_work_items.*jsonb_array_elements").
		WithArgs("groups/moonlit/generated/package", "proof-v1").
		WillReturnRows(teamWorkItemRows().AddRow(
			"22222222-2222-4222-8222-222222222222", "moonlit", "11111111-1111-4111-8111-111111111111", "", "", "proof-v1",
			"Build game", []byte(`[]`), "Soma", string(protocol.TeamExecutionShapeDelegatedWork), "team_async", workIntent,
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), "approved", string(protocol.TeamWorkStateOutputReady), []byte(`null`), false, "",
			[]byte(`[]`), outputs, []byte(`["proof-v1"]`), []byte(`[]`), now, now, "v1",
		))
	ctx := &chatContinuationContext{Kind: "output", Intent: "update", Title: "Moonlit Keep First Playable", Reference: "groups/moonlit/generated/package", Proof: "proof-v1"}
	resolved, err := s.resolveOutputContinuationOwnership(t.Context(), ctx)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resolved.OwnershipValidated || resolved.TeamID != "moonlit" || resolved.OutputID != "package-v1" {
		t.Fatalf("resolved = %#v", resolved)
	}
	if resolved.SourceDigest == "" || resolved.RevisionTarget != "groups/moonlit/generated/package-v2" {
		t.Fatalf("lineage = digest %q target %q", resolved.SourceDigest, resolved.RevisionTarget)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTrustedOutputRevisionPlanKeepsOwnerAndForksTarget(t *testing.T) {
	ctx := trustedRevisionContext()
	calls, ok := buildTrustedOutputRevisionPlan(ctx, "Add a moon-shard counter and collectible.")
	if !ok || len(calls) != 2 {
		t.Fatalf("calls = %#v, want retained handoff and delegation", calls)
	}
	if calls[0].Name != "write_file" || calls[1].Name != "delegate_task" {
		t.Fatalf("tools = %s, %s", calls[0].Name, calls[1].Name)
	}
	if got := firstStringArgument(calls[1].Arguments, "team_id"); got != ctx.TeamID {
		t.Fatalf("team = %q, want owning team %q", got, ctx.TeamID)
	}
	ask := mapArgument(calls[1].Arguments["ask"])
	result := mapArgument(mapArgument(ask["context"])["result_contract"])
	if result["package_folder"] != ctx.RevisionTarget || result["package_folder"] == ctx.Reference {
		t.Fatalf("revision folder = %#v, source = %q", result["package_folder"], ctx.Reference)
	}
	lineage := mapArgument(result["source_lineage"])
	if lineage["work_item_id"] != ctx.WorkItemID || lineage["source_digest"] != ctx.SourceDigest {
		t.Fatalf("lineage = %#v", lineage)
	}
	criteria := confirmedActionStringSlice(result["acceptance_criteria"])
	for _, want := range ctx.SourceWorkIntent.OutputContract.AcceptanceCriteria {
		if !slices.Contains(criteria, want) {
			t.Fatalf("criteria lost inherited requirement %q: %#v", want, criteria)
		}
	}
	if !slices.Contains(criteria, revisionAcceptanceCriterion) {
		t.Fatalf("criteria missing revision validation contract: %#v", criteria)
	}
}

func TestTrustedOutputRevisionPlanRequiresValidatedOwnership(t *testing.T) {
	ctx := trustedRevisionContext()
	ctx.OwnershipValidated = false
	if calls, ok := buildTrustedOutputRevisionPlan(ctx, "Add a moon shard."); ok || len(calls) != 0 {
		t.Fatalf("unvalidated ownership planned mutation: %#v", calls)
	}
}

func TestMatchingContinuationOutputValidatesIdentity(t *testing.T) {
	ctx := trustedRevisionContext()
	item := protocol.TeamWorkItem{
		TeamID: ctx.TeamID, RunID: ctx.RunID,
		OutputRefs: []protocol.TeamOutputRef{{
			OutputID: ctx.OutputID, StorageRef: ctx.Reference, Entrypoint: "index.html", ProofRef: ctx.Proof,
		}},
	}
	if _, ok := matchingContinuationOutput(item, ctx); !ok {
		t.Fatal("matching owned output was rejected")
	}
	ctx.RunID = "99999999-9999-4999-8999-999999999999"
	if _, ok := matchingContinuationOutput(item, ctx); ok {
		t.Fatal("mismatched run identity must be rejected")
	}
}

func TestRevisionWorkIntentInheritsSemanticContract(t *testing.T) {
	ctx := trustedRevisionContext()
	current := &protocol.WorkIntent{Kind: "project", OutputContract: &protocol.WorkOutputContract{Shape: "document"}}
	intent := inheritRevisionWorkIntent(current, ctx, "Add a moon-shard counter.")
	if intent.OutputContract.PrimaryDeliverable != ctx.RevisionTarget || intent.OutputContract.Shape != "app_package" {
		t.Fatalf("revision output contract = %#v", intent.OutputContract)
	}
	if !intent.OutputContract.SemanticValidationRequired || len(intent.OutputContract.AcceptanceCriteria) < 3 {
		t.Fatalf("semantic criteria were weakened: %#v", intent.OutputContract.AcceptanceCriteria)
	}
}

func trustedRevisionContext() *chatContinuationContext {
	return &chatContinuationContext{
		Kind: "output", Intent: "update", Title: "Moonlit Keep First Playable",
		Reference: "groups/moonlit/generated/package", Proof: "proof-v1",
		TeamID: "moonlit", RunID: "11111111-1111-4111-8111-111111111111",
		WorkItemID: "22222222-2222-4222-8222-222222222222", OutputID: "package-v1",
		SourceDigest: "sha256:abcdef0123456789", SourceVersion: "v1",
		RevisionTarget: "groups/moonlit/generated/package-v2", OwnershipValidated: true,
		SourceWorkIntent: &protocol.WorkIntent{Kind: "project", OutputContract: &protocol.WorkOutputContract{
			Shape: "app_package", Retention: "user_deliverable", SemanticValidationRequired: true,
			PrimaryDeliverable: "Moonlit Keep First Playable",
			AcceptanceCriteria: []string{"Movement and attack work", "Win and restart are testable"},
		}},
	}
}
