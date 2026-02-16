import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import ActiveMissionsBar from '@/components/dashboard/ActiveMissionsBar';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

const MISSIONS = [
    { id: 'mission-alpha', intent: 'Scan papers', status: 'active', teams: 2, agents: 5 },
    { id: 'mission-beta', intent: 'Deploy report', status: 'completed', teams: 1, agents: 3 },
];

describe('ActiveMissionsBar', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        // Reset store state
        useCortexStore.setState({
            missions: [],
            isFetchingMissions: false,
            activeMissionId: null,
            missionStatus: 'idle',
            blueprint: null,
        });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders the missions bar container', () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });
        render(<ActiveMissionsBar />);
        expect(screen.getByTestId('active-missions-bar')).toBeDefined();
    });

    it('shows "No active missions" when API returns empty', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(screen.getByText('No active missions')).toBeDefined();
        });
    });

    it('renders mission chips when API returns missions', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => MISSIONS,
        });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(screen.getByText('mission-alpha')).toBeDefined();
            expect(screen.getByText('mission-beta')).toBeDefined();
        });
    });

    it('shows LIVE count badge for active missions', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => MISSIONS,
        });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(screen.getByText('1 LIVE')).toBeDefined();
        });
    });

    it('shows team/agent counts on mission chips', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => MISSIONS,
        });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(screen.getByText('2T/5A')).toBeDefined(); // mission-alpha
            expect(screen.getByText('1T/3A')).toBeDefined(); // mission-beta
        });
    });

    it('calls fetchMissions on mount', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/missions');
        });
    });

    it('polls for missions every 15 seconds', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<ActiveMissionsBar />);

        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        const initial = mockFetch.mock.calls.filter(c => c[0] === '/api/v1/missions').length;

        await act(async () => { await vi.advanceTimersByTimeAsync(15000); });
        const afterPoll = mockFetch.mock.calls.filter(c => c[0] === '/api/v1/missions').length;

        expect(afterPoll).toBeGreaterThan(initial);
    });

    it('shows Loading... while fetching with no existing data', async () => {
        // Start with fetching = true, empty missions
        useCortexStore.setState({ isFetchingMissions: true, missions: [] });
        mockFetch.mockReturnValue(new Promise(() => {})); // never resolves

        render(<ActiveMissionsBar />);

        expect(screen.getByText('Loading...')).toBeDefined();
    });

    it('merges live in-memory mission with fetched missions', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => [MISSIONS[1]], // only completed mission from API
        });

        // Set a live mission in store
        useCortexStore.setState({
            activeMissionId: 'mission-live',
            missionStatus: 'active',
            blueprint: {
                mission_id: 'mission-live',
                intent: 'Live mission',
                teams: [
                    { name: 'Team A', role: 'scan', agents: [{ id: 'a1', role: 'cognitive' }] },
                ],
            },
        });

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            // Should show both: live mission + completed from API
            expect(screen.getByText('mission-live')).toBeDefined();
            expect(screen.getByText('mission-beta')).toBeDefined();
        });
    });

    it('handles API failure gracefully', async () => {
        mockFetch.mockRejectedValue(new Error('Network error'));

        render(<ActiveMissionsBar />);

        await waitFor(() => {
            expect(screen.getByText('No active missions')).toBeDefined();
        });
    });
});
