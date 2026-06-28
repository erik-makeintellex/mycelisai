"use client";

import { ChevronDown, ExternalLink, RotateCcw, ShieldAlert, ShieldCheck, Sparkles } from "lucide-react";
import type React from "react";
import type { ChatArtifactRef, ExecutionSummaryData } from "@/store/useCortexStore";
import {
  actionableOutputWorkbenchItems,
  outputWorkbenchItems,
  projectPackageOutputs,
} from "./OutputWorkbench";
import {
  auditText,
  compactText,
  degradationLines,
  executionShapeLabel,
  executionSummaryHeading,
  linkHref,
  linkLabel,
  linkRunId,
  nextStepText,
  proofLinks,
  trustVerdict,
  type TrustVerdictTone,
} from "./ExecutionSummaryCardModel";
import ExecutionSummaryMediaPreview from "./ExecutionSummaryMediaPreview";
import { outputWorkbenchDigest, OutputWorkbenchCompactDigest } from "./OutputWorkbenchDigest";

function trustToneClass(tone: TrustVerdictTone) {
  if (tone === "trusted") return "border-cortex-success/25 bg-cortex-success/10 text-cortex-success";
  if (tone === "attention") return "border-red-400/30 bg-red-400/10 text-red-300";
  return "border-amber-400/25 bg-amber-400/10 text-amber-300";
}

function cardToneClass(tone: TrustVerdictTone) {
  if (tone === "attention") return "border-red-400/25 bg-red-400/[0.04]";
  if (tone === "trusted") return "border-cortex-success/20 bg-cortex-success/[0.04]";
  return "border-cortex-info/20 bg-cortex-info/5";
}

function statusToneClass(status: string | null, trustTone: TrustVerdictTone) {
  const normalized = status?.toLowerCase();
  if (trustTone === "attention" || normalized === "failed" || normalized === "blocked" || normalized === "cancelled") {
    return "border-red-400/25 bg-red-400/10 text-red-300";
  }
  if (trustTone === "trusted" || normalized === "complete" || normalized === "completed" || normalized === "verified") {
    return "border-cortex-success/20 bg-cortex-success/10 text-cortex-success";
  }
  return "border-amber-400/25 bg-amber-400/10 text-amber-300";
}

function SummaryRow({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-[16px_74px_minmax(0,1fr)] items-start gap-2">
      <div className="mt-0.5 text-cortex-info">{icon}</div>
      <div className="font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted">
        {label}
      </div>
      <div className="min-w-0 text-[11px] leading-5 text-cortex-text-main">{children}</div>
    </div>
  );
}

