package server

import (
	"context"
	"errors"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) persistConfirmedActionTeamWork(ctx context.Context, link confirmedActionTeamWorkLink, results []plannedToolExecutionResult) error {
	if s.getDB() == nil || link.Scope == nil {
		return nil
	}
	var errs []error
	if err := s.persistConfirmedCreateTeamItems(ctx, link, results); err != nil {
		errs = append(errs, err)
	}
	if err := s.persistConfirmedDelegatedWorkItems(ctx, link, results); err != nil {
		errs = append(errs, err)
	}
	if err := s.persistConfirmedDeliverableWorkItems(ctx, link, results); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *AdminServer) persistFailedConfirmedActionTeamWork(ctx context.Context, link confirmedActionTeamWorkLink, failure error) error {
	if s.getDB() == nil || link.Scope == nil {
		return nil
	}
	var errs []error
	for _, planned := range link.Scope.PlannedToolCalls {
		planned = normalizePlannedToolCall(planned)
		toolName := strings.TrimSpace(planned.Name)
		if !isTeamWorkTool(toolName) {
			continue
		}
		teamID := firstNonEmptyString(confirmedActionTeamID(planned.Arguments), confirmedActionCreatedTeamIDFromScope(link.Scope))
		if teamID == "" {
			continue
		}
		item := baseConfirmedActionWorkItem(link, teamID, objectiveForPlannedTool(toolName, planned.Arguments))
		item.ExecutionShape = executionShapeForTeamWorkTool(toolName)
		if item.ExecutionShape == protocol.TeamExecutionShapeCreateTeam {
			item.ExecutionShape = protocol.TeamExecutionShapeDelegatedWork
		}
		item.State = protocol.TeamWorkStateDegraded
		item.NeedsOperator = true
		item.DegradationState = "confirmed_action_failed"
		item.RecoveryOptions = []string{"Review the failed run proof, correct the blocker, and retry the confirmed action."}
		item.LastEvent = confirmedActionStatusEvent(link, item, protocol.TeamWorkStateDegraded, "Team work degraded", failureSummary(failure), "operator_attention", "Retry after reviewing the failed proof.")
		if err := s.persistTeamWorkItemWithLifecycle(ctx, &item, []protocol.TeamStatusEvent{*item.LastEvent}, confirmedActionInteraction(link, item, "degraded", failureSummary(failure), toolName, planned.Arguments)); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type confirmedActionTeamWorkLink struct {
	ProofID         string
	ContractID      string
	ProofArtifactID string
	RunID           string
	AuditID         string
	AuditUser       string
	Scope           *protocol.ScopeValidation
}

func (s *AdminServer) persistConfirmedCreateTeamItems(ctx context.Context, link confirmedActionTeamWorkLink, results []plannedToolExecutionResult) error {
	var errs []error
	for _, result := range results {
		if strings.TrimSpace(result.Name) != "create_team" {
			continue
		}
		teamID := confirmedActionTeamID(result.Arguments)
		if teamID == "" {
			continue
		}
		item := baseConfirmedActionWorkItem(link, teamID, "Create runtime team "+teamDisplayName(result.Arguments, teamID))
		item.ExecutionShape = protocol.TeamExecutionShapeCreateTeam
		item.State = protocol.TeamWorkStateNew
		item.ExpectedOutputs = []string{"Runtime team shell"}
		item.ExpectedProof = []string{"Confirmed proposal proof", "Team roster visibility"}
		item.CapabilityRequirements = confirmedActionStringSlice(mergedTeamArgs(result.Arguments)["allowed_capabilities"])
		item.OutputRefs = outputRefsForTeamWork(link, item.WorkItemID, teamID, executionOutputsForResult(result))
		item.LastEvent = confirmedActionStatusEvent(link, item, protocol.TeamWorkStateNew, "Team created; no active work started", "The confirmed action created the team shell. Start a delegated task or deliverable request before treating the team as active work.", "verified", "Ask Soma for the team's first work item.")
		interaction := confirmedActionInteraction(link, item, "create_team", "Soma created the runtime team shell without marking it as active execution.", result.Name, result.Arguments)
		if err := s.persistTeamWorkItemWithLifecycle(ctx, &item, []protocol.TeamStatusEvent{*item.LastEvent}, interaction); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *AdminServer) persistConfirmedDelegatedWorkItems(ctx context.Context, link confirmedActionTeamWorkLink, results []plannedToolExecutionResult) error {
	var errs []error
	for _, result := range results {
		if !isDelegateTool(result.Name) {
			continue
		}
		teamID := firstNonEmptyString(confirmedActionTeamID(result.Arguments), confirmedActionCreatedTeamID(results))
		if teamID == "" {
			continue
		}
		item := baseConfirmedActionWorkItem(link, teamID, objectiveForPlannedTool(result.Name, result.Arguments))
		item.ExecutionShape = protocol.TeamExecutionShapeDelegatedWork
		item.State = protocol.TeamWorkStateQueued
		item.ExpectedOutputs = expectedOutputsFromDelegateArgs(result.Arguments)
		item.ExpectedProof = expectedProofFromDelegateArgs(result.Arguments)
		item.CapabilityRequirements = requiredCapabilitiesFromDelegateArgs(result.Arguments)
		item.LastEvent = confirmedActionStatusEvent(link, item, protocol.TeamWorkStateQueued, "Team work queued", "Soma delegated the confirmed ask to the team command channel.", "pending_team_response", "Wait for team status or result output.")
		interaction := confirmedActionInteraction(link, item, "delegate", firstNonEmptyString(result.Output, "Soma delegated a confirmed task to the team."), result.Name, result.Arguments)
		if err := s.persistTeamWorkItemWithLifecycle(ctx, &item, []protocol.TeamStatusEvent{*item.LastEvent}, interaction); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *AdminServer) persistConfirmedDeliverableWorkItems(ctx context.Context, link confirmedActionTeamWorkLink, results []plannedToolExecutionResult) error {
	var errs []error
	defaultTeamID := confirmedActionCreatedTeamID(results)
	for _, result := range results {
		if !isDeliverableTool(result.Name) {
			continue
		}
		teamID := firstNonEmptyString(confirmedActionTeamID(result.Arguments), defaultTeamID)
		if teamID == "" {
			continue
		}
		outputs := executionOutputsForResult(result)
		item := baseConfirmedActionWorkItem(link, teamID, objectiveForDeliverableResult(result))
		item.ExecutionShape = protocol.TeamExecutionShapeDeliverable
		item.State = protocol.TeamWorkStateOutputReady
		item.ExpectedOutputs = expectedOutputsFromDeliverableResult(result, outputs)
		item.ExpectedProof = []string{"Confirmed execution run", "Retained output reference"}
		item.OutputRefs = outputRefsForTeamWork(link, item.WorkItemID, teamID, outputs)
		item.LastEvent = confirmedActionStatusEvent(link, item, protocol.TeamWorkStateOutputReady, "Team output ready", "The confirmed deliverable request completed and produced retained output proof.", "verified", "Review the retained output and proof package.")
		events := []protocol.TeamStatusEvent{
			*confirmedActionStatusEvent(link, item, protocol.TeamWorkStateQueued, "Team work queued", "The confirmed deliverable request entered durable team work.", "verified", ""),
			*confirmedActionStatusEvent(link, item, protocol.TeamWorkStateRunning, "Team work running", "Soma executed the confirmed deliverable request for this team.", "verified", ""),
			*item.LastEvent,
		}
		interaction := confirmedActionInteraction(link, item, "output_ready", firstNonEmptyString(result.Output, "Confirmed deliverable output is ready."), result.Name, result.Arguments)
		if err := s.persistTeamWorkItemWithLifecycle(ctx, &item, events, interaction); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *AdminServer) persistTeamWorkItemWithLifecycle(ctx context.Context, item *protocol.TeamWorkItem, events []protocol.TeamStatusEvent, interaction protocol.TeamInteraction) error {
	item.LastEvent = nil
	if err := protocol.ValidateTeamWorkItem(*item); err != nil {
		return err
	}
	if err := s.insertTeamWorkItemDB(ctx, item); err != nil {
		return err
	}
	for i := range events {
		if err := s.insertTeamStatusEventDB(ctx, &events[i]); err != nil {
			return err
		}
	}
	if len(events) > 0 {
		if err := s.updateTeamWorkItemLastEventDB(ctx, item, events[len(events)-1]); err != nil {
			return err
		}
	}
	if strings.TrimSpace(interaction.Summary) != "" {
		if err := s.insertTeamInteractionDB(ctx, &interaction); err != nil {
			return err
		}
	}
	return nil
}
