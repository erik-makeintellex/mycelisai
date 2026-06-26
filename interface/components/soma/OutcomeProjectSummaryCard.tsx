import Link from "next/link";
import { Boxes, ExternalLink, FolderOpen, Users } from "lucide-react";

export type OutcomeProjectSummary = {
  title: string;
  detail: string;
  ownerLabel?: string;
  leadLabel?: string;
  registryOwnerLabel?: string;
  teamCount: number;
  workCount: number;
  outputCount: number;
  recoveryCount: number;
  href: string;
  hrefLabel: string;
  targetReference?: string;
  outputHref?: string;
  outputLabel?: string;
};

export function OutcomeProjectSummaryCard({ summary }: { summary: OutcomeProjectSummary }) {
  return (
    <section aria-label="Outcome project">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Outcome project</h3>
        <span className="rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-2 py-0.5 text-xs font-semibold text-cortex-primary">
          {summary.teamCount} team{summary.teamCount === 1 ? "" : "s"}
        </span>
      </div>
      <div className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3">
        <div className="flex items-start gap-2">
          <Boxes className="mt-0.5 h-4 w-4 shrink-0 text-cortex-primary" />
          <div className="min-w-0">
            <div className="font-semibold text-cortex-text-main">{summary.title}</div>
            <p className="mt-1 text-sm leading-5 text-cortex-text-muted">{summary.detail}</p>
          </div>
        </div>
        {summary.leadLabel ? (
          <div className="mt-3 flex items-center gap-2 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-muted">
            <Users className="h-3.5 w-3.5 text-cortex-primary" />
            Lead: <span className="font-semibold text-cortex-text-main">{summary.leadLabel}</span>
          </div>
        ) : null}
        {(summary.ownerLabel || summary.registryOwnerLabel) ? (
          <div className="mt-3 space-y-1.5 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-muted">
            {summary.ownerLabel ? (
              <div>
                <span className="font-semibold text-cortex-text-main">OutcomeProject owner:</span>{" "}
                {summary.ownerLabel}
              </div>
            ) : null}
            {summary.registryOwnerLabel ? (
              <div>
                <span className="font-semibold text-cortex-text-main">TeamRegistry owner:</span>{" "}
                {summary.registryOwnerLabel}
              </div>
            ) : null}
          </div>
        ) : null}
        <div className="mt-3 grid grid-cols-3 gap-2 text-center text-xs">
          <Metric label="Work" value={summary.workCount} />
          <Metric label="Outputs" value={summary.outputCount} />
          <Metric label="Recovery" value={summary.recoveryCount} tone={summary.recoveryCount > 0 ? "warning" : "neutral"} />
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <Link
            href={summary.href}
            data-target-reference={summary.targetReference}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary hover:border-cortex-primary/60"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            {summary.hrefLabel}
          </Link>
          {summary.outputHref && summary.outputLabel ? (
            <Link href={summary.outputHref} className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-border px-2.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/35">
              <FolderOpen className="h-3.5 w-3.5" />
              {summary.outputLabel}
            </Link>
          ) : null}
        </div>
      </div>
    </section>
  );
}

function Metric({ label, value, tone = "neutral" }: { label: string; value: number; tone?: "neutral" | "warning" }) {
  const valueClass = tone === "warning" ? "text-cortex-warning" : "text-cortex-text-main";
  return (
    <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2 py-2">
      <div className={`font-semibold ${valueClass}`}>{value}</div>
      <div className="mt-0.5 text-[10px] uppercase tracking-[0.12em] text-cortex-text-muted">{label}</div>
    </div>
  );
}
