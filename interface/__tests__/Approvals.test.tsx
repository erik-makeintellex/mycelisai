import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { mockFetch } from './setup';

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
    ShieldCheck: (props: any) => <svg data-testid="shield-check" {...props} />,
    ScrollText: (props: any) => <svg data-testid="scroll-text" {...props} />,
    Settings: (props: any) => <svg data-testid="settings" {...props} />,
    Plus: (props: any) => <svg data-testid="plus" {...props} />,
    Trash2: (props: any) => <svg data-testid="trash2" {...props} />,
    Save: (props: any) => <svg data-testid="save" {...props} />,
    ChevronDown: (props: any) => <svg data-testid="chev-down" {...props} />,
    ChevronRight: (props: any) => <svg data-testid="chev-right" {...props} />,
    Loader2: (props: any) => <svg data-testid="loader" {...props} />,
    Activity: (props: any) => <svg data-testid="activity" {...props} />,
}));

// Mock TrustSlider (makes API calls)
vi.mock('@/components/workspace/TrustSlider', () => ({
    default: () => <div data-testid="trust-slider">TrustSlider</div>,
}));

// Mock DecisionCard
vi.mock('@/components/approvals/DecisionCard', () => ({
    DecisionCard: ({ approval, onResolve }: any) => (
        <div data-testid={`decision-card-${approval.id}`}>
            <span>{approval.id}</span>
            <button onClick={() => onResolve(approval.id, true)}>Approve</button>
            <button onClick={() => onResolve(approval.id, false)}>Reject</button>
        </div>
    ),
}));

import ApprovalsPage from '../app/approvals/page';
import { useCortexStore } from '@/store/useCortexStore';

describe('Approvals Page (Governance)', () => {
    beforeEach(() => {
        // Default: empty approvals from API
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ([]),
        });

        useCortexStore.setState({
            pendingApprovals: [],
            isFetchingApprovals: false,
            policyConfig: null,
            isFetchingPolicy: false,
        });
    });

    it('renders Governance heading and tabs', async () => {
        await act(async () => {
            render(<ApprovalsPage />);
        });

        expect(screen.getByText('Governance')).toBeDefined();
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

    it('renders decision cards when approvals exist', async () => {
        const approvals = [
            { id: 'req-1', intent: 'file.write', agent: 'coder' },
            { id: 'req-2', intent: 'k8s.deploy', agent: 'deployer' },
        ];

        // The useEffect calls fetchPendingApprovals which fetches from API
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => approvals,
        });

        await act(async () => {
            render(<ApprovalsPage />);
        });

        expect(screen.getByText('2 pending requests')).toBeDefined();
        expect(screen.getByTestId('decision-card-req-1')).toBeDefined();
        expect(screen.getByTestId('decision-card-req-2')).toBeDefined();
    });

    it('switches to policy tab', async () => {
        // First fetch: pending approvals (Approvals tab). Second fetch: policy config (Policy tab)
        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ([]),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    groups: [],
                    defaults: { default_action: 'DENY' },
                }),
            });

        await act(async () => {
            render(<ApprovalsPage />);
        });

        const policyTab = screen.getByText('Policy Configuration');
        await act(async () => {
            fireEvent.click(policyTab);
        });

        expect(screen.getByText('Default Action')).toBeDefined();
    });
});
