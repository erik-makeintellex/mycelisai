package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/catalogue"
)

func failureTruthDB(t *testing.T) (sqlmock.Sqlmock, *InternalToolRegistry) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return mock, NewInternalToolRegistry(InternalToolDeps{DB: db})
}

func assertToolFailure(t *testing.T, name, out string, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s returned success %q, want failure", name, out)
	}
	if out != "" {
		t.Fatalf("%s leaked output %q alongside failure", name, out)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("%s err = %v, want %q", name, err, want)
	}
}

func TestInternalToolsUnavailableDependenciesAreFailures(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	ctx := context.Background()
	cases := []struct {
		name string
		call func() (string, error)
		want string
	}{
		{"search_memory", func() (string, error) { return r.handleSearchMemory(ctx, map[string]any{"query": "x"}) }, "search_memory unavailable"},
		{"list_teams", func() (string, error) { return r.handleListTeams(ctx, nil) }, "list_teams unavailable"},
		{"list_missions", func() (string, error) { return r.handleListMissions(ctx, nil) }, "list_missions unavailable"},
		{"list_catalogue", func() (string, error) { return r.handleListCatalogue(ctx, nil) }, "list_catalogue unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.call()
			assertToolFailure(t, tc.name, out, err, tc.want)
		})
	}
}

func TestInternalToolsStorageErrorsAreFailures(t *testing.T) {
	ctx := context.Background()
	storeErr := errors.New("connection reset")

	t.Run("store_artifact", func(t *testing.T) {
		mock, r := failureTruthDB(t)
		mock.ExpectQuery("INSERT INTO artifacts").WillReturnError(storeErr)
		out, err := r.handleStoreArtifact(ctx, map[string]any{"type": "document", "title": "Notes", "content": "body"})
		assertToolFailure(t, "store_artifact", out, err, "connection reset")
	})
	t.Run("remember", func(t *testing.T) {
		mock, r := failureTruthDB(t)
		mock.ExpectExec("INSERT INTO agent_memories").WillReturnError(storeErr)
		out, err := r.handleRemember(ctx, map[string]any{"category": "fact", "content": "x"})
		assertToolFailure(t, "remember", out, err, "connection reset")
	})
	t.Run("list_missions", func(t *testing.T) {
		mock, r := failureTruthDB(t)
		mock.ExpectQuery("FROM missions").WillReturnError(storeErr)
		out, err := r.handleListMissions(ctx, nil)
		assertToolFailure(t, "list_missions", out, err, "connection reset")
	})
	t.Run("list_catalogue", func(t *testing.T) {
		mock, r := failureTruthDB(t)
		r.catalogue = catalogue.NewService(r.db)
		mock.ExpectQuery("SELECT").WillReturnError(storeErr)
		out, err := r.handleListCatalogue(ctx, nil)
		assertToolFailure(t, "list_catalogue", out, err, "connection reset")
	})
	t.Run("generate_image", func(t *testing.T) {
		mock, r := failureTruthDB(t)
		mock.ExpectQuery("INSERT INTO artifacts").WillReturnError(storeErr)
		var resp generatedImageResponse
		if err := json.Unmarshal([]byte(`{"data":[{"b64_json":"aW1n"}]}`), &resp); err != nil {
			t.Fatal(err)
		}
		out, err := r.finishGeneratedImage(ctx, "logo", "512x512", resp)
		assertToolFailure(t, "generate_image", out, err, "could not store the generated image")
	})
}

func TestInternalToolsSuccessPathsUnchanged(t *testing.T) {
	ctx := context.Background()
	mock, r := failureTruthDB(t)
	mock.ExpectQuery("FROM missions").WillReturnRows(sqlmock.NewRows([]string{"id", "directive", "status", "teams", "agents"}).AddRow("m-1", "Ship", "active", 1, 2))
	out, err := r.handleListMissions(ctx, nil)
	if err != nil || !strings.Contains(out, `"id":"m-1"`) {
		t.Fatalf("list_missions = %q, %v", out, err)
	}

	var resp generatedImageResponse
	_ = json.Unmarshal([]byte(`{"data":[{"b64_json":"aW1n"}]}`), &resp)
	out, err = NewInternalToolRegistry(InternalToolDeps{}).finishGeneratedImage(ctx, "logo", "512x512", resp)
	if err != nil || !strings.Contains(out, `"cached":false`) || strings.Contains(out, "Cached for 60 minutes") {
		t.Fatalf("uncached image = %q, %v; want honest not-cached result", out, err)
	}
}

// delegate_task publishes on core NATS without an ack, so it reports the task
// as queued and never as delivered or accepted.
func TestDelegateTaskReportsQueuedNotDelivered(t *testing.T) {
	_, nc := startTestNATS(t)
	r := NewInternalToolRegistry(InternalToolDeps{NC: nc})
	out, err := r.handleDelegateTask(context.Background(), map[string]any{"team_id": "admin-core", "task": "inspect gate state"})
	if err != nil {
		t.Fatalf("delegate_task: %v", err)
	}
	if !strings.Contains(out, "queued for team admin-core") || !strings.Contains(out, "not yet confirmed") || strings.Contains(out, "delegated") {
		t.Fatalf("delegate_task output = %q, want an honest queued result", out)
	}
}
