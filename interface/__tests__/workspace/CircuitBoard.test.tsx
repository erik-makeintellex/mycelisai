import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock reactflow with factory (must return default export properly)
vi.mock('reactflow', () => {
    const React = require('react');
    const ReactFlow = React.forwardRef(({ children, ...props }: any, ref: any) =>
        React.createElement('div', { 'data-testid': 'react-flow', ref, ...props }, children)
    );
    ReactFlow.displayName = 'ReactFlow';
    return {
        __esModule: true,
        default: ReactFlow,
        ReactFlow,
        ReactFlowProvider: ({ children }: any) =>
            React.createElement('div', { 'data-testid': 'react-flow-provider' }, children),
        Background: (props: any) => React.createElement('div', { 'data-testid': 'react-flow-background', ...props }),
        BackgroundVariant: { Dots: 'dots', Lines: 'lines', Cross: 'cross' },
        Controls: (props: any) => React.createElement('div', { 'data-testid': 'react-flow-controls', ...props }),
        MiniMap: (props: any) => React.createElement('div', { 'data-testid': 'react-flow-minimap', ...props }),
        Handle: (props: any) => React.createElement('div', { 'data-testid': 'react-flow-handle', ...props }),
        Panel: ({ children, ...props }: any) => React.createElement('div', { 'data-testid': 'react-flow-panel', ...props }, children),
        Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
        MarkerType: { Arrow: 'arrow', ArrowClosed: 'arrowclosed' },
        useNodesState: (init: any[] = []) => {
            const [nodes, setNodes] = React.useState(init);
            return [nodes, setNodes, vi.fn()];
        },
        useEdgesState: (init: any[] = []) => {
            const [edges, setEdges] = React.useState(init);
            return [edges, setEdges, vi.fn()];
        },
        useReactFlow: () => ({
            getNodes: vi.fn(() => []),
            getEdges: vi.fn(() => []),
            setNodes: vi.fn(),
            setEdges: vi.fn(),
            fitView: vi.fn(),
            zoomIn: vi.fn(),
            zoomOut: vi.fn(),
            project: vi.fn((pos: any) => pos),
        }),
        addEdge: vi.fn((e: any, es: any[]) => [...es, e]),
        applyNodeChanges: vi.fn((_: any, n: any[]) => n),
        applyEdgeChanges: vi.fn((_: any, e: any[]) => e),
    };
});
vi.mock('reactflow/dist/style.css', () => ({}));
vi.mock('@/components/wiring/AgentNode', () => ({
    nodeTypes: { agentNode: () => <div data-testid="agent-node" /> },
}));
vi.mock('@/components/wiring/DataWire', () => ({
    edgeTypes: { dataWire: () => <div data-testid="data-wire" /> },
}));
vi.mock('@/components/wiring/WiringAgentEditor', () => ({
    default: () => <div data-testid="wiring-agent-editor" />,
}));
vi.mock('lucide-react', () => ({
    Zap: (props: any) => <svg data-testid="zap-icon" {...props} />,
    Loader2: (props: any) => <svg data-testid="loader-icon" {...props} />,
    Rocket: (props: any) => <svg data-testid="rocket-icon" {...props} />,
    XCircle: (props: any) => <svg data-testid="xcircle-icon" {...props} />,
}));

import CircuitBoard from '@/components/workspace/CircuitBoard';
import { useCortexStore } from '@/store/useCortexStore';

describe('CircuitBoard', () => {
    beforeEach(() => {
        useCortexStore.setState({
            nodes: [],
            edges: [],
            blueprint: null,
            missionStatus: 'idle',
            activeMissionId: null,
            isCommitting: false,
            selectedAgentNodeId: null,
            isAgentEditorOpen: false,
        });
    });

    it('mounts with ReactFlow mock', () => {
        render(<CircuitBoard />);
        expect(screen.getByTestId('react-flow')).toBeDefined();
    });

    it('shows empty state overlay when no nodes exist', () => {
        render(<CircuitBoard />);
        expect(screen.getByText('Awaiting blueprint')).toBeDefined();
        expect(screen.getByText('Negotiate an intent to generate a team DAG')).toBeDefined();
    });

    it('hides empty state overlay when nodes are present in Zustand store', () => {
        useCortexStore.setState({
            nodes: [
                {
                    id: 'team-0',
                    type: 'group',
                    position: { x: 80, y: 40 },
                    data: { label: '' },
                    style: { width: 280, height: 250 },
                },
                {
                    id: 'agent-0-0',
                    type: 'agentNode',
                    position: { x: 60, y: 80 },
                    parentNode: 'team-0',
                    data: { label: 'researcher', role: 'cognitive', status: 'offline' },
                },
            ],
            edges: [],
        });

        render(<CircuitBoard />);
        expect(screen.getByTestId('react-flow')).toBeDefined();
        expect(screen.queryByText('Awaiting blueprint')).toBeNull();
    });
});
