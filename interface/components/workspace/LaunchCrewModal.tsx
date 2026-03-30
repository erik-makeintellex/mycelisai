"use client";

import React, { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
    X,
    Users,
    Loader2,
    CheckCircle,
    ArrowRight,
    Zap,
    AlertTriangle,
    MessageSquareText,
    Sparkles,
    ExternalLink,
} from "lucide-react";
import { useCortexStore, type ChatArtifactRef, type ChatMessage, type ExecutionMode } from "@/store/useCortexStore";

const STEPS = [
    { id: 1, label: "Describe" },
    { id: 2, label: "Evaluate" },
    { id: 3, label: "Outcome" },
] as const;

const EXAMPLES = [
    "Analyze our data and generate a weekly summary report",
    "Research a topic and synthesize the findings",
    "Review and refactor a section of our codebase",
];

type LaunchCrewState =
    | "describe"
    | "waiting"
    | "proposal"
    | "confirmed_pending_execution"
    | "answer"
    | "execution_result"
    | "blocker";

interface Props {
    onClose: () => void;
}

function stepForState(state: LaunchCrewState): 1 | 2 | 3 {
    if (state === "describe") return 1;
    if (state === "waiting") return 2;
    return 3;
}

function StepIndicator({ current }: { current: 1 | 2 | 3 }) {
    return (
        <div className="mb-5 flex items-center gap-0">
            {STEPS.map((step, i) => {
                const done = step.id < current;
                const active = step.id === current;
                return (
                    <React.Fragment key={step.id}>
                        <div className="flex flex-col items-center">
                            <div
                                className={`flex h-6 w-6 items-center justify-center rounded-full text-[10px] font-mono font-bold transition-all ${
                                    done
                                        ? "bg-cortex-success text-white"
                                        : active
                                          ? "bg-cortex-primary text-white shadow-[0_0_8px_rgba(75,78,109,0.32)]"
                                          : "bg-cortex-border text-cortex-text-muted"
                                }`}
                            >
                                {done ? <CheckCircle className="h-3.5 w-3.5" /> : step.id}
                            </div>
                            <span
                                className={`mt-1 text-[9px] font-mono uppercase tracking-wider ${
                                    active ? "text-cortex-primary" : "text-cortex-text-muted/60"
                                }`}
                            >
                                {step.label}
                            </span>
                        </div>
                        {i < STEPS.length - 1 && (
                            <div
                                className={`mx-2 mb-4 h-px flex-1 transition-all ${
                                    done ? "bg-cortex-success" : "bg-cortex-border"
                                }`}
                            />
                        )}
                    </React.Fragment>
                );
            })}
        </div>
    );
}

function OutcomeBadge({ mode }: { mode: ExecutionMode | "confirmed_pending_execution" | "execution_result" | "blocker" }) {
    const tone =
        mode === "proposal"
            ? "border-amber-400/30 bg-amber-400/10 text-amber-400"
            : mode === "confirmed_pending_execution"
              ? "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary"
            : mode === "execution_result"
              ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
              : mode === "blocker"
                ? "border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger"
                : "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary";
    const label =
        mode === "proposal"
            ? "Proposal"
            : mode === "confirmed_pending_execution"
              ? "Awaiting Proof"
            : mode === "execution_result"
              ? "Execution Result"
              : mode === "blocker"
                ? "Blocker"
                : "Answer";

    return (
        <span className={`rounded-full border px-2 py-1 text-[9px] font-mono uppercase tracking-wider ${tone}`}>
            {label}
        </span>
    );
}

function artifactReference(artifact: ChatArtifactRef): string {
    if (artifact.saved_path?.trim()) return artifact.saved_path.trim();
    if (artifact.url?.trim()) return artifact.url.trim();
    if (artifact.id?.trim()) return `Artifact ID: ${artifact.id.trim()}`;
    if (artifact.content_type?.trim()) return artifact.content_type.trim();
    return `${artifact.type} output`;
}

function OutcomeArtifacts({ artifacts }: { artifacts: ChatArtifactRef[] }) {
    if (artifacts.length === 0) return null;

    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg px-3 py-3">
            <p className="mb-2 text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">
                Delivered outputs
            </p>
            <div className="space-y-2">
                {artifacts.map((artifact, index) => (
                    <div
                        key={artifact.id || artifact.url || `${artifact.title}-${index}`}
                        className="rounded-lg border border-cortex-border/60 bg-cortex-surface/40 px-2.5 py-2"
                    >
                        <div className="flex items-center gap-2">
                            <span className="flex-1 text-[10px] font-mono font-bold text-cortex-text-main">
                                {artifact.title}
                            </span>
                            <span className="rounded border border-cortex-primary/20 bg-cortex-primary/10 px-1.5 py-0.5 text-[8px] font-mono uppercase text-cortex-primary">
                                {artifact.type}
                            </span>
                            {artifact.url ? (
                                <a
                                    href={artifact.url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-cortex-primary transition-colors hover:text-cortex-primary/80"
                                    aria-label={`Open ${artifact.title}`}
                                >
                                    <ExternalLink className="h-3 w-3" />
                                </a>
                            ) : null}
                        </div>
                        <p className="mt-1 text-[10px] font-mono text-cortex-text-muted break-all">
                            {artifactReference(artifact)}
                        </p>
                    </div>
                ))}
            </div>
        </div>
    );
}

