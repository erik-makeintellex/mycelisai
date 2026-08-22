package swarm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleCreateTeamActivatedProfileHonorsConfirmationDeadline(t *testing.T) {
	_, nc := startTestNATS(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindWorkerProfile,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: "temporary-worker", Name: "Temporary worker", Version: "1.0.0", OwnerID: "owner-1", Enabled: true,
			Scope:  protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-7"},
			Source: protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceSoma, Ref: "session-1"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel: protocol.ConfigDocumentRiskLow, ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: []byte(`{"role":"worker","system_prompt":"Complete the confirmed temporary assignment.","capability_refs":["artifact.read"],"outputs":["result"]}`),
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		t.Fatal(err)
	}
	expectNoActiveProfile(mock, document.Metadata.ID, protocol.ConfigDocumentScopeOperator, "owner-user")
	mock.ExpectQuery("FROM config_document_activations").
		WithArgs("default", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
		WillReturnRows(activeProfileRows("22222222-2222-2222-2222-222222222222", document, digest))

	registry := NewInternalToolRegistry(InternalToolDeps{NC: nc, Catalogue: catalogue.NewService(db)})
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromManifests(nil), nil, nil, nil, registry)
	store := &contextBlockingDurableTeamStore{saveStarted: make(chan struct{})}
	soma.SetDurableTeamStore(store)
	t.Cleanup(soma.Shutdown)

	invocationCtx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		OperatorID: "owner-user", WorkspaceID: document.Metadata.Scope.Ref,
	})
	ctx, cancel := context.WithTimeout(invocationCtx, 200*time.Millisecond)
	defer cancel()

	type createResult struct {
		output string
		err    error
	}
	resultCh := make(chan createResult, 1)
	go func() {
		output, createErr := registry.handleCreateTeam(ctx, map[string]any{
			"team_id": "temporary-worker-team", "profile_ref": document.Metadata.ID,
			"profile_scope": map[string]any{"workspace_ref": document.Metadata.Scope.Ref},
		})
		resultCh <- createResult{output: output, err: createErr}
	}()

	select {
	case <-store.saveStarted:
	case <-time.After(time.Second):
		soma.cancel()
		<-resultCh
		t.Fatal("create_team did not reach durable persistence")
	}
	teamsCh := make(chan []*TeamManifest, 1)
	go func() { teamsCh <- soma.ListTeams() }()
	select {
	case teams := <-teamsCh:
		if len(teams) != 0 {
			t.Fatalf("unacknowledged runtime teams: %#v", teams)
		}
	case <-time.After(100 * time.Millisecond):
		soma.cancel()
		<-resultCh
		t.Fatal("blocked durable persistence held the team registry lock")
	}

	var result createResult
	select {
	case result = <-resultCh:
	case <-time.After(2 * time.Second):
		soma.cancel()
		<-resultCh
		t.Fatal("create_team outlived the confirmation deadline while durable persistence was blocked")
	}
	if result.output != "" || !errors.Is(result.err, context.DeadlineExceeded) {
		t.Fatalf("create_team = (%q, %v), want confirmation deadline error", result.output, result.err)
	}
	if teams := soma.ListTeams(); len(teams) != 0 {
		t.Fatalf("failed confirmation registered runtime teams: %#v", teams)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type contextBlockingDurableTeamStore struct {
	saveStarted chan struct{}
}

func (s *contextBlockingDurableTeamStore) LoadRuntimeTeams(context.Context) ([]*TeamManifest, error) {
	return nil, nil
}

func (s *contextBlockingDurableTeamStore) SaveRuntimeTeam(ctx context.Context, _ *TeamManifest) error {
	close(s.saveStarted)
	<-ctx.Done()
	return ctx.Err()
}

func (s *contextBlockingDurableTeamStore) DeleteRuntimeTeam(context.Context, string) error {
	return nil
}
