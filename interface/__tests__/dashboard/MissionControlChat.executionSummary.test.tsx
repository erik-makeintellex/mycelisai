import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => import('../mocks/reactflow'));

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    CTS_CHAT_RESPONSE,
    errorText,
    okJson,
    requestUrl,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat execution summary', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue(okJson({ ok: true, data: COUNCIL_MEMBERS }));
    });

    it('renders directed execution summary details from Soma chat responses', async () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', {
            value: { writeText },
            configurable: true,
        });

        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    payload: {
                        text: 'Directed execution completed.',
                        execution_summary: {
                            intent: 'Launch the onboarding workflow',
                            understanding: 'Create a compact execution path for the new team.',
                            execution: {
                                shape: 'directed_execution',
                                status: 'complete',
                                summary: 'Routed through Soma with one specialist team.',
                            },
                            capability_use: {
                                capabilities: ['workflow.launch'],
                                teams: ['Operations Team'],
                            },
                            outputs: [{ title: 'Onboarding run package', url: '/runs/run-123' }],
                            proof: [{ label: 'Audit proof', url: '/proof/proof-123' }],
                            audit_recovery: 'Audit event recorded; recovery snapshot available.',
                            next_step: 'Review the generated package before notifying operators.',
                        },
                    },
                },
            });
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Launch onboarding' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Operator trust package')).toBeDefined();
            expect(screen.getByText('Result needs review')).toBeDefined();
            expect(screen.getByText('Completed work')).toBeDefined();
            expect(screen.getByText('workflow.launch')).toBeDefined();
            expect(screen.getByText('Operations Team')).toBeDefined();
            expect(screen.getByRole('link', { name: /Audit proof/i }).getAttribute('href')).toBe('/proof/proof-123');
            expect(screen.getByText('Onboarding run package')).toBeDefined();
            expect(screen.getByRole('button', { name: /Open file Onboarding run package/i })).toBeDefined();
            expect(screen.getByText('Review the generated package before notifying operators.')).toBeDefined();
        });

        fireEvent.click(screen.getByRole('button', { name: /Copy output quote for Onboarding run package/i }));

        await waitFor(() => {
            expect(writeText).toHaveBeenCalledWith('> Onboarding run package\n/runs/run-123');
            expect(screen.getByRole('button', { name: /Copied output quote/i })).toBeDefined();
        });
    });

    it('renders direct search tool-assisted Soma result details', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    mode: 'execution_result',
                    payload: {
                        text: 'Soma searched the web and found a current answer.',
                        tools_used: ['web_search'],
                        execution_summary: {
                            intent: 'Search for current release notes',
                            understanding: 'Use direct Soma search before answering.',
                            execution: {
                                shape: 'tool_assisted_work',
                                status: 'completed',
                                summary: 'Soma used browser-visible search capability proof.',
                            },
                            capability_use: [{
                                id: 'web_search',
                                label: 'web_search',
                                kind: 'tool',
                                reason: 'Search source: Local Mycelis context',
                            }],
                            proof: [{ label: 'Search proof', url: '/runs/search-proof', run_id: 'search-run-123' }],
                            next_step: 'Share the direct search result with the operator.',
                        },
                    },
                },
            });
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Search the web for current release notes' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByTestId('execution-summary-card')).toBeDefined();
            expect(screen.getByText('Result verified')).toBeDefined();
            expect(screen.getByRole('link', { name: /Run search-r/i }).getAttribute('href')).toBe('/runs/search-run-123');
            expect(screen.getByText('Tool-assisted work')).toBeDefined();
            expect(screen.getAllByText('web_search').length).toBeGreaterThan(0);
            expect(screen.getByText('Source')).toBeDefined();
            expect(screen.getByText('Search source: Local Mycelis context')).toBeDefined();
            expect(screen.getByText(/completed/i)).toBeDefined();
            expect(screen.getByRole('link', { name: /Search proof/i }).getAttribute('href')).toBe('/runs/search-proof');
        });
    });

    it('renders confirmed generated file outputs as openable links on system run messages', async () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        const openWindow = vi.spyOn(window, 'open').mockReturnValue(null);
        Object.defineProperty(navigator, 'clipboard', {
            value: { writeText },
            configurable: true,
        });
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (url.includes('/api/v1/workspace/files/reveal')) {
                return okJson({ ok: true, data: { workspace_path: 'logs/qa_team_click_game.html' } });
            }
            return errorText(404, 'not found');
        });
        const filePath = 'workspace/logs/qa_team_click_game.html';
        const href = '/api/v1/workspace/files/view?path=workspace%2Flogs%2Fqa_team_click_game.html';

        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Mission activated',
                mode: 'execution_result',
                run_id: 'run-game-123456',
                execution_summary: {
                    execution: {
                        shape: 'team_execution',
                        status: 'verified',
                        summary: 'Generated a browser click game and retained it for operator review.',
                    },
                    outputs: [{
                        id: filePath,
                        kind: 'code',
                        title: filePath,
                        href,
                        retained: true,
                    }],
                    proof: [{ run_id: 'run-game-123456' }],
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(await screen.findByText(filePath)).toBeDefined();
        expect(screen.getByText('Result saved')).toBeDefined();
        expect(screen.getByRole('link', { name: /Mission activated/i }).getAttribute('href')).toBe('/runs/run-game-123456');

        fireEvent.click(screen.getByRole('button', { name: new RegExp(`Open file ${filePath} in a new browser window`) }));
        expect(openWindow).toHaveBeenCalledWith(href, '_blank', 'noopener,noreferrer');

        fireEvent.click(screen.getByRole('button', { name: new RegExp(`Open local folder for ${filePath}`) }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=workspace%2Flogs%2Fqa_team_click_game.html', { method: 'POST' });
        });

        fireEvent.click(screen.getByRole('button', { name: new RegExp(`Copy output quote for ${filePath}`) }));

        await waitFor(() => {
            expect(writeText).toHaveBeenCalledWith(`> ${filePath}\n${href}`);
        });
    });

    it('renders generated project packages as deliverable review cards', async () => {
        const openWindow = vi.spyOn(window, 'open').mockReturnValue(null);
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (url.includes('/api/v1/workspace/files/reveal')) {
                return okJson({ ok: true, data: { workspace_path: 'workspace/generated/coin-runner' } });
            }
            return errorText(404, 'not found');
        });
        const href = '/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Findex.html';

        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Game package generated',
                mode: 'execution_result',
                run_id: 'run-package-123456',
                execution_summary: {
                    execution: {
                        shape: 'team_execution',
                        status: 'verified',
                        summary: 'Generated a playable browser game package.',
                    },
                    outputs: [{
                        id: 'workspace/generated/coin-runner',
                        kind: 'project_package',
                        title: 'Coin Runner Game',
                        href,
                        entrypoint: 'workspace/generated/coin-runner/index.html',
                        folder: 'workspace/generated/coin-runner',
                        files: ['index.html', 'game.js', 'styles.css', 'README.md'],
                        validation: 'Browser opened and score increased after click.',
                        retained: true,
                    }],
                    proof: [{ run_id: 'run-package-123456' }],
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        render(<MissionControlChat simpleMode />);

        expect(await screen.findByText('Coin Runner Game')).toBeDefined();
        expect(screen.getByText('entry: workspace/generated/coin-runner/index.html')).toBeDefined();
        expect(screen.getByText('folder: workspace/generated/coin-runner')).toBeDefined();
        expect(screen.getByText('game.js')).toBeDefined();
        expect(screen.getByText('Browser opened and score increased after click.')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /Open Game Coin Runner Game in a new browser window/i }));
        expect(openWindow).toHaveBeenCalledWith(href, '_blank', 'noopener,noreferrer');

        fireEvent.click(screen.getByRole('button', { name: /Open local folder for Coin Runner Game/i }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=workspace%2Fgenerated%2Fcoin-runner', { method: 'POST' });
        });
    });

    it('renders plain failed-run boundaries for failed run messages', async () => {
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
        expect(screen.getByText('Failed: tool unavailable')).toBeDefined();
        expect(screen.getByText('Still available: The failed run record remains trusted.')).toBeDefined();
        expect(screen.getByText('Not reliable: No completed output should be trusted.')).toBeDefined();
        expect(screen.getByText('Safe next: Review the failed run and retry.')).toBeDefined();
        expect(screen.getAllByRole('link', { name: /Run run-fail/i })
            .some((link) => link.getAttribute('href') === '/runs/run-failed-123456')).toBe(true);
        expect(screen.queryByText('Result saved')).toBeNull();
        expect(screen.queryByText('Result verified')).toBeNull();
    });
});
