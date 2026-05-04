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
});
