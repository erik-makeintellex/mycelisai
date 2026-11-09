package server

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

type stubEmbedAdapter struct{}

func (s *stubEmbedAdapter) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok"}, nil
}

func (s *stubEmbedAdapter) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func (s *stubEmbedAdapter) Embed(_ context.Context, _ string, _ string) ([]float64, error) {
	return []float64{0.1, 0.2, 0.3}, nil
}

func TestHandleMemorySearch_ArbitrationPrecedenceHaltsAndSupersession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s := newTestServer()
	s.Mem = memory.NewServiceWithDB(db)
	s.Cognitive = &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"embed": "mock"}},
		Adapters: map[string]cognitive.LLMProvider{"mock": &stubEmbedAdapter{}},
	}

	now := time.Now().UTC()
	metaReflection, _ := json.Marshal(map[string]any{
		"knowledge_class": "reflection_synthesis",
		"reflection_kind": "lesson",
		"source_kind":     "synthesis_note",
	})
	metaDoctrine, _ := json.Marshal(map[string]any{
		"knowledge_class": "company_knowledge",
		"artifact_id":     "art-doc-10",
		"source_kind":     "governance_policy",
		"trust_class":     "trusted_internal",
	})
	metaSuperseded, _ := json.Marshal(map[string]any{
		"knowledge_class": "customer_context",
		"superseded_by":   "vec-doctrine",
	})
	metaContradicted, _ := json.Marshal(map[string]any{
		"knowledge_class": "reflection_synthesis",
		"contradicts_id":  "vec-doctrine",
	})

	rows := sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
		AddRow("vec-reflection", "Candidate reflection note", metaReflection, 0.96, now).
		AddRow("vec-superseded", "Old superseded customer note", metaSuperseded, 0.94, now).
		AddRow("vec-contradicted", "Contradicting claim", metaContradicted, 0.92, now).
		AddRow("vec-doctrine", "Governed architecture standard", metaDoctrine, 0.85, now)

	mock.ExpectQuery("SELECT id, content, metadata").
		WillReturnRows(rows)

	rr := doRequest(t, http.HandlerFunc(s.HandleMemorySearch), http.MethodGet, "/api/v1/memory/search?q=architecture&limit=5", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		Query       string                  `json:"query"`
		Count       int                     `json:"count"`
		Results     []memory.VectorResult   `json:"results"`
		Arbitration memory.ArbitrationSummary `json:"arbitration"`
	}
	assertJSON(t, rr, &resp)

	if resp.Count != 2 || len(resp.Results) != 2 {
		t.Fatalf("expected 2 active results after filtering, got count=%d len=%d", resp.Count, len(resp.Results))
	}
	if resp.Results[0].ID != "vec-doctrine" || resp.Results[1].ID != "vec-reflection" {
		t.Fatalf("expected doctrine to outrank reflection, got [%s, %s]", resp.Results[0].ID, resp.Results[1].ID)
	}
	if resp.Results[0].Metadata["memory_layer"] != string(memory.LayerGovernedDoctrine) {
		t.Fatalf("expected GOVERNED_DOCTRINE layer on first result, got %v", resp.Results[0].Metadata["memory_layer"])
	}
	if resp.Arbitration.SupersededFiltered != 1 {
		t.Fatalf("expected 1 superseded vector filtered, got %d", resp.Arbitration.SupersededFiltered)
	}
	if !resp.Arbitration.HaltRecommended || len(resp.Arbitration.ContradictionHalts) != 1 {
		t.Fatalf("expected contradiction halt in arbitration summary, got %+v", resp.Arbitration)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleMemorySearch_BoundedGrowthPerLayerParam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s := newTestServer()
	s.Mem = memory.NewServiceWithDB(db)
	s.Cognitive = &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"embed": "mock"}},
		Adapters: map[string]cognitive.LLMProvider{"mock": &stubEmbedAdapter{}},
	}

	now := time.Now().UTC()
	metaDoc, _ := json.Marshal(map[string]any{"knowledge_class": "company_knowledge", "artifact_id": "a1"})
	rows := sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
		AddRow("doc-1", "Rule 1", metaDoc, 0.95, now).
		AddRow("doc-2", "Rule 2", metaDoc, 0.90, now)

	mock.ExpectQuery("SELECT id, content, metadata").
		WillReturnRows(rows)

	rr := doRequest(t, http.HandlerFunc(s.HandleMemorySearch), http.MethodGet, "/api/v1/memory/search?q=rules&limit=5&max_per_layer=1", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		Count       int                       `json:"count"`
		Results     []memory.VectorResult     `json:"results"`
		Arbitration memory.ArbitrationSummary `json:"arbitration"`
	}
	assertJSON(t, rr, &resp)

	if resp.Count != 1 || len(resp.Results) != 1 || resp.Results[0].ID != "doc-1" {
		t.Fatalf("expected max_per_layer=1 to keep only doc-1, got %+v", resp.Results)
	}
	if resp.Arbitration.OverflowTrimmed != 1 {
		t.Fatalf("expected 1 overflow trimmed, got %d", resp.Arbitration.OverflowTrimmed)
	}
}

