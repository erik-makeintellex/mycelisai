import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TokenBudgetsPanel from '@/components/settings/TokenBudgetsPanel';
import * as tokenBudgetsClient from '@/lib/tokenBudgets';
import type { TokenBudgetsResponse } from '@/lib/tokenBudgets';

vi.mock('@/lib/tokenBudgets', async () => {
    const actual = await vi.importActual<typeof import('@/lib/tokenBudgets')>('@/lib/tokenBudgets');
    return {
        ...actual,
        getTokenBudgets: vi.fn(),
        putTokenBudgetOverride: vi.fn(),
        deleteTokenBudgetOverride: vi.fn(),
    };
});

const BASE: TokenBudgetsResponse = {
    default_class: 'local_large',
    class_order: ['local_large', 'local_small', 'hosted_standard', 'hosted_premium'],
    global: { per_execution: 32000, per_run: 128000, per_team_day: 500000, per_agent_day: 250000, warn_pct: 80 },
    classes: {
        local_large: { per_execution: 64000, per_run: 256000, per_team_day: 2000000, per_agent_day: 1000000, warn_pct: 80 },
        hosted_premium: { per_execution: 24000, per_run: 96000, per_team_day: 150000, per_agent_day: 75000, warn_pct: 80 },
    },
    day_period: 'utc_day',
    min_limit: 1024,
    max_limit: 5000000,
    can_edit: false,
};

describe('TokenBudgetsPanel', () => {
    beforeEach(() => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockReset();
        vi.mocked(tokenBudgetsClient.putTokenBudgetOverride).mockReset();
        vi.mocked(tokenBudgetsClient.deleteTokenBudgetOverride).mockReset();
    });

    it('renders the effective class and global table for a non-admin', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: BASE });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        expect(await screen.findByText('local_large')).toBeDefined();
        expect(screen.getByText('hosted_premium')).toBeDefined();
        expect(screen.getAllByText('64,000').length).toBeGreaterThan(0);
        expect(screen.getByText(/global/i)).toBeDefined();
    });

    it('hides the provider table when providers is absent', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: BASE });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByText('local_large');
        expect(screen.queryByText(/provider/i)).toBeNull();
    });

    it('shows the providers table when present (root-admin full view)', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({
            ok: true,
            data: { ...BASE, providers: [{ provider_id: 'openai_prod', model_id: 'gpt-4-turbo', budget_class: 'hosted_premium', class_source: 'explicit' }] },
        });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        expect(await screen.findByText('openai_prod')).toBeDefined();
        expect(screen.getByText('gpt-4-turbo')).toBeDefined();
    });

    it('hides override controls when can_edit is false', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: { ...BASE, can_edit: false } });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByText('local_large');
        expect(screen.queryByRole('button', { name: /save override/i })).toBeNull();
    });

    it('shows override controls when can_edit is true', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: { ...BASE, can_edit: true, overrides: {} } });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        expect(await screen.findByRole('button', { name: /save override/i })).toBeDefined();
    });

    it('lists an existing team override with a remove control', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({
            ok: true,
            data: { ...BASE, can_edit: true, overrides: { team: { team_alpha: { per_execution: 2000 } } } },
        });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        expect(await screen.findByText('team_alpha')).toBeDefined();
        expect(screen.getByRole('button', { name: /remove/i })).toBeDefined();
    });

    it('validates client-side before calling the server (value below min_limit)', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: { ...BASE, can_edit: true, overrides: {} } });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByRole('button', { name: /save override/i });

        fireEvent.change(screen.getByLabelText(/ref/i), { target: { value: 'team_alpha' } });
        fireEvent.change(screen.getByLabelText(/per.execution/i), { target: { value: '100' } });
        fireEvent.click(screen.getByRole('button', { name: /save override/i }));

        expect(await screen.findByText(/1,024|min/i)).toBeDefined();
        expect(tokenBudgetsClient.putTokenBudgetOverride).not.toHaveBeenCalled();
    });

    it('submits a valid override and refreshes on success', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: { ...BASE, can_edit: true, overrides: {} } });
        vi.mocked(tokenBudgetsClient.putTokenBudgetOverride).mockResolvedValue({
            ok: true,
            data: { level: 'team', ref: 'team_alpha', changed: true, effective: { limits: {}, sources: {} }, audit_event_id: 'a1' },
        });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByRole('button', { name: /save override/i });

        fireEvent.change(screen.getByLabelText(/ref/i), { target: { value: 'team_alpha' } });
        fireEvent.change(screen.getByLabelText(/per.execution/i), { target: { value: '2000' } });
        fireEvent.click(screen.getByRole('button', { name: /save override/i }));

        await waitFor(() => {
            expect(tokenBudgetsClient.putTokenBudgetOverride).toHaveBeenCalledWith('team', 'team_alpha', expect.objectContaining({ per_execution: 2000 }));
        });
    });

    it('shows the honest server error when the override PUT fails, without pretending success', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: true, data: { ...BASE, can_edit: true, overrides: {} } });
        vi.mocked(tokenBudgetsClient.putTokenBudgetOverride).mockResolvedValue({ ok: false, code: 'admin_required', httpStatus: 403 });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByRole('button', { name: /save override/i });

        fireEvent.change(screen.getByLabelText(/ref/i), { target: { value: 'team_alpha' } });
        fireEvent.change(screen.getByLabelText(/per.execution/i), { target: { value: '2000' } });
        fireEvent.click(screen.getByRole('button', { name: /save override/i }));

        expect(await screen.findByRole('alert')).toBeDefined();
    });

    it('removes an override and refreshes', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({
            ok: true,
            data: { ...BASE, can_edit: true, overrides: { team: { team_alpha: { per_execution: 2000 } } } },
        });
        vi.mocked(tokenBudgetsClient.deleteTokenBudgetOverride).mockResolvedValue({
            ok: true,
            data: { level: 'team', ref: 'team_alpha', changed: true, effective: { limits: {}, sources: {} }, audit_event_id: 'a2' },
        });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        await screen.findByText('team_alpha');

        fireEvent.click(screen.getByRole('button', { name: /remove/i }));

        await waitFor(() => {
            expect(tokenBudgetsClient.deleteTokenBudgetOverride).toHaveBeenCalledWith('team', 'team_alpha');
        });
    });

    it('shows a blocker, not a blank panel, when the budgets fetch fails', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgets).mockResolvedValue({ ok: false, code: 'service_unavailable', httpStatus: 503 });
        await act(async () => {
            render(<TokenBudgetsPanel />);
        });
        expect(await screen.findByRole('alert')).toBeDefined();
    });
});
