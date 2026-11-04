import { councilLabel } from "@/lib/labels";
import type { SomaSuggestion } from "@/components/soma/SomaSuggestionBar";

export function teamSuggestions(teamName: string): SomaSuggestion[] {
    return [
        {
            label: "Plan this team",
            detail: `Shape the next priority for ${teamName}.`,
            prompt: `Plan the next move for ${teamName}`,
        },
        {
            label: "Review state",
            detail: `Check ${teamName} progress and current risk.`,
            prompt: `Review the current state of ${teamName}`,
        },
        {
            label: "Create deliverable",
            detail: `Draft an output with ${teamName} and store it visibly.`,
            prompt: `Work with ${teamName} to produce a concise brief we can review and keep in the workspace.`,
        },
        {
            label: "Check readiness",
            detail: `Verify ${teamName} agents, tools, and inputs.`,
            prompt: `Check whether ${teamName} has the agents, tools, and inputs needed for its next run.`,
        },
    ];
}

export function somaPlaceholder({
    assistantName,
    broadcastMode,
    currentTeamName,
    directTarget,
    showAdvancedRouting,
    simpleMode,
}: {
    assistantName: string;
    broadcastMode: boolean;
    currentTeamName?: string;
    directTarget: string | null;
    showAdvancedRouting: boolean;
    simpleMode: boolean;
}): string {
    if (showAdvancedRouting && broadcastMode) {
        return "Broadcast to all teams...";
    }
    if (showAdvancedRouting && directTarget) {
        return `Direct to ${councilLabel(directTarget, assistantName).name}...`;
    }
    if (currentTeamName?.trim()) {
        return `Direct the ${currentTeamName} workspace — Ask ${assistantName} about ${currentTeamName}...`;
    }
    if (simpleMode) {
        return `Describe the outcome you need from ${assistantName} — or tell ${assistantName} what you want to plan, review, create, or execute`;
    }
    return `Describe the outcome you need from ${assistantName} — Ask ${assistantName}... (or /all to broadcast)`;
}
