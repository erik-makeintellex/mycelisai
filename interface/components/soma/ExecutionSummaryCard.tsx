"use client";

import type React from "react";
import { CheckCircle2, ExternalLink, FileText, Gauge, RotateCcw, Route, ShieldCheck, Sparkles } from "lucide-react";
import type {
    ExecutionSummaryCapabilityUse,
    ExecutionSummaryData,
    ExecutionSummaryItem,
    ExecutionSummaryLink,
} from "@/store/useCortexStore";

type SummaryValue = string | ExecutionSummaryItem;

function compactText(value?: string | null) {
    const trimmed = value?.trim();
    return trimmed || null;
}

function itemText(item: SummaryValue): string | null {
    if (typeof item === "string") return compactText(item);
    return compactText(item.label)
        ?? compactText(item.title)
        ?? compactText(item.name)
        ?? compactText(item.summary)
        ?? compactText(item.value)
        ?? compactText(item.id)
        ?? null;
}

function itemUrl(item: SummaryValue): string | null {
    if (typeof item === "string") return null;
    return compactText(item.url) ?? compactText(item.href) ?? compactText(item.path) ?? null;
}

function intentLines(intent: ExecutionSummaryData["intent"]): string[] {
    if (!intent) return [];
    if (typeof intent === "string") return compactText(intent) ? [intent] : [];
    return [
        compactText(intent.original),
        compactText(intent.resolved) ? `Resolved: ${intent.resolved}` : null,
    ].filter(Boolean) as string[];
}

function understandingLines(understanding: ExecutionSummaryData["understanding"]): string[] {
    if (!understanding) return [];
    if (typeof understanding === "string") return compactText(understanding) ? [understanding] : [];
    return [
        compactText(understanding.summary),
        ...(understanding.assumptions ?? []).map((item) => `Assumption: ${item}`),
    ].filter(Boolean) as string[];
}

function asItems(value: ExecutionSummaryData["outputs"]): SummaryValue[] {
    if (!value) return [];
    if (typeof value === "string") return [value];
    return value;
}

function linkLabel(link: string | ExecutionSummaryLink): string | null {
    if (typeof link === "string") return compactText(link);
    return compactText(link.label)
        ?? compactText(link.title)
        ?? (compactText(link.run_id) ? `Run ${link.run_id}` : null)
        ?? (compactText(link.audit_event_id) ? `Audit ${link.audit_event_id}` : null)
        ?? (compactText(link.intent_proof_id) ? `Proof ${link.intent_proof_id}` : null)
        ?? compactText(link.id)
        ?? compactText(link.path)
        ?? compactText(link.url)
        ?? compactText(link.href)
        ?? null;
}

function linkHref(link: string | ExecutionSummaryLink): string | null {
    if (typeof link === "string") return link.startsWith("/") || link.startsWith("http") ? link : null;
    return compactText(link.url)
        ?? compactText(link.href)
        ?? compactText(link.path)
        ?? (compactText(link.run_id) ? `/runs/${link.run_id}` : null)
        ?? null;
}

function proofLinks(proof: ExecutionSummaryData["proof"]): Array<string | ExecutionSummaryLink> {
    if (!proof) return [];
    return Array.isArray(proof) ? proof : [proof];
}

function capabilityGroups(capabilityUse: ExecutionSummaryData["capability_use"]) {
    if (!capabilityUse) return [];
    if (Array.isArray(capabilityUse)) {
        const values = capabilityUse.map(itemText).filter(Boolean) as string[];
        return values.length ? [{ label: "Used", values }] : [];
    }

    const groups: Array<{ label: string; values: string[] }> = [];
    const source = capabilityUse as ExecutionSummaryCapabilityUse;
    const candidates: Array<[keyof ExecutionSummaryCapabilityUse, string]> = [
        ["capabilities", "Capabilities"],
        ["teams", "Teams"],
        ["agents", "Agents"],
        ["tools", "Tools"],
        ["used", "Used"],
    ];

    for (const [key, label] of candidates) {
        const values = source[key]?.map(itemText).filter(Boolean) as string[] | undefined;
        if (values?.length) groups.push({ label, values });
    }
    return groups;
}

