"use client";

import React from "react";
import { Shield, AlertTriangle, CheckCircle, Clock3, XCircle } from "lucide-react";
import { useCortexStore, type ChatMessage } from "@/store/useCortexStore";
import { brainBadge, toolLabel, sourceNodeLabel } from "@/lib/labels";

interface Props {
    message: ChatMessage;
}

export default function ProposedActionBlock({ message }: Props) {
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const assistantName = useCortexStore((s) => s.assistantName);

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
        ? "Approval required before execution"
        : approvalMode === "optional"
            ? "Approval optional for this action"
            : "Auto-approved within governance thresholds";
    const actionLabel = approvalRequired ? "Approve & Execute" : "Execute";

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
            {/* Header */}
            <div className="px-4 py-2 bg-amber-400/5 border-b border-amber-400/20 flex items-center gap-2">
                <Shield className="w-4 h-4 text-amber-400" />
                <span className="text-amber-400 font-mono text-xs font-bold tracking-wider">PROPOSED ACTION</span>
            </div>

            {/* Details grid */}
            <div className="px-4 py-3 space-y-2 text-xs font-mono">
                <div className={`flex items-center gap-2 rounded border px-2.5 py-2 text-[10px] ${lifecycleTone}`}>
                    <LifecycleIcon className="h-3.5 w-3.5 flex-shrink-0" />
                    <span>{lifecycleLabel}</span>
                </div>
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
                <div className="flex items-center gap-4">
                    <span className="text-cortex-text-muted w-16">Risk</span>
                    <span className={`px-2 py-0.5 rounded border text-[10px] ${riskColor}`}>
                        {proposal.risk_level?.toUpperCase() || "LOW"}
                    </span>
                </div>
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-16 pt-0.5">Approval</span>
                    <div className="flex-1 space-y-1">
                        <div className="text-cortex-text-main">{governanceSummary}</div>
                        <div className="flex flex-wrap gap-1">
                            <span className={`px-2 py-0.5 rounded border text-[10px] ${approvalRequired ? "text-amber-300 border-amber-400/30" : "text-cortex-success border-cortex-success/30"}`}>
                                {approvalMode.replaceAll("_", " ").toUpperCase()}
                            </span>
                            <span className={`px-2 py-0.5 rounded border text-[10px] ${capabilityRisk === "high" ? "text-red-300 border-red-400/30" : capabilityRisk === "medium" ? "text-amber-300 border-amber-400/30" : "text-cortex-success border-cortex-success/30"}`}>
                                Capability {capabilityRisk.toUpperCase()}
                            </span>
                            {proposal.external_data_use ? (
                                <span className="px-2 py-0.5 rounded border text-[10px] text-cortex-primary border-cortex-primary/30">
                                    External data
                                </span>
                            ) : null}
                            {typeof proposal.estimated_cost === "number" ? (
                                <span className="px-2 py-0.5 rounded border text-[10px] text-cortex-text-main border-cortex-border">
                                    Cost {proposal.estimated_cost.toFixed(2)}
                                </span>
                            ) : null}
                        </div>
                        {proposal.approval_reason ? (
                            <div className="text-cortex-text-muted text-[10px]">
                                Reason: {proposal.approval_reason.replaceAll("_", " ")}
                            </div>
                        ) : null}
                        {capabilityIDs.length > 0 ? (
                            <div className="flex flex-wrap gap-1">
                                {capabilityIDs.map((capability) => (
                                    <span
                                        key={capability}
                                        className="px-1.5 py-0.5 rounded border text-[10px] bg-cortex-bg/60 text-cortex-text-main border-cortex-border"
                                    >
                                        {capability.replaceAll("_", " ")}
                                    </span>
                                ))}
                            </div>
                        ) : null}
                    </div>
                </div>
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
                                                    {binding.adapter_kind ? ` (${binding.adapter_kind})` : ''}
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

            {/* Actions */}
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
