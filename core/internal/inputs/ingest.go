package inputs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	HeaderChannelKey      = "Mycelis-Channel-Key"
	HeaderRunID           = "Mycelis-Run-ID"
	HeaderTeamID          = "Mycelis-Team-ID"
	HeaderAgentID         = "Mycelis-Agent-ID"
	HeaderSourceKind      = "Mycelis-Source-Kind"
	HeaderSourceChannel   = "Mycelis-Source-Channel"
	HeaderPayloadKind     = "Mycelis-Payload-Kind"
	HeaderSourceTimestamp = "Mycelis-Source-Timestamp"
)

func EventFromBusMessage(source Source, subject string, data []byte, headers map[string][]string) (IngestEvent, error) {
	subject = strings.TrimSpace(subject)
	if source.ID == "" {
		return IngestEvent{}, fmt.Errorf("%w: source id is required for ingestion", ErrInvalidInput)
	}
	if source.Status != "" && source.Status != StatusAvailable {
		return IngestEvent{}, fmt.Errorf("%w: source %s is not available", ErrInvalidInput, source.ID)
	}
	if source.AllowedIngressSubject != "" && subject != source.AllowedIngressSubject {
		return IngestEvent{}, fmt.Errorf("%w: subject %s is not registered for source %s", ErrInvalidInput, subject, source.ID)
	}

	event := IngestEvent{
		SourceID:      source.ID,
		ChannelKey:    firstHeader(headers, HeaderChannelKey),
		RunID:         firstHeader(headers, HeaderRunID),
		TeamID:        firstHeader(headers, HeaderTeamID),
		AgentID:       firstHeader(headers, HeaderAgentID),
		SourceKind:    firstHeader(headers, HeaderSourceKind),
		SourceChannel: firstHeader(headers, HeaderSourceChannel),
		PayloadKind:   firstHeader(headers, HeaderPayloadKind),
		TenantID:      source.TenantID,
	}
	if event.ChannelKey == "" {
		event.ChannelKey = "default"
	}
	if event.SourceKind == "" {
		event.SourceKind = defaultSourceKind(source.AdapterKind)
	}
	if event.SourceChannel == "" {
		event.SourceChannel = subject
	}
	if event.PayloadKind == "" {
		event.PayloadKind = defaultPayloadKind(source.AdapterKind)
	}
	if event.TenantID == "" {
		event.TenantID = "default"
	}
	if rawTimestamp := firstHeader(headers, HeaderSourceTimestamp); rawTimestamp != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, rawTimestamp); err == nil {
			event.SourceTimestamp = &parsed
		}
	}

	payload, metaTimestamp, err := normalizedPayloadFromBusData(data, &event)
	if err != nil {
		return IngestEvent{}, err
	}
	if event.SourceTimestamp == nil && !metaTimestamp.IsZero() {
		event.SourceTimestamp = &metaTimestamp
	}
	event.Payload = payload
	if event.ChannelKey == "default" {
		if channelKey := payloadStringField(payload, "channel_key"); channelKey != "" {
			event.ChannelKey = channelKey
		}
	}
	event.PayloadHash = hashPayload(payload)
	return event, nil
}

func NormalizeNATSMessage(source Source, subject string, data []byte) IngestEvent {
	event, err := EventFromBusMessage(source, subject, data, nil)
	if err != nil {
		return IngestEvent{
			SourceID:      source.ID,
			ChannelKey:    "default",
			Payload:       json.RawMessage(`{}`),
			SourceKind:    defaultSourceKind(source.AdapterKind),
			SourceChannel: subject,
			PayloadKind:   defaultPayloadKind(source.AdapterKind),
			TenantID:      firstNonEmpty(source.TenantID, "default"),
		}
	}
	return event
}

func normalizedPayloadFromBusData(data []byte, event *IngestEvent) (json.RawMessage, time.Time, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return json.RawMessage(`{}`), time.Time{}, nil
	}

	var env protocol.SignalEnvelope
	if err := json.Unmarshal(trimmed, &env); err == nil && looksLikeSignalEnvelope(env) {
		applySignalMeta(event, env.Meta)
		switch {
		case len(bytes.TrimSpace(env.Payload)) > 0 && string(bytes.TrimSpace(env.Payload)) != "null":
			return append(json.RawMessage(nil), env.Payload...), env.Meta.Timestamp, nil
		case strings.TrimSpace(env.Text) != "":
			payload, err := json.Marshal(map[string]string{"text": env.Text})
			return payload, env.Meta.Timestamp, err
		default:
			return json.RawMessage(`{}`), env.Meta.Timestamp, nil
		}
	}

	if json.Valid(trimmed) {
		return append(json.RawMessage(nil), trimmed...), time.Time{}, nil
	}
	payload, err := json.Marshal(map[string]string{"text": string(data)})
	return payload, time.Time{}, err
}

func looksLikeSignalEnvelope(env protocol.SignalEnvelope) bool {
	return env.Meta.SourceKind != "" ||
		env.Meta.SourceChannel != "" ||
		env.Meta.PayloadKind != "" ||
		!env.Meta.Timestamp.IsZero() ||
		len(bytes.TrimSpace(env.Payload)) > 0 ||
		strings.TrimSpace(env.Text) != ""
}

func applySignalMeta(event *IngestEvent, meta protocol.SignalMeta) {
	if event == nil {
		return
	}
	if meta.SourceKind != "" {
		event.SourceKind = string(meta.SourceKind)
	}
	if meta.SourceChannel != "" {
		event.SourceChannel = meta.SourceChannel
	}
	if meta.PayloadKind != "" {
		event.PayloadKind = string(meta.PayloadKind)
	}
	if meta.RunID != "" {
		event.RunID = meta.RunID
	}
	if meta.TeamID != "" {
		event.TeamID = meta.TeamID
	}
	if meta.AgentID != "" {
		event.AgentID = meta.AgentID
	}
}

func firstHeader(headers map[string][]string, key string) string {
	for candidate, values := range headers {
		if !strings.EqualFold(candidate, key) || len(values) == 0 {
			continue
		}
		return strings.TrimSpace(values[0])
	}
	return ""
}

func payloadStringField(payload json.RawMessage, key string) string {
	var values map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return ""
	}
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func defaultSourceKind(adapterKind string) string {
	switch adapterKind {
	case AdapterMCP:
		return string(protocol.SourceKindMCP)
	case AdapterSensor:
		return string(protocol.SourceKindSensor)
	case AdapterDevice:
		return string(protocol.SourceKindIoT)
	case AdapterFile:
		return string(protocol.SourceKindSystem)
	default:
		return string(protocol.SourceKindWebAPI)
	}
}

func defaultPayloadKind(adapterKind string) string {
	switch adapterKind {
	case AdapterSensor, AdapterDevice:
		return string(protocol.PayloadKindTelemetry)
	default:
		return string(protocol.PayloadKindEvent)
	}
}

func hashPayload(payload json.RawMessage) string {
	sum := sha256.Sum256(bytes.TrimSpace(payload))
	return hex.EncodeToString(sum[:])
}