function auditText(value: ExecutionSummaryData["audit_recovery"]) {
    if (!value) return null;
    if (typeof value === "string") return compactText(value);
    const status = compactText(value.status) ?? compactText(value.approval_status);
    const recovery = compactText(value.recovery_state);
    const summary = compactText(value.summary) ?? compactText(value.value) ?? compactText(value.label);
    const blocker = compactText(value.blocker);
    return [status, recovery, summary, blocker].filter(Boolean).join(": ") || null;
}

function nextStepText(value: ExecutionSummaryData["next_step"]) {
    if (!value) return null;
    if (typeof value === "string") return compactText(value);
    return compactText(value.label)
        ?? compactText(value.title)
        ?? compactText(value.action)
        ?? compactText(value.href)
        ?? compactText(value.url)
        ?? null;
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

export default function ExecutionSummaryCard({ summary }: { summary?: ExecutionSummaryData }) {
    if (!summary) return null;

    const executionShape = compactText(summary.execution?.shape) ?? compactText(summary.execution_shape);
    const executionStatus = compactText(summary.execution?.status) ?? compactText(summary.execution_status);
    const executionSummary = compactText(summary.execution?.summary) ?? compactText(summary.execution_summary);
    const capabilities = capabilityGroups(summary.capability_use);
    const intent = intentLines(summary.intent);
    const understanding = understandingLines(summary.understanding);
    const outputs = asItems(summary.outputs)
        .map((item) => ({ text: itemText(item), url: itemUrl(item) }))
        .filter((item): item is { text: string; url: string | null } => Boolean(item.text));
    const proofs = proofLinks(summary.proof)
        .map((proof) => ({ text: linkLabel(proof), url: linkHref(proof) }))
        .filter((proof): proof is { text: string; url: string | null } => Boolean(proof.text));
    const audit = auditText(summary.audit_recovery);
    const nextStep = nextStepText(summary.next_step);

    const hasContent = intent.length
        || understanding.length
        || executionShape
        || executionStatus
        || executionSummary
        || capabilities.length
        || outputs.length
        || proofs.length
        || audit
        || nextStep;

    if (!hasContent) return null;

    return (
        <div className="rounded-lg border border-cortex-info/20 bg-cortex-info/5 px-3 py-2.5 shadow-sm" data-testid="execution-summary-card">
            <div className="mb-2 flex items-center justify-between gap-2">
                <div className="flex items-center gap-1.5 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-info">
                    <Sparkles className="h-3 w-3" />
                    Directed execution
                </div>
                {executionStatus && (
                    <span className="rounded border border-cortex-success/20 bg-cortex-success/10 px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase text-cortex-success">
                        {executionStatus}
                    </span>
                )}
            </div>
            <div className="space-y-2">
                {(intent.length > 0 || understanding.length > 0) && (
                    <SummaryRow icon={<Route className="h-3.5 w-3.5" />} label="Intent">
                        <div className="space-y-1">
                            {intent.map((line) => <div key={line}>{line}</div>)}
                            {understanding.map((line) => <div key={line} className="text-cortex-text-muted">{line}</div>)}
                        </div>
                    </SummaryRow>
                )}
                {(executionShape || executionSummary) && (
                    <SummaryRow icon={<Gauge className="h-3.5 w-3.5" />} label="Shape">
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
                {outputs.length > 0 && (
                    <SummaryRow icon={<FileText className="h-3.5 w-3.5" />} label="Outputs">
                        <div className="flex flex-wrap gap-x-3 gap-y-1">
                            {outputs.map((output) => (
                                output.url ? (
                                    <a key={`${output.text}-${output.url}`} href={output.url} className="inline-flex items-center gap-1 text-cortex-primary hover:underline">
                                        {output.text}
                                        <ExternalLink className="h-3 w-3" />
                                    </a>
                                ) : (
                                    <span key={output.text}>{output.text}</span>
                                )
                            ))}
                        </div>
                    </SummaryRow>
                )}
                {proofs.length > 0 && (
                    <SummaryRow icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Proof">
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
                    <SummaryRow icon={<RotateCcw className="h-3.5 w-3.5" />} label="Audit">
                        {audit}
                    </SummaryRow>
                )}
                {nextStep && (
                    <div className="rounded border border-cortex-border/60 bg-cortex-surface/60 px-2 py-1.5 text-[11px] leading-5 text-cortex-text-main">
                        <span className="mr-1 font-mono text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted">Next</span>
                        {nextStep}
                    </div>
                )}
            </div>
        </div>
    );
}
