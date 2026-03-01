package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/comms/providers
func (s *AdminServer) HandleCommsProviders(w http.ResponseWriter, r *http.Request) {
	if s.Comms == nil {
		respondAPIError(w, "Communications gateway offline", http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(s.Comms.ListProviders()))
}

// POST /api/v1/comms/send
// { "provider":"whatsapp|telegram|slack|webhook", "recipient":"...", "message":"...", "metadata":{...} }
func (s *AdminServer) HandleCommsSend(w http.ResponseWriter, r *http.Request) {
	if s.Comms == nil {
		respondAPIError(w, "Communications gateway offline", http.StatusServiceUnavailable)
		return
	}

	var req comms.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.Message) == "" {
		respondAPIError(w, "provider and message are required", http.StatusBadRequest)
		return
	}

	res, err := s.Comms.Send(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "failed") || strings.Contains(err.Error(), "returned") {
			status = http.StatusBadGateway
		}
		respondAPIError(w, "Failed to send communication: "+err.Error(), status)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(res))
}

// POST /api/v1/comms/inbound/{provider}
// Accepts external webhook messages and injects them into Soma global input bus.
func (s *AdminServer) HandleCommsInbound(w http.ResponseWriter, r *http.Request) {
	if s.NC == nil {
		respondAPIError(w, "NATS connection offline", http.StatusServiceUnavailable)
		return
	}

	provider := strings.TrimSpace(strings.ToLower(r.PathValue("provider")))
	if provider == "" {
		respondAPIError(w, "provider is required", http.StatusBadRequest)
		return
	}

	rawBody, _ := io.ReadAll(r.Body)
	var req struct {
		Sender   string         `json:"sender"`
		Message  string         `json:"message"`
		Metadata map[string]any `json:"metadata"`
	}
	_ = json.Unmarshal(rawBody, &req)

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(rawBody))
	}
	if msg == "" {
		respondAPIError(w, "message is required", http.StatusBadRequest)
		return
	}

	subject := fmt.Sprintf("swarm.global.input.%s", provider)
	payload := msg
	if req.Sender != "" {
		payload = fmt.Sprintf("[%s:%s] %s", provider, req.Sender, msg)
	}

	if err := s.NC.Publish(subject, []byte(payload)); err != nil {
		respondAPIError(w, "Failed to publish inbound message: "+err.Error(), http.StatusBadGateway)
		return
	}

	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
		"provider": provider,
		"subject":  subject,
		"status":   "queued",
	}))
}
