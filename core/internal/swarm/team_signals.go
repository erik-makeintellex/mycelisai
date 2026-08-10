package swarm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

const teamCommandCorrelationTTL = 30 * time.Minute

// handleTrigger receives an external signal and broadens it to the internal team bus.
func (t *Team) handleTrigger(msg *nats.Msg) {
	log.Printf("Team [%s] Triggered by [%s]", t.Manifest.Name, msg.Subject)
	internalSubject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, t.Manifest.ID)
	payload := normalizeCommandPayload(msg.Data)
	if correlation := extractTeamCommandCorrelation(t.Manifest.ID, msg.Data, payload); correlation != nil {
		accepted, err := t.acceptCommandCorrelation(context.Background(), *correlation, msg.Subject)
		if err != nil {
			log.Printf("Team [%s] could not durably accept command [%s]: %v", t.Manifest.Name, correlation.commandKey(), err)
			return
		}
		if !accepted {
			log.Printf("Team [%s] ignored replayed command [%s]", t.Manifest.Name, correlation.commandKey())
			return
		}
		if t.commandReceipts != nil {
			t.publishCommandAccepted(*correlation, msg.Subject)
		}
	}
	t.nc.Publish(internalSubject, payload)
}

func (t *Team) acceptCommandCorrelation(ctx context.Context, correlation teamCommandCorrelation, sourceChannel string) (bool, error) {
	if t.commandReceipts != nil {
		accepted, err := t.commandReceipts.AcceptCommand(ctx, correlation, sourceChannel)
		if err != nil || !accepted {
			return accepted, err
		}
	}
	return t.rememberCommandCorrelation(correlation), nil
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

func (t *Team) rememberCommandCorrelation(correlation teamCommandCorrelation) bool {
	if strings.TrimSpace(correlation.WorkItemID) == "" {
		return true
	}
	correlation.TeamID = firstNonEmptySignalString(correlation.TeamID, t.Manifest.ID)
	// Keep response correlation beyond the durable work recovery deadline.
	// Local-model package work can legitimately take more than five minutes;
	// expiring first would publish an orphan result and leave work queued.
	correlation.ExpiresAt = time.Now().UTC().Add(teamCommandCorrelationTTL)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pruneExpiredCorrelationsLocked(time.Now().UTC())
	key := correlation.commandKey()
	if expiry, exists := t.seenCommandKeys[key]; exists && time.Now().UTC().Before(expiry) {
		return false
	}
	if t.seenCommandKeys == nil {
		t.seenCommandKeys = map[string]time.Time{}
	}
	t.seenCommandKeys[key] = time.Now().UTC().Add(time.Hour)
	t.pendingCorrelations = append(t.pendingCorrelations, correlation)
	return true
}

func (t *Team) responseCommandCorrelation(raw []byte) *teamCommandCorrelation {
	if explicit := correlationFromPayload(raw); explicit != nil {
		explicit.TeamID = firstNonEmptySignalString(explicit.TeamID, t.Manifest.ID)
		t.forgetMatchingCommandCorrelation(*explicit)
		return explicit
	}
	return t.consumeCommandCorrelation()
}

func (t *Team) forgetMatchingCommandCorrelation(explicit teamCommandCorrelation) {
	key := explicit.commandKey()
	if strings.TrimSpace(key) == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, pending := range t.pendingCorrelations {
		if pending.commandKey() != key {
			continue
		}
		t.pendingCorrelations = append(t.pendingCorrelations[:i], t.pendingCorrelations[i+1:]...)
		return
	}
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
	for key, expiry := range t.seenCommandKeys {
		if now.After(expiry) {
			delete(t.seenCommandKeys, key)
		}
	}
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
		WorkItemID:     workItemID,
		TeamID:         signalString(values["team_id"]),
		RunID:          signalString(values["run_id"]),
		IdempotencyKey: signalString(values["idempotency_key"]),
	}
}

func (c teamCommandCorrelation) commandKey() string {
	return firstNonEmptySignalString(c.IdempotencyKey, c.WorkItemID)
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
		if teamResponseHasBlocker(payload) {
			payload["state"] = string(protocol.TeamWorkStateDegraded)
			payload["needs_operator"] = true
		} else {
			payload["state"] = string(protocol.TeamWorkStateOutputReady)
		}
	}
	payload["work_item_id"] = correlation.WorkItemID
	payload["idempotency_key"] = correlation.commandKey()
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

func teamResponseHasBlocker(payload map[string]any) bool {
	for _, key := range []string{"blocker", "blocked_by", "error"} {
		if signalString(payload[key]) != "" {
			return true
		}
	}
	for _, key := range []string{"blockers", "errors"} {
		if values, ok := payload[key].([]any); ok && len(values) > 0 {
			return true
		}
	}
	switch strings.ToLower(firstNonEmptySignalString(
		signalString(payload["status"]),
		signalString(payload["result"]),
	)) {
	case "blocked", "degraded", "failed", "failure", "incomplete", "needs_operator":
		return true
	}
	text := strings.ToLower(strings.TrimSpace(firstNonEmptySignalString(
		signalString(payload["text"]),
		signalString(payload["details"]),
	)))
	for _, prefix := range []string{"blocked:", "failed:", "incomplete:", "could not "} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}
