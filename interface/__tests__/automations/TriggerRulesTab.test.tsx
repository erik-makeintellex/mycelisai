import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, fireEvent } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/automations',
    useSearchParams: () => new URLSearchParams(),
}));

// Mock labels
vi.mock('@/lib/labels', () => ({
    brainDisplayName: (id: string) => id,
    brainLocationLabel: (l: string) => l,
}));

// ── Store mock ───────────────────────────────────────────────

const mockFetchTriggerRules = vi.fn();
const mockCreateTriggerRule = vi.fn();
const mockDeleteTriggerRule = vi.fn();
const mockToggleTriggerRule = vi.fn();

let storeState: Record<string, unknown> = {};

vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = {
            triggerRules: storeState.triggerRules ?? [],
            isFetchingTriggers: storeState.isFetchingTriggers ?? false,
            fetchTriggerRules: mockFetchTriggerRules,
            createTriggerRule: mockCreateTriggerRule,
            deleteTriggerRule: mockDeleteTriggerRule,
            toggleTriggerRule: mockToggleTriggerRule,
        };
        return selector(state);
    },
}));

import TriggerRulesTab from '@/components/automations/TriggerRulesTab';
import type { TriggerRule } from '@/store/useCortexStore';

// ── Test data ────────────────────────────────────────────────

const now = new Date().toISOString();

const sampleRules: TriggerRule[] = [
    {
        id: 'rule-1',
        tenant_id: 'default',
        name: 'Auto-Archive on Completion',
        description: 'Archives mission artifacts when a mission completes',
        event_pattern: 'mission.completed',
        condition: {},
        target_mission_id: 'mission-abc-12345678',
        mode: 'propose',
        cooldown_seconds: 120,
        max_depth: 3,
        max_active_runs: 2,
        is_active: true,
        created_at: now,
        updated_at: now,
    },
    {
        id: 'rule-2',
        tenant_id: 'default',
        name: 'Retry on Failure',
        event_pattern: 'mission.failed',
        condition: {},
        target_mission_id: 'mission-def-87654321',
        mode: 'auto_execute',
        cooldown_seconds: 60,
        max_depth: 5,
        max_active_runs: 1,
        is_active: true,
        last_fired_at: now,
        created_at: now,
        updated_at: now,
    },
    {
        id: 'rule-3',
        tenant_id: 'default',
        name: 'Disabled Rule',
        event_pattern: 'tool.completed',
        condition: {},
        target_mission_id: 'mission-ghi-11111111',
        mode: 'propose',
        cooldown_seconds: 30,
        max_depth: 2,
        max_active_runs: 5,
        is_active: false,
        created_at: now,
        updated_at: now,
    },
];

// ── Tests ────────────────────────────────────────────────────

describe('TriggerRulesTab', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        storeState = {};
    });

    it('shows loading state when fetching triggers', async () => {
        storeState = { isFetchingTriggers: true, triggerRules: [] };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Loader2 spinner is rendered — find the animate-spin element
        const spinner = document.querySelector('.animate-spin');
        expect(spinner).toBeDefined();
        expect(spinner).not.toBeNull();
    });

    it('shows empty state when no trigger rules', async () => {
        storeState = { triggerRules: [], isFetchingTriggers: false };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('No trigger rules defined')).toBeDefined();
        expect(screen.getByText(/Trigger rules fire automatically/)).toBeDefined();
    });

    it('renders trigger rule cards when data is available', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('Auto-Archive on Completion')).toBeDefined();
        expect(screen.getByText('Retry on Failure')).toBeDefined();
        expect(screen.getByText('Disabled Rule')).toBeDefined();
    });

    it('shows create form button (New Rule)', async () => {
        storeState = { triggerRules: [] };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('New Rule')).toBeDefined();
    });

    it('shows mode badge — propose vs auto', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Rule 1 and 3 have mode "propose", rule 2 has "auto_execute" shown as "auto"
        const proposeBadges = screen.getAllByText('propose');
        expect(proposeBadges.length).toBe(2);

        const autoBadge = screen.getByText('auto');
        expect(autoBadge).toBeDefined();
    });

    it('shows guard badges when rule is expanded (cooldown, max_depth, max_active_runs)', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Click the first rule to expand it
        const ruleHeader = screen.getByText('Auto-Archive on Completion');
        await act(async () => {
            fireEvent.click(ruleHeader);
        });

        // Guard badges should now be visible
        expect(screen.getByText('Cooldown:')).toBeDefined();
        expect(screen.getByText('120s')).toBeDefined();
        expect(screen.getByText('Max Depth:')).toBeDefined();
        expect(screen.getByText('3')).toBeDefined();
        expect(screen.getByText('Max Runs:')).toBeDefined();
        expect(screen.getByText('2')).toBeDefined();
    });

    it('calls fetchTriggerRules on mount', async () => {
        storeState = { triggerRules: [] };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(mockFetchTriggerRules).toHaveBeenCalled();
    });

    it('shows disabled badge for inactive rules', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Rule 3 is inactive — shows "disabled" badge
        expect(screen.getByText('disabled')).toBeDefined();
    });

    it('shows event pattern for each rule', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('mission.completed')).toBeDefined();
        expect(screen.getByText('mission.failed')).toBeDefined();
        expect(screen.getByText('tool.completed')).toBeDefined();
    });

    it('shows footer summary with active/total count', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // 2 active / 3 total rules
        expect(screen.getByText(/2 active/)).toBeDefined();
        expect(screen.getByText(/3 total rules/)).toBeDefined();
    });

    it('shows description in expanded rule', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Expand first rule which has a description
        const ruleHeader = screen.getByText('Auto-Archive on Completion');
        await act(async () => {
            fireEvent.click(ruleHeader);
        });

        expect(screen.getByText('Archives mission artifacts when a mission completes')).toBeDefined();
    });

    it('toggles create form when New Rule is clicked', async () => {
        storeState = { triggerRules: [] };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Click "New Rule" button
        const newRuleBtn = screen.getByText('New Rule');
        await act(async () => {
            fireEvent.click(newRuleBtn);
        });

        // Form should appear with "New Trigger Rule" header
        expect(screen.getByText('New Trigger Rule')).toBeDefined();
        expect(screen.getByText('Create Rule')).toBeDefined();
        expect(screen.getByText('Cancel')).toBeDefined();
    });

    it('shows last_fired_at in expanded rule when present', async () => {
        storeState = { triggerRules: sampleRules };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        // Expand rule-2 which has last_fired_at
        const ruleHeader = screen.getByText('Retry on Failure');
        await act(async () => {
            fireEvent.click(ruleHeader);
        });

        expect(screen.getByText(/Last fired:/)).toBeDefined();
    });

    it('renders page heading and subtitle', async () => {
        storeState = { triggerRules: [] };

        await act(async () => {
            render(<TriggerRulesTab />);
        });

        expect(screen.getByText('Trigger Rules')).toBeDefined();
        expect(screen.getByText(/Declarative IF\/THEN rules/)).toBeDefined();
    });
});
