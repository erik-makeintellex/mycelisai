import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// The dashboard page uses next/dynamic to lazy-load MissionControl with ssr:false.
// In jsdom there's no real SSR boundary, so mock MissionControl directly and
// mock next/dynamic to return the component synchronously.
vi.mock('@/components/dashboard/MissionControl', () => ({
    __esModule: true,
    default: () => <div data-testid="mission-control">MissionControl</div>,
}));

vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any) => {
        // Vitest hoists vi.mock, so the MissionControl mock is already in place.
        // We can't resolve the promise synchronously, so just return a wrapper
        // that renders the known mock directly.
        const MockComponent = (props: any) => <div data-testid="mission-control">MissionControl</div>;
        MockComponent.displayName = 'DynamicMock';
        return MockComponent;
    },
}));

import DashboardPage from '@/app/(app)/dashboard/page';

describe('Dashboard Page (/dashboard)', () => {
    it('mounts without crashing', () => {
        render(<DashboardPage />);
        expect(screen.getByTestId('mission-control')).toBeDefined();
    });

    it('renders the MissionControl layout component', () => {
        render(<DashboardPage />);
        expect(screen.getByText('MissionControl')).toBeDefined();
    });
});
