"use client";

import React from "react";
import { Shield, AlertTriangle, CheckCircle, XCircle, Edit3 } from "lucide-react";
import { useCortexStore, type ChatMessage } from "@/store/useCortexStore";
import { brainBadge, toolLabel, sourceNodeLabel } from "@/lib/labels";

interface Props {
    message: ChatMessage;
}

export default function ProposedActionBlock({ message }: Props) {
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);

    const proposal = message.proposal;
    if (!proposal) return null;

    const riskColor = proposal.risk_level === "high"
        ? "text-red-400 border-red-400/30"
        : proposal.risk_level === "medium"
            ? "text-amber-400 border-amber-400/30"
            : "text-cortex-success border-cortex-success/30";

    return (
        <div className="mt-3 rounded-lg border border-amber-400/30 bg-cortex-surface/80 overflow-hidden">
            {/* Header */}
            <div className="px-4 py-2 bg-amber-400/5 border-b border-amber-400/20 flex items-center gap-2">
                <Shield className="w-4 h-4 text-amber-400" />
                <span className="text-amber-400 font-mono text-xs font-bold tracking-wider">PROPOSED ACTION</span>
            </div>

            {/* Details grid */}
            <div className="px-4 py-3 space-y-2 text-xs font-mono">
                <div className="flex items-center gap-4">
                    <span className="text-cortex-text-muted w-16">Role</span>
                    <span className="text-cortex-text-main">{sourceNodeLabel(message.source_node || "admin")}</span>
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
            </div>

            {/* Actions */}
            <div className="px-4 py-3 border-t border-cortex-border flex items-center gap-2">
                <button
                    onClick={() => confirmProposal()}
                    className="px-3 py-1.5 rounded bg-cortex-success/20 border border-cortex-success/40 text-cortex-success text-xs font-mono hover:bg-cortex-success/30 transition-colors flex items-center gap-1.5"
                >
                    <CheckCircle className="w-3 h-3" />
                    Confirm &amp; Execute
                </button>
                <button
                    className="px-3 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10 transition-colors flex items-center gap-1.5"
                >
                    <Edit3 className="w-3 h-3" />
                    Modify Plan
                </button>
                <button
                    onClick={() => cancelProposal()}
                    className="px-3 py-1.5 rounded text-red-400 text-xs font-mono hover:bg-red-400/10 transition-colors flex items-center gap-1.5"
                >
                    <XCircle className="w-3 h-3" />
                    Cancel
                </button>
            </div>
        </div>
    );
}
