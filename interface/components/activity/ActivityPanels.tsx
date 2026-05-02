"use client";

import { useState } from "react";
import Link from "next/link";
import { RefreshCw, type LucideIcon } from "lucide-react";
import type { MissionRun, StreamSignal } from "@/store/useCortexStore";

export function timeAgo(iso?: string) {
  if (!iso) return "now";
  const diff = Date.now() - new Date(iso).getTime();
  if (!Number.isFinite(diff) || diff < 0) return "now";
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return `${Math.floor(diff / 86_400_000)}d ago`;
}

function runTone(status: MissionRun["status"]) {
  if (status === "completed") return "bg-cortex-success";
  if (status === "failed") return "bg-cortex-danger";
  if (status === "running") return "bg-cortex-primary animate-pulse";
  return "bg-cortex-text-muted/40";
}

function signalCategory(signal: StreamSignal) {
  const text =
    `${signal.type ?? ""} ${signal.message ?? ""} ${signal.source_kind ?? ""} ${signal.payload_kind ?? ""}`.toLowerCase();
  if (text.includes("error") || signal.level?.toLowerCase() === "error")
    return "error";
  if (text.includes("govern") || text.includes("approval")) return "governance";
  if (
    text.includes("tool") ||
    text.includes("mcp") ||
    text.includes("actuation")
  )
    return "tools";
  if (
    text.includes("artifact") ||
    text.includes("output") ||
    text.includes("complete")
  )
    return "output";
  return "status";
}

export function summarizeSignals(signals: StreamSignal[]) {
  return signals.reduce(
    (acc, signal) => {
      acc[signalCategory(signal)] += 1;
      return acc;
    },
    { status: 0, output: 0, tools: 0, governance: 0, error: 0 },
  );
}

export function SummaryCard({
  icon: Icon,
  label,
  value,
  href,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  href: string;
}) {
  return (
    <Link
      href={href}
      aria-label={`${label}: ${value}`}
      className="block rounded-3xl border border-cortex-border bg-cortex-surface p-4 transition hover:border-cortex-primary/40 hover:bg-cortex-primary/5 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
    >
      <div className="flex items-center justify-between gap-3">
        <Icon className="h-5 w-5 text-cortex-primary" />
        <span className="text-xl font-semibold">{value}</span>
      </div>
      <p className="mt-3 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
        {label}
      </p>
    </Link>
  );
}

export function RunList({
  runs,
  selectedRunId,
  isFetching,
  onSelect,
}: {
  runs: MissionRun[];
  selectedRunId: string | null;
  isFetching: boolean;
  onSelect: (runID: string) => void;
}) {
  return (
    <div
      id="workflow-runs"
      className="rounded-3xl border border-cortex-border bg-cortex-surface p-5"
    >
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
            Current workflows
          </p>
          <h2 className="mt-2 text-lg font-semibold">
            Runs needing admin awareness
          </h2>
        </div>
        {isFetching ? (
          <RefreshCw className="h-4 w-4 animate-spin text-cortex-text-muted" />
        ) : null}
      </div>
      <div className="mt-4 overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg">
        {runs.length === 0 ? (
          <p className="px-4 py-8 text-center text-sm text-cortex-text-muted">
            No runs yet.
          </p>
        ) : (
          runs.slice(0, 12).map((run) => (
            <button
              key={run.id}
              type="button"
              onClick={() => onSelect(run.id)}
              className={`flex w-full items-center gap-3 border-b border-cortex-border/60 px-4 py-3 text-left last:border-b-0 hover:bg-cortex-surface ${selectedRunId === run.id ? "bg-cortex-primary/10" : ""}`}
            >
              <span
                className={`h-2.5 w-2.5 rounded-full ${runTone(run.status)}`}
              />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-semibold">
                  {run.id}
                </span>
                <span className="block truncate text-[11px] font-mono text-cortex-text-muted">
                  mission {run.mission_id}
                </span>
              </span>
              <span className="text-[11px] font-mono uppercase text-cortex-text-muted">
                {run.status}
              </span>
              <span className="text-[11px] font-mono text-cortex-text-muted">
                {timeAgo(run.started_at)}
              </span>
            </button>
          ))
        )}
      </div>
    </div>
  );
}

export function BusActivityPanel({
  signals,
  summary,
  natsStatus,
  groupsStatus,
}: {
  signals: StreamSignal[];
  summary: ReturnType<typeof summarizeSignals>;
  natsStatus?: string;
  groupsStatus?: string;
}) {
  return (
    <div
      id="message-bus"
      className="rounded-3xl border border-cortex-border bg-cortex-surface p-5"
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
            Message bus activity
          </p>
          <h2 className="mt-2 text-lg font-semibold">Readable event stream</h2>
        </div>
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-[11px] font-mono text-cortex-text-muted">
          NATS {natsStatus ?? "unknown"} | Groups {groupsStatus ?? "unknown"}
        </div>
      </div>
      <div className="mt-4 grid grid-cols-5 gap-2">
        {Object.entries(summary).map(([label, count]) => (
          <div
            key={label}
            className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2"
          >
            <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
              {label}
            </p>
            <p className="mt-1 text-lg font-semibold">{count}</p>
          </div>
        ))}
      </div>
      <div className="mt-4 max-h-[510px] overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-bg">
        {signals.length === 0 ? (
          <p className="px-4 py-8 text-center text-sm text-cortex-text-muted">
            Waiting for live stream entries.
          </p>
        ) : (
          signals
            .slice(0, 24)
            .map((signal, index) => (
              <SignalRow
                key={`${signal.timestamp ?? "ts"}-${index}`}
                signal={signal}
              />
            ))
        )}
      </div>
    </div>
  );
}

function SignalRow({ signal }: { signal: StreamSignal }) {
  const [expanded, setExpanded] = useState(false);
  const category = signalCategory(signal);
  const message = signal.message?.trim() || signal.topic || "Event received";
  return (
    <article className="border-b border-cortex-border/60 px-4 py-3 last:border-b-0">
      <button
        type="button"
        onClick={() => setExpanded((current) => !current)}
        className="flex w-full items-start justify-between gap-3 text-left"
      >
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm font-semibold">
              {signal.source ?? signal.source_kind ?? "system"}
            </span>
            <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
              {category}
            </span>
          </div>
          <p className="mt-1 truncate text-sm text-cortex-text-muted">
            {message}
          </p>
          {signal.topic ? (
            <p className="mt-1 truncate text-[10px] font-mono text-cortex-text-muted/70">
              {signal.topic}
            </p>
          ) : null}
        </div>
        <span className="shrink-0 text-[11px] font-mono text-cortex-text-muted">
          {timeAgo(signal.timestamp)}
        </span>
      </button>
      <div className="mt-2 flex flex-wrap gap-2">
        {signal.run_id ? (
          <Link
            href={`/runs/${encodeURIComponent(signal.run_id)}`}
            className="rounded-lg border border-cortex-border px-2 py-1 text-[11px] font-mono text-cortex-primary hover:bg-cortex-primary/10"
          >
            Open run
          </Link>
        ) : null}
        <span className="rounded-lg border border-cortex-border px-2 py-1 text-[11px] font-mono text-cortex-text-muted">
          {expanded ? "Hide details" : "Show details"}
        </span>
      </div>
      {expanded ? (
        <pre className="mt-3 max-h-56 overflow-auto whitespace-pre-wrap rounded-xl border border-cortex-border bg-cortex-surface p-3 text-[11px] leading-5 text-cortex-text-muted">
          {JSON.stringify(signal, null, 2)}
        </pre>
      ) : null}
    </article>
  );
}
