//go:build integration
// +build integration

package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// TestAgentInteraction_Sentry verifies that the Sentry profile (Boolean Schema)
// correctly interacts with the local model (Ollama) and returns exact bools.
func TestAgentInteraction_Sentry(t *testing.T) {
	// 1. Load Config
	router, err := cognitive.NewRouter("../config/cognitive.yaml", nil)
	if err != nil {
		t.Fatalf("Failed to load brain config: %v", err)
	}

	// 2. Prepare Request
	// We ask a simple question that expects a boolean
	req := cognitive.InferRequest{
		Profile: "sentry",
		Prompt:  "Is the sky blue? Answer ONLY with true or false.",
	}

	t.Logf("🤖 Sentry Agent: Probing Model [%s]...", req.Profile)

	// 3. Execute with Timeout (Sentry is fast)
	start := time.Now()
	resp, err := router.Infer(req)
	elapsed := time.Since(start)

	// 4. Assertions
	if err != nil {
		t.Fatalf("Inference Failed: %v", err)
	}

	t.Logf("✅ Response in %s: %s", elapsed, resp.Text)

	lower := strings.ToLower(resp.Text)
	if lower != "true" && lower != "false" {
		t.Errorf("Validation Failure: Expected 'true' or 'false', got '%s'", resp.Text)
	}
}

// TestAgentInteraction_Coder verifies that the Coder profile (Strict JSON)
// can generate valid JSON structures.
func TestAgentInteraction_Coder(t *testing.T) {
	// 1. Load Config
	router, err := cognitive.NewRouter("../config/cognitive.yaml", nil)
	if err != nil {
		t.Fatalf("Failed to load brain config: %v", err)
	}

	// 2. Prepare Request
	req := cognitive.InferRequest{
		Profile: "coder",
		Prompt:  `Generate a JSON object describing a user with fields: "username" (string) and "id" (int). Return ONLY JSON.`,
	}

	t.Logf("🤖 Coder Agent: Probing Model [%s]...", req.Profile)

	// 3. Execute
	start := time.Now()
	resp, err := router.Infer(req)
	elapsed := time.Since(start)

	// 4. Assertions
	if err != nil {
		t.Fatalf("Inference Failed: %v", err)
	}

	t.Logf("✅ Response in %s: %s", elapsed, resp.Text)

	cleanText := strings.TrimSpace(resp.Text)
	if strings.HasPrefix(cleanText, "```json") {
		cleanText = strings.TrimPrefix(cleanText, "```json")
		cleanText = strings.TrimSuffix(cleanText, "```")
		cleanText = strings.TrimSpace(cleanText)
	} else if strings.HasPrefix(cleanText, "```") {
		cleanText = strings.TrimPrefix(cleanText, "```")
		cleanText = strings.TrimSuffix(cleanText, "```")
		cleanText = strings.TrimSpace(cleanText)
	}

	if !strings.HasPrefix(cleanText, "{") {
		t.Errorf("Validation Failure: Does not start with '{'")
	}
}
