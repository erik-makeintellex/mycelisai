"use client";
import { AlertTriangle, Bot, Brain, Eye, Globe, Megaphone, MessageSquareReply, User, Zap } from "lucide-react";
import {
    sourceNodeLabel,
    trustBadge,
    trustTooltip,
    brainBadge,
    MODE_LABELS,
} from "@/lib/labels";
import { useCortexStore, type ChatConsultation, type ChatMessage } from "@/store/useCortexStore";
import InlineArtifact from "./InlineArtifact";
import ProposedActionBlock from "./ProposedActionBlock";
import ExecutionSummaryCard from "@/components/soma/ExecutionSummaryCard";
import ExecutionSummaryReceipt, { shouldUseExecutionSummaryReceipt } from "@/components/soma/ExecutionSummaryReceipt";
import MissionControlThreadStateCard from "./MissionControlThreadStateCard";
import MissionControlResponseDepth from "./MissionControlResponseDepth";
import MissionControlToolsUsed from "./MissionControlToolsUsed";
import { requestSomaOutputContinuation } from "@/components/soma/outputContinuation";
import {
    artifactResultSummary,
    askClassBadge,
    consultationResultSummary,
    COUNCIL_META,
    trustColor,
} from "./missionControlChatHelpers";
function DelegationTrace({ consultations, assistantName }: { consultations: ChatConsultation[]; assistantName: string }) {
    if (!consultations?.length) return null;
    return (
        <div className="mt-2 pt-2 border-t border-cortex-border/40">
            <div className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider mb-1.5">
                {assistantName} consulted
            </div>
            <div className="flex flex-wrap gap-1.5">
                {consultations.map((c, i) => {
                    const meta = COUNCIL_META[c.member];
                    return (
                        <div key={i} className="bg-cortex-surface/60 border border-cortex-border/60 rounded px-2 py-1.5 max-w-[180px]">
                            <div className={`text-[9px] font-mono font-bold mb-0.5 ${meta?.color ?? "text-cortex-text-muted"}`}>
                                {meta?.label ?? c.member}
                            </div>
                            <div className="text-[9px] text-cortex-text-muted leading-tight line-clamp-2">
                                {c.summary}
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

function MessageMeta({ msg, assistantName }: { msg: ChatMessage; assistantName: string }) {
    const setInspected = useCortexStore((s) => s.setInspectedMessage);
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const askBadge = askClassBadge(msg.ask_class);
    if (!msg.source_node && !msg.brain) return null;
    return (
        <div className="flex items-center gap-1.5 px-1 flex-wrap">
            {msg.source_node && (
                <span className="text-[8px] font-bold uppercase tracking-widest text-cortex-info font-mono">
                    {sourceNodeLabel(msg.source_node, assistantName)}
                </span>
            )}
            {advancedMode && msg.brain && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span
                        className={`text-[8px] font-mono font-bold uppercase tracking-wide flex items-center gap-1 ${
                            msg.brain.location === "remote" ? "text-amber-400" : "text-cortex-text-muted"
                        }`}
                        title={`Model: ${msg.brain.model_id}\nProvider: ${msg.brain.provider_id}\nLocation: ${msg.brain.location}\nData: ${msg.brain.data_boundary}`}
                    >
                        {msg.brain.location === "remote" ? <Globe className="w-2.5 h-2.5" /> : <Brain className="w-2.5 h-2.5" />}
                        {brainBadge(msg.brain.provider_id, msg.brain.location)}
                    </span>
                </>
            )}
            {advancedMode && msg.mode && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold uppercase tracking-wide ${MODE_LABELS[msg.mode]?.color ?? "text-cortex-text-muted"}`}>
                        {MODE_LABELS[msg.mode]?.label ?? msg.mode}
                    </span>
                </>
            )}
            {askBadge && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold px-1.5 py-0.5 rounded border ${askBadge.tone}`}>
                        {askBadge.label}
                    </span>
                </>
            )}
            {msg.trust_score != null && msg.trust_score > 0 && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold ${trustColor(msg.trust_score)}`} title={trustTooltip(msg.trust_score)}>
                        {trustBadge(msg.trust_score)}
                    </span>
                </>
            )}
            {advancedMode && msg.brain?.location === "remote" && (
                <span className="flex items-center gap-0.5 text-[7px] font-mono text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded border border-amber-400/20">
                    <AlertTriangle className="w-2.5 h-2.5" />
                    External
                </span>
            )}
            {advancedMode && (
                <button
                    onClick={() => setInspected(msg)}
                    className="p-0.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors ml-1"
                    title="Inspect orchestration"
                >
                    <Eye className="w-3 h-3" />
                </button>
            )}
        </div>
    );
}

function isAnswerOnlyDepth(msg: ChatMessage) {
    return Boolean(
        !msg.proposal
        && !msg.artifacts?.length
        && !msg.consultations?.length
        && msg.mode !== "execution_result"
        && (
            msg.response_depth === "quick_box"
            || msg.response_depth === "structured_summary"
            || msg.response_depth === "decision_brief"
        ),
    );
}
function hasAuditableAnswerEvidence(msg: ChatMessage) {
    const summary = msg.execution_summary;
    const status = summary?.execution?.status?.trim().toLowerCase();
    return Boolean(
        summary?.execution?.shape === "tool_assisted_work"
        && (status === "complete" || status === "completed")
        && summary.proof,
    );
}
function shouldShowSimpleThreadState(msg: ChatMessage) {
    const state = msg.ui_response_state ?? msg.execution_summary?.ui_response_state;
    const hasThreadEvents = Boolean(msg.thread_event || msg.thread_events?.length);
    return Boolean(
        msg.proposal
        || msg.ask_class === "execution_blocker"
        || state?.tone === "danger"
        || state?.tone === "warning"
        || hasThreadEvents
    );
}

export default function MissionControlMessageBubble({
    msg,
    compactResult = false,
}: {
    msg: ChatMessage;
    compactResult?: boolean;
}) {
    const isUser = msg.role === "user";
    const isBroadcast = isUser && msg.content.startsWith("[BROADCAST]");
    const assistantName = useCortexStore((s) => s.assistantName);
    const artifactSummary = artifactResultSummary(msg.artifacts);
    const consultationSummary = consultationResultSummary(msg.consultations);
    const responseBadge = askClassBadge(msg.ask_class);
    const answerOnlyDepth = isAnswerOnlyDepth(msg);
    const auditableAnswerEvidence = answerOnlyDepth && hasAuditableAnswerEvidence(msg);
    const showExecutionSummary = Boolean(
        msg.execution_summary
        && (
            auditableAnswerEvidence || (
                !answerOnlyDepth
                && (
                    msg.mode === "execution_result"
                    || msg.run_id
                    || msg.artifacts?.length
                    || msg.ask_class === "execution_blocker"
                )
            )
        ),
    );
    const hasStructuredEvidence = Boolean(
        msg.artifacts?.length || msg.consultations?.length,
    );
    const showAnswerExtras = !answerOnlyDepth;
    const showMeta = !isUser && !compactResult;
    const showTraceExtras = !isUser && showAnswerExtras && !compactResult;
    const showStructuredEvidence = !isUser
        && showAnswerExtras
        && (!compactResult || hasStructuredEvidence);
    const showThreadState = !isUser && showAnswerExtras && (!compactResult || shouldShowSimpleThreadState(msg));
    const useReceipt = (compactResult || auditableAnswerEvidence) && showExecutionSummary && msg.execution_summary
        ? shouldUseExecutionSummaryReceipt({
            summary: msg.execution_summary,
            runId: msg.run_id,
            artifacts: msg.artifacts,
        })
        : false;

    if (msg.role === "system") {
        return (
            <div className="my-1.5 flex justify-center">
                <div className="flex w-full max-w-[720px] flex-col items-center gap-1.5 px-2">
                    {msg.run_id ? (
                        <button
                            type="button"
                            onClick={() => requestSomaOutputContinuation({
                                title: "this requested work",
                                reference: `run:${msg.run_id}`,
                                proof: msg.run_id,
                                sourceLabel: "run",
                            })}
                            className="flex items-center gap-1.5 rounded-full border border-cortex-success/30 bg-cortex-success/5 px-2.5 py-1 font-mono text-[9px] text-cortex-success transition-colors hover:bg-cortex-success/10"
                        >
                            <Zap className="w-3 h-3" />
                            Continue with Soma
                            <MessageSquareReply className="h-2.5 w-2.5 opacity-60" />
                        </button>
                    ) : msg.thread_event || msg.thread_events?.length ? null : (
                        <span className="text-[9px] font-mono text-cortex-text-muted px-2.5 py-1 rounded-full border border-cortex-border">
                            {msg.content}
                        </span>
                    )}
                    <MissionControlThreadStateCard msg={msg} />
                    {showExecutionSummary && msg.execution_summary && (
                        <div className="w-full">
                            {useReceipt ? (
                                <ExecutionSummaryReceipt
                                    summary={msg.execution_summary}
                                    runId={msg.run_id}
                                    artifacts={msg.artifacts}
                                />
                            ) : (
                                <ExecutionSummaryCard
                                    summary={msg.execution_summary}
                                    runId={msg.run_id}
                                    artifacts={msg.artifacts}
                                    compact={compactResult}
                                />
                            )}
                        </div>
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className={`flex gap-2.5 ${isUser ? "justify-end" : "justify-start"}`}>
            {!isUser && (
                <div className="w-6 h-6 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <Bot className="w-3.5 h-3.5 text-cortex-info" />
                </div>
            )}
            <div className="max-w-[85%] flex flex-col gap-0.5">
                {showMeta && <MessageMeta msg={msg} assistantName={assistantName} />}
                <MissionControlResponseDepth
                    msg={msg}
                    isBroadcast={isBroadcast}
                    isUser={isUser}
                    compactResult={compactResult}
                />
                {showThreadState && <MissionControlThreadStateCard msg={msg} />}
                {showTraceExtras && msg.consultations?.length ? (
                    <div className="px-3 pb-2">
                        <DelegationTrace consultations={msg.consultations} assistantName={assistantName} />
                    </div>
                ) : null}
                {!isUser && showExecutionSummary && msg.execution_summary && (
                    useReceipt ? (
                        <ExecutionSummaryReceipt
                            summary={msg.execution_summary}
                            runId={msg.run_id}
                            artifacts={msg.artifacts}
                        />
                    ) : (
                        <ExecutionSummaryCard
                            summary={msg.execution_summary}
                            runId={msg.run_id}
                            artifacts={msg.artifacts}
                            compact={compactResult}
                        />
                    )
                )}
                {!isUser && msg.proposal && <ProposedActionBlock message={msg} />}
                {showStructuredEvidence && artifactSummary && (
                    <div className="rounded-lg border border-cortex-primary/20 bg-cortex-primary/5 px-3 py-2">
                        <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-primary">
                            {compactResult
                                ? responseBadge?.label ?? "Returned output"
                                : "Returned output"}
                        </div>
                        <p className="mt-1 text-sm text-cortex-text-main leading-6">{artifactSummary}</p>
                    </div>
                )}
                {showStructuredEvidence && consultationSummary && (
                    <div className="rounded-lg border border-cortex-warning/20 bg-cortex-warning/5 px-3 py-2">
                        <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-warning">
                            {compactResult
                                ? "Specialist support"
                                : "Specialist context"}
                        </div>
                        <p className="mt-1 text-sm text-cortex-text-main leading-6">{consultationSummary}</p>
                    </div>
                )}
                {showStructuredEvidence && msg.artifacts?.length ? (
                    <div className="space-y-1">
                        {msg.artifacts.map((artifact, i) => <InlineArtifact key={artifact.id || `art-${i}`} artifact={artifact} />)}
                    </div>
                ) : null}
                {showTraceExtras && !msg.proposal && <MissionControlToolsUsed tools={msg.tools_used} />}
                {showTraceExtras && msg.tools_used && (msg.tools_used.includes("recall") || msg.tools_used.includes("search_memory")) && (
                    <div className="flex items-center gap-1 px-1 mt-0.5">
                        <span className="w-1 h-1 rounded-full bg-cortex-primary" />
                        <span className="text-[8px] font-mono text-cortex-primary/70 italic">recalled from memory</span>
                    </div>
                )}
            </div>
            {isUser && (
                <div className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 mt-0.5 ${
                    isBroadcast ? "bg-cortex-warning/10 border border-cortex-warning/30" : "bg-cortex-bg border border-cortex-border"
                }`}>
                    {isBroadcast ? <Megaphone className="w-3.5 h-3.5 text-cortex-warning" /> : <User className="w-3.5 h-3.5 text-cortex-text-muted" />}
                </div>
            )}
        </div>
    );
}
