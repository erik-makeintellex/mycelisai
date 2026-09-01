import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';
import React, { type ComponentType } from 'react';

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
    default: (loader: () => Promise<{ default: ComponentType<Record<string, unknown>> }>) => {
        const Component = React.lazy(loader);
        return (props: Record<string, unknown>) => {
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
vi.mock('@/components/resources/WorkspaceExplorer', () => ({
    __esModule: true,
    default: ({ initialPath }: { initialPath?: string | null }) => (
        <div data-testid="workspace-explorer">WorkspaceExplorer path={initialPath ?? ''}</div>
    ),
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

const mockFetchMCPServers = vi.fn();
const mockDeleteMCPServer = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: (state: {
        mcpServers: never[];
        isFetchingMCPServers: boolean;
        fetchMCPServers: typeof mockFetchMCPServers;
        deleteMCPServer: typeof mockDeleteMCPServer;
    }) => unknown) =>
        selector({
            mcpServers: [],
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
            deleteMCPServer: mockDeleteMCPServer,
        }),
}));

import ResourcesPage from '@/app/(app)/resources/page';

describe('Resources Page (operator support)', () => {
    beforeEach(() => {
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
    });

    it('renders page title', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText('Resources')).toBeDefined();
    });

    it('renders all tabs', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByRole('tablist', { name: 'Resource type menu' })).toBeDefined();
        expect(screen.getByRole('tab', { name: /Capabilities/i })).toBeDefined();
        expect(screen.getByRole('tab', { name: /Exchange/i })).toBeDefined();
        expect(screen.getByRole('tab', { name: /Deployment Context/i })).toBeDefined();
        expect(screen.getByRole('tab', { name: /Deliverables/i })).toBeDefined();
        expect(screen.getByRole('tab', { name: /AI Engines/i })).toBeDefined();
        expect(screen.getByRole('tab', { name: /Worker Profiles/i })).toBeDefined();
    });

    it('defaults to deliverables and keeps setup resources secondary', async () => {
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('tab', { name: /Deliverables/i }).getAttribute('aria-selected')).toBe('true');
        });
        expect(screen.getByText('Results')).toBeDefined();
        expect(screen.getByText('Advanced resources')).toBeDefined();
        expect(screen.getByText(/Open the files, packages, media, and other results Soma delivered for you/i)).toBeDefined();
    });

    it('keeps tab=tools as the capability catalog deep link', async () => {
        mockSearchParams.set('tab', 'tools');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('tab', { name: /Capabilities/i }).getAttribute('aria-selected')).toBe('true');
        });
        expect(screen.getByText(/What Soma can use, what needs repair, and what can be requested/i)).toBeDefined();
    });

    it('deep-links to worker profiles via the compatible roles search param', async () => {
        mockSearchParams.set('tab', 'roles');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('tab', { name: /Worker Profiles/i }).getAttribute('aria-selected')).toBe('true');
        });
    });

    it('deep-links to exchange tab via search param', async () => {
        mockSearchParams.set('tab', 'exchange');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('tab', { name: /Exchange/i }).getAttribute('aria-selected')).toBe('true');
        });
    });

    it('passes workspace path deep links to Deliverables', async () => {
        mockSearchParams.set('tab', 'workspace');
        mockSearchParams.set('path', 'workspace/generated/game');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByTestId('workspace-explorer').textContent).toContain('path=workspace/generated/game');
        });
    });

    it('deep-links to deployment context tab via search param', async () => {
        mockSearchParams.set('tab', 'deployment-context');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByRole('tab', { name: /Deployment Context/i }).getAttribute('aria-selected')).toBe('true');
        });
    });

    it('stays available without admin tools gating', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByRole('tablist', { name: 'Resource type menu' })).toBeDefined();
        expect(screen.queryByText(/Admin tools/i)).toBeNull();
    });

    it('uses refresh-safe resource links', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByRole('tab', { name: /Capabilities/i }).getAttribute('href')).toBe('/resources?tab=tools');
        expect(screen.getByRole('tab', { name: /Deliverables/i }).getAttribute('href')).toBe('/resources');
    });
});
