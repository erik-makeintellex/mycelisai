import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    errorText,
    okJson,
    requestUrl,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat error states', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

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
        await settleMissionControlChat();

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
        await settleMissionControlChat();
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
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'hello' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(useCortexStore.getState().activeMode).toBe('blocker');
        });
        expect(screen.getByText('Soma Chat Blocked')).toBeDefined();
    });

    it('renders a readable fallback instead of raw council tool JSON in Soma chat', async () => {
        const rawCouncilPayload = '{"error":"consult_council requires \\"member\\" and \\"question\\"","tool":"consult_council"}';

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-02-16T12:00:00Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.5,
                        payload: {
                            text: rawCouncilPayload,
                            ask_class: 'direct_answer',
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'summarize the current state' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText(/could not produce a readable reply/i)).toBeDefined();
        });
        expect(screen.queryByText(rawCouncilPayload)).toBeNull();
        expect(screen.queryByText(/consult_council requires/i)).toBeNull();
    });

    it('records Soma blocker mode when council roster is unavailable', async () => {
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return { ok: false, status: 503, text: async () => 'unavailable' } as any;
            }
            if (url.includes('/api/v1/chat')) {
                return errorText(500, '{"error":"Soma chat blocked (500)"}');
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'hello' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(useCortexStore.getState().activeMode).toBe('blocker');
        });
        expect(screen.getByText('Soma Chat Blocked')).toBeDefined();
        expect(screen.queryByText('Switch to Soma')).toBeNull();
        expect(mockFetch.mock.calls.some((call: any[]) => requestUrl(call[0]).includes('/api/v1/chat'))).toBe(true);
    });

    it('records council blocker mode and shows Soma fallback actions when direct council chat fails', async () => {
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/council/council-sentry/chat')) {
                return errorText(500, '{"error":"Council member failed"}');
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();
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
        expect(mockFetch.mock.calls.some((call: any[]) => requestUrl(call[0]).includes('/api/v1/council/council-sentry/chat'))).toBe(true);
    });

    it('hides raw council transport strings behind the blocker contract', async () => {
        const rawCouncilFailure = "consult_council requires 'member' and 'question'";

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/council/council-sentry/chat')) {
                return errorText(500, rawCouncilFailure);
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();
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
        expect(screen.queryByText(rawCouncilFailure)).toBeNull();
        expect(screen.queryByText(/consult_council requires/i)).toBeNull();
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
        await settleMissionControlChat();

        expect(screen.getByText(/did not respond/)).toBeDefined();
        expect(screen.getByText('Architect')).toBeDefined();
    });
});
