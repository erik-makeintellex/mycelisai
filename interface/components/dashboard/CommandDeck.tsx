"use client";

import React, { useState, useRef } from "react";
import { Send, Terminal, Bot, User } from "lucide-react";
import { useSignalStream, Signal } from "./SignalContext";

export default function CommandDeck() {
    const [input, setInput] = useState("");
    const [sending, setSending] = useState(false);
    const { signals } = useSignalStream();
    const scrollRef = useRef<HTMLDivElement>(null);

    const chatSignals = signals
        .filter((s) => {
            const topic = s.topic || "";
            const source = s.source || "";
            return (
                topic.includes("swarm.global.input.user") ||
                (source === "soma-core" && s.type !== "heartbeat") ||
                topic.includes("swarm.global.output")
            );
        })
        .reverse();

    const handleSend = async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        if (!input.trim() || sending) return;

        setSending(true);
        try {
            await fetch("/api/swarm/command", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    content: input,
                    source: "userspace",
                }),
            });
            setInput("");
        } catch (err) {
            console.error("Failed to send command", err);
        } finally {
            setSending(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-cortex-bg border-r border-cortex-border relative overflow-hidden">
            {/* Header */}
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur flex justify-between items-center">
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest flex items-center gap-2">
                    <Terminal className="w-3 h-3" />
                    Command Deck
                </h3>
                <span className="text-[10px] text-cortex-success font-mono flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-cortex-success animate-pulse" />
                    ONLINE
                </span>
            </div>

            {/* Chat History */}
            <div
                className="flex-1 overflow-y-auto p-4 space-y-4 flex flex-col-reverse"
                ref={scrollRef}
            >
                {chatSignals.map((sig, i) => (
                    <ChatMessage key={i} signal={sig} />
                ))}

                {chatSignals.length === 0 && (
                    <div className="text-center text-cortex-text-muted/50 font-mono text-xs mt-10">
                        AWAITING COMMAND INPUT...
                    </div>
                )}
            </div>

            {/* Input Area */}
            <div className="p-4 border-t border-cortex-border bg-cortex-surface/30">
                <form onSubmit={handleSend} className="relative">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder="Enter command..."
                        className="w-full bg-cortex-surface border border-cortex-border rounded p-3 pl-4 pr-12 text-cortex-text-main focus:border-cortex-primary focus:outline-none placeholder-cortex-text-muted font-mono text-sm shadow-inner"
                        disabled={sending}
                    />
                    <button
                        type="submit"
                        disabled={!input.trim() || sending}
                        className="absolute right-2 top-2 p-1.5 rounded bg-cortex-primary/20 text-cortex-primary hover:bg-cortex-primary/30 disabled:opacity-50 transition-colors"
                    >
                        <Send className="w-4 h-4" />
                    </button>
                </form>
            </div>
        </div>
    );
}

function ChatMessage({ signal }: { signal: Signal }) {
    const isUser = signal.topic?.includes("user");

    return (
        <div
            className={`flex gap-3 text-sm ${isUser ? "flex-row-reverse" : ""}`}
        >
            <div
                className={`mt-0.5 min-w-[24px] h-6 rounded flex items-center justify-center ${isUser ? "bg-blue-900/30 text-blue-400" : "bg-cortex-primary/20 text-cortex-primary"}`}
            >
                {isUser ? (
                    <User className="w-3 h-3" />
                ) : (
                    <Bot className="w-3 h-3" />
                )}
            </div>
            <div
                className={`flex flex-col max-w-[80%] ${isUser ? "items-end" : "items-start"}`}
            >
                <div
                    className={`px-3 py-2 rounded-lg font-mono text-xs ${isUser ? "bg-blue-950/40 border border-blue-800/50 text-blue-100" : "bg-cortex-surface border border-cortex-border text-cortex-text-main"}`}
                >
                    {signal.message}
                </div>
                <span className="text-[9px] text-cortex-text-muted font-mono mt-1">
                    {signal.timestamp
                        ? new Date(signal.timestamp).toLocaleTimeString([], {
                              hour12: false,
                          })
                        : "Now"}{" "}
                    &bull; {signal.source}
                </span>
            </div>
        </div>
    );
}
