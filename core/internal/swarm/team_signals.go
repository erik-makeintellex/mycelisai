package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// handleTrigger receives an external signal and broadens it to the internal team bus.
func (t *Team) handleTrigger(msg *nats.Msg) {
	log.Printf("Team [%s] Triggered by [%s]", t.Manifest.Name, msg.Subject)
	internalSubject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, t.Manifest.ID)
	payload := normalizeCommandPayload(msg.Data)
	if correlation := extractTeamCommandCorrelation(t.Manifest.ID, msg.Data, payload); correlation != nil {
		t.rememberCommandCorrelation(*correlation)
	}
	t.nc.Publish(internalSubject, payload)
}

func normalizeCommandPayload(data []byte) []byte {
	var env protocol.SignalEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return data
	}
	if env.Meta.PayloadKind != protocol.PayloadKindCommand {
		return data
	}
	if strings.TrimSpace(env.Text) != "" {
		return []byte(env.Text)
	}
	if len(env.Payload) == 0 {
		return data
	}

	trimmed := bytes.TrimSpace(env.Payload)
	var ask protocol.TeamAsk
	if err := json.Unmarshal(trimmed, &ask); err == nil && !ask.IsZero() {
		return append([]byte(nil), trimmed...)
	}
	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		return []byte(asString)
	}
	return append([]byte(nil), trimmed...)
}

// handleResponse receives an internal signal and broadens it to the external team bus.
func (t *Team) handleResponse(msg *nats.Msg) {
	log.Printf("Team [%s] Response: %s", t.Manifest.Name, string(msg.Data))
	correlation := t.responseCommandCorrelation(msg.Data)
	for _, subject := range t.Manifest.Deliveries {
		payload := msg.Data
		switch {
		case strings.HasSuffix(subject, ".signal.status"):
			correlatedPayload := correlatedTeamResponsePayload(msg.Data, correlation)
			wrapped, err := protocol.WrapSignalPayloadWithMeta(
				protocol.SourceKindSystem,
				fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID),
				protocol.PayloadKindStatus,
				correlationRunID(correlation),
				t.Manifest.ID,
				"",
				correlatedPayload,
			)
			if err != nil {
				log.Printf("Team [%s] failed to wrap status signal for [%s]: %v", t.Manifest.Name, subject, err)
			} else {
				payload = wrapped
			}
		case strings.HasSuffix(subject, ".signal.result"):
			correlatedPayload := correlatedTeamResponsePayload(msg.Data, correlation)
			wrapped, err := protocol.WrapSignalPayloadWithMeta(
				protocol.SourceKindSystem,
				fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID),
				protocol.PayloadKindResult,
				correlationRunID(correlation),
				t.Manifest.ID,
				"",
				correlatedPayload,
			)
			if err != nil {
				log.Printf("Team [%s] failed to wrap result signal for [%s]: %v", t.Manifest.Name, subject, err)
			} else {
				payload = wrapped
			}
		}
		t.nc.Publish(subject, payload)
	}
}

func (t *Team) rememberCommandCorrelation(correlation teamCommandCorrelation) {
	if strings.TrimSpace(correlation.WorkItemID) == "" {
		return
	}
	correlation.TeamID = firstNonEmptySignalString(correlation.TeamID, t.Manifest.ID)
	correlation.ExpiresAt = time.Now().UTC().Add(5 * time.Minute)
	t.mu.Lock()
	t.pruneExpiredCorrelationsLocked(time.Now().UTC())
	t.pendingCorrelations = append(t.pendingCorrelations, correlation)
	t.mu.Unlock()
}

func (t *Team) responseCommandCorrelation(raw []byte) *teamCommandCorrelation {
	if explicit := correlationFromPayload(raw); explicit != nil {
		explicit.TeamID = firstNonEmptySignalString(explicit.TeamID, t.Manifest.ID)
		return explicit
	}
	return t.consumeCommandCorrelation()
}

