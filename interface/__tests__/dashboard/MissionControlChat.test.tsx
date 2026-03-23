import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
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
        missionChatFailure: null,
        activeMode: 'answer',
        activeRole: '',
        assistantName: 'Soma',
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

        it('shows custom assistant name from settings', async () => {
            useCortexStore.setState({ assistantName: 'Atlas' });
            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Atlas')).toBeDefined();
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
        it('sends Workspace chat through the Soma route', async () => {
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
                    typeof c[0] === 'string' && c[0].includes('/api/v1/chat')
                );
                expect(chatCall).toBeDefined();
            });
        });

        it('sends direct specialist chat through the targeted council route', async () => {
            useCortexStore.setState({
                councilMembers: COUNCIL_MEMBERS,
            });

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
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            act(() => {
                useCortexStore.getState().setCouncilTarget('council-architect');
            });

            const input = screen.getByPlaceholderText(/Direct to Architect/i);
            fireEvent.change(input, { target: { value: 'Review the system architecture' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                const calls = mockFetch.mock.calls;
                const chatCall = calls.find((c: any[]) =>
                    typeof c[0] === 'string' && c[0].includes('/api/v1/council/council-architect/chat')
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

        it('renders specialist-generated artifacts returned through Soma chat', async () => {
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
                            payload: {
                                text: 'I prepared two sample outputs for review.',
                                artifacts: [
                                    {
                                        id: 'img-1',
                                        type: 'image',
                                        title: 'Homepage Moodboard',
                                        content_type: 'image/png',
                                        content: 'cG5n',
                                        cached: true,
                                    },
                                    {
                                        id: 'doc-1',
                                        type: 'document',
                                        title: 'Creative Brief',
                                        content_type: 'text/markdown',
                                        content: '# Brief',
                                    },
                                ],
                            },
                        },
                    }),
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'Create two sample outputs for the homepage' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(screen.getByText('Homepage Moodboard')).toBeDefined();
                expect(screen.getByText('Creative Brief')).toBeDefined();
                expect(screen.getByTitle('Save image to workspace/saved-media')).toBeDefined();
            });
        });

        it('uses a readable fallback when Soma returns no text but includes artifacts', async () => {
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
                            payload: {
                                text: '',
                                artifacts: [
                                    {
                                        id: 'img-2',
                                        type: 'image',
                                        title: 'System Snapshot',
                                        content_type: 'image/png',
                                        content: 'cG5n',
                                    },
                                ],
                            },
                        },
                    }),
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'Show me the current system state visually' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(screen.getByText('Soma prepared output for review below.')).toBeDefined();
                expect(screen.getByText('System Snapshot')).toBeDefined();
            });
        });

        it('uses a readable fallback when Soma returns an empty answer without artifacts', async () => {
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
                            payload: {
                                text: '',
                            },
                        },
                    }),
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'Any organizations launched?' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(screen.getByText(/could not produce a readable reply/i)).toBeDefined();
            });
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

        it('saves cached image artifact to workspace folder from inline card', async () => {
            useCortexStore.setState({
                missionChat: [
                    {
                        role: 'council',
                        content: 'Generated image',
                        source_node: 'admin',
                        artifacts: [
                            {
                                id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
                                type: 'image',
                                title: 'Generated: test',
                                content_type: 'image/png',
                                content: 'cG5n',
                                cached: true,
                            },
                        ],
                    },
                ],
            });

            mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
                const url = String(input);
                if (url.includes('/api/v1/artifacts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/save')) {
                    return {
                        ok: true,
                        json: async () => ({ id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', file_path: 'saved-media/test.png' }),
                    } as any;
                }
                return {
                    ok: true,
                    json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
                } as any;
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            fireEvent.click(screen.getByTitle('Save image to workspace/saved-media'));

            await waitFor(() => {
                expect(screen.getByText(/Saved to:/i)).toBeDefined();
            });
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
        it('displays a Soma-specific blocker card when Workspace chat is blocked', async () => {
            useCortexStore.setState({
                councilTarget: 'admin',
                missionChatError: 'Soma chat blocked (500)',
                missionChatFailure: buildMissionChatFailure({
                    assistantName: 'Soma',
                    targetId: 'admin',
                    message: 'Soma chat blocked (500)',
                    statusCode: 500,
                }),
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            expect(screen.getByText('Soma Chat Blocked')).toBeDefined();
            expect(screen.queryByText('Switch to Soma')).toBeNull();
        });

        it('displays structured council error card when missionChatError is set', async () => {
            useCortexStore.setState({
                missionChatError: 'Swarm offline',
                missionChatFailure: buildMissionChatFailure({
                    assistantName: 'Soma',
                    targetId: 'council-architect',
                    message: 'Swarm offline',
                }),
            });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            act(() => {
                useCortexStore.getState().setCouncilTarget('council-architect');
            });
            expect(screen.getByText('Council Call Failed')).toBeDefined();
            expect(screen.getByText('Copy Diagnostics')).toBeDefined();
        });

        it('records blocker mode when Soma chat request fails', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: COUNCIL_MEMBERS }) })
                .mockResolvedValueOnce({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Soma chat blocked (500)"}',
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'hello' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(useCortexStore.getState().activeMode).toBe('blocker');
            });
            expect(screen.getByText('Soma Chat Blocked')).toBeDefined();
        });

        it('records Soma blocker mode when council roster is unavailable', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: false })
                .mockResolvedValueOnce({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Soma chat blocked (500)"}',
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });

            const input = screen.getByPlaceholderText(/Ask Soma/i);
            fireEvent.change(input, { target: { value: 'hello' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(useCortexStore.getState().activeMode).toBe('blocker');
            });
            expect(screen.getByText('Soma Chat Blocked')).toBeDefined();
            expect(screen.queryByText('Switch to Soma')).toBeNull();
            expect(
                mockFetch.mock.calls.some((c: any[]) => typeof c[0] === 'string' && c[0].includes('/api/v1/chat'))
            ).toBe(true);
        });

        it('records council blocker mode and shows Soma fallback actions when direct council chat fails', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: COUNCIL_MEMBERS }) })
                .mockResolvedValueOnce({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Council member failed"}',
                });

            render(<MissionControlChat />);
            await act(async () => { await new Promise((r) => setTimeout(r, 0)); });
            act(() => {
                useCortexStore.getState().setCouncilTarget('council-sentry');
            });

            const input = screen.getByPlaceholderText(/Direct to Sentry/i);
            fireEvent.change(input, { target: { value: 'hello sentry' } });
            fireEvent.keyDown(input, { key: 'Enter' });

            await waitFor(() => {
                expect(useCortexStore.getState().activeMode).toBe('blocker');
            });
            expect(screen.getByText('Council Call Failed')).toBeDefined();
            expect(screen.getByText('Switch to Soma')).toBeDefined();
            expect(screen.getByText('Continue with Soma Only')).toBeDefined();
            expect(
                mockFetch.mock.calls.some((c: any[]) => typeof c[0] === 'string' && c[0].includes('/api/v1/council/council-sentry/chat'))
            ).toBe(true);
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
