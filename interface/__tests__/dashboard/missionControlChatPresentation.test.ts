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
