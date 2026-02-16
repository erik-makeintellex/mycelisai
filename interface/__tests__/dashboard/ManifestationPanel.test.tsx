import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act, fireEvent } from '@testing-library/react';
import ManifestationPanel from '@/components/dashboard/ManifestationPanel';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

const PROPOSALS_RESPONSE = {
    proposals: [
        {
            id: 'prop-1',
            name: 'Signal Analytics Squad',
            role: 'analytics',
            agents: [
                { id: 'a1', role: 'cognitive', system_prompt: 'Analyze signals' },
                { id: 'a2', role: 'actuation', system_prompt: 'Write reports' },
            ],
            reason: 'Detected pattern in incoming sensor data that requires analysis.',
            status: 'pending',
            created_at: new Date().toISOString(),
        },
        {
            id: 'prop-2',
            name: 'Knowledge Curator Team',
            role: 'curation',
            agents: [{ id: 'a3', role: 'cognitive' }],
            reason: 'Memory store growing; need curation agent.',
            status: 'pending',
            created_at: new Date().toISOString(),
        },
    ],
};

describe('ManifestationPanel', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        useCortexStore.setState({
            teamProposals: [],
            isFetchingProposals: false,
        });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders manifestation panel', () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ proposals: [] }) });
        render(<ManifestationPanel />);
        expect(screen.getByTestId('manifestation-panel')).toBeDefined();
    });

    it('shows empty state when no proposals', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ proposals: [] }) });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('No team proposals')).toBeDefined();
        });
    });

    it('renders proposal cards from API data', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('Signal Analytics Squad')).toBeDefined();
            expect(screen.getByText('Knowledge Curator Team')).toBeDefined();
        });
    });

    it('shows pending count badge', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('2 PENDING')).toBeDefined();
        });
    });

    it('shows agent count on each proposal card', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('2 agents proposed')).toBeDefined();
            expect(screen.getByText('1 agent proposed')).toBeDefined();
        });
    });

    it('shows MANIFEST and DISMISS buttons for pending proposals', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            const manifestButtons = screen.getAllByText('MANIFEST');
            const dismissButtons = screen.getAllByText('DISMISS');
            expect(manifestButtons).toHaveLength(2);
            expect(dismissButtons).toHaveLength(2);
        });
    });

    it('calls approveProposal when MANIFEST is clicked', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('Signal Analytics Squad')).toBeDefined();
        });

        // Mock the approve endpoint
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

        act(() => {
            fireEvent.click(screen.getAllByText('MANIFEST')[0]);
        });

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v1/proposals/prop-1/approve',
                { method: 'POST' }
            );
        });
    });

    it('calls rejectProposal when DISMISS is clicked', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('Signal Analytics Squad')).toBeDefined();
        });

        mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

        act(() => {
            fireEvent.click(screen.getAllByText('DISMISS')[0]);
        });

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v1/proposals/prop-1/reject',
                { method: 'POST' }
            );
        });
    });

    it('fetches proposals on mount', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/proposals');
        });
    });

    it('polls proposals every 10 seconds', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => PROPOSALS_RESPONSE });

        render(<ManifestationPanel />);

        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        const initial = mockFetch.mock.calls.filter(c => c[0] === '/api/v1/proposals').length;

        await act(async () => { await vi.advanceTimersByTimeAsync(10000); });
        const afterPoll = mockFetch.mock.calls.filter(c => c[0] === '/api/v1/proposals').length;

        expect(afterPoll).toBeGreaterThan(initial);
    });

    it('hides action buttons for non-pending proposals', async () => {
        const approved = {
            proposals: [{
                ...PROPOSALS_RESPONSE.proposals[0],
                status: 'approved',
            }],
        };
        mockFetch.mockResolvedValue({ ok: true, json: async () => approved });

        render(<ManifestationPanel />);

        await waitFor(() => {
            expect(screen.getByText('Signal Analytics Squad')).toBeDefined();
            expect(screen.getByText('APPROVED')).toBeDefined();
        });

        expect(screen.queryByText('MANIFEST')).toBeNull();
        expect(screen.queryByText('DISMISS')).toBeNull();
    });
});
