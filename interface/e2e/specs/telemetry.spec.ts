import { test, expect } from '@playwright/test';

test.skip(true, 'Raw telemetry endpoint checks are no longer part of the MVP route audit suite.');

test.describe('Phase 5.2 - System Telemetry Contract', () => {

    test('GET /api/v1/telemetry/compute returns valid metrics', async ({ request }) => {
        const response = await request.get('http://localhost:3000/api/v1/telemetry/compute');
        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body).toHaveProperty('goroutines');
        expect(body).toHaveProperty('heap_alloc_mb');
        expect(body).toHaveProperty('sys_mem_mb');
        expect(body).toHaveProperty('llm_tokens_sec');
        expect(body).toHaveProperty('timestamp');

        expect(typeof body.goroutines).toBe('number');
        expect(typeof body.heap_alloc_mb).toBe('number');
        expect(typeof body.sys_mem_mb).toBe('number');
        expect(typeof body.llm_tokens_sec).toBe('number');
        expect(typeof body.timestamp).toBe('string');
    });

});
