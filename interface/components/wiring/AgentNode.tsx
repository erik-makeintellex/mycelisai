"use client";

import React, { memo } from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import { Bot, Cpu, Shield, Eye, Database, Zap } from 'lucide-react';

// ── Node Category (Phase 5.2 — Universal Bus) ──────────────────
// Hardware and Software are structurally identical NATS publishers.
// Visual differentiation is by channel type, not physical form.
export type NodeCategory = 'cognitive' | 'sensory' | 'actuation' | 'ledger';

export interface AgentNodeData {
    label: string;
    role: string;
    status: 'online' | 'busy' | 'error' | 'offline';
    lastThought?: string;
    isThinking?: boolean;
    nodeType?: NodeCategory;
    trustScore?: number;
}

// ── Role → NodeCategory mapping ────────────────────────────────
const roleToNodeType: Record<string, NodeCategory> = {
    architect: 'cognitive',
    coder: 'cognitive',
    creative: 'cognitive',
    chat: 'cognitive',
    sentry: 'actuation',
    executor: 'actuation',
    observer: 'sensory',
    ingress: 'sensory',
    archivist: 'ledger',
    memory: 'ledger',
};

// ── Iconography by role ────────────────────────────────────────
const roleIcons: Record<string, React.ComponentType<{ className?: string }>> = {
    architect: Cpu,
    sentry: Shield,
    observer: Eye,
    archivist: Database,
    executor: Zap,
};

// ── Node type → left border accent (Vuexy Dark Protocol) ──────
// Cognitive (LLMs)     → cortex-primary (Purple #7367F0)
// Sensory  (Ingress)   → cortex-info    (Cyan   #00CFE8)
// Actuation (Exec)     → cortex-success (Green  #28C76F)
// Ledger   (Memory)    → cortex-text-muted      (#7983BB)
const nodeTypeBorders: Record<NodeCategory, string> = {
    cognitive: 'border-l-cortex-primary',
    sensory: 'border-l-cortex-info',
    actuation: 'border-l-cortex-success',
    ledger: 'border-l-cortex-text-muted',
};

const nodeTypeGlows: Record<NodeCategory, string> = {
    cognitive: 'shadow-[0_0_8px_rgba(115,103,240,0.15)]',
    sensory: 'shadow-[0_0_8px_rgba(0,207,232,0.15)]',
    actuation: 'shadow-[0_0_8px_rgba(40,199,111,0.15)]',
    ledger: 'shadow-[0_0_8px_rgba(121,131,187,0.15)]',
};

// ── Status indicators ──────────────────────────────────────────
const statusColors: Record<string, string> = {
    online: 'bg-cortex-success shadow-[0_0_6px_rgba(40,199,111,0.6)]',
    busy: 'bg-cortex-warning shadow-[0_0_6px_rgba(255,159,67,0.6)]',
    error: 'bg-cortex-danger shadow-[0_0_6px_rgba(234,84,85,0.6)]',
    offline: 'bg-cortex-text-muted',
};

// ── Role badge colors (mapped to node category) ────────────────
const categoryBadgeColors: Record<NodeCategory, string> = {
    cognitive: 'bg-cortex-primary/20 text-cortex-primary border-cortex-primary/30',
    sensory: 'bg-cortex-info/20 text-cortex-info border-cortex-info/30',
    actuation: 'bg-cortex-success/20 text-cortex-success border-cortex-success/30',
    ledger: 'bg-cortex-text-muted/20 text-cortex-text-muted border-cortex-text-muted/30',
};

function AgentNodeComponent({ data }: NodeProps<AgentNodeData>) {
    const nodeType = data.nodeType ?? roleToNodeType[data.role] ?? 'cognitive';
    const Icon = roleIcons[data.role] ?? Bot;
    const statusClass = statusColors[data.status] ?? statusColors.offline;
    const badgeClass = categoryBadgeColors[nodeType];
    const borderClass = nodeTypeBorders[nodeType];
    const glowClass = nodeTypeGlows[nodeType];
    const thought = data.lastThought
        ? data.lastThought.length > 80
            ? data.lastThought.slice(0, 77) + '...'
            : data.lastThought
        : null;

    const isThinking = data.isThinking ?? false;

    return (
        <div className="relative group">
            {/* Activity ring — visible when agent is thinking */}
            {isThinking && (
                <div className="absolute -inset-1.5 rounded-xl border-2 border-cortex-info/60 activity-ring pointer-events-none" />
            )}

            {/* Target Handle (left) */}
            <Handle
                type="target"
                position={Position.Left}
                className="!w-2.5 !h-2.5 !bg-cortex-text-muted !border-2 !border-cortex-border"
            />

            {/* Node body — left border accent by node category */}
            <div className={`bg-cortex-surface border border-l-4 ${borderClass} rounded-lg px-4 py-3 min-w-[160px] max-w-[220px] transition-all ${
                isThinking
                    ? 'border-cortex-info/50 shadow-[0_0_12px_rgba(0,207,232,0.2)]'
                    : `border-cortex-border hover:border-cortex-text-muted ${glowClass}`
            }`}>
                {/* Status dot + trust score */}
                <div className="absolute top-2 right-2 flex items-center gap-1">
                    {isThinking && (
                        <span className="text-[8px] font-mono text-cortex-info uppercase animate-pulse">
                            thinking
                        </span>
                    )}
                    {data.trustScore !== undefined && (
                        <span className={`text-[8px] font-mono ${
                            data.trustScore >= 0.7 ? 'text-cortex-success' :
                            data.trustScore >= 0.4 ? 'text-cortex-warning' :
                            'text-cortex-danger'
                        }`}>
                            {data.trustScore.toFixed(1)}
                        </span>
                    )}
                    <span className={`inline-block w-2 h-2 rounded-full ${statusClass}`} />
                </div>

                {/* Icon + Name */}
                <div className="flex items-center gap-2 mb-1.5">
                    <div className={`w-7 h-7 rounded-md flex items-center justify-center flex-shrink-0 ${
                        isThinking ? 'bg-cortex-info/10' : 'bg-cortex-bg'
                    }`}>
                        <Icon className={`w-4 h-4 ${isThinking ? 'text-cortex-info' : 'text-cortex-text-muted'}`} />
                    </div>
                    <span className="text-sm font-semibold text-cortex-text-main truncate">
                        {data.label}
                    </span>
                </div>

                {/* Role badge */}
                <div className="mb-1.5">
                    <span className={`text-[10px] font-mono font-bold px-1.5 py-0.5 rounded border ${badgeClass}`}>
                        {data.role}
                    </span>
                </div>

                {/* Thought bubble — enhanced when thinking */}
                {thought && (
                    <div className={`mt-2 rounded px-2 py-1.5 transition-colors ${
                        isThinking
                            ? 'bg-cortex-info/5 border border-cortex-info/20'
                            : 'bg-cortex-bg/60 border border-cortex-border'
                    }`}>
                        <p className={`text-[10px] leading-snug italic ${
                            isThinking ? 'text-cortex-info/80' : 'text-cortex-text-muted'
                        }`}>
                            &ldquo;{thought}&rdquo;
                        </p>
                    </div>
                )}
            </div>

            {/* Source Handle (right) */}
            <Handle
                type="source"
                position={Position.Right}
                className="!w-2.5 !h-2.5 !bg-cortex-text-muted !border-2 !border-cortex-border"
            />
        </div>
    );
}

const AgentNode = memo(AgentNodeComponent);
AgentNode.displayName = 'AgentNode';

export default AgentNode;

export const nodeTypes = {
    agentNode: AgentNode,
};
