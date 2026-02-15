package server

import (
	"encoding/json"
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

// HandleTrustThreshold reads or updates the Overseer's AutoExecuteThreshold.
// GET  /api/v1/trust/threshold — returns current threshold
// PUT  /api/v1/trust/threshold — updates threshold (0.0–1.0)
func (s *AdminServer) HandleTrustThreshold(w http.ResponseWriter, r *http.Request) {
	if s.Overseer == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Overseer not initialized — trust economy offline"}`, http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case "GET":
		respondJSON(w, map[string]float64{
			"threshold": s.Overseer.GetAutoExecuteThreshold(),
		})
	case "PUT":
		var body struct {
			Threshold float64 `json:"threshold"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		if body.Threshold < 0 || body.Threshold > 1 {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"threshold must be between 0.0 and 1.0"}`, http.StatusBadRequest)
			return
		}
		s.Overseer.SetAutoExecuteThreshold(body.Threshold)
		respondJSON(w, map[string]string{"status": "updated"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
