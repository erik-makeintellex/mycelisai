import { describe, expect, it } from 'vitest';
import { presentMissionChat } from '@/components/dashboard/missionControlChatPresentation';
import type { ChatMessage } from '@/store/useCortexStore';

function runningEvent(runId: string) {
    return {
        kind: 'execution_started' as const,
        label: 'Work started',
        detail: 'Soma handed this to the work bus. It is running, not complete, and you can keep talking here.',
        tone: 'info' as const,
        status: 'running',
        run_id: runId,
    };
}

function completedEvent(runId: string) {
    return {
        kind: 'result_ready' as const,
        label: 'Result verified',
        detail: 'Soma completed the approved action and saved its proof.',
        tone: 'success' as const,
        status: 'completed',
        run_id: runId,
    };
}

describe('presentMissionChat (live test L1: stale running card)', () => {
    it('collapses an earlier running card into a later completed card for the same run, even with a message in between', () => {
        const messages: ChatMessage[] = [
            {
                role: 'system',
                content: 'Work started - Soma accepted the approved work.',
                mode: 'execution_result',
                run_id: 'run-1',
                thread_events: [runningEvent('run-1')],
            },
            // A message in a different role sits between the two updates —
            // this used to reset the compaction block and strand the
            // running card on screen.
            { role: 'user', content: 'Also prepare a launch note.' },
            {
                role: 'system',
                content: 'Run run-1 completed. The verified result is available to review.',
                mode: 'execution_result',
                run_id: 'run-1',
                thread_events: [completedEvent('run-1')],
            },
        ];

        const presented = presentMissionChat(messages);

        expect(presented).toHaveLength(2);
        expect(presented[0].role).toBe('user');
        expect(presented[1].thread_events?.[0].status).toBe('completed');
        expect(presented.some((m) => m.thread_events?.[0]?.status === 'running')).toBe(false);
    });

    it('does not touch messages with no run_id or team_id (nothing to collapse)', () => {
        const messages: ChatMessage[] = [
            { role: 'user', content: 'hello' },
            { role: 'council', content: 'hi there' },
        ];
        expect(presentMissionChat(messages)).toEqual(messages);
    });

    it('E2E-T1: never collapses away a proposal card sharing a run_id with a later completion card', () => {
        // Regression: once confirmProposal() records a run_id on the
        // original proposal message (E2E-T1 packet, cortexStoreProposalExecutionSlice.ts),
        // that message and the appended terminal system message share the
        // same operational key. The proposal card (with its own fixed
        // identity per intent_proof_id) must stay visible — it is the only
        // place "Confirmed, waiting for result" / "Approved, still running"
        // is rendered, and dropping it lets the terminal card look like an
        // unqualified completion claim.
        const messages: ChatMessage[] = [
            {
                role: 'council',
                content: 'I can start that.',
                mode: 'execution_result',
                run_id: 'run-button-approval',
                proposal_status: 'confirmed_pending_execution',
                // cortexStoreProposalExecutionSlice.ts sets this thread_events
                // entry on the FIRST confirm-action update (approval sent),
                // before the run_id is known; it is not cleared by the SECOND
                // update that adds the run_id, so a real proposal message
                // carries both a run_id and a stale thread_events entry.
                thread_events: [runningEvent('proof-1')],
                proposal: {
                    intent: 'Create a simple python file named hello_world.py in the workspace.',
                    teams: 1,
                    agents: 1,
                    tools: [],
                    risk_level: 'low',
                    confirm_token: 'token-1',
                    intent_proof_id: 'proof-1',
                },
            },
            {
                role: 'system',
                content: 'Soma started the work.',
                mode: 'execution_result',
                run_id: 'run-button-approval',
                thread_events: [runningEvent('run-button-approval')],
            },
        ];

        const presented = presentMissionChat(messages);

        expect(presented).toHaveLength(2);
        expect(presented.some((m) => m.proposal)).toBe(true);
    });

    it('E2E-T1: still collapses an executed proposal card into its terminal completion card (no duplicate result)', () => {
        // Once the proposal reaches "executed", ProposedActionBlock renders
        // its own completed-state proof (first-demo-success.spec.ts,
        // ui-finalization-browser-package-retry.spec.ts), which duplicates
        // the appended terminal system card if both stay on screen. Keep
        // the existing collapse-by-run_id behavior here.
        const messages: ChatMessage[] = [
            {
                role: 'council',
                content: 'I can start that.',
                mode: 'execution_result',
                run_id: 'run-first-demo-package-success',
                proposal_status: 'executed',
                thread_events: [runningEvent('proof-2')],
                proposal: {
                    intent: 'Create the exact first-demo playable browser game package.',
                    teams: 1,
                    agents: 1,
                    tools: [],
                    risk_level: 'low',
                    confirm_token: 'token-2',
                    intent_proof_id: 'proof-2',
                },
            },
            {
                role: 'system',
                content: 'Run run-first-demo-package-success completed.',
                mode: 'execution_result',
                run_id: 'run-first-demo-package-success',
                thread_events: [completedEvent('run-first-demo-package-success')],
            },
        ];

        const presented = presentMissionChat(messages);

        expect(presented).toHaveLength(1);
        expect(presented[0].proposal).toBeUndefined();
    });

    it('still collapses within a single contiguous system-message run (existing behavior)', () => {
        const messages: ChatMessage[] = [
            {
                role: 'system',
                content: 'running',
                run_id: 'run-2',
                thread_events: [runningEvent('run-2')],
            },
            {
                role: 'system',
                content: 'completed',
                run_id: 'run-2',
                thread_events: [completedEvent('run-2')],
            },
        ];
        const presented = presentMissionChat(messages);
        expect(presented).toHaveLength(1);
        expect(presented[0].content).toBe('completed');
    });
});
