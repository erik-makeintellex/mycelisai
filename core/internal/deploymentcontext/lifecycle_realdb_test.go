package deploymentcontext

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/mycelis/core/internal/memory"
)

func recallArtifactIDs(t *testing.T, db *sql.DB, embedder memory.Embedder, query string) (map[string]bool, string) {
	t.Helper()
	hits, mode, err := memory.NewServiceWithDB(db).RecallGoverned(context.Background(), embedder, query, memory.SemanticSearchOptions{Limit: 10, AllowGlobal: true})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	ids := map[string]bool{}
	for _, hit := range hits {
		if id, _ := hit.Metadata["artifact_id"].(string); id != "" {
			ids[id] = true
		}
	}
	return ids, mode
}

func listedIDs(t *testing.T, svc *Service, includeArchived bool) map[string]string {
	t.Helper()
	entries, err := svc.ListEntries(context.Background(), 50, includeArchived, memory.GovernedReader{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	out := map[string]string{}
	for _, entry := range entries {
		out[entry.ArtifactID] = entry.LifecycleState
	}
	return out
}

func TestLifecycleRealDB_ArchiveHidesFromRecallAndListRestoreBringsBack(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	saved, err := svc.Ingest(context.Background(), IngestRequest{KnowledgeClass: KnowledgeClassCustomerContext, Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "global"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.Ingest(context.Background(), IngestRequest{Title: "Weekend hours", Content: "Weekend special hours: open until two on Sunday.", Visibility: "global"})
	if err != nil {
		t.Fatal(err)
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "weekend special scones"); !ids[saved.ArtifactID] {
		t.Fatal("saved entry must be recallable before archive")
	}

	if err := svc.Archive(context.Background(), saved.ArtifactID, "owner"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	ids, _ := recallArtifactIDs(t, db, nil, "weekend special scones")
	if ids[saved.ArtifactID] || !ids[other.ArtifactID] {
		t.Fatalf("archived entry must leave keyword recall immediately, others stay: %v", ids)
	}
	if listed := listedIDs(t, svc, false); listed[saved.ArtifactID] != "" || listed[other.ArtifactID] != LifecycleActive {
		t.Fatalf("default list must hide archived entries: %v", listed)
	}
	if listed := listedIDs(t, svc, true); listed[saved.ArtifactID] != LifecycleArchived {
		t.Fatalf("show archived must include it with its state: %v", listed)
	}
	record, err := svc.Lookup(context.Background(), saved.ArtifactID)
	if err != nil || record.LifecycleState != LifecycleArchived {
		t.Fatalf("lookup after archive: %+v err=%v", record, err)
	}

	if err := svc.Restore(context.Background(), saved.ArtifactID, "owner"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "weekend special scones"); !ids[saved.ArtifactID] {
		t.Fatal("restore must bring the entry back into recall")
	}
	if listed := listedIDs(t, svc, false); listed[saved.ArtifactID] != LifecycleActive {
		t.Fatalf("restored entry must be listed as active: %v", listed)
	}
}

func TestLifecycleRealDB_ArchivedEntryLeavesSemanticRecall(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	engine := &fakeEmbedEngine{dims: 768}
	svc := serviceWith(db, routerWith(engine))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "global"})
	if err != nil || saved.EmbeddingStatus != EmbeddingStatusEmbedded {
		t.Fatalf("embedded save: %+v err=%v", saved, err)
	}
	if ids, mode := recallArtifactIDs(t, db, svc.Cognitive, "promo copy"); !ids[saved.ArtifactID] || mode != memory.RecallModeSemantic {
		t.Fatalf("semantic recall must find the entry first: %v mode=%s", ids, mode)
	}
	if err := svc.Archive(context.Background(), saved.ArtifactID, "owner"); err != nil {
		t.Fatal(err)
	}
	if ids, _ := recallArtifactIDs(t, db, svc.Cognitive, "promo copy"); ids[saved.ArtifactID] {
		t.Fatal("archived entry must leave semantic recall immediately")
	}
}

func TestLifecycleRealDB_DeleteRemovesContentAndChunks(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "global"})
	if err != nil {
		t.Fatal(err)
	}
	record, err := svc.Lookup(context.Background(), saved.ArtifactID)
	if err != nil || record.ChunkCount != 1 || record.KnowledgeClass != KnowledgeClassCustomerContext {
		t.Fatalf("lookup before delete: %+v err=%v", record, err)
	}
	result, err := svc.Delete(context.Background(), saved.ArtifactID)
	if err != nil || result.ChunksRemoved != 1 {
		t.Fatalf("delete: %+v err=%v", result, err)
	}
	var artifactsLeft, chunksLeft, contentLeft int
	_ = db.QueryRow(`SELECT count(*) FROM artifacts WHERE id::text = $1`, saved.ArtifactID).Scan(&artifactsLeft)
	_ = db.QueryRow(`SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1`, saved.ArtifactID).Scan(&chunksLeft)
	_ = db.QueryRow(`SELECT count(*) FROM context_vectors WHERE content LIKE '%blueberry-lavender%'`).Scan(&contentLeft)
	if artifactsLeft != 0 || chunksLeft != 0 || contentLeft != 0 {
		t.Fatalf("delete must remove content and chunks: artifacts=%d chunks=%d content=%d", artifactsLeft, chunksLeft, contentLeft)
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "weekend special scones"); ids[saved.ArtifactID] {
		t.Fatal("deleted entry must not be recalled")
	}
	if _, err := svc.Lookup(context.Background(), saved.ArtifactID); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("deleted entry lookup must be not found, got %v", err)
	}
	if _, err := svc.Delete(context.Background(), saved.ArtifactID); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("second delete must be not found, got %v", err)
	}
	if err := svc.Archive(context.Background(), "not-a-uuid", "owner"); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("unknown id must be not found, got %v", err)
	}
}
