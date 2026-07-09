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

describe('MissionControlChat header and routing chrome', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('shows "Soma" header by default', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Soma')).toBeDefined();
    });

    it('shows custom assistant name from settings', async () => {
        useCortexStore.setState({ assistantName: 'Atlas' });
        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Atlas')).toBeDefined();
    });

    it('rehydrates organization-scoped chat history when organizationId is provided', async () => {
        localStorage.setItem('mycelis-workspace-chat:org-123', JSON.stringify([
            { role: 'user', content: 'Persisted org message' },
        ]));

        render(<MissionControlChat simpleMode organizationId="org-123" />);
        await settleMissionControlChat();

        expect(useCortexStore.getState().workspaceChatScope).toBe('org-123');
        expect(screen.getByText('Persisted org message')).toBeDefined();
    });

    it('renders malformed persisted chat without crashing or emitting React key warnings', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
        const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
        localStorage.setItem('mycelis-workspace-chat:org-123', JSON.stringify([
            { role: 'user', content: { text: 'legacy object prompt' } },
            {
                role: 'architect',
                source_node: 'admin',
                content: 'Legacy result with duplicate fields',
                tools_used: ['web_search', 'web_search'],
                proposal: {
                    intent: 'legacy proposal',
                    risk_level: 'medium',
                    confirm_token: 'token-1',
                    intent_proof_id: 'proof-1',
                    tools: ['delegate_task', 'delegate_task'],
                    affected_resources: ['team:Game Team', 'team:Game Team'],
                },
                execution_summary: {
                    execution: { status: 'complete', shape: 'team_execution', summary: 'legacy complete' },
                    understanding: { summary: 'legacy', assumptions: 'not an array' },
                    outputs: { label: 'not an array' },
                    proof: [
                        { label: 'Audit', run_id: 'run-1' },
                        { label: 'Audit', run_id: 'run-1' },
                    ],
                    capability_use: {
                        tools: 'not an array',
                        used: ['web_search', 'web_search'],
                    },
                },
            },
            { role: 'not-a-role', content: 'drop me' },
        ]));

        try {
            render(<MissionControlChat simpleMode organizationId="org-123" />);
            await settleMissionControlChat();

            expect(screen.getByText(/legacy object prompt/i)).toBeDefined();
            expect(screen.getByText(/Legacy result with duplicate fields/i)).toBeDefined();
            expect(screen.queryByText('drop me')).toBeNull();
            expect(consoleError).not.toHaveBeenCalled();
            expect(consoleWarn).not.toHaveBeenCalled();
        } finally {
            consoleError.mockRestore();
            consoleWarn.mockRestore();
        }
    });

    it('rehydrates team-scoped chat history when a focused team is provided', async () => {
        localStorage.setItem('mycelis-workspace-chat:org-123::team::team-alpha', JSON.stringify([
            { role: 'user', content: 'Persisted focused team message' },
        ]));
        useCortexStore.setState({
            teamsDetail: [{
                id: 'team-alpha',
                name: 'Alpha Team',
                role: 'delivery',
                type: 'mission',
                mission_id: null,
                mission_intent: null,
                inputs: [],
                deliveries: [],
                agents: [],
            }],
        });

        render(<MissionControlChat simpleMode organizationId="org-123" focusedTeamId="team-alpha" />);
        await settleMissionControlChat();

        expect(useCortexStore.getState().workspaceChatScope).toBe('org-123::team::team-alpha');
        expect(screen.getByText('Persisted focused team message')).toBeDefined();
        expect(screen.getByText(/Team chat for Alpha Team/i)).toBeDefined();
        expect(screen.getByText(/Soma can still reference other work when you ask/i)).toBeDefined();
    });

    it('clears all persisted chat scopes when opened with fresh reset flag', async () => {
        localStorage.setItem('mycelis-workspace-chat', JSON.stringify([
            { role: 'user', content: 'Persisted root message' },
        ]));
        localStorage.setItem('mycelis-workspace-chat:org-123::team::team-alpha', JSON.stringify([
            { role: 'user', content: 'Persisted team message' },
        ]));
        localStorage.setItem('mycelis-workspace-chat-session:org-123::team::team-alpha', 'session-1');
        window.history.pushState(null, '', '/dashboard?fresh=1');

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(localStorage.getItem('mycelis-workspace-chat')).toBeNull();
        expect(localStorage.getItem('mycelis-workspace-chat:org-123::team::team-alpha')).toBeNull();
        expect(localStorage.getItem('mycelis-workspace-chat-session:org-123::team::team-alpha')).toBeNull();
        expect(window.location.search).toBe('');
        expect(screen.queryByText('Persisted root message')).toBeNull();
        expect(screen.queryByText('Persisted team message')).toBeNull();
    });

    it('clears a stale loading lock when a new organization scope is applied', async () => {
        localStorage.setItem('mycelis-workspace-chat:org-123', JSON.stringify([
            { role: 'user', content: 'Persisted org message' },
        ]));
        useCortexStore.setState({
            isMissionChatting: true,
            missionChatError: 'stale loading state',
        });

        render(<MissionControlChat simpleMode organizationId="org-123" />);
        await settleMissionControlChat();

        const input = screen.getByRole('textbox');
        expect(useCortexStore.getState().workspaceChatScope).toBe('org-123');
        expect(useCortexStore.getState().isMissionChatting).toBe(false);
        expect(input.hasAttribute('disabled')).toBe(false);
    });

    it('hides advanced routing controls in simple Soma mode', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByText('Soma')).toBeDefined();
        expect(screen.getByText(/Ask for plans, changes, files, decisions, or follow-up work/i)).toBeDefined();
        expect(screen.queryByText('Direct')).toBeNull();
        expect(screen.queryByTitle(/Broadcast mode/)).toBeNull();
        expect(screen.getByPlaceholderText(/Tell Soma what you want to plan, review, create, or run/i)).toBeDefined();
    });

    it('shows "Broadcast" header in broadcast mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        const broadcastBtn = screen.getByTitle(/Broadcast mode/);
        fireEvent.click(broadcastBtn);

        expect(screen.getByText('Broadcast')).toBeDefined();
    });

    it('shows Direct council button for targeting specific members', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Direct')).toBeDefined();
    });

    it('shows Soma header when exiting broadcast mode', async () => {
        render(<MissionControlChat />);
        await settleMissionControlChat();

        const broadcastBtn = screen.getByTitle(/Broadcast mode/);
        fireEvent.click(broadcastBtn);
        fireEvent.click(broadcastBtn);

        expect(screen.getByText('Soma')).toBeDefined();
    });
});
