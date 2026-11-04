"use client";

import React, { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { formatHumanActivityEvent } from "@/lib/activityLabels";
import type { MissionEvent } from "@/store/useCortexStore";

// ── Event type → badge color ──────────────────────────────────

function eventColor(eventType: string): string {
    switch (eventType) {
        case "mission.started":
        case "mission.completed":  return "bg-cortex-success/15 text-cortex-success border-cortex-success/30";
        case "mission.failed":
        case "tool.failed":        return "bg-cortex-danger/15 text-cortex-danger border-cortex-danger/30";
        case "tool.invoked":       return "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/30";
        case "tool.completed":     return "bg-cortex-info/15 text-cortex-info border-cortex-info/30";
        case "agent.started":
        case "agent.stopped":      return "bg-cortex-text-muted/10 text-cortex-text-muted border-cortex-border";
        case "memory.stored":
        case "memory.recalled":
        case "artifact.created":   return "bg-cortex-warning/15 text-cortex-warning border-cortex-warning/30";
        case "trigger.fired":      return "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/30";
        default:                   return "bg-cortex-border/40 text-cortex-text-muted border-cortex-border";
    }
}

function dotColor(eventType: string): string {
    switch (eventType) {
        case "mission.started":
        case "mission.completed":  return "bg-cortex-success";
        case "mission.failed":
        case "tool.failed":        return "bg-cortex-danger";
        case "tool.invoked":       return "bg-cortex-primary";
        case "tool.completed":     return "bg-cortex-info";
        case "memory.stored":
        case "memory.recalled":
        case "artifact.created":   return "bg-cortex-warning";
        default:                   return "bg-cortex-text-muted/40";
    }
}

function relativeTime(iso: string): string {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diff < 60)   return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
}

function payloadSummary(payload?: Record<string, unknown>): string | null {
    if (!payload) return null;
    if (payload.package_kind === "codex_task_bundle") {
        const label = String(payload.package_slug ?? payload.base_path ?? "codex_task_bundle");
        const fileCount = Array.isArray(payload.files) ? ` · ${payload.files.length} files` : "";
        const proof = typeof payload.proof_path === "string" ? ` · proof: ${payload.proof_path}` : "";
        return `Codex task bundle: ${label}${fileCount}${proof}`;
    }
    if (typeof payload.error === "string")     return payload.error.slice(0, 80);
    if (typeof payload.tool === "string")      return payload.tool;
    if (typeof payload.tool_name === "string") return payload.tool_name;
    if (typeof payload.title === "string")     return payload.title;
    if (typeof payload.rule_name === "string") return payload.rule_name;
    return null;
}

// ── EventCard ─────────────────────────────────────────────────

interface Props {
    event: MissionEvent;
    isLast?: boolean;
    isExpanded?: boolean;
    onToggle?: () => void;
}

export default function EventCard({ event, isLast = false }: Props) {
    const [expanded, setExpanded] = useState(false);
    const humanSentence = formatHumanActivityEvent(event);
    const summary = payloadSummary(event.payload);
    const hasPayload = Boolean(event.payload && Object.keys(event.payload).length > 0);

    return (
        <div className="flex gap-3">
            {/* Timeline spine */}
            <div className="flex flex-col items-center w-4 flex-shrink-0">
                <div className={`w-2.5 h-2.5 rounded-full flex-shrink-0 mt-1 ${dotColor(event.event_type)}`} />
                {!isLast && <div className="flex-1 w-px bg-cortex-border/50 mt-1" />}
            </div>

            {/* Content */}
            <div className="flex-1 pb-4">
                <div className="flex items-start gap-2 flex-wrap">
                    {/* Human-readable event sentence */}
                    <span className="text-xs font-medium text-cortex-text-main leading-relaxed">
                        {humanSentence}
                    </span>

                    {/* Source agent chip */}
                    {event.source_agent && (
                        <span className="text-[9px] font-mono text-cortex-text-muted bg-cortex-surface border border-cortex-border px-1.5 py-0.5 rounded">
                            {event.source_agent}
                        </span>
                    )}

                    {/* Concise payload tag */}
                    {summary && (
                        <span className="text-[10px] font-mono text-cortex-text-muted truncate max-w-[260px]">
                            {summary}
                        </span>
                    )}

                    {/* Timestamp */}
                    <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                        {relativeTime(event.emitted_at)}
                    </span>
                </div>

                {/* Event type badge + View Technical Envelope toggle */}
                <div className="mt-1.5 flex items-center gap-2 flex-wrap">
                    <span className={`text-[9px] font-mono font-bold px-1.5 py-0.5 rounded border ${eventColor(event.event_type)}`}>
                        {event.event_type}
                    </span>

                    {hasPayload && (
                        <button
                            type="button"
                            data-testid="toggle-technical-envelope"
                            onClick={() => setExpanded((p) => !p)}
                            className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-mono text-cortex-text-muted hover:text-cortex-primary transition-colors"
                            aria-expanded={expanded}
                            title={expanded ? "Collapse payload" : "View Technical Envelope"}
                        >
                            {expanded
                                ? <ChevronDown className="w-3 h-3" />
                                : <ChevronRight className="w-3 h-3" />
                            }
                            <span>View Technical Envelope</span>
                        </button>
                    )}
                </div>

                {/* Full payload JSON */}
                {expanded && event.payload && (
                    <pre className="mt-1.5 bg-cortex-bg border border-cortex-border rounded p-2 text-[9px] font-mono text-cortex-text-muted overflow-x-auto max-h-40 overflow-y-auto leading-relaxed">
                        {JSON.stringify(event.payload, null, 2)}
                    </pre>
                )}
            </div>
        </div>
    );
}
