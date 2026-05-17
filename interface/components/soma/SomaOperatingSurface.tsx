"use client";

import type React from "react";
import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import type { ChatMessage } from "@/store/useCortexStore";
import { useCortexStore } from "@/store/useCortexStore";
import { OutputWorkbench, outputWorkbenchItems, projectPackageOutputs } from "./OutputWorkbench";
import { SomaCausalSummary } from "./SomaCausalSummary";
import { SomaEvidencePanel, type SomaEvidenceItem } from "./SomaEvidencePanel";
import { SomaHeader } from "./SomaHeader";
import { DEFAULT_SOMA_SUGGESTIONS, type SomaSuggestion } from "./SomaSuggestionBar";
import { SomaWorkspaceFrame } from "./SomaWorkspaceFrame";

function lastSomaMessage(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .find((message) => message.role !== "user" && message.role !== "system");
}

export function SomaOperatingSurface({
  organizationId,
  organizationName,
  activeMode,
  governancePosture,
  evidenceItems,
  activeWorkSlot,
  outputSlot,
  trustSlot,
  contextSlot,
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
  organizationId?: string;
  organizationName?: string | null;
  activeMode?: string | null;
  governancePosture?: string;
  evidenceItems?: SomaEvidenceItem[];
  activeWorkSlot?: React.ReactNode;
  outputSlot?: React.ReactNode;
  trustSlot?: React.ReactNode;
  contextSlot?: React.ReactNode;
  suggestions?: readonly SomaSuggestion[];
}) {
  const missionChat = useCortexStore((state) => state.missionChat);
  const evidence = evidenceItems ?? defaultEvidence;
  const latestSoma = lastSomaMessage(missionChat);
  const outputItems = outputWorkbenchItems(latestSoma?.execution_summary, latestSoma?.artifacts);
  const projectPackages = projectPackageOutputs(latestSoma?.execution_summary?.outputs);

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
      <div className="p-4 lg:p-5">
        <SomaWorkspaceFrame
          expression={(
            <div
              data-testid="central-soma-chat-frame"
              className="h-[68vh] min-h-[500px] max-h-[760px] overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg"
            >
              <MissionControlChat
                simpleMode
                autoFocus
                organizationId={organizationId}
                suggestions={suggestions}
              />
            </div>
          )}
          activeWork={activeWorkSlot ?? <ActiveWorkFallback activeMode={activeMode} />}
          trust={trustSlot ?? <SomaCausalSummary messages={missionChat} />}
          output={outputSlot ?? (
            <OutputWorkbench
              outputs={outputItems}
              projectPackages={projectPackages}
              emptyMessage="Soma has not returned a retained output package yet."
            />
          )}
          context={contextSlot ?? <SomaEvidencePanel items={evidence} />}
        />
      </div>
    </section>
  );
}

function ActiveWorkFallback({ activeMode }: { activeMode?: string | null }) {
  return (
    <div className="rounded-lg border border-cortex-border bg-cortex-surface/70 p-3 text-sm leading-6 text-cortex-text-muted">
      {activeMode
        ? `${activeMode} is the current workspace lane. No active team work is attached to this view yet.`
        : "No active team work is attached to this view yet."}
    </div>
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
