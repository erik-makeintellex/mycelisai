import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { SomaOfflineGuide } from '@/components/dashboard/MissionControlChatStates';
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

    it('uses source-mode recovery commands in the offline guide', () => {
        render(<SomaOfflineGuide assistantName="Soma" onRetry={vi.fn()} />);

        expect(screen.getByText(/Your Mycelis runtime isn't running yet/i)).toBeDefined();
        expect(screen.getByText('Source-mode recovery')).toBeDefined();
        expect(screen.getByText('uv run inv native-infra.status')).toBeDefined();
        expect(screen.getByText('uv run inv native-infra.up')).toBeDefined();
        expect(screen.getByText('uv run inv db.migrate')).toBeDefined();
        expect(screen.getByText('uv run inv lifecycle.up --frontend')).toBeDefined();
        expect(screen.queryByText(/neural organism/i)).toBeNull();
        expect(screen.queryByText('uv run inv lifecycle.up')).toBeNull();
        expect(screen.queryByText('inv lifecycle.up')).toBeNull();
    });

    it('keeps the message history in the bounded scroll region', async () => {
        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByTestId('mission-chat').className).toContain('min-h-0');
        expect(screen.getByTestId('soma-conversation-thread').className).toContain('min-h-0');
        expect(screen.getByTestId('soma-conversation-thread').className).toContain('overflow-y-auto');
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

        expect(screen.getByText(/Tell Soma what you want to plan, review, create, or run/i)).toBeDefined();
    });

    it('turns starter prompts into clickable guided actions in simple mode', async () => {
        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByText('Start with Soma')).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: /Research something.*Search or review sources/i }));
        expect(screen.getByDisplayValue('Research this, cite sources, and tell me what changed.')).toBeDefined();
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

        expect(screen.getByRole('button', { name: /Plan this team.*Marketing/i })).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: /Review state.*current risk/i }));
        expect(screen.getByDisplayValue('Review the current state of Marketing')).toBeDefined();
    });

    it('shows broadcast directive text in broadcast mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        const broadcastBtn = screen.getByTitle(/Broadcast mode/);
        fireEvent.click(broadcastBtn);

        expect(screen.getByText(/Broadcast directives/i)).toBeDefined();
    });
});
