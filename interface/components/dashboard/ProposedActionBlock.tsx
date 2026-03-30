"use client";

import React from "react";
import { Shield, AlertTriangle, CheckCircle, ChevronDown, ChevronUp, Clock3, XCircle } from "lucide-react";
import { useCortexStore, type ChatMessage, type ProposalData } from "@/store/useCortexStore";
import { brainBadge, toolLabel, sourceNodeLabel } from "@/lib/labels";

interface Props {
    message: ChatMessage;
}

function humanizeLabel(value: string): string {
    const normalized = value.replace(/[_-]+/g, " ").trim();
    if (!normalized) return "";
    return normalized.charAt(0).toUpperCase() + normalized.slice(1);
}

function fallbackOperatorSummary(proposal: ProposalData): string {
    if (proposal.tools.includes("write_file")) return "create or update files in your workspace.";
    if (proposal.tools.includes("generate_blueprint")) return "prepare a durable blueprint artifact.";
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) {
        return "send a governed signal through the platform.";
    }
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) {
        return "coordinate work through governed team execution.";
    }
    return "carry out a governed action.";
}

function fallbackExpectedResult(proposal: ProposalData): string {
    if (proposal.tools.includes("write_file")) return "The requested file change will be created after approval.";
    if (proposal.tools.includes("generate_blueprint")) return "A saved blueprint artifact will be returned in this conversation.";
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) {
        return "The governed signal will be sent and the outcome will be returned here.";
    }
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) {
        return "Soma will return the execution result in this conversation.";
    }
    return "A governed execution result will be returned in this conversation.";
}

function fallbackAffectedResources(proposal: ProposalData): string[] {
    if (proposal.tools.includes("write_file")) return ["Workspace files"];
    if (proposal.tools.includes("generate_blueprint")) return ["Blueprint artifact"];
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) return ["Governed signal"];
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) return ["Governed team workflow"];
    return [];
}

function explainApprovalPosture(proposal: ProposalData, approvalRequired: boolean, approvalMode: string): string {
    if (approvalRequired) {
        if (proposal.capability_ids?.includes("write_file")) {
            return "This action will change your workspace, so Soma needs your approval before running it.";
        }
        if (proposal.external_data_use) {
            return "This action may use external systems or data, so Soma needs your approval before running it.";
        }
        if (proposal.approval_reason === "cost_threshold") {
            return "This action may incur additional spend, so Soma needs your approval before running it.";
        }
        return "This action crosses a governed policy threshold, so Soma needs your approval before running it.";
    }

    if (approvalMode === "optional") {
        return "This action stays within current policy thresholds, but you can still review it before execution.";
    }

    return "This action is within current policy thresholds and can run without a mandatory approval.";
}

