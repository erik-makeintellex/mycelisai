import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
const mockSearchParams = new URLSearchParams();
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/dashboard',
    useSearchParams: () => mockSearchParams,
}));

// Mock next/dynamic — eagerly resolve the loader for testing
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any, _opts?: any) => {
        let Resolved: any = null;
        loader().then((mod: any) => { Resolved = mod.default || mod; });
        return (props: any) => {
            const React = require('react');
            return Resolved ? React.createElement(Resolved, props) : null;
        };
    },
}));

// Mock heavy child components to avoid import side-effects
vi.mock('@/components/workspace/Workspace', () => ({
    __esModule: true,
    default: () => <div data-testid="workspace">Workspace</div>,
}));
vi.mock('@/components/teams/TeamsPage', () => ({
    __esModule: true,
    default: () => <div data-testid="teams-page">TeamsPage</div>,
}));
vi.mock('@/components/automations/ApprovalsTab', () => ({
    __esModule: true,
    default: () => <div data-testid="approvals-tab">ApprovalsTab</div>,
}));
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
vi.mock('@/components/matrix/MatrixGrid', () => ({
    __esModule: true,
    default: () => <div data-testid="matrix-grid">MatrixGrid</div>,
}));

// Mock Zustand store
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = { advancedMode: false };
        return selector(state);
    },
}));

import AutomationsPage from '@/app/(app)/automations/page';
import ResourcesPage from '@/app/(app)/resources/page';
import SystemPage from '@/app/(app)/system/page';

describe('no_console_error_on_primary_routes', () => {
    let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
    let consoleWarnSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        // Mock fetch for health checks
        vi.spyOn(global, 'fetch').mockResolvedValue({
            ok: true,
            json: async () => ({ goroutines: 42, heap_alloc_mb: 18.5, sys_mem_mb: 52.0, llm_tokens_sec: 3.2, timestamp: new Date().toISOString() }),
        } as Response);

        consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        consoleErrorSpy.mockRestore();
        consoleWarnSpy.mockRestore();
    });

    it('/automations renders without console.error', async () => {
        await act(async () => { render(<AutomationsPage />); });
        const errors = consoleErrorSpy.mock.calls.filter(
            (call) => !String(call[0]).includes('act()')
        );
        expect(errors).toHaveLength(0);
    });

    it('/resources renders without console.error', async () => {
        await act(async () => { render(<ResourcesPage />); });
        const errors = consoleErrorSpy.mock.calls.filter(
            (call) => !String(call[0]).includes('act()')
        );
        expect(errors).toHaveLength(0);
    });

    it('/system renders without console.error', async () => {
        await act(async () => { render(<SystemPage />); });
        const errors = consoleErrorSpy.mock.calls.filter(
            (call) => !String(call[0]).includes('act()')
        );
        expect(errors).toHaveLength(0);
    });

    it('/automations renders with no bg-white in output', async () => {
        const { container } = await act(async () => render(<AutomationsPage />));
        expect(container.innerHTML).not.toContain('bg-white');
    });

    it('/resources renders with no bg-white in output', async () => {
        const { container } = await act(async () => render(<ResourcesPage />));
        expect(container.innerHTML).not.toContain('bg-white');
    });

    it('/system renders with no bg-white in output', async () => {
        const { container } = await act(async () => render(<SystemPage />));
        expect(container.innerHTML).not.toContain('bg-white');
    });
});
