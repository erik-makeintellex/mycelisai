"use client";

import React, { useRef, useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
    Send,
    Loader2,
    Bot,
    User,
    Trash2,
    Megaphone,
    Shield,
    Code,
    FileText,
    Image as ImageIcon,
    Music,
    Database,
    BarChart3,
    File,
    ExternalLink,
    Copy,
    Check,
} from "lucide-react";
import { useCortexStore, type ChatMessage, type ChatArtifactRef } from "@/store/useCortexStore";
import { ChartRenderer, type MycelisChartSpec } from "@/components/charts";

// ── Helpers ──────────────────────────────────────────────────

function trustColor(score?: number): string {
    if (score == null) return "";
    if (score >= 0.8) return "text-cortex-success";
    if (score >= 0.5) return "text-cortex-warning";
    return "text-cortex-danger";
}

function artifactIcon(type: string) {
    switch (type) {
        case "code":
            return Code;
        case "document":
            return FileText;
        case "image":
            return ImageIcon;
        case "audio":
            return Music;
        case "data":
            return Database;
        case "chart":
            return BarChart3;
        default:
            return File;
    }
}

// ── Copy Button ──────────────────────────────────────────────

function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
    };

    return (
        <button
            onClick={handleCopy}
            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
            title="Copy"
        >
            {copied ? <Check className="w-3 h-3 text-cortex-success" /> : <Copy className="w-3 h-3" />}
        </button>
    );
}

// ── Inline Artifact Card ─────────────────────────────────────

