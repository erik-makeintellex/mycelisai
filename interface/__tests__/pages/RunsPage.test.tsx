import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

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

    it('shows the advanced gate when advanced mode is off', () => {
        storeState = {
            ...storeState,
            advancedMode: false,
        };

        render(<RunsPage />);

        expect(screen.getByText('Run lists are an Advanced proof view')).toBeDefined();
        expect(screen.getByText(/inspect execution history across workflows/i)).toBeDefined();
        expect(screen.queryByText('run-alpha-123')).toBeNull();
    });

    it('fetches runs on mount and navigates to run details on click', () => {
        render(<RunsPage />);

        expect(fetchRecentRuns).toHaveBeenCalledTimes(1);
        expect(screen.getByText('run-alpha-123')).toBeDefined();
        expect(screen.getByText('running')).toBeDefined();

        const row = screen.getByText('run-alpha-123').closest('button');
        expect(row).toBeDefined();
        fireEvent.click(row!);

        expect(routerPush).toHaveBeenCalledWith('/runs/run-alpha-123');
    });

    it('renders empty state when no runs are available', () => {
        storeState = {
            ...storeState,
            recentRuns: [],
        };

        render(<RunsPage />);

        expect(screen.getByText('No runs yet')).toBeDefined();
        expect(screen.getByText(/Ask Soma for an outcome first/i)).toBeDefined();
    });

    it('filters to active runs from the status query', () => {
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

        expect(screen.getByText('Active Runs')).toBeDefined();
        expect(screen.getByText('run-alpha-123')).toBeDefined();
        expect(screen.queryByText('run-complete-123')).toBeNull();
    });
});

