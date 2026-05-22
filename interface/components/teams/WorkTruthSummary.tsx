"use client";

import { AlertTriangle, CheckCircle2, Radio } from "lucide-react";
import type { TeamWorkItem } from "@/store/useCortexStore";

export function WorkTruthSummary({ item }: { item: TeamWorkItem }) {
  const outputRefs = item.outputRefs ?? [];
  const outputCount = item.outputCount ?? outputRefs.length;
  const packageCount = outputRefs.filter(
    (output) => output.kind === "project_package" || Boolean(output.entrypoint),
  ).length;
  const outputProofRefs = outputRefs.flatMap((output) => [
    output.proof_ref,
    output.proof_id,
  ]).filter(Boolean);
  const proofCount =
    new Set([...(item.proofRefs ?? []), ...outputProofRefs]).size +
    (item.runId ? 1 : 0);
  const recoverAction = item.interactions.find(
    (action) => action.action === "recover",
  );
  const canRecover = Boolean(recoverAction && !recoverAction.disabled);
  const isActive = item.state === "running" || item.state === "reviewing";
  const isDegraded =
    item.state === "degraded" || item.state === "needs_operator";
  const summaryTone = isDegraded
    ? "border-amber-400/25 bg-amber-400/10 text-amber-200"
    : isActive
      ? "border-cortex-success/25 bg-cortex-success/10 text-cortex-success"
      : "border-cortex-border bg-cortex-surface text-cortex-text-muted";
  const stateText = isDegraded
    ? canRecover
      ? "Needs recovery"
      : "Degraded, recovery not connected"
    : isActive
      ? "Running, output may still change"
      : item.state === "output_ready"
        ? "Output ready for review"
        : "Work state retained";
  const outputText =
    packageCount > 0
      ? `${packageCount} package${packageCount === 1 ? "" : "s"} retained`
      : outputCount > 0
        ? `${outputCount} output${outputCount === 1 ? "" : "s"} retained`
        : "No retained output yet";
  const proofText = proofCount > 0 ? "Proof available" : "Proof pending";

  return (
    <div
      className="mt-3 flex flex-wrap gap-1.5 text-[11px]"
      aria-label={`Work truth for ${item.title}`}
    >
      <span
        className={`inline-flex max-w-full items-center gap-1 rounded border px-2 py-1 font-mono ${summaryTone}`}
      >
        {isDegraded ? (
          <AlertTriangle className="h-3 w-3 shrink-0" />
        ) : (
          <Radio className="h-3 w-3 shrink-0" />
        )}
        <span className="truncate">{stateText}</span>
      </span>
      <span className="inline-flex max-w-full items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-cortex-text-muted">
        <span className="truncate">{outputText}</span>
      </span>
      <span className="inline-flex max-w-full items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-cortex-text-muted">
        {proofCount > 0 ? (
          <CheckCircle2 className="h-3 w-3 shrink-0 text-cortex-success" />
        ) : null}
        <span className="truncate">{proofText}</span>
      </span>
      {isDegraded && item.recoveryOptions?.[0] ? (
        <span className="inline-flex max-w-full items-center gap-1 rounded border border-amber-400/25 bg-amber-400/10 px-2 py-1 font-mono text-amber-200">
          <span className="truncate">
            Recovery: {item.recoveryOptions[0]}
          </span>
        </span>
      ) : null}
    </div>
  );
}
