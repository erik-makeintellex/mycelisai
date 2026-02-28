import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

// Mock child components to isolate TeamsPage logic
vi.mock('@/components/teams/TeamDetailDrawer', () => ({
    __esModule: true,
    default: ({ team, onClose }: any) => (
        <div data-testid="team-detail-drawer">
            <span>Drawer: {team.name}</span>
            <button onClick={onClose}>Close</button>
        </div>
    ),
}));

import TeamsPage from '@/components/teams/TeamsPage';
import { useCortexStore } from '@/store/useCortexStore';

// Sample team data matching TeamDetailEntry
const mockTeams = [
    {
        id: 'team-alpha',
        name: 'Alpha Squad',
        role: 'action',
        type: 'standing' as const,
        mission_id: null,
        mission_intent: null,
        inputs: ['nats.input.alpha'],
        deliveries: ['nats.output.alpha'],
        agents: [
            { id: 'agent-1', role: 'cognitive', status: 1, last_heartbeat: new Date().toISOString(), tools: [], model: 'qwen' },
            { id: 'agent-2', role: 'sensory', status: 0, last_heartbeat: new Date().toISOString(), tools: [], model: 'qwen' },
        ],
    },
    {
        id: 'team-bravo',
        name: 'Bravo Ops',
        role: 'expression',
        type: 'mission' as const,
        mission_id: 'mission-001',
        mission_intent: 'Deploy sentinel network',
        inputs: [],
        deliveries: [],
        agents: [
            { id: 'agent-3', role: 'actuation', status: 2, last_heartbeat: new Date().toISOString(), tools: ['exec'], model: 'llama' },
        ],
    },
];

describe('TeamsPage', () => {
    beforeEach(() => {
        // Reset store state before each test
        useCortexStore.setState({
            teamsDetail: [],
            isFetchingTeamsDetail: false,
            selectedTeamId: null,
            isTeamDrawerOpen: false,
            teamsFilter: 'all',
            // Override fetchTeamsDetail to prevent actual fetching
            fetchTeamsDetail: vi.fn(),
        });
    });

    it('renders card grid of teams', () => {
        useCortexStore.setState({
            teamsDetail: mockTeams,
        });

        render(<TeamsPage />);

        // Both team cards should be rendered
        expect(screen.getByText('Alpha Squad')).toBeDefined();
        expect(screen.getByText('Bravo Ops')).toBeDefined();

        // Header should show team count
        expect(screen.getByText(/2 teams/)).toBeDefined();

        // Should show agent online stats (agent-1 is online=1, agent-3 is busy=2 → 2 online out of 3)
        expect(screen.getByText('2/3 agents online')).toBeDefined();
    });

    it('filter dropdown filters teams by type', () => {
        useCortexStore.setState({
            teamsDetail: mockTeams,
        });

        render(<TeamsPage />);

        // Initially shows all teams
        expect(screen.getByText('Alpha Squad')).toBeDefined();
        expect(screen.getByText('Bravo Ops')).toBeDefined();

        // Change filter to "standing"
        const filterSelect = screen.getByDisplayValue('All Teams');
        fireEvent.change(filterSelect, { target: { value: 'standing' } });

        // Only standing team should remain
        expect(screen.getByText('Alpha Squad')).toBeDefined();
        expect(screen.queryByText('Bravo Ops')).toBeNull();
    });

    it('clicking a team card opens the detail drawer', () => {
        useCortexStore.setState({
            teamsDetail: mockTeams,
            selectedTeamId: null,
            isTeamDrawerOpen: false,
        });

        render(<TeamsPage />);

        // No drawer initially
        expect(screen.queryByTestId('team-detail-drawer')).toBeNull();

        // Click on team-alpha card
        fireEvent.click(screen.getByRole('button', { name: /Alpha Squad/i }));

        // The store's selectTeam should set the selectedTeamId + open the drawer
        // Since we're using real store actions, check that the drawer opens
        expect(screen.getByTestId('team-detail-drawer')).toBeDefined();
        expect(screen.getByText('Drawer: Alpha Squad')).toBeDefined();
    });

    it('renders team quick action links', () => {
        useCortexStore.setState({
            teamsDetail: mockTeams,
        });

        render(<TeamsPage />);

        expect(screen.getByTestId('team-team-alpha-open-chat')).toBeDefined();
        expect(screen.getByTestId('team-team-alpha-view-runs')).toBeDefined();
        expect(screen.getByTestId('team-team-alpha-view-wiring')).toBeDefined();
        expect(screen.getByTestId('team-team-alpha-view-logs')).toBeDefined();
    });
});
