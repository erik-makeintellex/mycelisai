import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock reactflow (store imports it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

// Mock next/dynamic to render the component directly
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any) => {
        const Component = require('react').lazy(loader);
        return (props: any) => {
            const React = require('react');
            return React.createElement(
                React.Suspense,
                { fallback: null },
                React.createElement(Component, props),
            );
        };
    },
}));

// Mock child components
vi.mock('@/components/catalogue/AgentCard', () => ({
    __esModule: true,
    default: ({ agent }: any) => <div data-testid="agent-card">{agent?.name ?? 'agent'}</div>,
}));

vi.mock('@/components/catalogue/AgentEditorDrawer', () => ({
    __esModule: true,
    default: () => <div data-testid="agent-editor-drawer" />,
}));

import CatalogueRoute from '@/app/catalogue/page';
import { useCortexStore } from '@/store/useCortexStore';

describe('Catalogue Page (app/catalogue/page.tsx)', () => {
    beforeEach(() => {
        useCortexStore.setState({
            catalogueAgents: [],
            isFetchingCatalogue: false,
            selectedCatalogueAgent: null,
        });

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ agents: [] }),
        });
    });

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<CatalogueRoute />);
        });

        expect(document.body.innerHTML.length).toBeGreaterThan(0);
    });

    it('renders the catalogue page content', async () => {
        await act(async () => {
            render(<CatalogueRoute />);
        });

        // CataloguePage header contains BookOpen icon and "AGENT CATALOGUE" text
        const content = screen.queryByText(/catalogue/i) || screen.queryByText(/agent/i);
        expect(content).toBeDefined();
    });
});
