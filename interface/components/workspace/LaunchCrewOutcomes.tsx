"use client";

import React from "react";
import { Loader2, CheckCircle, AlertTriangle, Sparkles } from "lucide-react";
import { type ChatArtifactRef, type ChatMessage, type ProposalData } from "@/store/useCortexStore";
import { type LaunchCrewState, OutcomeArtifacts, OutcomeBadge } from "./LaunchCrewShared";

interface OutcomeBodyProps {
    state: LaunchCrewState;
    assistantName: string;
    intent: string;
    pendingProposal: ProposalData | null;
    outcomeMessage: ChatMessage | null;
    outcomeArtifacts: ChatArtifactRef[];
    activeRunId: string | null;
    blockerMessage: string;
}

export function LaunchCrewOutcomeBody({
    state,
    assistantName,
    intent,
    pendingProposal,
    outcomeMessage,
    outcomeArtifacts,
    activeRunId,
    blockerMessage,
}: OutcomeBodyProps) {
    return (
        <>
            {state === "waiting" && (
                <div className="space-y-4">
                    <div className="flex flex-col items-center justify-center gap-3 py-6">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full border border-cortex-primary/20 bg-cortex-primary/10">
                            <Loader2 className="h-5 w-5 animate-spin text-cortex-primary" />
                        </div>
                        <div className="text-center">
                            <p className="font-mono text-sm font-bold text-cortex-text-main">
                                {assistantName} is evaluating your request
                            </p>
                            <p className="mt-1 text-[10px] font-mono text-cortex-text-muted">
                                Launch Crew must return a proposal, direct answer, execution result, or blocker.
                            </p>
                        </div>
                    </div>
                    <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-2.5">
                        <p className="mb-1 text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Your intent</p>
                        <p className="text-[10px] font-mono leading-relaxed text-cortex-text-main">{intent}</p>
                    </div>
                </div>
            )}

            {state === "proposal" && pendingProposal && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                            <CheckCircle className="h-4 w-4 text-amber-400" />
                            <p className="font-mono text-sm font-bold text-cortex-text-main">
                                {assistantName} prepared a crew proposal
                            </p>
                        </div>
                        <OutcomeBadge mode="proposal" />
                    </div>

                    <div className="rounded-lg border border-cortex-border bg-cortex-bg">
                        <div className="flex items-center justify-between border-b border-cortex-border/50 px-3 py-2">
                            <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Intent</span>
                            <span className="max-w-[200px] truncate text-[10px] font-mono text-cortex-text-main">
                                {pendingProposal.intent}
                            </span>
                        </div>
                        <div className="flex items-center justify-between border-b border-cortex-border/50 px-3 py-2">
                            <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Risk</span>
                            <span className="text-[10px] font-mono font-bold uppercase text-amber-400">
                                {pendingProposal.risk_level}
                            </span>
                        </div>
                        <div className="px-3 py-2">
                            <span className="mb-2 block text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Tools</span>
                            <div className="flex flex-wrap gap-1">
                                {pendingProposal.tools.length > 0 ? (
                                    pendingProposal.tools.map((tool) => (
                                        <span
                                            key={tool}
                                            className="rounded border border-cortex-primary/20 bg-cortex-primary/10 px-1.5 py-0.5 text-[8px] font-mono text-cortex-primary"
                                        >
                                            {tool}
                                        </span>
                                    ))
                                ) : (
                                    <span className="text-[10px] font-mono text-cortex-text-muted">No explicit tool list provided.</span>
                                )}
                            </div>
                        </div>
                    </div>
                    <p className="text-[10px] font-mono text-cortex-text-muted">
                        Confirm to execute now, or review the same proposal in Workspace chat.
                    </p>
                </div>
            )}

            {state === "answer" && outcomeMessage && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                            <p className="font-mono text-sm font-bold text-cortex-text-main">
                                {assistantName} answered directly
                            </p>
                        </div>
                        <OutcomeBadge mode="answer" />
                    </div>
                    <div className="rounded-lg border border-cortex-border bg-cortex-bg px-3 py-3">
                        <p className="mb-2 text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">
                            No crew launch was required
                        </p>
                        <p className="text-[10px] font-mono leading-relaxed text-cortex-text-main">
                            {outcomeMessage.content}
                        </p>
                    </div>
                    <OutcomeArtifacts artifacts={outcomeArtifacts} />
                </div>
            )}

            {state === "execution_result" && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                            <CheckCircle className="h-4 w-4 text-cortex-success" />
                            <p className="font-mono text-sm font-bold text-cortex-text-main">Crew launch submitted</p>
                        </div>
                        <OutcomeBadge mode="execution_result" />
                    </div>
                    <div className="rounded-lg border border-cortex-success/30 bg-cortex-success/5 px-3 py-3">
                        <p className="mb-1 text-[9px] font-mono uppercase tracking-widest text-cortex-success">Backend transaction complete</p>
                        <p className="text-[10px] font-mono leading-relaxed text-cortex-text-main">
                            {outcomeMessage?.content?.trim()
                                ? outcomeMessage.content
                                : activeRunId
                                ? `Mission run ${activeRunId.slice(0, 8)}... was activated and is ready to inspect.`
                                : "The launch completed and is recorded in Workspace chat."}
                        </p>
                    </div>
                    <OutcomeArtifacts artifacts={outcomeArtifacts} />
                </div>
            )}

            {state === "confirmed_pending_execution" && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                            <Loader2 className="h-4 w-4 text-cortex-primary" />
                            <p className="font-mono text-sm font-bold text-cortex-text-main">Proposal confirmed</p>
                        </div>
                        <OutcomeBadge mode="confirmed_pending_execution" />
                    </div>
                    <div className="rounded-lg border border-cortex-primary/30 bg-cortex-primary/5 px-3 py-3">
                        <p className="mb-1 text-[9px] font-mono uppercase tracking-widest text-cortex-primary">Waiting for durable proof</p>
                        <p className="text-[10px] font-mono leading-relaxed text-cortex-text-main">
                            {outcomeMessage?.content || "The proposal was confirmed, but Workspace has not attached a run or durable execution proof yet."}
                        </p>
                    </div>
                    <p className="text-[10px] font-mono text-cortex-text-muted">
                        Keep tracking this request in Workspace chat until a verified run or execution record appears.
                    </p>
                </div>
            )}

            {state === "blocker" && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                            <AlertTriangle className="h-4 w-4 text-cortex-danger" />
                            <p className="font-mono text-sm font-bold text-cortex-text-main">Launch Crew is blocked</p>
                        </div>
                        <OutcomeBadge mode="blocker" />
                    </div>
                    <div className="rounded-lg border border-cortex-danger/30 bg-cortex-danger/5 px-3 py-3">
                        <p className="mb-1 text-[9px] font-mono uppercase tracking-widest text-cortex-danger">Action required</p>
                        <p className="text-[10px] font-mono leading-relaxed text-cortex-text-main">{blockerMessage}</p>
                    </div>
                    <p className="text-[10px] font-mono text-cortex-text-muted">
                        Retry after adjusting the request, or continue in Workspace chat for more context.
                    </p>
                </div>
            )}
        </>
    );
}