function InlineArtifact({ artifact }: { artifact: ChatArtifactRef }) {
    const [expanded, setExpanded] = useState(false);
    const Icon = artifactIcon(artifact.type);

    // Chart rendering
    if (artifact.type === "chart" && artifact.content) {
        try {
            const spec = JSON.parse(artifact.content) as MycelisChartSpec;
            if (spec.chart_type && spec.data) {
                return (
                    <div className="mt-2 border border-cortex-border rounded-lg overflow-hidden bg-cortex-bg">
                        <div className="flex items-center gap-2 px-3 py-1.5 bg-cortex-surface/50 border-b border-cortex-border">
                            <BarChart3 className="w-3 h-3 text-cortex-primary" />
                            <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                                {artifact.title}
                            </span>
                        </div>
                        <div className="p-2" style={{ maxHeight: 280 }}>
                            <ChartRenderer spec={spec} />
                        </div>
                    </div>
                );
            }
        } catch {
            /* fall through to generic card */
        }
    }

    // Image rendering
    if (artifact.type === "image") {
        const src = artifact.url || (artifact.content_type?.startsWith("image/") && artifact.content
            ? `data:${artifact.content_type};base64,${artifact.content}`
            : null);

        return (
            <div className="mt-2 border border-cortex-border rounded-lg overflow-hidden bg-cortex-bg">
                <div className="flex items-center gap-2 px-3 py-1.5 bg-cortex-surface/50 border-b border-cortex-border">
                    <ImageIcon className="w-3 h-3 text-cortex-primary" />
                    <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                        {artifact.title}
                    </span>
                    {artifact.url && (
                        <a href={artifact.url} target="_blank" rel="noopener noreferrer"
                           className="p-0.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors">
                            <ExternalLink className="w-3 h-3" />
                        </a>
                    )}
                </div>
                {src ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={src} alt={artifact.title} className="w-full max-h-80 object-contain bg-cortex-bg p-2" />
                ) : (
                    <div className="flex items-center justify-center h-32 text-cortex-text-muted/40">
                        <ImageIcon className="w-8 h-8" />
                    </div>
                )}
            </div>
        );
    }

    // Code rendering
    if (artifact.type === "code" && artifact.content) {
        return (
            <div className="mt-2 border border-cortex-border rounded-lg overflow-hidden bg-cortex-bg">
                <div className="flex items-center gap-2 px-3 py-1.5 bg-cortex-surface/50 border-b border-cortex-border">
                    <Code className="w-3 h-3 text-cortex-primary" />
                    <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                        {artifact.title}
                    </span>
                    {artifact.content_type && (
                        <span className="text-[8px] font-mono text-cortex-text-muted bg-cortex-bg px-1.5 py-0.5 rounded">
                            {artifact.content_type.replace("text/", "").replace("application/", "")}
                        </span>
                    )}
                    <CopyButton text={artifact.content} />
                </div>
                <pre className="p-3 overflow-x-auto text-[11px] font-mono text-cortex-text-main leading-relaxed max-h-64 overflow-y-auto">
                    {artifact.content}
                </pre>
            </div>
        );
    }

    // Audio rendering
    if (artifact.type === "audio" && artifact.url) {
        return (
            <div className="mt-2 border border-cortex-border rounded-lg overflow-hidden bg-cortex-bg">
                <div className="flex items-center gap-2 px-3 py-1.5 bg-cortex-surface/50 border-b border-cortex-border">
                    <Music className="w-3 h-3 text-cortex-primary" />
                    <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                        {artifact.title}
                    </span>
                </div>
                <div className="p-3">
                    <audio controls className="w-full" src={artifact.url}>
                        Your browser does not support the audio element.
                    </audio>
                </div>
            </div>
        );
    }

    // Data / JSON rendering
    if ((artifact.type === "data" || artifact.type === "document") && artifact.content) {
        let displayContent = artifact.content;
        if (artifact.content_type?.includes("json")) {
            try {
                displayContent = JSON.stringify(JSON.parse(artifact.content), null, 2);
            } catch { /* keep raw */ }
        }

        return (
            <div className="mt-2 border border-cortex-border rounded-lg overflow-hidden bg-cortex-bg">
                <div className="flex items-center gap-2 px-3 py-1.5 bg-cortex-surface/50 border-b border-cortex-border">
                    <Icon className="w-3 h-3 text-cortex-primary" />
                    <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                        {artifact.title}
                    </span>
                    <CopyButton text={displayContent} />
                    <button
                        onClick={() => setExpanded(!expanded)}
                        className="text-[8px] font-mono text-cortex-primary hover:text-cortex-primary/80 transition-colors"
                    >
                        {expanded ? "collapse" : "expand"}
                    </button>
                </div>
                <pre className={`p-3 overflow-x-auto text-[11px] font-mono text-cortex-text-main leading-relaxed overflow-y-auto ${expanded ? "max-h-96" : "max-h-32"}`}>
                    {displayContent}
                </pre>
            </div>
        );
    }

    // Generic file reference (no inline content)
    return (
        <div className="mt-2 border border-cortex-border rounded-lg bg-cortex-bg">
            <div className="flex items-center gap-2 px-3 py-2">
                <Icon className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-[10px] font-mono font-bold text-cortex-text-main flex-1 truncate">
                    {artifact.title}
                </span>
                <span className="text-[8px] font-mono text-cortex-text-muted uppercase">
                    {artifact.type}
                </span>
                {artifact.url && (
                    <a href={artifact.url} target="_blank" rel="noopener noreferrer"
                       className="p-0.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors">
                        <ExternalLink className="w-3 h-3" />
                    </a>
                )}
            </div>
        </div>
    );
}

// ── Markdown Renderer ────────────────────────────────────────

