import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
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

    it('renders a simple run confirmation by default and hides low-level mechanics', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/run confirmation/i)).toBeDefined();
        expect(screen.getByText(/run this now/i)).toBeDefined();
        expect(screen.getByText(/soma will start only after you confirm/i)).toBeDefined();
        expect(screen.getByText(/what soma will do/i)).toBeDefined();
        expect(screen.getByText(/create a hello_world\.py file in your workspace\./i)).toBeDefined();
        expect(screen.getByText(/a new python file will be saved to workspace\/logs\/hello_world\.py after approval\./i)).toBeDefined();
        expect(screen.getAllByText(/workspace\/logs\/hello_world\.py/i).length).toBe(1);
        expect(screen.queryByText(/this action will change your workspace, so soma needs your approval before running it\./i)).toBeNull();
        expect(screen.getByText(/confirmation needed/i)).toBeDefined();
        expect(screen.queryByText(/risk medium/i)).toBeNull();
        expect(screen.queryByText(/current team bus/i)).toBeNull();
        expect(screen.queryByText(/no bus connection/i)).toBeNull();
        expect(screen.queryByText(/unless you approve bus wiring/i)).toBeNull();
        expect(screen.getByRole('button', { name: /review run details/i })).toBeDefined();
        expect(screen.queryByText(/execute delegate through governed module binding/i)).toBeNull();
        expect(screen.queryByText(/capability_risk/i)).toBeNull();
        expect(screen.queryByText(/delegate \(internal\)/i)).toBeNull();
    });

    it('reveals advanced execution details only after inspection', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /review run details/i }));

        expect(screen.getByText(/this action will change your workspace/i)).toBeDefined();
        expect(screen.getByText(/risk: medium/i)).toBeDefined();
        expect(screen.getAllByText(/workspace\/logs\/hello_world\.py/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/execute delegate through tool step/i)).toBeDefined();
        expect(screen.getByText(/1 team plan step/i)).toBeDefined();
        expect(screen.getAllByText(/^delegate$/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/needs approval/i)).toBeDefined();
        expect(screen.getByText(/file changes/i)).toBeDefined();
    });

    it('dispatches confirm and cancel actions', async () => {
        const confirmProposal = vi.fn().mockResolvedValue({ ok: true, runId: null });
        const cancelProposal = vi.fn();
        useCortexStore.setState({ confirmProposal, cancelProposal });

        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /run now/i }));

        await waitFor(() => expect(confirmProposal).toHaveBeenCalledTimes(1));
        expect(confirmProposal).toHaveBeenCalledWith(expect.objectContaining({
            confirm_token: 'ct-123',
            intent_proof_id: 'ip-123',
        }));

        fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
        expect(cancelProposal).toHaveBeenCalledTimes(1);
    });

    it('shows immediate execution feedback after approval click', async () => {
        let resolveConfirm!: (value: { ok: boolean; runId: string | null }) => void;
        const confirmProposal = vi.fn(() => new Promise<{ ok: boolean; runId: string | null }>((resolve) => {
            resolveConfirm = resolve;
        }));
        useCortexStore.setState({ confirmProposal, cancelProposal: vi.fn() });

        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /run now/i }));

        expect(screen.getByRole('button', { name: /running/i })).toBeDefined();
        expect(screen.getByText(/starting now/i)).toBeDefined();
        resolveConfirm({ ok: true, runId: 'run-1' });

        await waitFor(() => expect(screen.queryByRole('button', { name: /running/i })).toBeNull());
    });

    it('renders terminal lifecycle messaging and hides actions for cancelled proposals', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'cancelled' })} />);

        expect(screen.getByText(/cancelled/i)).toBeDefined();
        expect(screen.getByText(/no action executed/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /run now/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /cancel/i })).toBeNull();
    });

    it('shows pending-proof messaging after confirmation without execution proof', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'confirmed_pending_execution' })} />);

        expect(screen.getByText(/waiting for result/i)).toBeDefined();
        expect(screen.getByText(/approved, still running/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /run now/i })).toBeNull();
    });

    it('does not claim verification for an executed proposal until run proof exists', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed' })} />);

        expect(screen.getByText(/waiting for result/i)).toBeDefined();
        expect(screen.getByText(/approved, still running/i)).toBeDefined();
        expect(screen.queryByText(/action completed/i)).toBeNull();
    });

    it('renders a verified execution label when run proof exists', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed', run_id: 'run-123' })} />);

        expect(screen.getByText(/action completed/i)).toBeDefined();
        expect(screen.getByText(/result saved/i)).toBeDefined();
        expect(screen.getByRole('link', { name: /open run details/i }).getAttribute('href')).toBe('/runs/run-123');
        expect(screen.queryByRole('button', { name: /run now/i })).toBeNull();
    });

    it('renders a failed lifecycle without offering approval actions', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'failed' })} />);

        expect(screen.getAllByText(/could not run/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/nothing changed/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /run now/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /cancel/i })).toBeNull();
    });

    it('renders approval-required governance summary by default', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/run this now/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /run now/i })).toBeDefined();
    });

    it('shows scheduled or long-running task posture and bus scope after inspection', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                task_cadence: 'continuous',
                schedule_summary: 'Watch the incident channel every 5 minutes.',
                bus_scope: 'current_team',
                nats_subjects: ['swarm.team.ops.signal.status'],
            },
        })} />);

        expect(screen.queryByText(/when it runs/i)).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: /review run details/i }));

        expect(screen.getByText(/when it runs/i)).toBeDefined();
        expect(screen.getByText(/keep running/i)).toBeDefined();
        expect(screen.getByText(/watch the incident channel every 5 minutes/i)).toBeDefined();
        expect(screen.getByText(/team connection/i)).toBeDefined();
        expect(screen.getByText(/current team/i)).toBeDefined();
        expect(screen.getByText('swarm.team.ops.signal.status')).toBeDefined();
    });

    it('shows no-approval-needed posture for low-risk actions with plain detail labels', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                tools: ['generate_blueprint'],
                operator_summary: 'prepare a reusable implementation blueprint.',
                expected_result: 'A saved blueprint artifact will be returned in this conversation.',
                affected_resources: ['Blueprint artifact'],
                risk_level: 'low',
                approval_required: false,
                approval_mode: 'auto_allowed',
                approval_reason: 'auto_approve',
                capability_risk: 'low',
                capability_ids: ['planning'],
                estimated_cost: 0.2,
            },
        })} />);

        expect(screen.getByText(/let soma run this now/i)).toBeDefined();
        expect(screen.getAllByText(/ready/i).length).toBeGreaterThan(0);
        expect(screen.queryByText(/risk low/i)).toBeNull();
        expect(screen.queryByText(/auto approve/i)).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: /review run details/i }));

        expect(screen.getByText(/within current policy thresholds and can run without a mandatory approval/i)).toBeDefined();
        expect(screen.getByText(/risk: low, estimated cost 0\.20/i)).toBeDefined();
        expect(screen.getByText(/low-risk action/i)).toBeDefined();
        expect(screen.getByText(/planning/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /run now/i })).toBeDefined();
    });

    it('keeps a proposal visible but blocks execution when executable proof is missing', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                confirm_token: '',
                intent_proof_id: '',
            },
        })} />);

        expect(screen.getByRole('button', { name: /cannot run yet/i }).hasAttribute('disabled')).toBe(true);
        expect(screen.getByText(/missing the information soma needs to run it/i)).toBeDefined();
    });

    it('blocks execution when a token exists but proof linkage is missing', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                confirm_token: 'ct-present',
                intent_proof_id: '',
            },
        })} />);

        expect(screen.getByRole('button', { name: /cannot run yet/i }).hasAttribute('disabled')).toBe(true);
        expect(screen.getByText(/missing the information soma needs to run it/i)).toBeDefined();
    });
});
