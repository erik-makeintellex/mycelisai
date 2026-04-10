import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('@/components/teams/TeamCreationPage', () => ({
    __esModule: true,
    default: () => <div>Guided team creation route</div>,
}));

import CreateTeamRoutePage from '@/app/(app)/teams/create/page';

describe('Create Team route page', () => {
    it('renders the guided team creation workspace', () => {
        render(<CreateTeamRoutePage />);
        expect(screen.getByText('Guided team creation route')).toBeDefined();
    });
});
