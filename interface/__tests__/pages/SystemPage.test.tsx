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

const mockAdvancedMode = vi.fn(() => true);
const mockFetchServicesStatus = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) =>
        selector({
            advancedMode: mockAdvancedMode(),
            servicesStatus: [],
            isFetchingServicesStatus: false,
            servicesStatusUpdatedAt: null,
            fetchServicesStatus: mockFetchServicesStatus,
        }),
}));

import SystemPage from '@/app/(app)/system/page';

describe('System Page (V8.1 advanced diagnostics)', () => {
    beforeEach(() => {
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        mockAdvancedMode.mockReturnValue(true);
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
        expect(screen.getByText('Runtime Health')).toBeDefined();
        expect(screen.getByText('Event Bus')).toBeDefined();
        expect(screen.getByText('Storage')).toBeDefined();
        expect(screen.getByText('Services')).toBeDefined();
    });

    it('defaults to Runtime Health tab', async () => {
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('LIVE')).toBeDefined();
    });

    it('deep-links to services tab via search param', async () => {
        mockSearchParams.set('tab', 'services');
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('Services')).toBeDefined();
    });

    it('shows the advanced gate when advanced mode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText(/System diagnostics are hidden until you open Advanced mode/i)).toBeDefined();
    });

    it('shows lifecycle commands through the supported invoke contract', async () => {
        mockSearchParams.set('tab', 'services');
        await act(async () => { render(<SystemPage />); });
        expect(screen.getByText('uv run inv lifecycle.up --build --frontend')).toBeDefined();
        expect(screen.getByText('uv run inv lifecycle.down')).toBeDefined();
        expect(screen.queryByText(/uvx inv lifecycle/i)).toBeNull();
    });
});
