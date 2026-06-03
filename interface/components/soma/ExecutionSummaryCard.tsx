"use client";

import type React from "react";
import { CheckCircle2, ChevronDown, ExternalLink, FileText, Gauge, RotateCcw, Route, ShieldCheck, Sparkles } from "lucide-react";
import type { ChatArtifactRef, ExecutionSummaryData } from "@/store/useCortexStore";
import {
  actionableOutputWorkbenchItems,
  OutputWorkbench,
  outputWorkbenchItems,
  projectPackageOutputs,
} from "./OutputWorkbench";
import {
    auditText,
    capabilityGroups,
    compactText,
    degradationLines,
    executionShapeLabel,
    intentLines,
    linkHref,
    linkLabel,
    linkRunId,
    nextStepText,
    proofLinks,
    searchSourceLines,
    trustVerdict,
    type TrustVerdictTone,
    understandingLines,
} from "./ExecutionSummaryCardModel";

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
        <div className="grid grid-cols-[16px_74px_minmax(0,1fr)] gap-2 items-start">
            <div className="mt-0.5 text-cortex-info">{icon}</div>
            <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                {label}
            </div>
            <div className="min-w-0 text-[11px] leading-5 text-cortex-text-main">{children}</div>
        </div>
    );
}

function ChipList({ values }: { values: string[] }) {
    return (
        <div className="flex flex-wrap gap-1">
            {values.map((value) => (
                <span
                    key={value}
                    className="max-w-full truncate rounded border border-cortex-info/20 bg-cortex-info/10 px-1.5 py-0.5 text-[9px] font-mono text-cortex-info"
                    title={value}
                >
                    {value}
                </span>
            ))}
        </div>
    );
}

function trustToneClass(tone: TrustVerdictTone) {
    if (tone === "trusted") return "border-cortex-success/25 bg-cortex-success/10 text-cortex-success";
    if (tone === "attention") return "border-red-400/30 bg-red-400/10 text-red-300";
    return "border-amber-400/25 bg-amber-400/10 text-amber-300";
}

