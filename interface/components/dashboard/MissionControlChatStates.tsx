"use client";

import { Brain, Megaphone } from "lucide-react";
import { councilLabel } from "@/lib/labels";
import {
    SomaSuggestionBar,
    type SomaSuggestion,
} from "@/components/soma/SomaSuggestionBar";

export function BroadcastModeIndicator() {
    return (
        <div className="px-3 py-1 bg-cortex-warning/10 border-b border-cortex-warning/30 flex items-center gap-1.5">
            <Megaphone className="w-3 h-3 text-cortex-warning" />
            <p className="text-[9px] text-cortex-warning font-mono font-bold uppercase tracking-wider">
                Messages go to ALL active teams
            </p>
        </div>
    );
}

export function SomaOfflineGuide({ onRetry, assistantName }: { onRetry: () => void; assistantName: string }) {
    return (
        <div className="flex flex-col items-center justify-center h-full px-6">
            <div className="w-full max-w-sm border border-cortex-border rounded-xl bg-cortex-surface/60 p-5 space-y-4">
                <div className="flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-cortex-danger" />
                    <span className="text-xs font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                        {assistantName} Offline
                    </span>
                </div>

                <p className="text-sm font-mono text-cortex-text-main leading-relaxed">
                    Your Mycelis runtime isn&apos;t running yet. Start it to talk with {assistantName}, advisor support, and your crews.
                </p>

                <div>
                    <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1.5">
                        Source-mode recovery
                    </p>
                    <div className="space-y-1 bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 font-mono text-xs text-cortex-primary">
                        <div className="select-all">uv run inv native-infra.status</div>
                        <div className="select-all">uv run inv native-infra.up</div>
                        <div className="select-all">uv run inv db.migrate</div>
                        <div className="select-all">uv run inv lifecycle.up --frontend</div>
                    </div>
                    <p className="text-[9px] font-mono text-cortex-text-muted/60 mt-1.5">
                        Use the first line to inspect, then start missing services and the frontend.
                    </p>
                </div>

                <div className="space-y-1">
                    <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1">
                        What comes online
                    </p>
                    {[
                        { label: assistantName, detail: "your orchestrator" },
                        { label: "Advisor support", detail: "Architect, Coder, Creative, Sentry" },
                        { label: "Cognitive provider", detail: "Ollama (local) or remote" },
                    ].map((item) => (
                        <div key={item.label} className="flex items-center gap-2">
                            <span className="text-cortex-success text-[10px]">&#10003;</span>
                            <span className="text-[10px] font-mono text-cortex-text-muted">
                                {item.label} &mdash; {item.detail}
                            </span>
                        </div>
                    ))}
                </div>

                <button
                    onClick={onRetry}
                    className="w-full py-2 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 hover:bg-cortex-primary/20 text-cortex-primary text-sm font-mono transition-all"
                >
                    Retry Connection
                </button>
            </div>
        </div>
    );
}

export function MissionControlEmptyState({
    assistantName,
    broadcastMode,
    currentTeamName,
    directTarget,
    onStarterPrompt,
    showAdvancedRouting,
    simpleMode,
    suggestions,
}: {
    assistantName: string;
    broadcastMode: boolean;
    currentTeamName?: string;
    directTarget: string | null;
    onStarterPrompt: (prompt: string) => void;
    showAdvancedRouting: boolean;
    simpleMode: boolean;
    suggestions: readonly SomaSuggestion[];
}) {
    const message = showAdvancedRouting && broadcastMode
        ? "Broadcast directives to all active teams"
        : showAdvancedRouting && directTarget
        ? `Direct message to ${councilLabel(directTarget, assistantName).name}...`
        : currentTeamName
        ? `Tell ${assistantName} what you want to plan, review, create, or run for ${currentTeamName}.`
        : `Tell ${assistantName} what you want to plan, review, create, or run.`;

    return (
        <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
            {showAdvancedRouting && broadcastMode ? (
                <Megaphone className="w-8 h-8 mb-2 opacity-20" />
            ) : (
                <Brain className="w-8 h-8 mb-2 opacity-25 text-cortex-primary" />
            )}
            <p className="max-w-md text-center text-sm leading-6">{message}</p>
            {simpleMode ? (
                <div className="mt-4 space-y-2">
                    <SomaSuggestionBar suggestions={suggestions} onSelect={onStarterPrompt} />
                </div>
            ) : null}
        </div>
    );
}
