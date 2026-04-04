import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';

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
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) =>
        selector({
            advancedMode: mockAdvancedMode(),
            mcpServers: [],
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
            deleteMCPServer: mockDeleteMCPServer,
        }),
}));

import ResourcesPage from '@/app/(app)/resources/page';

describe('Resources Page (V8.1 advanced support)', () => {
    beforeEach(() => {
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        mockAdvancedMode.mockReturnValue(true);
    });

    it('renders page title', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText('Resources')).toBeDefined();
    });

    it('renders all tabs', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText('Connected Tools')).toBeDefined();
        expect(screen.getByText('Exchange')).toBeDefined();
        expect(screen.getByText('Deployment Context')).toBeDefined();
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
            expect(screen.getByRole('button', { name: 'Deployment Context' }).className).toContain('border-cortex-primary');
        });
    });

    it('shows the advanced gate when advanced mode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText(/Advanced resources stay tucked away by default/i)).toBeDefined();
    });
});
