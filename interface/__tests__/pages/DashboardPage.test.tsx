import { describe, it, expect, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// The dashboard page renders MissionControlLayout which has deep dependencies
// (SignalProvider with EventSource, TelemetryRow, OperationsBoard, etc.)
// Mock the heavy component to isolate the page smoke test.
vi.mock('@/components/dashboard/MissionControl', () => ({
    __esModule: true,
    default: () => <div data-testid="mission-control">MissionControl</div>,
}));

import DashboardPage from '@/app/(app)/dashboard/page';

describe('Dashboard Page (/dashboard)', () => {
    it('mounts without crashing', async () => {
        await act(async () => {
            render(<DashboardPage />);
        });

        expect(screen.getByTestId('mission-control')).toBeDefined();
    });

    it('renders the MissionControl layout component', async () => {
        await act(async () => {
            render(<DashboardPage />);
        });

        expect(screen.getByText('MissionControl')).toBeDefined();
    });
});
