import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => import('../mocks/reactflow'));

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    okJson,
    resetMissionControlChatStore,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat failed execution summaries', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue(okJson({ ok: true, data: COUNCIL_MEMBERS }));
    });

    it('renders compact failed-run boundaries without trusted-output language', async () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'council',
                content: 'Soma hit a server-side failure while handling the request.',
                mode: 'blocker',
                run_id: 'run-failed-123456',
                execution_summary: {
                    execution: {
                        shape: 'guided_proposal',
                        status: 'failed',
                        summary: 'Soma could not complete the approved proposal.',
                    },
                    proof: [{ run_id: 'run-failed-123456', proof_class: 'run_and_audit', verified: false }],
                    audit_recovery: {
                        recovery_state: 'failed',
                        blocker: 'tool unavailable',
                        degradation: {
                            code: 'approved_execution_failed',
                            what_failed: 'tool unavailable',
                            trusted_state: 'The failed run record remains trusted.',
                            invalidated_proof: 'No completed output should be trusted.',
                            safe_continuation: 'Review the failed run and retry.',
                            requires_attention: true,
                        },
                    },
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(await screen.findByText('Needs review')).toBeDefined();
        expect(screen.getByText('Could not run')).toBeDefined();
        expect(screen.getByText('failed').className).toContain('text-red-300');
        expect(screen.getByText('Details and proof')).toBeDefined();
        expect(screen.getAllByText('Failed: tool unavailable').length).toBeGreaterThan(0);
        expect(screen.getAllByText('Still available: The failed run record remains trusted.').length).toBeGreaterThan(0);
        expect(screen.getAllByText('Not reliable: No completed output should be trusted.').length).toBeGreaterThan(0);
        expect(screen.getAllByText('Safe next: Review the failed run and retry.').length).toBeGreaterThan(0);
        expect(screen.getAllByRole('link', { name: /Run run-fail/i })
            .some((link) => link.getAttribute('href') === '/runs/run-failed-123456')).toBe(true);
        expect(screen.queryByText('Result saved')).toBeNull();
        expect(screen.queryByText('Result verified')).toBeNull();
    });
});
