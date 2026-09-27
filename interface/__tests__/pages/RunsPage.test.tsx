import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

const routerPush = vi.fn();
const fetchRecentRuns = vi.fn();
const toggleAdvancedMode = vi.fn();
const searchParams = new URLSearchParams();

type StoreState = {
    advancedMode: boolean;
    toggleAdvancedMode: () => void;
    recentRuns: Array<{
        id: string;
        mission_id: string;
        status: string;
        started_at: string;
    }>;
    isFetchingRuns: boolean;
    fetchRecentRuns: () => void;
    assistantName: string;
};

let storeState: StoreState;

vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: routerPush,
        replace: vi.fn(),
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    usePathname: () => '/runs',
    useSearchParams: () => searchParams,
}));

vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: (state: StoreState) => unknown) => selector(storeState),
}));

import RunsPage from '@/app/(app)/runs/page';

describe('RunsPage', () => {
    beforeEach(() => {
        routerPush.mockReset();
        fetchRecentRuns.mockReset();
        toggleAdvancedMode.mockReset();
        searchParams.delete('status');
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: 'admin' } } }) });
        storeState = {
            advancedMode: true,
            toggleAdvancedMode,
            recentRuns: [
                {
                    id: 'run-alpha-123',
                    mission_id: 'mission-alpha-123',
                    status: 'running',
                    started_at: new Date(Date.now() - 20 * 1000).toISOString(),
                },
            ],
            isFetchingRuns: false,
            fetchRecentRuns,
            assistantName: 'Soma',
        };
    });

    it('shows the advanced gate when advanced mode is off', async () => {
        storeState = {
            ...storeState,
            advancedMode: false,
        };

        render(<RunsPage />);

        expect(await screen.findByText('Run lists are in Admin tools')).toBeDefined();
        expect(screen.getByText(/inspect execution history across workflows/i)).toBeDefined();
        expect(screen.queryByText('run-alpha-123')).toBeNull();
    });

    it('fetches runs on mount and exposes run detail links', async () => {
        render(<RunsPage />);

        expect(await screen.findByText('run-alpha-123')).toBeDefined();
        expect(fetchRecentRuns).toHaveBeenCalledTimes(1);
        expect(screen.getByLabelText('Outcome health: Running')).toBeDefined();

        const row = screen.getByText('run-alpha-123').closest('a');
        expect(row).toBeDefined();

        expect(row?.getAttribute('href')).toBe('/runs/run-alpha-123');
    });

    it('renders empty state when no runs are available', async () => {
        storeState = {
            ...storeState,
            recentRuns: [],
        };

        render(<RunsPage />);

        expect(await screen.findByText('No runs yet')).toBeDefined();
        expect(screen.getByText(/Ask Soma for an outcome first/i)).toBeDefined();
    });

    it('filters to active runs from the status query', async () => {
        searchParams.set('status', 'running');
        storeState = {
            ...storeState,
            recentRuns: [
                ...storeState.recentRuns,
                {
                    id: 'run-complete-123',
                    mission_id: 'mission-complete-123',
                    status: 'completed',
                    started_at: new Date().toISOString(),
                },
            ],
        };

        render(<RunsPage />);

        expect(await screen.findByText('Active Runs')).toBeDefined();
        expect(screen.getByText('run-alpha-123')).toBeDefined();
        expect(screen.queryByText('run-complete-123')).toBeNull();
    });
});

