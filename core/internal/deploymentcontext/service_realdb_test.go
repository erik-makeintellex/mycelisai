package deploymentcontext

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Real PostgreSQL proof. MYCELIS_MEMORY_TEST_DSN must point at a disposable
// database with 001_current_schema.sql installed. Skip is not proof.
func openDeploymentContextTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYCELIS_MEMORY_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_MEMORY_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`SELECT 1 FROM context_vectors LIMIT 1`); err != nil {
		t.Fatalf("DSN must have 001_current_schema.sql installed: %v", err)
	}
	cleanup := func() {
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE metadata->>'knowledge_store' = 'governed_context_store'`)
		_, _ = db.Exec(`DELETE FROM artifacts WHERE metadata->>'knowledge_store' = 'governed_context_store'`)
	}
	cleanup()
	t.Cleanup(cleanup)
	return db
}

type fakeEmbedEngine struct {
	calls atomic.Int32
	fail  atomic.Bool
	dims  int
}

func (f *fakeEmbedEngine) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok"}, nil
}
func (f *fakeEmbedEngine) Probe(context.Context) (bool, error) { return true, nil }
func (f *fakeEmbedEngine) Embed(context.Context, string, string) ([]float64, error) {
	f.calls.Add(1)
	if f.fail.Load() {
		return nil, errors.New("embedding failed")
	}
	vec := make([]float64, f.dims)
	vec[0] = 1
	return vec, nil
}

func routerWith(engine *fakeEmbedEngine) *cognitive.Router {
	return &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"chat": "stub", "embed": "stub"}, Providers: map[string]cognitive.ProviderConfig{"stub": {Enabled: true}}},
		Adapters: map[string]cognitive.LLMProvider{"stub": engine},
	}
}

func serviceWith(db *sql.DB, router *cognitive.Router) *Service {
	return NewService(artifacts.NewService(db, os.TempDir()), memory.NewServiceWithDB(db), router)
}

const bakeryContent = "Weekend special: blueberry-lavender scones, 3 for $10, Saturday and Sunday. Sign-off line: See you at the counter."

func chunkStates(t *testing.T, db *sql.DB, artifactID string) (total, nullEmbedding int, statuses map[string]int) {
	t.Helper()
	rows, err := db.Query(`SELECT embedding IS NULL, COALESCE(metadata->>'embedding_status','') FROM context_vectors WHERE metadata->>'artifact_id' = $1`, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	statuses = map[string]int{}
	for rows.Next() {
		var isNull bool
		var status string
		if err := rows.Scan(&isNull, &status); err != nil {
			t.Fatal(err)
		}
		total++
		if isNull {
			nullEmbedding++
		}
		statuses[status]++
	}
	return total, nullEmbedding, statuses
}

func TestIngestRealDB_NoEmbedderStoresPendingLexicalRows(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	engine := &fakeEmbedEngine{dims: 768}
	engine.fail.Store(true)
	result, err := serviceWith(db, routerWith(engine)).Ingest(context.Background(), IngestRequest{
		KnowledgeClass: KnowledgeClassCompanyKnowledge, Title: "Juniper & Rye Bakery", Content: bakeryContent,
	})
	if err != nil {
		t.Fatalf("save without embeddings must succeed: %v", err)
	}
	if result.EmbeddingStatus != EmbeddingStatusPending || result.VectorCount != 0 || result.ChunkCount != 1 {
		t.Fatalf("honest status required: %+v", result)
	}
	if len(result.RetrievalModes) != 1 || result.RetrievalModes[0] != "keyword" || result.StatusMessage == "" {
		t.Fatalf("retrieval modes must be keyword only: %+v", result)
	}
	total, nulls, statuses := chunkStates(t, db, result.ArtifactID)
	if total != 1 || nulls != 1 || statuses[EmbeddingStatusPending] != 1 {
		t.Fatalf("chunk rows: total=%d null=%d statuses=%v", total, nulls, statuses)
	}
	entries, err := serviceWith(db, nil).List(context.Background(), 10)
	if err != nil || len(entries) != 1 || entries[0].EmbeddingStatus != EmbeddingStatusPending || entries[0].VectorCount != 0 {
		t.Fatalf("list after reload must show the pending entry without a false vector count: %+v err=%v", entries, err)
	}
}

func TestIngestRealDB_WorkingEmbedderMarksEmbedded(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	result, err := serviceWith(db, routerWith(&fakeEmbedEngine{dims: 768})).Ingest(context.Background(), IngestRequest{Title: "Menu", Content: bakeryContent})
	if err != nil {
		t.Fatal(err)
	}
	if result.EmbeddingStatus != EmbeddingStatusEmbedded || result.VectorCount != 1 || len(result.RetrievalModes) != 2 {
		t.Fatalf("working embedder: %+v", result)
	}
	if _, nulls, statuses := chunkStates(t, db, result.ArtifactID); nulls != 0 || statuses[EmbeddingStatusEmbedded] != 1 {
		t.Fatalf("chunk must be embedded: nulls=%d statuses=%v", nulls, statuses)
	}
}

func TestIngestRealDB_DimensionMismatchStaysKeywordSearchable(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	result, err := serviceWith(db, routerWith(&fakeEmbedEngine{dims: 2})).Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent})
	if err != nil {
		t.Fatal(err)
	}
	if result.VectorCount != 0 || result.EmbeddingStatus != EmbeddingStatusFailedDimension {
		t.Fatalf("a 2-dim vector must never count as embedded: %+v", result)
	}
	if _, nulls, statuses := chunkStates(t, db, result.ArtifactID); nulls != 1 || statuses[EmbeddingStatusFailedDimension] != 1 {
		t.Fatalf("mismatched vector must not be inserted: nulls=%d statuses=%v", nulls, statuses)
	}
	hits, mode, err := memory.NewServiceWithDB(db).RecallGoverned(context.Background(), nil, "weekend special", memory.SemanticSearchOptions{Limit: 3})
	if err != nil || mode != memory.RecallModeKeyword || len(hits) == 0 {
		t.Fatalf("row must stay keyword-searchable: hits=%d mode=%q err=%v", len(hits), mode, err)
	}
}

func TestBackfillRealDB_PendingToEmbeddedAndConcurrentRunsSkipLocked(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svcDown := serviceWith(db, routerWith(down))
	var ids []string
	for i := 0; i < 6; i++ {
		result, err := svcDown.Ingest(context.Background(), IngestRequest{Title: "Pending note", Content: bakeryContent})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, result.ArtifactID)
	}
	if res, err := svcDown.BackfillEmbeddings(context.Background(), 20); err != nil || res.Embedded != 0 || res.Remaining != 6 || res.Status != BackfillStatusUnavailable {
		t.Fatalf("backfill without an engine must say so: %+v err=%v", res, err)
	}

	up := &fakeEmbedEngine{dims: 768}
	svcUp := serviceWith(db, routerWith(up))
	if !svcUp.Cognitive.EmbeddingAvailable(context.Background()) {
		t.Fatal("engine should be available")
	}
	probeCalls := up.calls.Load()
	var wg sync.WaitGroup
	var embedded atomic.Int32
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := svcUp.BackfillEmbeddings(context.Background(), 20)
			if err != nil {
				t.Errorf("backfill: %v", err)
				return
			}
			embedded.Add(int32(res.Embedded))
		}()
	}
	wg.Wait()
	if embedded.Load() != 6 || up.calls.Load()-probeCalls != 6 {
		t.Fatalf("SKIP LOCKED must embed each row once: embedded=%d embedCalls=%d", embedded.Load(), up.calls.Load()-probeCalls)
	}
	for _, id := range ids {
		if _, nulls, statuses := chunkStates(t, db, id); nulls != 0 || statuses[EmbeddingStatusEmbedded] != 1 {
			t.Fatalf("artifact %s not embedded: %v", id, statuses)
		}
	}
	entries, _ := svcUp.List(context.Background(), 10)
	for _, entry := range entries {
		if entry.EmbeddingStatus != EmbeddingStatusEmbedded || entry.VectorCount != 1 {
			t.Fatalf("artifact metadata must follow backfill: %+v", entry)
		}
	}
}

func TestIngestRealDB_OpportunisticBackfillAfterSuccessfulEmbed(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	pending, err := serviceWith(db, routerWith(down)).Ingest(context.Background(), IngestRequest{Title: "Earlier note", Content: bakeryContent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serviceWith(db, routerWith(&fakeEmbedEngine{dims: 768})).Ingest(context.Background(), IngestRequest{Title: "Later note", Content: "Opening hours are seven to three."}); err != nil {
		t.Fatal(err)
	}
	if _, nulls, _ := chunkStates(t, db, pending.ArtifactID); nulls != 0 {
		t.Fatal("successful ingest must opportunistically backfill earlier pending rows")
	}
}
