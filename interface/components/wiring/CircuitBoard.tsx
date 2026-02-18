"use client";

import React, { useState, useCallback } from 'react';
import ReactFlow, {
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    Node,
    Edge,
    Position,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { nodeTypes } from './AgentNode';
import { edgeTypes } from './DataWire';
import type { AgentNodeData } from './AgentNode';
import { Send, Loader2, Zap } from 'lucide-react';

/** Blueprint types matching the Go structs */
interface Constraint {
    constraint_id?: string;
    description: string;
}

interface AgentManifest {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
    inputs?: string[];
    outputs?: string[];
}

interface BlueprintTeam {
    name: string;
    role: string;
    agents: AgentManifest[];
}

interface MissionBlueprint {
    mission_id: string;
    intent: string;
    teams: BlueprintTeam[];
    constraints?: Constraint[];
}

/** Layout constants */
const TEAM_WIDTH = 280;
const TEAM_PADDING = 60;
const AGENT_SPACING_Y = 130;
const AGENT_OFFSET_X = 60;
const TEAM_HEADER_HEIGHT = 80;

/** Convert a MissionBlueprint into ReactFlow nodes and edges */
function blueprintToGraph(bp: MissionBlueprint): { nodes: Node[]; edges: Edge[] } {
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    // Track output-to-input topic mappings for edge generation
    const outputMap: Map<string, string> = new Map(); // topic -> agentNodeId

    let teamX = 80;

    bp.teams.forEach((team, tIdx) => {
        // Team group node (background)
        const teamId = `team-${tIdx}`;
        const teamHeight = TEAM_HEADER_HEIGHT + team.agents.length * AGENT_SPACING_Y + 40;

        nodes.push({
            id: teamId,
            type: 'group',
            position: { x: teamX, y: 40 },
            data: { label: '' },
            style: {
                width: TEAM_WIDTH,
                height: teamHeight,
                background: 'rgba(30, 41, 59, 0.4)',
                border: '1px solid rgba(71, 85, 105, 0.4)',
                borderRadius: '12px',
                padding: '8px',
            },
        });

        // Team label node
        nodes.push({
            id: `${teamId}-label`,
            type: 'default',
            position: { x: 12, y: 8 },
            parentNode: teamId,
            extent: 'parent' as const,
            draggable: false,
            data: { label: team.name },
            style: {
                background: 'transparent',
                border: 'none',
                color: '#94a3b8',
                fontSize: '11px',
                fontWeight: 700,
                textTransform: 'uppercase' as const,
                letterSpacing: '0.1em',
                width: TEAM_WIDTH - 24,
                pointerEvents: 'none' as const,
            },
        });

        // Agent nodes within team
        team.agents.forEach((agent, aIdx) => {
            const agentId = `agent-${tIdx}-${aIdx}`;

            nodes.push({
                id: agentId,
                type: 'agentNode',
                position: {
                    x: AGENT_OFFSET_X,
                    y: TEAM_HEADER_HEIGHT + aIdx * AGENT_SPACING_Y,
                },
                parentNode: teamId,
                extent: 'parent' as const,
                data: {
                    label: agent.id,
                    role: agent.role,
                    status: 'online',
                    lastThought: agent.system_prompt
                        ? agent.system_prompt.slice(0, 60)
                        : undefined,
                } as AgentNodeData,
                sourcePosition: Position.Right,
                targetPosition: Position.Left,
            });

            // Register outputs for edge matching
            agent.outputs?.forEach((topic) => {
                outputMap.set(topic, agentId);
            });
        });

        teamX += TEAM_WIDTH + TEAM_PADDING;
    });

    // Generate edges: connect agent outputs to agent inputs by matching topic names
    bp.teams.forEach((team, tIdx) => {
        team.agents.forEach((agent, aIdx) => {
            const targetId = `agent-${tIdx}-${aIdx}`;
            agent.inputs?.forEach((topic) => {
                const sourceId = outputMap.get(topic);
                if (sourceId && sourceId !== targetId) {
                    edges.push({
                        id: `edge-${sourceId}-${targetId}-${topic}`,
                        source: sourceId,
                        target: targetId,
                        type: 'dataWire',
                        data: { type: 'output' },
                        animated: true,
                    });
                }
            });
        });
    });

    return { nodes, edges };
}

export default function CircuitBoard() {
    const [intent, setIntent] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [blueprint, setBlueprint] = useState<MissionBlueprint | null>(null);
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const handleNegotiate = useCallback(async () => {
        if (!intent.trim()) return;

        setLoading(true);
        setError(null);

        try {
            const res = await fetch('/api/v1/intent/negotiate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ intent }),
            });

            const data = await res.json();

            if (data.error) {
                setError(data.error);
                return;
            }

            setBlueprint(data);
            const graph = blueprintToGraph(data);
            setNodes(graph.nodes);
            setEdges(graph.edges);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Request failed');
        } finally {
            setLoading(false);
        }
    }, [intent, setNodes, setEdges]);

    return (
        <div className="h-full flex flex-col bg-zinc-950">
            {/* Intent bar */}
            <div className="px-6 py-4 border-b border-zinc-800 flex gap-3 items-center">
                <Zap className="w-4 h-4 text-cyan-500 flex-shrink-0" />
                <input
                    type="text"
                    value={intent}
                    onChange={(e) => setIntent(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && !loading && handleNegotiate()}
                    placeholder="Describe your mission intent..."
                    className="flex-1 bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2.5 text-sm text-zinc-200 placeholder-zinc-600 font-mono focus:outline-none focus:border-cyan-600 focus:ring-1 focus:ring-cyan-600/30"
                />
                <button
                    onClick={handleNegotiate}
                    disabled={loading || !intent.trim()}
                    className="flex items-center gap-2 px-4 py-2.5 bg-cyan-600 hover:bg-cyan-500 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded-lg text-sm font-medium transition-colors"
                >
                    {loading ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                        <Send className="w-4 h-4" />
                    )}
                    Negotiate
                </button>
            </div>

            {/* Error display */}
            {error && (
                <div className="px-6 py-2 bg-red-950/50 border-b border-red-800/50">
                    <p className="text-xs text-red-400 font-mono">{error}</p>
                </div>
            )}

            {/* Blueprint metadata bar */}
            {blueprint && (
                <div className="px-6 py-2 border-b border-zinc-800 flex items-center gap-4 bg-zinc-900/50">
                    <span className="text-[10px] font-mono text-zinc-500 uppercase">
                        {blueprint.mission_id}
                    </span>
                    <span className="text-[10px] font-mono text-cyan-500">
                        {blueprint.teams.length} team{blueprint.teams.length !== 1 ? 's' : ''}
                    </span>
                    <span className="text-[10px] font-mono text-green-500">
                        {blueprint.teams.reduce((sum, t) => sum + t.agents.length, 0)} agents
                    </span>
                    {blueprint.constraints && blueprint.constraints.length > 0 && (
                        <span className="text-[10px] font-mono text-amber-500">
                            {blueprint.constraints.length} constraint{blueprint.constraints.length !== 1 ? 's' : ''}
                        </span>
                    )}
                </div>
            )}

            {/* ReactFlow canvas */}
            <div className="flex-1">
                {nodes.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-zinc-600">
                        <div className="text-center">
                            <Zap className="w-12 h-12 mx-auto mb-3 opacity-20" />
                            <p className="text-sm font-mono">Enter a mission intent above</p>
                            <p className="text-xs font-mono mt-1 text-zinc-700">
                                The Meta-Architect will decompose it into a team DAG
                            </p>
                        </div>
                    </div>
                ) : (
                    <ReactFlow
                        nodes={nodes}
                        edges={edges}
                        onNodesChange={onNodesChange}
                        onEdgesChange={onEdgesChange}
                        nodeTypes={nodeTypes}
                        edgeTypes={edgeTypes}
                        fitView
                        fitViewOptions={{ padding: 0.3 }}
                        proOptions={{ hideAttribution: true }}
                    >
                        <Background color="#27272a" gap={20} size={1} />
                        <Controls
                            position="bottom-right"
                            style={{ background: '#18181b', border: '1px solid #3f3f46', borderRadius: '8px' }}
                        />
                        <MiniMap
                            nodeColor="#3f3f46"
                            maskColor="rgba(0,0,0,0.7)"
                            style={{ background: '#09090b', border: '1px solid #27272a' }}
                        />
                    </ReactFlow>
                )}
            </div>
        </div>
    );
}
