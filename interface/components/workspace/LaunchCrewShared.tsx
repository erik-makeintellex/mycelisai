"use client";

import React from "react";
import { CheckCircle, ExternalLink, Zap } from "lucide-react";
import { type ChatArtifactRef, type ChatMessage, type ExecutionMode } from "@/store/useCortexStore";

export const STEPS = [
    { id: 1, label: "Describe" },
    { id: 2, label: "Evaluate" },
    { id: 3, label: "Outcome" },
] as const;

export const EXAMPLES = [
    "Analyze our data and generate a weekly summary report",
    "Research a topic and synthesize the findings",
    "Review and refactor a section of our codebase",
];

export type LaunchCrewState =
    | "describe"
    | "waiting"
    | "proposal"
    | "confirmed_pending_execution"
    | "answer"
    | "execution_result"
    | "blocker";

export function stepForState(state: LaunchCrewState): 1 | 2 | 3 {
    if (state === "describe") return 1;
    if (state === "waiting") return 2;
    return 3;
}

export function lastOutcomeMessage(messages: ChatMessage[], startIndex: number | null): ChatMessage | null {
    if (startIndex == null) return null;
    const tail = messages.slice(startIndex).filter((msg) => msg.role !== "user");
    if (tail.length === 0) return null;
    return tail[tail.length - 1];
}

export function StepIndicator({ current }: { current: 1 | 2 | 3 }) {
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

export function OutcomeBadge({ mode }: { mode: ExecutionMode | "confirmed_pending_execution" | "execution_result" | "blocker" }) {
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

export function OutcomeArtifacts({ artifacts }: { artifacts: ChatArtifactRef[] }) {
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

export function LaunchCrewDescribeStep({
    textareaRef,
    intent,
    onChange,
    onKeyDown,
}: {
    textareaRef: React.RefObject<HTMLTextAreaElement | null>;
    intent: string;
    onChange: (value: string) => void;
    onKeyDown: (e: React.KeyboardEvent) => void;
}) {
    return (
        <div className="space-y-4">
            <div>
                <label className="mb-2 block text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted">
                    What should this crew accomplish?
                </label>
                <textarea
                    ref={textareaRef}
                    value={intent}
                    onChange={(e) => onChange(e.target.value)}
                    onKeyDown={onKeyDown}
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
                            onClick={() => onChange(example)}
                            className="w-full rounded-lg border border-cortex-border/60 bg-cortex-bg/50 px-2.5 py-2 text-left text-[10px] font-mono text-cortex-text-muted transition-all hover:border-cortex-primary/40 hover:bg-cortex-primary/5 hover:text-cortex-text-main"
                        >
                            <Zap className="mr-1.5 inline-block h-2.5 w-2.5 text-cortex-primary/60" />
                            {example}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}
