package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func (t *Team) acceptCommand(ctx context.Context, correlation teamCommandCorrelation, sourceChannel string, steering bool) (bool, error) {
	if !steering {
		return t.acceptCommandCorrelation(ctx, correlation, sourceChannel)
	}
	if t.commandReceipts != nil {
		accepted, err := t.commandReceipts.AcceptCommand(ctx, correlation, sourceChannel)
		if err != nil || !accepted {
			return accepted, err
		}
	}
	return t.rememberSteeringCommand(correlation), nil
}

func (t *Team) rememberSteeringCommand(correlation teamCommandCorrelation) bool {
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
	return true
}

func (t *Team) publishAgentInterjections(guidance string) {
	guidance = strings.TrimSpace(guidance)
	if guidance == "" {
		return
	}
	for _, member := range t.Manifest.Members {
		subject := fmt.Sprintf(protocol.TopicAgentInterjectionFmt, member.ID)
		if err := t.nc.Publish(subject, []byte(guidance)); err != nil {
			log.Printf("Team [%s] could not steer agent [%s]: %v", t.Manifest.Name, member.ID, err)
		}
	}
}

func extractTeamSteering(payload []byte) (string, bool) {
	var command map[string]any
	if err := json.Unmarshal(payload, &command); err != nil {
		return "", false
	}
	contextValue, _ := command["context"].(map[string]any)
	if !strings.EqualFold(signalString(contextValue["action"]), string(protocol.TeamWorkActionSteer)) {
		return "", false
	}
	guidance := signalString(command["guidance"])
	if guidance == "" {
		guidance = signalString(command["goal"])
	}
	return guidance, true
}
