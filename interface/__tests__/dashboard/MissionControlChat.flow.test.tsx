import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { requestSomaOutputContinuation } from '@/components/soma/outputContinuation';
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
            const chatCall = calls.find((call) => requestUrl(call[0]).includes('/api/v1/chat'));
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

    it('confirms pending work when the operator replies naturally', async () => {
        const proposal = {
            intent: 'create-output',
            teams: 1,
            agents: 1,
            tools: ['delegate'],
            risk_level: 'medium',
            confirm_token: 'confirm-123',
            intent_proof_id: 'proof-123',
            approval_required: true,
        };
        const confirmProposal = vi.fn().mockResolvedValue({ ok: true, runId: 'run-123' });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();
        act(() => {
            useCortexStore.setState({
                missionChat: [{ role: 'council', content: 'I can start that.', proposal, proposal_status: 'active' }],
                pendingProposal: proposal,
                activeConfirmToken: proposal.confirm_token,
                confirmProposal,
            });
        });

        const input = screen.getByRole('textbox');
        fireEvent.change(input, { target: { value: 'approve' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => expect(confirmProposal).toHaveBeenCalledWith(proposal, 'approve'));
        expect(mockFetch.mock.calls.some((call) => requestUrl(call[0]).includes('/api/v1/chat'))).toBe(false);
    });

    it('keeps requested proposal changes in the Soma conversation', async () => {
        const proposal = {
            intent: 'create-output',
            teams: 1,
            agents: 1,
            tools: ['delegate'],
            risk_level: 'medium',
            confirm_token: 'confirm-123',
            intent_proof_id: 'proof-123',
            approval_required: true,
        };
        const confirmProposal = vi.fn().mockResolvedValue({ ok: true, runId: 'run-123' });
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            if (requestUrl(input).includes('/api/v1/chat')) return okJson(CTS_CHAT_RESPONSE);
            return okJson({ ok: true, data: COUNCIL_MEMBERS });
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();
        act(() => {
            useCortexStore.setState({
                missionChat: [{ role: 'council', content: 'I can start that.', proposal, proposal_status: 'active' }],
                pendingProposal: proposal,
                activeConfirmToken: proposal.confirm_token,
                confirmProposal,
            });
        });

        const input = screen.getByRole('textbox');
        fireEvent.change(input, { target: { value: 'Use a smaller team and keep the same output.' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(mockFetch.mock.calls.some((call) => requestUrl(call[0]).includes('/api/v1/chat'))).toBe(true);
        });
        expect(confirmProposal).not.toHaveBeenCalled();
    });

    it('submits typed continuation context when replying to delivered output', async () => {
        mockFetch.mockImplementation(async () => okJson({
            ok: true,
            data: {
                ...CTS_CHAT_RESPONSE.data,
                meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'admin' },
            },
        }));

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        const detail = {
            title: 'Trusted Outcome Kit',
            reference: 'workspace/generated/trusted-outcome-kit',
            proof: 'proof-artifact-trusted-outcome',
        };
        act(() => requestSomaOutputContinuation(detail));

        expect(screen.getByText(/Continuing from/i)).toBeDefined();
        expect(screen.getByText('Trusted Outcome Kit')).toBeDefined();
        const input = screen.getByRole('textbox') as HTMLTextAreaElement;
        expect(input.value).toBe('');
        fireEvent.change(input, { target: { value: 'Make the launch page clearer.' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            const chatCall = mockFetch.mock.calls.find((call) => requestUrl(call[0]).includes('/api/v1/chat'));
            expect(chatCall).toBeDefined();
            const body = JSON.parse(String(chatCall?.[1]?.body ?? '{}'));
            expect(body.continuation_context).toEqual({
                kind: 'output',
                title: 'Trusted Outcome Kit',
                reference: 'workspace/generated/trusted-outcome-kit',
                proof: 'proof-artifact-trusted-outcome',
            });
        });
    });

    it('lets the operator clear delivered-output continuation context', async () => {
        mockFetch.mockResolvedValue(okJson({
            ok: true,
            data: {
                ...CTS_CHAT_RESPONSE.data,
                meta: { ...CTS_CHAT_RESPONSE.data.meta, source_node: 'admin' },
            },
        }));
        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        act(() => requestSomaOutputContinuation({ title: 'Trusted Outcome Kit', reference: 'workspace/generated/trusted-outcome-kit' }));
        fireEvent.click(screen.getByRole('button', { name: /Clear continuation from Trusted Outcome Kit/i }));
        fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Ordinary follow-up.' } });
        fireEvent.keyDown(screen.getByRole('textbox'), { key: 'Enter' });

        await waitFor(() => {
            const chatCall = mockFetch.mock.calls.find((call) => requestUrl(call[0]).includes('/api/v1/chat'));
            const body = JSON.parse(String(chatCall?.[1]?.body ?? '{}'));
            expect(body).not.toHaveProperty('continuation_context');
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
            const chatCall = calls.find((call) => requestUrl(call[0]).includes('/api/v1/council/council-architect/chat'));
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
