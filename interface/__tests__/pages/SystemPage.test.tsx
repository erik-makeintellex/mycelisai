import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
const mockSearchParams = new URLSearchParams();
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/system',
    useSearchParams: () => mockSearchParams,
}));

// Mock MatrixGrid (heavy component)
vi.mock('@/components/matrix/MatrixGrid', () => ({
    __esModule: true,
    default: () => <div data-testid="matrix-grid">MatrixGrid</div>,
}));

import SystemPage from '@/app/(app)/system/page';

describe('System Page (V7 — Advanced)', () => {
    beforeEach(() => {
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        // Mock fetch for health checks
        vi.spyOn(global, 'fetch').mockResolvedValue({
            ok: true,
            json: async () => ({ goroutines: 42, heap_alloc_mb: 18.5, sys_mem_mb: 52.0, llm_tokens_sec: 3.2, timestamp: new Date().toISOString() }),
        } as Response);
    });

    it('renders page title', async () => {
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('System')).toBeDefined();
    });

    it('renders Advanced badge', async () => {
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('Advanced')).toBeDefined();
    });

    it('renders all tabs', async () => {
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('Event Health')).toBeDefined();
        expect(screen.getByText('NATS Status')).toBeDefined();
        expect(screen.getByText('Database')).toBeDefined();
        expect(screen.getByText('Cognitive Matrix')).toBeDefined();
        expect(screen.getByText('Debug')).toBeDefined();
    });

    it('defaults to Event Health tab', async () => {
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('LIVE')).toBeDefined();
    });

    it('deep-links to matrix tab via search param', async () => {
        mockSearchParams.set('tab', 'matrix');
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByTestId('matrix-grid')).toBeDefined();
    });
});
