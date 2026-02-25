import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/navigation
vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/runs/test-run-1',
    useSearchParams: () => new URLSearchParams(),
}));

// Mock labels
vi.mock('@/lib/labels', () => ({
    brainDisplayName: (id: string) => id,
    brainLocationLabel: (l: string) => l,
}));

// Mock Zustand store
const mockFetchRunConversation = vi.fn();
const mockInterjectInRun = vi.fn();
const mockTurns = [
    {
        id: 'turn-1', session_id: 'sess-1', agent_id: 'admin', turn_index: 0,
        role: 'user', content: 'Hello agent', created_at: new Date().toISOString(),
    },
    {
        id: 'turn-2', session_id: 'sess-1', agent_id: 'admin', turn_index: 1,
        role: 'assistant', content: 'Hi there!', provider_id: 'ollama', model_used: 'qwen2.5',
        created_at: new Date().toISOString(),
    },
    {
        id: 'turn-3', session_id: 'sess-1', agent_id: 'council-architect', turn_index: 2,
        role: 'tool_call', content: '{"tool_call": ...}', tool_name: 'consult_council',
        tool_args: { member: 'council-architect' }, created_at: new Date().toISOString(),
    },
    {
        id: 'turn-4', session_id: 'sess-1', agent_id: 'admin', turn_index: 3,
        role: 'interjection', content: 'Focus on database', created_at: new Date().toISOString(),
    },
];

let storeState: Record<string, unknown> = {};

vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = {
            conversationTurns: storeState.conversationTurns ?? null,
            isFetchingConversation: storeState.isFetchingConversation ?? false,
            fetchRunConversation: mockFetchRunConversation,
            interjectInRun: mockInterjectInRun,
        };
        return selector(state);
    },
}));

import ConversationLog from '@/components/runs/ConversationLog';

describe('ConversationLog', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        storeState = {};
    });

    it('shows loading state when fetching and no turns', async () => {
        storeState = { isFetchingConversation: true, conversationTurns: null };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('Loading conversation...')).toBeDefined();
    });

    it('shows empty state when turns array is empty', async () => {
        storeState = { conversationTurns: [] };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('No conversation data')).toBeDefined();
    });

    it('renders turns when data is available', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('Hello agent')).toBeDefined();
        expect(screen.getByText('Hi there!')).toBeDefined();
    });

    it('renders interjection turn with label', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('Operator Interjection')).toBeDefined();
        expect(screen.getByText('Focus on database')).toBeDefined();
    });

    it('shows agent filter buttons when multiple agents', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('All')).toBeDefined();
        // "admin" appears on both filter button and turn badges; verify at least 2 occurrences
        const adminElements = screen.getAllByText('admin');
        expect(adminElements.length).toBeGreaterThanOrEqual(2);
        // council-architect appears on filter + turn badge
        const architectElements = screen.getAllByText('council-architect');
        expect(architectElements.length).toBeGreaterThanOrEqual(1);
    });

    it('shows interjection input when run is running', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" runStatus="running" />); });
        const input = screen.getByPlaceholderText('Interject in this run...');
        expect(input).toBeDefined();
        expect(screen.getByText('Interject')).toBeDefined();
    });

    it('hides interjection input when run is not running', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" runStatus="completed" />); });
        expect(screen.queryByPlaceholderText('Interject in this run...')).toBeNull();
    });

    it('calls fetchRunConversation on mount', async () => {
        storeState = { conversationTurns: [] };
        await act(async () => { render(<ConversationLog runId="run-42" />); });
        expect(mockFetchRunConversation).toHaveBeenCalledWith('run-42', undefined);
    });

    it('renders tool name badge for tool_call', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('consult_council')).toBeDefined();
    });

    it('renders provider/model badge for assistant', async () => {
        storeState = { conversationTurns: mockTurns };
        await act(async () => { render(<ConversationLog runId="run-1" />); });
        expect(screen.getByText('ollama / qwen2.5')).toBeDefined();
    });
});
