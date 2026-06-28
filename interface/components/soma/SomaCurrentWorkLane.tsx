"use client";

import { FileText, PanelRightOpen, Radio, Route } from "lucide-react";
import {
  OutputWorkbenchCompactDigest,
  type OutputWorkbenchDigest,
} from "./OutputWorkbenchDigest";

export function SomaCurrentWorkLane({
  digest,
  isPanelOpen,
  onTogglePanel,
  panelId,
  primaryKind,
  recoveryReviewCount = 0,
  reviewCount,
  reviewLabel,
  showOutputDigest,
}: {
  digest: OutputWorkbenchDigest | null;
  isPanelOpen: boolean;
  onTogglePanel: () => void;
  panelId: string;
  primaryKind: "work" | "output";
  recoveryReviewCount?: number;
  reviewCount: number;
  reviewLabel: string;
  showOutputDigest: boolean;
}) {
  const activeLabel = primaryKind === "work" ? "Work needs review" : "Output ready";
  const nextAction = primaryKind === "work" ? "Review work" : "Review output";
  const visibleDigest = digest && showOutputDigest ? digest : null;
  const hasOutputDigest = Boolean(visibleDigest);
  const recoveryHint = primaryKind === "output" && recoveryReviewCount > 0
    ? `${recoveryReviewCount} recovery ${recoveryReviewCount === 1 ? "item" : "items"} also ${recoveryReviewCount === 1 ? "needs" : "need"} review.`
    : null;

  return (
    <section
      className="flex min-w-0 flex-col gap-2 rounded-xl border border-cortex-border bg-cortex-bg/80 px-3 py-2 md:flex-row md:items-center md:justify-between"
      data-testid="soma-current-work-lane"
      aria-label="Current work"
    >
      <div className="flex min-w-0 flex-1 flex-col gap-1 md:flex-row md:items-center md:gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <Route className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
          {primaryKind === "work" ? (
            <Radio className="h-3.5 w-3.5 shrink-0 text-amber-300" />
          ) : (
            <FileText className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
          )}
          <span className="truncate text-sm font-semibold text-cortex-text-main">
            {activeLabel}
          </span>
        </div>
        {recoveryHint ? (
          <p className="truncate text-xs text-amber-200 md:max-w-[18rem]">
            {recoveryHint}
          </p>
        ) : null}

        {visibleDigest ? (
          <div className="min-w-0 md:max-w-[28rem]" data-testid="soma-current-output-slot">
            <OutputWorkbenchCompactDigest digest={visibleDigest} />
          </div>
        ) : null}
      </div>

      <button
        type="button"
        aria-controls={panelId}
        aria-expanded={isPanelOpen}
        data-testid="soma-workbench-panel-toggle"
        onClick={onTogglePanel}
        className="inline-flex min-h-9 shrink-0 items-center justify-center gap-2 rounded-lg border border-cortex-primary/30 bg-cortex-surface px-3 py-1.5 text-xs font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/60"
      >
        <PanelRightOpen className="h-3.5 w-3.5 text-cortex-primary" />
        <span>{isPanelOpen ? "Hide review" : nextAction}</span>
        <span
          className="inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-1.5 text-[10px] font-bold text-cortex-primary"
          aria-label={`${reviewCount} review ${reviewCount === 1 ? "item" : "items"}`}
        >
          {reviewCount}
        </span>
        <span className="sr-only">{reviewLabel}</span>
      </button>
    </section>
  );
}
