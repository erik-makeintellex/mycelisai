"use client";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Loader2, MessageSquare, Send, Filter } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import TurnCard from "./TurnCard";

interface Props {
    runId: string;
    runStatus?: string;
}

export default function ConversationLog({ runId, runStatus }: Props) {
    const conversationTurns = useCortexStore((s) => s.conversationTurns);
    const isFetching = useCortexStore((s) => s.isFetchingConversation);
    const fetchRunConversation = useCortexStore((s) => s.fetchRunConversation);
    const interjectInRun = useCortexStore((s) => s.interjectInRun);

    const [agentFilter, setAgentFilter] = useState<string | undefined>(undefined);
    const [interjectionText, setInterjectionText] = useState("");
    const [isSending, setIsSending] = useState(false);
    const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const scrollRef = useRef<HTMLDivElement>(null);

    const isRunning = runStatus === "running";

    // Fetch conversation (with optional agent filter)
    const doFetch = useCallback(() => {
        fetchRunConversation(runId, agentFilter);
    }, [fetchRunConversation, runId, agentFilter]);

    // Initial fetch
    useEffect(() => {
        doFetch();
    }, [doFetch]);

    // Auto-poll every 5s while running
    useEffect(() => {
        if (!isRunning) {
            if (pollingRef.current) clearInterval(pollingRef.current);
            return;
        }
        pollingRef.current = setInterval(doFetch, 5000);
        return () => {
            if (pollingRef.current) clearInterval(pollingRef.current);
        };
    }, [isRunning, doFetch]);

    // Auto-scroll to bottom when new turns arrive
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [conversationTurns]);

    // Derive unique agent IDs for filter buttons
    const agentIds = useMemo(() => {
        if (!conversationTurns) return [];
        const ids = new Set<string>();
        for (const turn of conversationTurns) {
            if (turn.agent_id) ids.add(turn.agent_id);
        }
        return Array.from(ids).sort();
    }, [conversationTurns]);

    // Handle interjection submission
    const handleInterject = async () => {
        const trimmed = interjectionText.trim();
        if (!trimmed) return;
        setIsSending(true);
        try {
            await interjectInRun(runId, trimmed);
            setInterjectionText("");
        } finally {
            setIsSending(false);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleInterject();
        }
    };

    // Loading state
    if (isFetching && conversationTurns === null) {
        return (
            <div className="flex items-center justify-center py-12 gap-2 text-cortex-text-muted">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="text-sm font-mono">Loading conversation...</span>
            </div>
        );
    }

    // Empty state
    if (!conversationTurns || conversationTurns.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 gap-2 text-cortex-text-muted">
                <MessageSquare className="w-8 h-8 opacity-20" />
                <p className="text-sm font-mono">No conversation data</p>
                {isRunning && (
                    <p className="text-[10px] font-mono opacity-60">Auto-refreshing every 5s</p>
                )}
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full">
            {/* Agent filter bar */}
            {agentIds.length > 1 && (
                <div className="flex items-center gap-1.5 px-1 pb-3 flex-wrap">
                    <Filter className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
                    <button
                        onClick={() => setAgentFilter(undefined)}
                        className={`text-[9px] font-mono px-2 py-0.5 rounded-full border transition-colors ${
                            !agentFilter
                                ? "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/30"
                                : "bg-cortex-surface text-cortex-text-muted border-cortex-border hover:text-cortex-text-main"
                        }`}
                    >
                        All
                    </button>
                    {agentIds.map((id) => (
                        <button
                            key={id}
                            onClick={() => setAgentFilter(agentFilter === id ? undefined : id)}
                            className={`text-[9px] font-mono px-2 py-0.5 rounded-full border transition-colors ${
                                agentFilter === id
                                    ? "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/30"
                                    : "bg-cortex-surface text-cortex-text-muted border-cortex-border hover:text-cortex-text-main"
                            }`}
                        >
                            {id}
                        </button>
                    ))}
                </div>
            )}

            {/* Turns list */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto space-y-0.5">
                {conversationTurns.map((turn) => (
                    <TurnCard key={turn.id} turn={turn} />
                ))}
            </div>

            {/* Interjection input bar — only visible when running */}
            {isRunning && (
                <div className="mt-3 pt-3 border-t border-cortex-border">
                    <div className="flex items-center gap-2">
                        <input
                            type="text"
                            value={interjectionText}
                            onChange={(e) => setInterjectionText(e.target.value)}
                            onKeyDown={handleKeyDown}
                            placeholder="Interject in this run..."
                            disabled={isSending}
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-3 py-2 text-sm font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary/50 disabled:opacity-50"
                        />
                        <button
                            onClick={handleInterject}
                            disabled={isSending || !interjectionText.trim()}
                            className="flex items-center gap-1.5 px-3 py-2 bg-red-500/15 text-red-400 border border-red-500/30 rounded text-[10px] font-mono font-bold hover:bg-red-500/25 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                        >
                            {isSending ? (
                                <Loader2 className="w-3 h-3 animate-spin" />
                            ) : (
                                <Send className="w-3 h-3" />
                            )}
                            Interject
                        </button>
                    </div>
                    <p className="text-[9px] font-mono text-cortex-text-muted/50 mt-1">
                        Interjections are injected into the active agent context as operator overrides.
                    </p>
                </div>
            )}
        </div>
    );
}
