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

    it('renders structured thread state without exposing raw routing subjects', async () => {
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
            expect(screen.getByTestId('soma-thread-state-card')).toBeDefined();
            expect(screen.getByText('Waiting for executive review')).toBeDefined();
            expect(screen.getByText('Current team route')).toBeDefined();
            expect(screen.getByText('Review the generated action card before Soma starts the background work.')).toBeDefined();
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
                    detail: 'Soma accepted the approved work and saved the run receipt for proof.',
                    tone: 'info',
                    status: 'running',
                    href: '/runs/run-thread-123',
                    href_label: 'Open run receipt',
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
        expect(screen.getByText('Soma accepted the approved work and saved the run receipt for proof.')).toBeDefined();
        expect(screen.getByRole('link', { name: /Open run receipt/i }).getAttribute('href')).toBe('/runs/run-thread-123');
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
                    detail: 'Soma accepted the approved work.',
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
        expect(screen.getByText('Soma accepted the approved work.')).toBeDefined();
        expect(screen.queryByText('Execution started - Soma accepted the approved work.')).toBeNull();
    });
});
