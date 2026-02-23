import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import CatalogueRedirect from '@/app/(app)/catalogue/page';

describe('Catalogue Page (redirect)', () => {
    it('redirects to /resources?tab=catalogue', () => {
        try { CatalogueRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/resources?tab=catalogue');
    });
});
