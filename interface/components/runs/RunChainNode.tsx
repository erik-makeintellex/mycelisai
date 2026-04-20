"use client";

import React, { useMemo, useState } from "react";
import { ChevronDown, ChevronRight, GitBranch, Clock3, CircleDot } from "lucide-react";
import type { MissionRun } from "@/types/events";

function statusClass(status: MissionRun["status"]): string {
    switch (status) {
        case "completed":
            return "bg-cortex-success/15 text-cortex-success border-cortex-success/30";
        case "failed":
            return "bg-cortex-danger/15 text-cortex-danger border-cortex-danger/30";
        case "running":
            return "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/30";
        default:
            return "bg-cortex-border/40 text-cortex-text-muted border-cortex-border";
    }
}

function timeLabel(iso: string): string {
    const started = new Date(iso);
    if (Number.isNaN(started.getTime())) return iso;
    return started.toLocaleString();
}

export type ChainNode = {
    run: MissionRun;
    children: ChainNode[];
};

interface Props {
    run: MissionRun;
    children?: ChainNode[];
}

export default function RunChainNode({ run, children = [] }: Props) {
    const [expanded, setExpanded] = useState(true);
    const hasChildren = children.length > 0;

    const metadataEntries = useMemo(() => {
        const entries = run.metadata ? Object.entries(run.metadata) : [];
        return entries.filter(([, value]) => value !== undefined && value !== null);
    }, [run.metadata]);

    return (
        <div className="border border-cortex-border rounded-md bg-cortex-surface overflow-hidden">
            <div className="px-3 py-2 flex items-start gap-2">
                <div className="mt-0.5 flex h-4 w-4 items-center justify-center">
                    <CircleDot className="h-3 w-3 text-cortex-primary" />
                </div>

                <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <span className={`text-[9px] font-mono font-bold px-1.5 py-0.5 rounded border uppercase ${statusClass(run.status)}`}>
                            {run.status}
                        </span>
                        <span className="text-[10px] font-mono font-bold text-cortex-text-main break-all">
                            {run.id}
                        </span>
                        <span className="text-[9px] font-mono text-cortex-text-muted/70">
                            mission: {run.mission_id}
                        </span>
                    </div>

                    <div className="mt-1 flex flex-wrap items-center gap-2 text-[9px] font-mono text-cortex-text-muted/70">
                        <span className="flex items-center gap-1">
                            <GitBranch className="h-3 w-3" />
                            depth {run.run_depth}
                        </span>
                        <span className="flex items-center gap-1">
                            <Clock3 className="h-3 w-3" />
                            started {timeLabel(run.started_at)}
                        </span>
                        {run.completed_at && (
                            <span>
                                completed {timeLabel(run.completed_at)}
                            </span>
                        )}
                        {run.parent_run_id && (
                            <span>
                                parent {run.parent_run_id}
                            </span>
                        )}
                    </div>

                    {metadataEntries.length > 0 && (
                        <div className="mt-1 flex flex-wrap gap-1.5">
                            {metadataEntries.map(([key, value]) => (
                                <span
                                    key={key}
                                    className="text-[9px] font-mono px-1.5 py-0.5 rounded border border-cortex-border bg-cortex-bg text-cortex-text-muted break-all"
                                >
                                    {key}: {String(value)}
                                </span>
                            ))}
                        </div>
                    )}
                </div>

                {hasChildren && (
                    <button
                        type="button"
                        onClick={() => setExpanded((p) => !p)}
                        className="mt-0.5 flex items-center gap-1 rounded px-1.5 py-1 text-[9px] font-mono text-cortex-text-muted hover:text-cortex-primary transition-colors"
                        aria-label={expanded ? "Collapse child runs" : "Expand child runs"}
                    >
                        {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                        {children.length}
                    </button>
                )}
            </div>

            {hasChildren && expanded && (
                <div className="border-t border-cortex-border/70 bg-cortex-bg/40 px-3 py-3 pl-6">
                    <div className="space-y-2">
                        {children.map((child) => (
                            <RunChainNode key={child.run.id} run={child.run} children={child.children} />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
}