function lastOutcomeMessage(messages: ChatMessage[], startIndex: number | null): ChatMessage | null {
    if (startIndex == null) return null;
    const tail = messages.slice(startIndex).filter((msg) => msg.role !== "user");
    if (tail.length === 0) return null;
    return tail[tail.length - 1];
}

export default function LaunchCrewModal({ onClose }: Props) {
    const router = useRouter();
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const pendingProposal = useCortexStore((s) => s.pendingProposal);
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);
    const assistantName = useCortexStore((s) => s.assistantName);
    const missionChat = useCortexStore((s) => s.missionChat);
    const missionChatError = useCortexStore((s) => s.missionChatError);
    const activeMode = useCortexStore((s) => s.activeMode);
    const activeRunId = useCortexStore((s) => s.activeRunId);

    const [intent, setIntent] = useState("");
    const [state, setState] = useState<LaunchCrewState>("describe");
    const [requestStartIndex, setRequestStartIndex] = useState<number | null>(null);
    const [confirmError, setConfirmError] = useState("");
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    const currentStep = stepForState(state);
    const outcomeMessage = useMemo(
        () => lastOutcomeMessage(missionChat, requestStartIndex),
        [missionChat, requestStartIndex],
    );
    const outcomeArtifacts = outcomeMessage?.artifacts ?? [];
    const blockerMessage = confirmError || missionChatError || outcomeMessage?.content || "Launch Crew hit a blocker.";

    useEffect(() => {
        setCouncilTarget("admin");
        setState("describe");
        setIntent("");
        setConfirmError("");
        setRequestStartIndex(null);
        textareaRef.current?.focus();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    useEffect(() => {
        if (state !== "waiting" || isMissionChatting) return;

        if (pendingProposal || activeMode === "proposal") {
            setState(pendingProposal ? "proposal" : "confirmed_pending_execution");
            return;
        }

        if (missionChatError || activeMode === "blocker") {
            setState("blocker");
            return;
        }

        if (activeMode === "execution_result" || (outcomeMessage?.role === "system" && outcomeMessage.run_id)) {
            setState("execution_result");
            return;
        }

        if (outcomeMessage) {
            setState(outcomeMessage.mode === "blocker" ? "blocker" : "answer");
        }
    }, [activeMode, isMissionChatting, missionChatError, outcomeMessage, pendingProposal, state]);

    const handleSend = () => {
        const trimmed = intent.trim();
        if (!trimmed || isMissionChatting) return;
        setCouncilTarget("admin");
        setConfirmError("");
        setRequestStartIndex(missionChat.length);
        sendMissionChat(trimmed);
        setState("waiting");
    };

    const handleConfirm = async () => {
        setConfirmError("");
        const result = await confirmProposal();
        if (result.ok) {
            setState(result.runId ? "execution_result" : "confirmed_pending_execution");
            return;
        }
        setConfirmError(result.error || "Launch Crew failed.");
        setState("blocker");
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
            handleSend();
        }
    };

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
            onClick={(e) => {
                if (e.target === e.currentTarget) onClose();
            }}
        >
            <div className="mx-4 w-full max-w-md overflow-hidden rounded-2xl border border-cortex-border bg-cortex-surface shadow-2xl">
                <div className="flex items-center justify-between border-b border-cortex-border px-5 pb-4 pt-5">
                    <div className="flex items-center gap-2.5">
                        <div className="flex h-7 w-7 items-center justify-center rounded-lg border border-cortex-primary/30 bg-cortex-primary/15">
                            <Users className="h-4 w-4 text-cortex-primary" />
                        </div>
                        <div>
                            <h2 className="font-mono text-sm font-bold text-cortex-text-main">Launch a Crew</h2>
                            <p className="text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                                {assistantName} must return a real outcome, not a planning stub
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="rounded-lg p-1.5 text-cortex-text-muted transition-colors hover:bg-cortex-border hover:text-cortex-text-main"
                    >
                        <X className="h-4 w-4" />
                    </button>
                </div>

                <div className="px-5 py-5">
                    <StepIndicator current={currentStep} />

                    {state === "describe" && (
                        <div className="space-y-4">
                            <div>
                                <label className="mb-2 block text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">
                                    What should this crew accomplish?
                                </label>
                                <textarea
                                    ref={textareaRef}
                                    value={intent}
                                    onChange={(e) => setIntent(e.target.value)}
                                    onKeyDown={handleKeyDown}
                                    placeholder="Describe the outcome you need..."
                                    rows={4}
                                    className="w-full resize-none rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2.5 text-sm font-mono leading-relaxed text-cortex-text-main placeholder-cortex-text-muted/40 focus:border-cortex-primary focus:outline-none focus:ring-1 focus:ring-cortex-primary/30"
                                />
                                <p className="mt-1 text-right text-[9px] font-mono text-cortex-text-muted/50">⌘ Enter to send</p>
                            </div>

                            <div>
                                <p className="mb-2 text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Examples</p>
                                <div className="space-y-1.5">
                                    {EXAMPLES.map((example) => (
                                        <button
                                            key={example}
                                            onClick={() => setIntent(example)}
                                            className="w-full rounded-lg border border-cortex-border/60 bg-cortex-bg/50 px-2.5 py-2 text-left text-[10px] font-mono text-cortex-text-muted transition-all hover:border-cortex-primary/40 hover:bg-cortex-primary/5 hover:text-cortex-text-main"
                                        >
                                            <Zap className="mr-1.5 inline-block h-2.5 w-2.5 text-cortex-primary/60" />
                                            {example}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

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
                </div>

                <div className="flex justify-end gap-2 px-5 pb-5">
                    {state === "describe" && (
                        <>
                            <button
                                onClick={onClose}
                                className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:border-cortex-border/80 hover:text-cortex-text-main"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSend}
                                disabled={!intent.trim() || isMissionChatting}
                                className="flex items-center gap-2 rounded-lg bg-cortex-primary px-4 py-2 text-sm font-mono font-bold text-white transition-all hover:bg-cortex-primary/80 disabled:cursor-not-allowed disabled:opacity-40"
                            >
                                Send to {assistantName}
                                <ArrowRight className="h-3.5 w-3.5" />
                            </button>
                        </>
                    )}

                    {state === "waiting" && (
                        <button
                            onClick={onClose}
                            className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                        >
                            View in Workspace chat
                        </button>
                    )}

                    {state === "proposal" && (
                        <>
                            <button
                                onClick={() => {
                                    cancelProposal();
                                    setState("describe");
                                }}
                                className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                            >
                                Back to intent
                            </button>
                            <button
                                onClick={onClose}
                                className="rounded-lg border border-cortex-primary/40 px-4 py-2 text-sm font-mono text-cortex-primary transition-colors hover:bg-cortex-primary/10"
                            >
                                Review in chat
                            </button>
                            <button
                                onClick={handleConfirm}
                                className="flex items-center gap-2 rounded-lg bg-cortex-success px-4 py-2 text-sm font-mono font-bold text-white transition-all hover:bg-cortex-success/80"
                            >
                                <CheckCircle className="h-3.5 w-3.5" />
                                Launch Crew
                            </button>
                        </>
                    )}

                    {state === "answer" && (
                        <>
                            <button
                                onClick={() => setState("describe")}
                                className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                            >
                                Try another brief
                            </button>
                            <button
                                onClick={onClose}
                                className="rounded-lg border border-cortex-primary/40 px-4 py-2 text-sm font-mono text-cortex-primary transition-colors hover:bg-cortex-primary/10"
                            >
                                Review in chat
                            </button>
                        </>
                    )}

                    {state === "execution_result" && (
                        <>
                            <button
                                onClick={onClose}
                                className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                            >
                                Close
                            </button>
                            {activeRunId && (
                                <button
                                    onClick={() => {
                                        router.push(`/runs/${activeRunId}`);
                                        onClose();
                                    }}
                                    className="flex items-center gap-2 rounded-lg bg-cortex-success px-4 py-2 text-sm font-mono font-bold text-white transition-all hover:bg-cortex-success/80"
                                >
                                    <MessageSquareText className="h-3.5 w-3.5" />
                                    View Run
                                </button>
                            )}
                        </>
                    )}

                    {state === "confirmed_pending_execution" && (
                        <button
                            onClick={onClose}
                            className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                        >
                            Review in chat
                        </button>
                    )}

                    {state === "blocker" && (
                        <>
                            <button
                                onClick={() => setState("describe")}
                                className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-mono text-cortex-text-muted transition-colors hover:text-cortex-text-main"
                            >
                                Revise request
                            </button>
                            <button
                                onClick={onClose}
                                className="rounded-lg border border-cortex-danger/40 px-4 py-2 text-sm font-mono text-cortex-danger transition-colors hover:bg-cortex-danger/10"
                            >
                                Continue in chat
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
