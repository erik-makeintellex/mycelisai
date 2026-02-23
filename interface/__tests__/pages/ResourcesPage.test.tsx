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
    default: (loader: any, _opts?: any) => {
        return (props: any) => {
            const React = require('react');
            const [Comp, setComp] = React.useState<any>(null);
            React.useEffect(() => {
                loader().then((mod: any) => setComp(() => mod.default || mod));
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

import ResourcesPage from '@/app/(app)/resources/page';

describe('Resources Page (V7)', () => {
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
        expect(screen.getByText('Brains')).toBeDefined();
        expect(screen.getByText('MCP Tools')).toBeDefined();
        expect(screen.getByText('Workspace Explorer')).toBeDefined();
        expect(screen.getByText('Capabilities')).toBeDefined();
    });

    it('defaults to Brains tab', async () => {
        await act(async () => { render(<ResourcesPage />); });
        expect(screen.getByTestId('brains-page')).toBeDefined();
    });

    it('deep-links to catalogue tab via search param', async () => {
        mockSearchParams.set('tab', 'catalogue');
        await act(async () => { render(<ResourcesPage />); });
        await waitFor(() => {
            expect(screen.getByTestId('catalogue-page')).toBeDefined();
        });
    });
});
