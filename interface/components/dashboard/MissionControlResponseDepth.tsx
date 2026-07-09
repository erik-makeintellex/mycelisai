"use client";

import type { ChatMessage, ResponseDepth } from "@/store/useCortexStore";
import MissionControlMarkdown from "./MissionControlMarkdown";

const RESPONSE_DEPTH_LABELS: Record<ResponseDepth, string> = {
    quick_box: "Quick answer",
    structured_summary: "Summary",
    decision_brief: "Decision brief",
    execution_proposal: "Proposed work",
};

function responseDepthTone(depth?: ResponseDepth) {
    switch (depth) {
        case "quick_box":
            return "border-cortex-primary/15 bg-cortex-bg/70";
        case "decision_brief":
            return "border-cortex-warning/25 bg-cortex-warning/5";
        case "execution_proposal":
            return "border-cortex-warning/20 bg-cortex-warning/5";
        case "structured_summary":
        default:
            return "border-cortex-info/20 bg-cortex-info/5";
    }
}

function shouldShowDepthLabel(msg: ChatMessage) {
    return Boolean(msg.response_depth && msg.response_depth !== "execution_proposal");
}

export default function MissionControlResponseDepth({
    msg,
    isBroadcast,
    isUser,
    compactResult,
}: {
    msg: ChatMessage;
    isBroadcast: boolean;
    isUser: boolean;
    compactResult: boolean;
}) {
    const label = msg.response_depth ? RESPONSE_DEPTH_LABELS[msg.response_depth] : null;
    const labelVisible = !isUser && shouldShowDepthLabel(msg);
    const boundedClass = compactResult && !isUser
        ? "max-h-[420px] overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border"
        : "";
    const toneClass = isBroadcast
        ? "border-cortex-warning/30 bg-cortex-warning/10"
        : isUser
            ? "border-cortex-border bg-cortex-bg"
            : responseDepthTone(msg.response_depth);

    return (
        <div className={`rounded-lg border px-3 py-2 text-sm leading-relaxed text-cortex-text-main ${toneClass} ${boundedClass}`}>
            {labelVisible ? (
                <div className="mb-1.5 text-[9px] font-mono font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
                    {label}
                </div>
            ) : null}
            {isUser ? msg.content : <MissionControlMarkdown content={msg.content} />}
        </div>
    );
}
