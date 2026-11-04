"use client";

import React from "react";
import { AlertTriangle } from "lucide-react";
import type { ChatMessage, ProposalData } from "@/store/useCortexStore";
import { brainBadge, sourceNodeLabel, toolLabel } from "@/lib/labels";

function humanizeLabel(value: string): string {
    const normalized = value.replace(/[_-]+/g, " ").trim();
    if (!normalized) return "";
    return normalized.charAt(0).toUpperCase() + normalized.slice(1);
}

export default function ProposedActionDetails({
    message,
    proposal,
    assistantName,
}: {
    message: ChatMessage;
    proposal: ProposalData;
    assistantName: string;
}) {
    const capabilityRisk = proposal.capability_risk ?? proposal.risk_level ?? "low";
    const capabilityIDs = proposal.capability_ids ?? [];
    const expressions = proposal.team_expressions ?? [];
    const bindingCount = expressions.reduce((sum, expr) => sum + (expr.module_bindings?.length ?? 0), 0);

    return (
        <div
            className="space-y-2.5 rounded border border-cortex-border bg-cortex-bg/40 px-3 py-3 text-xs font-mono"
            data-testid="proposal-proof-details"
        >
            <div className="flex flex-wrap gap-1.5 pb-1 border-b border-cortex-border/60">
                <span
                    className={`px-2 py-0.5 rounded border text-[10px] font-mono ${
                        capabilityRisk === "high"
                            ? "text-cortex-danger border-cortex-danger/30"
                            : capabilityRisk === "medium"
                            ? "text-cortex-warning border-cortex-warning/30"
                            : "text-cortex-success border-cortex-success/30"
                    }`}
                >
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

            <div className="flex items-center gap-4">
                <span className="text-cortex-text-muted w-20">Role</span>
                <span className="text-cortex-text-main">{sourceNodeLabel(message.source_node || "admin", assistantName)}</span>
            </div>

            {message.brain && (
                <div className="flex items-center gap-4">
                    <span className="text-cortex-text-muted w-20">Brain</span>
                    <span className="text-cortex-text-main">{brainBadge(message.brain.provider_id, message.brain.location)}</span>
                    {message.brain.location === "remote" && (
                        <span className="text-cortex-warning text-[10px] flex items-center gap-1">
                            <AlertTriangle className="w-3 h-3" /> External
                        </span>
                    )}
                </div>
            )}

            {proposal.tools.length > 0 && (
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-20 pt-0.5">Tools</span>
                    <div className="flex flex-wrap gap-1">
                        {proposal.tools.map((t) => (
                            <span
                                key={t}
                                className="px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary text-[10px] border border-cortex-primary/20"
                            >
                                {toolLabel(t)}
                            </span>
                        ))}
                    </div>
                </div>
            )}

            <div className="flex items-center gap-4">
                <span className="text-cortex-text-muted w-20">Scope</span>
                <span className="text-cortex-text-main">
                    {proposal.teams} team{proposal.teams !== 1 ? "s" : ""}, {proposal.agents} agent{proposal.agents !== 1 ? "s" : ""}
                </span>
            </div>

            {proposal.approval_reason ? (
                <div className="flex items-center gap-4">
                    <span className="text-cortex-text-muted w-20">Reason</span>
                    <span className="text-cortex-text-main">{humanizeLabel(proposal.approval_reason)}</span>
                </div>
            ) : null}

            {capabilityIDs.length > 0 ? (
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-20 pt-0.5">Capabilities</span>
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

            {(proposal.intent_proof_id || proposal.confirm_token || message.run_id) && (
                <div className="flex flex-wrap items-center gap-3 pt-1 text-[10px] text-cortex-text-muted">
                    {proposal.intent_proof_id ? <span>Proof ID: {proposal.intent_proof_id}</span> : null}
                    {proposal.confirm_token ? <span>Token: {proposal.confirm_token}</span> : null}
                    {message.run_id ? <span>Run ID: {message.run_id}</span> : null}
                </div>
            )}

            {expressions.length > 0 && (
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-20 pt-0.5">Expressions</span>
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
    );
}
