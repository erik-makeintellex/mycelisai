"use client"

import { useCallback } from 'react';
import ReactFlow, {
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    addEdge,
    Connection,
    Edge,
    Node,
} from 'reactflow';
import 'reactflow/dist/style.css';

const initialNodes: Node[] = [
    { id: '1', position: { x: 0, y: 0 }, data: { label: 'Mycelis Core' }, type: 'input' },
    { id: '2', position: { x: 0, y: 100 }, data: { label: 'Active Agent Loop' } },
];
const initialEdges: Edge[] = [{ id: 'e1-2', source: '1', target: '2', animated: true }];

export function Canvas() {
    const [nodes, , onNodesChange] = useNodesState(initialNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    const onConnect = useCallback(
        (params: Connection | Edge) => setEdges((eds) => addEdge(params, eds)),
        [setEdges],
    );

    return (
        <div className="w-full h-full bg-[rgb(var(--background))]">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                onConnect={onConnect}
                fitView
            >
                <Background color="#E4E4E7" gap={16} size={1} />
                <Controls className="bg-white border border-zinc-200 shadow-sm" />
                <MiniMap className="bg-zinc-50 border border-zinc-200" />
            </ReactFlow>
        </div>
    );
}
