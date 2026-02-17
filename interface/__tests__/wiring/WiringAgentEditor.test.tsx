import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

import WiringAgentEditor from '@/components/wiring/WiringAgentEditor';
import type { AgentManifest, MissionStatus } from '@/store/useCortexStore';

function makeAgent(overrides?: Partial<AgentManifest>): AgentManifest {
    return {
        id: 'test-agent',
        role: 'cognitive',
        system_prompt: 'You are a test agent.',
        model: 'qwen2.5-coder:7b-instruct',
        tools: ['read_file', 'write_file'],
        inputs: ['team.input.general'],
        outputs: ['team.output.general'],
        ...overrides,
    };
}

describe('WiringAgentEditor', () => {
    it('renders form fields for agent editing', () => {
        const agent = makeAgent();

        render(
            <WiringAgentEditor
                teamIdx={0}
                agentIdx={0}
                agent={agent}
                missionStatus={'draft' as MissionStatus}
                onClose={vi.fn()}
                onSave={vi.fn()}
                onDelete={vi.fn()}
            />,
        );

        // Header should show "Edit Agent"
        expect(screen.getByText('Edit Agent')).toBeDefined();

        // Agent ID input should be pre-populated
        const agentIdInput = screen.getByPlaceholderText('agent-name') as HTMLInputElement;
        expect(agentIdInput.value).toBe('test-agent');

        // Role select should exist with options
        const roleSelect = screen.getByDisplayValue('cognitive') as HTMLSelectElement;
        expect(roleSelect).toBeDefined();

        // System prompt textarea
        const promptTextarea = screen.getByPlaceholderText('You are a...') as HTMLTextAreaElement;
        expect(promptTextarea.value).toBe('You are a test agent.');

        // Model input
        const modelInput = screen.getByPlaceholderText('qwen2.5-coder:7b-instruct') as HTMLInputElement;
        expect(modelInput.value).toBe('qwen2.5-coder:7b-instruct');

        // Save and Cancel buttons
        expect(screen.getByText('Save')).toBeDefined();
        expect(screen.getByText('Cancel')).toBeDefined();
    });

    it('save action calls onSave with updated values', () => {
        const agent = makeAgent();
        const onSave = vi.fn();

        render(
            <WiringAgentEditor
                teamIdx={1}
                agentIdx={2}
                agent={agent}
                missionStatus={'draft' as MissionStatus}
                onClose={vi.fn()}
                onSave={onSave}
                onDelete={vi.fn()}
            />,
        );

        // Click the Save button
        fireEvent.click(screen.getByText('Save'));

        // onSave should have been called with teamIdx, agentIdx, and updates
        expect(onSave).toHaveBeenCalledTimes(1);
        expect(onSave).toHaveBeenCalledWith(
            1,
            2,
            expect.objectContaining({
                id: 'test-agent',
                role: 'cognitive',
            }),
        );
    });

    it('delete action triggers confirmation then calls onDelete', () => {
        const agent = makeAgent();
        const onDelete = vi.fn();

        render(
            <WiringAgentEditor
                teamIdx={0}
                agentIdx={1}
                agent={agent}
                missionStatus={'draft' as MissionStatus}
                onClose={vi.fn()}
                onSave={vi.fn()}
                onDelete={onDelete}
            />,
        );

        // The delete button (Trash2 icon) is rendered but does not show "Confirm?" initially
        const deleteButtons = screen.queryAllByText('Confirm?');
        expect(deleteButtons.length).toBe(0);

        // Find the delete button by its bg-cortex-danger styling (unique to footer)
        // Tool-remove buttons have hover:text-cortex-danger but NOT bg-cortex-danger
        const allButtons = screen.getAllByRole('button');
        const deleteBtn = allButtons.find(
            (btn) => btn.className.includes('bg-cortex-danger'),
        );
        expect(deleteBtn).toBeDefined();

        // First click triggers confirmation state
        fireEvent.click(deleteBtn!);
        expect(deleteBtn!.textContent).toContain('Confirm?');
        expect(onDelete).not.toHaveBeenCalled();

        // Second click actually deletes
        fireEvent.click(deleteBtn!);
        expect(onDelete).toHaveBeenCalledTimes(1);
        expect(onDelete).toHaveBeenCalledWith(0, 1);
    });
});
