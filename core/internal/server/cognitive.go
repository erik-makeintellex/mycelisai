package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/cognitive/status
// Returns health and configuration of all cognitive engines (vLLM text + Diffusers media).
func (s *AdminServer) HandleCognitiveStatus(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondJSON(w, map[string]any{"text": map[string]string{"status": "offline"}, "media": map[string]string{"status": "offline"}})
		return
	}

	type engineStatus struct {
		Status   string `json:"status"`
		Endpoint string `json:"endpoint,omitempty"`
		Model    string `json:"model,omitempty"`
	}

	result := map[string]*engineStatus{
		"text":  {Status: "offline"},
		"media": {Status: "offline"},
	}

	// Probe all openai_compatible text engines (vLLM, Ollama, LM Studio, etc.)
	cfg := s.Cognitive.Config
	for provID, prov := range cfg.Providers {
		if prov.Type != "openai_compatible" || prov.Endpoint == "" {
			continue
		}
		adapter, ok := s.Cognitive.Adapters[provID]
		if !ok {
			continue
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		alive, _ := adapter.Probe(ctx)
		cancel()
		if alive {
			result["text"] = &engineStatus{
				Status:   "online",
				Endpoint: prov.Endpoint,
				Model:    prov.ModelID,
			}
			break
		}
	}

	// Probe media engine
	if cfg.Media != nil && cfg.Media.Endpoint != "" {
		healthURL := cfg.Media.Endpoint[:len(cfg.Media.Endpoint)-3] + "/health" // strip /v1, add /health
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		resp, err := http.DefaultClient.Do(httpReq)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				result["media"] = &engineStatus{
					Status:   "online",
					Endpoint: cfg.Media.Endpoint,
					Model:    cfg.Media.ModelID,
				}
			}
		}
	}

	respondJSON(w, result)
}

// POST /api/v1/cognitive/infer
func (s *AdminServer) handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req cognitive.InferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.Cognitive.Infer(req)
	if err != nil {
		log.Printf("Inference Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp)
}

// GET /api/v1/cognitive/config
// Returns the current Cognitive Configuration (Profiles + Providers)
func (s *AdminServer) HandleCognitiveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	// Return the raw config struct
	respondJSON(w, s.Cognitive.Config)
}

// POST /api/v1/chat
// Routes user messages exclusively through the Admin agent via NATS request-reply.
// The Admin agent has its full system prompt, tools, and council access.
// No raw LLM fallback — if the swarm is offline, the endpoint returns an error.
func (s *AdminServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Vercel Request: { messages: [ { role, content }, ... ] }
	type VercelMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var req struct {
		Messages []VercelMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	lastMsg := req.Messages[len(req.Messages)-1]

	// 2. NATS must be available — the Admin agent is the ONLY path for chat.
	// No raw LLM fallback: agents must always operate within their context
	// (system prompt, tools, input/output rules).
	if s.NC == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Swarm offline — Admin agent unavailable. Start the organism first."}`, http.StatusServiceUnavailable)
		return
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	reqCtx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, []byte(lastMsg.Content))
	if err != nil {
		log.Printf("Chat via Admin agent failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"Admin agent did not respond: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(msg.Data)
}

// PUT /api/v1/cognitive/profiles
// Updates which provider each cognitive profile uses.
// Persists to cognitive.yaml and updates the in-memory config.
func (s *AdminServer) HandleUpdateProfiles(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Profiles map[string]string `json:"profiles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.Profiles) == 0 {
		http.Error(w, "No profiles provided", http.StatusBadRequest)
		return
	}

	// Validate: every profile value must reference an existing provider
	for profile, providerID := range req.Profiles {
		if _, ok := s.Cognitive.Config.Providers[providerID]; !ok {
			http.Error(w, fmt.Sprintf("Unknown provider '%s' for profile '%s'", providerID, profile), http.StatusBadRequest)
			return
		}
	}

	// Update in-memory config
	for profile, providerID := range req.Profiles {
		s.Cognitive.Config.Profiles[profile] = providerID
	}

	// Persist to YAML
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist cognitive config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	log.Printf("Cognitive profiles updated: %v", req.Profiles)
	respondJSON(w, s.Cognitive.Config)
}

// PUT /api/v1/cognitive/providers/{id}
// Updates a provider's configuration (endpoint, model_id, api_key_env).
// Reinitializes the adapter if the endpoint or type changes.
func (s *AdminServer) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	providerID := r.PathValue("id")
	if providerID == "" {
		http.Error(w, "Missing provider ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Endpoint   string `json:"endpoint,omitempty"`
		ModelID    string `json:"model_id,omitempty"`
		APIKey     string `json:"api_key,omitempty"`     // Direct key (stored in-memory only, not persisted to YAML)
		APIKeyEnv  string `json:"api_key_env,omitempty"` // Env var name (persisted to YAML)
		Type       string `json:"type,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	existing, ok := s.Cognitive.Config.Providers[providerID]
	if !ok {
		// Create new provider
		existing = cognitive.ProviderConfig{}
	}

	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Endpoint != "" {
		existing.Endpoint = req.Endpoint
	}
	if req.ModelID != "" {
		existing.ModelID = req.ModelID
	}
	if req.APIKeyEnv != "" {
		existing.AuthKeyEnv = req.APIKeyEnv
	}
	if req.APIKey != "" {
		existing.AuthKey = req.APIKey
	}

	s.Cognitive.Config.Providers[providerID] = existing

	// Persist to YAML (AuthKey/AuthKeyEnv are json:"-" so won't leak)
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist cognitive config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	log.Printf("Provider '%s' updated: endpoint=%s model=%s", providerID, existing.Endpoint, existing.ModelID)

	// Return sanitized provider info (no secrets)
	respondJSON(w, map[string]any{
		"id":       providerID,
		"type":     existing.Type,
		"endpoint": existing.Endpoint,
		"model_id": existing.ModelID,
		"configured": existing.AuthKey != "" || existing.AuthKeyEnv != "",
	})
}
