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
  reviewCount,
  reviewLabel,
  showOutputDigest,
}: {
  digest: OutputWorkbenchDigest | null;
  isPanelOpen: boolean;
  onTogglePanel: () => void;
  panelId: string;
  primaryKind: "work" | "output";
  reviewCount: number;
  reviewLabel: string;
  showOutputDigest: boolean;
}) {
  const activeLabel = primaryKind === "work" ? "Work needs review" : "Output ready";
  const nextAction = primaryKind === "work" ? "Review work" : "Review output";

  return (
    <section
      className="grid gap-2 rounded-xl border border-cortex-border bg-cortex-bg/85 p-2.5 md:grid-cols-[minmax(0,0.9fr)_minmax(0,1.2fr)_auto] md:items-center"
      data-testid="soma-current-work-lane"
      aria-label="Current work"
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2 font-mono text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-primary">
          <Route className="h-3.5 w-3.5" />
          Current workflow
        </div>
        <div className="mt-1 flex min-w-0 items-center gap-2">
          {primaryKind === "work" ? (
            <Radio className="h-3.5 w-3.5 shrink-0 text-amber-300" />
          ) : (
            <FileText className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
          )}
          <span className="truncate text-sm font-semibold text-cortex-text-main">
            {activeLabel}
          </span>
        </div>
      </div>

      <div className="min-w-0" data-testid="soma-current-output-slot">
        {digest && showOutputDigest ? (
          <OutputWorkbenchCompactDigest digest={digest} />
        ) : (
          <div className="rounded-lg border border-cortex-border bg-cortex-surface/75 px-3 py-2">
            <div className="font-mono text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
              Latest output
            </div>
            <div className="mt-0.5 truncate text-xs font-semibold text-cortex-text-main">
              Not produced yet
            </div>
          </div>
        )}
      </div>

      <button
        type="button"
        aria-controls={panelId}
        aria-expanded={isPanelOpen}
        data-testid="soma-workbench-panel-toggle"
        onClick={onTogglePanel}
        className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-cortex-primary/30 bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/60"
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
