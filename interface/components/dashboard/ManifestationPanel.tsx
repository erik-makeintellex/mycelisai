"use client";

import React, { useEffect } from "react";
import { Sparkles, Users, Check, X, Clock } from "lucide-react";
import { useCortexStore, type TeamProposal } from "@/store/useCortexStore";

function ProposalCard({ proposal }: { proposal: TeamProposal }) {
    const approve = useCortexStore((s) => s.approveProposal);
    const reject = useCortexStore((s) => s.rejectProposal);

    const isPending = proposal.status === "pending";

    return (
        <div className={`
            p-3 rounded-lg border transition-all
            ${isPending
                ? "bg-cortex-bg border-cortex-primary/30 hover:border-cortex-primary/60"
                : proposal.status === "approved"
                    ? "bg-cortex-success/5 border-cortex-success/20 opacity-60"
                    : "bg-cortex-bg border-cortex-border opacity-40"
            }
        `}>
            {/* Header */}
            <div className="flex items-center gap-2 mb-2">
                <Users className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-xs font-mono font-bold text-cortex-text-main truncate">
                    {proposal.name}
                </span>
                <span className={`ml-auto text-[10px] font-mono px-1.5 py-0.5 rounded ${
                    isPending
                        ? "bg-cortex-primary/10 text-cortex-primary"
                        : proposal.status === "approved"
                            ? "bg-cortex-success/10 text-cortex-success"
                            : "bg-cortex-border text-cortex-text-muted"
                }`}>
                    {proposal.status.toUpperCase()}
                </span>
            </div>

            {/* Role */}
            <div className="text-[10px] text-cortex-text-muted font-mono mb-1">
                Role: {proposal.role}
            </div>

            {/* Reason */}
            <p className="text-xs text-cortex-text-main mb-2 line-clamp-2">
                {proposal.reason}
            </p>

            {/* Agent count */}
            <div className="text-[10px] text-cortex-text-muted font-mono mb-2">
                {proposal.agents.length} agent{proposal.agents.length !== 1 ? "s" : ""} proposed
            </div>

            {/* Actions */}
            {isPending && (
                <div className="flex gap-2">
                    <button
                        onClick={() => approve(proposal.id)}
                        className="flex-1 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded bg-cortex-success/10 border border-cortex-success/30 text-cortex-success text-xs font-mono hover:bg-cortex-success/20 transition-colors"
                    >
                        <Check className="w-3 h-3" />
                        MANIFEST
                    </button>
                    <button
                        onClick={() => reject(proposal.id)}
                        className="flex-1 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-mono hover:bg-red-500/20 transition-colors"
                    >
                        <X className="w-3 h-3" />
                        DISMISS
                    </button>
                </div>
            )}
        </div>
    );
}

export default function ManifestationPanel() {
    const proposals = useCortexStore((s) => s.teamProposals);
    const isFetching = useCortexStore((s) => s.isFetchingProposals);
    const fetchProposals = useCortexStore((s) => s.fetchProposals);

    const pendingCount = proposals.filter((p) => p.status === "pending").length;

    useEffect(() => {
        fetchProposals();
        const interval = setInterval(fetchProposals, 10000);
        return () => clearInterval(interval);
    }, [fetchProposals]);

    return (
        <div className="h-full flex flex-col bg-cortex-surface" data-testid="manifestation-panel">
            {/* Header */}
            <div className="px-3 py-2 border-b border-cortex-border flex items-center gap-2">
                <Sparkles className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-xs font-mono font-bold text-cortex-text-muted">TEAM MANIFESTATION</span>
                {pendingCount > 0 && (
                    <span className="ml-auto text-[10px] font-mono text-cortex-primary bg-cortex-primary/10 px-1.5 py-0.5 rounded animate-pulse">
                        {pendingCount} PENDING
                    </span>
                )}
            </div>

            {/* Proposals List */}
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
                {isFetching && proposals.length === 0 && (
                    <div className="space-y-2">
                        {[1, 2].map((i) => (
                            <div key={i} className="h-28 rounded-lg bg-cortex-bg animate-pulse" />
                        ))}
                    </div>
                )}

                {!isFetching && proposals.length === 0 && (
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        <Clock className="w-8 h-8 mb-2 opacity-30" />
                        <p className="text-xs font-mono">No team proposals</p>
                        <p className="text-[10px] mt-1 opacity-50">Mother Brain is thinking...</p>
                    </div>
                )}

                {proposals.map((proposal) => (
                    <ProposalCard key={proposal.id} proposal={proposal} />
                ))}
            </div>
        </div>
    );
}
