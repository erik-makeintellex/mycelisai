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
        await act(async () => { render(<AutomationsPage />); });
        expect(screen.getByTestId('approvals-tab')).toBeDefined();
    });
});
