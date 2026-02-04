"use client";

import { useEffect, useState } from 'react';
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
import { Network, ArrowRight } from "lucide-react";

// Types
interface WiringData {
    inputs: string[];
    outputs: string[];
}

interface Team {
    id: string;
    name: string;
}

export default function WireGraph() {
    const [team, setTeam] = useState<Team | null>(null);
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    // 1. Fetch Default Team
    useEffect(() => {
        fetch("/api/v1/teams")
            .then(res => res.json())
            .then(data => {
                // response is { teams: [...], count: N } or just array depending on implementation
                // admin.go handleTeams returns map
                // "teams": [...]
                if (data.teams && data.teams.length > 0) {
                    setTeam(data.teams[0]);
                } else if (Array.isArray(data) && data.length > 0) {
                    // Fallback
                    setTeam(data[0]);
                }
            })
            .catch(err => console.error("Failed to fetch teams", err));
    }, []);

    // 2. Fetch Wiring
    useEffect(() => {
        if (!team) return;

        fetch(`/api/v1/teams/${team.id}/wiring`)
            .then(res => res.json())
            .then((data: WiringData) => {
                if (!data) return;

                const newNodes: Node[] = [];
                const newEdges: Edge[] = [];
                let yOffset = 0;

                // Center Node: The Team
                newNodes.push({
                    id: 'team-node',
                    position: { x: 400, y: 200 },
                    data: { label: team.name },
                    type: 'default', // Using default style for now
                    style: {
                        background: '#fff',
                        border: '2px solid #000',
                        padding: 10,
                        fontWeight: 'bold',
                        width: 150,
                        textAlign: 'center'
                    },
                });

                // Inputs (Left)
                data.inputs.forEach((topic, idx) => {
                    const id = `in-${idx}`;
                    newNodes.push({
                        id: id,
                        position: { x: 100, y: 100 + idx * 100 },
                        data: { label: topic },
                        sourcePosition: Position.Right,
                        targetPosition: Position.Left,
                        style: { background: '#f0fdf4', border: '1px solid #16a34a', color: '#166534' }
                    });
                    newEdges.push({
                        id: `e-${id}`,
                        source: id,
                        target: 'team-node',
                        label: 'publishes',
                        animated: true,
                        style: { stroke: '#16a34a' }
                    });
                });

                // Outputs (Right)
                data.outputs.forEach((topic, idx) => {
                    const id = `out-${idx}`;
                    newNodes.push({
                        id: id,
                        position: { x: 700, y: 100 + idx * 100 },
                        data: { label: topic },
                        sourcePosition: Position.Right,
                        targetPosition: Position.Left,
                        style: { background: '#eff6ff', border: '1px solid #2563eb', color: '#1e40af' }
                    });
                    newEdges.push({
                        id: `e-${id}`,
                        source: 'team-node',
                        target: id,
                        label: 'subscribes',
                        animated: true,
                        style: { stroke: '#2563eb' }
                    });
                });

                setNodes(newNodes);
                setEdges(newEdges);
            });

    }, [team, setNodes, setEdges]);

    return (
        <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            fitView
        >
            <Background color="#E4E4E7" gap={16} size={1} />
            <Controls />
            <MiniMap />
        </ReactFlow>
    );
}
