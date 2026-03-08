import { defineConfig, devices } from '@playwright/test';

const interfaceHost = process.env.INTERFACE_HOST ?? '127.0.0.1';
const interfacePort = process.env.INTERFACE_PORT ?? '3000';
const baseURL = `http://${interfaceHost}:${interfacePort}`;
const webServerCommand =
    process.env.PLAYWRIGHT_UI_SERVER_COMMAND ??
    `npm run dev -- --hostname ${interfaceHost} --port ${interfacePort}`;

/**
 * Playwright E2E configuration for Mycelis Interface.
 *
 * Playwright owns the Next.js dev server lifecycle so `uv run inv interface.e2e`
 * remains reproducible and does not depend on a manually started UI process.
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
        baseURL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        video: 'retain-on-failure',
    },

    webServer: {
        command: webServerCommand,
        cwd: __dirname,
        url: baseURL,
        reuseExistingServer: !process.env.CI,
        stdout: 'ignore',
        stderr: 'pipe',
        timeout: 120_000,
    },

    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
            testIgnore: /.*mobile\.spec\.ts/,
        },
        {
            name: 'firefox',
            use: { ...devices['Desktop Firefox'] },
            testIgnore: /.*mobile\.spec\.ts/,
        },
        {
            name: 'webkit',
            use: { ...devices['Desktop Safari'] },
            testIgnore: /.*mobile\.spec\.ts/,
        },
        {
            name: 'mobile-chromium',
            use: { ...devices['Pixel 5'] },
            testMatch: /.*mobile\.spec\.ts/,
        },
    ],
});
