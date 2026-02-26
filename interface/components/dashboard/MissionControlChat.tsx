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
    AlertTriangle,
    Brain,
    Globe,
    Eye,
    Zap,
} from "lucide-react";
import { useCortexStore, type ChatMessage, type ChatArtifactRef, type ChatConsultation } from "@/store/useCortexStore";
import { ChartRenderer, type MycelisChartSpec } from "@/components/charts";
import ProposedActionBlock from "./ProposedActionBlock";
import OrchestrationInspector from "./OrchestrationInspector";
import {
    sourceNodeLabel,
    trustBadge,
    trustTooltip,
    toolLabel,
    councilLabel,
    brainBadge,
    brainDisplayName,
    toolOrigin,
    MODE_LABELS,
} from "@/lib/labels";
import CouncilCallErrorCard from "./CouncilCallErrorCard";

// ── Helpers ──────────────────────────────────────────────────

function trustColor(score?: number): string {
    if (score == null) return "";
    if (score >= 0.8) return "text-cortex-success";
    if (score >= 0.5) return "text-cortex-warning";
    return "text-cortex-danger";
}

const COUNCIL_META: Record<string, { label: string; color: string }> = {
    'council-architect': { label: 'Architect', color: 'text-cortex-info' },
    'council-coder':     { label: 'Coder',     color: 'text-cortex-success' },
    'council-creative':  { label: 'Creative',  color: 'text-cortex-warning' },
    'council-sentry':    { label: 'Sentry',    color: 'text-cortex-danger' },
};

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

// ── Delegation Trace ─────────────────────────────────────────

