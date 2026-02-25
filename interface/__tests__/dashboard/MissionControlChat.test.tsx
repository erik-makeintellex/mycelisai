import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';

// ── Helpers ───────────────────────────────────────────────────

const COUNCIL_MEMBERS = [
    { id: 'admin', role: 'admin', team: 'admin-core' },
    { id: 'council-architect', role: 'architect', team: 'council-core' },
    { id: 'council-coder', role: 'coder', team: 'council-core' },
    { id: 'council-creative', role: 'creative', team: 'council-core' },
    { id: 'council-sentry', role: 'sentry', team: 'council-core' },
];

const CTS_CHAT_RESPONSE = {
    ok: true,
    data: {
        meta: { source_node: 'admin', timestamp: '2026-02-16T12:00:00Z' },
        signal_type: 'chat_response',
        trust_score: 0.5,
        payload: { text: 'Hello from admin agent', consultations: null, tools_used: null },
    },
};

function resetStore() {
    useCortexStore.setState({
        missionChat: [],
        isMissionChatting: false,
        missionChatError: null,
        councilTarget: 'admin',
        councilMembers: [],
        isBroadcasting: false,
        lastBroadcastResult: null,
        streamLogs: [],
    });
}

// ── Tests ─────────────────────────────────────────────────────

describe('MissionControlChat', () => {
    beforeEach(() => {
        resetStore();
        // Default: council members fetch returns the members
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    // ── Header & Target Display ────────────────────────────────

    describe('Header & Target Display', () => {
        it('shows "Soma" header by default', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            // The header shows "Soma" as the default target
            expect(screen.getByText('Soma')).toBeDefined();
        });

        it('shows "Broadcast" header in broadcast mode', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            // Toggle broadcast mode via the Megaphone button
            const broadcastBtn = screen.getByTitle(/Broadcast mode/);
            fireEvent.click(broadcastBtn);

            expect(screen.getByText('Broadcast')).toBeDefined();
        });

        it('shows Direct council button for targeting specific members', async () => {
            useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            // "Direct" button exists for targeting specific council members
            expect(screen.getByText('Direct')).toBeDefined();
        });

        it('shows Soma header when exiting broadcast mode', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const broadcastBtn = screen.getByTitle(/Broadcast mode/);
            fireEvent.click(broadcastBtn); // ON
            fireEvent.click(broadcastBtn); // OFF

            expect(screen.getByText('Soma')).toBeDefined();
        });
    });

    // ── Chat Flow ─────────────────────────────────────────────

    describe('Chat Flow', () => {
        it('sends message to council endpoint with selected target', async () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
                councilTarget: 'admin',
            });

            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: COUNCIL_MEMBERS }) })
                .mockResolvedValueOnce({
                    ok: true,
                    json: async () => ({
                        ok: true,
                        data: {
                            ...CTS_CHAT_RESPONSE.data,
                            meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'admin' },
                        },
                    }),
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'Design a new mission' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                const calls = mockFetch.mock.calls;
                const chatCall = calls.find((c: any[]) =>
                    typeof c[0] === 'string' && c[0].includes('/council/admin/chat')
                );
                expect(chatCall).toBeDefined();
            });
        });

        it('renders user message as right-aligned bubble', async () => {
            useCortexStore.setState({
                missionChat: [{ role: 'user', content: 'Hello world' }],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Hello world')).toBeDefined();
        });

        it('renders council response with source label', async () => {
            useCortexStore.setState({
                missionChat: [
                    { role: 'user', content: 'Hi' },
                    {
                        role: 'council',
                        content: 'Hello from architect',
                        source_node: 'council-architect',
                        trust_score: 0.5,
                    },
                ],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            expect(screen.getByText('Hello from architect')).toBeDefined();
            expect(screen.getByText('Architect')).toBeDefined();
        });

        it('renders trust badge with correct score', async () => {
            useCortexStore.setState({
                missionChat: [
                    {
                        role: 'council',
                        content: 'Response',
                        source_node: 'admin',
                        trust_score: 0.5,
                    },
                ],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('C:0.5')).toBeDefined();
        });

        it('renders tools-used pills when present', async () => {
            useCortexStore.setState({
                missionChat: [
                    {
                        role: 'council',
                        content: 'I searched memory',
                        source_node: 'admin',
                        trust_score: 0.5,
                        tools_used: ['search_memory', 'list_teams'],
                    },
                ],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Search Memory')).toBeDefined();
            expect(screen.getByText('View Teams')).toBeDefined();
        });

        it('does not render tools pills when tools_used is empty', async () => {
            useCortexStore.setState({
                missionChat: [
                    {
                        role: 'council',
                        content: 'Simple response',
                        source_node: 'admin',
                        trust_score: 0.5,
                        tools_used: [],
                    },
                ],
            });

            const { container } = render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            // No tool pills should be rendered
            const pills = container.querySelectorAll('[class*="cortex-primary/10"]');
            expect(pills).toHaveLength(0);
        });
    });

    // ── Dynamic Placeholder ───────────────────────────────────

    describe('Dynamic Placeholder', () => {
        it('shows "Ask Soma..." when admin is selected', async () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
                councilTarget: 'admin',
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByPlaceholderText(/Ask Soma/i)).toBeDefined();
        });

        it('shows broadcast placeholder in broadcast mode', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const broadcastBtn = screen.getByTitle(/Broadcast mode/);
            fireEvent.click(broadcastBtn);

            expect(screen.getByPlaceholderText(/Broadcast to all teams/i)).toBeDefined();
        });
    });

    // ── Error States ──────────────────────────────────────────

    describe('Error States', () => {
        it('displays error bar when missionChatError is set', async () => {
            useCortexStore.setState({
                missionChatError: 'Swarm offline',
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Swarm offline')).toBeDefined();
        });

        it('shows error as chat bubble with source_node on API failure', async () => {
            useCortexStore.setState({
                councilTarget: 'council-architect',
                missionChat: [
                    {
                        role: 'council',
                        content: 'Council member council-architect did not respond',
                        source_node: 'council-architect',
                    },
                ],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText(/did not respond/)).toBeDefined();
            expect(screen.getByText('Architect')).toBeDefined();
        });
    });

    // ── Clear Chat ────────────────────────────────────────────

    describe('Clear Chat', () => {
        it('clears messages when trash button is clicked', async () => {
            useCortexStore.setState({
                missionChat: [
                    { role: 'user', content: 'Hello' },
                    { role: 'council', content: 'Hi there', source_node: 'admin' },
                ],
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Hello')).toBeDefined();

            const clearBtn = screen.getByTitle('Clear chat');
            fireEvent.click(clearBtn);

            expect(useCortexStore.getState().missionChat).toHaveLength(0);
        });

        it('hides trash button when chat is empty', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.queryByTitle('Clear chat')).toBeNull();
        });
    });

    // ── Loading State ─────────────────────────────────────────

    describe('Loading State', () => {
        it('shows bouncing dots while chatting', async () => {
            useCortexStore.setState({ isMissionChatting: true });

            const { container } = render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            const dots = container.querySelectorAll('.animate-bounce');
            expect(dots.length).toBe(3);
        });

        it('disables input while loading', async () => {
            useCortexStore.setState({ isMissionChatting: true });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            const input = screen.getByRole('textbox');
            expect(input.hasAttribute('disabled')).toBe(true);
        });
    });

    // ── Empty State ───────────────────────────────────────────

    describe('Empty State', () => {
        it('shows prompt about asking Soma in normal mode', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText(/Ask Soma/i)).toBeDefined();
        });

        it('shows broadcast directive text in broadcast mode', async () => {
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const broadcastBtn = screen.getByTitle(/Broadcast mode/);
            fireEvent.click(broadcastBtn);

            expect(screen.getByText(/Broadcast directives/i)).toBeDefined();
        });
    });
});
