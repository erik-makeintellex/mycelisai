"use client";

import React, { useCallback } from 'react';
import ReactFlow, { Background, BackgroundVariant, Controls, MiniMap, type Node } from 'reactflow';
import 'reactflow/dist/style.css';
import { nodeTypes } from '@/components/wiring/AgentNode';
import { edgeTypes } from '@/components/wiring/DataWire';
import { useCortexStore } from '@/store/useCortexStore';
import { Zap, Loader2, Rocket } from 'lucide-react';

export default function CircuitBoard() {
    const nodes = useCortexStore((s) => s.nodes);
    const edges = useCortexStore((s) => s.edges);
    const onNodesChange = useCortexStore((s) => s.onNodesChange);
    const onEdgesChange = useCortexStore((s) => s.onEdgesChange);
    const blueprint = useCortexStore((s) => s.blueprint);
    const missionStatus = useCortexStore((s) => s.missionStatus);
    const activeMissionId = useCortexStore((s) => s.activeMissionId);
    const isCommitting = useCortexStore((s) => s.isCommitting);
    const instantiateMission = useCortexStore((s) => s.instantiateMission);
    const enterSquadRoom = useCortexStore((s) => s.enterSquadRoom);

    const onNodeDoubleClick = useCallback(
        (_event: React.MouseEvent, node: Node) => {
            if (node.type === 'group') {
                enterSquadRoom(node.id);
            }
        },
        [enterSquadRoom],
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg relative">
            {/* Blueprint metadata bar */}
            {blueprint && (
                <div className="px-4 py-1.5 border-b border-cortex-border flex items-center gap-4 bg-cortex-surface/50">
                    <span className="text-[10px] font-mono text-cortex-text-muted uppercase">
                        {activeMissionId ?? blueprint.mission_id}
                    </span>
                    <span className="text-[10px] font-mono text-cortex-primary">
                        {blueprint.teams.length} team{blueprint.teams.length !== 1 ? 's' : ''}
                    </span>
                    <span className="text-[10px] font-mono text-cortex-success">
                        {blueprint.teams.reduce((sum, t) => sum + t.agents.length, 0)} agents
                    </span>
                    {blueprint.constraints && blueprint.constraints.length > 0 && (
                        <span className="text-[10px] font-mono text-cortex-warning">
                            {blueprint.constraints.length} constraint{blueprint.constraints.length !== 1 ? 's' : ''}
                        </span>
                    )}
                    <span
                        className={`ml-auto text-[9px] font-mono uppercase font-bold ${
                            missionStatus === 'active'
                                ? 'text-cortex-success'
                                : 'text-cortex-text-muted'
                        }`}
                    >
                        {missionStatus === 'active'
                            ? `ACTIVE - ${activeMissionId?.slice(0, 8)}`
                            : 'DRAFT'}
                    </span>
                </div>
            )}

            {/* ReactFlow canvas — always mounted */}
            <div className="flex-1 relative bg-cortex-bg">
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    onNodeDoubleClick={onNodeDoubleClick}
                    nodeTypes={nodeTypes}
                    edgeTypes={edgeTypes}
                    fitView
                    fitViewOptions={{ padding: 0.3 }}
                    proOptions={{ hideAttribution: true }}
                >
                    <Background color="#7983BB" variant={BackgroundVariant.Dots} gap={20} size={1} style={{ backgroundColor: '#2F3349' }} />
                    <Controls
                        position="bottom-right"
                        style={{
                            background: '#2F3349',
                            border: '1px solid #434968',
                            borderRadius: '8px',
                        }}
                    />
                    <MiniMap
                        nodeColor="#434968"
                        maskColor="rgba(37,41,60,0.7)"
                        style={{
                            background: '#25293C',
                            border: '1px solid #434968',
                        }}
                    />
                </ReactFlow>

                {/* Empty-state overlay — sits on top of the live canvas grid */}
                {nodes.length === 0 && (
                    <div className="absolute inset-0 flex items-center justify-center z-10 pointer-events-none">
                        <div className="text-center text-cortex-text-muted">
                            <Zap className="w-12 h-12 mx-auto mb-3 opacity-20" />
                            <p className="text-sm font-mono">Awaiting blueprint</p>
                            <p className="text-xs font-mono mt-1 opacity-60">
                                Negotiate an intent to generate a team DAG
                            </p>
                        </div>
                    </div>
                )}

                {/* INSTANTIATE SWARM button — only visible when draft */}
                {blueprint && missionStatus === 'draft' && (
                    <div className="absolute bottom-6 left-1/2 -translate-x-1/2 z-50">
                        <button
                            onClick={instantiateMission}
                            disabled={isCommitting}
                            className="flex items-center gap-2.5 px-6 py-3 rounded-xl font-mono text-sm font-bold uppercase tracking-wider transition-all duration-300 shadow-lg border border-cortex-success/40 bg-cortex-success/90 hover:bg-cortex-success hover:shadow-[0_0_25px_rgba(40,199,111,0.4)] text-white disabled:opacity-60 disabled:cursor-wait"
                        >
                            {isCommitting ? (
                                <>
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                    Instantiating...
                                </>
                            ) : (
                                <>
                                    <Rocket className="w-4 h-4" />
                                    Instantiate Swarm
                                </>
                            )}
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
