package main

import "testing"

// This package-level integration probe is intentionally skipped in normal CI.
// Use core/probe.go for manual local Ollama connectivity checks.
func TestLocalOllamaProbe_ManualOnly(t *testing.T) {
	t.Skip("manual integration probe; run core/probe.go directly when needed")
}
