package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// handleTrigger receives an external signal and broadens it to the internal team bus.
func (t *Team) handleTrigger(msg *nats.Msg) {
	log.Printf("Team [%s] Triggered by [%s]", t.Manifest.Name, msg.Subject)
	internalSubject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, t.Manifest.ID)
	t.nc.Publish(internalSubject, normalizeCommandPayload(msg.Data))
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
	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		return []byte(asString)
	}
	return append([]byte(nil), trimmed...)
}

// handleResponse receives an internal signal and broadens it to the external team bus.
func (t *Team) handleResponse(msg *nats.Msg) {
	log.Printf("Team [%s] Response: %s", t.Manifest.Name, string(msg.Data))
	for _, subject := range t.Manifest.Deliveries {
		payload := msg.Data
		switch {
		case strings.HasSuffix(subject, ".signal.status"):
			wrapped, err := protocol.WrapSignalPayload(
				protocol.SourceKindSystem,
				fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID),
				protocol.PayloadKindStatus,
				t.Manifest.ID,
				msg.Data,
			)
			if err != nil {
				log.Printf("Team [%s] failed to wrap status signal for [%s]: %v", t.Manifest.Name, subject, err)
			} else {
				payload = wrapped
			}
		case strings.HasSuffix(subject, ".signal.result"):
			wrapped, err := protocol.WrapSignalPayload(
				protocol.SourceKindSystem,
				fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID),
				protocol.PayloadKindResult,
				t.Manifest.ID,
				msg.Data,
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
