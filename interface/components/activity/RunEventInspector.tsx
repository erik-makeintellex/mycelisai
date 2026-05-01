"use client";

import Link from "next/link";
import { FileJson, RefreshCw } from "lucide-react";
import type { MissionEvent, MissionRun } from "@/store/useCortexStore";

function timeAgo(iso?: string) {
  if (!iso) return "now";
  const diff = Date.now() - new Date(iso).getTime();
  if (!Number.isFinite(diff) || diff < 0) return "now";
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return `${Math.floor(diff / 86_400_000)}d ago`;
}

function payloadPreview(payload?: Record<string, unknown>) {
  if (!payload || Object.keys(payload).length === 0) return "No payload";
  try {
    return JSON.stringify(payload, null, 2);
  } catch {
    return "Payload could not be rendered";
  }
}

function eventTitle(event: MissionEvent) {
  return event.event_type?.trim() || "run.event";
}

export function RunEventInspector({
  run,
  events,
  isFetching,
}: {
  run: MissionRun | null;
  events: MissionEvent[];
  isFetching: boolean;
}) {
  const scopedEvents = run
    ? events.filter((event) => event.run_id === run.id)
    : [];

  return (
    <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
            Run events
          </p>
          <h2 className="mt-2 text-lg font-semibold">
            {run ? run.id : "Select a run"}
          </h2>
          {run ? (
            <p className="mt-1 text-xs font-mono uppercase text-cortex-text-muted">
              {run.status} | mission {run.mission_id}
            </p>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          {isFetching ? (
            <RefreshCw className="h-4 w-4 animate-spin text-cortex-text-muted" />
          ) : null}
          {run ? (
            <Link
              href={`/runs/${run.id}`}
              className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-[11px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted hover:border-cortex-primary/30 hover:text-cortex-primary"
            >
              Full timeline
            </Link>
          ) : null}
        </div>
      </div>

      <div className="mt-4 max-h-[430px] overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-bg">
        {!run ? (
          <p className="px-4 py-8 text-center text-sm text-cortex-text-muted">
            Choose a run to inspect its events.
          </p>
        ) : scopedEvents.length === 0 ? (
          <p className="px-4 py-8 text-center text-sm text-cortex-text-muted">
            No events recorded for this run yet.
          </p>
        ) : (
          scopedEvents.map((event) => (
            <InlineEvent key={event.id} event={event} />
          ))
        )}
      </div>
    </section>
  );
}

function InlineEvent({ event }: { event: MissionEvent }) {
  return (
    <article className="border-b border-cortex-border/60 px-4 py-3 last:border-b-0">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm font-semibold">{eventTitle(event)}</span>
            <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
              {event.severity || "info"}
            </span>
          </div>
          <p className="mt-1 text-[11px] font-mono text-cortex-text-muted">
            {event.source_agent || event.source_team || "system"} |{" "}
            {timeAgo(event.emitted_at)}
          </p>
        </div>
        <FileJson className="h-4 w-4 shrink-0 text-cortex-primary" />
      </div>
      <pre className="mt-3 max-h-48 overflow-auto whitespace-pre-wrap rounded-2xl border border-cortex-border bg-cortex-surface px-3 py-2 text-[11px] leading-5 text-cortex-text-muted">
        {payloadPreview(event.payload)}
      </pre>
    </article>
  );
}
