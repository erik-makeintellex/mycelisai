import Link from "next/link";
import { ExternalLink, FolderOpen } from "lucide-react";
import type { OutcomeProjectSummary } from "./OutcomeProjectSummary";

export function DeliveredOutcomeSummary({ summary }: { summary: OutcomeProjectSummary }) {
  const primaryLabel = summary.recoveryCount > 0 ? "Review recovery" : "Revisit work";
  return (
    <section aria-label="Delivered outcome">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Outcome ready to revisit</h3>
        {summary.recoveryCount > 0 ? (
          <span className="rounded-full border border-cortex-warning/45 bg-cortex-warning/10 px-2 py-0.5 text-xs font-semibold text-cortex-warning">
            Recovery attention
          </span>
        ) : null}
      </div>
      <div className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3">
        <div className="font-semibold text-cortex-text-main">{summary.title}</div>
        <p className="mt-1 text-sm leading-5 text-cortex-text-muted">{summary.detail}</p>
        <div className="mt-3 grid grid-cols-3 gap-2 text-center text-xs">
          <OutcomeMetric label="Background" value={summary.workCount} />
          <OutcomeMetric label="Saved" value={summary.outputCount} />
          <OutcomeMetric label="Recovery" value={summary.recoveryCount} tone={summary.recoveryCount > 0 ? "warning" : "neutral"} />
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <Link
            href={summary.href}
            data-target-reference={summary.targetReference}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary hover:border-cortex-primary/60"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            {primaryLabel}
          </Link>
          {summary.outputHref ? (
            <Link href={summary.outputHref} className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-border px-2.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/35">
              <FolderOpen className="h-3.5 w-3.5" />
              Open saved outcomes
            </Link>
          ) : null}
        </div>
      </div>
    </section>
  );
}

function OutcomeMetric({ label, value, tone = "neutral" }: { label: string; value: number; tone?: "neutral" | "warning" }) {
  const valueClass = tone === "warning" ? "text-cortex-warning" : "text-cortex-text-main";
  return (
    <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2 py-2">
      <div className={`font-semibold ${valueClass}`}>{value}</div>
      <div className="mt-0.5 text-[10px] uppercase tracking-[0.12em] text-cortex-text-muted">{label}</div>
    </div>
  );
}
