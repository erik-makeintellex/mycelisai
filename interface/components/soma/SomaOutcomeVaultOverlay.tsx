"use client";

import { FolderOpen } from "lucide-react";
import { useEffect, useRef } from "react";
import { SomaOutcomeVaultPanel, type DashboardRailAlert } from "./SomaOutcomeVaultPanel";
import type { OutcomeProjectSummary } from "./OutcomeProjectSummary";
import type { OutputWorkbenchDigest } from "./OutputWorkbenchDigest";

export function SomaOutcomeVaultHeaderButton({
  attentionCount,
  expanded = false,
  onOpen,
}: {
  attentionCount: number;
  expanded?: boolean;
  onOpen: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onOpen}
      className="inline-flex h-8 items-center gap-1.5 rounded-full border border-cortex-primary/35 bg-cortex-primary/10 px-3 text-xs font-semibold text-cortex-primary transition hover:border-cortex-primary/60 hover:bg-cortex-primary/15 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
      aria-label="Open Outcome Vault"
      aria-expanded={expanded}
      aria-haspopup="dialog"
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
  const dialogRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;

    previousFocusRef.current = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;
    const closeButton = dialogRef.current?.querySelector<HTMLButtonElement>(
      'button[aria-label="Close Outcome Vault"]',
    );
    closeButton?.focus();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
        return;
      }
      if (event.key !== "Tab") return;

      const focusable = Array.from(dialogRef.current?.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
      ) ?? []);
      if (focusable.length === 0) {
        event.preventDefault();
        dialogRef.current?.focus();
        return;
      }
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement;
      if (event.shiftKey && (active === first || !dialogRef.current?.contains(active))) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && (active === last || !dialogRef.current?.contains(active))) {
        event.preventDefault();
        first.focus();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      previousFocusRef.current?.focus();
    };
  }, [open, onClose]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-y-0 right-0 left-[68px] z-30 flex justify-end p-0 sm:left-0 sm:p-3 md:p-4 lg:absolute lg:inset-0" data-testid="soma-outcome-vault-overlay">
      <button
        type="button"
        className="absolute inset-0 cursor-default bg-cortex-bg/35 backdrop-blur-[2px]"
        aria-label="Close Outcome Vault backdrop"
        onClick={onClose}
      />
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label="Outcome Vault"
        tabIndex={-1}
        className="relative z-10 h-full w-full sm:max-w-[420px]"
      >
        <SomaOutcomeVaultPanel
          className="h-full w-full rounded-none sm:rounded-3xl"
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
    </div>
  );
}