function MessageContent({ content }: { content: string }) {
    return (
        <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
                // Links: open in new tab, styled
                a: ({ href, children }) => (
                    <a
                        href={href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-cortex-primary hover:text-cortex-primary/80 underline underline-offset-2 decoration-cortex-primary/30 hover:decoration-cortex-primary/60 transition-colors inline-flex items-center gap-0.5"
                    >
                        {children}
                        <ExternalLink className="w-2.5 h-2.5 inline-block flex-shrink-0" />
                    </a>
                ),
                // Code blocks with copy button
                pre: ({ children }) => (
                    <div className="relative group my-2">
                        <pre className="bg-cortex-bg border border-cortex-border rounded-lg p-3 overflow-x-auto text-[11px] leading-relaxed max-h-64 overflow-y-auto">
                            {children}
                        </pre>
                    </div>
                ),
                // Inline code
                code: ({ children, className }) => {
                    const isBlock = className?.includes("language-");
                    if (isBlock) {
                        return <code className="font-mono">{children}</code>;
                    }
                    return (
                        <code className="font-mono text-[11px] px-1 py-0.5 rounded bg-cortex-bg border border-cortex-border text-cortex-primary">
                            {children}
                        </code>
                    );
                },
                // Headings
                h1: ({ children }) => (
                    <h1 className="text-sm font-bold text-cortex-text-main mt-3 mb-1 first:mt-0">{children}</h1>
                ),
                h2: ({ children }) => (
                    <h2 className="text-xs font-bold text-cortex-text-main mt-2.5 mb-1 first:mt-0">{children}</h2>
                ),
                h3: ({ children }) => (
                    <h3 className="text-xs font-bold text-cortex-text-muted mt-2 mb-0.5 first:mt-0">{children}</h3>
                ),
                // Paragraphs
                p: ({ children }) => <p className="mb-1.5 last:mb-0">{children}</p>,
                // Lists
                ul: ({ children }) => <ul className="list-disc list-inside space-y-0.5 mb-1.5 ml-1">{children}</ul>,
                ol: ({ children }) => <ol className="list-decimal list-inside space-y-0.5 mb-1.5 ml-1">{children}</ol>,
                li: ({ children }) => <li className="text-cortex-text-main">{children}</li>,
                // Tables (GFM)
                table: ({ children }) => (
                    <div className="overflow-x-auto my-2 border border-cortex-border rounded-lg">
                        <table className="w-full text-[10px] font-mono">{children}</table>
                    </div>
                ),
                thead: ({ children }) => (
                    <thead className="bg-cortex-surface/50 border-b border-cortex-border">{children}</thead>
                ),
                th: ({ children }) => (
                    <th className="px-2 py-1.5 text-left font-bold text-cortex-text-muted uppercase tracking-wider">{children}</th>
                ),
                td: ({ children }) => (
                    <td className="px-2 py-1.5 border-t border-cortex-border/50 text-cortex-text-main">{children}</td>
                ),
                // Blockquotes
                blockquote: ({ children }) => (
                    <blockquote className="border-l-2 border-cortex-primary/40 pl-3 my-1.5 text-cortex-text-muted italic">
                        {children}
                    </blockquote>
                ),
                // Horizontal rule
                hr: () => <hr className="border-cortex-border my-2" />,
                // Strong / em
                strong: ({ children }) => <strong className="font-bold text-cortex-text-main">{children}</strong>,
                em: ({ children }) => <em className="italic text-cortex-text-main/80">{children}</em>,
                // Images in markdown
                img: ({ src, alt }) => (
                    <span className="block my-2">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                            src={src}
                            alt={alt || ""}
                            className="max-w-full max-h-80 rounded-lg border border-cortex-border object-contain"
                        />
                    </span>
                ),
            }}
        >
            {content}
        </ReactMarkdown>
    );
}

// ── Message Bubble ───────────────────────────────────────────

