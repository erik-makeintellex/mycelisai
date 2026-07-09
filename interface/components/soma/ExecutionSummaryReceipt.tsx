"use client";

import { CheckCircle2, ExternalLink, ShieldAlert } from "lucide-react";
import type { ChatArtifactRef, ExecutionSummaryData } from "@/store/useCortexStore";
import {
  actionableOutputWorkbenchItems,
  outputWorkbenchItems,
  projectPackageOutputs,
} from "./OutputWorkbench";
import { outputWorkbenchDigest, OutputWorkbenchCompactDigest } from "./OutputWorkbenchDigest";
import { proofLinks, linkRunId, trustVerdict } from "./ExecutionSummaryCardModel";
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
            {digest?.isProjectPackage
              ? "App/package output is ready. Open it, browse it in Resources, or reply to Soma for changes."
              : digest
                ? "Latest output is ready. Use Open file or open the review panel for proof."
                : trust.detail}
          </p>
        </div>
        {summaryRunId ? (
          <a
            href={`/runs/${summaryRunId}`}
            className="inline-flex shrink-0 items-center gap-1 rounded-lg border border-cortex-border px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.08em] text-cortex-text-main hover:border-cortex-primary/40"
          >
            Run
            <ExternalLink className="h-3 w-3" />
          </a>
        ) : null}
      </div>
      {digest ? (
        <div className="mt-2 rounded-lg border border-cortex-border/70 bg-cortex-bg/80 px-2.5 py-2">
          <OutputWorkbenchCompactDigest digest={digest} />
        </div>
      ) : null}
      {outputs.length > 0 ? <ExecutionSummaryMediaPreview outputs={outputs} compact /> : null}
    </div>
  );
}
