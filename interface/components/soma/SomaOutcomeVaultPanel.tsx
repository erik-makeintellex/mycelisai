"use client";

import Link from "next/link";
import { FolderOpen, Radio } from "lucide-react";
import type { OutputWorkbenchDigest } from "./OutputWorkbenchDigest";
import { OutputWorkbenchCompactDigest } from "./OutputWorkbenchDigest";

export function SomaOutcomeVaultPanel({
  operationCount,
  latestOutput,
  recoveryCount = 0,
}: {
  operationCount: number;
  latestOutput?: OutputWorkbenchDigest | null;
  recoveryCount?: number;
}) {
  return (
    <aside
      className="flex min-h-[360px] min-w-0 flex-col overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(0,0,0,0.18)]"
      aria-label="Outcomes and Vault"
      data-testid="soma-outcome-vault"
    >
      <div className="border-b border-cortex-border px-5 py-4">
        <h2 className="text-xl font-semibold tracking-tight text-cortex-text-main">Outcomes & Vault</h2>
        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
          Background work and retained outputs stay here without crowding the Soma thread.
        </p>
      </div>
      <div className="min-h-0 flex-1 space-y-6 overflow-y-auto px-5 py-5">
        <section>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Running in background</h3>
            {operationCount > 0 ? (
              <span className="rounded-full border border-cortex-success/35 bg-cortex-success/10 px-2 py-0.5 text-xs font-semibold text-cortex-success">
                {operationCount}
              </span>
            ) : null}
          </div>
          <div className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <div className="flex items-center gap-2 font-semibold text-cortex-text-main">
              <Radio className={`h-4 w-4 ${operationCount > 0 ? "text-cortex-success" : "text-cortex-text-muted"}`} />
              {operationCount > 0
                ? `${operationCount} item${operationCount === 1 ? " needs" : "s need"} review`
                : "No background work running"}
            </div>
            <p className="mt-1 text-sm leading-5 text-cortex-text-muted">
              {operationCount > 0
                ? "Open the review lane in Soma to recover, approve, or inspect work."
                : "Quick actions and approved Soma work will appear here while they run."}
            </p>
            {recoveryCount > 0 ? (
              <p className="mt-2 text-xs font-semibold text-cortex-warning">
                {recoveryCount} recovery item{recoveryCount === 1 ? "" : "s"} also need attention.
              </p>
            ) : null}
          </div>
        </section>
        <section>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Recent deliverables</h3>
            <Link href="/resources?tab=workspace" className="text-xs font-semibold text-cortex-primary hover:underline">
              Open vault
            </Link>
          </div>
          {latestOutput ? (
            <OutputWorkbenchCompactDigest digest={latestOutput} />
          ) : (
            <div className="flex items-center justify-between gap-4 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3">
              <div>
                <div className="font-semibold text-cortex-text-main">No retained deliverables yet</div>
                <div className="text-sm text-cortex-text-muted">Ask Soma to create or review something and save the output.</div>
              </div>
              <FolderOpen className="h-5 w-5 shrink-0 text-cortex-text-muted" />
            </div>
          )}
        </section>
      </div>
    </aside>
  );
}
