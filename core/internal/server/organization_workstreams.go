package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
)

func buildCompactExecutionWorkstreams(teamName string, targetOutputs []string) []TeamLeadExecutionWorkstream {
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Planning lane",
			OwnerLabel:    fmt.Sprintf("%s lead", teamName),
			Status:        "ACTIVE",
			Summary:       "Turn the request into a bounded plan, sequencing, and output contract before the rest of the team expands the work.",
			NextStep:      "Confirm the scope, sequence the deliverables, and hand the package to the focused builder.",
			TargetOutputs: compactOutputSlice(targetOutputs, 0),
		},
		{
			Label:         "Production lane",
			OwnerLabel:    "Focused builder",
			Status:        "NEXT",
			Summary:       "Produce the main retained outputs for the team without inflating the roster or losing the review trail.",
			NextStep:      "Create the first retained deliverable package and keep it ready for review.",
			TargetOutputs: compactOutputSlice(targetOutputs, 1),
		},
		{
			Label:         "Review lane",
			OwnerLabel:    "Delivery reviewer",
			Status:        "NEXT",
			Summary:       "Check the retained package for readiness, risks, and handoff clarity before the team closes the loop.",
			NextStep:      "Review the retained outputs, call out gaps, and decide whether the package is ready to hand off.",
			TargetOutputs: compactOutputSlice(targetOutputs, len(targetOutputs)-1),
		},
	}
}

func buildCreativeExecutionWorkstreams(teamName string, targetOutputs []string) []TeamLeadExecutionWorkstream {
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Creative direction lane",
			OwnerLabel:    fmt.Sprintf("%s lead", teamName),
			Status:        "ACTIVE",
			Summary:       "Shape the visual direction, prompt inputs, and acceptance notes before generation starts.",
			NextStep:      "Lock the concept direction and hand the approved prompt package to the generation specialist.",
			TargetOutputs: compactOutputSlice(targetOutputs, 1),
		},
		{
			Label:         "Artifact generation lane",
			OwnerLabel:    "Image generation specialist",
			Status:        "NEXT",
			Summary:       "Generate the image artifact and keep the first strong candidate reviewable inside the retained package.",
			NextStep:      "Produce the first reviewable artifact candidate and attach it to the workflow package.",
			TargetOutputs: compactOutputSlice(targetOutputs, 0),
		},
		{
			Label:         "Review lane",
			OwnerLabel:    "Artifact reviewer",
			Status:        "NEXT",
			Summary:       "Check the generated artifact against the brief and decide whether the team should hand off or iterate.",
			NextStep:      "Approve or refine the generated artifact before closing the lane.",
			TargetOutputs: compactOutputSlice(targetOutputs, 0),
		},
	}
}

func buildMultiTeamExecutionWorkstreams(targetOutputs []string) []TeamLeadExecutionWorkstream {
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Planning lane",
			OwnerLabel:    "Planning lane lead",
			Status:        "ACTIVE",
			Summary:       "Define deploy posture, scope, acceptance, and the output contract that the other lanes will work against.",
			NextStep:      "Lock the scope and acceptance criteria, then publish the retained planning package.",
			TargetOutputs: compactOutputSlice(targetOutputs, 0),
		},
		{
			Label:         "Validation lane",
			OwnerLabel:    "Validation lane lead",
			Status:        "NEXT",
			Summary:       "Run the runtime and environment proof sequence against the bounded plan instead of widening the ask.",
			NextStep:      "Turn the retained plan into a concrete validation checklist and capture the first proof pass.",
			TargetOutputs: compactOutputSlice(targetOutputs, 1),
		},
		{
			Label:         "Review lane",
			OwnerLabel:    "Review lane lead",
			Status:        "NEXT",
			Summary:       "Assess the retained planning and validation outputs, then call out remaining risk and the next owner.",
			NextStep:      "Review the lane outputs together and name the lead that owns the next follow-through step.",
			TargetOutputs: compactOutputSlice(targetOutputs, len(targetOutputs)-1),
		},
	}
}

