//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/provisioning"
)

func TestProvisioning_Draft(t *testing.T) {
	// 1. Setup (Load Brain)
	router, err := cognitive.NewRouter("../config/cognitive.yaml", nil)
	if err != nil {
		t.Fatalf("Failed to load brain: %v", err)
	}

	engine := provisioning.NewEngine(router)

	// 2. Draft Intent
	intent := "Create a database monitor that watches for slow queries and logs them."
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Architect is slow
	defer cancel()

	manifest, err := engine.Draft(ctx, intent)
	if err != nil {
		t.Fatalf("Provisioning Draft Failed: %v", err)
	}

	// 3. Verify
	t.Logf("Generated Manifest: %+v", manifest)

	if manifest.Name == "" {
		t.Error("Manifest Name is empty")
	}
	if len(manifest.Connectivity.Subscriptions) == 0 {
		t.Error("Manifest has no subscriptions")
	}
}
