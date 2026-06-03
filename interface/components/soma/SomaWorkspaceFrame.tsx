"use client";

import React, { useState } from "react";
import Link from "next/link";
import { Boxes, FileText, PanelRightOpen, Radio, ShieldCheck, Sparkles, X } from "lucide-react";
import {
  OutputWorkbenchCompactDigest,
  outputWorkbenchDigest,
} from "./OutputWorkbenchDigest";
import { OutputWorkbench } from "./OutputWorkbench";

type PanelKey = "work" | "output" | "trust" | "context";

function SlotPanel({
  icon,
  label,
  description,
  children,
  className = "",
  showHeader = true,
}: {
  icon: React.ReactNode;
  label: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
  showHeader?: boolean;
}) {
  return (
    <section className={`min-w-0 rounded-xl border border-cortex-border bg-cortex-bg p-2.5 ${className}`}>
      {showHeader ? (
        <div className="mb-2 flex items-center gap-2 text-[10px] font-mono font-bold uppercase tracking-[0.16em] text-cortex-text-muted">
          <span className="text-cortex-primary">{icon}</span>
          {label}
        </div>
      ) : null}
      {description ? (
        <p className="sr-only">{description}</p>
      ) : null}
      {children}
    </section>
  );
}

export function SomaWorkspaceFrame({
  expression,
  activeWork,
  output,
  trust,
  context,
  primaryPanel,
  reviewCount,
  showOutputDigest = true,
}: {
  expression: React.ReactNode;
  activeWork?: React.ReactNode;
  output?: React.ReactNode;
  trust?: React.ReactNode;
  context?: React.ReactNode;
  primaryPanel?: Extract<PanelKey, "work" | "output">;
  reviewCount?: number;
  showOutputDigest?: boolean;
}) {
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [activePanel, setActivePanel] = useState<PanelKey>("work");
  const sideRailId = "soma-workbench-panel";
  const panels = [
    {
      key: "work" as const,
      icon: <Radio className="h-3.5 w-3.5" />,
      label: "Work",
      title: "Work to review",
      description: "Work that needs a decision, check, or follow-up.",
      href: "/teams",
      content: activeWork,
    },
    {
      key: "output" as const,
      icon: <FileText className="h-3.5 w-3.5" />,
      label: "Output",
      title: "Output",
      description: "Files, media, and packages Soma can open for you.",
      href: "/resources?tab=workspace",
      content: output,
    },
    {
      key: "trust" as const,
      icon: <ShieldCheck className="h-3.5 w-3.5" />,
      label: "Trust",
      title: "Trust",
      description: "Proof, recovery, and the safe next step.",
      href: "/activity",
      content: trust,
    },
    {
      key: "context" as const,
      icon: <Boxes className="h-3.5 w-3.5" />,
      label: "Context",
      title: "Context",
      description: "Tools, saved context, and setup.",
      href: "/resources",
      content: context,
    },
  ].filter((panel) => Boolean(panel.content));
  const outputPanel = panels.find((panel) => panel.key === "output");
  const outputDigest = digestFromOutputNode(output);
  const primaryPanelKey = primaryPanel ?? (outputPanel ? "output" : "work");
  const primaryReviewPanel = panels.find((panel) => panel.key === primaryPanelKey) ?? outputPanel ?? panels[0];
  const selectedPanel = panels.find((panel) => panel.key === activePanel) ?? outputPanel ?? panels[0];
  const hasPanels = panels.length > 0;
  const visibleReviewCount = reviewCount ?? outputDigest?.count ?? panels.length;
  const reviewLabel = primaryReviewPanel?.key === "output" ? "Review output" : "Review work";

  const togglePanel = () => {
    setIsPanelOpen((open) => {
      if (!open && primaryReviewPanel) {
        setActivePanel(primaryReviewPanel.key);
      }
      return !open;
    });
  };

  return (
    <div
      className="relative grid gap-3 lg:h-[calc(100vh-260px)] lg:min-h-[420px] 2xl:min-h-[560px]"
      data-testid="soma-workspace-frame"
    >
      <div className="min-h-0 min-w-0">
        <SlotPanel
          icon={<Sparkles className="h-3.5 w-3.5" />}
          label="Soma"
          description="Ask Soma for the next plan, change, file, decision, or follow-up."
          className="flex h-full min-h-0 flex-col"
          showHeader={false}
        >
          {expression}
        </SlotPanel>
      </div>
      {hasPanels ? (
        <>
          <div className="absolute right-3 top-3 z-30 flex max-w-[min(92vw,620px)] flex-col items-end gap-2 sm:flex-row sm:items-start">
            {outputDigest && showOutputDigest && !isPanelOpen ? <OutputWorkbenchCompactDigest digest={outputDigest} /> : null}
            <button
              type="button"
              aria-controls={sideRailId}
              aria-expanded={isPanelOpen}
              data-testid="soma-workbench-panel-toggle"
              onClick={togglePanel}
              className="inline-flex items-center gap-2 rounded-lg border border-cortex-primary/30 bg-cortex-surface/95 px-3 py-2 text-xs font-semibold text-cortex-text-main shadow-lg shadow-black/10 backdrop-blur transition-colors hover:border-cortex-primary/60"
            >
              <PanelRightOpen className="h-3.5 w-3.5 text-cortex-primary" />
              <span>{isPanelOpen ? "Hide review" : reviewLabel}</span>
              <span
                className="inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-1.5 text-[10px] font-bold text-cortex-primary"
                aria-label={`${visibleReviewCount} review ${visibleReviewCount === 1 ? "item" : "items"}`}
              >
                {visibleReviewCount}
              </span>
            </button>
          </div>
          <div
            id={sideRailId}
            aria-hidden={!isPanelOpen}
            className={`fixed inset-y-0 right-0 z-40 flex w-[min(92vw,440px)] min-w-0 flex-col overflow-hidden border-l border-cortex-border bg-cortex-surface/95 p-3 shadow-2xl shadow-black/20 backdrop-blur transition duration-200 lg:absolute lg:inset-y-3 lg:right-3 lg:rounded-2xl lg:border ${
              isPanelOpen ? "translate-x-0 opacity-100" : "pointer-events-none translate-x-full opacity-0"
            }`}
            data-testid="soma-workbench-side-rail"
            tabIndex={0}
          >
        <div className="shrink-0 rounded-xl border border-cortex-border bg-cortex-bg/95 p-2 backdrop-blur">
          <div className="flex items-center justify-between gap-2 px-1">
            <div>
              <p className="font-mono text-[11px] font-bold uppercase tracking-[0.16em] text-cortex-text-muted">
                Review
              </p>
              <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                Check the most useful details here. Open the full page for more.
              </p>
            </div>
            <button
              type="button"
              onClick={() => setIsPanelOpen(false)}
              className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-cortex-border text-cortex-text-main hover:border-cortex-primary/40"
              aria-label="Close work panel"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
          <div
            className="mt-3 grid grid-cols-4 gap-1"
            role="tablist"
            aria-label="Soma review sections"
          >
            {panels.map((panel) => (
              <button
                key={panel.key}
                type="button"
                role="tab"
                aria-selected={selectedPanel?.key === panel.key}
                onClick={() => setActivePanel(panel.key)}
                className={`inline-flex items-center justify-center gap-1 rounded-lg border px-2 py-1.5 text-[10px] font-semibold uppercase tracking-[0.08em] transition-colors ${
                  selectedPanel?.key === panel.key
                    ? "border-cortex-primary/40 bg-cortex-primary/15 text-cortex-primary"
                    : "border-cortex-border text-cortex-text-muted hover:border-cortex-primary/30 hover:text-cortex-text-main"
                }`}
              >
                <span className="hidden sm:inline">{panel.label}</span>
                <span className="sm:hidden">{panel.icon}</span>
              </button>
            ))}
          </div>
        </div>
        {selectedPanel ? (
          <div
            className="mt-3 min-h-0 flex-1 overflow-y-auto overscroll-contain pr-1"
            data-testid="soma-workbench-panel-scroll"
          >
            <SlotPanel
              icon={selectedPanel.icon}
              label={selectedPanel.title}
              description={selectedPanel.description}
            >
              <div className="mb-3 flex items-start justify-between gap-3">
                <p className="text-xs leading-5 text-cortex-text-muted">
                  {selectedPanel.description}
                </p>
                <Link
                  href={selectedPanel.href}
                  className="shrink-0 rounded-lg border border-cortex-border px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.08em] text-cortex-text-main hover:border-cortex-primary/40"
                >
                  Open page
                </Link>
              </div>
              {selectedPanel.content}
            </SlotPanel>
          </div>
        ) : (
          <div className="mt-3 rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm text-cortex-text-muted">
            No review content is available yet.
          </div>
        )}
          </div>
        </>
      ) : null}
    </div>
  );
}

function digestFromOutputNode(output?: React.ReactNode) {
  if (!React.isValidElement(output) || output.type !== OutputWorkbench) return null;
  return outputWorkbenchDigest(output.props as React.ComponentProps<typeof OutputWorkbench>);
}
