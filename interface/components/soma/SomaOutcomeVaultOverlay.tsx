"use client";

import { FolderOpen } from "lucide-react";
import { SomaOutcomeVaultPanel, type DashboardRailAlert } from "./SomaOutcomeVaultPanel";
import type { OutcomeProjectSummary } from "./OutcomeProjectSummary";
import type { OutputWorkbenchDigest } from "./OutputWorkbenchDigest";

export function SomaOutcomeVaultHeaderButton({
  attentionCount,
  onOpen,
}: {
  attentionCount: number;
  onOpen: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onOpen}
      className="inline-flex h-8 items-center gap-1.5 rounded-full border border-cortex-primary/35 bg-cortex-primary/10 px-3 text-xs font-semibold text-cortex-primary transition hover:border-cortex-primary/60 hover:bg-cortex-primary/15 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
      aria-label="Open Outcome Vault"
    >
      <FolderOpen className="h-3.5 w-3.5" />
      Outcomes
      {attentionCount > 0 ? (
        <span className="rounded-full bg-cortex-primary/15 px-1.5 py-0.5 text-[10px]">
          {attentionCount}
        </span>
      ) : null}
    </button>
  );
}

export function SomaOutcomeVaultOverlay({
  open,
  operationCount,
  latestOutput,
  projectSummary,
  recoveryCount,
  alerts,
  onClose,
}: {
  open: boolean;
  operationCount: number;
  latestOutput?: OutputWorkbenchDigest | null;
  projectSummary?: OutcomeProjectSummary | null;
  recoveryCount: number;
  alerts: DashboardRailAlert[];
  onClose: () => void;
}) {
  if (!open) {
    return null;
  }

  return (
    <div className="absolute inset-0 z-30 flex justify-end bg-cortex-bg/35 p-3 backdrop-blur-[2px] md:p-4" data-testid="soma-outcome-vault-overlay">
      <SomaOutcomeVaultPanel
        className="h-full w-full max-w-[420px]"
        operationCount={operationCount}
        latestOutput={latestOutput}
        projectSummary={projectSummary}
        recoveryCount={recoveryCount}
        alerts={alerts}
        collapsed={false}
        onCollapsedChange={onClose}
        closeLabel="Close Outcome Vault"
      />
    </div>
  );
}
