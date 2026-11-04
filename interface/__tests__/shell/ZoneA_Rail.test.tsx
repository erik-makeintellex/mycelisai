import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';

// Mock next/navigation with a configurable pathname
const mockPathname = vi.fn(() => '/dashboard');
vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: vi.fn(),
        replace: vi.fn(),
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    usePathname: () => mockPathname(),
}));

// Mock Zustand store
const mockAdvancedMode = vi.fn(() => false);
const mockToggleAdvancedMode = vi.fn();
const mockSetStatusDrawerOpen = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = {
            advancedMode: mockAdvancedMode(),
            toggleAdvancedMode: mockToggleAdvancedMode,
            setStatusDrawerOpen: mockSetStatusDrawerOpen,
        };
        return selector(state);
    },
}));

import { ZoneA } from '@/components/shell/ZoneA_Rail';

const DEFAULT_NAV_ENTRIES = [
    { href: '/dashboard', label: 'Soma Work' },
    { href: '/teams', label: 'Teams' },
    { href: '/resources', label: 'Durable Files' },
];

describe('ZoneA_Rail (V8.2 Operator vs. Builder Navigation)', () => {
    beforeEach(() => {
        mockPathname.mockReturnValue('/dashboard');
        mockAdvancedMode.mockReturnValue(false);
        mockSetStatusDrawerOpen.mockClear();
        window.localStorage.clear();
    });

    it('renders the Mycelis brand identity', () => {
        render(<ZoneA />);
        expect(screen.getByText('Mycelis')).toBeDefined();
    });

    it('renders streamlined operator navigation entries when advancedMode is off', () => {
        render(<ZoneA />);
        for (const entry of DEFAULT_NAV_ENTRIES) {
            expect(screen.getByText(entry.label)).toBeDefined();
        }
    });

    it('renders Settings in footer', () => {
        render(<ZoneA />);
        expect(screen.getByText('Settings')).toBeDefined();
    });

    it('renders Status drawer trigger in footer', () => {
        render(<ZoneA />);
        expect(screen.getByRole('button', { name: /Status/i })).toBeDefined();
    });

    it('renders a return link for the remembered organization when outside the workspace', async () => {
        window.localStorage.setItem('mycelis-last-organization-id', 'org-123');
        window.localStorage.setItem('mycelis-last-organization-name', 'Northstar Labs');

        render(<ZoneA />);

        await waitFor(() => {
            expect(screen.getByTestId('current-organization-nav')).toBeDefined();
        });
        expect(screen.getByText('Return to Organization')).toBeDefined();
        expect(screen.getByText('Northstar Labs')).toBeDefined();
    });

    it('labels the remembered organization as Current Organization when on its route', async () => {
        window.localStorage.setItem('mycelis-last-organization-id', 'org-123');
        window.localStorage.setItem('mycelis-last-organization-name', 'Northstar Labs');
        mockPathname.mockReturnValue('/organizations/org-123');

        render(<ZoneA />);

        await waitFor(() => {
            expect(screen.getByText('Current Organization')).toBeDefined();
        });
    });

    it('highlights Soma Work when on /dashboard', () => {
        mockPathname.mockReturnValue('/dashboard');
        const { container } = render(<ZoneA />);
        const dashboardLink = container.querySelector('a[href="/dashboard"]');
        expect(dashboardLink?.className).toContain('bg-cortex-primary');
    });

    it('keeps non-active operator entries muted when on /dashboard', () => {
        mockPathname.mockReturnValue('/dashboard');
        const { container } = render(<ZoneA />);
        const resourcesLink = container.querySelector('a[href="/resources"]');
        expect(resourcesLink?.className).toContain('text-cortex-text-muted');
    });

    it('hides builder/runtime links (/activity, /runs, /groups, /system, /docs, /memory) when advancedMode is off', () => {
        mockAdvancedMode.mockReturnValue(false);
        const { container } = render(<ZoneA />);
        expect(container.querySelector('a[href="/groups"]')).toBeNull();
        expect(container.querySelector('a[href="/activity"]')).toBeNull();
        expect(container.querySelector('a[href="/runs"]')).toBeNull();
        expect(container.querySelector('a[href="/system"]')).toBeNull();
        expect(container.querySelector('a[href="/memory"]')).toBeNull();
        expect(container.querySelector('a[href="/docs"]')).toBeNull();
    });

    it('reveals full builder and governance surfaces when advancedMode is on', () => {
        mockAdvancedMode.mockReturnValue(true);
        render(<ZoneA />);
        expect(screen.getByText('Activity History')).toBeDefined();
        expect(screen.getByText('Runs')).toBeDefined();
        expect(screen.getByText('Memory & Continuity')).toBeDefined();
        expect(screen.getByText('Groups')).toBeDefined();
        expect(screen.getByText('Deployments')).toBeDefined();
        expect(screen.getByText('Automations')).toBeDefined();
        expect(screen.getByText('System')).toBeDefined();
        expect(screen.getByText('Docs')).toBeDefined();
    });

    it('links Automations and Runs to real surfaces when advancedMode is on', () => {
        mockAdvancedMode.mockReturnValue(true);
        const { container } = render(<ZoneA />);
        expect(container.querySelector('a[href="/automations"]')).not.toBeNull();
        expect(container.querySelector('a[href="/runs"]')).not.toBeNull();
    });

    it('renders Advanced toggle button', () => {
        render(<ZoneA />);
        expect(screen.getByText('Advanced: Off')).toBeDefined();
    });

    it('does not use bg-white', () => {
        const { container } = render(<ZoneA />);
        expect(container.innerHTML).not.toContain('bg-white');
    });

    it('does not show old navigation entries', () => {
        render(<ZoneA />);
        const oldLabels = ['Neural Wiring', 'Team Management', 'Agent Catalogue', 'Skills Market', 'Governance', 'System Status', 'Cognitive Matrix'];
        for (const label of oldLabels) {
            expect(screen.queryByText(label)).toBeNull();
        }
    });
});
