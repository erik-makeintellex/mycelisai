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

// Mock MCPToolRegistry (loaded via next/dynamic, has deep store dependencies)
vi.mock('@/components/settings/MCPToolRegistry', () => ({
    __esModule: true,
    default: () => <div data-testid="mcp-tool-registry">MCPToolRegistry</div>,
}));

// Mock MatrixGrid (fetches cognitive matrix config from API)
vi.mock('@/components/matrix/MatrixGrid', () => ({
    __esModule: true,
    default: () => <div data-testid="matrix-grid">MatrixGrid</div>,
}));

import SettingsPage from '@/app/settings/page';

describe('Settings Page (app/settings/page.tsx)', () => {
    beforeEach(() => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({}),
        });
    });

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        expect(screen.getByText('Settings')).toBeDefined();
    });

    it('renders tab navigation', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        expect(screen.getByText('Profile')).toBeDefined();
        expect(screen.getByText('Teams')).toBeDefined();
        expect(screen.getByText('Cognitive Matrix')).toBeDefined();
        expect(screen.getByText('MCP Tools')).toBeDefined();
    });

    it('defaults to profile tab with appearance settings', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        expect(screen.getByText('Appearance')).toBeDefined();
        expect(screen.getByText('Vuexy Dark')).toBeDefined();
    });
});
