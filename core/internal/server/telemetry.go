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
// GET  /api/v1/trust/threshold — returns current threshold (signed-in users)
// PUT  /api/v1/trust/threshold — updates threshold (0.0–1.0); root admin +
// governance:write, audited first; the threshold decides what auto-executes,
// so it is governance policy (AUTH-C1 A4). Denials happen before any read.
func (s *AdminServer) HandleTrustThreshold(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		if _, ok := requireRootAdminScope(w, r, scopeGovernanceWrite); !ok {
			return
		}
	}
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
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		if body.Threshold < 0 || body.Threshold > 1 {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"threshold must be between 0.0 and 1.0"}`, http.StatusBadRequest)
			return
		}
		auditCtx := map[string]any{"action": "trust_threshold_update", "previous_threshold": s.Overseer.GetAutoExecuteThreshold(),
			"new_threshold": body.Threshold, "result_status": "requested"}
		auditID := s.auditGovernance(r, "trust-threshold", "Trust threshold update requested", auditCtx)
		if auditID == "" {
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable,
				"The audit record could not be written, so the threshold was not changed.", nil)
			return
		}
		s.Overseer.SetAutoExecuteThreshold(body.Threshold)
		auditCtx["result_status"] = "applied"
		s.auditGovernance(r, "trust-threshold", "Trust threshold update applied", auditCtx)
		respondJSON(w, map[string]string{"status": "updated", "audit_id": auditID})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
