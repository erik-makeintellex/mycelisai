package server

import (
	"net/http"
	"runtime"
	"time"
)

// TelemetrySnapshot is the JSON response for GET /api/v1/telemetry/compute.
// It provides real-time system metrics for the Mission Control Panopticon.
type TelemetrySnapshot struct {
	Goroutines   int     `json:"goroutines"`
	HeapAllocMB  float64 `json:"heap_alloc_mb"`
	SysMemMB     float64 `json:"sys_mem_mb"`
	LLMTokensSec float64 `json:"llm_tokens_sec"`
	Timestamp    string  `json:"timestamp"`
}

// HandleTelemetry returns a real-time compute telemetry snapshot.
// GET /api/v1/telemetry/compute
func (s *AdminServer) HandleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var tokenRate float64
	if s.Cognitive != nil {
		tokenRate = s.Cognitive.TokenRate()
	}

	snap := TelemetrySnapshot{
		Goroutines:   runtime.NumGoroutine(),
		HeapAllocMB:  float64(memStats.HeapAlloc) / 1024 / 1024,
		SysMemMB:     float64(memStats.Sys) / 1024 / 1024,
		LLMTokensSec: tokenRate,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	respondJSON(w, snap)
}
