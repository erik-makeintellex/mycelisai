"use client";

import { useState } from "react";
import { Activity, AlertTriangle, CheckCircle2, ChevronDown, ChevronRight, FileText, RotateCcw, ShieldCheck } from "lucide-react";
import type { MissionEvent } from "@/store/useCortexStore";

type ReceiptStatus = "running" | "completed" | "failed";

type Receipt = {
  status: ReceiptStatus;
  headline: string;
  result: string;
  trust: string;
  next: string;
  outputRefs: string[];
  proofRefs: string[];
  failure?: string;
};

const TERMINAL_EVENTS = new Set(["mission.completed", "mission.failed", "mission.cancelled"]);

function text(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function compactRef(value: string) {
  if (value.length <= 28) return value;
  return `${value.slice(0, 18)}...${value.slice(-6)}`;
}

function payloadText(event: MissionEvent, keys: string[]) {
  const payload = event.payload ?? {};
  for (const key of keys) {
    const value = text(payload[key]);
    if (value) return value;
  }
  return undefined;
}

function unique(values: Array<string | undefined>) {
  return [...new Set(values.filter(Boolean) as string[])];
}

export function buildRunReceipt(events: MissionEvent[], runId: string): Receipt {
  const terminal = [...events].reverse().find((event) => TERMINAL_EVENTS.has(event.event_type));
  const failedEvent = events.find((event) => event.event_type === "mission.failed" || event.event_type === "tool.failed");
  const completed = terminal?.event_type === "mission.completed";
  const failed = Boolean(failedEvent) || terminal?.event_type === "mission.failed";
  const status: ReceiptStatus = completed ? "completed" : failed ? "failed" : "running";
  const outputEvents = events.filter((event) => /artifact|output|file|media/i.test(event.event_type));
  const proofEvents = events.filter((event) => /proof|audit|completed|failed/i.test(event.event_type) || event.audit_event_id);
  const outputRefs = unique(
    outputEvents.flatMap((event) => {
      const payload = event.payload ?? {};
      return [
        text(payload.title),
        text(payload.path),
        text(payload.storage_ref),
        text(payload.artifact_id),
        text(payload.output_id),
        text(payload.url),
      ];
    }),
  );
  const proofRefs = unique(
    proofEvents.flatMap((event) => {
      const payload = event.payload ?? {};
      return [
        event.audit_event_id,
        text(payload.proof_id),
        text(payload.proof_artifact_id),
        text(payload.intent_proof_id),
        text(payload.audit_event_id),
        text(payload.contract_id),
      ];
    }),
  );
  const result =
    payloadText(terminal ?? events[events.length - 1], ["operator_summary", "summary", "message", "result"]) ??
    (completed
      ? "The run completed. Review retained outputs and proof before treating the work as accepted."
      : failed
        ? "The run stopped before it produced a trustworthy completed result."
        : "The run is still active. New events may change the outcome.");
  const failure = failedEvent ? payloadText(failedEvent, ["error", "operator_summary", "message", "reason"]) : undefined;

  return {
    status,
    headline: completed ? "Run completed" : failed ? "Run needs recovery" : "Run in progress",
    result,
    trust: completed
      ? "Completed run evidence is available. Use the event stream only when you need deeper audit detail."
      : failed
        ? "The run record and failure evidence remain trusted. Completed output proof is not reliable for this attempt."
        : "This receipt is provisional until the run reaches a terminal state.",
    next: completed
      ? "Review the output and proof, then return to Soma or the owning workflow."
      : failed
        ? "Review the failure, adjust the request or dependency, then retry from Soma or the owning workflow."
        : "Wait for completion, or inspect events if the run stalls.",
    outputRefs,
    proofRefs,
    failure,
  };
}

function statusClasses(status: ReceiptStatus) {
  if (status === "completed") return "border-cortex-success/30 bg-cortex-success/10 text-cortex-success";
  if (status === "failed") return "border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger";
  return "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary";
}

function StatusIcon({ status }: { status: ReceiptStatus }) {
  if (status === "completed") return <CheckCircle2 className="h-4 w-4" />;
  if (status === "failed") return <AlertTriangle className="h-4 w-4" />;
  return <Activity className="h-4 w-4" />;
}

export default function RunReceipt({ events, runId }: { events: MissionEvent[]; runId: string }) {
  const [inspectOpen, setInspectOpen] = useState(false);
  const receipt = buildRunReceipt(events, runId);

  return (
    <section className="mb-5 rounded-lg border border-cortex-border bg-cortex-surface/60 p-4" aria-label="Run receipt">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2 text-[10px] font-mono font-bold uppercase tracking-widest text-cortex-primary">
            <ShieldCheck className="h-3.5 w-3.5" />
            Run receipt
          </div>
          <h2 className="mt-2 text-lg font-semibold text-cortex-text-main">{receipt.headline}</h2>
        </div>
        <span className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-mono font-bold uppercase ${statusClasses(receipt.status)}`}>
          <StatusIcon status={receipt.status} />
          {receipt.status}
        </span>
      </div>

      <div className="mt-4 grid gap-3 md:grid-cols-3">
        <div className="rounded-md border border-cortex-border bg-cortex-bg/60 p-3">
          <p className="text-[10px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">What happened</p>
          <p className="mt-2 text-sm leading-6 text-cortex-text-main">{receipt.result}</p>
        </div>
        <div className="rounded-md border border-cortex-border bg-cortex-bg/60 p-3">
          <p className="text-[10px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">What to trust</p>
          <p className="mt-2 text-sm leading-6 text-cortex-text-main">{receipt.trust}</p>
        </div>
        <div className="rounded-md border border-cortex-border bg-cortex-bg/60 p-3">
          <p className="text-[10px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">Next step</p>
          <p className="mt-2 text-sm leading-6 text-cortex-text-main">{receipt.next}</p>
        </div>
      </div>

      {receipt.failure ? (
        <div className="mt-3 rounded-md border border-cortex-danger/30 bg-cortex-danger/10 p-3 text-sm text-cortex-danger">
          <span className="font-mono text-[10px] font-bold uppercase tracking-widest">Failure: </span>
          {receipt.failure}
        </div>
      ) : null}

      <div className="mt-3 flex flex-wrap gap-2 text-[10px] font-mono">
        <span className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-bg/70 px-2 py-1 text-cortex-text-muted">
          <FileText className="h-3 w-3" />
          {receipt.outputRefs.length ? `${receipt.outputRefs.length} output ref${receipt.outputRefs.length === 1 ? "" : "s"}` : "No retained output yet"}
        </span>
        <span className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-bg/70 px-2 py-1 text-cortex-text-muted">
          <ShieldCheck className="h-3 w-3" />
          {receipt.proofRefs.length ? `${receipt.proofRefs.length} proof ref${receipt.proofRefs.length === 1 ? "" : "s"}` : "Proof pending"}
        </span>
        {receipt.status === "failed" ? (
          <span className="inline-flex items-center gap-1 rounded border border-cortex-danger/30 bg-cortex-danger/10 px-2 py-1 text-cortex-danger">
            <RotateCcw className="h-3 w-3" />
            Recovery needed
          </span>
        ) : null}
      </div>

      <button
        type="button"
        onClick={() => setInspectOpen((value) => !value)}
        className="mt-3 inline-flex items-center gap-1.5 rounded border border-cortex-border px-2.5 py-1.5 text-[10px] font-mono font-bold text-cortex-text-muted transition-colors hover:border-cortex-primary/40 hover:text-cortex-primary"
      >
        {inspectOpen ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
        Inspect receipt evidence
      </button>

      {inspectOpen ? (
        <div className="mt-3 grid gap-3 rounded-md border border-cortex-border bg-cortex-bg/70 p-3 text-[11px] font-mono text-cortex-text-muted md:grid-cols-3">
          <div>
            <p className="mb-1 font-bold uppercase tracking-widest text-cortex-text-main">Run</p>
            <p className="break-all">{runId}</p>
          </div>
          <div>
            <p className="mb-1 font-bold uppercase tracking-widest text-cortex-text-main">Outputs</p>
            {receipt.outputRefs.length ? receipt.outputRefs.slice(0, 4).map((ref) => <p key={ref}>{compactRef(ref)}</p>) : <p>None recorded yet</p>}
          </div>
          <div>
            <p className="mb-1 font-bold uppercase tracking-widest text-cortex-text-main">Proof</p>
            {receipt.proofRefs.length ? receipt.proofRefs.slice(0, 4).map((ref) => <p key={ref}>{compactRef(ref)}</p>) : <p>Pending terminal proof</p>}
          </div>
        </div>
      ) : null}
    </section>
  );
}
