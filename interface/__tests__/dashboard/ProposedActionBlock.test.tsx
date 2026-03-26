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
                teams: 1,
                agents: 1,
                tools: ['delegate'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
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

    it('renders team expression and module binding details', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/proposed action/i)).toBeDefined();
        expect(screen.getByText(/execute delegate through governed module binding/i)).toBeDefined();
        expect(screen.getByText(/1 expression/i)).toBeDefined();
        expect(screen.getByText(/delegate \(internal\)/i)).toBeDefined();
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

        expect(screen.getByText(/approval required before execution/i)).toBeDefined();
        expect(screen.getByText(/capability medium/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /approve & execute/i })).toBeDefined();
    });

    it('shows auto-approved execution details for low-risk actions', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                tools: ['generate_blueprint'],
                risk_level: 'low',
                approval_required: false,
                approval_mode: 'auto_allowed',
                approval_reason: 'auto_approve',
                capability_risk: 'low',
                capability_ids: ['planning'],
                estimated_cost: 0.2,
            },
        })} />);

        expect(screen.getByText(/auto-approved within governance thresholds/i)).toBeDefined();
        expect(screen.getByText(/reason: auto approve/i)).toBeDefined();
        expect(screen.getByText(/planning/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^execute$/i })).toBeDefined();
    });
});
