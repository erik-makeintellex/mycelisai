import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('@/components/teams/TeamsPage', () => ({
    __esModule: true,
    default: () => <div>Teams workspace route</div>,
}));

import TeamsRoutePage from '@/app/(app)/teams/page';

describe('Teams route page', () => {
    it('renders the direct teams workspace', () => {
        render(<TeamsRoutePage />);
        expect(screen.getByText('Teams workspace route')).toBeDefined();
    });
});
