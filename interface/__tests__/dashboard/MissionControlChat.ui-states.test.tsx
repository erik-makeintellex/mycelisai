import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat UI states', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('shows "Ask Soma..." when admin is selected', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByPlaceholderText(/Ask Soma/i)).toBeDefined();
    });

    it('shows a selected team placeholder in simple Soma mode', async () => {
        useCortexStore.setState({
            selectedTeamId: 'marketing-team',
            teamsDetail: [
                {
                    id: 'marketing-team',
                    name: 'Marketing',
                    role: 'campaigns',
                    type: 'standing',
                    mission_id: null,
                    mission_intent: null,
                    inputs: [],
                    deliveries: [],
                    agents: [],
                },
            ],
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByPlaceholderText(/Ask Soma about Marketing/i)).toBeDefined();
    });

    it('shows broadcast placeholder in broadcast mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        const broadcastBtn = screen.getByTitle(/Broadcast mode/);
        fireEvent.click(broadcastBtn);

        expect(screen.getByPlaceholderText(/Broadcast to all teams/i)).toBeDefined();
    });

    it('clears messages when trash button is clicked', async () => {
        useCortexStore.setState({
            missionChat: [
                { role: 'user', content: 'Hello' },
                { role: 'council', content: 'Hi there', source_node: 'admin' },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();
        expect(screen.getByText('Hello')).toBeDefined();

        const clearBtn = screen.getByTitle('Clear chat');
        fireEvent.click(clearBtn);

        expect(useCortexStore.getState().missionChat).toHaveLength(0);
    });

    it('hides trash button when chat is empty', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.queryByTitle('Clear chat')).toBeNull();
    });

    it('shows bouncing dots while chatting', async () => {
        useCortexStore.setState({ isMissionChatting: true });

        const { container } = render(<MissionControlChat />);
        await settleMissionControlChat();

        const dots = container.querySelectorAll('.animate-bounce');
        expect(dots.length).toBe(3);
    });

    it('disables input while loading', async () => {
        useCortexStore.setState({ isMissionChatting: true });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByRole('textbox');
        expect(input.hasAttribute('disabled')).toBe(true);
    });

    it('shows prompt about asking Soma in normal mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText(/Tell Soma what you want to plan, review, create, or execute/i)).toBeDefined();
    });

    it('turns starter prompts into clickable guided actions in simple mode', async () => {
        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByText('Choose a starter prompt')).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: 'Run a governed change' }));
        expect(screen.getByDisplayValue('Run a governed change')).toBeDefined();
    });

    it('shows team-specific starter prompts in simple mode when a team is selected', async () => {
        useCortexStore.setState({
            selectedTeamId: 'marketing-team',
            teamsDetail: [
                {
                    id: 'marketing-team',
                    name: 'Marketing',
                    role: 'campaigns',
                    type: 'standing',
                    mission_id: null,
                    mission_intent: null,
                    inputs: [],
                    deliveries: [],
                    agents: [],
                },
            ],
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByRole('button', { name: 'Plan the next move for Marketing' })).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: 'Summarize Marketing in one sentence' }));
        expect(screen.getByDisplayValue('Summarize Marketing in one sentence')).toBeDefined();
    });

    it('shows broadcast directive text in broadcast mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        const broadcastBtn = screen.getByTitle(/Broadcast mode/);
        fireEvent.click(broadcastBtn);

        expect(screen.getByText(/Broadcast directives/i)).toBeDefined();
    });
});
