import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock child components to isolate CataloguePage logic
vi.mock('@/components/catalogue/AgentCard', () => ({
    __esModule: true,
    default: ({ agent, onSelect, onDelete }: any) => (
        <div data-testid={`agent-card-${agent.id}`}>
            <span>{agent.name}</span>
            <span>{agent.role}</span>
            <button data-testid={`select-${agent.id}`} onClick={() => onSelect(agent)}>
                Select
            </button>
            <button data-testid={`delete-${agent.id}`} onClick={() => onDelete(agent.id)}>
                Delete
            </button>
        </div>
    ),
}));

vi.mock('@/components/catalogue/AgentEditorDrawer', () => ({
    __esModule: true,
    default: ({ agent, onClose, onSave }: any) => (
        <div data-testid="editor-drawer">
            <span>{agent ? `Editing: ${agent.name}` : 'Creating new agent'}</span>
            <button onClick={onClose}>Close</button>
        </div>
    ),
}));

import CataloguePage from '@/components/catalogue/CataloguePage';
import { useCortexStore } from '@/store/useCortexStore';
import type { CatalogueAgent } from '@/store/useCortexStore';

const mockAgents: CatalogueAgent[] = [
    {
        id: 'agent-001',
        name: 'Coder Bot',
        role: 'cognitive',
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

        // Both agent cards should render
        expect(screen.getByTestId('agent-card-agent-001')).toBeDefined();
        expect(screen.getByTestId('agent-card-agent-002')).toBeDefined();

        // Agent names should be visible
        expect(screen.getByText('Coder Bot')).toBeDefined();
        expect(screen.getByText('Sensor Node')).toBeDefined();

        // Header should show "Agent Catalogue"
        expect(screen.getByText('Agent Catalogue')).toBeDefined();
    });

    it('New Agent button is present and opens the editor drawer', () => {
        useCortexStore.setState({
            catalogueAgents: mockAgents,
        });

        render(<CataloguePage />);

        // "New Agent" button should be present
        const newBtn = screen.getByText('New Agent');
        expect(newBtn).toBeDefined();

        // No drawer initially
        expect(screen.queryByTestId('editor-drawer')).toBeNull();

        // Click New Agent
        fireEvent.click(newBtn);

        // Drawer should now be open in "create" mode
        expect(screen.getByTestId('editor-drawer')).toBeDefined();
        expect(screen.getByText('Creating new agent')).toBeDefined();
    });

    it('delete action on an agent card calls the store delete action', () => {
        const deleteFn = vi.fn();
        useCortexStore.setState({
            catalogueAgents: mockAgents,
            deleteCatalogueAgent: deleteFn,
        });

        render(<CataloguePage />);

        // Click the delete button on agent-001
        fireEvent.click(screen.getByTestId('delete-agent-001'));

        // The store's deleteCatalogueAgent should have been called
        expect(deleteFn).toHaveBeenCalledTimes(1);
        expect(deleteFn).toHaveBeenCalledWith('agent-001');
    });
});
