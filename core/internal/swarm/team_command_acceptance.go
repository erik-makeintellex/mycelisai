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

// reportUndeliveredCommand makes every refused command visible (TPD). A
// command this Team instance already forwarded is a duplicate of live work
// (dispatcher retry, bus redelivery): it is logged and its original delivery
// keeps the work item's status. A non-steering command that the durable
// receipt refused but this instance never forwarded was accepted by an earlier
// Core process (restart mid-run). Its outcome there is unknown, so it is not
// re-executed; the work item is reported as needing recovery instead of being
// left queued until the recovery deadline.
func (t *Team) reportUndeliveredCommand(correlation teamCommandCorrelation, sourceChannel string, steering bool) {
	if steering || t.commandReceipts == nil || t.commandDeliveredHere(correlation) {
		log.Printf("component=team_command action=drop_duplicate team=%s work_item_id=%s key=%s steering=%t reason=already_delivered_by_this_process",
			t.Manifest.ID, correlation.WorkItemID, correlation.commandKey(), steering)
		return
	}
	log.Printf("component=team_command action=report_needs_recovery team=%s work_item_id=%s key=%s reason=accepted_before_restart_outcome_unknown",
		t.Manifest.ID, correlation.WorkItemID, correlation.commandKey())
	payload, err := json.Marshal(map[string]any{
		"work_item_id": correlation.WorkItemID,
		// A distinct key: the projection dedupes status on kind plus key, and
		// the acceptance status already used the command key.
		"idempotency_key":   correlation.commandKey() + ":redelivered",
		"state":             string(protocol.TeamWorkStateDegraded),
		"needs_operator":    true,
		"degradation_state": "team_command_outcome_unknown",
		"blocked_by":        []string{"command_accepted_before_restart"},
		"headline":          "Team work needs recovery",
		"details":           "The team accepted this approved task before Core restarted. Core did not run it again, because the earlier attempt's outcome is unknown.",
		"recovery_options": []string{
			"Check the team's retained output for this task before deciding whether a retry is safe.",
			"Ask Soma to retry this work with the same team if nothing was produced.",
			"Archive the work if it is no longer needed.",
		},
	})
	if err != nil {
		log.Printf("Team [%s] could not encode recovery status: %v", t.Manifest.Name, err)
		return
	}
	wrapper, err := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindSystem, sourceChannel, protocol.PayloadKindStatus, correlation.RunID, t.Manifest.ID, "", payload)
	if err != nil {
		log.Printf("Team [%s] could not wrap recovery status: %v", t.Manifest.Name, err)
		return
	}
	if err := t.nc.Publish(fmt.Sprintf(protocol.TopicTeamSignalStatus, t.Manifest.ID), wrapper); err != nil {
		log.Printf("Team [%s] could not publish recovery status: %v", t.Manifest.Name, err)
	}
}
