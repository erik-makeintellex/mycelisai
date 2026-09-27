package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

// Real PostgreSQL proof. MYCELIS_MEMORY_TEST_DSN must point at a disposable
// database with 001_current_schema.sql installed. Skip is not proof.
func openMemoryTestDB(t *testing.T) *sql.DB {
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
	var table sql.NullString
	if err := db.QueryRow(`SELECT to_regclass('public.context_vectors')::text`).Scan(&table); err != nil || !table.Valid {
		t.Fatalf("DSN must have 001_current_schema.sql installed (context_vectors): %v", err)
	}
	return db
}

// seedLexicalRow inserts a pending (embedding NULL) governed chunk in an
// isolated tenant so parallel packages never see each other's rows.
func seedLexicalRow(t *testing.T, db *sql.DB, tenant, title, content string, meta map[string]any) {
	t.Helper()
	base := map[string]any{
		"tenant_id": tenant, "artifact_id": uuid.NewString(), "artifact_title": title,
		"knowledge_class": "company_knowledge", "type": "company_knowledge",
		"knowledge_store": "governed_context_store", "visibility": "global",
		"sensitivity_class": "role_scoped", "embedding_status": "pending",
	}
	for key, value := range meta {
		base[key] = value
	}
	raw, _ := json.Marshal(base)
	if _, err := db.Exec(`INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, NULL, $2)`, content, raw); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM context_vectors WHERE metadata->>'tenant_id' = $1`, tenant) })
}

const bakeryFacts = "Juniper & Rye Bakery. Weekend special: blueberry-lavender scones, 3 for $10, Saturday and Sunday. Sign-off line: See you at the counter."

func seedBakeryWithDistractors(t *testing.T, db *sql.DB, tenant string) {
	t.Helper()
	distractors := []string{
		"Our quarterly tax filing checklist for the accounting team.",
		"Write access to the shared drive is managed by IT.",
		"The line manager approves vacation requests for our staff.",
		"Two-factor authentication is required for every account.",
		"Our office is closed on public holidays.",
		"The special projects budget is reviewed each quarter.",
		"Deployment pipeline runs nightly for all services.",
		"Our supplier contract renews every January.",
		"Customer support hours are 9 to 5 on weekdays.",
		"Password rotation happens every ninety days.",
		"The team retrospective is held on Fridays.",
		"Write clear commit messages for every change.",
		"Inventory counts happen at month end.",
		"Our brand colors are forest green and cream.",
		"The delivery van is serviced twice a year.",
		"New hires complete safety training in week one.",
		"Our loyalty program gives a free coffee after ten visits.",
		"The espresso machine is descaled monthly.",
		"Parking is available behind the building.",
		"Payroll runs on the fifteenth and the last day of the month.",
	}
	for i, text := range distractors {
		seedLexicalRow(t, db, tenant, fmt.Sprintf("Operations note %d", i+1), text, nil)
	}
	seedLexicalRow(t, db, tenant, "Juniper & Rye Bakery", bakeryFacts, nil)
}

func TestRecallGovernedRealDB_KeywordRanksBakeryFirstWithoutEmbeddings(t *testing.T) {
	db := openMemoryTestDB(t)
	tenant := "tenant-" + uuid.NewString()
	seedBakeryWithDistractors(t, db, tenant)
	svc := NewServiceWithDB(db)

	results, mode, err := svc.RecallGoverned(context.Background(), nil, "Write a two-line promo for our weekend special", SemanticSearchOptions{Limit: 5, TenantID: tenant})
	if err != nil {
		t.Fatalf("RecallGoverned: %v", err)
	}
	if mode != RecallModeKeyword {
		t.Fatalf("mode = %q, want keyword", mode)
	}
	if len(results) == 0 || results[0].Metadata["artifact_title"] != "Juniper & Rye Bakery" {
		t.Fatalf("bakery chunk must rank first, got %+v", results)
	}
	if results[0].RetrievalMode != RecallModeKeyword || results[0].Score <= 0 {
		t.Fatalf("first hit must carry keyword mode and a positive rank: %+v", results[0])
	}
}

func TestRecallGovernedRealDB_StopWordOnlyQueryReturnsNothing(t *testing.T) {
	db := openMemoryTestDB(t)
	tenant := "tenant-" + uuid.NewString()
	seedBakeryWithDistractors(t, db, tenant)
	svc := NewServiceWithDB(db)
	for _, query := range []string{"our for the a", "and to of"} {
		results, _, err := svc.RecallGoverned(context.Background(), nil, query, SemanticSearchOptions{Limit: 5, TenantID: tenant})
		if err != nil || len(results) != 0 {
			t.Fatalf("stop-word query %q returned %d rows (err %v)", query, len(results), err)
		}
	}
	legacy, err := svc.TextSearchWithOptions(context.Background(), "our for the", SemanticSearchOptions{Limit: 5, TenantID: tenant})
	if err != nil || len(legacy) != 0 {
		t.Fatalf("TextSearchWithOptions stop words returned %d rows (err %v)", len(legacy), err)
	}
}

func TestRecallGovernedRealDB_ScopeNegatives(t *testing.T) {
	db := openMemoryTestDB(t)
	tenant := "tenant-" + uuid.NewString()
	svc := NewServiceWithDB(db)
	seedLexicalRow(t, db, tenant, "Open weekend special", "Weekend special for everyone.", nil)
	seedLexicalRow(t, db, tenant, "Team A weekend special", "Weekend special for team alpha only.",
		map[string]any{"visibility": "team", "team_id": "team-a"})
	seedLexicalRow(t, db, tenant, "Restricted weekend special", "Weekend special restricted pricing.",
		map[string]any{"sensitivity_class": "restricted", "team_id": "team-a"})
	seedLexicalRow(t, db, tenant, "Goal weekend special", "Weekend special for the launch goal.",
		map[string]any{"target_goal_sets": []string{"spring-launch"}})
	seedLexicalRow(t, db, tenant, "Operating weekend special", "Weekend special operating guidance.",
		map[string]any{"knowledge_class": "soma_operating_context", "type": "soma_operating_context"})

	titles := func(opts SemanticSearchOptions) map[string]bool {
		t.Helper()
		opts.Limit, opts.TenantID = 20, tenant
		results, _, err := svc.RecallGoverned(context.Background(), nil, "weekend special", opts)
		if err != nil {
			t.Fatalf("RecallGoverned: %v", err)
		}
		out := map[string]bool{}
		for _, result := range results {
			out[fmt.Sprint(result.Metadata["artifact_title"])] = true
		}
		return out
	}

	teamB := titles(SemanticSearchOptions{TeamID: "team-b", AgentID: "worker-b", AllowGlobal: true,
		Types: []string{"customer_context", "company_knowledge"}, ExcludeSensitivity: []string{"restricted"}})
	if !teamB["Open weekend special"] {
		t.Fatalf("team B must see the open row: %v", teamB)
	}
	for _, hidden := range []string{"Team A weekend special", "Restricted weekend special", "Goal weekend special", "Operating weekend special"} {
		if teamB[hidden] {
			t.Fatalf("team B worker received scoped-out row %q: %v", hidden, teamB)
		}
	}
	teamA := titles(SemanticSearchOptions{TeamID: "team-a", AgentID: "worker-a", AllowGlobal: true,
		Types: []string{"customer_context", "company_knowledge"}, ExcludeSensitivity: []string{"restricted"}})
	if !teamA["Team A weekend special"] || !teamA["Restricted weekend special"] {
		t.Fatalf("team A must see its own team and restricted rows: %v", teamA)
	}
	withGoal := titles(SemanticSearchOptions{GoalSets: []string{"spring-launch"}})
	if !withGoal["Goal weekend special"] {
		t.Fatalf("intersecting goal set must include the goal-scoped row: %v", withGoal)
	}
	otherGoal := titles(SemanticSearchOptions{GoalSets: []string{"autumn"}, ExcludeClasses: []string{"soma_operating_context"}})
	if otherGoal["Goal weekend special"] || otherGoal["Operating weekend special"] {
		t.Fatalf("non-intersecting goal set or excluded class leaked: %v", otherGoal)
	}
}

type countingEmbedder struct {
	available bool
	calls     atomic.Int32
	vec       []float64
}

func (c *countingEmbedder) EmbeddingAvailable(context.Context) bool { return c.available }
func (c *countingEmbedder) Embed(context.Context, string, string) ([]float64, error) {
	c.calls.Add(1)
	if !c.available {
		return nil, errors.New("embedding failed")
	}
	return c.vec, nil
}

func TestRecallGovernedRealDB_UnavailableEmbedderIsNeverCalled(t *testing.T) {
	db := openMemoryTestDB(t)
	tenant := "tenant-" + uuid.NewString()
	seedBakeryWithDistractors(t, db, tenant)
	embedder := &countingEmbedder{available: false}
	results, mode, err := NewServiceWithDB(db).RecallGoverned(context.Background(), embedder, "weekend special promo", SemanticSearchOptions{Limit: 3, TenantID: tenant})
	if err != nil || mode != RecallModeKeyword || len(results) == 0 {
		t.Fatalf("keyword recall failed: mode=%q n=%d err=%v", mode, len(results), err)
	}
	if embedder.calls.Load() != 0 {
		t.Fatalf("unavailable embedder was called %d times", embedder.calls.Load())
	}
}

func TestRecallGovernedRealDB_HybridMergesPendingRowsWhenEmbeddingsAvailable(t *testing.T) {
	db := openMemoryTestDB(t)
	tenant := "tenant-" + uuid.NewString()
	seedBakeryWithDistractors(t, db, tenant)
	vec := make([]float64, 768)
	vec[0] = 1
	embedded, _ := json.Marshal(map[string]any{"tenant_id": tenant, "artifact_id": uuid.NewString(), "artifact_title": "Embedded menu",
		"knowledge_class": "company_knowledge", "type": "company_knowledge", "visibility": "global", "embedding_status": "embedded"})
	if _, err := db.Exec(`INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, $2::vector, $3)`, "Menu board layout", formatVector(vec), embedded); err != nil {
		t.Fatal(err)
	}
	embedder := &countingEmbedder{available: true, vec: vec}
	results, mode, err := NewServiceWithDB(db).RecallGoverned(context.Background(), embedder, "weekend special", SemanticSearchOptions{Limit: 5, TenantID: tenant})
	if err != nil || mode != RecallModeHybrid {
		t.Fatalf("mode=%q err=%v", mode, err)
	}
	seen := map[string]string{}
	for _, result := range results {
		seen[fmt.Sprint(result.Metadata["artifact_title"])] = result.RetrievalMode
	}
	if seen["Embedded menu"] != RecallModeSemantic || seen["Juniper & Rye Bakery"] != RecallModeKeyword {
		t.Fatalf("hybrid must merge semantic embedded rows with keyword pending rows: %v", seen)
	}
}
