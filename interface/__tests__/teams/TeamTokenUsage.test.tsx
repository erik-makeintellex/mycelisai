import { act, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TeamTokenUsage from '@/components/teams/TeamTokenUsage';
import * as tokenBudgetsClient from '@/lib/tokenBudgets';

vi.mock('@/lib/tokenBudgets', async () => {
    const actual = await vi.importActual<typeof import('@/lib/tokenBudgets')>('@/lib/tokenBudgets');
    return { ...actual, getTokenBudgetUsage: vi.fn() };
});

describe('TeamTokenUsage', () => {
    beforeEach(() => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockReset();
    });

    it('formats k/M and shows the D10 example line for a run plus daily usage', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({
            ok: true,
            data: { scope: 'team_day', ref: 'team_alpha', used: 180000, limit: 2000000, remaining: 1820000, warn: false, warn_pct: 80, period: 'utc_day', usage_reported: true },
        });
        await act(async () => {
            render(<TeamTokenUsage teamId="team_alpha" runUsed={12000} runLimit={64000} />);
        });
        expect(await screen.findByText('This team used 12k of 64k tokens this run · 180k of 2M today.')).toBeDefined();
    });

    it('labels since_restart usage instead of claiming a full day of history', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({
            ok: true,
            data: { scope: 'team_day', ref: 'team_alpha', used: 5000, limit: 2000000, remaining: 1995000, warn: false, warn_pct: 80, period: 'since_restart', usage_reported: true },
        });
        await act(async () => {
            render(<TeamTokenUsage teamId="team_alpha" runUsed={1000} runLimit={64000} />);
        });
        expect(await screen.findByText(/since restart/i)).toBeDefined();
    });

    it('shows "at least" when the daily usage was not provider-reported', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({
            ok: true,
            data: { scope: 'team_day', ref: 'team_alpha', used: 40000, limit: 2000000, remaining: 1960000, warn: false, warn_pct: 80, period: 'utc_day', usage_reported: false },
        });
        await act(async () => {
            render(<TeamTokenUsage teamId="team_alpha" runUsed={2000} runLimit={64000} />);
        });
        expect(await screen.findByText(/at least 40k/i)).toBeDefined();
    });

    it('shows reserved tokens when the field is present and > 0 (B1R-A)', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({
            ok: true,
            data: { scope: 'team_day', ref: 'team_alpha', used: 180000, limit: 2000000, remaining: 1818000, warn: false, warn_pct: 80, period: 'utc_day', usage_reported: true, reserved: 2000 },
        });
        await act(async () => {
            render(<TeamTokenUsage teamId="team_alpha" runUsed={12000} runLimit={64000} />);
        });
        expect(await screen.findByText('This team used 12k of 64k tokens this run · 180k of 2M today · 2k reserved.')).toBeDefined();
    });

    it('omits the reserved clause when the field is absent or zero', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({
            ok: true,
            data: { scope: 'team_day', ref: 'team_alpha', used: 180000, limit: 2000000, remaining: 1820000, warn: false, warn_pct: 80, period: 'utc_day', usage_reported: true, reserved: 0 },
        });
        await act(async () => {
            render(<TeamTokenUsage teamId="team_alpha" runUsed={12000} runLimit={64000} />);
        });
        expect(await screen.findByText('This team used 12k of 64k tokens this run · 180k of 2M today.')).toBeDefined();
        expect(screen.queryByText(/reserved/i)).toBeNull();
    });

    it('renders nothing on a 403 (non-member), never a fake zero', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({ ok: false, code: 'token_budget_usage_forbidden', httpStatus: 403 });
        const { container } = render(<TeamTokenUsage teamId="team_alpha" runUsed={2000} runLimit={64000} />);
        await waitFor(() => {
            expect(tokenBudgetsClient.getTokenBudgetUsage).toHaveBeenCalled();
        });
        expect(container.textContent).toBe('');
    });

    it('renders nothing while loading (no placeholder flash)', () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockReturnValue(new Promise(() => {}));
        const { container } = render(<TeamTokenUsage teamId="team_alpha" runUsed={2000} runLimit={64000} />);
        expect(container.textContent).toBe('');
    });

    it('renders nothing on any other fetch failure rather than a broken line', async () => {
        vi.mocked(tokenBudgetsClient.getTokenBudgetUsage).mockResolvedValue({ ok: false, code: 'request_failed', httpStatus: 0 });
        const { container } = render(<TeamTokenUsage teamId="team_alpha" runUsed={2000} runLimit={64000} />);
        await waitFor(() => {
            expect(tokenBudgetsClient.getTokenBudgetUsage).toHaveBeenCalled();
        });
        expect(container.textContent).toBe('');
    });
});
