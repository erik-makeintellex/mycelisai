import { councilLabel } from "@/lib/labels";
import type { SomaSuggestion } from "@/components/soma/SomaSuggestionBar";

export function teamSuggestions(teamName: string, teamId?: string): SomaSuggestion[] {
    return [
        {
            label: "Plan & Execute Next Steps",
            detail: `Shape the execution plan and action matrix for ${teamName}.`,
            prompt: `Plan for team ${teamName}${teamId ? ` (team_id: ${teamId})` : ""} and execute next steps to produce its execution plan, action matrix, and handoff deliverable package.`,
        },
        {
            label: "Build Interacting Companion App",
            detail: `Have another team build an app that interacts with ${teamName}.`,
            prompt: `Have a second team work on another application that interacts with ${teamName}${teamId ? ` (team_id: ${teamId})` : ""} and connect their live outputs.`,
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

export function buildTeamActionPrefillPrompt(searchParams: URLSearchParams): string {
    const chainChannel = searchParams.get("chain_channel")?.trim();
    if (chainChannel) {
        return `Chain the output published on channel ${chainChannel} into a downstream team that listens on ${chainChannel} and generates the next deliverable package.`;
    }
    const teamAction = searchParams.get("team_action")?.trim();
    if (teamAction === "add_mcp") {
        const teamId = searchParams.get("team_id")?.trim() || "mission-delivery-team";
        const teamName = searchParams.get("team_name")?.trim() || teamId;
        const mcpId = searchParams.get("mcp_id")?.trim() || "filesystem";
        const mcpTitle = searchParams.get("mcp_title")?.trim() || mcpId;
        const mcpGroup = searchParams.get("mcp_group")?.trim() === "custom" ? "custom" : "standard";
        const groupLabel = mcpGroup === "custom" ? "Custom Developed MCP" : "Standard Deployment MCP";
        return `Add ${groupLabel} ${mcpTitle} (mcp_id: ${mcpId}, mcp_group: ${mcpGroup}, mcp_title: ${mcpTitle}) to team ${teamName} (team_id: ${teamId}) as an MCP addition and generate its deployment binding manifest.`;
    }
    if (teamAction === "companion_app") {
        const teamId = searchParams.get("team_id")?.trim() || "primary-core-app-team";
        const teamName = searchParams.get("team_name")?.trim() || teamId;
        return `Have a team work on another application that interacts with ${teamName} (team_id: ${teamId}) and connect their live outputs.`;
    }
    if (teamAction === "plan_execute" || (teamAction && searchParams.get("team_id")?.trim())) {
        const teamId = searchParams.get("team_id")?.trim() || "mission-delivery-team";
        const teamName = searchParams.get("team_name")?.trim() || teamId;
        return `Plan for team ${teamName} (team_id: ${teamId}) and execute next steps to produce its execution plan, action matrix, and handoff deliverable package.`;
    }
    return "";
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
