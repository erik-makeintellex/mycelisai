package swarm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

type fakeMemoryProvider struct {
	inferText string
	embedVec  []float64
}

func (f fakeMemoryProvider) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{
		Text:      f.inferText,
		ModelUsed: "stub-model",
		Provider:  "stub",
	}, nil
}

func (f fakeMemoryProvider) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func (f fakeMemoryProvider) Embed(_ context.Context, _ string, _ string) ([]float64, error) {
	return f.embedVec, nil
}

func newFakeBrain(provider fakeMemoryProvider) *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{
				"stub": {Enabled: true, ModelID: "stub-model"},
			},
			Profiles: map[string]string{
				"chat":  "stub",
				"embed": "stub",
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"stub": provider,
		},
	}
}

func TestHandleSearchMemory_UsesInvocationTeamScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	nowRows := sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
		AddRow("vec-1", "team memory", `{"team_id":"alpha","visibility":"team"}`, 0.88, time.Now())

	mock.ExpectQuery("SELECT id, content, metadata, 1 - \\(embedding <=> \\$1::vector\\) AS score, created_at").
		WithArgs(sqlmock.AnyArg(), "default", "alpha", "lead-alpha", 5).
		WillReturnRows(nowRows)

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{embedVec: []float64{0.1, 0.2}}),
		Mem:   mem,
	})
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		TeamID:  "alpha",
		AgentID: "lead-alpha",
	})

	out, err := registry.handleSearchMemory(ctx, map[string]any{"query": "planning memory"})
	if err != nil {
		t.Fatalf("handleSearchMemory: %v", err)
	}
	if out == "" {
		t.Fatal("expected search output")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAutoSummarize_StoresTempPlanningCheckpoint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", "team.alpha.planning", "lead-alpha", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tmp-1"))

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{
			inferText: `{"summary":"Current work is narrowing toward an execution plan.","key_topics":["execution","planning"],"user_preferences":{"tone":"direct"},"personality_notes":"keep the plan practical","data_references":[]}`,
			embedVec:  []float64{0.1, 0.2},
		}),
		Mem: mem,
	})

	registry.AutoSummarize(context.Background(), "lead-alpha", "alpha", []cognitive.ChatMessage{
		{Role: "user", Content: "Let's think through the plan first."},
		{Role: "assistant", Content: "I'll outline the planning options."},
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleSummarizeConversation_PromotesDurableMemory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	mock.ExpectQuery("INSERT INTO conversation_summaries").
		WithArgs("lead-alpha", "Saved summary", sqlmock.AnyArg(), sqlmock.AnyArg(), "treat the user as a collaborator", sqlmock.AnyArg(), 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("sum-1"))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs("[conversation] Saved summary", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{
			inferText: `{"summary":"Saved summary","key_topics":["governance"],"user_preferences":{"detail":"high"},"personality_notes":"treat the user as a collaborator","data_references":[]}`,
			embedVec:  []float64{0.1, 0.2},
		}),
		Mem: mem,
	})
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		TeamID:  "alpha",
		AgentID: "lead-alpha",
		RunID:   "run-42",
	})

	out, err := registry.handleSummarizeConversation(ctx, map[string]any{"messages": "Important durable conversation"})
	if err != nil {
		t.Fatalf("handleSummarizeConversation: %v", err)
	}
	if out == "" {
		t.Fatal("expected summary id output")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleLoadDeploymentContext_PersistsArtifactAndVectors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"lead-alpha", sqlmock.AnyArg(),
			sqlmock.AnyArg(), "Operator Notes", "text/markdown",
			"Mycelis should keep MCP web access reviewable.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", time.Now()))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{embedVec: []float64{0.1, 0.2}}),
		Mem:   mem,
		DB:    db,
	})
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		TeamID:  "alpha",
		AgentID: "lead-alpha",
	})

	out, err := registry.handleLoadDeploymentContext(ctx, map[string]any{
		"knowledge_class": "company_knowledge",
		"title":           "Operator Notes",
		"content":         "Mycelis should keep MCP web access reviewable.",
		"tags":            []any{"deployment", "security"},
	})
	if err != nil {
		t.Fatalf("handleLoadDeploymentContext: %v", err)
	}
	if !strings.Contains(out, "company_knowledge") || !strings.Contains(out, "governed_knowledge") {
		t.Fatalf("expected deployment context payload, got %s", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBuildContext_IncludesDeploymentContextRecall(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	mock.ExpectQuery("SELECT id, content, metadata, 1 - \\(embedding <=> \\$1::vector\\) AS score, created_at").
		WithArgs(sqlmock.AnyArg(), "default", "customer_context", "company_knowledge", "alpha", 3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
			AddRow("vec-ctx", "[customer_context] secure web access via MCP policy gates", `{"artifact_title":"Security Brief","source_label":"operator brief","visibility":"global","knowledge_class":"customer_context"}`, 0.93, time.Now()).
			AddRow("vec-company", "[company_knowledge] approved deployment playbook", `{"artifact_title":"Deployment Playbook","source_label":"approved company guide","visibility":"global","knowledge_class":"company_knowledge"}`, 0.9, time.Now()))

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{embedVec: []float64{0.1, 0.2}}),
		Mem:   mem,
	})

	text := registry.BuildContext("lead-alpha", "alpha", "lead", nil, nil, "Summarize our MCP security posture.")
	if !strings.Contains(text, "Customer Context Store") || !strings.Contains(text, "Company Knowledge Store") || !strings.Contains(text, "Security Brief") || !strings.Contains(text, "Deployment Playbook") {
		t.Fatalf("expected deployment context in build context, got %s", text)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
