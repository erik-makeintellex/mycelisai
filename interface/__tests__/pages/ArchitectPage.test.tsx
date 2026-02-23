import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import ArchitectRedirect from '@/app/(app)/architect/page';

describe('Architect Page (redirect)', () => {
    it('redirects to /automations?tab=wiring', () => {
        try { ArchitectRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/automations?tab=wiring');
    });
});
