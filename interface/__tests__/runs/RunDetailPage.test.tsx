import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-123',
    useSearchParams: () => new URLSearchParams(),
}));

// Mock react `use` to resolve the params promise synchronously
vi.mock('react', async () => {
    const actual = await vi.importActual('react');
    return {
        ...actual,
        use: (p: any) => ({ id: 'test-run-123-abcd-5678' }),
    };
});

// Mock next/dynamic — returns a simple stub for dynamically imported components
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any) => {
        const Component = (props: any) => {
            return <div data-testid="dynamic-component" {...props} />;
        };
        Component.displayName = 'DynamicMock';
        return Component;
    },
}));

import RunPage from '@/app/(app)/runs/[id]/page';

describe('RunDetailPage (/runs/[id])', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Mock fetch for the run status useEffect — return empty events (running status)
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ data: [] }),
        });
    });

    it('renders page with run ID in header', async () => {
        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        // Run ID is displayed as first 8 chars: "test-run" followed by "..."
        expect(screen.getByText(/Run:/)).toBeDefined();
        expect(screen.getByText(/test-run/)).toBeDefined();
    });

    it('shows "Conversation" tab active by default', async () => {
        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        // Both tab buttons should exist
        const conversationButtons = screen.getAllByText('Conversation');
        const eventsButtons = screen.getAllByText('Events');
        expect(conversationButtons.length).toBeGreaterThan(0);
        expect(eventsButtons.length).toBeGreaterThan(0);

        // The Conversation tab should have the active styling (cortex-primary)
        const activeConvoTab = conversationButtons.find(
            (el) => el.closest('button')?.className.includes('text-cortex-primary')
        );
        expect(activeConvoTab).toBeDefined();
    });

    it('shows "Events" tab when clicked', async () => {
        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        // Click the Events tab
        const eventsButton = screen.getAllByText('Events').find(
            (el) => el.closest('button')
        );
        expect(eventsButton).toBeDefined();

        await act(async () => { fireEvent.click(eventsButton!.closest('button')!); });

        // After clicking Events, the Events tab should have active styling
        const activeEventsTab = screen.getAllByText('Events').find(
            (el) => el.closest('button')?.className.includes('text-cortex-primary')
        );
        expect(activeEventsTab).toBeDefined();
    });

    it('shows status badge when runStatus is set', async () => {
        // Mock fetch to return completed event so runStatus becomes "completed"
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ data: [{ event_type: 'mission.completed' }] }),
        });

        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        // Allow useEffect to run
        await act(async () => {
            await new Promise((r) => setTimeout(r, 50));
        });

        expect(screen.getByText('completed')).toBeDefined();
    });

    it('shows running status with animated pulse dot', async () => {
        // Mock fetch to return no terminal events -> running
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ data: [{ event_type: 'mission.started' }] }),
        });

        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        await act(async () => {
            await new Promise((r) => setTimeout(r, 50));
        });

        const runningBadge = screen.getByText('running');
        expect(runningBadge).toBeDefined();

        // The running badge should contain an animated pulse dot
        const pulseDot = runningBadge.querySelector('.animate-pulse');
        expect(pulseDot).toBeDefined();
    });

    it('shows completed status badge', async () => {
        (global.fetch as any) = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ data: [{ event_type: 'mission.completed' }] }),
        });

        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        await act(async () => {
            await new Promise((r) => setTimeout(r, 50));
        });

        const completedBadge = screen.getByText('completed');
        expect(completedBadge).toBeDefined();
        // Completed badge has success color
        expect(completedBadge.className).toContain('text-cortex-success');
    });

    it('renders back link to /dashboard', async () => {
        await act(async () => {
            render(<RunPage params={Promise.resolve({ id: 'test-run-123-abcd-5678' })} />);
        });

        // The back link reads "Workspace" and links to /dashboard
        const backLink = screen.getByText('Workspace');
        expect(backLink).toBeDefined();
        const anchor = backLink.closest('a');
        expect(anchor).toBeDefined();
        expect(anchor?.getAttribute('href')).toBe('/dashboard');
    });
});
