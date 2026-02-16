import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock next/navigation with a configurable pathname
const mockPathname = vi.fn(() => '/');
vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: vi.fn(),
        replace: vi.fn(),
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    usePathname: () => mockPathname(),
}));

import { ZoneA } from '@/components/shell/ZoneA_Rail';

const NAV_ENTRIES = [
    { href: '/', label: 'Mission Control' },
    { href: '/dashboard', label: 'Dashboard' },
    { href: '/architect', label: 'Swarm Architect' },
    { href: '/matrix', label: 'Cognitive Matrix' },
    { href: '/wiring', label: 'Neural Wiring' },
    { href: '/teams', label: 'Team Management' },
    { href: '/catalogue', label: 'Agent Catalogue' },
    { href: '/marketplace', label: 'Skills Market' },
    { href: '/memory', label: 'Memory' },
    { href: '/telemetry', label: 'System Status' },
    { href: '/approvals', label: 'Governance' },
    { href: '/settings', label: 'Settings' },
];

describe('ZoneA_Rail', () => {
    beforeEach(() => {
        mockPathname.mockReturnValue('/');
    });

    it('renders the Mycelis brand name', () => {
        render(<ZoneA />);
        expect(screen.getByText('Mycelis')).toBeDefined();
    });

    it('renders all navigation links', () => {
        render(<ZoneA />);
        for (const entry of NAV_ENTRIES) {
            expect(screen.getByText(entry.label)).toBeDefined();
        }
    });

    it('highlights active route with cortex-primary', () => {
        mockPathname.mockReturnValue('/');
        const { container } = render(<ZoneA />);
        const homeLink = container.querySelector('a[href="/"]');
        expect(homeLink?.className).toContain('bg-cortex-primary');
        expect(homeLink?.className).toContain('text-white');
    });

    it('uses muted text for inactive routes', () => {
        mockPathname.mockReturnValue('/');
        const { container } = render(<ZoneA />);
        const teamsLink = container.querySelector('a[href="/teams"]');
        expect(teamsLink?.className).toContain('text-cortex-text-muted');
    });

    it('highlights /teams when active', () => {
        mockPathname.mockReturnValue('/teams');
        const { container } = render(<ZoneA />);
        const teamsLink = container.querySelector('a[href="/teams"]');
        expect(teamsLink?.className).toContain('bg-cortex-primary');
    });

    it('does not use bg-white', () => {
        const { container } = render(<ZoneA />);
        expect(container.innerHTML).not.toContain('bg-white');
    });
});
