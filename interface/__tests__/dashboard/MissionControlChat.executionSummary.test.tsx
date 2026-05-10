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
            expect(screen.getByText('Directed execution package')).toBeDefined();
            expect(screen.getByText('directed_execution')).toBeDefined();
            expect(screen.getByText('workflow.launch')).toBeDefined();
            expect(screen.getByText('Operations Team')).toBeDefined();
            expect(screen.getByRole('link', { name: /Audit proof/i }).getAttribute('href')).toBe('/proof/proof-123');
            expect(screen.getByRole('link', { name: /Onboarding run package/i }).getAttribute('href')).toBe('/runs/run-123');
            expect(screen.getByText('Review the generated package before notifying operators.')).toBeDefined();
        });
    });

    it('renders direct search tool-assisted Soma execution proof', async () => {
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
                            capability_use: {
                                capabilities: ['web_search'],
                                tools: ['web_search'],
                            },
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
            expect(screen.getByRole('link', { name: /Run search-r/i }).getAttribute('href')).toBe('/runs/search-run-123');
            expect(screen.getByText('tool_assisted_work')).toBeDefined();
            expect(screen.getAllByText('web_search').length).toBeGreaterThan(0);
            expect(screen.getByText(/completed/i)).toBeDefined();
            expect(screen.getByRole('link', { name: /Search proof/i }).getAttribute('href')).toBe('/runs/search-proof');
        });
    });
});
