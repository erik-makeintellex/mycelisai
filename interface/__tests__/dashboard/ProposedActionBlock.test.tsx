import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
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

    it('renders the user-facing approval summary by default and hides low-level mechanics', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/proposed action/i)).toBeDefined();
        expect(screen.getByText(/soma wants to/i)).toBeDefined();
        expect(screen.getByText(/create a hello_world\.py file in your workspace\./i)).toBeDefined();
        expect(screen.getByText(/a new python file will be saved to workspace\/logs\/hello_world\.py after approval\./i)).toBeDefined();
        expect(screen.getAllByText(/workspace\/logs\/hello_world\.py/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/this action will change your workspace, so soma needs your approval before running it\./i)).toBeDefined();
        expect(screen.getByText(/approval required/i)).toBeDefined();
        expect(screen.getByText(/risk medium/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /show details/i })).toBeDefined();
        expect(screen.queryByText(/execute delegate through governed module binding/i)).toBeNull();
        expect(screen.queryByText(/delegate \(internal\)/i)).toBeNull();
    });

    it('reveals advanced execution details only after inspection', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /show details/i }));

        expect(screen.getByText(/execute delegate through governed module binding/i)).toBeDefined();
        expect(screen.getByText(/1 expression/i)).toBeDefined();
        expect(screen.getByText(/delegate \(internal\)/i)).toBeDefined();
        expect(screen.getByText(/capability risk/i)).toBeDefined();
        expect(screen.getByText(/write file/i)).toBeDefined();
    });

    it('dispatches confirm and cancel actions', () => {
        const confirmProposal = vi.fn().mockResolvedValue({ ok: true, runId: null });
        const cancelProposal = vi.fn();
        useCortexStore.setState({ confirmProposal, cancelProposal });

        render(<ProposedActionBlock message={buildMessage()} />);

        fireEvent.click(screen.getByRole('button', { name: /approve & execute/i }));
        fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

        expect(confirmProposal).toHaveBeenCalledTimes(1);
        expect(cancelProposal).toHaveBeenCalledTimes(1);
    });

    it('renders terminal lifecycle messaging and hides actions for cancelled proposals', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'cancelled' })} />);

        expect(screen.getByText(/cancelled/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /approve & execute|execute/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /cancel/i })).toBeNull();
    });

    it('shows pending-proof messaging after confirmation without execution proof', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'confirmed_pending_execution' })} />);

        expect(screen.getByText(/awaiting execution proof/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /approve & execute|execute/i })).toBeNull();
    });

    it('does not claim verification for an executed proposal until run proof exists', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed' })} />);

        expect(screen.getByText(/awaiting execution proof/i)).toBeDefined();
        expect(screen.queryByText(/execution verified/i)).toBeNull();
    });

    it('renders a verified execution label when run proof exists', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'executed', run_id: 'run-123' })} />);

        expect(screen.getByText(/execution verified/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /approve & execute|execute/i })).toBeNull();
    });

    it('renders a failed lifecycle without offering approval actions', () => {
        render(<ProposedActionBlock message={buildMessage({ proposal_status: 'failed' })} />);

        expect(screen.getByText(/confirmation failed/i)).toBeDefined();
        expect(screen.queryByRole('button', { name: /approve & execute|execute/i })).toBeNull();
        expect(screen.queryByRole('button', { name: /cancel/i })).toBeNull();
    });

    it('renders approval-required governance details by default', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/why approval is needed/i)).toBeDefined();
        expect(screen.getByText(/this action will change your workspace/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /approve & execute/i })).toBeDefined();
    });

    it('shows scheduled or long-running task posture and bus scope before approval', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                task_cadence: 'continuous',
                schedule_summary: 'Watch the incident channel every 5 minutes.',
                bus_scope: 'current_team',
                nats_subjects: ['swarm.team.ops.signal.status'],
            },
        })} />);

        expect(screen.getByText(/task lifecycle/i)).toBeDefined();
        expect(screen.getByText(/keep running/i)).toBeDefined();
        expect(screen.getByText(/watch the incident channel every 5 minutes/i)).toBeDefined();
        expect(screen.getByText(/team \/ nats connection/i)).toBeDefined();
        expect(screen.getByText(/current team bus/i)).toBeDefined();
        expect(screen.getByText('swarm.team.ops.signal.status')).toBeDefined();
    });

    it('shows auto-approved execution posture for low-risk actions and keeps raw reasons in details', () => {
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

        expect(screen.getByText(/execution posture/i)).toBeDefined();
        expect(screen.getByText(/within current policy thresholds and can run without a mandatory approval/i)).toBeDefined();
        expect(screen.getByText(/auto-approved/i)).toBeDefined();
        expect(screen.getByText(/risk low/i)).toBeDefined();
        expect(screen.queryByText(/auto approve/i)).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: /show details/i }));

        expect(screen.getByText(/auto approve/i)).toBeDefined();
        expect(screen.getByText(/planning/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^execute$/i })).toBeDefined();
    });
});
