import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-1',
    useSearchParams: () => new URLSearchParams(),
}));

vi.mock('@/lib/labels', () => ({
    brainDisplayName: (id: string) => id,
    brainLocationLabel: (l: string) => l,
}));

import RunTimeline from '@/components/runs/RunTimeline';
import type { MissionEvent } from '@/store/useCortexStore';

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

async function renderTimeline(runId = 'run-abc') {
    await act(async () => {
        render(<RunTimeline runId={runId} />);
    });
    await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
    });
}

describe('RunTimeline', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('shows loading state when fetching timeline', async () => {
        (global.fetch as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
        await act(async () => {
            render(<RunTimeline runId="run-abc" />);
        });
        expect(screen.getByText('Loading events...')).toBeDefined();
    });

    it('shows empty state when no events', async () => {
        mockFetchOk([]);
        await renderTimeline();
        expect(screen.getByText(/No events yet/)).toBeDefined();
        expect(screen.getByText(/Auto-refreshing every 5s/)).toBeDefined();
    });

    it('renders events and EventCard details when data is available', async () => {
        mockFetchOk(mockEvents);
        await renderTimeline();

        expect(screen.getByText('mission.started')).toBeDefined();
        expect(screen.getByText('tool.invoked')).toBeDefined();
        expect(screen.getByText('tool.completed')).toBeDefined();
        expect(screen.getAllByText('admin').length).toBe(3);
        expect(screen.getAllByText('consult_council').length).toBeGreaterThanOrEqual(2);
    });

    it('calls fetch on mount and shows truncated run ID in header', async () => {
        mockFetchOk([]);
        await renderTimeline('run-abc-def-012345');

        expect(global.fetch).toHaveBeenCalledWith('/api/v1/runs/run-abc-def-012345/events');
        expect(screen.getByText(/Run: run-abc-/)).toBeDefined();
    });

    it('shows auto-refresh while running and hides it when terminal', async () => {
        mockFetchOk(mockEvents);
        const { unmount } = render(<RunTimeline runId="run-abc" />);
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });
        expect(screen.getByText('auto-refresh')).toBeDefined();
        unmount();

        mockFetchOk(terminalEvents);
        await renderTimeline();
        expect(screen.queryByText('auto-refresh')).toBeNull();
        expect(screen.getByText('completed')).toBeDefined();
    });

    it('summarizes Codex task package artifact events', async () => {
        mockFetchOk([
            {
                id: 'evt-package',
                run_id: 'run-abc',
                tenant_id: 'default',
                event_type: 'artifact.created',
                severity: 'info',
                source_agent: 'admin',
                emitted_at: now,
                payload: {
                    package_kind: 'codex_task_bundle',
                    package_slug: 'myc-rc-001',
                    proof_path: 'codex-task-packages/myc-rc-001/proof.md',
                    files: [
                        { path: 'codex-task-packages/myc-rc-001/implementation_task.md' },
                        { path: 'codex-task-packages/myc-rc-001/proof.md' },
                    ],
                },
            },
        ]);
        await renderTimeline();
        expect(screen.getByText('Codex task bundle: myc-rc-001 · 2 files · proof: codex-task-packages/myc-rc-001/proof.md')).toBeDefined();
    });

    it('shows error state on fetch failure and polls every 5s when non-terminal', async () => {
        mockFetchError(503);
        const { unmount } = render(<RunTimeline runId="run-abc" />);
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });
        expect(screen.getByText('Failed to load events (503)')).toBeDefined();
        expect(screen.getByText('Retry')).toBeDefined();
        unmount();

        mockFetchOk(mockEvents);
        await renderTimeline();
        const initialCalls = (global.fetch as ReturnType<typeof vi.fn>).mock.calls.length;
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });
        expect((global.fetch as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(initialCalls);
    });

    it('renders run-stage-progress-strip and reflects active, completed, and failed stages', async () => {
        mockFetchOk(mockEvents);
        const { unmount } = render(<RunTimeline runId="run-abc" />);
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });

        expect(screen.getByTestId('run-stage-progress-strip')).toBeDefined();
        expect(screen.getByTestId('run-stage-intent_received').getAttribute('data-stage-status')).toBe('completed');
        expect(screen.getByTestId('run-stage-tool_action_execution').getAttribute('data-stage-status')).toBe('active');
        expect(screen.getByTestId('run-stage-package_generated').getAttribute('data-stage-status')).toBe('pending');
        unmount();

        mockFetchOk(terminalEvents);
        const { unmount: unmountCompleted } = render(<RunTimeline runId="run-abc" />);
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });
        expect(screen.getByTestId('run-stage-proof_generated').getAttribute('data-stage-status')).toBe('completed');
        expect(screen.getByTestId('run-stage-recovery_next_actions').getAttribute('data-stage-status')).toBe('completed');
        unmountCompleted();

        mockFetchOk([
            ...mockEvents,
            { id: 'evt-fail', run_id: 'run-abc', tenant_id: 'default', event_type: 'tool.failed', severity: 'error', emitted_at: now },
        ]);
        await renderTimeline();
        expect(screen.getByTestId('run-stage-tool_action_execution').getAttribute('data-stage-status')).toBe('failed');
        expect(screen.getByTestId('run-stage-recovery_next_actions').getAttribute('data-stage-status')).toBe('failed');
    });
});
