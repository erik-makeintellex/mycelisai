package signal

import (
	"encoding/json"
	"strings"
)

var sensitiveOperatorFields = map[string]struct{}{
	"api_key":       {},
	"authorization": {},
	"credential":    {},
	"password":      {},
	"access_token":  {},
	"refresh_token": {},
	"secret":        {},
	"token":         {},
}

func sanitizeOperatorPayload(payload string) string {
	var value any
	if err := json.Unmarshal([]byte(payload), &value); err != nil {
		return payload
	}
	redactOperatorValue(value)
	raw, err := json.Marshal(value)
	if err != nil {
		return payload
	}
	return string(raw)
}

func redactOperatorValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if _, sensitive := sensitiveOperatorFields[strings.ToLower(strings.TrimSpace(key))]; sensitive {
				typed[key] = "[REDACTED]"
				continue
			}
			redactOperatorValue(child)
		}
	case []any:
		for _, child := range typed {
			redactOperatorValue(child)
		}
	}
}
