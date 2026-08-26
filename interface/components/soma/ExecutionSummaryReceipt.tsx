"use client";

import { CheckCircle2, ExternalLink, ShieldAlert } from "lucide-react";
import type { ChatArtifactRef, ExecutionSummaryData } from "@/store/useCortexStore";
import {
  actionableOutputWorkbenchItems,
  outputWorkbenchItems,
  projectPackageOutputs,
} from "./OutputWorkbench";
import { outputWorkbenchDigest, OutputWorkbenchCompactDigest } from "./OutputWorkbenchDigest";
import { asItems, compactText, proofLinks, linkRunId, trustVerdict } from "./ExecutionSummaryCardModel";
import ExecutionSummaryMediaPreview from "./ExecutionSummaryMediaPreview";

export function shouldUseExecutionSummaryReceipt({
  summary,
  runId,
  artifacts,
}: {
  summary: ExecutionSummaryData;
  runId?: string;
  artifacts?: ChatArtifactRef[];
}) {
  return trustVerdict(summary, runId, artifacts).tone !== "attention";
}

export default function ExecutionSummaryReceipt({
  summary,
  runId,
  artifacts,
}: {
  summary: ExecutionSummaryData;
  runId?: string;
  artifacts?: ChatArtifactRef[];
}) {
  const trust = trustVerdict(summary, runId, artifacts);
  const outputs = actionableOutputWorkbenchItems(outputWorkbenchItems(summary, artifacts));
  const packages = projectPackageOutputs(summary.outputs);
  const digest = outputWorkbenchDigest({ outputs, projectPackages: packages });
  const summaryRunId = runId ?? proofLinks(summary.proof).map(linkRunId).find(Boolean) ?? null;
  const configStateSummary = asItems(summary.outputs).some((output) => (
    typeof output !== "string" && output.kind?.startsWith("config_")
  )) ? compactText(summary.execution?.summary) : null;

  return (
    <div
      className="rounded-lg border border-cortex-primary/25 bg-cortex-primary/5 px-3 py-2"
      data-testid="execution-summary-receipt"
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            {trust.tone === "trusted" ? (
              <CheckCircle2 className="h-3.5 w-3.5 text-cortex-success" />
            ) : (
              <ShieldAlert className="h-3.5 w-3.5 text-amber-300" />
            )}
            <p className="font-mono text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-primary">
              {trust.label}
            </p>
          </div>
          <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
            {configStateSummary
              ? configStateSummary
              : digest?.isProjectPackage
              ? `${digest.text} is ready. Open the app now; follow-up and proof stay available in details.`
              : digest
                ? `${digest.text} is ready. Open it now; follow-up and proof stay available in details.`
                : trust.detail}
          </p>
        </div>
      </div>
      {digest ? (
        <div className="mt-2 rounded-lg border border-cortex-border/70 bg-cortex-bg/80 px-2.5 py-2">
          <OutputWorkbenchCompactDigest digest={digest} />
        </div>
      ) : null}
      {outputs.length > 0 ? <ExecutionSummaryMediaPreview outputs={outputs} compact /> : null}
      {summaryRunId ? (
        <details className="mt-2 border-t border-cortex-border/60 pt-2 text-[10px] text-cortex-text-muted">
          <summary className="cursor-pointer font-semibold uppercase tracking-[0.08em] hover:text-cortex-text-main">
            Proof and execution details
          </summary>
          <a
            href={`/runs/${summaryRunId}`}
            className="mt-2 inline-flex items-center gap-1 text-cortex-primary hover:underline"
          >
            Inspect run receipt
            <ExternalLink className="h-3 w-3" />
          </a>
        </details>
      ) : null}
    </div>
  );
}
