import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, fireEvent } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

const mockFetchTriggerRules = vi.fn();

let storeState: Record<string, unknown> = {};

vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = {
            triggerRules: storeState.triggerRules ?? [],
            isFetchingTriggers: storeState.isFetchingTriggers ?? false,
            triggerRulesLoadError: storeState.triggerRulesLoadError ?? null,
            fetchTriggerRules: mockFetchTriggerRules,
            createTriggerRule: vi.fn(),
            deleteTriggerRule: vi.fn(),
            toggleTriggerRule: vi.fn(),
        };
        return selector(state);
    },
}));

import TriggerRulesTab from '@/components/automations/TriggerRulesTab';

describe('TriggerRulesTab load failure', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        storeState = {};
    });

    it('shows a load failure with Retry instead of the empty state when the fetch failed', async () => {
        storeState = {
            triggerRules: [],
            isFetchingTriggers: false,
            triggerRulesLoadError: 'Could not load trigger rules (HTTP 500).',
        };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.queryByText('No trigger rules defined')).toBeNull();
        const failure = screen.getByTestId('trigger-rules-load-failure');
        expect(failure.textContent).toMatch(/Could not load trigger rules/);

        await act(async () => {
            fireEvent.click(screen.getByRole('button', { name: /Retry/i }));
        });
        expect(mockFetchTriggerRules).toHaveBeenCalledTimes(2);
    });

    it('shows the empty state (not a failure) after a successful empty fetch', async () => {
        storeState = { triggerRules: [], isFetchingTriggers: false, triggerRulesLoadError: null };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('No trigger rules defined')).toBeDefined();
        expect(screen.queryByTestId('trigger-rules-load-failure')).toBeNull();
    });
});
