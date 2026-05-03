package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// POST /api/v1/intent/negotiate
func (s *AdminServer) handleIntentNegotiate(w http.ResponseWriter, r *http.Request) {
	if s.MetaArchitect == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Cognitive engine offline — MetaArchitect not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Intent string `json:"intent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.Intent == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"intent is required"}`, http.StatusBadRequest)
		return
	}

	var blueprint *protocol.MissionBlueprint
	if s.NC != nil {
		bp, err := s.negotiateViaAdmin(r.Context(), req.Intent)
		if err == nil && bp != nil {
			blueprint = bp
		} else {
			log.Printf("Admin-routed negotiate failed, falling back to direct MetaArchitect: %v", err)
		}
	}

	if blueprint == nil {
		bp, err := s.MetaArchitect.GenerateBlueprint(r.Context(), req.Intent)
		if err != nil {
			log.Printf("Intent negotiation failed: %v", err)
			respondError(w, "Cognitive engine error: "+err.Error(), http.StatusBadGateway)
			return
		}
		blueprint = bp
	}

	scope := buildScopeFromBlueprint(blueprint)
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "negotiate",
		fmt.Sprintf("Blueprint negotiation: %s", req.Intent),
		map[string]any{"intent": req.Intent, "teams": len(blueprint.Teams), "scope": scope},
	)

	proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, req.Intent, scope, auditEventID)
	var confirmToken *protocol.ConfirmToken
	if proof != nil {
		confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
	}

	templateSpec := protocol.TemplateRegistry[protocol.TemplateChatToProposal]
	respondJSON(w, protocol.NegotiateResponse{
		Blueprint:    blueprint,
		IntentProof:  proof,
		ConfirmToken: confirmToken,
		Template:     &templateSpec,
	})
}

func (s *AdminServer) negotiateViaAdmin(ctx context.Context, intent string) (*protocol.MissionBlueprint, error) {
	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	directive := fmt.Sprintf(
		"The user wants to negotiate a mission blueprint. Their intent is:\n\n%s\n\n"+
			"Use research_for_blueprint to gather context first, then use generate_blueprint "+
			"to create the mission blueprint. Return ONLY the blueprint JSON — no commentary.",
		intent,
	)

	reqCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, []byte(directive))
	if err != nil {
		return nil, fmt.Errorf("admin agent did not respond: %w", err)
	}
	return extractBlueprintFromResponse(string(msg.Data))
}

func extractBlueprintFromResponse(response string) (*protocol.MissionBlueprint, error) {
	text := response
	if idx := strings.Index(text, "```json"); idx >= 0 {
		text = text[idx+7:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	} else if idx := strings.Index(text, "```"); idx >= 0 {
		text = text[idx+3:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	}

	text = strings.TrimSpace(text)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("no JSON object found in admin response")
	}
	text = text[start : end+1]

	var bp protocol.MissionBlueprint
	if err := json.Unmarshal([]byte(text), &bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint JSON: %w", err)
	}
	if bp.MissionID == "" || len(bp.Teams) == 0 {
		return nil, fmt.Errorf("invalid blueprint: missing mission_id or teams")
	}
	return &bp, nil
}
