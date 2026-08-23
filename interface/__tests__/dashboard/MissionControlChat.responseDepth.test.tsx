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

describe('MissionControlChat adaptive response depth', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue(okJson({ ok: true, data: COUNCIL_MEMBERS }));
    });

    it('renders adaptive answer depth without turning answers into proposal controls', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS, councilTarget: 'admin' });
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    payload: {
                        text: '| Source | Use |\n| --- | --- |\n| Docs | Explain setup |',
                        response_depth: 'quick_box',
                    },
                },
            });
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();
        const input = screen.getByPlaceholderText(/Tell Soma/i);
        fireEvent.change(input, { target: { value: 'Give me a table of setup sources' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Quick answer')).toBeDefined();
            expect(screen.getByRole('table')).toBeDefined();
            expect(screen.queryByRole('button', { name: /Start this/i })).toBeNull();
            expect(screen.queryByRole('button', { name: /Approve/i })).toBeNull();
        });
    });

    it('renders decision brief depth as guidance without approval state', async () => {
        useCortexStore.setState({ councilMembers: COUNCIL_MEMBERS, councilTarget: 'admin' });
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/council/members')) return okJson({ ok: true, data: COUNCIL_MEMBERS });
            if (!url.includes('/api/v1/chat')) return errorText(404, 'not found');
            return okJson({
                ok: true,
                data: {
                    ...CTS_CHAT_RESPONSE.data,
                    payload: {
                        text: '**Recommendation:** keep Soma conversational and expand only on request.',
                        response_depth: 'decision_brief',
                    },
                },
            });
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();
        const input = screen.getByPlaceholderText(/Tell Soma/i);
        fireEvent.change(input, { target: { value: 'What should we do about output depth?' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('Decision brief')).toBeDefined();
            expect(screen.getByText('Recommendation:')).toBeDefined();
            expect(screen.queryByText(/Waiting for executive review/i)).toBeNull();
            expect(screen.queryByRole('button', { name: /Start this/i })).toBeNull();
        });
    });
});
