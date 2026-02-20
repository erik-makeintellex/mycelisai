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
    });
}

// ── Tests ─────────────────────────────────────────────────────

describe('MissionControlChat', () => {
    beforeEach(() => {
        resetStore();
    });

    // ── Council Selector ──────────────────────────────────────

    describe('Council Member Selector', () => {
        it('shows fallback "Admin" when members not loaded', () => {
            mockFetch.mockResolvedValue({ ok: false });
            render(<MissionControlChat />);

            const select = screen.getByRole('combobox');
            expect(select).toBeDefined();
            expect(select.querySelector('option')?.textContent).toContain('Soma');
        });

        it('populates dropdown with council members from API', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
            });

            render(<MissionControlChat />);

            await waitFor(() => {
                const options = screen.getAllByRole('option');
                expect(options).toHaveLength(5);
            });

            const options = screen.getAllByRole('option');
            expect(options[0].textContent).toContain('Soma');
            expect(options[1].textContent).toContain('Architect');
        });

        it('changes councilTarget when selecting a different member', async () => {
            // Pre-populate members
            useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });

            render(<MissionControlChat />);

            const select = screen.getByRole('combobox');
            fireEvent.change(select, { target: { value: 'council-architect' } });

            expect(useCortexStore.getState().councilTarget).toBe('council-architect');
        });

        it('hides selector in broadcast mode', () => {
            useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });
            render(<MissionControlChat />);

            // Toggle broadcast
            const broadcastBtn = screen.getByTitle(/Toggle broadcast/);
            fireEvent.click(broadcastBtn);

            // Selector should be gone, "Broadcast" label should appear
            expect(screen.queryByRole('combobox')).toBeNull();
            expect(screen.getByText('Broadcast')).toBeDefined();
        });

        it('restores selector when exiting broadcast mode', () => {
            useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS });
            render(<MissionControlChat />);

            const broadcastBtn = screen.getByTitle(/Toggle broadcast/);
            fireEvent.click(broadcastBtn); // ON
            fireEvent.click(broadcastBtn); // OFF

            expect(screen.getByRole('combobox')).toBeDefined();
        });
    });

    // ── Chat Flow ─────────────────────────────────────────────

    describe('Chat Flow', () => {
        it('sends message to council endpoint with selected target', async () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
                councilTarget: 'council-architect',
            });

            // First call: fetchCouncilMembers, second: sendMissionChat
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: COUNCIL_MEMBERS }) })
                .mockResolvedValueOnce({
                    ok: true,
                    json: async () => ({
                        ok: true,
                        data: {
                            ...CTS_CHAT_RESPONSE.data,
                            meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'council-architect' },
                        },
                    }),
                });

            render(<MissionControlChat />);

            const input = screen.getByPlaceholderText(/Ask Architect/i);
            fireEvent.change(input, { target: { value: 'Design a new mission' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                // Verify fetch was called with council endpoint
                const calls = mockFetch.mock.calls;
                const chatCall = calls.find((c: any[]) =>
                    typeof c[0] === 'string' && c[0].includes('/council/council-architect/chat')
                );
                expect(chatCall).toBeDefined();
            });
        });

        it('renders user message as right-aligned bubble', async () => {
            useCortexStore.setState({
                missionChat: [{ role: 'user', content: 'Hello world' }],
            });

            render(<MissionControlChat />);
            expect(screen.getByText('Hello world')).toBeDefined();
        });

        it('renders council response with source label', () => {
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

            expect(screen.getByText('Hello from architect')).toBeDefined();
            expect(screen.getByText('Architect')).toBeDefined();
        });

        it('renders trust badge with correct score', () => {
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
            expect(screen.getByText('C:0.5')).toBeDefined();
        });

        it('renders tools-used pills when present', () => {
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
            expect(screen.getByText('Search Memory')).toBeDefined();
            expect(screen.getByText('View Teams')).toBeDefined();
        });

        it('does not render tools pills when tools_used is empty', () => {
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
            // No tool pills should be rendered
            const pills = container.querySelectorAll('[class*="cortex-primary/10"]');
            expect(pills).toHaveLength(0);
        });
    });

    // ── Dynamic Placeholder ───────────────────────────────────

    describe('Dynamic Placeholder', () => {
        it('shows "Ask Soma..." when admin is selected', () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
                councilTarget: 'admin',
            });

            render(<MissionControlChat />);
            expect(screen.getByPlaceholderText(/Ask Soma/i)).toBeDefined();
        });

        it('shows "Ask Architect..." when architect is selected', () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
                councilTarget: 'council-architect',
            });

            render(<MissionControlChat />);
            expect(screen.getByPlaceholderText(/Ask Architect/i)).toBeDefined();
        });

        it('shows broadcast placeholder in broadcast mode', () => {
            render(<MissionControlChat />);

            const broadcastBtn = screen.getByTitle(/Toggle broadcast/);
            fireEvent.click(broadcastBtn);

            expect(screen.getByPlaceholderText(/Broadcast to all teams/i)).toBeDefined();
        });
    });

    // ── Error States ──────────────────────────────────────────

    describe('Error States', () => {
        it('displays error bar when missionChatError is set', () => {
            useCortexStore.setState({
                missionChatError: 'Swarm offline',
            });

            render(<MissionControlChat />);
            expect(screen.getByText('Swarm offline')).toBeDefined();
        });

        it('shows error as chat bubble with source_node on API failure', () => {
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
            expect(screen.getByText(/did not respond/)).toBeDefined();
            expect(screen.getByText('Architect')).toBeDefined();
        });
    });

    // ── Clear Chat ────────────────────────────────────────────

    describe('Clear Chat', () => {
        it('clears messages when trash button is clicked', () => {
            useCortexStore.setState({
                missionChat: [
                    { role: 'user', content: 'Hello' },
                    { role: 'council', content: 'Hi there', source_node: 'admin' },
                ],
            });

            render(<MissionControlChat />);
            expect(screen.getByText('Hello')).toBeDefined();

            const clearBtn = screen.getByTitle('Clear chat');
            fireEvent.click(clearBtn);

            expect(useCortexStore.getState().missionChat).toHaveLength(0);
        });

        it('hides trash button when chat is empty', () => {
            render(<MissionControlChat />);
            expect(screen.queryByTitle('Clear chat')).toBeNull();
        });
    });

    // ── Loading State ─────────────────────────────────────────

    describe('Loading State', () => {
        it('shows bouncing dots while chatting', () => {
            useCortexStore.setState({ isMissionChatting: true });

            const { container } = render(<MissionControlChat />);
            const dots = container.querySelectorAll('.animate-bounce');
            expect(dots.length).toBe(3);
        });

        it('disables input while loading', () => {
            useCortexStore.setState({ isMissionChatting: true });

            render(<MissionControlChat />);
            const input = screen.getByRole('textbox');
            expect(input.hasAttribute('disabled')).toBe(true);
        });
    });

    // ── Empty State ───────────────────────────────────────────

    describe('Empty State', () => {
        it('shows shield icon and prompt in normal mode', () => {
            render(<MissionControlChat />);
            expect(screen.getByText(/Ask Soma/i)).toBeDefined();
        });

        it('shows megaphone icon in broadcast mode', () => {
            render(<MissionControlChat />);

            const broadcastBtn = screen.getByTitle(/Toggle broadcast/);
            fireEvent.click(broadcastBtn);

            expect(screen.getByText(/Broadcast directives/i)).toBeDefined();
        });
    });
});