func buildExternalWorkflowWorkstreams(target string) []TeamLeadExecutionWorkstream {
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Workflow handoff lane",
			OwnerLabel:    target,
			Status:        "NEXT",
			Summary:       "Shape the external automation so invocation posture and result return stay explicit.",
			NextStep:      "Prepare the external workflow handoff and confirm the normalized result format before activation.",
			TargetOutputs: []string{"Normalized workflow result"},
		},
		{
			Label:         "Review return lane",
			OwnerLabel:    "Soma review",
			Status:        "NEXT",
			Summary:       "Bring the external result back through one readable Mycelis return path instead of leaving the operator inside raw workflow detail.",
			NextStep:      "Review the returned artifact or execution note and decide whether a follow-up action is needed.",
			TargetOutputs: []string{"Linked artifact or execution note"},
		},
	}
}

func buildContinuityExecutionWorkstreams() []TeamLeadExecutionWorkstream {
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Completed work lane",
			OwnerLabel:    "Retained outputs reviewer",
			Status:        "COMPLETE",
			Summary:       "Use the durable outputs to confirm which work is already done before anyone starts rebuilding the package.",
			NextStep:      "Keep the finished retained outputs linked to the package so the next handoff starts from durable evidence.",
			TargetOutputs: []string{"Completed work snapshot"},
		},
		{
			Label:         "Continuity briefing lane",
			OwnerLabel:    "Team Lead continuity",
			Status:        "ACTIVE",
			Summary:       "Summarize the retained package so the next operator move is obvious after a reboot, reload, or interruption.",
			NextStep:      "Publish the continuity summary that ties the finished outputs to the remaining work.",
			TargetOutputs: []string{"Retained package continuity summary"},
		},
		{
			Label:         "Next-step handoff lane",
			OwnerLabel:    "Next lane lead",
			Status:        "NEXT",
			Summary:       "Assign the remaining work to the right lead and keep the handoff explicit instead of flattening it into chat history.",
			NextStep:      "Name the next owner and continue from the recorded checkpoint instead of rebuilding finished work.",
			TargetOutputs: []string{"Remaining work checklist"},
		},
	}
}

func buildContinuityExecutionWorkstreamsFromPackage(group CollaborationGroup, outputs []artifacts.Artifact, nextOwner string) []TeamLeadExecutionWorkstream {
	titles := retainedOutputTitles(outputs, 3)
	coordinator := retainedCoordinatorLabel(group)
	groupName := strings.TrimSpace(group.Name)
	if groupName == "" {
		groupName = "retained package"
	}
	return []TeamLeadExecutionWorkstream{
		{
			Label:         "Completed work lane",
			OwnerLabel:    coordinator,
			Status:        "COMPLETE",
			Summary:       fmt.Sprintf("Use durable outputs like %s as the completed baseline already captured in %s.", humanJoin(titles), groupName),
			NextStep:      fmt.Sprintf("Keep %s linked to %s so the restart point stays anchored in retained work.", humanJoin(retainedOutputTitles(outputs, 2)), groupName),
			TargetOutputs: compactOutputSlice(titles, 0),
		},
		{
			Label:         "Continuity briefing lane",
			OwnerLabel:    coordinator,
			Status:        "ACTIVE",
			Summary:       fmt.Sprintf("Turn %s into a readable continuity brief so the rebooted workspace can resume without rebuilding finished work.", groupName),
			NextStep:      fmt.Sprintf("Publish the retained package summary for %s and highlight %s as the next owner.", groupName, nextOwner),
			TargetOutputs: compactOutputSlice(titles, 1),
		},
		{
			Label:         "Next-step handoff lane",
			OwnerLabel:    nextOwner,
			Status:        "NEXT",
			Summary:       fmt.Sprintf("Hand the retained package to %s with the finished outputs and the remaining step made explicit.", nextOwner),
			NextStep:      fmt.Sprintf("Continue from %s after %s reviews the retained outputs.", titles[0], nextOwner),
			TargetOutputs: compactOutputSlice(titles, len(titles)-1),
		},
	}
}
