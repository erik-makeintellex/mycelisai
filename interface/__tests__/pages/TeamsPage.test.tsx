import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import TeamsRedirect from '@/app/(app)/teams/page';

describe('Teams Page (redirect)', () => {
    it('redirects to /automations?tab=teams', () => {
        try { TeamsRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/automations?tab=teams');
    });
});
