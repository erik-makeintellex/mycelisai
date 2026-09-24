import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore confirm proposal pending proof', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('keeps the proposal in a pending-proof state when confirmation succeeds without a run id', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            },
            activeConfirmToken: 'ct-123',
            missionChat: [{
                role: 'council',
                content: 'Proposed execution path',
                mode: 'proposal',
                proposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ data: { confirmed: true, run_id: null } }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({ ok: true, runId: null });
        expect(useCortexStore.getState().activeMode).toBe('proposal');
        expect(useCortexStore.getState().activeRunId).toBeNull();
        expect(useCortexStore.getState().pendingProposal).toBeNull();
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'confirmed_pending_execution',
            mode: 'proposal',
            ui_response_state: {
                kind: 'proposal',
                label: 'Approval recorded',
                detail: 'Completion has not been verified.',
                tone: 'info',
            },
            thread_events: [{
                kind: 'execution_update',
                label: 'Approval sent',
                detail: 'Soma is starting the handoff.',
                tone: 'info',
                status: 'confirming',
                source_kind: 'workspace_ui',
                source_channel: 'soma.proposal.confirm',
                payload_kind: 'soma_thread_event',
            }],
        });
        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            role: 'system',
            mode: 'proposal',
            content: 'Proposal approved. Approval was recorded. Completion has not been verified.',
            ui_response_state: {
                kind: 'proposal',
                label: 'Approval recorded',
                detail: 'Completion has not been verified.',
                tone: 'info',
            },
            thread_events: [{
                kind: 'execution_update',
                label: 'Approval recorded',
                detail: 'Soma received the approval. Completion has not been verified.',
                tone: 'info',
                status: 'confirmed',
                source_kind: 'web_api',
                source_channel: 'api.intent.confirm-action',
                payload_kind: 'soma_thread_event',
            }],
        });
    });

    it('exposes a conversational started state while confirmation is still in flight', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            },
            activeConfirmToken: 'ct-123',
            missionChat: [{
                role: 'council',
                content: 'Proposed execution path',
                mode: 'proposal',
                proposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                proposal_status: 'active',
            }],
            activeMode: 'proposal',
        });
        let resolveFetch!: (value: Response) => void;
        mockFetch.mockImplementation(() => new Promise<Response>((resolve) => {
            resolveFetch = resolve;
        }));

        const pending = useCortexStore.getState().confirmProposal();

        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'confirmed_pending_execution',
            mode: 'proposal',
            ui_response_state: {
                kind: 'proposal',
                label: 'Approval sent',
                detail: 'Soma is checking the approval.',
                tone: 'info',
            },
        });

        resolveFetch({
            ok: true,
            json: async () => ({ data: { confirmed: true, run_id: 'run-1' } }),
        } as Response);
        await pending;
    });

    it('presents synchronous template storage as a completed conversation instead of bus work', async () => {
        const proposal = {
            intent: 'Save the approved Outcome Template',
            teams: 0,
            agents: 0,
            tools: ['store_config_document'],
            risk_level: 'medium',
            confirm_token: 'ct-config',
            intent_proof_id: 'ip-config',
        };
        useCortexStore.setState({
            pendingProposal: proposal,
            activeConfirmToken: proposal.confirm_token,
            missionChat: [{ role: 'council', content: 'Save this template?', mode: 'proposal', proposal, proposal_status: 'active' }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: true,
            status: 200,
            json: async () => ({ data: {
                confirmed: true,
                run_id: 'run-config',
                verified: true,
                run_status: 'completed',
                execution_state: 'verified',
                execution_summary: { execution: {
                    status: 'completed',
                    summary: 'Outcome Template "Delivery Brief" v2 saved with digest sha256:abc.',
                }, proof: { run_id: 'run-config', verified: true, proof_id: 'proof-config' } },
            } }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({ ok: true, runId: 'run-config' });
        const state = useCortexStore.getState();
        expect(state.activeRunId).toBeNull();
        expect(state.activeMode).toBe('execution_result');
        expect(state.missionChat.at(-1)).toMatchObject({
            content: 'Outcome Template "Delivery Brief" v2 saved with digest sha256:abc.',
            mode: 'execution_result',
            ui_response_state: {
                kind: 'execution_result',
                label: 'Template saved',
                detail: 'It is saved but remains inactive until you activate it.',
                tone: 'success',
            },
            thread_events: [{ kind: 'result_ready', label: 'Template saved', status: 'completed' }],
        });
        expect(JSON.stringify(state.missionChat)).not.toMatch(/work bus|Work started|Run run-config started/i);
    });

    it.each([
        ['failed execution state', { execution_state: 'failed' }],
        ['cancelled run', { run_status: 'cancelled' }],
        ['unknown execution state', { execution_state: 'unknown' }],
        ['contradictory execution status', { execution_status: 'failed' }],
        ['missing proof', { execution_summary: { execution: { status: 'completed' }, proof: { verified: false } } }],
        ['unconfirmed response', { confirmed: false }],
    ])('does not claim a template was saved for %s', async (_caseName, override) => {
        const proposal = {
            intent: 'Save the template', teams: 0, agents: 0,
            tools: ['store_config_document'], risk_level: 'medium',
            confirm_token: 'ct-config-negative', intent_proof_id: 'ip-config-negative',
        };
        useCortexStore.setState({
            pendingProposal: proposal, activeConfirmToken: proposal.confirm_token,
            missionChat: [{ role: 'council', content: 'Save?', mode: 'proposal', proposal, proposal_status: 'active' }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValueOnce({
            ok: true, status: 200,
            json: async () => ({ data: {
                confirmed: true, verified: true, run_id: 'run-config-negative',
                run_status: 'completed', execution_state: 'verified',
                execution_summary: { execution: { status: 'completed' }, proof: { verified: true } },
                ...override,
            } }),
        }).mockResolvedValueOnce({ ok: true, json: async () => ([]) });

        await useCortexStore.getState().confirmProposal();

        const state = useCortexStore.getState();
        expect(state.missionChat[0].proposal_status).toBe('confirmed_pending_execution');
        expect(state.missionChat.at(-1)?.thread_events?.[0]).toMatchObject({ label: 'Approval recorded', status: 'confirmed' });
        expect(JSON.stringify(state.missionChat)).not.toMatch(/Template saved|Template active|Result verified|Result saved/);
    });

    it('keeps a config-only 202 response visible as active work', async () => {
        const proposal = {
            intent: 'Save the template', teams: 0, agents: 0,
            tools: ['store_config_document'], risk_level: 'medium',
            confirm_token: 'ct-running', intent_proof_id: 'ip-running',
        };
        useCortexStore.setState({
            pendingProposal: proposal, activeConfirmToken: proposal.confirm_token,
            missionChat: [{ role: 'council', content: 'Save?', mode: 'proposal', proposal, proposal_status: 'active' }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: true, status: 202,
            json: async () => ({ data: {
                confirmed: true, verified: false, run_id: 'run-config-active', run_status: 'running', execution_state: 'running',
                execution_summary: { execution: { status: 'running' } },
            } }),
        });

        await useCortexStore.getState().confirmProposal();

        const state = useCortexStore.getState();
        expect(state.activeRunId).toBe('run-config-active');
        expect(state.missionChat.at(-1)).toMatchObject({
            run_id: 'run-config-active',
            thread_events: [{ kind: 'execution_update', status: 'queued' }],
        });
        expect(JSON.stringify(state.missionChat)).not.toMatch(/Template saved|Template active/);
    });
});
