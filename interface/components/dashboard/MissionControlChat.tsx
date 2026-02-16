"use client";

import React, { useRef, useEffect, useState } from 'react';
import { Send, Loader2, Bot, User, Trash2, Megaphone, Shield } from 'lucide-react';
import { useCortexStore, type ChatMessage } from '@/store/useCortexStore';

function MessageBubble({ msg }: { msg: ChatMessage }) {
    const isUser = msg.role === 'user';
    const isBroadcast = isUser && msg.content.startsWith('[BROADCAST]');

    return (
        <div className={`flex gap-2.5 ${isUser ? 'justify-end' : 'justify-start'}`}>
            {!isUser && (
                <div className="w-6 h-6 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <Bot className="w-3.5 h-3.5 text-cortex-info" />
                </div>
            )}

            <div
                className={`max-w-[85%] px-3 py-2 rounded-lg text-xs font-mono leading-relaxed ${
                    isBroadcast
                        ? 'bg-cortex-warning/10 text-cortex-text-main border border-cortex-warning/30'
                        : isUser
                        ? 'bg-cortex-bg text-cortex-text-main border border-cortex-border'
                        : 'bg-cortex-info/5 text-cortex-text-main border border-cortex-info/20'
                }`}
            >
                {msg.content}
            </div>

            {isUser && (
                <div className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 mt-0.5 ${
                    isBroadcast
                        ? 'bg-cortex-warning/10 border border-cortex-warning/30'
                        : 'bg-cortex-bg border border-cortex-border'
                }`}>
                    {isBroadcast ? (
                        <Megaphone className="w-3.5 h-3.5 text-cortex-warning" />
                    ) : (
                        <User className="w-3.5 h-3.5 text-cortex-text-muted" />
                    )}
                </div>
            )}
        </div>
    );
}

export default function MissionControlChat() {
    const missionChat = useCortexStore((s) => s.missionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const missionChatError = useCortexStore((s) => s.missionChatError);
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const clearMissionChat = useCortexStore((s) => s.clearMissionChat);
    const broadcastToSwarm = useCortexStore((s) => s.broadcastToSwarm);
    const isBroadcasting = useCortexStore((s) => s.isBroadcasting);

    const [input, setInput] = useState('');
    const [broadcastMode, setBroadcastMode] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);

    const isLoading = isMissionChatting || isBroadcasting;

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [missionChat.length]);

    const handleSubmit = () => {
        if (!input.trim() || isLoading) return;

        const isBroadcast = broadcastMode || input.trimStart().startsWith('/all ');
        const content = isBroadcast && input.trimStart().startsWith('/all ')
            ? input.trimStart().slice(5).trim()
            : input.trim();

        if (!content) return;

        if (isBroadcast) {
            broadcastToSwarm(content);
        } else {
            sendMissionChat(content);
        }
        setInput('');
    };

    return (
        <div className="h-full flex flex-col" data-testid="mission-chat">
            {/* Header */}
            <div className="h-8 px-3 border-b border-cortex-border flex items-center gap-2 flex-shrink-0">
                {broadcastMode ? (
                    <Megaphone className="w-3.5 h-3.5 text-cortex-warning" />
                ) : (
                    <Shield className="w-3.5 h-3.5 text-cortex-info" />
                )}
                <span className={`text-[9px] font-bold uppercase tracking-widest ${
                    broadcastMode ? 'text-cortex-warning' : 'text-cortex-text-muted'
                }`}>
                    {broadcastMode ? 'Broadcast' : 'Admin'}
                </span>

                <div className="ml-auto flex items-center gap-1">
                    {isLoading && (
                        <span className="flex items-center gap-1 text-[9px] font-mono text-cortex-info">
                            <Loader2 className="w-3 h-3 animate-spin" />
                        </span>
                    )}
                    <button
                        onClick={() => setBroadcastMode((prev) => !prev)}
                        className={`p-1 rounded transition-colors ${
                            broadcastMode
                                ? 'bg-cortex-warning/20 text-cortex-warning'
                                : 'hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main'
                        }`}
                        title={broadcastMode ? 'Broadcast mode ON (messages go to ALL teams)' : 'Toggle broadcast mode'}
                    >
                        <Megaphone className="w-3 h-3" />
                    </button>
                    {missionChat.length > 0 && !isLoading && (
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

            {/* Broadcast mode indicator */}
            {broadcastMode && (
                <div className="px-3 py-1 bg-cortex-warning/10 border-b border-cortex-warning/30 flex items-center gap-1.5">
                    <Megaphone className="w-3 h-3 text-cortex-warning" />
                    <p className="text-[9px] text-cortex-warning font-mono font-bold uppercase tracking-wider">
                        Messages go to ALL active teams
                    </p>
                </div>
            )}

            {/* Error bar */}
            {missionChatError && (
                <div className="px-3 py-1 bg-cortex-danger/10 border-b border-cortex-danger/30">
                    <p className="text-[9px] text-cortex-danger font-mono">{missionChatError}</p>
                </div>
            )}

            {/* Chat log */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto px-3 py-3 space-y-2.5 scrollbar-thin scrollbar-thumb-cortex-border">
                {missionChat.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        {broadcastMode ? (
                            <Megaphone className="w-8 h-8 mb-2 opacity-20" />
                        ) : (
                            <Shield className="w-8 h-8 mb-2 opacity-20" />
                        )}
                        <p className="text-[10px] font-mono text-center">
                            {broadcastMode
                                ? 'Broadcast directives to all active teams'
                                : 'Ask about team state, mission progress, or direct the council'}
                        </p>
                    </div>
                ) : (
                    missionChat.map((msg, i) => (
                        <MessageBubble key={i} msg={msg} />
                    ))
                )}

                {/* Drafting indicator */}
                {isLoading && (
                    <div className="flex gap-2 justify-start">
                        <div className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 ${
                            isBroadcasting
                                ? 'bg-cortex-warning/10 border border-cortex-warning/20'
                                : 'bg-cortex-info/10 border border-cortex-info/20'
                        }`}>
                            {isBroadcasting ? (
                                <Megaphone className="w-3.5 h-3.5 text-cortex-warning animate-pulse" />
                            ) : (
                                <Bot className="w-3.5 h-3.5 text-cortex-info animate-pulse" />
                            )}
                        </div>
                        <div className={`px-3 py-2 rounded-lg ${
                            isBroadcasting
                                ? 'bg-cortex-warning/5 border border-cortex-warning/20'
                                : 'bg-cortex-info/5 border border-cortex-info/20'
                        }`}>
                            <div className="flex gap-1">
                                <span className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? 'bg-cortex-warning' : 'bg-cortex-info'}`} style={{ animationDelay: '0ms' }} />
                                <span className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? 'bg-cortex-warning' : 'bg-cortex-info'}`} style={{ animationDelay: '150ms' }} />
                                <span className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? 'bg-cortex-warning' : 'bg-cortex-info'}`} style={{ animationDelay: '300ms' }} />
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Input */}
            <div className="px-3 py-2 border-t border-cortex-border flex-shrink-0">
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
                        placeholder={broadcastMode ? 'Broadcast to all teams...' : 'Ask the admin... (or /all to broadcast)'}
                        disabled={isLoading}
                        className={`flex-1 bg-cortex-bg border rounded-lg px-2.5 py-1.5 text-xs text-cortex-text-main placeholder-cortex-text-muted/50 font-mono focus:outline-none focus:ring-1 disabled:opacity-50 ${
                            broadcastMode
                                ? 'border-cortex-warning/40 focus:border-cortex-warning focus:ring-cortex-warning/30'
                                : 'border-cortex-border focus:border-cortex-primary focus:ring-cortex-primary/30'
                        }`}
                    />
                    <button
                        onClick={handleSubmit}
                        disabled={isLoading || !input.trim()}
                        className={`flex items-center justify-center w-8 h-8 disabled:bg-cortex-border disabled:text-cortex-text-muted text-white rounded-lg transition-colors ${
                            broadcastMode
                                ? 'bg-cortex-warning hover:bg-cortex-warning/80'
                                : 'bg-cortex-primary hover:bg-cortex-primary/80'
                        }`}
                    >
                        {isLoading ? (
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        ) : (
                            <Send className="w-3.5 h-3.5" />
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
}
