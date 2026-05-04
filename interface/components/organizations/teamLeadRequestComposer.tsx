"use client";

import type { RefObject } from "react";
import { ArrowRight, Loader2 } from "lucide-react";
import type { TeamLeadGuidedAction } from "@/lib/organizations";
import { GUIDED_ACTIONS } from "@/components/organizations/teamLeadGuidanceNormalization";
import type { TeamLeadPromptSuggestion } from "@/components/organizations/TeamLeadInteractionPanel";

export function TeamLeadPanelHeader({
    organizationName,
    somaName,
    teamLeadName,
    embedded,
}: {
    organizationName: string;
    somaName: string;
    teamLeadName: string;
    embedded: boolean;
}) {
    return (
        <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div>
                <h2 className="text-xl font-semibold text-cortex-text-main">Focused team design with Soma</h2>
                <p className="mt-1 text-sm text-cortex-text-muted">
                    Use this lane when you want Soma to shape a team, delivery lane, or first execution structure for {organizationName}.
                </p>
            </div>
            {!embedded ? (
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Primary counterpart</p>
                    <p className="mt-1">{somaName}</p>
                    <p className="mt-3 font-medium text-cortex-text-main">Operational lead</p>
                    <p className="mt-1">{teamLeadName}</p>
                </div>
            ) : (
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                    <p className="font-medium text-cortex-text-main">How this mode works</p>
                    <p className="mt-1 leading-6">Stay in the same Soma workspace while focusing this mode on team design, delivery lanes, and execution structure.</p>
                </div>
            )}
        </div>
    );
}

export function TeamLeadRequestComposer({
    promptRef,
    draftPrompt,
    promptSuggestions,
    isLoading,
    onPromptChange,
    onSubmit,
}: {
    promptRef: RefObject<HTMLTextAreaElement | null>;
    draftPrompt: string;
    promptSuggestions: TeamLeadPromptSuggestion[];
    isLoading: boolean;
    onPromptChange: (value: string) => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-5 rounded-2xl border border-cortex-primary/30 bg-cortex-primary/10 p-4">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                <div className="space-y-2">
                    <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Advanced team design mode</p>
                    <p className="text-sm leading-6 text-cortex-text-main">
                        Tell Soma what kind of team, delivery lane, or execution structure you want to create. This panel keeps that request focused on team design instead of general organization conversation.
                    </p>
                </div>
                <button type="button" onClick={onSubmit} disabled={isLoading} className="inline-flex items-center justify-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90 disabled:cursor-not-allowed disabled:opacity-60">
                    {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                    Start team design
                </button>
            </div>
            <div className="mt-4 space-y-2">
                <label htmlFor="soma-guided-prompt" className="text-sm font-medium text-cortex-text-main">Tell Soma what team or delivery lane you want to create</label>
                <textarea
                    id="soma-guided-prompt"
                    ref={promptRef}
                    value={draftPrompt}
                    onChange={(event) => onPromptChange(event.target.value)}
                    onKeyDown={(event) => {
                        if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
                            event.preventDefault();
                            onSubmit();
                        }
                    }}
                    placeholder="Tell Soma what team or delivery lane you want to create"
                    className="min-h-[104px] w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                />
                <p className="text-xs leading-6 text-cortex-text-muted">Soma will use this request to choose the right first team-design move. Leave it blank to start with a quick strategy check right away.</p>
                {promptSuggestions.length > 0 ? (
                    <div className="flex flex-wrap gap-2 pt-1">
                        {promptSuggestions.map((suggestion) => (
                            <button key={suggestion.label} type="button" onClick={() => onPromptChange(suggestion.prompt)} className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary">
                                {suggestion.label}
                            </button>
                        ))}
                    </div>
                ) : null}
            </div>
        </div>
    );
}

export function TeamLeadActionList({
    selectedAction,
    isLoading,
    onAction,
}: {
    selectedAction: TeamLeadGuidedAction | null;
    isLoading: boolean;
    onAction: (action: TeamLeadGuidedAction) => void;
}) {
    return (
        <div className="mt-5 grid gap-3">
            {GUIDED_ACTIONS.map((item) => (
                <button
                    key={item.action}
                    onClick={() => onAction(item.action)}
                    disabled={isLoading}
                    className={`w-full rounded-2xl border p-4 text-left transition-colors ${selectedAction === item.action ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"} disabled:cursor-not-allowed disabled:opacity-70`}
                    aria-busy={selectedAction === item.action && isLoading}
                >
                    <div className="flex items-start justify-between gap-3">
                        <div>
                            <p className="text-sm font-semibold text-cortex-text-main">{item.label}</p>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{item.detail}</p>
                        </div>
                        <ArrowRight className={`mt-0.5 h-4 w-4 ${selectedAction === item.action ? "text-cortex-primary" : "text-cortex-text-muted"}`} />
                    </div>
                </button>
            ))}
        </div>
    );
}
