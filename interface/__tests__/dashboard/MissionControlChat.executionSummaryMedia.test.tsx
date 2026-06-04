import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('reactflow', async () => import('../mocks/reactflow'));

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import { COUNCIL_MEMBERS, resetMissionControlChatStore } from './support/missionControlChatTestUtils';

describe('MissionControlChat execution summary media previews', () => {
    it('previews linked media outputs inside the result card', async () => {
        resetMissionControlChatStore();
        const imageHref = '/api/v1/workspace/files/view?path=workspace%2Fsaved-media%2Flaunch-hero.png';
        const audioHref = '/api/v1/workspace/files/view?path=workspace%2Fsaved-media%2Flaunch-voiceover.wav';

        useCortexStore.setState({
            missionChat: [{
                role: 'system',
                content: 'Mission activated',
                mode: 'execution_result',
                run_id: 'run-media-123456',
                execution_summary: {
                    execution: {
                        shape: 'team_execution',
                        status: 'verified',
                        summary: 'Generated media outputs and retained them for operator review.',
                    },
                    outputs: [
                        { id: 'workspace/saved-media/launch-hero.png', kind: 'image', title: 'Launch hero image', href: imageHref, retained: true },
                        { id: 'workspace/saved-media/launch-voiceover.wav', kind: 'audio', title: 'Launch voiceover', href: audioHref, retained: true },
                    ],
                    proof: [{ run_id: 'run-media-123456' }],
                },
            }],
            councilMembers: COUNCIL_MEMBERS,
            councilTarget: 'admin',
        });

        const { container } = render(<MissionControlChat simpleMode />);

        expect((await screen.findByAltText('Launch hero image')).getAttribute('src')).toBe(imageHref);
        expect(container.querySelector(`audio[src="${audioHref}"]`)).toBeTruthy();
        expect(screen.getByRole('button', { name: /Open local folder for Launch hero image/i })).toBeDefined();
        expect(screen.queryByRole('button', { name: /Open local folder for Launch voiceover/i })).toBeNull();
    });
});
