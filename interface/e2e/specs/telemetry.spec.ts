import { test, expect } from '@playwright/test';

test.describe('Phase 5.2 - System Telemetry Contract', () => {

    test('GET /api/v1/telemetry/compute returns valid metrics', async ({ request }) => {
        // 1. Execute API Request
        const response = await request.get('http://localhost:3000/api/v1/telemetry/compute');

        // 2. Assert Status Code
        expect(response.status()).toBe(200);

        // 3. Assert Response Body Structure
        const body = await response.json();
        console.log('Telemetry Response:', body);

        // Validation of the Cortex Telemetry Standard (CTS) for Compute
        expect(body).toHaveProperty('cpu_usage');
        expect(body).toHaveProperty('memory_usage');
        expect(body).toHaveProperty('active_nodes');
        expect(body).toHaveProperty('trust_index');

        // Type checks
        expect(typeof body.cpu_usage).toBe('number');
        expect(typeof body.memory_usage).toBe('number');
        expect(typeof body.active_nodes).toBe('number');
        expect(typeof body.trust_index).toBe('number');
    });

});
