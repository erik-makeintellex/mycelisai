import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    CTS_CHAT_RESPONSE,
    errorText,
    okJson,
    requestUrl,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat flow contracts', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('sends Workspace chat through the Soma route', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'admin' },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Design a new mission' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            const calls = mockFetch.mock.calls;
            const chatCall = calls.find((call: any[]) => requestUrl(call[0]).includes('/api/v1/chat'));
            expect(chatCall).toBeDefined();
        });
    });

    it('includes organization and selected team context in Soma chat requests', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
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

        mockFetch.mockImplementation(async () => okJson({
            ok: true,
            data: {
                ...CTS_CHAT_RESPONSE.data,
                meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'admin' },
            },
        }));

        render(<MissionControlChat simpleMode organizationId="org-123" />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma about Marketing/i);
        fireEvent.change(input, { target: { value: 'Plan the next move' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            const chatCall = mockFetch.mock.calls.at(-1);
            expect(chatCall).toBeDefined();

            const body = JSON.parse(String(chatCall?.[1]?.body ?? '{}'));
            expect(body.organization_id).toBe('org-123');
            expect(body.team_id).toBe('marketing-team');
            expect(body.team_name).toBe('Marketing');
        });
    });

    it('sends direct specialist chat through the targeted council route', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/council/council-architect/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'council-architect' },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();
        act(() => {
            useCortexStore.getState().setCouncilTarget('council-architect');
        });

        const input = screen.getByPlaceholderText(/Direct to Architect/i);
        fireEvent.change(input, { target: { value: 'Review the system architecture' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            const calls = mockFetch.mock.calls;
            const chatCall = calls.find((call: any[]) => requestUrl(call[0]).includes('/api/v1/council/council-architect/chat'));
            expect(chatCall).toBeDefined();
        });
    });

    it('renders user message as right-aligned bubble', async () => {
        useCortexStore.setState({
            missionChat: [{ role: 'user', content: 'Hello world' }],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

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
        await settleMissionControlChat();

        expect(screen.getByText('Hello from architect')).toBeDefined();
        expect(screen.getByText('Architect')).toBeDefined();
    });

    it('renders specialist-generated artifacts returned through Soma chat', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        payload: {
                            text: 'I prepared two sample outputs for review.',
                            ask_class: 'governed_artifact',
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
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Create two sample outputs for the homepage' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Homepage Moodboard')).toBeDefined();
            expect(screen.getByText('Creative Brief')).toBeDefined();
            expect(screen.getByTitle('Save image to workspace/saved-media')).toBeDefined();
            expect(screen.getByText('Artifact result')).toBeDefined();
            expect(screen.getByText('Soma prepared 2 artifacts for review: Homepage Moodboard and Creative Brief.')).toBeDefined();
        });
    });

    it('shows a specialist-support badge for consulted answers', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'The architect reviewed the tradeoffs and recommends the safer route.',
                    source_node: 'admin',
                    ask_class: 'specialist_consultation',
                    consultations: [
                        { member: 'council-architect', summary: 'Prefer the safer route.' },
                    ],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Specialist support')).toBeDefined();
        expect(screen.getByText('Soma checked with Architect while shaping this answer: Prefer the safer route.')).toBeDefined();
    });

    it('uses a readable fallback when Soma returns no text but includes artifacts', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
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
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Show me the current system state visually' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Soma prepared output for review below.')).toBeDefined();
            expect(screen.getByText('System Snapshot')).toBeDefined();
            expect(screen.getByText('Soma prepared 1 artifact for review: System Snapshot.')).toBeDefined();
        });
    });

    it('uses a readable fallback when Soma returns an empty answer without artifacts', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        payload: {
                            text: '',
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

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
        await settleMissionControlChat();

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
        await settleMissionControlChat();

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
        await settleMissionControlChat();

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
        await settleMissionControlChat();

        fireEvent.click(screen.getByTitle('Save image to workspace/saved-media'));

        await waitFor(() => {
            expect(screen.getByText(/Saved to:/i)).toBeDefined();
        });

        const savedLink = screen.getByRole('link', { name: 'saved-media/test.png' });
        expect(savedLink.getAttribute('href')).toBe('/api/v1/artifacts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/download');
    });

    it('shows a clickable saved path for binary file artifacts', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'The packaged audio file is ready for download.',
                    source_node: 'admin',
                    artifacts: [
                        {
                            id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
                            type: 'file',
                            title: 'campaign-voiceover.wav',
                            content_type: 'audio/wav',
                            saved_path: 'saved-media/campaign-voiceover.wav',
                        },
                    ],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const savedObjectLink = screen.getByRole('link', { name: 'saved-media/campaign-voiceover.wav' });
        expect(savedObjectLink.getAttribute('href')).toBe('/api/v1/artifacts/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb/download');
        expect(screen.getByText(/Saved object:/i)).toBeDefined();
    });
});