func (t *Team) consumeCommandCorrelation() *teamCommandCorrelation {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now().UTC()
	t.pruneExpiredCorrelationsLocked(now)
	if len(t.pendingCorrelations) == 0 {
		return nil
	}
	copied := t.pendingCorrelations[0]
	t.pendingCorrelations = append([]teamCommandCorrelation(nil), t.pendingCorrelations[1:]...)
	return &copied
}

func (t *Team) pruneExpiredCorrelationsLocked(now time.Time) {
	if len(t.pendingCorrelations) == 0 {
		return
	}
	kept := t.pendingCorrelations[:0]
	for _, correlation := range t.pendingCorrelations {
		if correlation.ExpiresAt.IsZero() || !now.After(correlation.ExpiresAt) {
			kept = append(kept, correlation)
		}
	}
	t.pendingCorrelations = kept
}

func extractTeamCommandCorrelation(teamID string, rawEnvelope, normalizedPayload []byte) *teamCommandCorrelation {
	var env protocol.SignalEnvelope
	correlation := &teamCommandCorrelation{TeamID: teamID}
	if err := json.Unmarshal(rawEnvelope, &env); err == nil {
		correlation.TeamID = firstNonEmptySignalString(env.Meta.TeamID, teamID)
		correlation.RunID = env.Meta.RunID
		if fromPayload := correlationFromPayload(env.Payload); fromPayload != nil {
			fromPayload.TeamID = firstNonEmptySignalString(fromPayload.TeamID, correlation.TeamID)
			fromPayload.RunID = firstNonEmptySignalString(fromPayload.RunID, correlation.RunID)
			return fromPayload
		}
	}
	if fromPayload := correlationFromPayload(normalizedPayload); fromPayload != nil {
		fromPayload.TeamID = firstNonEmptySignalString(fromPayload.TeamID, correlation.TeamID)
		fromPayload.RunID = firstNonEmptySignalString(fromPayload.RunID, correlation.RunID)
		return fromPayload
	}
	return nil
}

func correlationFromPayload(raw []byte) *teamCommandCorrelation {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	var ask protocol.TeamAsk
	if err := json.Unmarshal(raw, &ask); err == nil && !ask.IsZero() {
		return correlationFromMap(ask.Context)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	if contextValue, ok := payload["context"].(map[string]any); ok {
		return correlationFromMap(contextValue)
	}
	return correlationFromMap(payload)
}

func correlationFromMap(values map[string]any) *teamCommandCorrelation {
	if values == nil {
		return nil
	}
	workItemID := signalString(values["work_item_id"])
	if workItemID == "" {
		return nil
	}
	return &teamCommandCorrelation{
		WorkItemID: workItemID,
		TeamID:     signalString(values["team_id"]),
		RunID:      signalString(values["run_id"]),
	}
}

func correlatedTeamResponsePayload(raw []byte, correlation *teamCommandCorrelation) []byte {
	if correlation == nil || strings.TrimSpace(correlation.WorkItemID) == "" {
		return raw
	}
	payload := map[string]any{}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && json.Valid(trimmed) {
		if err := json.Unmarshal(trimmed, &payload); err != nil {
			payload = map[string]any{"text": string(raw)}
		}
	} else if len(trimmed) > 0 {
		payload["text"] = string(raw)
		payload["summary"] = "Team result ready"
		payload["details"] = string(raw)
	}
	if signalString(payload["state"]) == "" {
		payload["state"] = "output_ready"
	}
	payload["work_item_id"] = correlation.WorkItemID
	payload["team_id"] = firstNonEmptySignalString(correlation.TeamID, "")
	if correlation.RunID != "" {
		payload["run_id"] = correlation.RunID
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return raw
	}
	return out
}

func correlationRunID(correlation *teamCommandCorrelation) string {
	if correlation == nil {
		return ""
	}
	return correlation.RunID
}

func signalString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func firstNonEmptySignalString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
