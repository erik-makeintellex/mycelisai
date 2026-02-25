import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-1',
    useSearchParams: () => new URLSearchParams(),
}));

// Mock labels (EventCard may import it)
vi.mock('@/lib/labels', () => ({
    brainDisplayName: (id: string) => id,
    brainLocationLabel: (l: string) => l,
}));

import RunTimeline from '@/components/runs/RunTimeline';
import type { MissionEvent } from '@/store/useCortexStore';

// ── Test data ────────────────────────────────────────────────

const now = new Date().toISOString();

const mockEvents: MissionEvent[] = [
    {
        id: 'evt-1',
        run_id: 'run-abc',
        tenant_id: 'default',
        event_type: 'mission.started',
        severity: 'info',
        source_agent: 'admin',
        emitted_at: now,
        payload: { mission_id: 'mission-xyz' },
    },
    {
        id: 'evt-2',
        run_id: 'run-abc',
        tenant_id: 'default',
        event_type: 'tool.invoked',
        severity: 'info',
        source_agent: 'admin',
        emitted_at: now,
        payload: { tool: 'consult_council' },
    },
    {
        id: 'evt-3',
        run_id: 'run-abc',
        tenant_id: 'default',
        event_type: 'tool.completed',
        severity: 'info',
        source_agent: 'admin',
        emitted_at: now,
        payload: { tool: 'consult_council' },
    },
];

const terminalEvents: MissionEvent[] = [
    ...mockEvents,
    {
        id: 'evt-4',
        run_id: 'run-abc',
        tenant_id: 'default',
        event_type: 'mission.completed',
        severity: 'info',
        emitted_at: now,
    },
];

// ── Helpers ──────────────────────────────────────────────────

/** Resolve a fetch mock that returns an events list */
function mockFetchOk(events: MissionEvent[]) {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
        ok: true,
        json: async () => ({ data: events }),
    });
}

function mockFetchError(status = 500) {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
        ok: false,
        status,
        json: async () => ({}),
    });
}

// ── Tests ────────────────────────────────────────────────────

describe('RunTimeline', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('shows loading state when fetching timeline', async () => {
        // Fetch never resolves — stays in loading
        (global.fetch as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });

        expect(screen.getByText('Loading events...')).toBeDefined();
    });

    it('shows empty state when no events', async () => {
        mockFetchOk([]);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        // Flush the fetch promise
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(screen.getByText(/No events yet/)).toBeDefined();
        expect(screen.getByText(/Auto-refreshing every 5s/)).toBeDefined();
    });

    it('renders events when data is available', async () => {
        mockFetchOk(mockEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        // EventCard renders event_type as badge text
        expect(screen.getByText('mission.started')).toBeDefined();
        expect(screen.getByText('tool.invoked')).toBeDefined();
        expect(screen.getByText('tool.completed')).toBeDefined();
    });

    it('calls fetch on mount with the runId prop', async () => {
        mockFetchOk([]);

        await act(async () => {
            render(<RunTimeline runId="run-42" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(global.fetch).toHaveBeenCalledWith('/api/v1/runs/run-42/events');
    });

    it('shows auto-refresh indicator when run is not terminal', async () => {
        // Events without a terminal event
        mockFetchOk(mockEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(screen.getByText('auto-refresh')).toBeDefined();
    });

    it('hides auto-refresh indicator when run reaches terminal state', async () => {
        mockFetchOk(terminalEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        // Terminal state = no auto-refresh text
        expect(screen.queryByText('auto-refresh')).toBeNull();
    });

    it('renders EventCard components for each event', async () => {
        mockFetchOk(mockEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        // Each event renders its source_agent chip
        const adminBadges = screen.getAllByText('admin');
        expect(adminBadges.length).toBe(mockEvents.filter(e => e.source_agent === 'admin').length);

        // Each event renders a payload summary — "consult_council" from tool events
        const toolTexts = screen.getAllByText('consult_council');
        expect(toolTexts.length).toBeGreaterThanOrEqual(2);
    });

    it('shows run ID in header (truncated)', async () => {
        mockFetchOk([]);

        await act(async () => {
            render(<RunTimeline runId="run-abc-def-012345" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        // Header shows "Run: {runId.slice(0,8)}..."
        expect(screen.getByText(/Run: run-abc-/)).toBeDefined();
    });

    it('shows status badge "completed" for terminal events', async () => {
        mockFetchOk(terminalEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(screen.getByText('completed')).toBeDefined();
    });

    it('shows error state on fetch failure', async () => {
        mockFetchError(503);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(screen.getByText('Failed to load events (503)')).toBeDefined();
        expect(screen.getByText('Retry')).toBeDefined();
    });

    it('polls every 5s while run is not terminal', async () => {
        mockFetchOk(mockEvents);

        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        const initialCallCount = (global.fetch as ReturnType<typeof vi.fn>).mock.calls.length;

        // Advance 5s — should trigger another fetch
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        expect((global.fetch as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(initialCallCount);
    });
});
