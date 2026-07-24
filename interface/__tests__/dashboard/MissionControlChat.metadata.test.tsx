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

    it('hides engine trace chrome in the simplified Soma dashboard mode', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Here is the concise answer a new user should read first.',
                    source_node: 'admin',
                    trust_score: 0.5,
                    ask_class: 'specialist_consultation',
                    tools_used: ['search_memory', 'list_teams'],
                    consultations: [
                        { member: 'council-architect', summary: 'Prefer a calmer first screen.' },
                    ],
                },
            ],
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByText('Here is the concise answer a new user should read first.')).toBeDefined();
        expect(screen.queryByText('C:0.5')).toBeNull();
        expect(screen.queryByText('Search Memory')).toBeNull();
        expect(screen.queryByText('View Teams')).toBeNull();
        expect(screen.getByText('Specialist support')).toBeDefined();
        expect(screen.getByText(/Soma checked with Architect/i)).toBeDefined();
    });

    it('keeps artifact and specialist evidence together in simplified mode', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'The launch package is ready for review.',
                    source_node: 'admin',
                    ask_class: 'governed_artifact',
                    tools_used: ['store_artifact'],
                    consultations: [
                        { member: 'council-creative', summary: 'Keep the usage note concise.' },
                    ],
                    artifacts: [
                        {
                            id: 'launch-brief',
                            type: 'document',
                            title: 'Launch brief',
                            content_type: 'text/markdown',
                            content: '# Launch brief',
                        },
                    ],
                },
            ],
        });

        render(<MissionControlChat simpleMode />);
        await settleMissionControlChat();

        expect(screen.getByText('Artifact result')).toBeDefined();
        expect(screen.getByText('Specialist support')).toBeDefined();
        expect(screen.getByText(
            'Soma checked with Creative while shaping this answer: Keep the usage note concise.',
        )).toBeDefined();
        expect(screen.queryByText('Store Artifact')).toBeNull();
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
