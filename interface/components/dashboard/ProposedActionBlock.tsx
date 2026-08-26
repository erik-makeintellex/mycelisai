"use client";

import { useEffect, useRef, useState } from "react";
import { Check, ChevronDown, ChevronUp } from "lucide-react";
import { useCortexStore, type ChatMessage } from "@/store/useCortexStore";
import ProposedActionDetails from "./ProposedActionDetails";
import ProposalLifecycleProof from "./ProposalLifecycleProof";
import {
    explainApprovalPosture,
    fallbackAffectedResources,
    fallbackExpectedResult,
    fallbackOperatorSummary,
    plainExecutionText,
} from "./proposedActionCopy";

function primaryTeamLabel(proposal: ChatMessage["proposal"]): string {
    const affectedTeam = proposal?.affected_resources
        ?.map((value) => value.trim())
        .find((value) => value.toLowerCase().startsWith("team:"))
        ?.slice("team:".length)
        .trim();
    if (affectedTeam) return plainExecutionText(affectedTeam);

    const expression = proposal?.team_expressions?.find((item) => {
        const teamId = item.team_id?.trim();
        return (teamId && teamId !== "admin-core") || item.objective?.trim();
    });
    const teamId = expression?.team_id?.trim();
    if (teamId && teamId !== "admin-core") return plainExecutionText(teamId);
    return "the right team";
}

function sentence(value: string): string {
    const trimmed = value.trim();
    return trimmed ? `${trimmed.charAt(0).toUpperCase()}${trimmed.slice(1)}` : "";
}

function approvalPlanBullets(proposal: ChatMessage["proposal"], operatorSummary: string, expectedResult: string): string[] {
    const tools = proposal?.tools ?? [];
    const teamLabel = primaryTeamLabel(proposal);
    const bullets: string[] = [];
    if (expectedResult) bullets.push(sentence(expectedResult));
    if (operatorSummary) bullets.push(sentence(operatorSummary));
    if (tools.some((tool) => /^(create_team|delegate|delegate_task)$/i.test(tool))) {
        bullets.push(`Bring in ${teamLabel} and keep their work connected to this conversation.`);
    }
    if (!bullets.length) bullets.push("Start the work and bring the result back to this conversation.");
    return [...new Set(bullets)].slice(0, 3);
}

function isDelegatedOrAsync(proposal: NonNullable<ChatMessage["proposal"]>): boolean {
    return proposal.tools.some((tool) => /^(create_team|delegate|delegate_task)$/i.test(tool))
        || proposal.execution_mode === "team_async"
        || proposal.execution_mode === "schedule_handoff"
        || ["scheduled", "continuous", "event_driven"].includes(proposal.work_intent?.cadence ?? proposal.task_cadence ?? "");
}

function hasVerifiedTerminalResult(message: ChatMessage): boolean {
    const status = (message.execution_summary?.execution?.status ?? message.execution_summary?.execution_status ?? "").trim().toLowerCase();
    if (status === "verified") return true;
    if (!["complete", "completed", "success", "succeeded"].includes(status)) return false;

    const proof = message.execution_summary?.proof;
    const proofs = Array.isArray(proof) ? proof : proof ? [proof] : [];
    return proofs.some((item) => typeof item === "object" && item.verified === true);
}

