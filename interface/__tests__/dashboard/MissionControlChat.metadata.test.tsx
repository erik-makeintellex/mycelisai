import { describe, it, expect, vi, beforeEach } from 'vitest';
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

describe('MissionControlChat metadata contracts', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('shows a specialist-support badge for consulted answers', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'The architect reviewed the tradeoffs and recommends the safer route.',
                    source_node: 'admin',
                    ask_class: 'specialist_consultation',
                    consultations: [
                        { member: 'council-architect', summary: 'Prefer the safer route.' },
                    ],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Specialist support')).toBeDefined();
        expect(screen.getByText('Soma checked with Architect while shaping this answer: Prefer the safer route.')).toBeDefined();
    });

    it('renders trust badge with correct score', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Response',
                    source_node: 'admin',
                    trust_score: 0.5,
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('C:0.5')).toBeDefined();
    });

    it('renders tools-used pills when present', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'I searched memory',
                    source_node: 'admin',
                    trust_score: 0.5,
                    tools_used: ['search_memory', 'list_teams'],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByText('Search Memory')).toBeDefined();
        expect(screen.getByText('View Teams')).toBeDefined();
    });

    it('does not render tools pills when tools_used is empty', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Simple response',
                    source_node: 'admin',
                    trust_score: 0.5,
                    tools_used: [],
                },
            ],
        });

        const { container } = render(<MissionControlChat />);
        await settleMissionControlChat();

        const pills = container.querySelectorAll('[class*="cortex-primary/10"]');
        expect(pills).toHaveLength(0);
    });
});
