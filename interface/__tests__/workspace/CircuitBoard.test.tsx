import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { ReactNode, SVGProps } from 'react';

type ChildrenProps = {
    children?: ReactNode;
};

type FlowPoint = {
    x: number;
    y: number;
};

// Mock reactflow with factory (must return default export properly)
vi.mock('reactflow', async () => {
    const React = await import('react');
    const ReactFlow = React.forwardRef<HTMLDivElement, ChildrenProps>(({ children }, ref) =>
        React.createElement('div', { 'data-testid': 'react-flow', ref }, children)
    );
    ReactFlow.displayName = 'ReactFlow';
    return {
        __esModule: true,
        default: ReactFlow,
        ReactFlow,
        ReactFlowProvider: ({ children }: ChildrenProps) =>
            React.createElement('div', { 'data-testid': 'react-flow-provider' }, children),
        Background: () => React.createElement('div', { 'data-testid': 'react-flow-background' }),
        BackgroundVariant: { Dots: 'dots', Lines: 'lines', Cross: 'cross' },
        Controls: () => React.createElement('div', { 'data-testid': 'react-flow-controls' }),
        MiniMap: () => React.createElement('div', { 'data-testid': 'react-flow-minimap' }),
        Handle: () => React.createElement('div', { 'data-testid': 'react-flow-handle' }),
        Panel: ({ children }: ChildrenProps) => React.createElement('div', { 'data-testid': 'react-flow-panel' }, children),
        Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
        MarkerType: { Arrow: 'arrow', ArrowClosed: 'arrowclosed' },
        useNodesState: (initialNodes: unknown[] = []) => {
            const [nodes, setNodes] = React.useState<unknown[]>(initialNodes);
            return [nodes, setNodes, vi.fn()];
        },
        useEdgesState: (initialEdges: unknown[] = []) => {
            const [edges, setEdges] = React.useState<unknown[]>(initialEdges);
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
            project: vi.fn((position: FlowPoint) => position),
        }),
        addEdge: vi.fn((edge: unknown, edges: unknown[]) => [...edges, edge]),
        applyNodeChanges: vi.fn((_changes: unknown[], nodes: unknown[]) => nodes),
        applyEdgeChanges: vi.fn((_changes: unknown[], edges: unknown[]) => edges),
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
    Zap: (props: SVGProps<SVGSVGElement>) => <svg data-testid="zap-icon" {...props} />,
    Loader2: (props: SVGProps<SVGSVGElement>) => <svg data-testid="loader-icon" {...props} />,
    Rocket: (props: SVGProps<SVGSVGElement>) => <svg data-testid="rocket-icon" {...props} />,
    XCircle: (props: SVGProps<SVGSVGElement>) => <svg data-testid="xcircle-icon" {...props} />,
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
