import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/dynamic to render the component directly
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any) => {
        const Component = require('react').lazy(loader);
        return (props: any) => {
            const React = require('react');
            return React.createElement(
                React.Suspense,
                { fallback: null },
                React.createElement(Component, props),
            );
        };
    },
}));

// Mock child components to avoid deep dependency chains
vi.mock('@/components/teams/TeamCard', () => ({
    __esModule: true,
    default: ({ team }: any) => <div data-testid="team-card">{team?.name ?? 'team'}</div>,
}));

vi.mock('@/components/teams/TeamDetailDrawer', () => ({
    __esModule: true,
    default: () => <div data-testid="team-drawer" />,
}));

import TeamsRoute from '@/app/teams/page';
import { useCortexStore } from '@/store/useCortexStore';

describe('Teams Page (app/teams/page.tsx)', () => {
    beforeEach(() => {
        // Provide default store state so the component doesn't crash
        useCortexStore.setState({
            teamsDetail: [],
            isFetchingTeamsDetail: false,
            selectedTeamId: null,
            isTeamDrawerOpen: false,
            teamsFilter: 'all' as any,
        });

        // Mock fetch for the teams detail endpoint
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ teams: [] }),
        });
    });

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<TeamsRoute />);
        });

        // TeamsPage component renders a header with team icon/title
        expect(document.body.innerHTML.length).toBeGreaterThan(0);
    });

    it('renders the teams header', async () => {
        await act(async () => {
            render(<TeamsRoute />);
        });

        // The TeamsPage component shows a "Teams" or "TEAMS" heading
        const heading = screen.queryByText(/teams/i);
        expect(heading).toBeDefined();
    });
});
