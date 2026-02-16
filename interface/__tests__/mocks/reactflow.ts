/**
 * ReactFlow mock for jsdom test environment.
 * ReactFlow relies on DOM APIs (ResizeObserver, SVG rendering) unavailable in jsdom.
 * This mock provides minimal stubs so components using ReactFlow can mount.
 */
import { vi } from 'vitest';

// Stub ResizeObserver (ReactFlow dependency)
if (typeof window !== 'undefined' && !window.ResizeObserver) {
    window.ResizeObserver = class {
        observe() {}
        unobserve() {}
        disconnect() {}
    } as any;
}

// Mock the reactflow module
vi.mock('reactflow', () => {
    const React = require('react');

    const ReactFlow = React.forwardRef(({ children, ...props }: any, ref: any) =>
        React.createElement('div', { 'data-testid': 'react-flow', ref, ...props }, children)
    );
    ReactFlow.displayName = 'ReactFlow';

    const Background = (props: any) =>
        React.createElement('div', { 'data-testid': 'react-flow-background', ...props });

    const Controls = (props: any) =>
        React.createElement('div', { 'data-testid': 'react-flow-controls', ...props });

    const MiniMap = (props: any) =>
        React.createElement('div', { 'data-testid': 'react-flow-minimap', ...props });

    const Handle = (props: any) =>
        React.createElement('div', { 'data-testid': 'react-flow-handle', ...props });

    const Panel = ({ children, ...props }: any) =>
        React.createElement('div', { 'data-testid': 'react-flow-panel', ...props }, children);

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
    const useNodesState = (initialNodes: any[] = []) => {
        const [nodes, setNodes] = React.useState(initialNodes);
        const onNodesChange = vi.fn();
        return [nodes, setNodes, onNodesChange];
    };

    const useEdgesState = (initialEdges: any[] = []) => {
        const [edges, setEdges] = React.useState(initialEdges);
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
        project: vi.fn((pos: any) => pos),
    });

    const useOnConnect = vi.fn();
    const addEdge = vi.fn((edge: any, edges: any[]) => [...edges, edge]);
    const applyNodeChanges = vi.fn((changes: any[], nodes: any[]) => nodes);
    const applyEdgeChanges = vi.fn((changes: any[], edges: any[]) => edges);

    return {
        __esModule: true,
        default: ReactFlow,
        ReactFlow,
        ReactFlowProvider: ({ children }: any) =>
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
