"use client";

import Link from "next/link";
import { ExternalLink, FolderOpen, PanelRightClose, PanelRightOpen, Radio } from "lucide-react";
import { resourcesWorkspaceHref } from "@/lib/outputPackageModel";
import { OutcomeProjectSummaryCard, type OutcomeProjectSummary } from "./OutcomeProjectSummaryCard";
import type { OutputWorkbenchDigest } from "./OutputWorkbenchDigest";
import OutputAccessActions, { workspacePathFromOutputUrl } from "./OutputAccessActions";

const WORK_REVIEW_LANE_HREF = "/teams?view=work";

export type DashboardRailAlertTarget = {
  type: "run" | "work" | "recovery" | "capability" | "output" | "outcome_project";
  id: string;
  href: string;
  label: string;
};

export type DashboardRailAlert = {
  id: string;
  kind: "work_review" | "recovery" | "output_ready" | "run_failed" | "capability_review";
  severity: "info" | "success" | "warning" | "danger";
  title: string;
  detail?: string;
  count?: number;
  href: string;
  actionLabel: string;
  targetReference: string;
  target?: DashboardRailAlertTarget;
  updatedAt?: string | null;
};

export function SomaOutcomeVaultPanel({
  operationCount,
  latestOutput,
  projectSummary,
  recoveryCount = 0,
  alerts = [],
  className = "",
  collapsed = false,
  onCollapsedChange,
  closeLabel = "Collapse Outcomes and Vault",
}: {
  operationCount: number;
  latestOutput?: OutputWorkbenchDigest | null;
  projectSummary?: OutcomeProjectSummary | null;
  recoveryCount?: number;
  alerts?: DashboardRailAlert[];
  className?: string;
  collapsed?: boolean;
  onCollapsedChange?: (collapsed: boolean) => void;
  closeLabel?: string;
}) {
  const attentionCount = operationCount + recoveryCount + (latestOutput ? 1 : 0) + (projectSummary ? 1 : 0);
  const hasWorkReviewTarget = operationCount > 0 || recoveryCount > 0;
  const latestOutputTarget = latestOutput ? deliverableTarget(latestOutput) : null;
  const workAlerts = alerts.filter((alert) => alert.kind !== "output_ready").slice(0, 3);

  if (collapsed) {
    return (
      <aside
        className={`flex min-h-[360px] min-w-0 flex-col items-center gap-3 overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface px-2 py-4 shadow-[0_18px_40px_rgba(0,0,0,0.18)] ${className}`}
        aria-label="Outcomes and Vault collapsed"
        data-state="collapsed"
        data-testid="soma-outcome-vault"
      >
        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-cortex-border bg-cortex-bg text-cortex-primary transition hover:border-cortex-primary/40 hover:bg-cortex-primary/10 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
          aria-label="Open Outcomes and Vault"
          title="Open Outcomes and Vault"
          onClick={() => onCollapsedChange?.(false)}
        >
          <PanelRightOpen className="h-4 w-4" />
        </button>
        <div className="flex flex-col items-center gap-2 text-cortex-text-muted" aria-hidden="true">
          <FolderOpen className="h-5 w-5" />
          {attentionCount > 0 ? (
            <span className="rounded-full border border-cortex-success/35 bg-cortex-success/10 px-2 py-0.5 text-xs font-semibold text-cortex-success">
              {attentionCount}
            </span>
          ) : null}
        </div>
      </aside>
    );
  }

  return (
    <aside
      className={`flex min-h-[360px] min-w-0 flex-col overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(0,0,0,0.18)] ${className}`}
      aria-label="Outcomes and Vault"
      data-state="expanded"
      data-testid="soma-outcome-vault"
    >
      <div className="border-b border-cortex-border px-5 py-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h2 className="text-xl font-semibold tracking-tight text-cortex-text-main">Outcomes & Vault</h2>
            <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
              Background work and retained outputs stay here without crowding the Soma thread.
            </p>
          </div>
          <button
            type="button"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-cortex-border bg-cortex-bg text-cortex-text-muted transition hover:border-cortex-primary/40 hover:text-cortex-primary focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
            aria-label={closeLabel}
            title={closeLabel}
            onClick={() => onCollapsedChange?.(true)}
          >
            <PanelRightClose className="h-4 w-4" />
          </button>
        </div>
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
            {hasWorkReviewTarget ? (
              <div className="mt-3 flex flex-wrap gap-2">
                <Link
                  href={WORK_REVIEW_LANE_HREF}
                  data-target-reference={WORK_REVIEW_LANE_HREF}
                  aria-label="Open work needing attention"
                  className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
                >
                  <ExternalLink className="h-3.5 w-3.5" />
                  Review items
                </Link>
              </div>
            ) : null}
            {workAlerts.length > 0 ? (
              <div className="mt-3 space-y-2" aria-label="Referenced work alerts">
                {workAlerts.map((alert) => (
                  <RailAlertLink key={alert.id} alert={alert} />
                ))}
              </div>
            ) : null}
          </div>
        </section>
        {projectSummary ? <OutcomeProjectSummaryCard summary={projectSummary} /> : null}
        <section>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Recent deliverables</h3>
            <Link href="/resources?tab=workspace" className="text-xs font-semibold text-cortex-primary hover:underline">
              Open vault
            </Link>
          </div>
          {latestOutput ? (
            <div className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3" aria-label="Recent deliverable">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  {latestOutputTarget ? (
                    <a
                      href={latestOutputTarget.href}
                      data-target-reference={latestOutputTarget.reference}
                      target={latestOutputTarget.external ? "_blank" : undefined}
                      rel={latestOutputTarget.external ? "noopener noreferrer" : undefined}
                      className="inline-flex max-w-full items-center gap-1.5 font-semibold text-cortex-primary hover:underline focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
                      aria-label={`Open latest deliverable ${latestOutput.text}`}
                    >
                      <span className="truncate">{latestOutput.text}</span>
                      <ExternalLink className="h-3.5 w-3.5 shrink-0" />
                    </a>
                  ) : (
                    <div className="font-semibold text-cortex-text-main">{latestOutput.text}</div>
                  )}
                  {deliverablePath(latestOutput) ? (
                    <code className="mt-1 block max-w-64 truncate font-mono text-xs text-cortex-text-muted">
                      {deliverablePath(latestOutput)}
                    </code>
                  ) : (
                    <div className="mt-1 text-sm text-cortex-text-muted">
                      Saved output ready to revisit.
                    </div>
                  )}
                </div>
                <OutputAccessActions
                  label={latestOutput.text}
                  url={latestOutput.url}
                  storagePath={latestOutput.storagePath}
                  openLabel="Open"
                  folderLabel="Folder"
                />
              </div>
              {latestOutput.count > 1 ? (
                <div className="mt-2 text-xs text-cortex-text-muted">
                  {latestOutput.count} retained item{latestOutput.count === 1 ? "" : "s"} in this outcome.
                </div>
              ) : null}
            </div>
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

function RailAlertLink({ alert }: { alert: DashboardRailAlert }) {
  return (
    <Link
      href={alert.href}
      data-alert-id={alert.id}
      data-alert-kind={alert.kind}
      data-target-reference={alert.targetReference}
      data-target-type={alert.target?.type}
      data-target-id={alert.target?.id}
      aria-label={`${alert.actionLabel}: ${alert.title}${alert.target?.label ? ` (${alert.target.label})` : ""}`}
      className={`block rounded-lg border px-3 py-2 text-left text-xs transition focus:outline-none focus:ring-2 ${alertClassName(alert.severity)}`}
    >
      <span className="block font-semibold">{alert.title}</span>
      {alert.detail ? (
        <span className="mt-0.5 block leading-4 opacity-80">{alert.detail}</span>
      ) : null}
      <span className="mt-1 inline-flex items-center gap-1 font-semibold">
        <ExternalLink className="h-3.5 w-3.5" />
        {alert.actionLabel}
      </span>
    </Link>
  );
}

function alertClassName(severity: DashboardRailAlert["severity"]) {
  if (severity === "success") return "border-cortex-success/35 bg-cortex-success/10 text-cortex-success focus:ring-cortex-success/40";
  if (severity === "warning") return "border-cortex-warning/45 bg-cortex-warning/10 text-cortex-warning focus:ring-cortex-warning/40";
  if (severity === "danger") return "border-cortex-danger/45 bg-cortex-danger/10 text-cortex-danger focus:ring-cortex-danger/40";
  return "border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary focus:ring-cortex-primary/40";
}

function deliverablePath(output: OutputWorkbenchDigest) {
  return output.storagePath?.trim() || workspacePathFromOutputUrl(output.url);
}

function deliverableTarget(output: OutputWorkbenchDigest) {
  const url = output.url?.trim();
  const path = deliverablePath(output);
  if (url) {
    return {
      href: url,
      reference: path || url,
      external: /^(https?:)?\/\//i.test(url),
    };
  }
  const resourcesHref = resourcesWorkspaceHref(path);
  return resourcesHref ? {
    href: resourcesHref,
    reference: path ?? resourcesHref,
    external: false,
  } : null;
}
