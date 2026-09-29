package swarm

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"unicode"

	"github.com/mycelis/core/pkg/protocol"
)

// triggerAuthorityKeys grant execution posture (run_id, contract_id,
// intent_proof_id, work_item_id, idempotency_key; see
// trust.VerifyExecutionClaim) or command/result correlation to a trigger.
// Only Core's confirmed dispatch may set them (F16, F16b, F16c).
var triggerAuthorityKeys = []string{"run_id", "contract_id", "intent_proof_id", "work_item_id", "idempotency_key"}

// isTriggerAuthorityKey matches a triggerAuthorityKeys entry in any case and
// with or without separators (run_id, RUN_ID, runId, run-id, "run id").
func isTriggerAuthorityKey(key string) bool {
	folded := strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == '.' || unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, key)
	for _, authority := range triggerAuthorityKeys {
		if folded == strings.ReplaceAll(authority, "_", "") {
			return true
		}
	}
	return false
}

func stripTriggerAuthorityKeys(values map[string]any) {
	for key := range values {
		if isTriggerAuthorityKey(key) {
			delete(values, key)
		}
	}
}

// planningOnlyTriggerPayload removes triggerAuthorityKeys from a JSON object
// trigger, at the top level and inside every key matching "context" without
// regard to case (encoding/json binds TeamAsk.Context case-insensitively).
// Non-object payloads carry no such fields and are returned unchanged.
func planningOnlyTriggerPayload(payload []byte) []byte {
	var object map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(payload), &object); err != nil || object == nil {
		return payload
	}
	stripTriggerAuthorityKeys(object)
	for key, value := range object {
		if nested, ok := value.(map[string]any); ok && strings.EqualFold(key, "context") {
			stripTriggerAuthorityKeys(nested)
		}
	}
	out, err := json.Marshal(object)
	if err != nil {
		return []byte("{}")
	}
	return out
}

// delegateAskAuthority is the delegate_task ingress guard (F16c). Tool
// arguments are model-written, so the ask's context loses every
// triggerAuthorityKeys entry unless ctx is Core's confirmed dispatch for the
// same run, whose arguments Core annotated from the approved plan
// (server.annotateConfirmedDelegationCall).
func delegateAskAuthority(ctx context.Context, ask protocol.TeamAsk) protocol.TeamAsk {
	if len(ask.Context) == 0 || confirmedDispatchInvocation(ctx, ask.Context) {
		return ask
	}
	clean := make(map[string]any, len(ask.Context))
	for key, value := range ask.Context {
		clean[key] = value
	}
	stripTriggerAuthorityKeys(clean)
	ask.Context = clean
	return ask
}

// confirmedDispatchInvocation recognizes Core's confirmed dispatch by the
// Core-only marker (WithConfirmedDispatchToolContext, TPD), never by the
// source-channel string, which any caller could copy.
func confirmedDispatchInvocation(ctx context.Context, askContext map[string]any) bool {
	inv, ok := ToolInvocationContextFromContext(ctx)
	if !ok || !inv.confirmedDispatch || inv.PlanningOnly || inv.SourceKind != protocol.SourceKindWebAPI {
		return false
	}
	runID := strings.TrimSpace(inv.RunID)
	return runID != "" && signalString(askContext["run_id"]) == runID
}

// planningOnlyTeamInputMessage is the publish_signal ingress guard (F16c): a
// model-written message bound for a team's internal.command or
// internal.trigger subject loses triggerAuthorityKeys. For internal.command it
// also strips inside the one JSON-string layer the team decodes
// (normalizeCommandPayload). Other subjects are returned unchanged.
func planningOnlyTeamInputMessage(subject string, message []byte) []byte {
	subject = strings.TrimSpace(subject)
	switch {
	case matchesTeamTopic(subject, protocol.TopicTeamInternalTrigger):
		return planningOnlyTriggerPayload(message)
	case matchesTeamTopic(subject, protocol.TopicTeamInternalCommand):
		var inner string
		if json.Unmarshal(bytes.TrimSpace(message), &inner) == nil {
			out, _ := json.Marshal(string(planningOnlyTriggerPayload([]byte(inner))))
			return out
		}
		return planningOnlyTriggerPayload(message)
	default:
		return message
	}
}

// matchesTeamTopic reports whether subject is format (a protocol team topic
// with one %s) for some non-empty team id, including ids that contain dots.
func matchesTeamTopic(subject, format string) bool {
	prefix, suffix, ok := strings.Cut(format, "%s")
	return ok && len(subject) > len(prefix)+len(suffix) && strings.HasPrefix(subject, prefix) && strings.HasSuffix(subject, suffix)
}
