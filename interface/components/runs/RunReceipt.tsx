"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, FileText, RotateCcw, ShieldCheck } from "lucide-react";
import type { MissionEvent } from "@/store/useCortexStore";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";

type ReceiptStatus = "running" | "completed" | "degraded" | "failed";

type ApprovedWork = {
  executionMode?: string;
  kind?: string;
  objective?: string;
  outputShape?: string;
  primaryDeliverable?: string;
  retention?: string;
  launchHint?: string;
  validation: string[];
  stopAction?: string;
  retryAction?: string;
  recoveryAction?: string;
  controlSummary?: string;
};

type Receipt = {
  status: ReceiptStatus;
  headline: string;
  result: string;
  trust: string;
  next: string;
  outputRefs: string[];
  proofRefs: string[];
  approvedWork?: ApprovedWork;
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

function record(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : undefined;
}

function textList(value: unknown) {
  return Array.isArray(value) ? unique(value.map(text)) : [];
}

function outputRefsFrom(value: unknown) {
  if (!Array.isArray(value)) return [];
  return value.flatMap((item) => {
    if (typeof item === "string") return text(item) ? [item.trim()] : [];
    const ref = record(item);
    if (!ref) return [];
    const canonical = text(ref.storage_ref) ?? text(ref.path) ?? text(ref.file_path) ?? text(ref.url)
      ?? text(ref.output_id) ?? text(ref.artifact_id) ?? text(ref.entrypoint) ?? text(ref.title);
    return canonical ? [canonical] : [];
  });
}

function approvedWorkFrom(events: MissionEvent[]): ApprovedWork | undefined {
  let intent: Record<string, unknown> | undefined;
  let executionMode: string | undefined;
  for (const event of [...events].reverse()) {
    const payload = event.payload ?? {};
    const context = record(payload.context);
    executionMode ??= text(payload.execution_mode) ?? text(context?.execution_mode);
    intent ??= record(payload.work_intent) ?? record(context?.work_intent);
    if (intent && executionMode) break;
  }
  if (!intent && !executionMode) return undefined;
  const output = record(intent?.output_contract);
  const lifecycle = record(intent?.lifecycle);
  return {
    executionMode,
    kind: text(intent?.kind),
    objective: text(intent?.objective),
    outputShape: text(output?.shape),
    primaryDeliverable: text(output?.primary_deliverable),
    retention: text(output?.retention),
    launchHint: text(output?.launch_hint),
    validation: textList(output?.validation),
    stopAction: text(lifecycle?.stop_action),
    retryAction: text(lifecycle?.retry_action),
    recoveryAction: text(lifecycle?.recovery_action),
    controlSummary: text(lifecycle?.control_summary),
  };
}

export function buildRunReceipt(events: MissionEvent[], runId: string): Receipt {
  const terminal = [...events].reverse().find((event) => TERMINAL_EVENTS.has(event.event_type));
  const failedEvent = events.find((event) => event.event_type === "mission.failed" || event.event_type === "tool.failed");
  const completed = terminal?.event_type === "mission.completed";
  const failed = Boolean(failedEvent) || terminal?.event_type === "mission.failed";
  const approvedWork = approvedWorkFrom(events);
  const outputEvents = events.filter((event) => /artifact|output|file|media/i.test(event.event_type));
  const proofEvents = events.filter((event) => /proof|audit|completed|failed/i.test(event.event_type) || event.audit_event_id);
  const outputRefs = unique(
    events.flatMap((event) => {
      const payload = event.payload ?? {};
      const direct = outputEvents.includes(event)
        ? [text(payload.title), text(payload.path), text(payload.storage_ref), text(payload.artifact_id), text(payload.output_id), text(payload.url)]
        : [];
      return [...direct, ...outputRefsFrom(payload.output_refs), ...outputRefsFrom(payload.outputs)];
    }),
  );
  const expectsRetainedOutput = Boolean(
    approvedWork?.primaryDeliverable || approvedWork?.retention?.toLowerCase() === "user_deliverable",
  );
  const missingRequiredOutput = completed && expectsRetainedOutput && outputRefs.length === 0;
  const status: ReceiptStatus = failed ? "failed" : missingRequiredOutput ? "degraded" : completed ? "completed" : "running";
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
  const result = missingRequiredOutput
    ? "The run completed, but the approved deliverable was not retained."
    : payloadText(terminal ?? events[events.length - 1], ["operator_summary", "summary", "message", "result"]) ??
    (completed
      ? "The run completed. Review retained outputs and proof before treating the work as accepted."
      : failed
        ? "The run stopped before it produced a trustworthy completed result."
        : "The run is still active. New events may change the outcome.");
  const failure = failedEvent ? payloadText(failedEvent, ["error", "operator_summary", "message", "reason"]) : undefined;

  return {
    status,
    headline: failed ? "Run needs recovery" : missingRequiredOutput ? "Run needs output recovery" : completed ? "Run completed" : "Run in progress",
    result,
    trust: missingRequiredOutput
      ? "Completion evidence remains trusted. The required output is missing, so the result is not ready to rely on."
      : completed
      ? "Completed run evidence is available. Use the event stream only when you need deeper audit detail."
      : failed
        ? "The run record and failure evidence remain trusted. Completed output proof is not reliable for this attempt."
        : "This receipt is provisional until the run reaches a terminal state.",
    next: missingRequiredOutput
      ? "Recover or rerun the owning work and retain the approved output before acceptance."
      : completed
      ? "Review the output and proof, then return to Soma or the owning workflow."
      : failed
        ? "Review the failure, adjust the request or dependency, then retry from Soma or the owning workflow."
        : "Wait for completion, or inspect events if the run stalls.",
    outputRefs,
    proofRefs,
    approvedWork,
    failure,
  };
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
        <OutcomeHealthBadge health={receipt.status === "failed" ? "blocked" : receipt.status} />
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
        {receipt.status === "failed" || receipt.status === "degraded" ? (
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
        <div className="mt-3 grid gap-3 rounded-md border border-cortex-border bg-cortex-bg/70 p-3 text-[11px] font-mono text-cortex-text-muted md:grid-cols-2">
          <div>
            <p className="mb-1 font-bold uppercase tracking-widest text-cortex-text-main">Run</p>
            <p className="break-all">{runId}</p>
          </div>
          <div>
            <p className="mb-1 font-bold uppercase tracking-widest text-cortex-text-main">Approved work</p>
            {receipt.approvedWork ? (
              <>
                <p>{[receipt.approvedWork.kind, receipt.approvedWork.executionMode].filter(Boolean).join(" · ") || "Typed work"}</p>
                <p>{receipt.approvedWork.primaryDeliverable || receipt.approvedWork.outputShape || "No named deliverable"}</p>
                {receipt.approvedWork.controlSummary ? <p>{receipt.approvedWork.controlSummary}</p> : null}
              </>
            ) : <p>No approved work contract recorded</p>}
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
