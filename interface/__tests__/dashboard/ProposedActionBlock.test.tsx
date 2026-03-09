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

    function buildMessage(): ChatMessage {
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

        fireEvent.click(screen.getByRole('button', { name: /confirm & execute/i }));
        fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

        expect(confirmProposal).toHaveBeenCalledTimes(1);
        expect(cancelProposal).toHaveBeenCalledTimes(1);
    });
});
