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

// mutationTools are tools that trigger proposal mode when used in chat.
// If an agent uses any of these, the response switches from answer → proposal.
var mutationTools = map[string]bool{
	"generate_blueprint": true,
	"delegate":           true,
	"write_file":         true,
	"publish_signal":     true,
	"broadcast":          true,
}

// hasMutationTools checks if any tools in the list are mutation tools.
func hasMutationTools(tools []string) (bool, []string) {
	var mutations []string
	for _, t := range tools {
		if mutationTools[t] {
			mutations = append(mutations, t)
		}
	}
	return len(mutations) > 0, mutations
}

// chatToolRisk estimates risk level from mutation tools used.
func chatToolRisk(tools []string) string {
	for _, t := range tools {
		if t == "generate_blueprint" || t == "delegate" || t == "broadcast" {
			return "medium"
		}
	}
	return "low"
}

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
//
// The full conversation history is forwarded as JSON so the admin agent can
// maintain multi-turn context. The NATS payload is a JSON array of
// {role, content} objects; the agent's handleDirectRequest detects JSON arrays
// and reconstructs prior turns.
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

	// 2. NATS must be available — the Admin agent is the ONLY path for chat.
	// No raw LLM fallback: agents must always operate within their context
	// (system prompt, tools, input/output rules).
	if s.NC == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Swarm offline — Admin agent unavailable. Start the organism first."}`, http.StatusServiceUnavailable)
		return
	}

	// 3. Forward FULL conversation history to admin agent via NATS.
	// Serialize the entire messages array so the agent can reconstruct context.
	payload, err := json.Marshal(req.Messages)
	if err != nil {
		http.Error(w, "Failed to serialize messages", http.StatusInternalServerError)
		return
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	reqCtx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, payload)
	if err != nil {
		log.Printf("Chat via Admin agent failed: %v", err)
		respondError(w, "Admin agent did not respond: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Agent returns structured JSON (ProcessResult with text, tools_used, artifacts, brain info).
	// Wrap in CTS envelope so artifacts flow to the frontend.
	var agentResult struct {
		Text       string                     `json:"text"`
		ToolsUsed  []string                   `json:"tools_used,omitempty"`
		Artifacts  []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
		ProviderID string                     `json:"provider_id,omitempty"`
		ModelUsed  string                     `json:"model_used,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &agentResult); err != nil || agentResult.Text == "" {
		agentResult.Text = string(msg.Data)
		agentResult.ToolsUsed = nil
		agentResult.Artifacts = nil
	}

	chatPayload := protocol.ChatResponsePayload{
		Text:      agentResult.Text,
		ToolsUsed: agentResult.ToolsUsed,
		Artifacts: agentResult.Artifacts,
	}

	// Phase 19: Build brain provenance from agent's inference metadata
	if agentResult.ProviderID != "" && s.Cognitive != nil {
		brain := &protocol.BrainProvenance{
			ProviderID: agentResult.ProviderID,
			ModelID:    agentResult.ModelUsed,
		}
		if s.Cognitive.Config != nil {
			if pCfg, ok := s.Cognitive.Config.Providers[agentResult.ProviderID]; ok {
				brain.ProviderName = agentResult.ProviderID
				brain.Location = pCfg.Location
				brain.DataBoundary = pCfg.DataBoundary
				if brain.Location == "" {
					brain.Location = "local"
				}
				if brain.DataBoundary == "" {
					brain.DataBoundary = "local_only"
				}
			}
		}
		chatPayload.Brain = brain
	}

	// Phase 19-B: Detect mutation tools → switch to proposal mode
	isMutation, mutTools := hasMutationTools(agentResult.ToolsUsed)

	var templateID protocol.TemplateID
	var mode protocol.ExecutionMode

	if isMutation {
		templateID = protocol.TemplateChatToProposal
		mode = protocol.ModeProposal

		// Build scope from mutation tools
		scope := &protocol.ScopeValidation{
			Tools:             mutTools,
			AffectedResources: []string{"state"},
			RiskLevel:         chatToolRisk(mutTools),
		}

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToProposal, "admin",
			"Chat mutation detected",
			map[string]any{"tools": agentResult.ToolsUsed, "mutations": mutTools},
		)

		proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, "chat-action", scope, auditEventID)
		var confirmToken *protocol.ConfirmToken
		if proof != nil {
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
		}

		chatPayload.Proposal = &protocol.ChatProposal{
			Intent:    "chat-action",
			Tools:     mutTools,
			RiskLevel: chatToolRisk(mutTools),
		}
		if proof != nil {
			chatPayload.Proposal.IntentProofID = proof.ID
		}
		if confirmToken != nil {
			chatPayload.Proposal.ConfirmToken = confirmToken.Token
		}

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
	} else {
		templateID = protocol.TemplateChatToAnswer
		mode = protocol.ModeAnswer

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, "admin",
			"Admin chat",
			map[string]any{"tools": agentResult.ToolsUsed},
		)
		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
	}

	payloadBytes, _ := json.Marshal(chatPayload)

	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "admin",
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

// ---------------------------------------------------------------------------
// Council Chat API — standardized, CTS-enveloped council interaction
// ---------------------------------------------------------------------------

// CouncilMemberInfo is returned by HandleListCouncilMembers.
type CouncilMemberInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Team string `json:"team"`
}

// respondAPIJSON writes a protocol.APIResponse as JSON with an explicit status code.
func respondAPIJSON(w http.ResponseWriter, status int, resp protocol.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// respondAPIError writes a structured APIResponse error.
func respondAPIError(w http.ResponseWriter, msg string, status int) {
	respondAPIJSON(w, status, protocol.NewAPIError(msg))
}

// isCouncilMember checks whether memberID belongs to a standing council team
// (admin-core or council-core). Returns the team ID and role on match.
// Dynamic: add a new member to the YAML, restart, done.
func (s *AdminServer) isCouncilMember(memberID string) (teamID string, role string, ok bool) {
	if s.Soma == nil {
		return "", "", false
	}
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			if m.ID == memberID {
				return tm.ID, m.Role, true
			}
		}
	}
	return "", "", false
}

// GET /api/v1/council/members
// Returns all addressable council members from standing teams.
func (s *AdminServer) HandleListCouncilMembers(w http.ResponseWriter, r *http.Request) {
	if s.Soma == nil {
		respondAPIError(w, "Swarm offline", http.StatusServiceUnavailable)
		return
	}

	var members []CouncilMemberInfo
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			members = append(members, CouncilMemberInfo{
				ID:   m.ID,
				Role: m.Role,
				Team: tm.ID,
			})
		}
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(members))
}

// POST /api/v1/council/{member}/chat
// Routes user conversation to a specific council member via NATS request-reply.
// Returns a CTS envelope wrapped in APIResponse with trust score and provenance.
func (s *AdminServer) HandleCouncilChat(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("member")
	if memberID == "" {
		respondAPIError(w, "Missing council member ID", http.StatusBadRequest)
		return
	}

	// Validate member exists in standing council teams
	teamID, _, ok := s.isCouncilMember(memberID)
	if !ok {
		respondAPIError(w, fmt.Sprintf("Unknown council member: %s", memberID), http.StatusNotFound)
		return
	}

	// Parse Vercel-format messages: { messages: [{role, content}] }
	type VercelMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var req struct {
		Messages []VercelMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		respondAPIError(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	// NATS must be available
	if s.NC == nil {
		respondAPIError(w, "Swarm offline — council agents unavailable. Start the organism first.", http.StatusServiceUnavailable)
		return
	}

	// Forward conversation to the council member's personal NATS topic
	payload, err := json.Marshal(req.Messages)
	if err != nil {
		respondAPIError(w, "Failed to serialize messages", http.StatusInternalServerError)
		return
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, memberID)
	reqCtx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, payload)
	if err != nil {
		log.Printf("Council chat with %s failed: %v", memberID, err)
		respondAPIError(w, fmt.Sprintf("Council member %s did not respond: %s", memberID, err.Error()), http.StatusBadGateway)
		return
	}

	// Parse structured agent response (ProcessResult JSON with text, tools_used, artifacts, brain info).
	// Falls back gracefully to raw text if the agent returns plain text.
	var agentResult struct {
		Text       string                     `json:"text"`
		ToolsUsed  []string                   `json:"tools_used,omitempty"`
		Artifacts  []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
		ProviderID string                     `json:"provider_id,omitempty"`
		ModelUsed  string                     `json:"model_used,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &agentResult); err != nil || agentResult.Text == "" {
		// Fallback: treat entire response as plain text
		agentResult.Text = string(msg.Data)
		agentResult.ToolsUsed = nil
		agentResult.Artifacts = nil
	}

	// Wrap response in CTS envelope with trust score, provenance, and tool metadata
	chatPayload := protocol.ChatResponsePayload{
		Text:      agentResult.Text,
		ToolsUsed: agentResult.ToolsUsed,
		Artifacts: agentResult.Artifacts,
	}

	// Phase 19: Build brain provenance from agent's inference metadata
	if agentResult.ProviderID != "" && s.Cognitive != nil {
		brain := &protocol.BrainProvenance{
			ProviderID: agentResult.ProviderID,
			ModelID:    agentResult.ModelUsed,
		}
		// Enrich with provider config metadata (location, data boundary)
		if s.Cognitive.Config != nil {
			if pCfg, ok := s.Cognitive.Config.Providers[agentResult.ProviderID]; ok {
				brain.ProviderName = agentResult.ProviderID // Use ID as display name
				brain.Location = pCfg.Location
				brain.DataBoundary = pCfg.DataBoundary
				if brain.Location == "" {
					brain.Location = "local" // default to local
				}
				if brain.DataBoundary == "" {
					brain.DataBoundary = "local_only"
				}
			}
		}
		chatPayload.Brain = brain
	}

	// Phase 19-B: Detect mutation tools → switch to proposal mode
	isMutation, mutTools := hasMutationTools(agentResult.ToolsUsed)

	var templateID protocol.TemplateID
	var mode protocol.ExecutionMode

	if isMutation {
		templateID = protocol.TemplateChatToProposal
		mode = protocol.ModeProposal

		// Build scope from mutation tools
		scope := &protocol.ScopeValidation{
			Tools:             mutTools,
			AffectedResources: []string{"state"},
			RiskLevel:         chatToolRisk(mutTools),
		}

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToProposal, memberID,
			fmt.Sprintf("Council chat mutation detected from %s", memberID),
			map[string]any{"tools": agentResult.ToolsUsed, "mutations": mutTools, "member": memberID, "team": teamID},
		)

		proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, "chat-action", scope, auditEventID)
		var confirmToken *protocol.ConfirmToken
		if proof != nil {
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
		}

		chatPayload.Proposal = &protocol.ChatProposal{
			Intent:    "chat-action",
			Tools:     mutTools,
			RiskLevel: chatToolRisk(mutTools),
		}
		if proof != nil {
			chatPayload.Proposal.IntentProofID = proof.ID
		}
		if confirmToken != nil {
			chatPayload.Proposal.ConfirmToken = confirmToken.Token
		}

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
	} else {
		templateID = protocol.TemplateChatToAnswer
		mode = protocol.ModeAnswer

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, memberID,
			fmt.Sprintf("Council chat with %s", memberID),
			map[string]any{"tools": agentResult.ToolsUsed, "member": memberID, "team": teamID},
		)
		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
	}

	payloadBytes, _ := json.Marshal(chatPayload)

	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: memberID,
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
	log.Printf("Council chat: member=%s team=%s trust=%.1f tools=%v template=%s", memberID, teamID, envelope.TrustScore, agentResult.ToolsUsed, envelope.TemplateID)
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
