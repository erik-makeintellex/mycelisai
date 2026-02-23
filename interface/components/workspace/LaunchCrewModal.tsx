"use client";

import React, { useEffect, useRef, useState } from "react";
import { X, Users, Loader2, CheckCircle, ArrowRight, Zap } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

// ── Step Indicator ────────────────────────────────────────────

const STEPS = [
    { id: 1, label: "Describe" },
    { id: 2, label: "Proposal" },
    { id: 3, label: "Confirm" },
];

function StepIndicator({ current }: { current: 1 | 2 | 3 }) {
    return (
        <div className="flex items-center gap-0 mb-5">
            {STEPS.map((step, i) => {
                const done = step.id < current;
                const active = step.id === current;
                return (
                    <React.Fragment key={step.id}>
                        <div className="flex flex-col items-center">
                            <div
                                className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-mono font-bold transition-all ${
                                    done
                                        ? "bg-cortex-success text-white"
                                        : active
                                        ? "bg-cortex-primary text-white shadow-[0_0_8px_rgba(6,182,212,0.4)]"
                                        : "bg-cortex-border text-cortex-text-muted"
                                }`}
                            >
                                {done ? <CheckCircle className="w-3.5 h-3.5" /> : step.id}
                            </div>
                            <span
                                className={`text-[9px] font-mono uppercase tracking-wider mt-1 ${
                                    active ? "text-cortex-primary" : "text-cortex-text-muted/60"
                                }`}
                            >
                                {step.label}
                            </span>
                        </div>
                        {i < STEPS.length - 1 && (
                            <div
                                className={`flex-1 h-px mx-2 mb-4 transition-all ${
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

// ── Example Prompts ───────────────────────────────────────────

const EXAMPLES = [
    "Analyze our data and generate a weekly summary report",
    "Research a topic and synthesize the findings",
    "Review and refactor a section of our codebase",
];

// ── LaunchCrewModal ───────────────────────────────────────────

interface Props {
    onClose: () => void;
}

export default function LaunchCrewModal({ onClose }: Props) {
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const pendingProposal = useCortexStore((s) => s.pendingProposal);
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);

    const [step, setStep] = useState<1 | 2 | 3>(1);
    const [intent, setIntent] = useState("");
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    // Clear stale proposal and reset to Soma on open
    useEffect(() => {
        cancelProposal();
        setCouncilTarget('admin');
        setStep(1);
        setIntent('');
        textareaRef.current?.focus();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Advance to step 3 when Soma has a proposal ready
    useEffect(() => {
        if (step === 2 && pendingProposal) {
            setStep(3);
        }
    }, [pendingProposal, step]);

    const handleSend = () => {
        const trimmed = intent.trim();
        if (!trimmed || isMissionChatting) return;
        setCouncilTarget('admin'); // always route to Soma
        sendMissionChat(trimmed);
        setStep(2);
    };

    const handleConfirm = () => {
        confirmProposal();
        onClose();
    };

    const handleCancel = () => {
        cancelProposal();
        onClose();
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
            handleSend();
        }
    };

    return (
        // Backdrop
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
            onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
        >
            <div className="w-full max-w-md mx-4 bg-cortex-surface border border-cortex-border rounded-2xl shadow-2xl overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-5 pt-5 pb-4 border-b border-cortex-border">
                    <div className="flex items-center gap-2.5">
                        <div className="w-7 h-7 rounded-lg bg-cortex-primary/15 border border-cortex-primary/30 flex items-center justify-center">
                            <Users className="w-4 h-4 text-cortex-primary" />
                        </div>
                        <div>
                            <h2 className="text-sm font-bold text-cortex-text-main font-mono">
                                Launch a Crew
                            </h2>
                            <p className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider">
                                Soma will design a team for your intent
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1.5 rounded-lg hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                {/* Body */}
                <div className="px-5 py-5">
                    <StepIndicator current={step} />

                    {/* Step 1: Intent input */}
                    {step === 1 && (
                        <div className="space-y-4">
                            <div>
                                <label className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted block mb-2">
                                    What should this crew accomplish?
                                </label>
                                <textarea
                                    ref={textareaRef}
                                    value={intent}
                                    onChange={(e) => setIntent(e.target.value)}
                                    onKeyDown={handleKeyDown}
                                    placeholder="Describe the outcome you need..."
                                    rows={4}
                                    className="w-full bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2.5 text-sm font-mono text-cortex-text-main placeholder-cortex-text-muted/40 focus:outline-none focus:border-cortex-primary focus:ring-1 focus:ring-cortex-primary/30 resize-none leading-relaxed"
                                />
                                <p className="text-[9px] font-mono text-cortex-text-muted/50 mt-1 text-right">
                                    ⌘ Enter to send
                                </p>
                            </div>

                            <div>
                                <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-2">
                                    Examples
                                </p>
                                <div className="space-y-1.5">
                                    {EXAMPLES.map((ex) => (
                                        <button
                                            key={ex}
                                            onClick={() => setIntent(ex)}
                                            className="w-full text-left px-2.5 py-2 rounded-lg border border-cortex-border/60 bg-cortex-bg/50 hover:border-cortex-primary/40 hover:bg-cortex-primary/5 text-[10px] font-mono text-cortex-text-muted hover:text-cortex-text-main transition-all"
                                        >
                                            <Zap className="w-2.5 h-2.5 inline-block mr-1.5 text-cortex-primary/60" />
                                            {ex}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Step 2: Waiting for proposal */}
                    {step === 2 && (
                        <div className="space-y-4">
                            <div className="flex flex-col items-center justify-center py-6 gap-3">
                                <div className="w-10 h-10 rounded-full bg-cortex-primary/10 border border-cortex-primary/20 flex items-center justify-center">
                                    <Loader2 className="w-5 h-5 text-cortex-primary animate-spin" />
                                </div>
                                <div className="text-center">
                                    <p className="text-sm font-mono font-bold text-cortex-text-main">
                                        Soma is designing your crew
                                    </p>
                                    <p className="text-[10px] font-mono text-cortex-text-muted mt-1">
                                        Analysing intent, selecting agents, and wiring the blueprint...
                                    </p>
                                </div>
                            </div>

                            <div className="bg-cortex-bg/60 border border-cortex-border rounded-lg px-3 py-2.5">
                                <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1">
                                    Your intent
                                </p>
                                <p className="text-[10px] font-mono text-cortex-text-main leading-relaxed">
                                    {intent}
                                </p>
                            </div>

                            <p className="text-[9px] font-mono text-cortex-text-muted/60 text-center">
                                The proposal will also appear in your Workspace chat
                            </p>
                        </div>
                    )}

                    {/* Step 3: Proposal ready */}
                    {step === 3 && pendingProposal && (
                        <div className="space-y-4">
                            <div className="flex items-center gap-2 mb-1">
                                <CheckCircle className="w-4 h-4 text-cortex-success" />
                                <p className="text-sm font-mono font-bold text-cortex-text-main">
                                    Soma has a proposal
                                </p>
                            </div>

                            <div className="bg-cortex-bg border border-cortex-border rounded-lg divide-y divide-cortex-border/50">
                                <div className="px-3 py-2 flex justify-between">
                                    <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Intent</span>
                                    <span className="text-[10px] font-mono text-cortex-text-main truncate max-w-[200px]">
                                        {pendingProposal.intent}
                                    </span>
                                </div>
                                <div className="px-3 py-2 flex justify-between">
                                    <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Risk</span>
                                    <span className={`text-[10px] font-mono font-bold uppercase ${
                                        pendingProposal.risk_level === "high"
                                            ? "text-red-400"
                                            : pendingProposal.risk_level === "medium"
                                            ? "text-amber-400"
                                            : "text-cortex-success"
                                    }`}>
                                        {pendingProposal.risk_level}
                                    </span>
                                </div>
                                <div className="px-3 py-2 flex justify-between">
                                    <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">Tools</span>
                                    <div className="flex flex-wrap gap-1 justify-end max-w-[200px]">
                                        {pendingProposal.tools.slice(0, 4).map((t) => (
                                            <span key={t} className="text-[8px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20">
                                                {t}
                                            </span>
                                        ))}
                                        {pendingProposal.tools.length > 4 && (
                                            <span className="text-[8px] font-mono text-cortex-text-muted">
                                                +{pendingProposal.tools.length - 4}
                                            </span>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer actions */}
                <div className="px-5 pb-5 flex gap-2 justify-end">
                    {step === 1 && (
                        <>
                            <button
                                onClick={onClose}
                                className="px-4 py-2 rounded-lg border border-cortex-border text-sm font-mono text-cortex-text-muted hover:text-cortex-text-main hover:border-cortex-border/80 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSend}
                                disabled={!intent.trim() || isMissionChatting}
                                className="px-4 py-2 rounded-lg bg-cortex-primary text-white text-sm font-mono font-bold flex items-center gap-2 hover:bg-cortex-primary/80 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
                            >
                                Send to Soma
                                <ArrowRight className="w-3.5 h-3.5" />
                            </button>
                        </>
                    )}

                    {step === 2 && (
                        <button
                            onClick={onClose}
                            className="px-4 py-2 rounded-lg border border-cortex-border text-sm font-mono text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        >
                            View in Workspace chat
                        </button>
                    )}

                    {step === 3 && (
                        <>
                            <button
                                onClick={handleCancel}
                                className="px-4 py-2 rounded-lg border border-cortex-border text-sm font-mono text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                            >
                                Discard
                            </button>
                            <button
                                onClick={onClose}
                                className="px-4 py-2 rounded-lg border border-cortex-primary/40 text-cortex-primary text-sm font-mono hover:bg-cortex-primary/10 transition-colors"
                            >
                                Review in chat
                            </button>
                            <button
                                onClick={handleConfirm}
                                className="px-4 py-2 rounded-lg bg-cortex-success text-white text-sm font-mono font-bold flex items-center gap-2 hover:bg-cortex-success/80 transition-all"
                            >
                                <CheckCircle className="w-3.5 h-3.5" />
                                Launch Crew
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
