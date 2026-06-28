"use client";

import { useState } from "react";
import { Shield, AlertTriangle, CheckCircle, ChevronDown, ChevronUp, Clock3, Loader2, XCircle } from "lucide-react";
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

function approvalPlanBullets(proposal: ChatMessage["proposal"], operatorSummary: string, expectedResult: string): string[] {
    const tools = proposal?.tools ?? [];
    const teamLabel = primaryTeamLabel(proposal);
    const bullets: string[] = [];
    if (tools.some((tool) => /^(create_team|delegate|delegate_task)$/i.test(tool))) {
        bullets.push(`Hand the work to ${teamLabel} through the team bus.`);
    }
    if (operatorSummary) bullets.push(operatorSummary);
    if (expectedResult) bullets.push(expectedResult);
    if (!bullets.length) bullets.push("Start the approved work and keep the result tied to this conversation.");
    return [...new Set(bullets)].slice(0, 3);
}

export default function ProposedActionBlock({ message }: { message: ChatMessage }) {
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const assistantName = useCortexStore((s) => s.assistantName);
    const [detailsOpen, setDetailsOpen] = useState(false);
    const [confirming, setConfirming] = useState(false);
    const [confirmError, setConfirmError] = useState<string | null>(null);

    const proposal = message.proposal;
    if (!proposal) return null;
    const lifecycle = message.proposal_status ?? "active";
    const hasRunProof = Boolean(message.run_id?.trim());
    const renderedLifecycle = lifecycle === "executed" && !hasRunProof ? "confirmed_pending_execution" : lifecycle;
    const isActionable = renderedLifecycle === "active";
    const hasConfirmToken = Boolean(proposal.confirm_token?.trim());
    const hasIntentProof = Boolean(proposal.intent_proof_id?.trim());
    const canRunProposal = hasConfirmToken && hasIntentProof;
    const approvalRequired = proposal.approval_required ?? true;
    const approvalMode = proposal.approval_mode ?? (approvalRequired ? "required" : "auto_allowed");
    const capabilityRisk = proposal.capability_risk ?? proposal.risk_level ?? "low";
    const governanceSummary = approvalRequired ? "Needs approval" : approvalMode === "optional" ? "Ready if you want" : "Ready";
    const actionLabel = approvalRequired ? "Approve" : "Start";
    const operatorSummary = plainExecutionText(proposal.operator_summary?.trim() || fallbackOperatorSummary(proposal));
    const expectedResult = plainExecutionText(proposal.expected_result?.trim() || fallbackExpectedResult(proposal));
    const affectedResources = (proposal.affected_resources ?? []).filter((value) => value.trim().length > 0);
    const visibleResources = (affectedResources.length > 0 ? affectedResources : fallbackAffectedResources(proposal)).map(plainExecutionText);
    const approvalExplanation = explainApprovalPosture(proposal, approvalRequired, approvalMode);
    const planBullets = approvalPlanBullets(proposal, operatorSummary, expectedResult);
    const runQuestion = approvalRequired ? "Approve this?" : "Start this?";
    const runHelp = approvalRequired
        ? "I will hand this to the bus after approval and keep this thread open for questions while the team works."
        : "This stays inside current policy. I will start the handoff and keep the thread live.";
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
            ? "Confirmed, waiting for result"
            : renderedLifecycle === "executed"
                ? "Action completed"
                : renderedLifecycle === "failed"
                    ? "Could not run"
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
    const handleConfirm = async () => {
        if (!canRunProposal || confirming) return;
        setConfirming(true);
        setConfirmError(null);
        const result = await confirmProposal(proposal);
        setConfirming(false);
        if (!result.ok) {
            setConfirmError(plainExecutionText(result.error || "Soma could not run this. Review the blocker below."));
        }
    };

    const handleCancel = () => {
        if (confirming) return;
        cancelProposal();
    };

    return (
        <div className="mt-3 max-w-[min(100%,680px)] overflow-hidden rounded-xl border border-cortex-border bg-cortex-surface/90 shadow-sm">
            <div className="border-b border-cortex-border px-3.5 py-3">
                <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0">
                        <div className="text-xs font-semibold text-cortex-text-muted">Soma</div>
                        <div className="mt-0.5 text-base font-semibold text-cortex-text-main">
                            I can start that.
                        </div>
                    </div>
                    <span className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-semibold ${lifecycleTone}`}>
                        <LifecycleIcon className="h-3.5 w-3.5 flex-shrink-0" />
                        {lifecycleLabel}
                    </span>
                </div>
                <span className="mt-2 inline-flex rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-1 text-[11px] font-semibold text-cortex-text-muted">
                    {governanceSummary}
                </span>
            </div>

            <div className="space-y-3 px-3.5 py-3.5">
                {!isActionable ? <ProposalLifecycleProof lifecycle={renderedLifecycle} runId={message.run_id} /> : null}

                <div className="space-y-2">
                    {isActionable ? (
                        <div>
                            <h3 className="text-base font-semibold text-cortex-text-main">{runQuestion}</h3>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{runHelp}</p>
                        </div>
                    ) : null}
                    <ul className="space-y-1.5 text-sm leading-6 text-cortex-text-main">
                        {planBullets.map((item) => (
                            <li key={item} className="flex gap-2">
                                <span className="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-cortex-primary" />
                                <span>{item}</span>
                            </li>
                        ))}
                    </ul>
                </div>

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
                    <div className="space-y-2">
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
                                    {visibleResources.map((resource) => (
                                        <span
                                            key={resource}
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

            {isActionable ? (
                <div className="space-y-2 border-t border-cortex-border px-3.5 py-3">
                    <div className="flex flex-wrap items-center gap-2">
                        <button
                            onClick={() => void handleConfirm()}
                            disabled={!canRunProposal || confirming}
                            className="flex items-center gap-1.5 rounded-lg border border-cortex-success/40 bg-cortex-success/15 px-3 py-1.5 text-sm font-semibold text-cortex-success transition-colors hover:bg-cortex-success/25 disabled:cursor-not-allowed disabled:border-cortex-border disabled:bg-cortex-bg/40 disabled:text-cortex-text-muted"
                        >
                            {confirming ? <Loader2 className="w-3 h-3 animate-spin" /> : <CheckCircle className="w-3 h-3" />}
                            {confirming ? "Starting..." : canRunProposal ? actionLabel : "Cannot run yet"}
                        </button>
                        <button
                            onClick={handleCancel}
                            disabled={confirming}
                            className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-semibold text-cortex-text-muted transition-colors hover:bg-cortex-bg hover:text-red-300 disabled:cursor-not-allowed disabled:text-cortex-text-muted"
                        >
                            <XCircle className="w-3 h-3" />
                            Adjust
                        </button>
                        {!canRunProposal ? (
                            <span className="text-[11px] text-cortex-text-muted">
                                This proposal is missing the information Soma needs to run it. Ask Soma to regenerate it.
                            </span>
                        ) : null}
                    </div>
                    {confirming ? (
                        <p className="text-[11px] leading-5 text-cortex-text-muted">
                            Handoff starting. You can keep talking to Soma while the team works.
                        </p>
                    ) : null}
                    {confirmError ? (
                        <p className="text-[11px] leading-5 text-red-300">{confirmError}</p>
                    ) : null}
                </div>
            ) : null}
        </div>
    );
}
