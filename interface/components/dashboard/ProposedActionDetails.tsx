"use client";

import { AlertTriangle } from "lucide-react";
import { brainBadge, toolLabel, sourceNodeLabel } from "@/lib/labels";
import type { ChatMessage, ProposalData } from "@/store/useCortexStore";
import ProposalRunIntent from "./ProposalRunIntent";
import { humanizeLabel, plainExecutionText } from "./proposedActionCopy";

export default function ProposedActionDetails({
    assistantName,
    message,
    proposal,
}: {
    assistantName: string;
    message: ChatMessage;
    proposal: ProposalData;
}) {
    const expressions = proposal.team_expressions ?? [];
    const bindingCount = expressions.reduce((sum, expr) => sum + (expr.module_bindings?.length ?? 0), 0);
    const capabilityIDs = proposal.capability_ids ?? [];

    return (
        <div className="space-y-2 rounded border border-cortex-border bg-cortex-bg/40 px-3 py-3 text-xs font-mono">
            <ProposalRunIntent proposal={proposal} />
            <div className="flex items-center gap-4">
                <span className="text-cortex-text-muted w-16">Handled by</span>
                <span className="text-cortex-text-main">{sourceNodeLabel(message.source_node || "admin", assistantName)}</span>
            </div>
            {message.brain && (
                <div className="flex items-center gap-4">
                    <span className="text-cortex-text-muted w-16">AI</span>
                    <span className="text-cortex-text-main">{brainBadge(message.brain.provider_id, message.brain.location)}</span>
                    {message.brain.location === "remote" && (
                        <span className="text-amber-400 text-[10px] flex items-center gap-1">
                            <AlertTriangle className="w-3 h-3" /> External
                        </span>
                    )}
                </div>
            )}
            {proposal.tools.length > 0 && (
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-16 pt-0.5">Can use</span>
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
                    <span className="text-cortex-text-muted w-16">Why</span>
                    <span className="text-cortex-text-main">{humanizeLabel(proposal.approval_reason)}</span>
                </div>
            ) : null}
            {capabilityIDs.length > 0 ? (
                <div className="flex items-start gap-4">
                    <span className="text-cortex-text-muted w-16 pt-0.5">Allowed</span>
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
                    <span className="text-cortex-text-muted w-16 pt-0.5">Team plan</span>
                    <div className="flex-1 space-y-1.5">
                        <div className="text-cortex-text-muted text-[10px]">
                            {expressions.length} team plan step{expressions.length !== 1 ? "s" : ""}, {bindingCount} tool link{bindingCount !== 1 ? "s" : ""}
                        </div>
                        {expressions.map((expr, idx) => (
                            <div key={expr.expression_id || `expr-${idx}`} className="rounded border border-cortex-border px-2 py-1.5 bg-cortex-bg/40">
                                <div className="text-cortex-text-main text-[10px]">{plainExecutionText(expr.objective)}</div>
                                {expr.module_bindings && expr.module_bindings.length > 0 && (
                                    <div className="flex flex-wrap gap-1 mt-1">
                                        {expr.module_bindings.map((binding, bIdx) => (
                                            <span
                                                key={binding.binding_id || `${binding.module_id}-${bIdx}`}
                                                className="px-1.5 py-0.5 rounded border text-[10px] bg-cortex-primary/10 text-cortex-primary border-cortex-primary/20"
                                                title={binding.operation || binding.module_id}
                                            >
                                                {toolLabel(binding.module_id)}
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
