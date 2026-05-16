package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func hasTriggerCondition(condition json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(condition))
	return trimmed != "" && trimmed != "{}" && trimmed != "null"
}

func (e *Engine) evaluateCondition(
	ctx context.Context,
	condition json.RawMessage,
	eventID string,
	sourceRunID string,
	eventType string,
) (bool, string, error) {
	var expected map[string]interface{}
	if err := json.Unmarshal(condition, &expected); err != nil {
		return false, "", fmt.Errorf("condition JSON is invalid: %w", err)
	}
	if len(expected) == 0 {
		return true, "", nil
	}

	event := protocol.MissionEventEnvelope{
		ID:        eventID,
		RunID:     sourceRunID,
		EventType: protocol.EventType(eventType),
		Payload:   map[string]interface{}{},
	}
	if e.events != nil && eventID != "" {
		loaded, err := e.events.GetEvent(ctx, eventID)
		if err != nil {
			return false, "", fmt.Errorf("condition event payload unavailable: %w", err)
		}
		event = *loaded
	}

	for key, want := range expected {
		got, ok := conditionValue(event, key)
		if !ok {
			return false, fmt.Sprintf("%s was not present", key), nil
		}
		if !conditionEqual(got, want) {
			return false, fmt.Sprintf("%s was %v", key, got), nil
		}
	}
	return true, "", nil
}

func conditionValue(event protocol.MissionEventEnvelope, key string) (interface{}, bool) {
	switch key {
	case "id", "event_id", "mission_event_id":
		return event.ID, event.ID != ""
	case "run_id":
		return event.RunID, event.RunID != ""
	case "event_type":
		return string(event.EventType), event.EventType != ""
	case "severity":
		return string(event.Severity), event.Severity != ""
	case "source_agent":
		return event.SourceAgent, event.SourceAgent != ""
	case "source_team":
		return event.SourceTeam, event.SourceTeam != ""
	case "payload":
		return event.Payload, event.Payload != nil
	}

	if strings.HasPrefix(key, "payload.") {
		return nestedValue(event.Payload, strings.TrimPrefix(key, "payload."))
	}
	if got, ok := nestedValue(event.Payload, key); ok {
		return got, true
	}
	return false, false
}

func nestedValue(source map[string]interface{}, dottedPath string) (interface{}, bool) {
	if source == nil || dottedPath == "" {
		return nil, false
	}
	var current interface{} = source
	for _, part := range strings.Split(dottedPath, ".") {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = obj[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func conditionEqual(got interface{}, want interface{}) bool {
	switch expected := want.(type) {
	case []interface{}:
		for _, item := range expected {
			if conditionEqual(got, item) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		gotMap, ok := got.(map[string]interface{})
		if !ok {
			return false
		}
		for key, value := range expected {
			gotValue, ok := nestedValue(gotMap, key)
			if !ok || !conditionEqual(gotValue, value) {
				return false
			}
		}
		return true
	}

	if reflect.DeepEqual(got, want) {
		return true
	}
	gotNumber, gotOK := numberValue(got)
	wantNumber, wantOK := numberValue(want)
	if gotOK && wantOK {
		return gotNumber == wantNumber
	}
	return fmt.Sprint(got) == fmt.Sprint(want)
}

func numberValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case json.Number:
		parsed, err := v.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
