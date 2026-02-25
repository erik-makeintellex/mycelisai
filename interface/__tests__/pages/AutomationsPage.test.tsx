import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
const mockSearchParams = new URLSearchParams();
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/automations',
    useSearchParams: () => mockSearchParams,
}));

// Mock next/dynamic — return a sync wrapper that renders the loading fallback
// then the resolved component. Since vi.mock is hoisted, child mocks are ready.
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any, opts?: any) => {
        // Store for lazy resolution
        let Comp: any = null;
        const p = loader().then((mod: any) => { Comp = mod.default || mod; }).catch(() => {});
        const Dynamic = (props: any) => {
            if (Comp) return <Comp {...props} />;
            return opts?.loading ? opts.loading() : null;
        };
        // Attach the resolution promise so tests can await it
        (Dynamic as any).__resolvePromise = p;
        return Dynamic;
    },
}));

// Mock heavy child components
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
vi.mock('@/components/automations/TriggerRulesTab', () => ({
    __esModule: true,
    default: () => <div data-testid="triggers-tab">TriggerRulesTab</div>,
}));
vi.mock('@/components/shared/DegradedState', () => ({
    __esModule: true,
    default: ({ title }: any) => <div data-testid="degraded-state">{title}</div>,
}));

// Mock Zustand store
const mockAdvancedMode = vi.fn(() => false);
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = { advancedMode: mockAdvancedMode() };
        return selector(state);
    },
}));

import AutomationsPage from '@/app/(app)/automations/page';

describe('Automations Page (V7)', () => {
    beforeEach(() => {
        mockAdvancedMode.mockReturnValue(false);
        // Reset search params
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
    });

    it('renders page title', async () => {
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.getByText('Automations')).toBeDefined();
    });

    it('renders all standard tabs', async () => {
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.getByText('Active Automations')).toBeDefined();
        expect(screen.getByText('Draft Blueprints')).toBeDefined();
        expect(screen.getByText('Trigger Rules')).toBeDefined();
        expect(screen.getByText('Approvals')).toBeDefined();
        expect(screen.getByText('Teams')).toBeDefined();
    });

    it('hides Neural Wiring tab when advancedMode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.queryByText('Neural Wiring')).toBeNull();
    });

    it('shows Neural Wiring tab when advancedMode is on', async () => {
        mockAdvancedMode.mockReturnValue(true);
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.getByText('Neural Wiring')).toBeDefined();
    });

    it('defaults to Active Automations tab', async () => {
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.getByText('Scheduled Missions')).toBeDefined();
    });

    it('deep-links to approvals tab via search param', async () => {
        mockSearchParams.set('tab', 'approvals');
        // Allow dynamic import promises to resolve
        await new Promise((r) => setTimeout(r, 10));
        await act(async () => { render(<AutomationsPage />); });
        // If dynamic hasn't resolved yet, flush again
        await act(async () => { await new Promise((r) => setTimeout(r, 10)); });
        expect(screen.getByTestId('approvals-tab')).toBeDefined();
    });
});
