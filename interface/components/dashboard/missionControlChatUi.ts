import type { SomaSuggestion } from "@/components/soma/SomaSuggestionBar";
import { councilLabel } from "@/lib/labels";

export function teamSuggestions(teamName: string): SomaSuggestion[] {
  return [
    {
      label: "Plan this team",
      detail: `Shape next steps for ${teamName}.`,
      prompt: `Plan the next move for ${teamName}`,
    },
    {
      label: "Review state",
      detail: "Check current risk and readiness.",
      prompt: `Review the current state of ${teamName}`,
    },
    {
      label: "Schedule review",
      detail: "Ask before making it recurring.",
      prompt: `Schedule a recurring check for ${teamName}`,
    },
    {
      label: "Monitor changes",
      detail: "Keep the lane visible over time.",
      prompt: `Keep monitoring ${teamName} and report changes here`,
    },
    {
      label: "Governed change",
      detail: "Create a proposal before execution.",
      prompt: `Run a governed change for ${teamName}`,
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
}) {
  if (showAdvancedRouting && broadcastMode) return "Broadcast to all teams...";
  if (showAdvancedRouting && directTarget) {
    return `Direct to ${councilLabel(directTarget, assistantName).name}... (or /all to broadcast)`;
  }
  if (currentTeamName) return `Ask ${assistantName} about ${currentTeamName}...`;
  if (simpleMode) {
    return `Tell ${assistantName} what you want to plan, review, create, or execute`;
  }
  return `Ask ${assistantName}... (or /all to broadcast)`;
}
