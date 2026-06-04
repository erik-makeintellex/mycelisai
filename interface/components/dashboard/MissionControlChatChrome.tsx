"use client";

import { useState } from "react";
import { Brain, Loader2, Megaphone, Trash2, Zap } from "lucide-react";
import { councilLabel } from "@/lib/labels";
import { useCortexStore } from "@/store/useCortexStore";
import { COUNCIL_META, toolToActivity } from "./missionControlChatHelpers";

export function SomaActivityIndicator({ isBroadcasting, assistantName }: { isBroadcasting: boolean; assistantName: string }) {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const recentTool = isBroadcasting ? null : [...streamLogs].reverse().find((l) => l.type === "tool.invoked");
    const label = isBroadcasting
        ? "Broadcasting to all teams..."
        : recentTool
        ? toolToActivity(recentTool)
        : `${assistantName} is thinking...`;

    return (
        <div className="flex items-center gap-2 px-3 py-2.5">
            <Loader2 className={`w-3 h-3 animate-spin flex-shrink-0 ${isBroadcasting ? "text-cortex-warning" : "text-cortex-primary"}`} />
            <span className={`text-sm font-mono ${isBroadcasting ? "text-cortex-warning/70" : "text-cortex-text-muted"}`}>
                {label}
            </span>
            <span className="flex gap-0.5 ml-0.5">
                {[0, 1, 2].map((i) => (
                    <span
                        key={i}
                        className={`inline-block w-1 h-1 rounded-full animate-bounce ${isBroadcasting ? "bg-cortex-warning/60" : "bg-cortex-primary/60"}`}
                        style={{ animationDelay: `${i * 150}ms` }}
                    />
                ))}
            </span>
        </div>
    );
}

export function DirectCouncilButton({
    councilMembers,
    directTarget,
    setDirectTarget,
    setCouncilTarget,
    assistantName,
}: {
    councilMembers: { id: string; role: string }[];
    directTarget: string | null;
    setDirectTarget: (id: string | null) => void;
    setCouncilTarget: (id: string) => void;
    assistantName: string;
}) {
    const [open, setOpen] = useState(false);
    const members = councilMembers.filter((m) => m.id !== "admin");

    return (
        <div className="relative">
            <button
                onClick={() => setOpen((p) => !p)}
                className={`flex items-center gap-1 px-1.5 py-0.5 rounded text-[8px] font-mono transition-colors ${
                    directTarget
                        ? "bg-cortex-warning/10 text-cortex-warning border border-cortex-warning/30"
                        : "text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border"
                }`}
                title="Direct message to a specific council member"
            >
                <Zap className="w-2.5 h-2.5" />
                Direct
            </button>
            {open && (
                <div className="absolute top-full left-0 mt-1 bg-cortex-surface border border-cortex-border rounded-lg shadow-xl z-50 min-w-[140px] py-1">
                    <button
                        className="w-full px-3 py-1.5 text-left text-[10px] font-mono text-cortex-primary hover:bg-cortex-border/50 transition-colors"
                        onClick={() => {
                            setDirectTarget(null);
                            setCouncilTarget("admin");
                            setOpen(false);
                        }}
                    >
                        &larr; {assistantName}
                    </button>
                    <div className="border-t border-cortex-border/40 my-0.5" />
                    {members.map((m) => {
                        const meta = COUNCIL_META[m.id];
                        return (
                            <button
                                key={m.id}
                                className={`w-full px-3 py-1.5 text-left text-[10px] font-mono hover:bg-cortex-border/50 transition-colors ${
                                    directTarget === m.id ? "font-bold" : ""
                                } ${meta?.color ?? "text-cortex-text-muted"}`}
                                onClick={() => {
                                    setDirectTarget(m.id);
                                    setCouncilTarget(m.id);
                                    setOpen(false);
                                }}
                            >
                                {meta?.label ?? m.role}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

export function MissionControlChatHeader({
    assistantName,
    broadcastMode,
    clearMissionChat,
    councilMembers,
    directTarget,
    focusedTeamName,
    isLoading,
    messageCount,
    setBroadcastMode,
    setCouncilTarget,
    setDirectTarget,
    showAdvancedRouting,
    simpleMode,
}: {
    assistantName: string;
    broadcastMode: boolean;
    clearMissionChat: () => void;
    councilMembers: { id: string; role: string }[];
    directTarget: string | null;
    focusedTeamName?: string | null;
    isLoading: boolean;
    messageCount: number;
    setBroadcastMode: (next: (prev: boolean) => boolean) => void;
    setCouncilTarget: (id: string) => void;
    setDirectTarget: (id: string | null) => void;
    showAdvancedRouting: boolean;
    simpleMode: boolean;
}) {
    return (
        <div className="h-10 px-4 border-b border-cortex-border flex items-center gap-2 flex-shrink-0 bg-cortex-surface/50">
            {showAdvancedRouting && broadcastMode ? (
                <>
                    <Megaphone className="w-3.5 h-3.5 text-cortex-warning" />
                    <span className="text-[9px] font-bold uppercase tracking-widest text-cortex-warning">Broadcast</span>
                </>
            ) : (
                <>
                    <Brain className="w-3.5 h-3.5 text-cortex-primary" />
                    <div className="min-w-0">
                        <span className="block text-[9px] font-bold uppercase tracking-widest text-cortex-primary font-mono">
                            {assistantName}
                        </span>
                        {simpleMode ? (
                            <span className="block text-[11px] text-cortex-text-muted">
                                {focusedTeamName
                                    ? `Team chat for ${focusedTeamName}. Soma can still reference other work when you ask.`
                                    : "Ask for plans, changes, files, decisions, or follow-up work."}
                            </span>
                        ) : null}
                    </div>
                    {showAdvancedRouting && directTarget && (
                        <span className="text-[9px] font-mono text-cortex-warning">
                            &rarr; {councilLabel(directTarget, assistantName).name}
                        </span>
                    )}
                    {showAdvancedRouting ? (
                        <DirectCouncilButton
                            councilMembers={councilMembers}
                            directTarget={directTarget}
                            setDirectTarget={setDirectTarget}
                            setCouncilTarget={setCouncilTarget}
                            assistantName={assistantName}
                        />
                    ) : null}
                </>
            )}

            <div className="ml-auto flex items-center gap-1">
                {isLoading && (
                    <span className="flex items-center gap-1 text-[9px] font-mono text-cortex-info">
                        <Loader2 className="w-3 h-3 animate-spin" />
                    </span>
                )}
                {showAdvancedRouting ? (
                    <button
                        onClick={() => setBroadcastMode((prev) => !prev)}
                        className={`p-1 rounded transition-colors ${
                            broadcastMode
                                ? "bg-cortex-warning/20 text-cortex-warning"
                                : "hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
                        }`}
                        title={broadcastMode ? "Broadcast mode ON \u2014 messages go to ALL active teams" : "Broadcast mode \u2014 click to send to all teams instead of one agent"}
                    >
                        <Megaphone className="w-3 h-3" />
                    </button>
                ) : null}
                {messageCount > 0 && !isLoading && (
                    <button
                        onClick={clearMissionChat}
                        className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Clear chat"
                    >
                        <Trash2 className="w-3 h-3" />
                    </button>
                )}
            </div>
        </div>
    );
}
