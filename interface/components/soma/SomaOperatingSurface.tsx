"use client";

import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { useCortexStore } from "@/store/useCortexStore";
import { SomaCausalSummary } from "./SomaCausalSummary";
import { SomaEvidencePanel, type SomaEvidenceItem } from "./SomaEvidencePanel";
import { SomaHeader } from "./SomaHeader";
import { DEFAULT_SOMA_SUGGESTIONS, type SomaSuggestion } from "./SomaSuggestionBar";

export function SomaOperatingSurface({
  organizationId,
  organizationName,
  activeMode,
  governancePosture,
  evidenceItems,
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
  organizationId?: string;
  organizationName?: string | null;
  activeMode?: string | null;
  governancePosture?: string;
  evidenceItems?: SomaEvidenceItem[];
  suggestions?: readonly SomaSuggestion[];
}) {
  const missionChat = useCortexStore((state) => state.missionChat);
  const evidence = evidenceItems ?? defaultEvidence;

  return (
    <section
      className="overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(148,163,184,0.16)]"
      data-testid="soma-operating-surface"
    >
      <SomaHeader
        organizationName={organizationName}
        activeMode={activeMode}
        governancePosture={governancePosture}
      />
      <div className="space-y-4 p-5">
        <div
          data-testid="central-soma-chat-frame"
          className="h-[72vh] min-h-[560px] max-h-[760px] overflow-hidden rounded-3xl border border-cortex-border bg-cortex-bg"
        >
          <MissionControlChat
            simpleMode
            autoFocus
            organizationId={organizationId}
            suggestions={suggestions}
          />
        </div>
        <SomaCausalSummary messages={missionChat} />
        <SomaEvidencePanel items={evidence} />
      </div>
    </section>
  );
}

const defaultEvidence: SomaEvidenceItem[] = [
  {
    title: "Approval queue",
    detail: "Review gated actions, risk decisions, and pending confirmations.",
    href: "/approvals",
    icon: <CheckSquare className="h-4 w-4" />,
  },
  {
    title: "Activity and runs",
    detail: "See progress, events, and recent outcomes behind Soma actions.",
    href: "/activity",
    icon: <ListChecks className="h-4 w-4" />,
  },
  {
    title: "Learning and context",
    detail: "Inspect retained patterns, artifacts, and continuity evidence.",
    href: "/memory",
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: "Tool readiness",
    detail: "Check connected tools, search configuration, and MCP capability status.",
    href: "/resources?tab=tools",
    icon: <Wrench className="h-4 w-4" />,
  },
];
