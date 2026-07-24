/**
 * ReactFlow mock for jsdom test environment.
 * ReactFlow relies on DOM APIs (ResizeObserver, SVG rendering) unavailable in jsdom.
 * This mock provides minimal stubs so components using ReactFlow can mount.
 */
import { vi } from 'vitest';
import type { ReactNode } from 'react';

type ChildrenProps = {
    children?: ReactNode;
};

type FlowPoint = {
    x: number;
    y: number;
};

// Stub ResizeObserver (ReactFlow dependency)
if (typeof window !== 'undefined' && !window.ResizeObserver) {
    window.ResizeObserver = class implements ResizeObserver {
        observe() {}
        unobserve() {}
        disconnect() {}
    };
}

// Mock the reactflow module
vi.mock('reactflow', async () => {
    const React = await import('react');

    const ReactFlow = React.forwardRef<HTMLDivElement, ChildrenProps>(({ children }, ref) =>
        React.createElement('div', { 'data-testid': 'react-flow', ref }, children)
    );
    ReactFlow.displayName = 'ReactFlow';

    const Background = () =>
        React.createElement('div', { 'data-testid': 'react-flow-background' });

    const Controls = () =>
        React.createElement('div', { 'data-testid': 'react-flow-controls' });

    const MiniMap = () =>
        React.createElement('div', { 'data-testid': 'react-flow-minimap' });

    const Handle = () =>
        React.createElement('div', { 'data-testid': 'react-flow-handle' });

    const Panel = ({ children }: ChildrenProps) =>
        React.createElement('div', { 'data-testid': 'react-flow-panel' }, children);

    // Position enum
    const Position = {
        Left: 'left',
        Right: 'right',
        Top: 'top',
        Bottom: 'bottom',
    };

    // MarkerType enum
    const MarkerType = {
        Arrow: 'arrow',
        ArrowClosed: 'arrowclosed',
    };

    // Hooks
    const useNodesState = (initialNodes: unknown[] = []) => {
        const [nodes, setNodes] = React.useState<unknown[]>(initialNodes);
        const onNodesChange = vi.fn();
        return [nodes, setNodes, onNodesChange];
    };

    const useEdgesState = (initialEdges: unknown[] = []) => {
        const [edges, setEdges] = React.useState<unknown[]>(initialEdges);
        const onEdgesChange = vi.fn();
        return [edges, setEdges, onEdgesChange];
    };

    const useReactFlow = () => ({
        getNodes: vi.fn(() => []),
        getEdges: vi.fn(() => []),
        setNodes: vi.fn(),
        setEdges: vi.fn(),
        fitView: vi.fn(),
        zoomIn: vi.fn(),
        zoomOut: vi.fn(),
        project: vi.fn((position: FlowPoint) => position),
    });

    const useOnConnect = vi.fn();
    const addEdge = vi.fn((edge: unknown, edges: unknown[]) => [...edges, edge]);
    const applyNodeChanges = vi.fn((_changes: unknown[], nodes: unknown[]) => nodes);
    const applyEdgeChanges = vi.fn((_changes: unknown[], edges: unknown[]) => edges);

    return {
        __esModule: true,
        default: ReactFlow,
        ReactFlow,
        ReactFlowProvider: ({ children }: ChildrenProps) =>
            React.createElement('div', { 'data-testid': 'react-flow-provider' }, children),
        Background,
        Controls,
        MiniMap,
        Handle,
        Panel,
        Position,
        MarkerType,
        useNodesState,
        useEdgesState,
        useReactFlow,
        useOnConnect,
        addEdge,
        applyNodeChanges,
        applyEdgeChanges,
    };
});
