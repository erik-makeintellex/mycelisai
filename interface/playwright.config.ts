import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E configuration for Mycelis Interface.
 *
 * Expects the dev server to already be running (`uvx inv interface.dev`).
 * Run tests: `uvx inv interface.e2e`
 */
export default defineConfig({
    testDir: './e2e/specs',
    timeout: 30_000,
    expect: { timeout: 10_000 },
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: process.env.CI ? 'github' : 'list',

    use: {
        baseURL: `http://${process.env.INTERFACE_HOST ?? 'localhost'}:${process.env.INTERFACE_PORT ?? '3000'}`,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        video: 'retain-on-failure',
    },

    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
});
