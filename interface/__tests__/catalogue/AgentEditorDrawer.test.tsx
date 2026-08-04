import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import AgentEditorDrawer from '@/components/catalogue/AgentEditorDrawer';
import type { CatalogueAgent } from '@/store/useCortexStore';

const builtIn: CatalogueAgent = {
    id: 'profile-research',
    profile_key: 'default.researcher',
    name: 'Research Specialist',
    description: 'Finds source-aware evidence.',
    role: 'researcher',
    source: 'built_in',
    locked: true,
    tools: ['web_search'],
    capability_refs: ['web_search'],
    context_bindings: [{ kind: 'public_web', access: 'search' }],
    usage_policy: { selection: 'automatic', scope: 'workspace' },
    inputs: [],
    outputs: ['research_summary'],
    verification_rubric: ['Sources are attributable'],
    created_at: '2026-08-04T00:00:00Z',
    updated_at: '2026-08-04T00:00:00Z',
};

describe('AgentEditorDrawer', () => {
    it('keeps a ready-made profile read-only and exposes governed access before copying', () => {
        const duplicate = vi.fn();
        render(<AgentEditorDrawer agent={builtIn} onClose={vi.fn()} onSave={vi.fn()} onDuplicate={duplicate} />);

        expect(screen.getByRole('dialog', { name: 'Research Specialist profile' })).toBeDefined();
        expect((screen.getByLabelText('Name') as HTMLInputElement).disabled).toBe(true);
        fireEvent.click(screen.getByRole('tab', { name: 'Access & context' }));
        expect(screen.getByText('web_search')).toBeDefined();
        expect((screen.getByLabelText('Context type 1') as HTMLSelectElement).disabled).toBe(true);
        fireEvent.click(screen.getByRole('button', { name: 'Copy profile' }));
        expect(duplicate).toHaveBeenCalledWith(builtIn);
    });

    it('returns capabilities, context, scope, and quality for a new user profile', () => {
        const save = vi.fn();
        render(<AgentEditorDrawer agent={null} onClose={vi.fn()} onSave={save} />);

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Customer Data Researcher' } });
        fireEvent.change(screen.getByLabelText('Purpose'), { target: { value: 'Researches approved customer sources.' } });
        fireEvent.click(screen.getByRole('tab', { name: 'Access & context' }));
        fireEvent.click(screen.getByRole('button', { name: 'Add source' }));
        fireEvent.change(screen.getByLabelText('Context type 1'), { target: { value: 'private_api' } });
        fireEvent.change(screen.getByLabelText('Context reference 1'), { target: { value: 'customer-crm' } });
        fireEvent.change(screen.getByLabelText('Default scope'), { target: { value: 'outcome' } });
        fireEvent.click(screen.getByRole('tab', { name: 'Quality' }));
        fireEvent.change(screen.getByLabelText('Review criteria'), { target: { value: 'Sources named, Claims supported' } });
        fireEvent.click(screen.getByRole('button', { name: 'Create profile' }));

        expect(save).toHaveBeenCalledWith(expect.objectContaining({
            name: 'Customer Data Researcher',
            description: 'Researches approved customer sources.',
            context_bindings: [{ kind: 'private_api', ref: 'customer-crm', access: 'read' }],
            usage_policy: { selection: 'soma_or_manual', scope: 'outcome' },
            verification_rubric: ['Sources named', 'Claims supported'],
        }));
    });
});
