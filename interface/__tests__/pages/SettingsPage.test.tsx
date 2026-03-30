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

const mockSearchParams = new URLSearchParams();
const mockAdvancedMode = vi.fn(() => true);
const mockUpdateAssistantName = vi.fn(async () => true);
const mockUpdateTheme = vi.fn(async () => true);
vi.mock('next/navigation', () => ({
    useSearchParams: () => mockSearchParams,
}));
vi.mock('@/store/useCortexStore', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/store/useCortexStore')>();
    return {
        ...actual,
        useCortexStore: (selector: any) => selector({
            advancedMode: mockAdvancedMode(),
            assistantName: 'Soma',
            theme: 'aero-light',
            updateAssistantName: mockUpdateAssistantName,
            updateTheme: mockUpdateTheme,
        }),
    };
});

import SettingsPage from '@/app/(app)/settings/page';

describe('Settings Page (app/settings/page.tsx)', () => {
    beforeEach(() => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({}),
        });
        mockUpdateAssistantName.mockClear();
        mockUpdateTheme.mockClear();
        for (const key of [...mockSearchParams.keys()]) {
            mockSearchParams.delete(key);
        }
        mockAdvancedMode.mockReturnValue(true);
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
        expect(screen.getByText('Mission Profiles')).toBeDefined();
        expect(screen.getByText('People & Access')).toBeDefined();
        expect(screen.getByText('AI Engines')).toBeDefined();
        expect(screen.getByText('Connected Tools')).toBeDefined();
    });

    it('defaults to profile tab with appearance settings', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        expect(screen.getByText('Appearance')).toBeDefined();
        expect(screen.getByRole('option', { name: 'Aero Light' })).toBeDefined();
        expect(screen.getByRole('option', { name: 'Midnight Cortex' })).toBeDefined();
        expect(screen.getByRole('option', { name: 'System' })).toBeDefined();
    });

    it('hides advanced tabs when advanced mode is off', async () => {
        mockAdvancedMode.mockReturnValue(false);
        await act(async () => {
            render(<SettingsPage />);
        });

        expect(screen.queryByText('AI Engines')).toBeNull();
        expect(screen.queryByText('Connected Tools')).toBeNull();
    });

    it('saves the selected theme', async () => {
        const { fireEvent } = await import('@testing-library/react');

        await act(async () => {
            render(<SettingsPage />);
        });

        fireEvent.change(screen.getByLabelText('Theme'), { target: { value: 'midnight-cortex' } });
        await act(async () => {
            fireEvent.click(screen.getAllByRole('button', { name: 'Save' })[1]);
        });

        expect(mockUpdateTheme).toHaveBeenCalledWith('midnight-cortex');
    });
});