export default function ExecutionSummaryCompactCard({
  summary,
  runId,
  artifacts,
}: {
  summary: ExecutionSummaryData;
  runId?: string;
  artifacts?: ChatArtifactRef[];
}) {
  const executionStatus = compactText(summary.execution?.status) ?? compactText(summary.execution_status);
  const executionSummary = compactText(summary.execution?.summary) ?? compactText(summary.execution_summary);
  const executionShape = executionShapeLabel(summary.execution?.shape) ?? executionShapeLabel(summary.execution_shape);
  const projectPackages = projectPackageOutputs(summary.outputs);
  const allOutputs = actionableOutputWorkbenchItems(outputWorkbenchItems(summary, artifacts));
  const proofs = proofLinks(summary.proof)
    .map((proof) => ({ text: linkLabel(proof), url: linkHref(proof) }))
    .filter((proof): proof is { text: string; url: string | null } => Boolean(proof.text));
  const summaryRunId = runId ?? proofLinks(summary.proof).map(linkRunId).find(Boolean) ?? null;
  const audit = auditText(summary.audit_recovery);
  const degradation = degradationLines(summary.audit_recovery);
  const nextStep = nextStepText(summary.next_step);
  const trust = trustVerdict(summary, summaryRunId ?? runId, artifacts);
  const outputDigest = outputWorkbenchDigest({ outputs: allOutputs, projectPackages });
  const hasDetails = executionShape || proofs.length || audit || degradation.length || nextStep;
  const heading = executionSummaryHeading(summary, allOutputs.length + projectPackages.length);

  return (
    <div className={`rounded-lg border px-2.5 py-2 shadow-sm ${cardToneClass(trust.tone)}`} data-testid="execution-summary-card">
      <div className="mb-1.5 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-info">
          {trust.tone === "attention" ? (
            <ShieldAlert className="h-3 w-3 text-red-300" />
          ) : (
            <Sparkles className="h-3 w-3" />
          )}
          {heading}
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <span className={`rounded border px-1.5 py-0.5 font-mono text-[9px] font-bold uppercase ${trustToneClass(trust.tone)}`}>
            {trust.label}
          </span>
          {summaryRunId ? (
            <a href={`/runs/${summaryRunId}`} className="inline-flex items-center gap-1 rounded border border-cortex-info/20 bg-cortex-info/10 px-1.5 py-0.5 font-mono text-[9px] font-bold uppercase text-cortex-info hover:underline">
              Run {summaryRunId.slice(0, 8)}
              <ExternalLink className="h-2.5 w-2.5" />
            </a>
          ) : null}
          {executionStatus ? (
            <span className={`rounded border px-1.5 py-0.5 font-mono text-[9px] font-bold uppercase ${statusToneClass(executionStatus, trust.tone)}`}>
              {executionStatus}
            </span>
          ) : null}
        </div>
      </div>
      <div className="space-y-1.5">
        {outputDigest ? <OutputWorkbenchCompactDigest digest={outputDigest} /> : null}
        {!outputDigest && executionSummary ? <p className="line-clamp-2 text-xs leading-5 text-cortex-text-main">{executionSummary}</p> : null}
        {allOutputs.length > 0 ? <ExecutionSummaryMediaPreview outputs={allOutputs} compact /> : null}
        <div className="rounded border border-cortex-border/60 bg-cortex-surface/60 px-2 py-1.5 text-[11px] leading-4 text-cortex-text-muted">
          <span className="mr-1 font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted">Trust</span>
          {trust.detail}
        </div>
        {hasDetails ? (
          <details className="rounded border border-cortex-border/60 bg-cortex-surface/60 px-2 py-1.5">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-2 font-mono text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
              <span>Details and proof</span>
              <ChevronDown className="h-3.5 w-3.5" />
            </summary>
            <div className="mt-2 space-y-2">
              {proofs.length > 0 ? (
                <SummaryRow icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Evidence">
                  <div className="flex flex-wrap gap-x-3 gap-y-1">
                    {proofs.map((proof) => (
                      proof.url ? (
                        <a key={`${proof.text}-${proof.url}`} href={proof.url} className="inline-flex items-center gap-1 text-cortex-primary hover:underline">
                          {proof.text}
                          <ExternalLink className="h-3 w-3" />
                        </a>
                      ) : <span key={proof.text}>{proof.text}</span>
                    ))}
                  </div>
                </SummaryRow>
              ) : null}
              {audit ? <SummaryRow icon={<RotateCcw className="h-3.5 w-3.5" />} label="Recovery">{audit}</SummaryRow> : null}
              {degradation.length > 0 ? (
                <SummaryRow icon={<RotateCcw className="h-3.5 w-3.5" />} label="Blocked">
                  <div className="space-y-1">
                    {degradation.map((line) => <div key={line}>{line}</div>)}
                  </div>
                </SummaryRow>
              ) : null}
              {nextStep ? (
                <div className="rounded border border-cortex-border/60 bg-cortex-bg/70 px-2 py-1.5 text-[11px] leading-5 text-cortex-text-main">
                  <span className="mr-1 font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted">Next</span>
                  {nextStep}
                </div>
              ) : null}
            </div>
          </details>
        ) : null}
      </div>
    </div>
  );
}
