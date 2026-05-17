"use client";

import { PlayCircle } from "lucide-react";
import type { ChatMessage, ExecutionSummaryData, ExecutionSummaryItem } from "@/store/useCortexStore";

function valueText(value: unknown): string {
    if (typeof value === "string") return value;
    if (!value || typeof value !== "object") return "";
    const item = value as ExecutionSummaryItem;
    return item.title ?? item.label ?? item.name ?? item.id ?? item.value ?? "";
}

function summaryItems(value: ExecutionSummaryData["outputs"]): Array<string | ExecutionSummaryItem> {
    if (!value) return [];
    return Array.isArray(value) ? value : [value];
}

function summaryCapabilityNames(value: ExecutionSummaryData["capability_use"]): string[] {
    if (!value) return [];
    if (Array.isArray(value)) return value.map(valueText).filter(Boolean);
    return [
        ...(value.capabilities ?? []),
        ...(value.tools ?? []),
        ...(value.used ?? []),
    ].map(valueText).filter(Boolean);
}

function teamOnlyCreation(summary?: ExecutionSummaryData) {
    if (!summary) return null;
    const capabilities = summaryCapabilityNames(summary.capability_use).map((value) => value.toLowerCase());
    if (!capabilities.some((value) => value.includes("create_team"))) return null;
    if (capabilities.some((value) => value.includes("write_file"))) return null;

    const outputs = summaryItems(summary.outputs);
    const hasDeliverable = outputs.some((output) => {
        if (typeof output === "string") return false;
        const kind = (output.kind ?? output.type ?? "").toLowerCase();
        return ["code", "file", "project_package", "artifact", "document"].includes(kind);
    });
    if (hasDeliverable) return null;

    const teamOutput = outputs.find((output) => typeof output !== "string" && ((output.kind ?? "").toLowerCase() === "team"));
    return valueText(teamOutput) || "the new team";
}

function latestTeamOnlyCreation(messages: ChatMessage[]) {
    for (let i = messages.length - 1; i >= 0; i -= 1) {
        const teamName = teamOnlyCreation(messages[i].execution_summary);
        if (teamName) return teamName;
    }
    return null;
}

export default function MissionControlTeamContinuationPrompt({
    messages,
    disabled,
    onStarterPrompt,
}: {
    messages: ChatMessage[];
    disabled?: boolean;
    onStarterPrompt: (prompt: string) => void;
}) {
    const teamNeedingWork = latestTeamOnlyCreation(messages);
    if (!teamNeedingWork || disabled) return null;

    return (
        <div className="mb-2 flex flex-wrap items-center justify-between gap-2 rounded-md border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2">
            <div className="min-w-0">
                <div className="text-[11px] font-semibold text-cortex-text-main">
                    {teamNeedingWork} is ready. No work item has started yet.
                </div>
                <div className="text-[10px] leading-4 text-cortex-text-muted">
                    Start a concrete deliverable so outputs and proof appear here.
                </div>
            </div>
            <button
                type="button"
                onClick={() => onStarterPrompt(`Have ${teamNeedingWork} build the first playable browser game prototype and save it as a reviewable output.`)}
                className="inline-flex items-center gap-1.5 rounded-md border border-cortex-info/30 bg-cortex-info/10 px-2.5 py-1.5 text-[10px] font-mono font-bold uppercase text-cortex-info transition-colors hover:bg-cortex-info/15"
            >
                <PlayCircle className="h-3.5 w-3.5" />
                Start work
            </button>
        </div>
    );
}
