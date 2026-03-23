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

// Mock next/dynamic — resolve synchronously using React.useState + useEffect
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: () => Promise<any>, _opts?: any) => {
        return (props: Record<string, unknown>) => {
            const React = require('react') as typeof import('react');
            const [Comp, setComp] = React.useState<import('react').ComponentType<any> | null>(null);
            React.useEffect(() => {
                let mounted = true;
                loader().then((mod: any) => {
                    if (!mounted) {
                        return;
                    }
                    setComp(() => (mod.default || mod) as import('react').ComponentType<any>);
                });
                return () => {
                    mounted = false;
                };
            }, []);
            return Comp ? React.createElement(Comp, props) : null;
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
vi.mock('@/components/catalogue/CataloguePage', () => ({
    __esModule: true,
    default: () => <div data-testid="catalogue-page">CataloguePage</div>,
}));

const mockAdvancedMode = vi.fn(() => true);
const mockFetchMCPServers = vi.fn();
const mockInstallMCPServer = vi.fn();
const mockDeleteMCPServer = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) =>
        selector({
            advancedMode: mockAdvancedMode(),
            mcpServers: [],
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
            installMCPServer: mockInstallMCPServer,
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
        expect(screen.getByText('Workspace Files')).toBeDefined();
        expect(screen.getByText('AI Engines')).toBeDefined();
        expect(screen.getByText('Role Library')).toBeDefined();
    });

    it('defaults to connected tools tab', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(await screen.findByText('MCP Tool Registry', {}, { timeout: 5000 })).toBeDefined();
    });

    it('deep-links to role library tab via search param', async () => {
        mockSearchParams.set('tab', 'roles');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByTestId('catalogue-page')).toBeDefined();
        });
    });

    it('shows the advanced gate when advanced mode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByText(/Advanced resources stay tucked away by default/i)).toBeDefined();
    });
});
