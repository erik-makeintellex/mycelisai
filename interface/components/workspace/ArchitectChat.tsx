"use client";

import React, { useRef, useEffect, useState } from 'react';
import { Send, Loader2, Bot, User, FileJson } from 'lucide-react';
import { useCortexStore, type ChatMessage } from '@/store/useCortexStore';
import TrustSlider from './TrustSlider';
import { WORKSPACE_LABELS } from '@/lib/labels';

function MessageBubble({ msg }: { msg: ChatMessage }) {
    const isUser = msg.role === 'user';

    return (
        <div className={`flex gap-2.5 ${isUser ? 'justify-end' : 'justify-start'}`}>
            {!isUser && (
                <div className="w-7 h-7 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <Bot className="w-4 h-4 text-cortex-info" />
                </div>
            )}

            <div
                className={`max-w-[80%] px-3 py-2 rounded-lg text-sm font-mono leading-relaxed ${
                    isUser
                        ? 'bg-cortex-bg text-cortex-text-main border border-cortex-border'
                        : 'bg-cortex-info/5 text-cortex-text-main border border-cortex-info/20'
                }`}
            >
                {msg.content}
            </div>

            {isUser && (
                <div className="w-7 h-7 rounded-md bg-cortex-bg border border-cortex-border flex items-center justify-center flex-shrink-0 mt-0.5">
                    <User className="w-4 h-4 text-cortex-text-muted" />
                </div>
            )}
        </div>
    );
}

export default function ArchitectChat() {
    const chatHistory = useCortexStore((s) => s.chatHistory);
    const isDrafting = useCortexStore((s) => s.isDrafting);
    const error = useCortexStore((s) => s.error);
    const submitIntent = useCortexStore((s) => s.submitIntent);
    const toggleBlueprintDrawer = useCortexStore((s) => s.toggleBlueprintDrawer);

    const [input, setInput] = useState('');
    const scrollRef = useRef<HTMLDivElement>(null);

    // Auto-scroll on new messages
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [chatHistory.length]);

    const handleSubmit = () => {
        if (!input.trim() || isDrafting) return;
        submitIntent(input);
        setInput('');
    };

    return (
        <div className="h-full flex flex-col bg-cortex-surface">
            {/* Header */}
            <div className="px-4 py-3 border-b border-cortex-border flex items-center gap-2">
                <Bot className="w-4 h-4 text-cortex-info" />
                <span className="text-[11px] font-bold uppercase tracking-widest text-cortex-text-muted">
                    {WORKSPACE_LABELS.metaArchitect}
                </span>
                <div className="ml-auto flex items-center gap-2">
                    {isDrafting && (
                        <span className="flex items-center gap-1.5 text-[10px] font-mono text-cortex-info">
                            <Loader2 className="w-3 h-3 animate-spin" />
                            drafting...
                        </span>
                    )}
                    <button
                        onClick={toggleBlueprintDrawer}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
                        title="Blueprint Library"
                    >
                        <FileJson className="w-3.5 h-3.5" />
                    </button>
                </div>
            </div>

            {/* Trust Economy — Autonomy Threshold */}
            <TrustSlider />

            {/* Error bar */}
            {error && (
                <div className="px-4 py-1.5 bg-cortex-danger/10 border-b border-cortex-danger/30">
                    <p className="text-[10px] text-cortex-danger font-mono">{error}</p>
                </div>
            )}

            {/* Chat log */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3 scrollbar-thin scrollbar-thumb-cortex-border">
                {chatHistory.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        <Bot className="w-10 h-10 mb-2 opacity-20" />
                        <p className="text-xs font-mono">Describe your mission intent</p>
                        <p className="text-[10px] font-mono mt-1 opacity-60">
                            The Mission Architect will decompose it into teams
                        </p>
                    </div>
                ) : (
                    chatHistory.map((msg, i) => (
                        <MessageBubble key={i} msg={msg} />
                    ))
                )}

                {/* Drafting indicator */}
                {isDrafting && (
                    <div className="flex gap-2.5 justify-start">
                        <div className="w-7 h-7 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0">
                            <Bot className="w-4 h-4 text-cortex-info animate-pulse" />
                        </div>
                        <div className="px-3 py-2 rounded-lg bg-cortex-info/5 border border-cortex-info/20">
                            <div className="flex gap-1">
                                <span className="w-1.5 h-1.5 rounded-full bg-cortex-info animate-bounce" style={{ animationDelay: '0ms' }} />
                                <span className="w-1.5 h-1.5 rounded-full bg-cortex-info animate-bounce" style={{ animationDelay: '150ms' }} />
                                <span className="w-1.5 h-1.5 rounded-full bg-cortex-info animate-bounce" style={{ animationDelay: '300ms' }} />
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Input */}
            <div className="px-4 py-3 border-t border-cortex-border">
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
                        placeholder="Describe your mission intent..."
                        disabled={isDrafting}
                        className="flex-1 bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 text-sm text-cortex-text-main placeholder-cortex-text-muted/50 font-mono focus:outline-none focus:border-cortex-primary focus:ring-1 focus:ring-cortex-primary/30 disabled:opacity-50"
                    />
                    <button
                        onClick={handleSubmit}
                        disabled={isDrafting || !input.trim()}
                        className="flex items-center justify-center w-9 h-9 bg-cortex-primary hover:bg-cortex-primary/80 disabled:bg-cortex-border disabled:text-cortex-text-muted text-white rounded-lg transition-colors"
                    >
                        {isDrafting ? (
                            <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                            <Send className="w-4 h-4" />
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
}