export default function ExecutionSummaryCard({
    summary,
    runId,
    artifacts,
}: {
    summary?: ExecutionSummaryData;
    runId?: string;
    artifacts?: ChatArtifactRef[];
}) {
    if (!summary) return null;

    const executionShape = executionShapeLabel(summary.execution?.shape) ?? executionShapeLabel(summary.execution_shape);
    const executionStatus = compactText(summary.execution?.status) ?? compactText(summary.execution_status);
    const executionSummary = compactText(summary.execution?.summary) ?? compactText(summary.execution_summary);
    const capabilities = capabilityGroups(summary.capability_use);
    const searchSources = searchSourceLines(summary.capability_use);
    const intent = intentLines(summary.intent);
    const understanding = understandingLines(summary.understanding);
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

    const hasContent = intent.length
        || understanding.length
        || executionShape
        || executionStatus
        || executionSummary
        || capabilities.length
        || searchSources.length
        || allOutputs.length
        || projectPackages.length
        || proofs.length
        || summaryRunId
        || audit
        || degradation.length
        || nextStep;

    if (!hasContent) return null;

    const hasReviewDetails = intent.length
        || understanding.length
        || executionShape
        || executionSummary
        || capabilities.length
        || searchSources.length
        || proofs.length
        || audit
        || degradation.length
        || nextStep;

    return (
        <div className="rounded-lg border border-cortex-info/20 bg-cortex-info/5 px-3 py-2.5 shadow-sm" data-testid="execution-summary-card">
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-1.5 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-info">
                    <Sparkles className="h-3 w-3" />
                    Result
                </div>
                <div className="flex flex-wrap items-center gap-1.5">
                    <span className={`rounded border px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase ${trustToneClass(trust.tone)}`}>
                        {trust.label}
                    </span>
                    {summaryRunId && (
                        <a href={`/runs/${summaryRunId}`} className="inline-flex items-center gap-1 rounded border border-cortex-info/20 bg-cortex-info/10 px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase text-cortex-info hover:underline">
                            Run {summaryRunId.slice(0, 8)}
                            <ExternalLink className="h-2.5 w-2.5" />
                        </a>
                    )}
                    {executionStatus && (
                        <span className="rounded border border-cortex-success/20 bg-cortex-success/10 px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase text-cortex-success">
                            {executionStatus}
                        </span>
                    )}
                </div>
            </div>
            <div className="space-y-2">
                {(allOutputs.length > 0 || projectPackages.length > 0) && (
                    <SummaryRow icon={<FileText className="h-3.5 w-3.5" />} label="Outputs">
                        <OutputWorkbench outputs={allOutputs} projectPackages={projectPackages} projectOpenLabel="Open Game" />
                    </SummaryRow>
                )}
                <SummaryRow icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Trust">
                    {trust.detail}
                </SummaryRow>
                {hasReviewDetails ? (
                    <details className="rounded border border-cortex-border/60 bg-cortex-surface/60 px-2 py-1.5">
                        <summary className="flex cursor-pointer list-none items-center justify-between gap-2 text-[10px] font-mono font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
                            <span>Review request, proof, and recovery</span>
                            <ChevronDown className="h-3.5 w-3.5" />
                        </summary>
                        <div className="mt-2 space-y-2">
                            {intent.length > 0 && (
                                <SummaryRow icon={<Route className="h-3.5 w-3.5" />} label="Request">
                                    <div className="space-y-1">
                                        {intent.map((line) => <div key={line}>{line}</div>)}
                                    </div>
                                </SummaryRow>
                            )}
                            {understanding.length > 0 && (
                                <SummaryRow icon={<Sparkles className="h-3.5 w-3.5" />} label="Understood">
                                    <div className="space-y-1">
                                        {understanding.map((line) => <div key={line}>{line}</div>)}
                                    </div>
                                </SummaryRow>
                            )}
                            {(executionShape || executionSummary) && (
                                <SummaryRow icon={<Gauge className="h-3.5 w-3.5" />} label="Progress">
                                    <div className="space-y-1">
                                        {executionShape && <div className="font-mono text-[10px] text-cortex-info">{executionShape}</div>}
                                        {executionSummary && <div>{executionSummary}</div>}
                                    </div>
                                </SummaryRow>
                            )}
                            {capabilities.length > 0 && (
                                <SummaryRow icon={<CheckCircle2 className="h-3.5 w-3.5" />} label="Used">
                                    <div className="space-y-1.5">
                                        {capabilities.map((group) => (
                                            <div key={group.label} className="flex flex-wrap items-center gap-1.5">
                                                <span className="text-[9px] font-mono uppercase text-cortex-text-muted">{group.label}</span>
                                                <ChipList values={group.values} />
                                            </div>
                                        ))}
                                    </div>
                                </SummaryRow>
                            )}
                            {searchSources.length > 0 && (
                                <SummaryRow icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Source">
                                    <div className="space-y-1">
                                        {searchSources.map((line) => <div key={line}>{line}</div>)}
                                    </div>
                                </SummaryRow>
                            )}
                            {proofs.length > 0 && (
                                <SummaryRow icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Evidence">
                                    <div className="flex flex-wrap gap-x-3 gap-y-1">
                                        {proofs.map((proof) => (
                                            proof.url ? (
                                                <a key={`${proof.text}-${proof.url}`} href={proof.url} className="inline-flex items-center gap-1 text-cortex-primary hover:underline">
                                                    {proof.text}
                                                    <ExternalLink className="h-3 w-3" />
                                                </a>
                                            ) : (
                                                <span key={proof.text}>{proof.text}</span>
                                            )
                                        ))}
                                    </div>
                                </SummaryRow>
                            )}
                            {audit && (
                                <SummaryRow icon={<RotateCcw className="h-3.5 w-3.5" />} label="Recovery">
                                    {audit}
                                </SummaryRow>
                            )}
                            {degradation.length > 0 && (
                                <SummaryRow icon={<RotateCcw className="h-3.5 w-3.5" />} label="Degraded">
                                    <div className="space-y-1">
                                        {degradation.map((line) => <div key={line}>{line}</div>)}
                                    </div>
                                </SummaryRow>
                            )}
                            {nextStep && (
                                <div className="rounded border border-cortex-border/60 bg-cortex-bg/70 px-2 py-1.5 text-[11px] leading-5 text-cortex-text-main">
                                    <span className="mr-1 font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted">Next</span>
                                    {nextStep}
                                </div>
                            )}
                        </div>
                    </details>
                ) : null}
            </div>
        </div>
    );
}
