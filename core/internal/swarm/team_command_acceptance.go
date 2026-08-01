package swarm

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mycelis/core/pkg/protocol"
)

func (t *Team) publishCommandAccepted(correlation teamCommandCorrelation, sourceChannel string) {
	payload, err := json.Marshal(map[string]any{
		"work_item_id":    correlation.WorkItemID,
		"idempotency_key": correlation.commandKey(),
		"state":           string(protocol.TeamWorkStateRunning),
		"headline":        "Team accepted work",
		"details":         "The team durably accepted the command and started its response lane.",
		"next_action":     "Keep working with Soma while the team prepares a status or result.",
	})
	if err != nil {
		log.Printf("Team [%s] could not encode command acceptance: %v", t.Manifest.Name, err)
		return
	}
	wrapper, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindSystem,
		sourceChannel,
		protocol.PayloadKindStatus,
		correlation.RunID,
		t.Manifest.ID,
		"",
		payload,
	)
	if err != nil {
		log.Printf("Team [%s] could not wrap command acceptance: %v", t.Manifest.Name, err)
		return
	}
	subject := fmt.Sprintf(protocol.TopicTeamSignalStatus, t.Manifest.ID)
	if err := t.nc.Publish(subject, wrapper); err != nil {
		log.Printf("Team [%s] could not publish command acceptance: %v", t.Manifest.Name, err)
	}
}
