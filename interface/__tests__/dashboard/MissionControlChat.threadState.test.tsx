import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => import('../mocks/reactflow'));

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

describe('MissionControlChat thread state cards', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue(okJson({ ok: true, data: COUNCIL_MEMBERS }));
    });

    it('keeps proposal route metadata inside the approval card details', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS, councilTarget: 'admin' });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    template_id: 'chat-to-proposal',
                    mode: 'proposal',
                    ui_response_state: {
                        kind: 'awaiting_approval',
                        label: 'Waiting for executive review',
                        detail: 'Review the generated action card before Soma starts the background work.',
                        tone: 'warning',
                    },
                    payload: {
                        text: 'I prepared a governed action card.',
                        proposal: {
                            intent: 'build-review-package',
                            tools: ['delegate_task'],
                            risk_level: 'medium',
                            confirm_token: 'ct-structured',
                            intent_proof_id: 'proof-structured',
                            bus_scope: 'current_team',
                            nats_subjects: ['swarm.team.ops.internal.command'],
                        },
                    },
                },
            });
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Build the review package' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('I can start that.')).toBeDefined();
            expect(screen.queryByTestId('soma-thread-state-card')).toBeNull();
            expect(screen.queryByText('Current team route')).toBeNull();
        });

        expect(screen.queryByText('swarm.team.ops.internal.command')).toBeNull();
    });

    it('renders compact typed thread events without exposing transport subjects', () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Execution started - Soma accepted the approved work.',
                mode: 'execution_result',
                run_id: 'run-thread-123',
                thread_events: [{
                    kind: 'execution_started',
                    label: 'Execution started',
                    detail: 'Soma handed this to the work bus. You can keep talking here while work continues.',
                    tone: 'info',
                    status: 'running',
                    target_reference: 'run:run-thread-123',
                    source_kind: 'web_api',
                    source_channel: 'api.intent.confirm-action',
                    payload_kind: 'soma_thread_event',
                }],
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getByTestId('soma-thread-state-card')).toBeDefined();
        expect(screen.getByText('Execution started')).toBeDefined();
        expect(screen.getByText('running')).toBeDefined();
        expect(screen.getByText('Soma handed this to the work bus. You can keep talking here while work continues.')).toBeDefined();
        expect(screen.getByRole('button', { name: /Continue with Soma/i })).toBeDefined();
        expect(screen.queryByRole('link', { name: /Open run receipt/i })).toBeNull();
        expect(screen.queryByText('api.intent.confirm-action')).toBeNull();
    });

    it('does not duplicate plain system text when a structured thread event is present', () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Execution started - Soma accepted the approved work.',
                mode: 'execution_result',
                thread_event: {
                    kind: 'execution_started',
                    label: 'Execution started',
                    detail: 'Soma handed this to the work bus.',
                    tone: 'info',
                    status: 'running',
                    source_kind: 'web_api',
                    source_channel: 'api.intent.confirm-action',
                    payload_kind: 'soma_thread_event',
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getByTestId('soma-thread-state-card')).toBeDefined();
        expect(screen.getByText('Execution started')).toBeDefined();
        expect(screen.getByText('Soma handed this to the work bus.')).toBeDefined();
        expect(screen.queryByText('Execution started - Soma accepted the approved work.')).toBeNull();
    });

    it('returns completed team work as a concise directly openable result', () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Work complete - The team built and validated the app.',
                mode: 'execution_result',
                thread_event: {
                    kind: 'result_ready',
                    label: 'Work complete',
                    detail: 'The team built and validated the app. One deliverable is ready to open.',
                    tone: 'success',
                    status: 'completed',
                    href: '/api/v1/workspace/files/view?path=groups%2Fapp-team%2Fgenerated%2Fapp%2Findex.html',
                    href_label: 'Open app',
                    source_kind: 'system',
                    source_channel: 'team-work.result-projection',
                    payload_kind: 'thread_event',
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getByText('Work complete')).toBeDefined();
        expect(screen.getByText(/One deliverable is ready to open/i)).toBeDefined();
        expect(screen.getByRole('link', { name: /Open app/i }).getAttribute('href')).toContain('groups%2Fapp-team');
        expect(screen.queryByText('team-work.result-projection')).toBeNull();
    });

    it('renders adaptive answer depth without turning answers into proposal controls', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS, councilTarget: 'admin' });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    payload: {
                        text: '| Source | Use |\n| --- | --- |\n| Docs | Explain setup |',
                        response_depth: 'quick_box',
                    },
                },
            });
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Tell Soma/i);
        fireEvent.change(input, { target: { value: 'Give me a table of setup sources' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Quick answer')).toBeDefined();
            expect(screen.getByRole('table')).toBeDefined();
            expect(screen.queryByRole('button', { name: /Start this/i })).toBeNull();
            expect(screen.queryByRole('button', { name: /Approve/i })).toBeNull();
        });
    });

    it('renders decision brief depth as guidance without approval state', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS, councilTarget: 'admin' });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    payload: {
                        text: '**Recommendation:** keep Soma conversational and expand only on request.',
                        response_depth: 'decision_brief',
                    },
                },
            });
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Tell Soma/i);
        fireEvent.change(input, { target: { value: 'What should we do about output depth?' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Decision brief')).toBeDefined();
            expect(screen.getByText('Recommendation:')).toBeDefined();
            expect(screen.queryByText(/Waiting for executive review/i)).toBeNull();
            expect(screen.queryByRole('button', { name: /Start this/i })).toBeNull();
        });
    });
});
