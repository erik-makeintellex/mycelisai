"use client";

import React from "react";
import { Shield, AlertTriangle, CheckCircle, ChevronDown, ChevronUp, Clock3, ExternalLink, Eye, FileText, XCircle } from "lucide-react";
import { useCortexStore, type ChatMessage, type ProposalData } from "@/store/useCortexStore";
import { formatProposalHeadline } from "@/lib/labels";
import ProposalRunIntent from "./ProposalRunIntent";
import ProposedActionDetails from "./ProposedActionDetails";
import { artifactDownloadHref } from "./missionControlChatHelpers";

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
        if (proposal.capability_ids?.includes("write_file")) return "This action will change your workspace, so Soma needs your approval before running it.";
        if (proposal.external_data_use) return "This action may use external systems or data, so Soma needs your approval before running it.";
        if (proposal.approval_reason === "cost_threshold") return "This action may incur additional spend, so Soma needs your approval before running it.";
        return "This action crosses a governed policy threshold, so Soma needs your approval before running it.";
    }

    return approvalMode === "optional"
        ? "This action stays within current policy thresholds, but you can still review it before execution."
        : "This action is within current policy thresholds and can run without a mandatory approval.";
}

export default function ProposedActionBlock({ message, simpleMode = false }: { message: ChatMessage; simpleMode?: boolean }) {
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const assistantName = useCortexStore((s) => s.assistantName);
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const isSimpleDevView = simpleMode && !advancedMode;
    const [detailsOpen, setDetailsOpen] = React.useState(false);
    const [previewingArtifactId, setPreviewingArtifactId] = React.useState<string | null>(null);

    const proposal = message.proposal;
    if (!proposal) return null;
    const lifecycle = message.proposal_status ?? "active";
    const hasRunProof = Boolean(message.run_id?.trim());
    const renderedLifecycle = lifecycle === "executed" && !hasRunProof ? "confirmed_pending_execution" : lifecycle;
    const isActionable = renderedLifecycle === "active";
    const approvalRequired = proposal.approval_required ?? true;
    const approvalMode = proposal.approval_mode ?? (approvalRequired ? "required" : "auto_allowed");
    const governanceSummary = approvalRequired ? "Approval required" : approvalMode === "optional" ? "Approval optional" : "Auto-approved";
    const actionLabel = approvalRequired ? "Approve & Execute" : "Execute";
    const headline = formatProposalHeadline(proposal, assistantName);
    const operatorSummary = proposal.operator_summary?.trim() || fallbackOperatorSummary(proposal);
    const expectedResult = proposal.expected_result?.trim() || fallbackExpectedResult(proposal);
    const affectedResources = (proposal.affected_resources ?? []).filter((value) => value.trim().length > 0);
    const visibleResources = affectedResources.length > 0 ? affectedResources : fallbackAffectedResources(proposal);
    const approvalExplanation = explainApprovalPosture(proposal, approvalRequired, approvalMode);
    const attachedArtifacts = message.artifacts ?? [];

    const lifecycleTone = renderedLifecycle === "cancelled"
        ? "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted"
        : renderedLifecycle === "confirmed_pending_execution"
            ? "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary"
            : renderedLifecycle === "executed"
                ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
                : renderedLifecycle === "failed"
                    ? "border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger"
                    : "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning";

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
        <div className="mt-3 rounded-lg border border-cortex-warning/30 bg-cortex-surface/80 overflow-hidden">
            <div className="px-4 py-2.5 bg-cortex-warning/5 border-b border-cortex-warning/20 flex items-center justify-between gap-2 flex-wrap">
                <div className="flex items-center gap-2">
                    <Shield className="w-4 h-4 text-cortex-warning" />
                    <span className="text-cortex-warning font-mono text-xs font-bold tracking-wider">PROPOSED ACTION</span>
                </div>
                <span className={`px-2 py-0.5 rounded border text-[10px] font-mono ${approvalRequired ? "text-cortex-warning border-cortex-warning/30" : "text-cortex-success border-cortex-success/30"}`}>
                    {governanceSummary.toUpperCase()}
                </span>
            </div>

            <div className="px-4 py-3 space-y-3">
                <div className={`flex items-center gap-2 rounded border px-2.5 py-2 text-[10px] ${lifecycleTone}`}>
                    <LifecycleIcon className="h-3.5 w-3.5 flex-shrink-0" />
                    <span>{lifecycleLabel}</span>
                </div>

                <div className="space-y-1">
                    <div className="text-xs font-semibold text-cortex-text-main">{headline}</div>
                    <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Soma wants to</div>
                    <p className="text-sm leading-6 text-cortex-text-main">{operatorSummary}</p>
                </div>

                {!isSimpleDevView && <ProposalRunIntent proposal={proposal} />}

                {isSimpleDevView ? (
                    <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">What happens when you click {actionLabel}</div>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-main">{expectedResult}</p>
                    </div>
                ) : (
                    <div className="grid gap-3 md:grid-cols-2">
                        <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                            <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{approvalRequired ? "Why approval is needed" : "Execution posture"}</div>
                            <p className="mt-1.5 text-sm leading-6 text-cortex-text-main">{approvalExplanation}</p>
                        </div>
                        <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                            <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Expected result</div>
                            <p className="mt-1.5 text-sm leading-6 text-cortex-text-main">{expectedResult}</p>
                        </div>
                    </div>
                )}

                {visibleResources.length > 0 ? (
                    <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                            {isSimpleDevView ? "Files & teams to build" : "What will change"}
                        </div>
                        <div className="mt-1.5 flex flex-wrap gap-1.5">
                            {visibleResources.map((resource) => (
                                <span
                                    key={resource}
                                    className="rounded border border-cortex-border bg-cortex-bg/70 px-2 py-0.5 text-xs text-cortex-text-main"
                                >
                                    {resource}
                                </span>
                            ))}
                        </div>
                    </div>
                ) : null}

                {attachedArtifacts.length > 0 ? (
                    <div className="space-y-2 rounded border border-cortex-primary/30 bg-cortex-primary/5 p-3">
                        <div className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                            Deliverable Previews
                        </div>
                        {attachedArtifacts.map((artifact, idx) => {
                            const key = artifact.id || `prop-art-${idx}`;
                            const href = artifactDownloadHref(artifact);
                            const isPreviewOpen = previewingArtifactId === key;
                            return (
                                <div key={key} className="rounded border border-cortex-border bg-cortex-bg p-2.5 space-y-2">
                                    <div className="flex items-center justify-between gap-2 flex-wrap">
                                        <div className="flex items-center gap-2 min-w-0">
                                            <FileText className="w-3.5 h-3.5 text-cortex-primary flex-shrink-0" />
                                            <span className="text-xs font-medium text-cortex-text-main truncate">{artifact.title}</span>
                                        </div>
                                        <div className="flex items-center gap-1.5">
                                            {artifact.content ? (
                                                <button
                                                    type="button"
                                                    onClick={() => setPreviewingArtifactId(isPreviewOpen ? null : key)}
                                                    className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-[10px] font-mono text-cortex-text-main hover:border-cortex-primary/40"
                                                >
                                                    <Eye className="w-3 h-3" />
                                                    Preview
                                                </button>
                                            ) : null}
                                            {href ? (
                                                <a
                                                    href={href}
                                                    download
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    className="inline-flex items-center gap-1 rounded border border-cortex-primary/40 bg-cortex-primary/10 px-2 py-1 text-[10px] font-mono text-cortex-primary hover:bg-cortex-primary/20"
                                                >
                                                    <ExternalLink className="w-3 h-3" />
                                                    Download / Open
                                                </a>
                                            ) : null}
                                        </div>
                                    </div>
                                    {isPreviewOpen && artifact.content ? (
                                        <pre className="max-h-40 overflow-auto rounded border border-cortex-border/60 bg-cortex-surface/60 p-2 text-[11px] font-mono text-cortex-text-main">
                                            {artifact.content}
                                        </pre>
                                    ) : null}
                                </div>
                            );
                        })}
                    </div>
                ) : null}

                <button
                    type="button"
                    onClick={() => setDetailsOpen((open) => !open)}
                    className="inline-flex items-center gap-1.5 text-xs font-mono text-cortex-primary hover:text-cortex-primary/80 transition-colors"
                    aria-expanded={detailsOpen}
                >
                    {detailsOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                    {detailsOpen ? "Hide Proof Details" : "View Proof Details"}
                </button>

                {detailsOpen ? (
                    <ProposedActionDetails message={message} proposal={proposal} assistantName={assistantName} />
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
                        className="px-3 py-1.5 rounded text-cortex-danger text-xs font-mono hover:bg-cortex-danger/10 transition-colors flex items-center gap-1.5"
                    >
                        <XCircle className="w-3 h-3" />
                        Cancel
                    </button>
                </div>
            ) : null}
        </div>
    );
}
