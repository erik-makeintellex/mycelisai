"use client";

import React from 'react';
import { Brain, Radio, Zap, BookOpen, Trash2, Wrench } from 'lucide-react';
import type { CatalogueAgent } from '@/store/useCortexStore';

// ── Role accent borders (Midnight Cortex) ──────────────────────
const roleBorders: Record<string, string> = {
    cognitive: 'border-l-cortex-primary',
    sensory: 'border-l-cortex-info',
    actuation: 'border-l-cortex-success',
    ledger: 'border-l-cortex-text-muted',
};

// ── Role icons ─────────────────────────────────────────────────
const roleIcons: Record<string, React.ComponentType<{ className?: string }>> = {
    cognitive: Brain,
    sensory: Radio,
    actuation: Zap,
    ledger: BookOpen,
};

// ── Role badge colors ──────────────────────────────────────────
const roleBadgeColors: Record<string, string> = {
    cognitive: 'bg-cortex-primary/20 text-cortex-primary border-cortex-primary/30',
    sensory: 'bg-cortex-info/20 text-cortex-info border-cortex-info/30',
    actuation: 'bg-cortex-success/20 text-cortex-success border-cortex-success/30',
    ledger: 'bg-cortex-text-muted/20 text-cortex-text-muted border-cortex-text-muted/30',
};

interface AgentCardProps {
    agent: CatalogueAgent;
    onSelect: (agent: CatalogueAgent) => void;
    onDelete: (id: string) => void;
}

export default function AgentCard({ agent, onSelect, onDelete }: AgentCardProps) {
    const borderClass = roleBorders[agent.role] ?? 'border-l-cortex-border';
    const Icon = roleIcons[agent.role] ?? Brain;
    const badgeClass = roleBadgeColors[agent.role] ?? roleBadgeColors.cognitive;

    const handleDelete = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (window.confirm(`Delete agent "${agent.name}"? This action cannot be undone.`)) {
            onDelete(agent.id);
        }
    };

    return (
        <div
            onClick={() => onSelect(agent)}
            className={`bg-cortex-surface border border-l-4 ${borderClass} border-cortex-border rounded-xl p-4 cursor-pointer group hover:border-cortex-text-muted transition-all`}
        >
            {/* Header: Icon + Name + Delete */}
            <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2 min-w-0">
                    <div className="w-7 h-7 rounded-md flex items-center justify-center flex-shrink-0 bg-cortex-bg">
                        <Icon className="w-4 h-4 text-cortex-text-muted" />
                    </div>
                    <span className="text-sm font-mono font-semibold text-cortex-text-main truncate">
                        {agent.name}
                    </span>
                </div>
                <button
                    onClick={handleDelete}
                    className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-cortex-danger/20 text-cortex-text-muted hover:text-cortex-danger transition-all flex-shrink-0"
                >
                    <Trash2 className="w-3.5 h-3.5" />
                </button>
            </div>

            {/* Role badge */}
            <div className="flex items-center gap-2 mb-3">
                <span className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded border ${badgeClass}`}>
                    {agent.role}
                </span>
            </div>

            {/* Model badge */}
            {agent.model && (
                <div className="mb-3">
                    <span className="text-[10px] font-mono text-cortex-text-muted bg-cortex-bg px-1.5 py-0.5 rounded">
                        {agent.model}
                    </span>
                </div>
            )}

            {/* Tool count chip */}
            {agent.tools.length > 0 && (
                <div className="flex items-center gap-1 text-cortex-text-muted">
                    <Wrench className="w-3 h-3" />
                    <span className="text-[10px] font-mono">
                        {agent.tools.length} tool{agent.tools.length !== 1 ? 's' : ''}
                    </span>
                </div>
            )}
        </div>
    );
}