function MessageBubble({ msg }: { msg: ChatMessage }) {
    const isUser = msg.role === "user";
    const isBroadcast = isUser && msg.content.startsWith("[BROADCAST]");

    return (
        <div className={`flex gap-2.5 ${isUser ? "justify-end" : "justify-start"}`}>
            {!isUser && (
                <div className="w-6 h-6 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <Bot className="w-3.5 h-3.5 text-cortex-info" />
                </div>
            )}

            <div className="max-w-[85%] flex flex-col gap-0.5">
                {/* Source label + trust badge */}
                {!isUser && msg.source_node && (
                    <div className="flex items-center gap-1.5 px-1">
                        <span className="text-[8px] font-bold uppercase tracking-widest text-cortex-info font-mono">
                            {msg.source_node.replace("council-", "").toUpperCase()}
                        </span>
                        {msg.trust_score != null && msg.trust_score > 0 && (
                            <span className={`text-[8px] font-mono font-bold ${trustColor(msg.trust_score)}`}>
                                T:{msg.trust_score.toFixed(1)}
                            </span>
                        )}
                    </div>
                )}

                <div
                    className={`px-3 py-2 rounded-lg text-xs font-mono leading-relaxed ${
                        isBroadcast
                            ? "bg-cortex-warning/10 text-cortex-text-main border border-cortex-warning/30"
                            : isUser
                            ? "bg-cortex-bg text-cortex-text-main border border-cortex-border"
                            : "bg-cortex-info/5 text-cortex-text-main border border-cortex-info/20"
                    }`}
                >
                    {isUser ? (
                        msg.content
                    ) : (
                        <MessageContent content={msg.content} />
                    )}
                </div>

                {/* Inline artifacts */}
                {!isUser && msg.artifacts && msg.artifacts.length > 0 && (
                    <div className="space-y-1">
                        {msg.artifacts.map((artifact, i) => (
                            <InlineArtifact key={artifact.id || `art-${i}`} artifact={artifact} />
                        ))}
                    </div>
                )}

                {/* Tools used pills */}
                {!isUser && msg.tools_used && msg.tools_used.length > 0 && (
                    <div className="flex flex-wrap gap-1 px-1 mt-0.5">
                        {msg.tools_used.map((tool) => (
                            <span
                                key={tool}
                                className="text-[7px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20"
                            >
                                {tool}
                            </span>
                        ))}
                    </div>
                )}
            </div>

            {isUser && (
                <div
                    className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 mt-0.5 ${
                        isBroadcast
                            ? "bg-cortex-warning/10 border border-cortex-warning/30"
                            : "bg-cortex-bg border border-cortex-border"
                    }`}
                >
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

// ── Main Component ───────────────────────────────────────────

export default function MissionControlChat() {
    const missionChat = useCortexStore((s) => s.missionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const missionChatError = useCortexStore((s) => s.missionChatError);
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const clearMissionChat = useCortexStore((s) => s.clearMissionChat);
    const broadcastToSwarm = useCortexStore((s) => s.broadcastToSwarm);
    const isBroadcasting = useCortexStore((s) => s.isBroadcasting);
    const councilTarget = useCortexStore((s) => s.councilTarget);
    const councilMembers = useCortexStore((s) => s.councilMembers);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);
    const fetchCouncilMembers = useCortexStore((s) => s.fetchCouncilMembers);

    const [input, setInput] = useState("");
    const [broadcastMode, setBroadcastMode] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);

    const isLoading = isMissionChatting || isBroadcasting;

    useEffect(() => {
        fetchCouncilMembers();
    }, [fetchCouncilMembers]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [missionChat.length]);

    const handleSubmit = () => {
        if (!input.trim() || isLoading) return;

        const isBroadcast = broadcastMode || input.trimStart().startsWith("/all ");
        const content =
            isBroadcast && input.trimStart().startsWith("/all ")
                ? input.trimStart().slice(5).trim()
                : input.trim();

        if (!content) return;

        if (isBroadcast) {
            broadcastToSwarm(content);
        } else {
            sendMissionChat(content);
        }
        setInput("");
    };

    const targetMember = councilMembers.find((m) => m.id === councilTarget);
    const targetLabel = targetMember?.role?.toUpperCase() || councilTarget.toUpperCase();

    return (
        <div className="h-full flex flex-col" data-testid="mission-chat">
            {/* Header */}
            <div className="h-8 px-3 border-b border-cortex-border flex items-center gap-2 flex-shrink-0">
                {broadcastMode ? (
                    <>
                        <Megaphone className="w-3.5 h-3.5 text-cortex-warning" />
                        <span className="text-[9px] font-bold uppercase tracking-widest text-cortex-warning">
                            Broadcast
                        </span>
                    </>
                ) : (
                    <>
                        <Shield className="w-3.5 h-3.5 text-cortex-info" />
                        <select
                            value={councilTarget}
                            onChange={(e) => setCouncilTarget(e.target.value)}
                            className="bg-transparent text-[9px] font-bold uppercase tracking-widest text-cortex-text-muted border-none outline-none cursor-pointer font-mono"
                        >
                            {councilMembers.length === 0 ? (
                                <option value="admin">Admin</option>
                            ) : (
                                councilMembers.map((m) => (
                                    <option key={m.id} value={m.id}>
                                        {m.role.toUpperCase()} ({m.team.replace("-core", "")})
                                    </option>
                                ))
                            )}
                        </select>
                    </>
                )}

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
                                ? "bg-cortex-warning/20 text-cortex-warning"
                                : "hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
                        }`}
                        title={broadcastMode ? "Broadcast mode ON (messages go to ALL teams)" : "Toggle broadcast mode"}
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
                                ? "Broadcast directives to all active teams"
                                : `Ask the ${targetLabel.toLowerCase()} about team state, missions, or direct the council`}
                        </p>
                    </div>
                ) : (
                    missionChat.map((msg, i) => <MessageBubble key={i} msg={msg} />)
                )}

                {/* Drafting indicator */}
                {isLoading && (
                    <div className="flex gap-2 justify-start">
                        <div
                            className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 ${
                                isBroadcasting
                                    ? "bg-cortex-warning/10 border border-cortex-warning/20"
                                    : "bg-cortex-info/10 border border-cortex-info/20"
                            }`}
                        >
                            {isBroadcasting ? (
                                <Megaphone className="w-3.5 h-3.5 text-cortex-warning animate-pulse" />
                            ) : (
                                <Bot className="w-3.5 h-3.5 text-cortex-info animate-pulse" />
                            )}
                        </div>
                        <div
                            className={`px-3 py-2 rounded-lg ${
                                isBroadcasting
                                    ? "bg-cortex-warning/5 border border-cortex-warning/20"
                                    : "bg-cortex-info/5 border border-cortex-info/20"
                            }`}
                        >
                            <div className="flex gap-1">
                                <span
                                    className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? "bg-cortex-warning" : "bg-cortex-info"}`}
                                    style={{ animationDelay: "0ms" }}
                                />
                                <span
                                    className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? "bg-cortex-warning" : "bg-cortex-info"}`}
                                    style={{ animationDelay: "150ms" }}
                                />
                                <span
                                    className={`w-1.5 h-1.5 rounded-full animate-bounce ${isBroadcasting ? "bg-cortex-warning" : "bg-cortex-info"}`}
                                    style={{ animationDelay: "300ms" }}
                                />
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
                        onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
                        placeholder={
                            broadcastMode
                                ? "Broadcast to all teams..."
                                : `Ask the ${targetLabel.toLowerCase()}... (or /all to broadcast)`
                        }
                        disabled={isLoading}
                        className={`flex-1 bg-cortex-bg border rounded-lg px-2.5 py-1.5 text-xs text-cortex-text-main placeholder-cortex-text-muted/50 font-mono focus:outline-none focus:ring-1 disabled:opacity-50 ${
                            broadcastMode
                                ? "border-cortex-warning/40 focus:border-cortex-warning focus:ring-cortex-warning/30"
                                : "border-cortex-border focus:border-cortex-primary focus:ring-cortex-primary/30"
                        }`}
                    />
                    <button
                        onClick={handleSubmit}
                        disabled={isLoading || !input.trim()}
                        className={`flex items-center justify-center w-8 h-8 disabled:bg-cortex-border disabled:text-cortex-text-muted text-white rounded-lg transition-colors ${
                            broadcastMode
                                ? "bg-cortex-warning hover:bg-cortex-warning/80"
                                : "bg-cortex-primary hover:bg-cortex-primary/80"
                        }`}
                    >
                        {isLoading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
                    </button>
                </div>
            </div>
        </div>
    );
}
