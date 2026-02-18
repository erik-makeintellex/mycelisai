import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock child components that have their own complex rendering
vi.mock('@/components/approvals/DecisionCard', () => ({
    __esModule: true,
    DecisionCard: ({ approval }: any) => (
        <div data-testid="decision-card">{approval?.id ?? 'card'}</div>
    ),
}));

vi.mock('@/components/workspace/TrustSlider', () => ({
    __esModule: true,
    default: () => <div data-testid="trust-slider">TrustSlider</div>,
}));

import ApprovalsPage from '@/app/approvals/page';
import { useCortexStore } from '@/store/useCortexStore';

describe('Approvals Page (app/approvals/page.tsx)', () => {
    beforeEach(() => {
        useCortexStore.setState({
            pendingApprovals: [],
            isFetchingApprovals: false,
            policyConfig: null,
            isFetchingPolicy: false,
        });

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ proposals: [], config: { groups: [], defaults: { default_action: 'DENY' } } }),
        });
    });

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<ApprovalsPage />);
        });

        expect(screen.getByText('Governance')).toBeDefined();
    });

    it('renders tab navigation', async () => {
        await act(async () => {
            render(<ApprovalsPage />);
        });

        expect(screen.getByText('Approvals Queue')).toBeDefined();
        expect(screen.getByText('Policy Configuration')).toBeDefined();
    });

    it('shows empty state when no pending approvals', async () => {
        await act(async () => {
            render(<ApprovalsPage />);
        });

        expect(screen.getByText('All Clear')).toBeDefined();
        expect(screen.getByText('No pending governance requests.')).toBeDefined();
    });
});
