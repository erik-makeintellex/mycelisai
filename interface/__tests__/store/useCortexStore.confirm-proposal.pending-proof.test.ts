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
                kind: 'running',
                label: 'Started',
                detail: 'Soma started the work. You can keep talking here while updates arrive.',
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
            content: 'Proposal approved. Soma started the work. You can keep talking here while updates arrive.',
            ui_response_state: {
                kind: 'running',
                label: 'Started',
                detail: 'Soma started the work. You can keep talking here while updates arrive.',
                tone: 'info',
            },
            thread_events: [{
                kind: 'execution_update',
                label: 'Work queued',
                detail: 'Soma handed off the approved work. It is queued until the team accepts it.',
                tone: 'info',
                status: 'queued',
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
                kind: 'running',
                label: 'Started',
                detail: 'Soma started the work. You can keep talking here while updates arrive.',
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
                execution_summary: { execution: {
                    status: 'completed',
                    summary: 'Outcome Template "Delivery Brief" v2 saved with digest sha256:abc.',
                } },
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
                confirmed: true, verified: false, run_id: 'run-config-active', run_status: 'running',
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
