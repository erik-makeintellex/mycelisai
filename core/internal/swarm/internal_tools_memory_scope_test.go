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

// expectNarrowEngineDeploymentSave expects an M1 save whose engine returns
// 2-dim vectors: the atomic insert, the chunk marked failed_dimension (never
// stored as a vector), and the artifact status refresh.
func expectNarrowEngineDeploymentSave(mock sqlmock.Sqlmock, agentID, title, content, artifactID string) {
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(agentID, title, "text/markdown", content, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(artifactID, time.Now()))
	mock.ExpectQuery("INSERT INTO context_vectors").WithArgs(content, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))
	mock.ExpectCommit()
	mock.ExpectExec(`UPDATE context_vectors SET metadata = metadata \|\| '\{"embedding_status": "failed_dimension"\}'`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("SELECT id FROM artifacts").WithArgs("{" + artifactID + "}").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE artifacts a SET metadata").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "embedded"}).AddRow(artifactID, "failed_dimension", 0))
	mock.ExpectCommit()
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

	// The 2-dim fake engine cannot serve the 768-dim store, so recall is keyword.
	mock.ExpectQuery(`SELECT id, content, metadata, ts_rank_cd\(`).
		// SRU: a call with no requesting user reads with the empty reader;
		// MEM-LANES adds the owner-lane rule (lane types, user, team keys).
		WithArgs("default", "alpha", "lead-alpha", "", "", "{}", `{"agent_memory","conversation"}`, "", "{}", "planning memory", 5).
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

	registry.AutoSummarize(context.Background(), RecallAccess{}, "lead-alpha", "alpha", []cognitive.ChatMessage{
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
	expectNarrowEngineDeploymentSave(mock, "lead-alpha", "Operator Notes", "Mycelis should keep MCP web access reviewable.", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

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

func TestHandlePromoteDeploymentContext_PromotesCustomerContextToCompanyKnowledge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	sourceMeta := `{"knowledge_class":"customer_context","source_label":"customer brief","source_kind":"user_document","visibility":"global","sensitivity_class":"role_scoped","trust_class":"user_provided","tags":["deployment","security"]}`
	// SRU: the requesting user must be able to read the source first.
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM artifacts a WHERE a.id::text = \$1`).
		WithArgs("ctx-1", "owner-user-id", "owner-user", "{}").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT title, content, metadata FROM artifacts WHERE id::text = \\$1").
		WithArgs("ctx-1").
		WillReturnRows(sqlmock.NewRows([]string{"title", "content", "metadata"}).
			AddRow("Customer Deployment Brief", "Use reviewed MCP web access only.", []byte(sourceMeta)))
	expectNarrowEngineDeploymentSave(mock, "lead-alpha", "Approved Deployment Guidance", "Use reviewed MCP web access only.", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{embedVec: []float64{0.1, 0.2}}),
		Mem:   mem,
		DB:    db,
	})
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		TeamID:    "alpha",
		AgentID:   "lead-alpha",
		UserLabel: "owner-user",
		Recall:    RecallAccess{User: true, Reader: memory.GovernedReader{UserID: "owner-user-id", Label: "owner-user"}},
	})

	out, err := registry.handlePromoteDeploymentContext(ctx, map[string]any{
		"source_artifact_id": "ctx-1",
		"title":              "Approved Deployment Guidance",
	})
	if err != nil {
		t.Fatalf("handlePromoteDeploymentContext: %v", err)
	}
	if !strings.Contains(out, "company_knowledge") || !strings.Contains(out, "source_artifact_id") {
		t.Fatalf("expected promotion payload, got %s", out)
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
	// The 2-dim fake engine cannot serve the 768-dim store, so recall is keyword.
	mock.ExpectQuery(`SELECT id, content, metadata, ts_rank_cd\(`).
		WithArgs("default", "customer_context", "company_knowledge", "soma_operating_context", "user_private_context", "reflection_synthesis", "alpha", "lead-alpha", "", "", "{}", `{"agent_memory","conversation"}`, "", "{}", "Summarize our MCP security posture.", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
			AddRow("vec-ctx", "[customer_context] secure web access via MCP policy gates", `{"artifact_title":"Security Brief","source_label":"operator brief","visibility":"global","knowledge_class":"customer_context"}`, 0.93, time.Now()).
			AddRow("vec-company", "[company_knowledge] approved deployment playbook", `{"artifact_title":"Deployment Playbook","source_label":"approved company guide","visibility":"global","knowledge_class":"company_knowledge"}`, 0.9, time.Now()).
			AddRow("vec-soma", "[soma_operating_context] executive investor-facing output by default", `{"artifact_title":"Soma Output Contract","source_label":"root admin guidance","visibility":"global","knowledge_class":"soma_operating_context"}`, 0.89, time.Now()).
			AddRow("vec-private", "[user_private_context] Q2 cash-flow notes for tax planning", `{"artifact_title":"Personal Finance Notes","source_label":"private user content","visibility":"private","knowledge_class":"user_private_context"}`, 0.88, time.Now()).
			AddRow("vec-reflection", "[reflection_synthesis] user trajectory shifted toward investor-ready media workflows", `{"artifact_title":"Investor Workflow Shift","source_label":"reflection synthesis","visibility":"private","knowledge_class":"reflection_synthesis"}`, 0.87, time.Now()))

	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: newFakeBrain(fakeMemoryProvider{embedVec: []float64{0.1, 0.2}}),
		Mem:   mem,
	})

	text := registry.BuildContext("lead-alpha", "alpha", "lead", nil, nil, "Summarize our MCP security posture.")
	if !strings.Contains(text, "Customer Context Store") || !strings.Contains(text, "Company Knowledge Store") || !strings.Contains(text, "Admin-Shaped Soma Context") || !strings.Contains(text, "User-Private Context Store") || !strings.Contains(text, "Reflection / Synthesis Memory") || !strings.Contains(text, "Security Brief") || !strings.Contains(text, "Deployment Playbook") || !strings.Contains(text, "Soma Output Contract") || !strings.Contains(text, "Personal Finance Notes") || !strings.Contains(text, "Investor Workflow Shift") {
		t.Fatalf("expected deployment context in build context, got %s", text)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