export default function ProposedActionBlock({ message }: Props) {
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const assistantName = useCortexStore((s) => s.assistantName);
    const [detailsOpen, setDetailsOpen] = React.useState(false);

    const proposal = message.proposal;
    if (!proposal) return null;
    const lifecycle = message.proposal_status ?? "active";
    const hasRunProof = Boolean(message.run_id?.trim());
    const renderedLifecycle = lifecycle === "executed" && !hasRunProof ? "confirmed_pending_execution" : lifecycle;
    const expressions = proposal.team_expressions ?? [];
    const bindingCount = expressions.reduce((sum, expr) => sum + (expr.module_bindings?.length ?? 0), 0);
    const isActionable = renderedLifecycle === "active";
    const approvalRequired = proposal.approval_required ?? true;
    const approvalMode = proposal.approval_mode ?? (approvalRequired ? "required" : "auto_allowed");
    const capabilityRisk = proposal.capability_risk ?? proposal.risk_level ?? "low";
    const capabilityIDs = proposal.capability_ids ?? [];
    const governanceSummary = approvalRequired
        ? "Approval required"
        : approvalMode === "optional"
            ? "Approval optional"
            : "Auto-approved";
    const actionLabel = approvalRequired ? "Approve & Execute" : "Execute";
    const operatorSummary = proposal.operator_summary?.trim() || fallbackOperatorSummary(proposal);
    const expectedResult = proposal.expected_result?.trim() || fallbackExpectedResult(proposal);
    const affectedResources = (proposal.affected_resources ?? []).filter((value) => value.trim().length > 0);
    const visibleResources = affectedResources.length > 0 ? affectedResources : fallbackAffectedResources(proposal);
    const approvalExplanation = explainApprovalPosture(proposal, approvalRequired, approvalMode);

    const riskColor = proposal.risk_level === "high"
        ? "text-red-400 border-red-400/30"
        : proposal.risk_level === "medium"
            ? "text-amber-400 border-amber-400/30"
            : "text-cortex-success border-cortex-success/30";

    const lifecycleTone = renderedLifecycle === "cancelled"
        ? "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted"
        : renderedLifecycle === "confirmed_pending_execution"
            ? "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary"
            : renderedLifecycle === "executed"
                ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
                : renderedLifecycle === "failed"
                    ? "border-red-400/30 bg-red-400/10 text-red-300"
                    : "border-amber-400/20 bg-amber-400/5 text-amber-300";

    const lifecycleLabel = renderedLifecycle === "cancelled"
        ? "Cancelled"
        : renderedLifecycle === "confirmed_pending_execution"
            ? "Confirmed, awaiting execution proof"
            : renderedLifecycle === "executed"
                ? "Execution verified"
                : renderedLifecycle === "failed"
                    ? "Confirmation failed"
                    : "Awaiting approval";

    const LifecycleIcon = renderedLifecycle === "cancelled"
        ? XCircle
        : renderedLifecycle === "confirmed_pending_execution"
            ? Clock3
            : renderedLifecycle === "executed"
                ? CheckCircle
                : renderedLifecycle === "failed"
                    ? AlertTriangle
                    : Shield;

    return (
        <div className="mt-3 rounded-lg border border-amber-400/30 bg-cortex-surface/80 overflow-hidden">
            <div className="px-4 py-2 bg-amber-400/5 border-b border-amber-400/20 flex items-center gap-2">
                <Shield className="w-4 h-4 text-amber-400" />
                <span className="text-amber-400 font-mono text-xs font-bold tracking-wider">PROPOSED ACTION</span>
            </div>

            <div className="px-4 py-3 space-y-3">
                <div className={`flex items-center gap-2 rounded border px-2.5 py-2 text-[10px] ${lifecycleTone}`}>
                    <LifecycleIcon className="h-3.5 w-3.5 flex-shrink-0" />
                    <span>{lifecycleLabel}</span>
                </div>

                <div className="space-y-1">
                    <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Soma wants to</div>
                    <p className="text-sm leading-6 text-cortex-text-main">{operatorSummary}</p>
                </div>

                <div className="grid gap-3 md:grid-cols-2">
                    <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                            {approvalRequired ? "Why approval is needed" : "Execution posture"}
                        </div>
                        <p className="mt-1.5 text-sm leading-6 text-cortex-text-main">{approvalExplanation}</p>
                    </div>
                    <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Expected result</div>
                        <p className="mt-1.5 text-sm leading-6 text-cortex-text-main">{expectedResult}</p>
                    </div>
                </div>

                {visibleResources.length > 0 ? (
                    <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">What will change</div>
                        <div className="mt-2 flex flex-wrap gap-1.5">
                            {visibleResources.map((resource) => (
                                <span
                                    key={resource}
                                    className="rounded border border-cortex-border bg-cortex-bg/70 px-2 py-1 text-xs text-cortex-text-main"
                                >
                                    {resource}
                                </span>
                            ))}
                        </div>
                    </div>
                ) : null}

                <div className="flex flex-wrap gap-1.5">
                    <span className={`px-2 py-0.5 rounded border text-[10px] font-mono ${approvalRequired ? "text-amber-300 border-amber-400/30" : "text-cortex-success border-cortex-success/30"}`}>
                        {governanceSummary.toUpperCase()}
                    </span>
                    <span className={`px-2 py-0.5 rounded border text-[10px] font-mono ${capabilityRisk === "high" ? "text-red-300 border-red-400/30" : capabilityRisk === "medium" ? "text-amber-300 border-amber-400/30" : "text-cortex-success border-cortex-success/30"}`}>
                        RISK {proposal.risk_level?.toUpperCase() || "LOW"}
                    </span>
                    {proposal.external_data_use ? (
                        <span className="px-2 py-0.5 rounded border text-[10px] font-mono text-cortex-primary border-cortex-primary/30">
                            EXTERNAL DATA
                        </span>
                    ) : null}
                    {typeof proposal.estimated_cost === "number" ? (
                        <span className="px-2 py-0.5 rounded border text-[10px] font-mono text-cortex-text-main border-cortex-border">
                            EST. COST {proposal.estimated_cost.toFixed(2)}
                        </span>
                    ) : null}
                </div>

                <button
                    type="button"
                    onClick={() => setDetailsOpen((open) => !open)}
                    className="inline-flex items-center gap-1.5 text-xs font-mono text-cortex-primary hover:text-cortex-primary/80 transition-colors"
                    aria-expanded={detailsOpen}
                >
                    {detailsOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                    {detailsOpen ? "Hide details" : "Show details"}
                </button>

                {detailsOpen ? (
                    <div className="space-y-2 rounded border border-cortex-border bg-cortex-bg/40 px-3 py-3 text-xs font-mono">
                        <div className="flex items-center gap-4">
                            <span className="text-cortex-text-muted w-16">Role</span>
                            <span className="text-cortex-text-main">{sourceNodeLabel(message.source_node || "admin", assistantName)}</span>
                        </div>
                        {message.brain && (
                            <div className="flex items-center gap-4">
                                <span className="text-cortex-text-muted w-16">Brain</span>
                                <span className="text-cortex-text-main">
                                    {brainBadge(message.brain.provider_id, message.brain.location)}
                                </span>
                                {message.brain.location === "remote" && (
                                    <span className="text-amber-400 text-[10px] flex items-center gap-1">
                                        <AlertTriangle className="w-3 h-3" /> External
                                    </span>
                                )}
                            </div>
                        )}
                        {proposal.tools.length > 0 && (
                            <div className="flex items-start gap-4">
                                <span className="text-cortex-text-muted w-16 pt-0.5">Tools</span>
                                <div className="flex flex-wrap gap-1">
                                    {proposal.tools.map((t) => (
                                        <span key={t} className="px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary text-[10px] border border-cortex-primary/20">
                                            {toolLabel(t)}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}
                        <div className="flex items-center gap-4">
                            <span className="text-cortex-text-muted w-16">Scope</span>
                            <span className="text-cortex-text-main">
                                {proposal.teams} team{proposal.teams !== 1 ? "s" : ""}, {proposal.agents} agent{proposal.agents !== 1 ? "s" : ""}
                            </span>
                        </div>
                        {proposal.approval_reason ? (
                            <div className="flex items-center gap-4">
                                <span className="text-cortex-text-muted w-16">Reason</span>
                                <span className="text-cortex-text-main">{humanizeLabel(proposal.approval_reason)}</span>
                            </div>
                        ) : null}
                        {capabilityIDs.length > 0 ? (
                            <div className="flex items-start gap-4">
                                <span className="text-cortex-text-muted w-16 pt-0.5">Capabilities</span>
                                <div className="flex flex-wrap gap-1">
                                    {capabilityIDs.map((capability) => (
                                        <span
                                            key={capability}
                                            className="px-1.5 py-0.5 rounded border text-[10px] bg-cortex-bg/60 text-cortex-text-main border-cortex-border"
                                        >
                                            {humanizeLabel(capability)}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        ) : null}
                        {expressions.length > 0 && (
                            <div className="flex items-start gap-4">
                                <span className="text-cortex-text-muted w-16 pt-0.5">Expressions</span>
                                <div className="flex-1 space-y-1.5">
                                    <div className="text-cortex-text-muted text-[10px]">
                                        {expressions.length} expression{expressions.length !== 1 ? "s" : ""}, {bindingCount} module binding{bindingCount !== 1 ? "s" : ""}
                                    </div>
                                    {expressions.map((expr, idx) => (
                                        <div key={expr.expression_id || `expr-${idx}`} className="rounded border border-cortex-border px-2 py-1.5 bg-cortex-bg/40">
                                            <div className="text-cortex-text-main text-[10px]">{expr.objective}</div>
                                            {expr.module_bindings && expr.module_bindings.length > 0 && (
                                                <div className="flex flex-wrap gap-1 mt-1">
                                                    {expr.module_bindings.map((binding, bIdx) => (
                                                        <span
                                                            key={binding.binding_id || `${binding.module_id}-${bIdx}`}
                                                            className="px-1.5 py-0.5 rounded border text-[10px] bg-cortex-primary/10 text-cortex-primary border-cortex-primary/20"
                                                            title={binding.operation || binding.module_id}
                                                        >
                                                            {binding.module_id}
                                                            {binding.adapter_kind ? ` (${binding.adapter_kind})` : ""}
                                                        </span>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                ) : null}
            </div>

            {isActionable ? (
                <div className="px-4 py-3 border-t border-cortex-border flex items-center gap-2">
                    <button
                        onClick={() => confirmProposal()}
                        className="px-3 py-1.5 rounded bg-cortex-success/20 border border-cortex-success/40 text-cortex-success text-xs font-mono hover:bg-cortex-success/30 transition-colors flex items-center gap-1.5"
                    >
                        <CheckCircle className="w-3 h-3" />
                        {actionLabel}
                    </button>
                    <button
                        onClick={() => cancelProposal()}
                        className="px-3 py-1.5 rounded text-red-400 text-xs font-mono hover:bg-red-400/10 transition-colors flex items-center gap-1.5"
                    >
                        <XCircle className="w-3 h-3" />
                        Cancel
                    </button>
                </div>
            ) : null}
        </div>
    );
}
