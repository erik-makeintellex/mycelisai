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
                    mode: 'execution_result',
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
            expect(screen.getByText('Output ready')).toBeDefined();
            expect(screen.getByText('Result needs review')).toBeDefined();
            expect(screen.getByText('Review request, proof, and recovery')).toBeDefined();
            expect(screen.getByText('Completed work')).toBeDefined();
            expect(screen.getByText('workflow.launch')).toBeDefined();
            expect(screen.getByText('Operations Team')).toBeDefined();
            expect(screen.getByRole('link', { name: /Audit proof/i }).getAttribute('href')).toBe('/proof/proof-123');
            expect(screen.getByText('Onboarding run package')).toBeDefined();
            expect(screen.getByRole('button', { name: /Open output Onboarding run package in a new browser window/i })).toBeDefined();
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
                            proof: [{ label: 'Search proof', url: '/runs/search-proof', run_id: 'search-run-123', verified: true }],
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
            expect(screen.getByRole('link', { name: /Inspect run receipt search-r/i }).getAttribute('href')).toBe('/runs/search-run-123');
            expect(screen.getByText('Tool-assisted work')).toBeDefined();
            expect(screen.getAllByText('Soma Search').length).toBeGreaterThan(0);
            expect(screen.queryByText('web_search')).toBeNull();
            expect(screen.getByText('Source')).toBeDefined();
            expect(screen.getByText('Search source: Local Mycelis context')).toBeDefined();
            expect(screen.getByText(/completed/i)).toBeDefined();
            expect(screen.getByRole('link', { name: /Search proof/i }).getAttribute('href')).toBe('/runs/search-proof');
        });
    });

    it('keeps a compact trust receipt on quick tool-assisted answers', async () => {
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
                    mode: 'answer',
                    payload: {
                        text: 'Soma found current public-web sources.',
                        ask_class: 'direct_answer',
                        response_depth: 'quick_box',
                        execution_summary: {
                            execution: {
                                shape: 'tool_assisted_work',
                                status: 'completed',
                            },
                            capability_use: [{
                                id: 'web_search',
                                reason: 'Search source: External or public web provider; verify before relying',
                            }],
                            proof: {
                                run_class: 'no_run',
                                proof_class: 'audit_only',
                                audit_event_id: 'audit-search-123',
                                verified: true,
                            },
                        },
                    },
                },
            });
        });
        render(<MissionControlChat />);
        await settleMissionControlChat();
        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Find current sources' } });
        fireEvent.keyDown(input, { key: 'Enter' });
        await waitFor(() => {
            expect(screen.getByText('Soma found current public-web sources.')).toBeDefined();
            expect(screen.getByTestId('execution-summary-receipt')).toBeDefined();
            expect(screen.getByText('Result verified')).toBeDefined();
            expect(screen.queryByTestId('execution-summary-card')).toBeNull();
        });
    });

    it('renders confirmed generated file outputs as openable links on system run messages', async () => {
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
        expect(screen.getByRole('link', { name: /Inspect run receipt/i }).getAttribute('href')).toBe('/runs/run-game-123456');

        expect(screen.getByRole('button', { name: new RegExp(`Open file ${filePath} in Mycelis`) })).toBeDefined();

        fireEvent.click(screen.getByText('Details and follow-up'));
        fireEvent.click(screen.getByRole('button', { name: new RegExp(`Open local folder for ${filePath}`) }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=workspace%2Flogs%2Fqa_team_click_game.html', { method: 'POST' });
        });
    });

    it('renders generated project packages as deliverable review cards', async () => {
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
        expect(screen.getByText('App/package:')).toBeDefined();
        expect(screen.queryByText('entry: workspace/generated/coin-runner/index.html')).toBeNull();
        expect(screen.queryByText('game.js')).toBeNull();

        expect(screen.getByRole('button', { name: /Open app Coin Runner Game in Mycelis/i })).toBeDefined();

        fireEvent.click(screen.getByText('Details and follow-up'));
        expect(screen.getByRole('link', { name: /Open Coin Runner Game in Resources/i }).getAttribute('href')).toBe('/resources?tab=workspace&path=workspace%2Fgenerated%2Fcoin-runner');
        fireEvent.click(screen.getByRole('button', { name: /Open local folder for Coin Runner Game/i }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=workspace%2Fgenerated%2Fcoin-runner', { method: 'POST' });
        });
    });
});
