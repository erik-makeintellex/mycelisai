import { test, expect } from '@playwright/test';

test.describe('Phase 5.4 - Hippocampus (Memory) Contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');

    test('GET /api/v1/memory/search returns semantic SitReps', async ({ request }) => {
        const query = 'trust economy architecture';

        const response = await request.get(`/api/v1/memory/search`, {
            params: {
                q: query,
                limit: 5
            }
        });

        expect(response.status()).toBe(200);

        const body = await response.json();
        const results = Array.isArray(body) ? body : body.results ?? [];
        expect(Array.isArray(results)).toBeTruthy();

        if (results.length > 0) {
            const firstResult = results[0];
            expect(firstResult).toHaveProperty('content');
            expect(firstResult).toHaveProperty('score');
            expect(firstResult).toHaveProperty('id');
            expect(firstResult).toHaveProperty('created_at');
            expect(typeof firstResult.content).toBe('string');
            expect(typeof firstResult.score).toBe('number');
        }
    });

});
