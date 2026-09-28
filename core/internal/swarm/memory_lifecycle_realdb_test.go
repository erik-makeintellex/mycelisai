package swarm

import (
	"context"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
)

// A removed (archived or deleted) entry never reaches the model context and
// is never cited, even when the reply repeats its facts verbatim.
func TestSomaRecallRealDB_RemovedEntryIsNeverCited(t *testing.T) {
	for _, action := range []string{"archive", "delete"} {
		t.Run(action, func(t *testing.T) {
			db := openSwarmMemoryTestDB(t)
			provider := &retainedStackProvider{reply: "Blueberry-lavender scones, 3 for $10 this Saturday and Sunday only.\nSee you at the counter!"}
			router := retainedStackRouter(provider)
			seedBakeryAndScopedOutSources(t, db, router)
			svc := deploymentcontext.NewService(&artifacts.Service{DB: db}, memory.NewServiceWithDB(db), router)
			entries, err := svc.ListEntries(context.Background(), 20, false, memory.GovernedReader{})
			if err != nil {
				t.Fatal(err)
			}
			bakeryID := ""
			for _, entry := range entries {
				if entry.Title == bakeryTitle {
					bakeryID = entry.ArtifactID
				}
			}
			if bakeryID == "" {
				t.Fatal("bakery entry not saved")
			}
			if action == "archive" {
				err = svc.Archive(context.Background(), bakeryID, "admin")
			} else {
				_, err = svc.Delete(context.Background(), bakeryID)
			}
			if err != nil {
				t.Fatalf("%s: %v", action, err)
			}

			result := somaAgent(router, db).processMessageStructured(promoAsk, nil)
			if prompt := provider.allPrompts(); strings.Contains(prompt, "blueberry-lavender scones") || strings.Contains(prompt, bakeryTitle) {
				t.Fatalf("%s entry reached the model context", action)
			}
			for _, ref := range result.ContextSources {
				if ref.ArtifactID == bakeryID || ref.Title == bakeryTitle {
					t.Fatalf("%s entry was cited: %+v", action, result.ContextSources)
				}
			}
		})
	}
}
