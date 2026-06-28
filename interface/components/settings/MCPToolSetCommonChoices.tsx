"use client";

import type { MCPToolSetScopeKind } from "@/store/useCortexStore";

type CapabilityChoice = {
    label: string;
    description: string;
    refs: string[];
    recommendedScope: MCPToolSetScopeKind;
};

const capabilityChoices: CapabilityChoice[] = [
    {
        label: "Workspace files",
        description: "Read and write retained files inside the workspace boundary.",
        refs: ["mcp:filesystem/*"],
        recommendedScope: "all",
    },
    {
        label: "Web research",
        description: "Fetch public sources when search/research is approved.",
        refs: ["mcp:fetch/fetch", "tool:web_search"],
        recommendedScope: "group",
    },
    {
        label: "Team coordination",
        description: "Let Soma coordinate a bounded team or Outcome lane.",
        refs: ["capability:team_orchestration"],
        recommendedScope: "group",
    },
    {
        label: "Local host/media",
        description: "Use host-local services such as media or deployment helpers.",
        refs: ["host:invoke", "mcp:filesystem/*"],
        recommendedScope: "host",
    },
];

type Props = {
    onChoose: (refs: string[], recommendedScope: MCPToolSetScopeKind) => void;
};

export function MCPToolSetCommonChoices({ onChoose }: Props) {
    return (
        <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-surface/60 p-2">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                Common choices
            </p>
            <div className="mt-2 grid gap-1.5">
                {capabilityChoices.map((choice) => (
                    <button
                        key={choice.label}
                        type="button"
                        onClick={() => onChoose(choice.refs, choice.recommendedScope)}
                        className="rounded-md border border-cortex-border bg-cortex-bg px-2.5 py-2 text-left transition hover:border-cortex-primary/40"
                    >
                        <span className="flex items-center justify-between gap-2">
                            <span className="text-xs font-semibold text-cortex-text-main">{choice.label}</span>
                            <span className="rounded border border-cortex-border bg-cortex-surface px-1.5 py-0.5 text-[9px] font-mono uppercase text-cortex-text-muted">
                                {scopeLabel(choice.recommendedScope)}
                            </span>
                        </span>
                        <span className="mt-0.5 block text-[10px] leading-4 text-cortex-text-muted">
                            {choice.description}
                        </span>
                    </button>
                ))}
            </div>
        </div>
    );
}

function scopeLabel(scope: MCPToolSetScopeKind): string {
    if (scope === "all") return "Everyone";
    if (scope === "group") return "Group";
    if (scope === "host") return "Host";
    return scope;
}
