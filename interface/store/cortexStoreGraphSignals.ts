import type { Edge, Node } from 'reactflow';
import type { StreamSignal } from '@/store/cortexStoreTypes';

export function solidifyNodes(nodes: Node[]): Node[] {
    return nodes.map((node) => {
        if (!node.className?.includes('ghost-draft')) return node;

        const solidNode = { ...node, className: '' };
        if (node.type === 'group') {
            solidNode.style = {
                ...node.style,
                border: '1px solid rgba(71, 85, 105, 0.6)',
                boxShadow: '0 0 12px rgba(6, 182, 212, 0.15)',
            };
        }
        if (node.type === 'agentNode') {
            solidNode.data = { ...node.data, status: 'online' };
        }
        return solidNode;
    });
}

/** Dispatch an SSE signal to matching ReactFlow nodes. */
export function dispatchSignalToNodes(signal: StreamSignal, nodes: Node[]): Node[] | null {
    const src = signal.source;
    if (!src) return null;

    let changed = false;
    const updated = nodes.map((node) => {
        if (node.id !== src && node.data?.label !== src) return node;
        changed = true;

        if (signal.type === 'thought' || signal.type === 'cognitive') {
            return { ...node, data: { ...node.data, isThinking: true, lastThought: signal.message } };
        }
        if (signal.type === 'artifact' || signal.type === 'output') {
            return {
                ...node,
                data: { ...node.data, isThinking: false, lastThought: signal.message ?? node.data.lastThought },
            };
        }
        if (signal.type === 'error') {
            return {
                ...node,
                data: { ...node.data, status: 'error', isThinking: false, lastThought: signal.message },
            };
        }
        return node;
    });

    return changed ? updated : null;
}
