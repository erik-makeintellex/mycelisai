"use client";

import { FileText, ListChecks, PlayCircle, type LucideIcon } from "lucide-react";
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

type ContinuationAction = {
    label: string;
    detail: string;
    prompt: string;
    Icon: LucideIcon;
    primary?: boolean;
};

function continuationActions(teamName: string): ContinuationAction[] {
    return [
        {
            label: "Build playable prototype",
            detail: "Project package, README, validation notes",
            prompt: `Have ${teamName} build the first playable browser-game prototype as a reviewable project package. Save it in the team's group folder with README, validation notes, and output link.`,
            Icon: PlayCircle,
            primary: true,
        },
        {
            label: "Write design brief",
            detail: "Mechanics, roles, output shape, acceptance criteria",
            prompt: `Have ${teamName} write a concise game design brief with mechanics, art direction, team roles, acceptance criteria, and output path needed before build.`,
            Icon: FileText,
        },
        {
            label: "Draft delivery plan",
            detail: "Next tasks, tools, risks, expected review",
            prompt: `Have ${teamName} draft the next deliverable plan with output shape, owner roles, needed tools, risks, and expected review.`,
            Icon: ListChecks,
        },
    ];
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
    const actions = continuationActions(teamNeedingWork);

    return (
        <div className="mb-2 rounded-md border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2">
            <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[11px] font-semibold text-cortex-text-main">
                        {teamNeedingWork} is ready. Choose the first deliverable.
                    </span>
                    <span className="rounded-full border border-cortex-warning/25 bg-cortex-bg px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.12em] text-cortex-warning">
                        Needs first task
                    </span>
                </div>
                <div className="text-[10px] leading-4 text-cortex-text-muted">
                    Pick a starter, review it in Soma, then send. Soma will create a work item before anything runs.
                </div>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
                {actions.map(({ label, detail, prompt, Icon, primary }) => (
                    <button
                        key={label}
                        type="button"
                        onClick={() => onStarterPrompt(prompt)}
                        title={`${label}: ${detail}`}
                        className={
                            primary
                                ? "inline-flex items-center gap-1.5 rounded-md border border-cortex-info/30 bg-cortex-info/10 px-2.5 py-1.5 text-[10px] font-mono font-bold uppercase text-cortex-info transition-colors hover:bg-cortex-info/15"
                                : "inline-flex items-center gap-1.5 rounded-md border border-cortex-border bg-cortex-bg px-2.5 py-1.5 text-[10px] font-mono font-bold uppercase text-cortex-text-muted transition-colors hover:border-cortex-primary/30 hover:text-cortex-text-main"
                        }
                    >
                        <Icon className="h-3.5 w-3.5" />
                        {label}
                    </button>
                ))}
            </div>
        </div>
    );
}
