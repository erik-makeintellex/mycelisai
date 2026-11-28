import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, waitFor, fireEvent } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
const mockSearchParams = new URLSearchParams();
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/resources',
    useSearchParams: () => mockSearchParams,
}));

// Mock next/dynamic to render the component directly
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any) => {
        const Component = require('react').lazy(loader);
        return (props: any) => {
            const React = require('react');
            return React.createElement(
                React.Suspense,
                { fallback: null },
                React.createElement(Component, props),
            );
        };
    },
}));

// Mock child components
vi.mock('@/components/settings/BrainsPage', () => ({
    __esModule: true,
    default: () => <div data-testid="brains-page">BrainsPage</div>,
}));
vi.mock('@/components/settings/MCPToolRegistry', () => ({
    __esModule: true,
    default: () => <div data-testid="mcp-tools">MCPToolRegistry</div>,
}));
vi.mock('@/components/resources/ExchangeInspector', () => ({
    __esModule: true,
    default: () => <div data-testid="exchange-inspector">ExchangeInspector</div>,
}));
vi.mock('@/components/resources/DeploymentContextPanel', () => ({
    __esModule: true,
    default: () => <div data-testid="deployment-context-panel">DeploymentContextPanel</div>,
}));
vi.mock('@/components/catalogue/CataloguePage', () => ({
    __esModule: true,
    default: () => <div data-testid="catalogue-page">CataloguePage</div>,
}));

const mockAdvancedMode = vi.fn(() => true);
const mockFetchMCPServers = vi.fn();
const mockDeleteMCPServer = vi.fn();
const mockFetchArtifacts = vi.fn().mockResolvedValue(undefined);
const mockToggleAdvancedMode = vi.fn();
let mockArtifacts: Record<string, unknown>[] = [];

vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) =>
        selector({
            advancedMode: mockAdvancedMode(),
            assistantName: 'Soma',
            mcpServers: [],
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
            deleteMCPServer: mockDeleteMCPServer,
            isFetchingArtifacts: false,
            fetchArtifacts: mockFetchArtifacts,
            toggleAdvancedMode: mockToggleAdvancedMode,
            missionChat: [],
            artifacts: mockArtifacts,
        }),
}));

import ResourcesPage from '@/app/(app)/resources/page';

describe('Resources Page (V8.1 advanced support & Operator vs Builder View)', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        mockAdvancedMode.mockReturnValue(true);
        mockArtifacts = [
            {
                id: 'art-1',
                agent_id: 'admin',
                artifact_type: 'document',
                title: 'Q3 Launch Plan',
                content_type: 'text/markdown',
                content: '# Q3 Launch Plan\nReady for distribution.',
                file_path: 'workspace/deliverables/q3-launch-plan.md',
                metadata: {},
                status: 'approved',
                created_at: new Date().toISOString(),
            },
        ];
    });

    it('renders page title', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText('Advanced Resources')).toBeDefined();
    });

    it('renders all tabs', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText('Connected Tools')).toBeDefined();
        expect(screen.getByText('Exchange')).toBeDefined();
        expect(screen.getByText('Knowledge')).toBeDefined();
        expect(screen.getByText('Workspace Files')).toBeDefined();
        expect(screen.getByText('AI Engines')).toBeDefined();
        expect(screen.getByText('Role Library')).toBeDefined();
    });

    it('defaults to connected tools tab', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(await screen.findByTestId('mcp-tools', {}, { timeout: 5000 })).toBeDefined();
    });

    it('deep-links to role library tab via search param', async () => {
        mockSearchParams.set('tab', 'roles');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'Role Library' }).className).toContain('border-cortex-primary');
        });
    });

    it('deep-links to exchange tab via search param', async () => {
        mockSearchParams.set('tab', 'exchange');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'Exchange' }).className).toContain('border-cortex-primary');
        });
    });

    it('deep-links to deployment context tab via search param', async () => {
        mockSearchParams.set('tab', 'deployment-context');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'Knowledge' }).className).toContain('border-cortex-primary');
        });
    });

    it('shows the Durable Files & Deliverables operator view when advanced mode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByTestId('durable-files-operator-view')).toBeDefined();
        expect(screen.getByText('Q3 Launch Plan')).toBeDefined();
        expect(screen.getByText('workspace/deliverables/q3-launch-plan.md')).toBeDefined();
        expect(screen.getByText('Produced by Soma')).toBeDefined();

        const previewBtn = screen.getByRole('button', { name: /Preview/i });
        fireEvent.click(previewBtn);
        expect(screen.getByText(/Ready for distribution/i)).toBeDefined();

        const advancedBtn = screen.getByTestId('resources-switch-advanced');
        expect(advancedBtn.textContent).toContain('Connected Tools & AI Engines');
        fireEvent.click(advancedBtn);
        expect(mockToggleAdvancedMode).toHaveBeenCalledTimes(1);
    });
});
