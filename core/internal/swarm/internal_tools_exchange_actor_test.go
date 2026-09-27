package swarm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
)

func agentInvocation(agentID, teamID, role, runID string) context.Context {
	return WithToolInvocationContext(context.Background(), ToolInvocationContext{
		AgentID: agentID, TeamID: teamID, AgentRole: role, RunID: runID,
	})
}

func expectPlanningChannel(mock sqlmock.Sqlmock) {
	participants, _ := json.Marshal([]exchange.ChannelParticipant{
		{Role: "soma", CanRead: true, CanWrite: true},
		{Role: "team_lead", CanRead: true, CanWrite: true},
		{Role: "specialist", CanRead: true, CanWrite: false},
	})
	mock.ExpectQuery("FROM exchange_channels").
		WithArgs("organization.planning.work").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "channel_type", "owner", "participants", "reviewers", "schema_id", "retention_policy", "visibility", "sensitivity_class", "description", "metadata", "created_at"}).
			AddRow(uuid.New(), "organization.planning.work", "planning", "system", participants, []byte(`["review"]`), "PlanResult", "90d", "advanced", "role_scoped", "", []byte(`{}`), time.Now()))
}

// SECURITY: a model cannot mint exchange authority through tool arguments.
func TestPublishExchangeItemIgnoresForgedAdminRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectPlanningChannel(mock)
	registry := &InternalToolRegistry{exchange: exchange.NewService(db, nil, nil)}

	ctx := agentInvocation("research-analyst", "research", "researcher", "run-1")
	_, err = registry.handlePublishExchangeItem(ctx, map[string]any{
		"channel": "organization.planning.work", "schema_id": "PlanResult",
		"role": "admin", "created_by": "admin", "source_role": "admin",
		"payload": map[string]any{"summary": "forged", "status": "open", "priority": "high", "source_role": "admin", "target_role": "soma", "created_at": "2026-09-27T00:00:00Z"},
	})
	if err == nil || !strings.Contains(err.Error(), "role specialist cannot publish") {
		t.Fatalf("forged admin publish error = %v, want specialist write denial", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestExchangeActorComesFromInvocationIdentity(t *testing.T) {
	forged := map[string]any{"role": "admin", "created_by": "admin"}
	cases := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"specialist", agentInvocation("research-analyst", "research", "researcher", "run-1"), "specialist"},
		{"team lead", agentInvocation("research-lead", "research", "lead", "run-1"), "team_lead"},
		{"soma", agentInvocation("admin", "admin-core", "admin", ""), "soma"},
		{"runtime agent named soma", agentInvocation("soma-x", "research", "soma", "run-1"), "specialist"},
	}
	for _, tc := range cases {
		actor := exchangeActorForInvocation(tc.ctx, forged)
		if actor.Role != tc.want || actor.IsAdmin() {
			t.Fatalf("%s: actor = %#v, want role %s and no admin", tc.name, actor, tc.want)
		}
	}
}

type jsonMetadataMatcher func(map[string]any) bool

func (m jsonMetadataMatcher) Match(v driver.Value) bool {
	raw, ok := v.(string)
	if !ok {
		return false
	}
	var decoded map[string]any
	return json.Unmarshal([]byte(raw), &decoded) == nil && m(decoded)
}

func TestStoreArtifactRecordsInvocationProvenance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	registry := &InternalToolRegistry{db: db}
	attributed := jsonMetadataMatcher(func(meta map[string]any) bool {
		return meta["provenance"] == "attributed" && meta["provenance_agent_id"] == "research-lead" &&
			meta["provenance_team_id"] == "research" && meta["provenance_run_id"] == "run-7" &&
			meta["provenance_source"] == "tool_invocation" && meta["topic"] == "bakery" &&
			meta["provenance_sensitivity"] == "team_scoped"
	})
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs("research-lead", "document", "Fact sheet", "text/plain", "Juniper & Rye facts", attributed).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-4111-8111-111111111111"))

	ctx := agentInvocation("research-lead", "research", "lead", "run-7")
	_, err = registry.handleStoreArtifact(ctx, map[string]any{
		"type": "document", "title": "Fact sheet", "content": "Juniper & Rye facts",
		// Model-supplied provenance must be overwritten, never trusted.
		"metadata": map[string]any{"topic": "bakery", "provenance_team_id": "marketing", "provenance": "attributed", "sensitivity_class": "public", "provenance_sensitivity": "public"},
	})
	if err != nil {
		t.Fatalf("store_artifact: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreArtifactWithoutInvocationIsUnattributed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	registry := &InternalToolRegistry{db: db}
	unattributed := jsonMetadataMatcher(func(meta map[string]any) bool {
		_, hasTeam := meta["provenance_team_id"]
		return meta["provenance"] == "unattributed" && !hasTeam
	})
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs("internal", "document", "Note", "text/plain", "body", unattributed).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-4111-8111-111111111112"))
	if _, err := registry.handleStoreArtifact(context.Background(), map[string]any{"type": "document", "title": "Note", "content": "body", "metadata": map[string]any{"provenance_team_id": "research"}}); err != nil {
		t.Fatalf("store_artifact: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// A model may raise its artifact's classification but never lower it.
func TestStoreArtifactModelCanOnlyRaiseSensitivity(t *testing.T) {
	ctx := agentInvocation("research-lead", "research", "lead", "run-7")
	for label, want := range map[string]string{"public": "team_scoped", "restricted": "restricted", "admin_only": "admin_only", "": "team_scoped"} {
		_, raw := attributeArtifact(ctx, `{"sensitivity_class":"`+label+`"}`)
		var meta map[string]any
		_ = json.Unmarshal([]byte(raw), &meta)
		if meta["provenance_sensitivity"] != want {
			t.Fatalf("model label %q -> %v, want %s", label, meta["provenance_sensitivity"], want)
		}
	}
}
