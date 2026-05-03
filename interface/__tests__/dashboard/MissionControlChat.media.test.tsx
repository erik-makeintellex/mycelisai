import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat media output contracts', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('plays audio and video artifacts directly in the Soma chat window', async () => {
        useCortexStore.setState({
            missionChat: [{
                role: 'council',
                content: 'Media outputs are ready.',
                source_node: 'admin',
                artifacts: [
                    {
                        id: 'audio-1',
                        type: 'audio',
                        title: 'Narration draft',
                        content_type: 'audio/wav',
                        content: 'UklGRg==',
                    },
                    {
                        id: 'video-1',
                        type: 'video',
                        title: 'Workflow preview',
                        content_type: 'video/mp4',
                        url: 'https://example.test/preview.mp4',
                    },
                ],
            }],
        });

        const { container } = render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Narration draft')).toBeDefined();
        expect(screen.getByText('Workflow preview')).toBeDefined();
        expect(container.querySelector('audio[src^="data:audio/wav;base64,"]')).toBeTruthy();
        expect(container.querySelector('video[src="https://example.test/preview.mp4"]')).toBeTruthy();
    });
});
