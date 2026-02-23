import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

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
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => {
        const state = {
            advancedMode: mockAdvancedMode(),
            toggleAdvancedMode: mockToggleAdvancedMode,
        };
        return selector(state);
    },
}));

import { ZoneA } from '@/components/shell/ZoneA_Rail';

const V7_NAV_ENTRIES = [
    { href: '/dashboard', label: 'Mission Control' },
    { href: '/automations', label: 'Automations' },
    { href: '/resources', label: 'Resources' },
    { href: '/memory', label: 'Memory' },
];

describe('ZoneA_Rail (V7 Workflow-First Navigation)', () => {
    beforeEach(() => {
        mockPathname.mockReturnValue('/dashboard');
        mockAdvancedMode.mockReturnValue(false);
    });

    it('renders the Mycelis brand name', () => {
        render(<ZoneA />);
        expect(screen.getByText('Mycelis')).toBeDefined();
    });

    it('renders all primary navigation links', () => {
        render(<ZoneA />);
        for (const entry of V7_NAV_ENTRIES) {
            expect(screen.getByText(entry.label)).toBeDefined();
        }
    });

    it('renders Settings in footer', () => {
        render(<ZoneA />);
        expect(screen.getByText('Settings')).toBeDefined();
    });

    it('highlights active route with cortex-primary', () => {
        mockPathname.mockReturnValue('/dashboard');
        const { container } = render(<ZoneA />);
        const dashboardLink = container.querySelector('a[href="/dashboard"]');
        expect(dashboardLink?.className).toContain('bg-cortex-primary');
    });

    it('uses muted text for inactive routes', () => {
        mockPathname.mockReturnValue('/dashboard');
        const { container } = render(<ZoneA />);
        const automationsLink = container.querySelector('a[href="/automations"]');
        expect(automationsLink?.className).toContain('text-cortex-text-muted');
    });

    it('highlights /automations when active', () => {
        mockPathname.mockReturnValue('/automations');
        const { container } = render(<ZoneA />);
        const automationsLink = container.querySelector('a[href="/automations"]');
        expect(automationsLink?.className).toContain('bg-cortex-primary');
    });

    it('does not show System tab when advancedMode is off', () => {
        mockAdvancedMode.mockReturnValue(false);
        const { container } = render(<ZoneA />);
        const systemLink = container.querySelector('a[href="/system"]');
        expect(systemLink).toBeNull();
    });

    it('shows System tab when advancedMode is on', () => {
        mockAdvancedMode.mockReturnValue(true);
        render(<ZoneA />);
        expect(screen.getByText('System')).toBeDefined();
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
