package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/governance"
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

// F16 comms-inbound authority. Admins hold scopeCommsInbound through "*".
const (
	scopeCommsInbound     = "comms:inbound"
	commsAllowUserLaneEnv = "MYCELIS_COMMS_ALLOW_USER_LANE"
	commsInboundAuditSrc  = "comms-inbound"
	commsInboundIngress   = "comms_inbound"
)

// commsReservedLanes are global-input lanes owned by Core, never by a comms
// provider: Genesis's CLI lane and the telemetry sensor lanes. `user` is
// handled separately behind commsAllowUserLaneEnv.
var commsReservedLanes = map[string]struct{}{"cli": {}, "sensor": {}}

type commsInboundRequest struct {
	Sender   string         `json:"sender"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

// POST /api/v1/comms/inbound/{provider}
// Admin-only (F16). Accepts one external message for an allowed comms
// provider, wraps it in a signal envelope that names it as external input,
// audits it (fail closed), then publishes it on swarm.global.input.<provider>.
// Consumers treat it as data: teams receive it planning-only.
func (s *AdminServer) HandleCommsInbound(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, scopeCommsInbound)
	if !ok {
		return
	}
	if s.NC == nil {
		respondAPIError(w, "NATS connection offline", http.StatusServiceUnavailable)
		return
	}
	provider := r.PathValue("provider")
	subject := fmt.Sprintf(protocol.TopicGlobalInputFmt, provider)
	if reason := s.commsInboundProviderRejection(provider, subject); reason != "" {
		respondBlockerText(w, r, http.StatusBadRequest, codeCommsProviderRejected, commsProviderRejectedCopy, reason, nil)
		return
	}

	rawBody, err := io.ReadAll(io.LimitReader(r.Body, governance.MaxIngressBytes+1))
	if err != nil || len(rawBody) > governance.MaxIngressBytes {
		respondAPIError(w, "message body is unreadable or larger than 1 MiB", http.StatusBadRequest)
		return
	}
	var req commsInboundRequest
	if json.Unmarshal(rawBody, &req) != nil {
		req = commsInboundRequest{Message: string(rawBody)} // plain-text body
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		respondAPIError(w, "message is required", http.StatusBadRequest)
		return
	}

	data, err := commsInboundEnvelope(provider, subject, req, identity)
	if err != nil || len(data) > governance.MaxIngressBytes {
		respondAPIError(w, "message could not be wrapped for delivery", http.StatusBadRequest)
		return
	}
	sum := sha256.Sum256(data)
	record := map[string]any{
		"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "comms_inbound_accepted",
		"authority": scopeCommsInbound, "provider": provider, "subject": subject,
		"payload_sha256": hex.EncodeToString(sum[:]), "payload_bytes": len(data),
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, commsInboundAuditSrc, "Comms inbound message accepted: "+provider, attachActorIdentity(record, r))
	if err != nil || strings.TrimSpace(auditID) == "" {
		if err != nil {
			log.Printf("F16: comms inbound audit not recorded: %v", err)
		}
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "Comms inbound audit event could not be recorded; nothing was published", nil)
		return
	}
	if err := s.NC.Publish(subject, data); err != nil {
		record["action"], record["audit_event_id"] = "comms_inbound_publish_failed", auditID
		if _, auditErr := s.createAuditEvent(protocol.TemplateChatToProposal, commsInboundAuditSrc, "Comms inbound publish failed: "+provider, attachActorIdentity(record, r)); auditErr != nil {
			log.Printf("F16: comms inbound failure record not written (best-effort): %v", auditErr)
		}
		respondAPIError(w, "Failed to publish inbound message: "+err.Error(), http.StatusBadGateway)
		return
	}
	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
		"provider": provider, "subject": subject, "status": "queued", "audit_id": auditID,
	}))
}

// commsInboundProviderRejection returns "" when provider may publish, else an
// admin-readable reason. Checks, in order: token shape, the user switch,
// reserved lanes, registered input sources, the comms provider allowlist.
func (s *AdminServer) commsInboundProviderRejection(provider, subject string) string {
	if !governance.ValidIngressProviderToken(provider) {
		return "provider must match ^[a-z0-9_-]{1,32}$"
	}
	if provider == "user" {
		if on, err := strconv.ParseBool(os.Getenv(commsAllowUserLaneEnv)); err != nil || !on {
			return "the user lane is disabled (" + commsAllowUserLaneEnv + " is not true)"
		}
		return ""
	}
	if _, reserved := commsReservedLanes[provider]; reserved {
		return "provider names a reserved Core lane"
	}
	for _, source := range s.Inputs.List() {
		if source.ID == provider || source.AllowedIngressSubject == subject {
			return "provider is owned by registered input source " + source.ID
		}
	}
	if s.Comms != nil {
		for _, info := range s.Comms.ListProviders() {
			if info.Name == provider {
				return ""
			}
		}
	}
	return "provider is not a registered comms provider"
}

// commsInboundEnvelope wraps the message as external event input. Caller
// fields become data inside payload; nothing from the body reaches the top
// level, so it can never parse as a TeamAsk or a command envelope.
func commsInboundEnvelope(provider, subject string, req commsInboundRequest, identity *RequestIdentity) ([]byte, error) {
	payload, err := json.Marshal(map[string]any{
		"ingress": commsInboundIngress, "provider": provider, "sender": strings.TrimSpace(req.Sender),
		"message": req.Message, "metadata": req.Metadata,
		"principal": map[string]any{
			"user_id": identity.UserID, "username": identity.Username, "role": identity.Role,
			"principal_type": identity.PrincipalType, "auth_source": identity.AuthSource,
		},
	})
	if err != nil {
		return nil, err
	}
	return protocol.WrapSignalPayloadWithMeta(protocol.SourceKindWebAPI, subject, protocol.PayloadKindEvent, "", "", "", payload)
}
