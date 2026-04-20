import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-123/chain',
    useSearchParams: () => new URLSearchParams(),
}));

vi.mock('react', async () => {
    const actual = await vi.importActual<typeof import('react')>('react');
    return {
        ...actual,
        use: () => ({ id: 'test-run-123-abcd-5678' }),
    };
});

import RunChainPage from '@/app/(app)/runs/[id]/chain/page';

const now = new Date().toISOString();

describe('RunChainPage (/runs/[id]/chain)', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders the causal chain route with navigation back to the run', async () => {
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                run_id: 'test-run-123-abcd-5678',
                mission_id: 'mission-123',
                chain: [
                    {
                        id: 'test-run-123-abcd-5678',
                        mission_id: 'mission-123',
                        tenant_id: 'default',
                        status: 'completed',
                        run_depth: 0,
                        started_at: now,
                        completed_at: now,
                    },
                ],
            }),
        });

        await act(async () => {
            render(<RunChainPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        await act(async () => {
            await new Promise((resolve) => setTimeout(resolve, 0));
        });

        expect(screen.getByText(/Run Chain:/)).toBeDefined();
        expect(screen.getByText('Causal Chain')).toBeDefined();
        expect(screen.getByText('Workspace')).toBeDefined();

        const runLink = screen.getByText('Run').closest('a');
        expect(runLink).toBeDefined();
        expect(runLink?.getAttribute('href')).toBe('/runs/test-run-123-abcd-5678');
    });
});
