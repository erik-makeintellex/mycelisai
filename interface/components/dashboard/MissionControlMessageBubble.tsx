"use client";

import { AlertTriangle, Bot, Brain, ExternalLink, Eye, Globe, Megaphone, User, Zap } from "lucide-react";
import {
    sourceNodeLabel,
    trustBadge,
    trustTooltip,
    toolLabel,
    brainBadge,
    MODE_LABELS,
    toolOrigin,
} from "@/lib/labels";
import { useCortexStore, type ChatConsultation, type ChatMessage } from "@/store/useCortexStore";
import InlineArtifact from "./InlineArtifact";
import MissionControlMarkdown from "./MissionControlMarkdown";
import ProposedActionBlock from "./ProposedActionBlock";
import ExecutionSummaryCard from "@/components/soma/ExecutionSummaryCard";
import {
    artifactResultSummary,
    askClassBadge,
    consultationResultSummary,
    COUNCIL_META,
    trustColor,
} from "./missionControlChatHelpers";

function DelegationTrace({ consultations, assistantName }: { consultations: ChatConsultation[]; assistantName: string }) {
    if (!consultations?.length) return null;
    return (
        <div className="mt-2 pt-2 border-t border-cortex-border/40">
            <div className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider mb-1.5">
                {assistantName} consulted
            </div>
            <div className="flex flex-wrap gap-1.5">
                {consultations.map((c, i) => {
                    const meta = COUNCIL_META[c.member];
                    return (
                        <div key={i} className="bg-cortex-surface/60 border border-cortex-border/60 rounded px-2 py-1.5 max-w-[180px]">
                            <div className={`text-[9px] font-mono font-bold mb-0.5 ${meta?.color ?? "text-cortex-text-muted"}`}>
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

function MessageMeta({ msg, assistantName }: { msg: ChatMessage; assistantName: string }) {
    const setInspected = useCortexStore((s) => s.setInspectedMessage);
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const askBadge = askClassBadge(msg.ask_class);

    if (!msg.source_node && !msg.brain) return null;
    return (
        <div className="flex items-center gap-1.5 px-1 flex-wrap">
            {msg.source_node && (
                <span className="text-[8px] font-bold uppercase tracking-widest text-cortex-info font-mono">
                    {sourceNodeLabel(msg.source_node, assistantName)}
                </span>
            )}
            {advancedMode && msg.brain && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span
                        className={`text-[8px] font-mono font-bold uppercase tracking-wide flex items-center gap-1 ${
                            msg.brain.location === "remote" ? "text-amber-400" : "text-cortex-text-muted"
                        }`}
                        title={`Model: ${msg.brain.model_id}\nProvider: ${msg.brain.provider_id}\nLocation: ${msg.brain.location}\nData: ${msg.brain.data_boundary}`}
                    >
                        {msg.brain.location === "remote" ? <Globe className="w-2.5 h-2.5" /> : <Brain className="w-2.5 h-2.5" />}
                        {brainBadge(msg.brain.provider_id, msg.brain.location)}
                    </span>
                </>
            )}
            {advancedMode && msg.mode && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold uppercase tracking-wide ${MODE_LABELS[msg.mode]?.color ?? "text-cortex-text-muted"}`}>
                        {MODE_LABELS[msg.mode]?.label ?? msg.mode}
                    </span>
                </>
            )}
            {askBadge && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold px-1.5 py-0.5 rounded border ${askBadge.tone}`}>
                        {askBadge.label}
                    </span>
                </>
            )}
            {msg.trust_score != null && msg.trust_score > 0 && (
                <>
                    <span className="text-[7px] text-cortex-text-muted">&bull;</span>
                    <span className={`text-[8px] font-mono font-bold ${trustColor(msg.trust_score)}`} title={trustTooltip(msg.trust_score)}>
                        {trustBadge(msg.trust_score)}
                    </span>
                </>
            )}
            {advancedMode && msg.brain?.location === "remote" && (
                <span className="flex items-center gap-0.5 text-[7px] font-mono text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded border border-amber-400/20">
                    <AlertTriangle className="w-2.5 h-2.5" />
                    External
                </span>
            )}
            {advancedMode && (
                <button
                    onClick={() => setInspected(msg)}
                    className="p-0.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors ml-1"
                    title="Inspect orchestration"
                >
                    <Eye className="w-3 h-3" />
                </button>
            )}
        </div>
    );
}

function ToolsUsed({ tools }: { tools?: string[] }) {
    if (!tools?.length) return null;
    return (
        <div className="flex flex-wrap gap-1 px-1 mt-0.5">
            {tools.map((tool) => {
                const origin = toolOrigin(tool);
                return (
                    <span
                        key={tool}
                        className={`text-[7px] font-mono px-1.5 py-0.5 rounded border flex items-center gap-1 ${
                            origin === "external"
                                ? "bg-amber-400/10 text-amber-400 border-amber-400/20"
                                : origin === "sandboxed"
                                ? "bg-cortex-success/10 text-cortex-success border-cortex-success/20"
                                : "bg-cortex-primary/10 text-cortex-primary border-cortex-primary/20"
                        }`}
                        title={tool}
                    >
                        {toolLabel(tool)}
                        {origin === "external" && <span className="text-[6px] uppercase font-bold opacity-80">Ext</span>}
                        {origin === "sandboxed" && <span className="text-[6px] uppercase font-bold opacity-80">Box</span>}
                    </span>
                );
            })}
        </div>
    );
}

export default function MissionControlMessageBubble({ msg }: { msg: ChatMessage }) {
    const isUser = msg.role === "user";
    const isBroadcast = isUser && msg.content.startsWith("[BROADCAST]");
    const assistantName = useCortexStore((s) => s.assistantName);
    const artifactSummary = artifactResultSummary(msg.artifacts);
    const consultationSummary = consultationResultSummary(msg.consultations);

    if (msg.role === "system") {
        return (
            <div className="my-2 flex justify-center">
                <div className="flex w-full max-w-[85%] flex-col items-center gap-2">
                    {msg.run_id ? (
                        <a href={`/runs/${msg.run_id}`} className="flex items-center gap-2 px-3 py-1.5 rounded-full border border-cortex-success/30 bg-cortex-success/5 text-cortex-success text-[10px] font-mono hover:bg-cortex-success/10 transition-colors">
                            <Zap className="w-3 h-3" />
                            Mission activated &mdash; {msg.run_id.slice(0, 8)}...
                            <ExternalLink className="w-2.5 h-2.5 opacity-60" />
                        </a>
                    ) : (
                        <span className="text-[9px] font-mono text-cortex-text-muted px-3 py-1 rounded-full border border-cortex-border">
                            {msg.content}
                        </span>
                    )}
                    {msg.execution_summary && (
                        <div className="w-full">
                            <ExecutionSummaryCard
                                summary={msg.execution_summary}
                                runId={msg.run_id}
                                artifacts={msg.artifacts}
                            />
                        </div>
                    )}
                </div>
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
                {!isUser && <MessageMeta msg={msg} assistantName={assistantName} />}
                <div className={`px-3 py-2 rounded-lg text-sm font-mono leading-relaxed ${
                    isBroadcast
                        ? "bg-cortex-warning/10 text-cortex-text-main border border-cortex-warning/30"
                        : isUser
                        ? "bg-cortex-bg text-cortex-text-main border border-cortex-border"
                        : "bg-cortex-info/5 text-cortex-text-main border border-cortex-info/20"
                }`}>
                    {isUser ? msg.content : <MissionControlMarkdown content={msg.content} />}
                </div>
                {!isUser && msg.consultations?.length ? (
                    <div className="px-3 pb-2">
                        <DelegationTrace consultations={msg.consultations} assistantName={assistantName} />
                    </div>
                ) : null}
                {!isUser && msg.execution_summary && (
                    <ExecutionSummaryCard
                        summary={msg.execution_summary}
                        runId={msg.run_id}
                        artifacts={msg.artifacts}
                    />
                )}
                {!isUser && msg.proposal && <ProposedActionBlock message={msg} />}
                {!isUser && artifactSummary && (
                    <div className="rounded-lg border border-cortex-primary/20 bg-cortex-primary/5 px-3 py-2">
                        <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-primary">
                            Returned output
                        </div>
                        <p className="mt-1 text-sm text-cortex-text-main leading-6">{artifactSummary}</p>
                    </div>
                )}
                {!isUser && msg.ask_class === "specialist_consultation" && consultationSummary && (
                    <div className="rounded-lg border border-cortex-warning/20 bg-cortex-warning/5 px-3 py-2">
                        <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-warning">
                            Specialist context
                        </div>
                        <p className="mt-1 text-sm text-cortex-text-main leading-6">{consultationSummary}</p>
                    </div>
                )}
                {!isUser && msg.artifacts?.length ? (
                    <div className="space-y-1">
                        {msg.artifacts.map((artifact, i) => <InlineArtifact key={artifact.id || `art-${i}`} artifact={artifact} />)}
                    </div>
                ) : null}
                {!isUser && <ToolsUsed tools={msg.tools_used} />}
                {!isUser && msg.tools_used && (msg.tools_used.includes("recall") || msg.tools_used.includes("search_memory")) && (
                    <div className="flex items-center gap-1 px-1 mt-0.5">
                        <span className="w-1 h-1 rounded-full bg-cortex-primary" />
                        <span className="text-[8px] font-mono text-cortex-primary/70 italic">recalled from memory</span>
                    </div>
                )}
            </div>
            {isUser && (
                <div className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 mt-0.5 ${
                    isBroadcast ? "bg-cortex-warning/10 border border-cortex-warning/30" : "bg-cortex-bg border border-cortex-border"
                }`}>
                    {isBroadcast ? <Megaphone className="w-3.5 h-3.5 text-cortex-warning" /> : <User className="w-3.5 h-3.5 text-cortex-text-muted" />}
                </div>
            )}
        </div>
    );
}
