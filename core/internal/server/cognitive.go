package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
)

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
// Adapts Vercel AI SDK "messages" format to "InferRequest"
func (s *AdminServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
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

	// 2. Extract Prompt (Last User Message)
	// TODO: Context window management
	lastMsg := req.Messages[len(req.Messages)-1]

	// 3. Call Infer
	inferReq := cognitive.InferRequest{
		Profile: "default",
		Prompt:  lastMsg.Content,
	}

	// 4. Respond (Streaming Mimic)
	// We just write the full text for now.
	// Vercel SDK can handle non-streamed responses if headers are correct?
	// Actually we should write it as a stream of text.

	resp, err := s.Cognitive.Infer(inferReq)
	if err != nil {
		log.Printf("Chat Inference Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for SSE/Text stream compatibility if needed,
	// or just text/plain for simple useChat consumption.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(resp.Text))
}
