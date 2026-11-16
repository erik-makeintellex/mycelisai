"use client";

import React from "react";
import Link from "next/link";
import { Activity, Boxes, Brain, FolderCog, Megaphone, Users, Zap } from "lucide-react";
import { councilLabel } from "@/lib/labels";
import { useCortexStore } from "@/store/useCortexStore";

export interface WorkspaceContextHeaderProps {
    assistantName: string;
    currentTeamName?: string | null;
    agentCount?: number;
    directTarget?: string | null;
    broadcastMode?: boolean;
}

export default function WorkspaceContextHeader({
    assistantName,
    currentTeamName,
    agentCount = 0,
    directTarget,
    broadcastMode = false,
}: WorkspaceContextHeaderProps) {
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);
    const isTeamFocused = Boolean(currentTeamName?.trim());
    const isDirectSpecialist = Boolean(!isTeamFocused && directTarget);

    const workspaceTitle = broadcastMode
        ? "[Organization Broadcast Lane]"
        : isTeamFocused
        ? `[${currentTeamName} Workspace]`
        : isDirectSpecialist && directTarget
        ? `[${councilLabel(directTarget, assistantName).name} Workspace]`
        : `[${assistantName} Executive Workspace]`;

    const statusBadge = broadcastMode
        ? "All Active Teams Targeted"
        : isTeamFocused
        ? `${agentCount} ${agentCount === 1 ? "Agent" : "Agents"} Active`
        : isDirectSpecialist && directTarget
        ? `${councilLabel(directTarget, assistantName).subtitle || "Specialist"} Active`
        : "Standing Council Active";

    const Icon = broadcastMode
        ? Megaphone
        : isTeamFocused
        ? Users
        : isDirectSpecialist
        ? Zap
        : Brain;

    return (
        <div
            className="px-4 py-2 border-b border-cortex-border bg-cortex-bg/70 flex items-center justify-between gap-2 flex-wrap flex-shrink-0"
            data-testid="workspace-context-header"
        >
            <div className="flex items-center flex-wrap gap-2 min-w-0">
                <Icon className="w-3.5 h-3.5 text-cortex-primary flex-shrink-0" />
                <span className="text-xs font-mono font-semibold text-cortex-text-main truncate">
                    {workspaceTitle}
                </span>
                <span className="text-cortex-text-muted text-xs font-mono">|</span>
                <span className="inline-flex items-center gap-1.5 text-xs font-mono text-cortex-success">
                    <span className="w-1.5 h-1.5 rounded-full bg-cortex-success" />
                    {statusBadge}
                </span>
            </div>
            <div className="flex items-center gap-1.5 flex-wrap" data-testid="workspace-header-quick-actions">
                <button
                    type="button"
                    data-testid="workspace-header-status-btn"
                    onClick={() => setStatusDrawerOpen?.(true)}
                    className="inline-flex items-center gap-1 rounded-md border border-cortex-border bg-cortex-surface px-2 py-1 text-[11px] font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors"
                >
                    <Activity className="w-3 h-3" />
                    Service Status
                </button>
                <Link
                    href="/deployments"
                    data-testid="workspace-header-deployments-link"
                    className="inline-flex items-center gap-1 rounded-md border border-cortex-border bg-cortex-surface px-2 py-1 text-[11px] font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors"
                >
                    <Boxes className="w-3 h-3" />
                    Active Teams
                </Link>
                <Link
                    href="/resources"
                    data-testid="workspace-header-resources-link"
                    className="inline-flex items-center gap-1 rounded-md border border-cortex-border bg-cortex-surface px-2 py-1 text-[11px] font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors"
                >
                    <FolderCog className="w-3 h-3" />
                    Files &amp; Tools
                </Link>
            </div>
        </div>
    );
}
