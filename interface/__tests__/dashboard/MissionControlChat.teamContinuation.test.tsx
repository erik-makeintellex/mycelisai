import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => import('../mocks/reactflow'));

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    okJson,
    resetMissionControlChatStore,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat team continuation', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue(okJson({ ok: true, data: COUNCIL_MEMBERS }));
    });

    it('prompts the operator to start work after a team-only creation run', async () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Mission activated',
                mode: 'execution_result',
                run_id: 'run-team-123456',
                execution_summary: {
                    understanding: {
                        summary: 'Team created. No work item has started yet.',
                    },
                    execution: {
                        shape: 'guided_proposal',
                        status: 'completed',
                        summary: 'Soma created the governed team and recorded proof.',
                    },
                    capability_use: {
                        capabilities: ['create_team'],
                    },
                    outputs: [{
                        id: 'snes-style-browser-game-team',
                        kind: 'team',
                        title: 'SNES-Style Browser Game Team',
                        href: '/groups',
                        retained: true,
                    }],
                    proof: [{ run_id: 'run-team-123456' }],
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(await screen.findByText('SNES-Style Browser Game Team is ready. Choose the first deliverable.')).toBeDefined();
        expect(screen.getByRole('button', { name: /Write design brief/i })).toBeDefined();
        expect(screen.getByRole('button', { name: /Draft delivery plan/i })).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: /Build playable prototype/i }));

        expect((screen.getByPlaceholderText(/Tell Soma/i) as HTMLInputElement).value)
            .toBe("Have SNES-Style Browser Game Team build the first playable browser-game prototype as a reviewable project package. Save it in the team's group folder with README, validation notes, output link, and proof.");
    });
});
