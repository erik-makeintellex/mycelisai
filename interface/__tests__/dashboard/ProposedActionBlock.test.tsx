import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';
import ProposedActionBlock from '@/components/dashboard/ProposedActionBlock';
import { useCortexStore, type ChatMessage } from '@/store/useCortexStore';

describe('ProposedActionBlock', () => {
    beforeEach(() => {
        useCortexStore.setState({
            assistantName: 'Soma',
            confirmProposal: vi.fn().mockResolvedValue({ ok: true, runId: 'run-1' }),
            cancelProposal: vi.fn(),
        });
    });

    function buildMessage(overrides: Partial<ChatMessage> = {}): ChatMessage {
        return {
            role: 'council',
            content: 'Proposed execution path',
            mode: 'proposal',
            source_node: 'admin',
            proposal: {
                intent: 'chat-action',
                operator_summary: 'create a hello_world.py file in your workspace.',
                expected_result: 'A new Python file will be saved to workspace/logs/hello_world.py after approval.',
                affected_resources: ['workspace/logs/hello_world.py'],
                teams: 1,
                agents: 1,
                tools: ['delegate'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
                approval_required: true,
                approval_reason: 'capability_risk',
                approval_mode: 'required',
                capability_risk: 'medium',
                capability_ids: ['write_file'],
                team_expressions: [
                    {
                        expression_id: 'expr-1',
                        team_id: 'admin-core',
                        objective: 'Execute delegate through governed module binding',
                        role_plan: ['admin'],
                        module_bindings: [
                            {
                                binding_id: 'binding-1-delegate',
                                module_id: 'delegate',
                                adapter_kind: 'internal',
                                operation: 'delegate',
                            },
                        ],
                    },
                ],
            },
            proposal_status: 'active',
            ...overrides,
        };
    }

    it('renders a compact natural approval pause by default and hides low-level mechanics', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/i can start that/i)).toBeDefined();
        expect(screen.getByText(/once you approve, i’ll start the work and stay here for questions or changes/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^approve$/i })).toBeDefined();
        expect(screen.getByText(/or reply.*approve.*ask a question or request a change/i)).toBeDefined();
        expect(screen.queryByText(/what i will do/i)).toBeNull();
        expect(screen.getByText(/bring in the right team and keep their work connected to this conversation/i)).toBeDefined();
        expect(screen.getByText(/create a hello_world\.py file in your workspace\./i)).toBeDefined();
        expect(screen.getByText(/a new python file will be saved to workspace\/logs\/hello_world\.py after approval\./i)).toBeDefined();
        expect(screen.getAllByText(/workspace\/logs\/hello_world\.py/i).length).toBeGreaterThan(0);
        expect(screen.queryByText(/this action will change your workspace, so soma needs your approval before running it\./i)).toBeNull();
        expect(
            screen.getByText(/or reply.*approve.*ask a question or request a change/i),
        ).toBeTruthy();
        expect(screen.queryByText(/risk medium/i)).toBeNull();
        expect(screen.queryByText(/\bbus\b/i)).toBeNull();
        expect(screen.queryByText(/current team bus/i)).toBeNull();
        expect(screen.queryByText(/no bus connection/i)).toBeNull();
        expect(screen.queryByText(/unless you approve bus wiring/i)).toBeNull();
        expect(screen.getByRole('button', { name: /^details$/i })).toBeDefined();
        expect(screen.queryByText(/execute delegate through governed module binding/i)).toBeNull();
        expect(screen.queryByText(/capability_risk/i)).toBeNull();
        expect(screen.queryByText(/delegate \(internal\)/i)).toBeNull();
    });

    it('keeps a new team identifier behind Details when the named result is already clear', () => {
        const message = buildMessage();
        message.proposal = {
            ...message.proposal!,
            operator_summary: 'Create Moonlit Keep Game Team and start its first playable.',
            expected_result: 'Moonlit Keep First Playable will be ready to open after validation.',
            affected_resources: ['team:moonlit-keep-game-123'],
            tools: ['create_team', 'write_file', 'delegate_task'],
        };
        render(<ProposedActionBlock message={message} />);

        expect(screen.getByText(/moonlit keep first playable will be ready/i)).toBeDefined();
        expect(screen.queryByText(/moonlit-keep-game-123/i)).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: /^details$/i }));
        expect(screen.getByText(/moonlit-keep-game-123/i)).toBeDefined();
    });

    it('reveals advanced execution details only after inspection', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /^details$/i }));

        expect(screen.getByText(/this action will change your workspace/i)).toBeDefined();
        expect(screen.getByText(/risk: medium/i)).toBeDefined();
        expect(screen.getAllByText(/workspace\/logs\/hello_world\.py/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/execute delegate through tool step/i)).toBeDefined();
        expect(screen.getByText(/1 team plan step/i)).toBeDefined();
        expect(screen.getAllByText(/^delegate$/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/needs approval/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/file changes/i)).toBeDefined();
    });

    it('offers one primary approval action inside the conversation', async () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /^approve$/i }));
        // Live test L1: the button click must not echo "approve"/"start" as
        // a synthetic user chat message, so no second argument is passed.
        await waitFor(() => expect(useCortexStore.getState().confirmProposal).toHaveBeenCalledWith(buildMessage().proposal));
        expect(screen.queryByRole('button', { name: /adjust/i })).toBeNull();
        expect(screen.getByRole('button', { name: /^details$/i })).toBeDefined();
    });

    it('renders terminal lifecycle messaging and hides actions for cancelled proposals', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'cancelled' })} />);

        expect(screen.getByText(/cancelled/i)).toBeDefined();
        expect(screen.getByText(/no action executed/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /^approve$/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /adjust/i })).toBeNull();
    });

    it('shows pending-proof messaging after confirmation without execution proof', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'confirmed_pending_execution' })} />);

        expect(screen.getByText(/waiting for result/i)).toBeDefined();
        expect(screen.getByText(/approved, awaiting result/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /^approve$/i })).toBeNull();
    });

    it('does not claim approval succeeded while confirmation is still in flight', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal_status: 'confirmed_pending_execution',
            ui_response_state: { kind: 'proposal', label: 'Approval sent', tone: 'info' },
        })} />);

        expect(screen.getByText(/checking approval/i)).toBeDefined();
        expect(screen.getByText(/no change is confirmed yet/i)).toBeDefined();
        expect(screen.queryByText(/approved, awaiting result|approved, still running|action completed/i)).toBeNull();
    });

    it('does not claim verification for an executed proposal until run proof exists', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed' })} />);

        expect(screen.getByText(/waiting for result/i)).toBeDefined();
        expect(screen.getByText(/approved, awaiting result/i)).toBeDefined();
        expect(screen.queryByText(/action completed/i)).toBeNull();
    });

    it('keeps delegated work pending when approval only produced a run id', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed', run_id: 'run-123' })} />);

        expect(screen.getByText(/waiting for result/i)).toBeDefined();
        expect(screen.getByText(/approved, awaiting result/i)).toBeDefined();
        expect(screen.queryByText(/action completed/i)).toBeNull();
        expect(screen.queryByText(/result verified/i)).toBeNull();
        expect(screen.queryByText(/result saved/i)).toBeNull();
        expect(screen.queryByText(/proof is available in trust/i)).toBeNull();
        expect(screen.queryByRole('button', { name: /^approve$/i })).toBeNull();
    });

    it('keeps a nondelegated run pending without explicit verified terminal proof', () => {
        const message = buildMessage({ proposal_status: 'executed', run_id: 'run-123' });
        message.proposal = { ...message.proposal!, tools: ['write_file'], team_expressions: [] };
        render(<ProposedActionBlock message={message} />);

        expect(screen.getByText(/approved, awaiting result/i)).toBeDefined();
        expect(screen.queryByText(/action completed|result saved|proof is available in trust/i)).toBeNull();
    });

    it('shows a nondelegated result only with completed status and verified proof', () => {
        const message = buildMessage({
            proposal_status: 'executed', run_id: 'run-123',
            execution_summary: {
                execution: { shape: 'guided_proposal', status: 'completed' },
                proof: { run_id: 'run-123', verified: true },
            },
        });
        message.proposal = { ...message.proposal!, tools: ['write_file'], team_expressions: [] };
        render(<ProposedActionBlock message={message} />);

        expect(screen.getByText(/action completed/i)).toBeDefined();
        expect(screen.getByText(/result saved/i)).toBeDefined();
        expect(screen.getByText(/proof is available in trust/i)).toBeDefined();
    });

    it('shows a running state only when the execution summary says running', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal_status: 'confirmed_pending_execution', run_id: 'run-123',
            execution_summary: { execution: { status: 'running' }, proof: { verified: false } },
        })} />);

        expect(screen.getByText(/approved, still running/i)).toBeDefined();
        expect(screen.queryByText(/result saved|proof is available in trust/i)).toBeNull();
    });

    it('labels delegated work verified only with an explicit verified terminal result', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal_status: 'executed',
            run_id: 'run-123',
            execution_summary: {
                execution: { shape: 'team_execution', status: 'verified' },
                proof: [{ run_id: 'run-123', verified: true }],
            },
        })} />);

        expect(screen.getByText(/result verified/i)).toBeDefined();
        expect(screen.getByText(/result saved/i)).toBeDefined();
        expect(screen.getByText(/proof is available in trust/i)).toBeDefined();
        expect(screen.queryByRole('link', { name: /open run details/i })).toBeNull();
        expect(screen.queryByText(/action completed/i)).toBeNull();
    });

    it('renders a failed lifecycle without offering approval actions', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'failed' })} />);

        expect(screen.getAllByText(/could not run/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/nothing changed/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /^approve$/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /adjust/i })).toBeNull();
    });
});
