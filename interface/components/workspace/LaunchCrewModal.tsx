"use client";

import React, { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { X, Users, CheckCircle, ArrowRight, MessageSquareText } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { LaunchCrewDescribeStep, type LaunchCrewState, lastOutcomeMessage, stepForState, StepIndicator } from "./LaunchCrewShared";
import { LaunchCrewOutcomeBody } from "./LaunchCrewOutcomes";

interface Props {
    onClose: () => void;
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
                        <LaunchCrewDescribeStep
                            textareaRef={textareaRef}
                            intent={intent}
                            onChange={setIntent}
                            onKeyDown={handleKeyDown}
                        />
                    )}

                    <LaunchCrewOutcomeBody
                        state={state}
                        assistantName={assistantName}
                        intent={intent}
                        pendingProposal={pendingProposal}
                        outcomeMessage={outcomeMessage}
                        outcomeArtifacts={outcomeArtifacts}
                        activeRunId={activeRunId}
                        blockerMessage={blockerMessage}
                    />
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
