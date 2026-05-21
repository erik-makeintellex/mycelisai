import { describe, it, expect } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

import TeamDetailDrawer from '@/components/teams/TeamDetailDrawer';

const team = {
    id: 'marketing-team',
    name: 'Marketing',
    role: 'action',
    type: 'standing' as const,
    mission_id: null,
    mission_intent: null,
    inputs: ['campaign.requests'],
    deliveries: ['campaign.briefs'],
    agents: [
        {
            id: 'marketing-lead-agent',
            role: 'lead',
            status: 1,
            last_heartbeat: new Date().toISOString(),
            tools: ['recall'],
            model: 'qwen',
            system_prompt: 'Lead the marketing lane.',
        },
    ],
};

describe('TeamDetailDrawer', () => {
    it('explains the focused lead counterpart for the selected team', () => {
        render(<TeamDetailDrawer team={team} onClose={() => undefined} />);

        expect(screen.getByText('Primary counterpart')).toBeDefined();
        expect(screen.getByText('Marketing lead')).toBeDefined();
        expect(screen.getByText(/work starts with the lead/i)).toBeDefined();
        expect(screen.getByText(/agents stay summarized/i)).toBeDefined();
    });

    it('surfaces operator controls for the selected team inspector', () => {
        render(<TeamDetailDrawer team={team} onClose={() => undefined} />);

        expect(screen.getByText('Operator controls')).toBeDefined();
        expect(screen.getByRole('link', { name: 'Open lead workspace' }).getAttribute('href')).toBe('/dashboard?team_id=marketing-team');
        expect(screen.getByRole('link', { name: 'View runs' }).getAttribute('href')).toBe('/runs');
        expect(screen.getByRole('link', { name: 'View outputs' }).getAttribute('href')).toBe('/groups');
        expect(screen.getByText('Advanced coordination topics')).toBeDefined();
        expect(screen.getByRole('link', { name: 'View wiring' }).getAttribute('href')).toBe('/automations?tab=wiring');
        expect(screen.getByRole('link', { name: 'View system' }).getAttribute('href')).toBe('/system?tab=services');
    });

    it('summarizes larger agent rosters until the operator expands them', () => {
        const largeTeam = {
            ...team,
            agents: Array.from({ length: 7 }, (_, index) => ({
                id: `agent-${index + 1}`,
                role: index === 0 ? 'lead' : 'specialist',
                status: 1,
                last_heartbeat: new Date().toISOString(),
                tools: [],
                model: 'qwen',
            })),
        };

        render(<TeamDetailDrawer team={largeTeam} onClose={() => undefined} />);

        expect(screen.getByText('Team members (7) | Agent Roster')).toBeDefined();
        expect(screen.queryByText('agent-7')).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: 'Show 1 more members' }));

        expect(screen.getByText('agent-7')).toBeDefined();
    });
});
