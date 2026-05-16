import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

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

describe('MissionControlChat output contracts', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('renders specialist-generated artifacts returned through Soma chat', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        payload: {
                            text: 'I prepared two sample outputs for review.',
                            ask_class: 'governed_artifact',
                            artifacts: [
                                {
                                    id: 'img-1',
                                    type: 'image',
                                    title: 'Homepage Moodboard',
                                    content_type: 'image/png',
                                    content: 'cG5n',
                                    cached: true,
                                },
                                {
                                    id: 'doc-1',
                                    type: 'document',
                                    title: 'Creative Brief',
                                    content_type: 'text/markdown',
                                    content: '# Brief',
                                },
                            ],
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Create two sample outputs for the homepage' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Homepage Moodboard')).toBeDefined();
            expect(screen.getByText('Creative Brief')).toBeDefined();
            expect(screen.getByTitle('Save image to workspace/saved-media')).toBeDefined();
            expect(screen.getByText('Artifact result')).toBeDefined();
            expect(screen.getByText('Soma prepared 2 artifacts for review: Homepage Moodboard and Creative Brief.')).toBeDefined();
        });
    });

    it('uses a readable fallback when Soma returns no text but includes artifacts', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        payload: {
                            text: '',
                            artifacts: [
                                {
                                    id: 'img-2',
                                    type: 'image',
                                    title: 'System Snapshot',
                                    content_type: 'image/png',
                                    content: 'cG5n',
                                },
                            ],
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Show me the current system state visually' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Soma prepared output for review below.')).toBeDefined();
            expect(screen.getByText('System Snapshot')).toBeDefined();
            expect(screen.getByText('Soma prepared 1 artifact for review: System Snapshot.')).toBeDefined();
        });
    });

    it('uses a readable fallback when Soma returns an empty answer without artifacts', async () => {
        useCortexStore.setState({
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) {
                return okJson({ ok: true, data: COUNCIL_MEMBERS });
            }
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        ...CTS_CHAT_RESPONSE.data,
                        payload: { text: '' },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const input = screen.getByPlaceholderText(/Ask Soma/i);
        fireEvent.change(input, { target: { value: 'Any organizations launched?' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText(/could not produce a readable reply/i)).toBeDefined();
        });
    });

    it('saves cached image artifact to workspace folder from inline card', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Generated image',
                    source_node: 'admin',
                    artifacts: [
                        {
                            id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
                            type: 'image',
                            title: 'Generated: test',
                            content_type: 'image/png',
                            content: 'cG5n',
                            cached: true,
                        },
                    ],
                },
            ],
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = String(input);
            if (url.includes('/api/v1/artifacts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/save')) {
                return {
                    ok: true,
                    json: async () => ({ id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', file_path: 'saved-media/test.png' }),
                } as any;
            }
            if (url.includes('/api/v1/workspace/files/reveal')) {
                return {
                    ok: true,
                    json: async () => ({ ok: true, data: { workspace_path: 'saved-media/test.png' } }),
                } as any;
            }
            return {
                ok: true,
                json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
            } as any;
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        fireEvent.click(screen.getByTitle('Save image to workspace/saved-media'));

        await waitFor(() => {
            expect(screen.getByText(/Saved to:/i)).toBeDefined();
        });

        const savedLink = screen.getByRole('link', { name: 'saved-media/test.png' });
        expect(savedLink.getAttribute('href')).toBe('/api/v1/artifacts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/download');

        fireEvent.click(screen.getByRole('button', { name: 'Open local folder for saved-media/test.png' }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=saved-media%2Ftest.png', { method: 'POST' });
        });
    });

    it('shows a clickable saved path for binary file artifacts', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'The packaged audio file is ready for download.',
                    source_node: 'admin',
                    artifacts: [
                        {
                            id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
                            type: 'file',
                            title: 'campaign-voiceover.wav',
                            content_type: 'audio/wav',
                            saved_path: 'saved-media/campaign-voiceover.wav',
                        },
                    ],
                },
            ],
        });

        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = String(input);
            if (url.includes('/api/v1/workspace/files/reveal')) {
                return {
                    ok: true,
                    json: async () => ({ ok: true, data: { workspace_path: 'saved-media/campaign-voiceover.wav' } }),
                } as any;
            }
            return {
                ok: true,
                json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
            } as any;
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        const savedObjectLink = screen.getByRole('link', { name: 'saved-media/campaign-voiceover.wav' });
        expect(savedObjectLink.getAttribute('href')).toBe('/api/v1/artifacts/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb/download');
        expect(screen.getByText(/Saved object:/i)).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: 'Open local folder for saved-media/campaign-voiceover.wav' }));
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspace/files/reveal?path=saved-media%2Fcampaign-voiceover.wav', { method: 'POST' });
        });
    });
});
