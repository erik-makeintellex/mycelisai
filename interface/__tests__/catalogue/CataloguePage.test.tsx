import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import type { CatalogueAgent } from '@/store/useCortexStore';

type AgentCardProps = {
    agent: CatalogueAgent;
    onSelect: (agent: CatalogueAgent) => void;
};

// Mock child components to isolate CataloguePage logic
vi.mock('@/components/catalogue/AgentCard', () => ({
    __esModule: true,
    default: ({ agent, onSelect }: AgentCardProps) => (
        <div data-testid={`agent-card-${agent.id}`}>
            <span>{agent.name}</span>
            <span>{agent.role}</span>
            <button data-testid={`select-${agent.id}`} onClick={() => onSelect(agent)}>
                Select
            </button>
        </div>
    ),
}));

vi.mock('@/components/catalogue/AgentEditorDrawer', () => ({
    __esModule: true,
    default: ({ agent, onClose }: { agent: CatalogueAgent | null; onClose: () => void }) => (
        <div data-testid="editor-drawer">
            <span>{agent ? `Editing: ${agent.name}` : 'Creating new agent'}</span>
            <button onClick={onClose}>Close</button>
        </div>
    ),
}));

import CataloguePage from '@/components/catalogue/CataloguePage';
import { takePendingSomaPrompt } from '@/components/soma/somaPromptHandoff';
import { useCortexStore } from '@/store/useCortexStore';

const mockAgents: CatalogueAgent[] = [
    {
        id: 'agent-001',
        name: 'Coder Bot',
        role: 'cognitive',
        source: 'built_in',
        locked: true,
        system_prompt: 'You are a coder.',
        model: 'qwen2.5-coder:7b-instruct',
        tools: ['read_file', 'write_file'],
        inputs: ['team.input.code'],
        outputs: ['team.output.code'],
        verification_rubric: [],
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
    },
    {
        id: 'agent-002',
        name: 'Sensor Node',
        role: 'sensory',
        source: 'user',
        locked: false,
        tools: [],
        inputs: [],
        outputs: ['sensor.data'],
        verification_rubric: [],
        created_at: '2025-01-02T00:00:00Z',
        updated_at: '2025-01-02T00:00:00Z',
    },
];

describe('CataloguePage', () => {
    beforeEach(() => {
        window.history.replaceState({}, '', '/dashboard');
        window.sessionStorage.clear();
        useCortexStore.setState({
            catalogueAgents: [],
            isFetchingCatalogue: false,
            selectedCatalogueAgent: null,
            fetchCatalogue: vi.fn(),
            createCatalogueAgent: vi.fn(),
            updateCatalogueAgent: vi.fn(),
            deleteCatalogueAgent: vi.fn(),
            selectCatalogueAgent: vi.fn(),
        });
    });

    it('renders agent cards from the catalogue', () => {
        useCortexStore.setState({
            catalogueAgents: mockAgents,
        });

        render(<CataloguePage />);

        // The default view shows only ready-made profiles.
        expect(screen.getByTestId('agent-card-agent-001')).toBeDefined();
        expect(screen.getByText('Coder Bot')).toBeDefined();
        expect(screen.queryByTestId('agent-card-agent-002')).toBeNull();
        expect(screen.getByText('Worker profiles')).toBeDefined();

        expect(screen.queryByText('Sensor Node')).toBeNull();
        expect(screen.getByText('1 ready-made profiles')).toBeDefined();
    });

    it('hands custom profile creation to Soma without writing a legacy row', () => {
        useCortexStore.setState({
            catalogueAgents: mockAgents,
        });

        render(<CataloguePage />);

        const newBtn = screen.getByText('Create with Soma');
        expect(newBtn).toBeDefined();

        // No drawer initially
        expect(screen.queryByTestId('editor-drawer')).toBeNull();

        fireEvent.click(newBtn);
        expect(screen.queryByTestId('editor-drawer')).toBeNull();
        expect(takePendingSomaPrompt()).toContain('create a reusable Worker Profile');
    });

    it('opens ready-made profile inspection', () => {
        const selectAgent = vi.fn();
        useCortexStore.setState({
            catalogueAgents: mockAgents,
            selectCatalogueAgent: selectAgent,
        });

        render(<CataloguePage />);
        fireEvent.click(screen.getByTestId('select-agent-001'));
        expect(selectAgent).toHaveBeenCalledWith(mockAgents[0]);
    });
});
