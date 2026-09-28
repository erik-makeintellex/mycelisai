package deploymentcontext

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
)

func memStr(v string) *string { return &v }

func memChunkRows(t *testing.T, db *sql.DB, id string) (contents []string, titles []string, vectors int, statuses map[string]int) {
	t.Helper()
	rows, err := db.Query(`SELECT content, COALESCE(metadata->>'artifact_title',''), embedding IS NOT NULL, COALESCE(metadata->>'embedding_status','')
		FROM context_vectors WHERE metadata->>'artifact_id' = $1 ORDER BY (metadata->>'chunk_index')::int`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	statuses = map[string]int{}
	for rows.Next() {
		var content, title, status string
		var hasVector bool
		if err := rows.Scan(&content, &title, &hasVector, &status); err != nil {
			t.Fatal(err)
		}
		contents, titles = append(contents, content), append(titles, title)
		if hasVector {
			vectors++
		}
		statuses[status]++
	}
	return contents, titles, vectors, statuses
}

func memLookup(t *testing.T, svc *Service, id string) *EntryRecord {
	t.Helper()
	record, err := svc.Lookup(context.Background(), id)
	if err != nil {
		t.Fatalf("lookup %s: %v", id, err)
	}
	return record
}

func TestMemEditRealDB_ContentEditReplacesChunksVectorsAndRecall(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	engine := &fakeEmbedEngine{dims: 768}
	svc := serviceWith(db, routerWith(engine))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "global"})
	if err != nil || saved.EmbeddingStatus != EmbeddingStatusEmbedded {
		t.Fatalf("embedded save: %+v err=%v", saved, err)
	}
	engine.fail.Store(true) // the edit lands while embeddings are down: honest pending
	newContent := "Weekday special: cardamom buns, 2 for $7, Monday to Thursday. " + strings.Repeat("Fresh cardamom buns every morning at the counter. ", 60)
	result, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("Juniper & Rye weekday menu"),
		Content: memStr(newContent), SourceLabel: memStr("owner menu card"), Actor: "ada", Authorized: memLookup(t, svc, saved.ArtifactID)})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !result.Changed || !result.ContentChanged || result.ChunkCount < 2 || result.ChunksRemoved != 1 || result.EmbeddingStatus != EmbeddingStatusPending {
		t.Fatalf("edit result must report the re-chunk honestly: %+v", result)
	}
	contents, titles, vectors, statuses := memChunkRows(t, db, saved.ArtifactID)
	if len(contents) != result.ChunkCount || vectors != 0 || statuses[EmbeddingStatusPending] != len(contents) {
		t.Fatalf("chunks must be replaced and pending: n=%d vectors=%d statuses=%v", len(contents), vectors, statuses)
	}
	for i, c := range contents {
		if strings.Contains(c, "blueberry") || titles[i] != "Juniper & Rye weekday menu" {
			t.Fatalf("chunk %d kept stale text or title: %q / %q", i, c, titles[i])
		}
	}
	var title, content, status, label string
	var length int
	_ = db.QueryRow(`SELECT title, content, metadata->>'embedding_status', metadata->>'source_label', (metadata->>'content_length')::int FROM artifacts WHERE id::text = $1`,
		saved.ArtifactID).Scan(&title, &content, &status, &label, &length)
	if title != "Juniper & Rye weekday menu" || content != strings.TrimSpace(newContent) || status != EmbeddingStatusPending || label != "owner menu card" || length != len([]rune(strings.TrimSpace(newContent))) {
		t.Fatalf("artifact not updated: title=%q status=%q label=%q length=%d", title, status, label, length)
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "blueberry lavender scones"); ids[saved.ArtifactID] {
		t.Fatal("old text must no longer be recalled")
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "cardamom buns"); !ids[saved.ArtifactID] {
		t.Fatal("new text must be recalled at once")
	}

	// A later edit with the engine up is embedded right after commit (a fresh
	// router, since the first one cached the engine as unavailable).
	up := serviceWith(db, routerWith(&fakeEmbedEngine{dims: 768}))
	again, err := up.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Content: memStr("Short menu: rye loaves only."), Actor: "ada",
		Authorized: memLookup(t, svc, saved.ArtifactID)})
	if err != nil || again.ChunkCount != 1 || again.EmbeddingStatus != EmbeddingStatusEmbedded || again.Title != "Juniper & Rye weekday menu" || again.SourceLabel != "owner menu card" {
		t.Fatalf("second edit: %+v err=%v", again, err)
	}
	if contents, _, vectors, _ := memChunkRows(t, db, saved.ArtifactID); len(contents) != 1 || vectors != 1 {
		t.Fatalf("second edit chunks: n=%d vectors=%d", len(contents), vectors)
	}
}

func TestMemEditRealDB_TitleOnlyEditKeepsContentAndReChunksTitle(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "global"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("Quokka promo sheet"), Actor: "ada", Authorized: memLookup(t, svc, saved.ArtifactID)})
	if err != nil || !result.Changed || result.ContentChanged {
		t.Fatalf("title edit: %+v err=%v", result, err)
	}
	contents, titles, _, _ := memChunkRows(t, db, saved.ArtifactID)
	if len(contents) != 1 || contents[0] != bakeryContent || titles[0] != "Quokka promo sheet" {
		t.Fatalf("title edit must keep content and move the title: %v %v", contents, titles)
	}
	if ids, _ := recallArtifactIDs(t, db, nil, "quokka"); !ids[saved.ArtifactID] {
		t.Fatal("new title must be keyword-recallable at once")
	}
	same, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("Quokka promo sheet"), Actor: "ada", Authorized: memLookup(t, svc, saved.ArtifactID)})
	if err != nil || same.Changed {
		t.Fatalf("an identical patch must not claim a change: %+v err=%v", same, err)
	}
}