func TestHandleMemoryPromote_LayersAndValidation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s := newTestServer()
	s.Mem = memory.NewServiceWithDB(db)

	now := time.Now().UTC()
	metaPromoted, _ := json.Marshal(map[string]any{
		"memory_layer":        "GOVERNED_DOCTRINE",
		"verification_status": "verified",
		"verified":            true,
	})
	mock.ExpectQuery("UPDATE context_vectors").
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
			AddRow("vec-1", "Standard rule", metaPromoted, 1.0, now))

	rr := doRequest(t, http.HandlerFunc(s.HandleMemoryPromote), http.MethodPost, "/api/v1/memory/promote",
		`{"id":"vec-1","target_layer":"GOVERNED_DOCTRINE","verification_status":"verified","promoted_by":"operator"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		OK                 bool           `json:"ok"`
		ID                 string         `json:"id"`
		MemoryLayer        string         `json:"memory_layer"`
		VerificationStatus string         `json:"verification_status"`
		Metadata           map[string]any `json:"metadata"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK || resp.ID != "vec-1" || resp.MemoryLayer != "GOVERNED_DOCTRINE" || resp.VerificationStatus != "verified" {
		t.Fatalf("unexpected promote response: %+v", resp)
	}

	// Default target_layer = PROJECT_MEMORY in synthetic/no-DB fallback
	sNoDB := newTestServer()
	rrProj := doRequest(t, http.HandlerFunc(sNoDB.HandleMemoryPromote), http.MethodPost, "/api/v1/memory/promote", `{"id":"vec-2"}`)
	assertStatus(t, rrProj, http.StatusOK)
	assertJSON(t, rrProj, &resp)
	if resp.MemoryLayer != "PROJECT_MEMORY" || resp.VerificationStatus != "verified" {
		t.Fatalf("expected PROJECT_MEMORY & verified defaults, got %+v", resp)
	}

	// Validation errors
	rrMissingID := doRequest(t, http.HandlerFunc(s.HandleMemoryPromote), http.MethodPost, "/api/v1/memory/promote", `{"target_layer":"PROJECT_MEMORY"}`)
	assertStatus(t, rrMissingID, http.StatusBadRequest)

	rrBadLayer := doRequest(t, http.HandlerFunc(s.HandleMemoryPromote), http.MethodPost, "/api/v1/memory/promote", `{"id":"vec-1","target_layer":"INVALID_LAYER"}`)
	assertStatus(t, rrBadLayer, http.StatusBadRequest)
}

func TestHandleMemoryResolveConflict_KeepSupersedeAndValidation(t *testing.T) {
	s := newTestServer()

	rrKeep := doRequest(t, http.HandlerFunc(s.HandleMemoryResolveConflict), http.MethodPost, "/api/v1/memory/resolve-conflict",
		`{"id":"vec-conflict-1","resolution":"keep","reason":"Verified with architecture owner"}`)
	assertStatus(t, rrKeep, http.StatusOK)

	var resp struct {
		OK                 bool           `json:"ok"`
		ID                 string         `json:"id"`
		Resolution         string         `json:"resolution"`
		VerificationStatus string         `json:"verification_status"`
		Metadata           map[string]any `json:"metadata"`
	}
	assertJSON(t, rrKeep, &resp)
	if !resp.OK || resp.Resolution != "keep" || resp.VerificationStatus != "verified" || resp.Metadata["contradicted"] != false {
		t.Fatalf("unexpected keep resolution response: %+v", resp)
	}

	rrSup := doRequest(t, http.HandlerFunc(s.HandleMemoryResolveConflict), http.MethodPost, "/api/v1/memory/resolve-conflict",
		`{"id":"vec-conflict-2","resolution":"supersede","reason":"Outdated draft"}`)
	assertStatus(t, rrSup, http.StatusOK)
	assertJSON(t, rrSup, &resp)
	if !resp.OK || resp.Resolution != "supersede" || resp.VerificationStatus != "superseded" || resp.Metadata["superseded"] != true {
		t.Fatalf("unexpected supersede resolution response: %+v", resp)
	}

	rrBad := doRequest(t, http.HandlerFunc(s.HandleMemoryResolveConflict), http.MethodPost, "/api/v1/memory/resolve-conflict",
		`{"id":"vec-conflict-2","resolution":"ignore"}`)
	assertStatus(t, rrBad, http.StatusBadRequest)
}
