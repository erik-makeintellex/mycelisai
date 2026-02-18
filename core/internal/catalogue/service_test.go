package catalogue

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var catColumns = []string{
	"id", "name", "role", "system_prompt", "model",
	"tools", "inputs", "outputs",
	"verification_strategy", "verification_rubric", "validation_command",
	"created_at", "updated_at",
}

func TestCatalogueService_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)

	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(catColumns).
		AddRow(id1, "Code Reviewer", "cognitive", "Review code", "qwen2.5",
			[]byte(`["read_file"]`), []byte(`["swarm.data.>"]`), []byte(`["swarm.output.>"]`),
			"semantic", []byte(`["Check correctness"]`), nil,
			now, now).
		AddRow(id2, "Weather Sensor", "sensory", nil, nil,
			[]byte(`[]`), []byte(`[]`), []byte(`[]`),
			nil, []byte(`[]`), nil,
			now, now)

	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").WillReturnRows(rows)

	agents, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("Expected 2 agents, got %d", len(agents))
	}
	if agents[0].Name != "Code Reviewer" {
		t.Errorf("Expected 'Code Reviewer', got %q", agents[0].Name)
	}
	if agents[0].Role != "cognitive" {
		t.Errorf("Expected role 'cognitive', got %q", agents[0].Role)
	}
	if len(agents[0].Tools) != 1 || agents[0].Tools[0] != "read_file" {
		t.Errorf("Expected tools [read_file], got %v", agents[0].Tools)
	}
	if agents[1].Name != "Weather Sensor" {
		t.Errorf("Expected 'Weather Sensor', got %q", agents[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_ListEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	rows := sqlmock.NewRows(catColumns)
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").WillReturnRows(rows)

	agents, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if agents != nil && len(agents) != 0 {
		t.Errorf("Expected empty list, got %d agents", len(agents))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	newID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO agent_catalogue").
		WithArgs("Test Agent", "cognitive",
			sqlmock.AnyArg(), sqlmock.AnyArg(), // system_prompt, model (NullString)
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // tools, inputs, outputs JSON
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // ver_strat, rubric, val_cmd
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(newID, now, now))

	input := AgentTemplate{
		Name:         "Test Agent",
		Role:         "cognitive",
		SystemPrompt: "You are a test agent",
		Model:        "qwen2.5",
		Tools:        []string{"read_file", "write_file"},
	}

	result, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.ID != newID {
		t.Errorf("Expected ID %s, got %s", newID, result.ID)
	}
	if result.Name != "Test Agent" {
		t.Errorf("Expected name 'Test Agent', got %q", result.Name)
	}
	if len(result.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(result.Tools))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_CreateNilSlices(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	newID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO agent_catalogue").
		WithArgs("Minimal Agent", "sensory",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(newID, now, now))

	// nil slices should be initialized to empty
	input := AgentTemplate{
		Name: "Minimal Agent",
		Role: "sensory",
	}

	result, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.Tools == nil {
		t.Error("Expected Tools to be initialized, got nil")
	}
	if result.Inputs == nil {
		t.Error("Expected Inputs to be initialized, got nil")
	}
	if result.Outputs == nil {
		t.Error("Expected Outputs to be initialized, got nil")
	}
	if result.VerificationRubric == nil {
		t.Error("Expected VerificationRubric to be initialized, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	agentID := uuid.New()

	mock.ExpectExec("DELETE FROM agent_catalogue").
		WithArgs(agentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.Delete(context.Background(), agentID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_DeleteNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	agentID := uuid.New()

	mock.ExpectExec("DELETE FROM agent_catalogue").
		WithArgs(agentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = svc.Delete(context.Background(), agentID)
	if err == nil {
		t.Fatal("Expected error for non-existent agent, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCatalogueService_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db)
	agentID := uuid.New()
	now := time.Now()

	// Expect UPDATE
	mock.ExpectExec("UPDATE agent_catalogue").
		WithArgs("Updated Agent", "actuation",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			agentID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect the follow-up GET (Update calls Get internally)
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue WHERE id").
		WithArgs(agentID).
		WillReturnRows(sqlmock.NewRows(catColumns).
			AddRow(agentID, "Updated Agent", "actuation", "Updated prompt", "qwen2.5",
				[]byte(`["tool_a"]`), []byte(`[]`), []byte(`[]`),
				nil, []byte(`[]`), nil,
				now, now))

	input := AgentTemplate{
		Name:         "Updated Agent",
		Role:         "actuation",
		SystemPrompt: "Updated prompt",
		Model:        "qwen2.5",
		Tools:        []string{"tool_a"},
	}

	result, err := svc.Update(context.Background(), agentID, input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.Name != "Updated Agent" {
		t.Errorf("Expected 'Updated Agent', got %q", result.Name)
	}
	if result.Role != "actuation" {
		t.Errorf("Expected role 'actuation', got %q", result.Role)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}