func TestMemEditRealDB_RefusesArchivedMissingMovedAndInvalid(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent, Visibility: "private"})
	if err != nil {
		t.Fatal(err)
	}
	authorized := memLookup(t, svc, saved.ArtifactID)
	for name, req := range map[string]EditRequest{
		"empty patch":   {ArtifactID: saved.ArtifactID, Authorized: authorized},
		"blank title":   {ArtifactID: saved.ArtifactID, Title: memStr("  "), Authorized: authorized},
		"blank content": {ArtifactID: saved.ArtifactID, Content: memStr(""), Authorized: authorized},
		"blank label":   {ArtifactID: saved.ArtifactID, SourceLabel: memStr(" "), Authorized: authorized},
		"no authority":  {ArtifactID: saved.ArtifactID, Title: memStr("x")},
	} {
		if _, err := svc.Edit(context.Background(), req); !errors.Is(err, ErrEditInvalid) {
			t.Fatalf("%s: want ErrEditInvalid, got %v", name, err)
		}
	}
	// The authority decision was made on a private entry; if it became org-wide
	// in between (promotion), the edit refuses instead of using a stale decision.
	if _, err := db.Exec(`UPDATE artifacts SET metadata = metadata || '{"visibility":"global"}' WHERE id::text = $1`, saved.ArtifactID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("moved"), Authorized: authorized}); !errors.Is(err, ErrEntryChanged) {
		t.Fatalf("moved entry: want ErrEntryChanged, got %v", err)
	}
	if _, err := db.Exec(`UPDATE artifacts SET metadata = metadata || '{"visibility":"private"}' WHERE id::text = $1`, saved.ArtifactID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Archive(context.Background(), saved.ArtifactID, "ada"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("archived edit"), Authorized: authorized}); !errors.Is(err, ErrEntryArchived) {
		t.Fatalf("archived: want ErrEntryArchived, got %v", err)
	}
	if record := memLookup(t, svc, saved.ArtifactID); record.Title != "Juniper & Rye Bakery" {
		t.Fatalf("refused edits changed the title: %q", record.Title)
	}
	if _, err := svc.Delete(context.Background(), saved.ArtifactID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("ghost"), Authorized: authorized}); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("deleted: want ErrEntryNotFound, got %v", err)
	}
}

// Concurrent edit and delete serialize on the artifact row: either the edit
// lands first and the delete removes everything, or the delete wins and the
// edit finds nothing. Never an orphan chunk, never a deadlock.
func TestMemEditRealDB_ConcurrentEditAndDeleteLeaveNoOrphans(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	long := strings.Repeat("Concurrent edit text about rye starters and oven schedules. ", 80)
	for round := 0; round < 12; round++ {
		saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Race entry", Content: bakeryContent, Visibility: "private"})
		if err != nil {
			t.Fatal(err)
		}
		authorized := memLookup(t, svc, saved.ArtifactID)
		var wg sync.WaitGroup
		var editErr, deleteErr error
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, editErr = svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Content: memStr(long), Authorized: authorized})
		}()
		go func() {
			defer wg.Done()
			_, deleteErr = svc.Delete(context.Background(), saved.ArtifactID)
		}()
		go func() { // a competing edit on the same entry
			defer wg.Done()
			_, _ = svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Title: memStr("Race title"), Authorized: authorized})
		}()
		wg.Wait()
		if editErr != nil && !errors.Is(editErr, ErrEntryNotFound) {
			t.Fatalf("round %d: edit must succeed or find nothing, got %v", round, editErr)
		}
		if deleteErr != nil && !errors.Is(deleteErr, ErrEntryNotFound) {
			t.Fatalf("round %d: delete failed: %v", round, deleteErr)
		}
		var left int
		_ = db.QueryRow(`SELECT (SELECT count(*) FROM artifacts WHERE id::text = $1) + (SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1)`,
			saved.ArtifactID).Scan(&left)
		if left != 0 {
			t.Fatalf("round %d: %d rows survived edit+delete", round, left)
		}
	}
}

func TestMemEditRealDB_ConcurrentEditsKeepChunksConsistent(t *testing.T) {
	db := openDeploymentContextTestDB(t)
	down := &fakeEmbedEngine{dims: 768}
	down.fail.Store(true)
	svc := serviceWith(db, routerWith(down))
	saved, err := svc.Ingest(context.Background(), IngestRequest{Title: "Race entry", Content: bakeryContent, Visibility: "private"})
	if err != nil {
		t.Fatal(err)
	}
	authorized := memLookup(t, svc, saved.ArtifactID)
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := strings.Repeat("Variant "+string(rune('A'+i))+" of the menu text. ", 20*(i+1))
			_, errs[i] = svc.Edit(context.Background(), EditRequest{ArtifactID: saved.ArtifactID, Content: memStr(body), Authorized: authorized})
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("edit %d: %v", i, err)
		}
	}
	var content string
	var chunkCount int
	_ = db.QueryRow(`SELECT content, (metadata->>'chunk_count')::int FROM artifacts WHERE id::text = $1`, saved.ArtifactID).Scan(&content, &chunkCount)
	contents, _, _, _ := memChunkRows(t, db, saved.ArtifactID)
	want := chunkText(content, defaultChunkSize, defaultChunkOverlap)
	if len(contents) != chunkCount || len(contents) != len(want) {
		t.Fatalf("chunks must match the last committed content: rows=%d meta=%d want=%d", len(contents), chunkCount, len(want))
	}
	for i := range want {
		if contents[i] != want[i] {
			t.Fatalf("chunk %d does not belong to the final content", i)
		}
	}
}