export default function ProposedActionBlock({ message }: { message: ChatMessage }) {
    const assistantName = useCortexStore((s) => s.assistantName);
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const [detailsOpen, setDetailsOpen] = useState(false);
    const [approvalPending, setApprovalPending] = useState(false);
    const [approvalError, setApprovalError] = useState<string | null>(null);
    const detailsRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!detailsOpen) return;
        const frame = window.requestAnimationFrame(() => {
            detailsRef.current?.scrollIntoView?.({ block: "end", behavior: "smooth" });
        });
        return () => window.cancelAnimationFrame(frame);
    }, [detailsOpen]);

    const proposal = message.proposal;
    if (!proposal) return null;
    const lifecycle = message.proposal_status ?? "active";
    const hasRunProof = Boolean(message.run_id?.trim());
    const delegatedOrAsync = isDelegatedOrAsync(proposal);
    const verifiedTerminalResult = hasVerifiedTerminalResult(message);
    const renderedLifecycle = lifecycle === "executed" && (!hasRunProof || (delegatedOrAsync && !verifiedTerminalResult))
        ? "confirmed_pending_execution"
        : lifecycle;
    const isActionable = renderedLifecycle === "active";
    const hasConfirmToken = Boolean(proposal.confirm_token?.trim());
    const hasIntentProof = Boolean(proposal.intent_proof_id?.trim());
    const canRunProposal = hasConfirmToken && hasIntentProof;
    const approvalRequired = proposal.approval_required ?? true;
    const approvalMode = proposal.approval_mode ?? (approvalRequired ? "required" : "auto_allowed");
    const capabilityRisk = proposal.capability_risk ?? proposal.risk_level ?? "low";
    const operatorSummary = plainExecutionText(proposal.operator_summary?.trim() || fallbackOperatorSummary(proposal));
    const expectedResult = plainExecutionText(proposal.expected_result?.trim() || fallbackExpectedResult(proposal));
    const affectedResources = (proposal.affected_resources ?? []).filter((value) => value.trim().length > 0);
    const visibleResources = (affectedResources.length > 0 ? affectedResources : fallbackAffectedResources(proposal)).map(plainExecutionText);
    const approvalExplanation = explainApprovalPosture(proposal, approvalRequired, approvalMode);
    const planBullets = approvalPlanBullets(proposal, operatorSummary, expectedResult);
    const runHelp = approvalRequired
        ? "Once you approve, I’ll start the work and stay here for questions or changes while it runs."
        : "This is ready to start. I’ll stay here for questions or changes while it runs.";

    const approve = async () => {
        if (approvalPending || !canRunProposal) return;
        setApprovalPending(true);
        setApprovalError(null);
        const result = await confirmProposal(proposal, approvalRequired ? "approve" : "start");
        if (!result.ok) setApprovalError(result.error ?? "Soma could not start this proposal.");
        setApprovalPending(false);
    };

    const lifecycleLabel = renderedLifecycle === "cancelled"
            ? "Cancelled"
        : renderedLifecycle === "confirmed_pending_execution"
            ? "Confirmed, waiting for result"
            : renderedLifecycle === "executed"
                ? delegatedOrAsync ? "Result verified" : "Action completed"
                : renderedLifecycle === "failed"
                    ? "Could not run"
                    : "Awaiting approval";


    return (
        <div className="mt-3 max-w-[min(100%,680px)] border-l-2 border-cortex-primary/50 py-1 pl-4 pr-2">
            <div className="space-y-3">
                <div>
                    <div className="text-xs font-semibold text-cortex-text-muted">Soma</div>
                    <h3 className="mt-0.5 text-base font-semibold text-cortex-text-main">
                        {isActionable ? "I can start that." : lifecycleLabel}
                    </h3>
                </div>
                {!isActionable ? <ProposalLifecycleProof lifecycle={renderedLifecycle} runId={message.run_id} /> : null}

                <div className="space-y-2">
                    {isActionable ? (
                        <p className="text-sm leading-6 text-cortex-text-muted">{runHelp}</p>
                    ) : null}
                    <ul className="space-y-1.5 text-sm leading-6 text-cortex-text-main">
                        {planBullets.map((item, index) => (
                            <li key={`${item}-${index}`} className="flex gap-2">
                                <span className="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-cortex-primary" />
                                <span>{item}</span>
                            </li>
                        ))}
                    </ul>
                </div>

                {isActionable ? (
                    canRunProposal ? (
                        <div className="flex flex-wrap items-center gap-3">
                            <button
                                type="button"
                                onClick={() => void approve()}
                                disabled={approvalPending}
                                className="inline-flex items-center gap-2 rounded-lg bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90 disabled:cursor-wait disabled:opacity-60"
                            >
                                <Check className="h-4 w-4" />
                                {approvalPending ? "Starting…" : approvalRequired ? "Approve" : "Start"}
                            </button>
                            <p className="text-sm font-medium leading-6 text-cortex-primary">
                                Or reply “{approvalRequired ? "approve" : "start"}”. You can also ask a question or request a change.
                            </p>
                        </div>
                    ) : (
                        <p className="text-sm leading-6 text-red-300">
                            I cannot start this version yet. Ask me to regenerate the proposal before approving it.
                        </p>
                    )
                ) : null}

                {approvalError ? <p role="alert" className="text-sm leading-6 text-red-300">{approvalError}</p> : null}

                <button
                    type="button"
                    onClick={() => setDetailsOpen((open) => !open)}
                    className="inline-flex items-center gap-1.5 text-sm font-semibold text-cortex-primary transition-colors hover:text-cortex-primary/80"
                    aria-expanded={detailsOpen}
                >
                    {detailsOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                    {detailsOpen ? "Hide details" : "Details"}
                </button>

                {detailsOpen ? (
                    <div ref={detailsRef} className="space-y-2 scroll-mb-3">
                        <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-3 text-xs">
                            <div className="grid gap-2 md:grid-cols-2">
                                <div>
                                    <div className="font-mono uppercase tracking-[0.14em] text-cortex-text-muted text-[10px]">Confirmation</div>
                                    <p className="mt-1 text-cortex-text-main">{approvalExplanation}</p>
                                </div>
                                <div>
                                    <div className="font-mono uppercase tracking-[0.14em] text-cortex-text-muted text-[10px]">Risk and cost</div>
                                    <p className="mt-1 text-cortex-text-main">
                                        Risk: {plainExecutionText(capabilityRisk)}{typeof proposal.estimated_cost === "number" ? `, estimated cost ${proposal.estimated_cost.toFixed(2)}` : ""}
                                    </p>
                                </div>
                            </div>
                            {visibleResources.length > 0 || proposal.external_data_use ? (
                                <div className="mt-3 flex flex-wrap items-center gap-1.5">
                                    <span className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">Will touch</span>
                                    {visibleResources.map((resource, index) => (
                                        <span
                                            key={`${resource}-${index}`}
                                            className="rounded border border-cortex-border bg-cortex-bg/70 px-2 py-1 text-xs text-cortex-text-main"
                                        >
                                            {resource}
                                        </span>
                                    ))}
                                    {proposal.external_data_use ? (
                                        <span className="rounded border border-cortex-primary/30 px-2 py-1 text-[10px] font-mono text-cortex-primary">
                                            EXTERNAL DATA
                                        </span>
                                    ) : null}
                                </div>
                            ) : null}
                        </div>
                        <ProposedActionDetails assistantName={assistantName} message={message} proposal={proposal} />
                    </div>
                ) : null}
            </div>
        </div>
    );
}
