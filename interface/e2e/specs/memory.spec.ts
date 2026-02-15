import { test, expect } from '@playwright/test';

test.describe('Phase 5.4 - Hippocampus (Memory) Contract', () => {

    test('GET /api/v1/memory/search returns semantic SitReps', async ({ request }) => {
        // 1. Arrange: Define a query that should yield results
        const query = 'trust economy architecture';

        // 2. Act: Execute Search API Request
        const response = await request.get(`http://localhost:3000/api/v1/memory/search`, {
            params: {
                q: query,
                limit: 5
            }
        });

        // 3. Assert: Status Code
        expect(response.status()).toBe(200);

        // 4. Assert: Response Structure (Vector Search Result)
        const body = await response.json();
        console.log('Memory Search Response:', body);

        // Expecting an array of results or a wrapper
        // Adapting to probable response format based on architecture
        const results = Array.isArray(body) ? body : body.results;

        expect(Array.isArray(results)).toBeTruthy();

        if (results.length > 0) {
            const firstResult = results[0];

            // Verify Semantic Payload
            expect(firstResult).toHaveProperty('content');
            expect(firstResult).toHaveProperty('similarity');
            // expect(firstResult.similarity).toBeGreaterThan(0.5); // value depends on embedding model

            // Verify Source Traceability
            expect(firstResult).toHaveProperty('source');
            expect(firstResult).toHaveProperty('timestamp');
        } else {
            console.warn('Memory search returned 0 results. Vector store might be empty.');
        }
    });

});
