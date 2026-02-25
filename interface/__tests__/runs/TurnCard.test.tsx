import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';

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
}));

import TurnCard from '@/components/runs/TurnCard';
import type { ConversationTurn } from '@/types/conversations';

function makeTurn(overrides: Partial<ConversationTurn> = {}): ConversationTurn {
    return {
        id: 'turn-1',
        session_id: 'sess-1',
        agent_id: 'admin',
        turn_index: 0,
        role: 'user',
        content: 'Hello world',
        created_at: new Date().toISOString(),
        ...overrides,
    };
}

describe('TurnCard', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders user role with correct border color and label', async () => {
        const turn = makeTurn({ role: 'user', content: 'User message here' });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('USER')).toBeDefined();
        expect(screen.getByText('User message here')).toBeDefined();
        // Check the border class is applied (border-l-cyan-500 for user)
        const card = screen.getByText('USER').closest('div[class*="border-l"]');
        expect(card?.className).toContain('border-l-cyan-500');
    });

    it('renders assistant role with provider/model badge', async () => {
        const turn = makeTurn({
            role: 'assistant',
            content: 'Assistant reply',
            provider_id: 'ollama',
            model_used: 'qwen2.5',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('ASSISTANT')).toBeDefined();
        // brainDisplayName mock returns the id as-is, so "ollama / qwen2.5"
        expect(screen.getByText('ollama / qwen2.5')).toBeDefined();
        expect(screen.getByText('Assistant reply')).toBeDefined();
    });

    it('renders tool_call role with tool name badge', async () => {
        const turn = makeTurn({
            role: 'tool_call',
            content: 'Calling tool...',
            tool_name: 'consult_council',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('TOOL CALL')).toBeDefined();
        expect(screen.getByText('consult_council')).toBeDefined();
        expect(screen.getByText('Calling tool...')).toBeDefined();
    });

    it('renders tool_result role correctly', async () => {
        const turn = makeTurn({
            role: 'tool_result',
            content: 'Tool returned data',
            tool_name: 'search_memory',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('TOOL RESULT')).toBeDefined();
        expect(screen.getByText('search_memory')).toBeDefined();
        expect(screen.getByText('Tool returned data')).toBeDefined();
        // tool_result has ml-6 indent
        const card = screen.getByText('TOOL RESULT').closest('div[class*="border-l"]');
        expect(card?.className).toContain('ml-6');
    });

    it('renders interjection role with "Operator Interjection" label', async () => {
        const turn = makeTurn({
            role: 'interjection',
            content: 'Override: focus on security',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('Operator Interjection')).toBeDefined();
        expect(screen.getByText('Override: focus on security')).toBeDefined();
        // Interjection uses red border
        const card = screen.getByText('Operator Interjection').closest('div[class*="border-l"]');
        expect(card?.className).toContain('border-l-red-500');
    });

    it('system role renders collapsed by default (expandable)', async () => {
        const turn = makeTurn({
            role: 'system',
            content: 'System prompt: You are an AI assistant.',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText('SYSTEM')).toBeDefined();
        // Content is NOT visible because system is collapsed by default
        expect(screen.queryByText('System prompt: You are an AI assistant.')).toBeNull();

        // Click to expand
        const header = screen.getByText('SYSTEM').closest('div[class*="cursor-pointer"]');
        expect(header).toBeDefined();
        await act(async () => { fireEvent.click(header!); });

        // Now content is visible
        expect(screen.getByText('System prompt: You are an AI assistant.')).toBeDefined();
    });

    it('shows consultation_of badge when present', async () => {
        const turn = makeTurn({
            role: 'tool_result',
            content: 'Architect analysis complete',
            tool_name: 'consult_council',
            consultation_of: 'council-architect',
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        // The consultation badge renders "consulted: {consultation_of}"
        expect(screen.getByText(/consulted:/)).toBeDefined();
        expect(screen.getByText(/council-architect/)).toBeDefined();
    });

    it('shows tool_args when expanded (if present)', async () => {
        const turn = makeTurn({
            role: 'tool_call',
            content: 'Calling tool...',
            tool_name: 'consult_council',
            tool_args: { member: 'council-architect', question: 'Design review?' },
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        // Args are collapsed by default — the "arguments" toggle should be present
        expect(screen.getByText('arguments')).toBeDefined();

        // Args content should NOT be visible yet
        expect(screen.queryByText(/"member"/)).toBeNull();

        // Click the arguments toggle to expand
        const argsButton = screen.getByText('arguments');
        await act(async () => { fireEvent.click(argsButton); });

        // Now the JSON stringified args should be visible
        expect(screen.getByText(/"member"/)).toBeDefined();
        expect(screen.getByText(/"council-architect"/)).toBeDefined();
    });

    it('renders content text correctly', async () => {
        const content = 'This is a detailed response from the AI agent.';
        const turn = makeTurn({
            role: 'assistant',
            content,
        });
        await act(async () => { render(<TurnCard turn={turn} />); });

        expect(screen.getByText(content)).toBeDefined();
    });
});
