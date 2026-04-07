import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';

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
        expect(screen.getByText(/focused lead entity first/i)).toBeDefined();
        expect(screen.getByText(/coordinate back through Soma/i)).toBeDefined();
    });
});
