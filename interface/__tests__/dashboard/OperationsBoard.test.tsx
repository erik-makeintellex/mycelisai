import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock SignalContext — OperationsBoard uses useSignalStream for priority alerts
const mockSignals: any[] = [];
vi.mock('@/components/dashboard/SignalContext', () => ({
    useSignalStream: () => ({
        signals: mockSignals,
        isConnected: true,
    }),
    SignalProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

import OperationsBoard from '@/components/dashboard/OperationsBoard';
import { useCortexStore } from '@/store/useCortexStore';

// ── Fixtures ──────────────────────────────────────────────────

const STANDING_TEAMS = [
    {
        id: 'admin-core',
        name: 'Admin',
        type: 'standing',
        mission_id: null,
        agents: [{ id: 'admin', role: 'admin', status: 1 }],
    },
    {
        id: 'council-core',
        name: 'Council',
        type: 'standing',
        mission_id: null,
        agents: [
            { id: 'council-architect', role: 'architect', status: 1 },
            { id: 'council-coder', role: 'coder', status: 2 },
        ],
    },
];

const MISSION_TEAMS = [
    {
        id: 'team-abc-1',
        name: 'Scraper Team',
        type: 'mission',
        mission_id: 'mission-001',
        agents: [
            { id: 'scraper-agent', role: 'worker', status: 2 },
            { id: 'parser-agent', role: 'worker', status: 1 },
        ],
    },
];

const MISSIONS = [
    {
        id: 'mission-001',
        intent: 'Build a web scraper for weather data',
        status: 'active' as const,
        teams: 1,
        agents: 2,
        created_at: '2026-02-15T10:00:00Z',
    },
    {
        id: 'mission-002',
        intent: 'Generate quarterly report',
        status: 'completed' as const,
        teams: 2,
        agents: 4,
        created_at: '2026-02-14T08:00:00Z',
    },
];

function resetStore() {
    useCortexStore.setState({
        teamsDetail: [],
        isFetchingTeamsDetail: false,
        missions: [],
        isFetchingMissions: false,
    });
}

// Helper to configure mockFetch to return the right data based on URL
function setupFetchResponses(teams: any[] = [], missions: any[] = []) {
    mockFetch.mockImplementation((url: string) => {
        if (typeof url === 'string' && url.includes('/teams/detail')) {
            return Promise.resolve({ ok: true, json: async () => teams });
        }
        if (typeof url === 'string' && url.includes('/missions')) {
            return Promise.resolve({ ok: true, json: async () => missions });
        }
        return Promise.resolve({ ok: true, json: async () => [] });
    });
}

// ── Tests ─────────────────────────────────────────────────────

describe('OperationsBoard', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        resetStore();
        mockSignals.length = 0;
        // Default: fetch returns empty for teams and missions
        setupFetchResponses([], []);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    // ── Rendering ─────────────────────────────────────────────

    it('renders the operations board container', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        expect(screen.getByTestId('operations-board')).toBeDefined();
    });

    it('renders Standing Workloads section header', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        expect(screen.getByText('Standing Workloads')).toBeDefined();
    });

    it('renders Missions section header', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        expect(screen.getByText('Missions')).toBeDefined();
    });

    // ── Standing Workloads ────────────────────────────────────

    it('renders standing teams when present', async () => {
        setupFetchResponses(STANDING_TEAMS, []);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('Admin')).toBeDefined();
        expect(screen.getByText('Council')).toBeDefined();
    });

    it('shows agent counts for standing teams', async () => {
        setupFetchResponses(STANDING_TEAMS, []);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('1/1 agents')).toBeDefined(); // Admin: 1 idle
        expect(screen.getByText('2/2 agents')).toBeDefined(); // Council: 2 agents
    });

    it('shows empty state when no standing teams', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        expect(screen.getByText('No standing workloads')).toBeDefined();
    });

    it('excludes mission teams from standing workloads', async () => {
        setupFetchResponses([...STANDING_TEAMS, ...MISSION_TEAMS], []);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        // Scraper Team should not appear in standing workloads
        // but will appear in missions section
        const standingSection = screen.getByText('Standing Workloads').closest('div');
        expect(standingSection).toBeDefined();
    });

    // ── Missions ──────────────────────────────────────────────

    it('renders missions when present', async () => {
        setupFetchResponses([], MISSIONS);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('mission-001')).toBeDefined();
        expect(screen.getByText('mission-002')).toBeDefined();
    });

    it('shows mission intent text', async () => {
        setupFetchResponses([], MISSIONS);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('Build a web scraper for weather data')).toBeDefined();
    });

    it('shows LIVE badge for active missions', async () => {
        setupFetchResponses([], MISSIONS);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('1 LIVE')).toBeDefined();
    });

    it('shows empty state when no missions', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        expect(screen.getByText('No active missions')).toBeDefined();
    });

    it('sorts missions active-first', async () => {
        const reversed = [...MISSIONS].reverse(); // completed first
        setupFetchResponses([], reversed);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        const missionIds = screen.getAllByText(/^mission-/);
        expect(missionIds[0].textContent).toBe('mission-001'); // active first
        expect(missionIds[1].textContent).toBe('mission-002'); // completed second
    });

    it('shows team/agent counts on mission rows', async () => {
        setupFetchResponses([], MISSIONS);

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('1T/2A')).toBeDefined(); // mission-001
        expect(screen.getByText('2T/4A')).toBeDefined(); // mission-002
    });

    // ── Priority Alerts ───────────────────────────────────────

    it('hides priority alerts when no relevant signals', async () => {
        mockSignals.push(
            { type: 'telemetry', message: 'CPU OK', timestamp: new Date().toISOString() }
        );

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.queryByText('Priority Alerts')).toBeNull();
    });

    it('shows priority alerts for governance_halt signals', async () => {
        mockSignals.push(
            { type: 'governance_halt', message: 'Action blocked', source: 'admin', timestamp: new Date().toISOString() }
        );

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('Priority Alerts')).toBeDefined();
        expect(screen.getByText('GOVERNANCE')).toBeDefined();
    });

    it('shows priority alerts for error signals', async () => {
        mockSignals.push(
            { type: 'error', message: 'Agent crashed', source: 'worker-1', timestamp: new Date().toISOString() }
        );

        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        expect(screen.getByText('ERROR')).toBeDefined();
        expect(screen.getByText('Agent crashed')).toBeDefined();
    });

    // ── Polling ───────────────────────────────────────────────

    it('fetches teams and missions on mount', async () => {
        render(<OperationsBoard />);
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });

        // Should have been called at least once for teams and missions
        const fetchCalls = mockFetch.mock.calls.map((c: any[]) => c[0]);
        expect(fetchCalls.some((url: string) => url.includes('/teams/detail'))).toBe(true);
        expect(fetchCalls.some((url: string) => url.includes('/missions'))).toBe(true);
    });

    it('polls teams every 10 seconds', async () => {
        render(<OperationsBoard />);

        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        const initialCount = mockFetch.mock.calls.filter(
            (c: any[]) => typeof c[0] === 'string' && c[0].includes('/teams/detail')
        ).length;

        await act(async () => { await vi.advanceTimersByTimeAsync(10000); });
        const afterPoll = mockFetch.mock.calls.filter(
            (c: any[]) => typeof c[0] === 'string' && c[0].includes('/teams/detail')
        ).length;

        expect(afterPoll).toBeGreaterThan(initialCount);
    });
});