function DelegationTrace({ consultations }: { consultations: ChatConsultation[] }) {
    if (!consultations?.length) return null;
    return (
        <div className="mt-2 pt-2 border-t border-cortex-border/40">
            <div className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider mb-1.5">
                Soma consulted
            </div>
            <div className="flex flex-wrap gap-1.5">
                {consultations.map((c, i) => {
                    const meta = COUNCIL_META[c.member];
                    return (
                        <div
                            key={i}
                            className="bg-cortex-surface/60 border border-cortex-border/60 rounded px-2 py-1.5 max-w-[180px]"
                        >
                            <div className={`text-[9px] font-mono font-bold mb-0.5 ${meta?.color ?? 'text-cortex-text-muted'}`}>
                                {meta?.label ?? c.member}
                            </div>
                            <div className="text-[9px] text-cortex-text-muted leading-tight line-clamp-2">
                                {c.summary}
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

// ── Soma Activity Indicator ───────────────────────────────────

function toolToActivity(log: { type?: string; payload?: Record<string, string>; message?: string }): string {
    const payload = (log.payload ?? {}) as Record<string, string>;
    const tool = payload.tool ?? payload.tool_name ?? '';
    switch (tool) {
        case 'consult_council':        return `Consulting ${payload.member ?? 'council'}...`;
        case 'generate_blueprint':     return 'Generating mission blueprint...';
        case 'research_for_blueprint': return 'Researching past missions...';
        case 'write_file':             return `Writing ${payload.path ?? 'file'}...`;
        case 'read_file':              return `Reading ${payload.path ?? 'file'}...`;
        case 'search_memory':          return 'Searching memory...';
        case 'recall':                 return 'Recalling context...';
        case 'store_artifact':         return 'Storing artifact...';
        case 'list_teams':             return 'Checking active teams...';
        case 'get_system_status':      return 'Reading system status...';
        default:                       return tool ? `${tool.replace(/_/g, ' ')}...` : (log.message ?? 'Working...');
    }
}

function SomaActivityIndicator({ isBroadcasting }: { isBroadcasting: boolean }) {
    const streamLogs = useCortexStore((s) => s.streamLogs);

    const recentTool = isBroadcasting
        ? null
        : [...streamLogs].reverse().find((l) => l.type === 'tool.invoked');

    const label = isBroadcasting
        ? 'Broadcasting to all teams...'
        : recentTool
        ? toolToActivity(recentTool)
        : 'Soma is thinking...';

    return (
        <div className="flex items-center gap-2 px-3 py-2.5">
            <Loader2 className={`w-3 h-3 animate-spin flex-shrink-0 ${isBroadcasting ? 'text-cortex-warning' : 'text-cortex-primary'}`} />
            <span className={`text-sm font-mono ${isBroadcasting ? 'text-cortex-warning/70' : 'text-cortex-text-muted'}`}>
                {label}
            </span>
            <span className="flex gap-0.5 ml-0.5">
                {[0, 1, 2].map((i) => (
                    <span
                        key={i}
                        className={`inline-block w-1 h-1 rounded-full animate-bounce ${isBroadcasting ? 'bg-cortex-warning/60' : 'bg-cortex-primary/60'}`}
                        style={{ animationDelay: `${i * 150}ms` }}
                    />
                ))}
            </span>
        </div>
    );
}

// ── Direct Council Button ─────────────────────────────────────

function DirectCouncilButton({
    councilMembers,
    directTarget,
    setDirectTarget,
    setCouncilTarget,
}: {
    councilMembers: { id: string; role: string }[];
    directTarget: string | null;
    setDirectTarget: (id: string | null) => void;
    setCouncilTarget: (id: string) => void;
}) {
    const [open, setOpen] = useState(false);
    const members = councilMembers.filter((m) => m.id !== 'admin');

    return (
        <div className="relative">
            <button
                onClick={() => setOpen((p) => !p)}
                className={`flex items-center gap-1 px-1.5 py-0.5 rounded text-[8px] font-mono transition-colors ${
                    directTarget
                        ? 'bg-cortex-warning/10 text-cortex-warning border border-cortex-warning/30'
                        : 'text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border'
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
                        onClick={() => { setDirectTarget(null); setCouncilTarget('admin'); setOpen(false); }}
                    >
                        ← Soma
                    </button>
                    <div className="border-t border-cortex-border/40 my-0.5" />
                    {members.map((m) => {
                        const meta = COUNCIL_META[m.id];
                        return (
                            <button
                                key={m.id}
                                className={`w-full px-3 py-1.5 text-left text-[10px] font-mono hover:bg-cortex-border/50 transition-colors ${
                                    directTarget === m.id ? 'font-bold' : ''
                                } ${meta?.color ?? 'text-cortex-text-muted'}`}
                                onClick={() => { setDirectTarget(m.id); setCouncilTarget(m.id); setOpen(false); }}
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

// ── Message Bubble ───────────────────────────────────────────

function MessageBubble({ msg }: { msg: ChatMessage }) {
    const isUser = msg.role === "user";
    const isBroadcast = isUser && msg.content.startsWith("[BROADCAST]");
    const setInspected = useCortexStore((s) => s.setInspectedMessage);

    // System message: render as centered pill/link
    if (msg.role === 'system') {
        return (
            <div className="flex justify-center my-2">
                {msg.run_id ? (
                    <a
                        href={`/runs/${msg.run_id}`}
                        className="flex items-center gap-2 px-3 py-1.5 rounded-full border border-cortex-success/30 bg-cortex-success/5 text-cortex-success text-[10px] font-mono hover:bg-cortex-success/10 transition-colors"
                    >
                        <Zap className="w-3 h-3" />
                        Mission activated — {msg.run_id.slice(0, 8)}...
                        <ExternalLink className="w-2.5 h-2.5 opacity-60" />
                    </a>
                ) : (
                    <span className="text-[9px] font-mono text-cortex-text-muted px-3 py-1 rounded-full border border-cortex-border">
                        {msg.content}
                    </span>
                )}
            </div>
        );
    }

    return (
        <div className={`flex gap-2.5 ${isUser ? "justify-end" : "justify-start"}`}>
            {!isUser && (
                <div className="w-6 h-6 rounded-md bg-cortex-info/10 border border-cortex-info/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <Bot className="w-3.5 h-3.5 text-cortex-info" />
                </div>
            )}

            <div className="max-w-[85%] flex flex-col gap-0.5">
                {/* Brain provenance header bar */}
                {!isUser && (msg.source_node || msg.brain) && (
                    <div className="flex items-center gap-1.5 px-1 flex-wrap">
                        {/* Source node label */}
                        {msg.source_node && (
                            <span className="text-[8px] font-bold uppercase tracking-widest text-cortex-info font-mono">
                                {sourceNodeLabel(msg.source_node!)}
                            </span>
                        )}
                        {/* Brain provenance */}
                        {msg.brain && (
                            <>
                                <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                                <span
                                    className={`text-[8px] font-mono font-bold uppercase tracking-wide flex items-center gap-1 ${
                                        msg.brain.location === 'remote' ? 'text-amber-400' : 'text-cortex-text-muted'
                                    }`}
                                    title={`Model: ${msg.brain.model_id}\nProvider: ${msg.brain.provider_id}\nLocation: ${msg.brain.location}\nData: ${msg.brain.data_boundary}`}
                                >
                                    {msg.brain.location === 'remote' ? (
                                        <Globe className="w-2.5 h-2.5" />
                                    ) : (
                                        <Brain className="w-2.5 h-2.5" />
                                    )}
                                    {brainBadge(msg.brain.provider_id, msg.brain.location)}
                                </span>
                            </>
                        )}
                        {/* Mode label */}
                        {msg.mode && (
                            <>
                                <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                                <span className={`text-[8px] font-mono font-bold uppercase tracking-wide ${
                                    MODE_LABELS[msg.mode]?.color ?? 'text-cortex-text-muted'
                                }`}>
                                    {MODE_LABELS[msg.mode]?.label ?? msg.mode}
                                </span>
                            </>
                        )}
                        {/* Trust badge */}
                        {msg.trust_score != null && msg.trust_score > 0 && (
                            <>
                                <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                                <span
                                    className={`text-[8px] font-mono font-bold ${trustColor(msg.trust_score)}`}
                                    title={trustTooltip(msg.trust_score)}
                                >
                                    {trustBadge(msg.trust_score)}
                                </span>
                            </>
                        )}
                        {/* Remote data boundary warning */}
                        {msg.brain?.location === 'remote' && (
                            <span className="flex items-center gap-0.5 text-[7px] font-mono text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded border border-amber-400/20">
                                <AlertTriangle className="w-2.5 h-2.5" />
                                External
                            </span>
                        )}
                        {/* Inspect orchestration button */}
                        <button
                            onClick={() => setInspected(msg)}
                            className="p-0.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors ml-1"
                            title="Inspect orchestration"
                        >
                            <Eye className="w-3 h-3" />
                        </button>
                    </div>
                )}

                <div
                    className={`px-3 py-2 rounded-lg text-sm font-mono leading-relaxed ${
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

                {/* Delegation trace — which council members Soma consulted */}
                {!isUser && msg.consultations && msg.consultations.length > 0 && (
                    <div className="px-3 pb-2">
                        <DelegationTrace consultations={msg.consultations} />
                    </div>
                )}

                {/* Proposed action block for proposal-mode messages */}
                {!isUser && msg.mode === "proposal" && msg.proposal && (
                    <ProposedActionBlock message={msg} />
                )}

                {/* Inline artifacts */}
                {!isUser && msg.artifacts && msg.artifacts.length > 0 && (
                    <div className="space-y-1">
                        {msg.artifacts.map((artifact, i) => (
                            <InlineArtifact key={artifact.id || `art-${i}`} artifact={artifact} />
                        ))}
                    </div>
                )}

                {/* Tools used pills with origin badges */}
                {!isUser && msg.tools_used && msg.tools_used.length > 0 && (
                    <div className="flex flex-wrap gap-1 px-1 mt-0.5">
                        {msg.tools_used.map((tool) => {
                            const origin = toolOrigin(tool);
                            return (
                                <span
                                    key={tool}
                                    className={`text-[7px] font-mono px-1.5 py-0.5 rounded border flex items-center gap-1 ${
                                        origin === 'external'
                                            ? 'bg-amber-400/10 text-amber-400 border-amber-400/20'
                                            : origin === 'sandboxed'
                                            ? 'bg-cortex-success/10 text-cortex-success border-cortex-success/20'
                                            : 'bg-cortex-primary/10 text-cortex-primary border-cortex-primary/20'
                                    }`}
                                    title={tool}
                                >
                                    {toolLabel(tool)}
                                    {origin === 'external' && (
                                        <span className="text-[6px] uppercase font-bold opacity-80">Ext</span>
                                    )}
                                    {origin === 'sandboxed' && (
                                        <span className="text-[6px] uppercase font-bold opacity-80">Box</span>
                                    )}
                                </span>
                            );
                        })}
                    </div>
                )}

                {/* Recalled memory indicator */}
                {!isUser && msg.tools_used && (msg.tools_used.includes('recall') || msg.tools_used.includes('search_memory')) && (
                    <div className="flex items-center gap-1 px-1 mt-0.5">
                        <span className="w-1 h-1 rounded-full bg-cortex-primary" />
                        <span className="text-[8px] font-mono text-cortex-primary/70 italic">recalled from memory</span>
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

// ── Soma Offline Guide ────────────────────────────────────────

function SomaOfflineGuide({ onRetry }: { onRetry: () => void }) {
    return (
        <div className="flex flex-col items-center justify-center h-full px-6">
            <div className="w-full max-w-sm border border-cortex-border rounded-xl bg-cortex-surface/60 p-5 space-y-4">
                <div className="flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-cortex-danger" />
                    <span className="text-xs font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                        Soma Offline
                    </span>
                </div>

                <p className="text-sm font-mono text-cortex-text-main leading-relaxed">
                    Your neural organism isn&apos;t running yet. Start it to talk with Soma, the Council, and your crews.
                </p>

                <div>
                    <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1.5">
                        Start the organism
                    </p>
                    <div className="bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 font-mono text-sm text-cortex-primary select-all">
                        inv lifecycle.up
                    </div>
                    <p className="text-[9px] font-mono text-cortex-text-muted/60 mt-1.5">
                        Then verify: <span className="text-cortex-text-muted">inv lifecycle.health</span>
                    </p>
                </div>

                <div className="space-y-1">
                    <p className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1">
                        What comes online
                    </p>
                    {[
                        "Soma — your orchestrator",
                        "Council — Architect, Coder, Creative, Sentry",
                        "Cognitive provider — Ollama (local) or remote",
                    ].map((item) => (
                        <div key={item} className="flex items-center gap-2">
                            <span className="text-cortex-success text-[10px]">✓</span>
                            <span className="text-[10px] font-mono text-cortex-text-muted">{item}</span>
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
    const [fetchedMembers, setFetchedMembers] = useState(false);
    const [directTarget, setDirectTarget] = useState<string | null>(null);
    const scrollRef = useRef<HTMLDivElement>(null);

    const isLoading = isMissionChatting || isBroadcasting;
    const lastUserMessage = [...missionChat].reverse().find((m) => m.role === "user");

    // Ensure Soma is the default target on mount
    useEffect(() => {
        setCouncilTarget('admin');
    }, [setCouncilTarget]);

    // Keep local direct-target label synchronized with global council target.
    useEffect(() => {
        setDirectTarget(councilTarget === "admin" ? null : councilTarget);
    }, [councilTarget]);

    useEffect(() => {
        fetchCouncilMembers().then(() => setFetchedMembers(true));
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

    const retryLastMessage = () => {
        if (!lastUserMessage) return;
        const content = lastUserMessage.content.replace(/^\[BROADCAST\]\s*/i, "");
        if (content.trim()) sendMissionChat(content);
    };

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
                        <Brain className="w-3.5 h-3.5 text-cortex-primary" />
                        <span className="text-[9px] font-bold uppercase tracking-widest text-cortex-primary font-mono">
                            Soma
                        </span>
                        {directTarget && (
                            <span className="text-[9px] font-mono text-cortex-warning">
                                → {councilLabel(directTarget).name}
                            </span>
                        )}
                        <DirectCouncilButton
                            councilMembers={councilMembers}
                            directTarget={directTarget}
                            setDirectTarget={setDirectTarget}
                            setCouncilTarget={setCouncilTarget}
                        />
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
                        title={broadcastMode ? "Broadcast mode ON — messages go to ALL active teams" : "Broadcast mode — click to send to all teams instead of one agent"}
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

            {/* Chat log */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto px-3 py-3 space-y-2.5 scrollbar-thin scrollbar-thumb-cortex-border">
                {missionChatError && (
                    <CouncilCallErrorCard
                        member={directTarget || councilTarget}
                        errorMessage={missionChatError}
                        onRetry={retryLastMessage}
                        onSwitchToSoma={() => {
                            setDirectTarget(null);
                            setCouncilTarget("admin");
                            retryLastMessage();
                        }}
                        onContinueWithSoma={() => {
                            setDirectTarget(null);
                            setCouncilTarget("admin");
                        }}
                    />
                )}
                {missionChat.length === 0 ? (
                    fetchedMembers && councilMembers.length === 0 ? (
                        <SomaOfflineGuide onRetry={() => {
                            setFetchedMembers(false);
                            fetchCouncilMembers().then(() => setFetchedMembers(true));
                        }} />
                    ) : (
                        <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                            {broadcastMode ? (
                                <Megaphone className="w-8 h-8 mb-2 opacity-20" />
                            ) : (
                                <Shield className="w-8 h-8 mb-2 opacity-20" />
                            )}
                            <p className="text-[10px] font-mono text-center">
                                {broadcastMode
                                    ? "Broadcast directives to all active teams"
                                    : directTarget
                                    ? `Direct message to ${councilLabel(directTarget).name}...`
                                    : "Ask Soma about missions, teams, or your system"}
                            </p>
                        </div>
                    )
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
                                    : "bg-cortex-primary/10 border border-cortex-primary/20"
                            }`}
                        >
                            {isBroadcasting ? (
                                <Megaphone className="w-3.5 h-3.5 text-cortex-warning animate-pulse" />
                            ) : (
                                <Brain className="w-3.5 h-3.5 text-cortex-primary animate-pulse" />
                            )}
                        </div>
                        <div
                            className={`rounded-lg ${
                                isBroadcasting
                                    ? "bg-cortex-warning/5 border border-cortex-warning/20"
                                    : "bg-cortex-primary/5 border border-cortex-primary/20"
                            }`}
                        >
                            <SomaActivityIndicator isBroadcasting={isBroadcasting} />
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
                                : directTarget
                                ? `Direct to ${councilLabel(directTarget).name}... (or /all to broadcast)`
                                : "Ask Soma... (or /all to broadcast)"
                        }
                        disabled={isLoading}
                        className={`flex-1 bg-cortex-bg border rounded-lg px-2.5 py-1.5 text-sm text-cortex-text-main placeholder-cortex-text-muted/50 font-mono focus:outline-none focus:ring-1 disabled:opacity-50 ${
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

            {/* Orchestration Inspector drawer */}
            <OrchestrationInspector />
        </div>
    );
}
