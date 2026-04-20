import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-1/chain',
    useSearchParams: () => new URLSearchParams(),
}));

import ViewChain from '@/components/runs/ViewChain';

const now = new Date().toISOString();

const chainResponse = {
    run_id: 'run-root',
    mission_id: 'mission-123',
    chain: [
        {
            id: 'run-root',
            mission_id: 'mission-123',
            tenant_id: 'default',
            status: 'completed',
            run_depth: 0,
            started_at: now,
            completed_at: now,
        },
        {
            id: 'run-child',
            mission_id: 'mission-123',
            tenant_id: 'default',
            status: 'running',
            run_depth: 1,
            parent_run_id: 'run-root',
            started_at: now,
        },
    ],
};

describe('ViewChain', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders the causal chain tree from the run chain API', async () => {
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => chainResponse,
        });

        await act(async () => {
            render(<ViewChain runId="run-root" />);
        });

        await act(async () => {
            await new Promise((resolve) => setTimeout(resolve, 0));
        });

        expect(screen.getByText('Causal Chain')).toBeDefined();
        expect(screen.getByText('run-root')).toBeDefined();
        expect(screen.getByText('run-child')).toBeDefined();
        expect(screen.getByText('depth 0')).toBeDefined();
        expect(screen.getByText('depth 1')).toBeDefined();
        expect(screen.getByText('completed')).toBeDefined();
        expect(screen.getByText('running')).toBeDefined();
    });

    it('shows an error state when the chain request fails', async () => {
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: false,
            status: 503,
            json: async () => ({}),
        });

        await act(async () => {
            render(<ViewChain runId="run-root" />);
        });

        await act(async () => {
            await new Promise((resolve) => setTimeout(resolve, 0));
        });

        expect(screen.getByText('Failed to load causal chain (503)')).toBeDefined();
        expect(screen.getByText('Retry')).toBeDefined();
    });
});
