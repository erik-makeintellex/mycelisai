import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import MarketplaceRedirect from '@/app/(app)/marketplace/page';

describe('Marketplace Page (redirect)', () => {
    it('redirects to /resources?tab=catalogue', () => {
        try { MarketplaceRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/resources?tab=catalogue');
    });
});
