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

const realSendMissionChat = useCortexStore.getState().sendMissionChat;

describe('MissionControlChat thread state cards', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        useCortexStore.setState({ sendMissionChat: realSendMissionChat });
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

    it('renders accepted work as running rather than completed proof and keeps Soma available', () => {
        const sendMissionChat = vi.fn();
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Work started - Soma accepted the approved work.',
                mode: 'execution_result',
                run_id: 'run-thread-123',
                thread_events: [{
                    kind: 'execution_started',
                    label: 'Work started',
                    detail: 'Soma handed this to the work bus. It is running, not complete, and you can keep talking here.',
                    tone: 'info',
                    status: 'running',
                    run_id: 'run-thread-123',
                    team_id: 'launch-team',
                    work_item_id: 'work-thread-123',
                    target_reference: 'run:run-thread-123',
                    source_kind: 'web_api',
                    source_channel: 'api.intent.confirm-action',
                    payload_kind: 'soma_thread_event',
                }],
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
            isMissionChatting: false,
            isBroadcasting: false,
            sendMissionChat,
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getByTestId('soma-thread-state-card')).toBeDefined();
        expect(screen.getByText('Work started')).toBeDefined();
        expect(screen.queryByText('running')).toBeNull();
        expect(screen.getByText('Soma handed this to the work bus. It is running, not complete, and you can keep talking here.')).toBeDefined();
        expect(screen.queryByRole('button', { name: /Continue with Soma/i })).toBeNull();
        expect(screen.queryByRole('link', { name: /Open run receipt/i })).toBeNull();
        expect(screen.queryByText(/Work complete/i)).toBeNull();
        expect(screen.queryByText(/Result verified/i)).toBeNull();
        expect(screen.queryByText('api.intent.confirm-action')).toBeNull();

        const input = screen.getByPlaceholderText(/Tell Soma/i);
        expect((input as HTMLTextAreaElement).disabled).toBe(false);
        fireEvent.change(input, { target: { value: 'Also prepare a concise launch note.' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        expect(sendMissionChat).toHaveBeenCalledWith('Also prepare a concise launch note.', {
            active_work_context: {
                type: 'team_work',
                id: 'work-thread-123',
                run_id: 'run-thread-123',
                team_id: 'launch-team',
                work_item_id: 'work-thread-123',
            },
            continuation_context: undefined,
        });
        expect((input as HTMLTextAreaElement).value).toBe('');
    });

    it('shows one clear latest state when the same work emits repeated machine updates', () => {
        const threadEvent = (kind: 'execution_started' | 'attention_required', label: string, detail: string) => ({
            kind,
            label,
            detail,
            tone: kind === 'attention_required' ? 'warning' as const : 'info' as const,
            status: kind === 'attention_required' ? 'degraded' : 'running',
            run_id: 'run-repeated-1',
            source_kind: 'system',
            source_channel: 'team-work.result-projection',
            payload_kind: 'thread_event',
        });
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'system',
                    content: 'Work started',
                    mode: 'execution_result',
                    run_id: 'run-repeated-1',
                    thread_event: threadEvent('execution_started', 'Work started', 'Work is underway.'),
                },
                {
                    role: 'system',
                    content: 'Work needs attention',
                    mode: 'blocker',
                    run_id: 'run-repeated-1',
                    thread_event: threadEvent('attention_required', 'Work needs attention', 'The configured provider did not return a readable reply.'),
                },
                {
                    role: 'system',
                    content: 'Work needs attention again',
                    mode: 'blocker',
                    run_id: 'run-repeated-1',
                    thread_event: threadEvent('attention_required', 'Work needs attention', 'The configured provider timed out.'),
                },
            ],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getAllByTestId('soma-thread-state-card')).toHaveLength(1);
        expect(screen.getByText('Soma needs your direction')).toBeDefined();
        expect(screen.getByText(/Tell Soma to try again, use another available service, or change the request/i)).toBeDefined();
        expect(screen.queryByText('degraded')).toBeNull();
        expect(screen.queryByText('Work started')).toBeNull();
        expect(screen.queryByRole('button', { name: /Continue with Soma/i })).toBeNull();
        expect(screen.getByPlaceholderText(/Tell Soma/i)).toBeDefined();
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

    it('states when a contract-unsatisfied game output is not playable', () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Output is not playable yet',
                mode: 'blocker',
                run_id: 'run-game-failed',
                thread_event: {
                    kind: 'attention_required',
                    label: 'Output is not playable yet',
                    detail: 'The team stopped because the required runnable output was not validated.',
                    tone: 'warning',
                    status: 'result_contract_unsatisfied',
                    run_id: 'run-game-failed',
                    source_kind: 'web_api',
                    source_channel: 'api.intent.confirm-action',
                    payload_kind: 'soma_thread_event',
                    target_reference: 'result_contract_unsatisfied',
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(screen.getByText('Output is not playable yet')).toBeDefined();
        expect(screen.getByText('The team did not produce a validated runnable output. Nothing new should be trusted yet.')).toBeDefined();
        expect(screen.getByText(/Tell Soma to try again, use another available service, or change the request/i)).toBeDefined();
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
        const outputLink = screen.getByRole('link', { name: /Open app/i });
        expect(outputLink.getAttribute('href')).toContain('/outputs/view?');
        expect(outputLink.getAttribute('href')).toContain('groups%252Fapp-team');
        expect(outputLink.getAttribute('href')).toContain('return_to=%2Fdashboard');
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
