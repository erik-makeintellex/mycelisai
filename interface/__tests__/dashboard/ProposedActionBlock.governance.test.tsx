import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';
import ProposedActionBlock from '@/components/dashboard/ProposedActionBlock';
import { useCortexStore, type ChatMessage } from '@/store/useCortexStore';

describe('ProposedActionBlock governance and proof gates', () => {
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

    it('renders approval-required governance summary by default', () => {
        render(<ProposedActionBlock message={buildMessage()} />);

        expect(screen.getByText(/or reply.*approve.*ask a question/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^approve$/i })).toBeDefined();
    });

    it('shows scheduled or long-running task posture and bus scope after inspection', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                work_intent: {
					kind: 'service',
                    cadence: 'continuous',
                    schedule_summary: 'Watch the incident channel every 5 minutes.',
                    bus_scope: 'current_team',
                    nats_subjects: ['swarm.team.ops.signal.status'],
                    output_contract: {
                        shape: 'app_package',
                        primary_deliverable: 'Playable browser package with proof.',
                        launch_hint: 'Return an openable entrypoint and folder access.',
                    },
					lifecycle: {
						stop_action: 'stop_service',
						retry_action: 'restart_service',
						recovery_action: 'inspect_and_restart',
						control_summary: 'You can stop the service or inspect its last trusted state.',
					},
                },
            },
        })} />);

        expect(screen.queryByText(/when it runs/i)).toBeNull();
        expect(screen.queryByText(/expected output/i)).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: /^details$/i }));

        expect(screen.getByText(/when it runs/i)).toBeDefined();
        expect(screen.getByText(/keep running/i)).toBeDefined();
        expect(screen.getByText(/watch the incident channel every 5 minutes/i)).toBeDefined();
		expect(screen.getByText(/control:/i)).toBeDefined();
		expect(screen.getByText(/stop service/i)).toBeDefined();
		expect(screen.getByText(/inspect its last trusted state/i)).toBeDefined();
        expect(screen.getByText(/team connection/i)).toBeDefined();
        expect(screen.getByText(/current team/i)).toBeDefined();
        expect(screen.getByText('swarm.team.ops.signal.status')).toBeDefined();
        expect(screen.getByText(/expected output/i)).toBeDefined();
        expect(screen.getByText(/app or package/i)).toBeDefined();
        expect(screen.getByText(/playable browser package with proof/i)).toBeDefined();
        expect(screen.getByText(/openable entrypoint/i)).toBeDefined();
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

        expect(screen.getByText(/ready to start.*stay here for questions or changes/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^start$/i })).toBeDefined();
        expect(screen.getByText(/or reply.*start.*ask a question or request a change/i)).toBeDefined();
        expect(screen.queryByText(/risk low/i)).toBeNull();
        expect(screen.queryByText(/auto approve/i)).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: /^details$/i }));

        expect(screen.getByText(/within current policy thresholds and can run without a mandatory approval/i)).toBeDefined();
        expect(screen.getByText(/risk: low, estimated cost 0\.20/i)).toBeDefined();
        expect(screen.getByText(/low-risk action/i)).toBeDefined();
        expect(screen.getByText(/planning/i)).toBeDefined();
        expect(screen.getByRole('button', { name: /^start$/i })).toBeDefined();
    });

    it('keeps a proposal visible but blocks execution when executable proof is missing', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                confirm_token: '',
                intent_proof_id: '',
            },
        })} />);

        expect(screen.queryByRole('button', { name: /cannot run yet/i })).toBeNull();
        expect(screen.getByText(/cannot start this version yet/i)).toBeDefined();
    });

    it('blocks execution when a token exists but proof linkage is missing', () => {
        render(<ProposedActionBlock message={buildMessage({
            proposal: {
                ...buildMessage().proposal!,
                confirm_token: 'ct-present',
                intent_proof_id: '',
            },
        })} />);

        expect(screen.queryByRole('button', { name: /cannot run yet/i })).toBeNull();
        expect(screen.getByText(/cannot start this version yet/i)).toBeDefined();
    });

    describe('required_approver_role (deck top-10 #1)', () => {
        it('shows "Needs admin approval" and no Approve button for a standard user', async () => {
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: 'standard' } } }) });
            render(<ProposedActionBlock message={buildMessage({
                proposal: { ...buildMessage().proposal!, required_approver_role: 'admin' },
            })} />);

            await waitFor(() => expect(screen.getByText('Needs admin approval')).toBeDefined());
            expect(screen.queryByRole('button', { name: /approve/i })).toBeNull();
            expect(screen.getByText(/an admin can approve this/i)).toBeDefined();
        });

        it('lets an admin approve, labeled "Approve as admin", with a self-approval audit note', async () => {
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: 'admin' } } }) });
            render(<ProposedActionBlock message={buildMessage({
                proposal: { ...buildMessage().proposal!, required_approver_role: 'admin' },
            })} />);

            const button = await screen.findByRole('button', { name: 'Approve as admin' });
            expect(button).toBeDefined();
            expect(screen.getByText(/recorded in history/i)).toBeDefined();
        });

        it('shows the normal Approve button when no admin approval is required', async () => {
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: 'standard' } } }) });
            render(<ProposedActionBlock message={buildMessage()} />);

            expect(await screen.findByRole('button', { name: 'Approve' })).toBeDefined();
            expect(screen.queryByText('Needs admin approval')).toBeNull();
        });
    });

    describe('draft_previews (D2 contract)', () => {
        it('shows the drafted file preview before Approve, with a "first N lines" label when truncated', () => {
            render(<ProposedActionBlock message={buildMessage({
                proposal: {
                    ...buildMessage().proposal!,
                    tools: ['write_file'],
                    draft_previews: [
                        { path: 'workspace/logs/hello_world.py', preview: 'print("hello")\nprint("world")', full_draft: false, lines: 2, bytes: 30 },
                    ],
                },
            })} />);

            expect(screen.getByText('workspace/logs/hello_world.py')).toBeDefined();
            expect(screen.getByText('Preview — first 2 lines')).toBeDefined();
            expect(screen.getByText((_, el) => el?.tagName === 'PRE' && el.textContent === 'print("hello")\nprint("world")')).toBeDefined();
        });

        it('omits the "first N lines" label when the preview is the full draft', () => {
            render(<ProposedActionBlock message={buildMessage({
                proposal: {
                    ...buildMessage().proposal!,
                    tools: ['write_file'],
                    draft_previews: [
                        { path: 'workspace/logs/short.py', preview: 'pass', full_draft: true, lines: 1, bytes: 4 },
                    ],
                },
            })} />);

            expect(screen.getByText('workspace/logs/short.py')).toBeDefined();
            expect(screen.queryByText(/Preview — first/)).toBeNull();
        });

        it('renders as today when draft_previews is absent', () => {
            render(<ProposedActionBlock message={buildMessage()} />);
            expect(screen.queryByText(/Preview — first/)).toBeNull();
        });
    });

    describe('team-plan proposals (create_team)', () => {
        it('points to the Workspace canvas instead of a guaranteed-failing Approve click', () => {
            render(<ProposedActionBlock message={buildMessage({
                proposal: { ...buildMessage().proposal!, tools: ['create_team'] },
            })} />);

            expect(screen.queryByRole('button', { name: /approve/i })).toBeNull();
            expect(screen.getByText(/isn.t available yet/i)).toBeDefined();
            expect(screen.getByRole('link', { name: 'Open Workspace' }).getAttribute('href')).toBe('/automations');
        });
    });
});
