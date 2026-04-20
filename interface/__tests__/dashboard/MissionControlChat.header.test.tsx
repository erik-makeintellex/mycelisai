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

        expect(screen.getByText('Soma conversation')).toBeDefined();
        expect(screen.queryByText('Direct')).toBeNull();
        expect(screen.queryByTitle(/Broadcast mode/)).toBeNull();
        expect(screen.getByPlaceholderText(/Tell Soma what you want to plan, review, create, or execute/i)).toBeDefined();
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
