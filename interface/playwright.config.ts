import { defineConfig, devices } from '@playwright/test';

const configDir = typeof __dirname === 'string' ? __dirname : process.cwd();
const interfaceHost = process.env.INTERFACE_HOST ?? '127.0.0.1';
const interfacePort = process.env.INTERFACE_PORT ?? '3000';
const baseURL = `http://${interfaceHost}:${interfacePort}`;
const shouldManageWebServer = !process.env.PLAYWRIGHT_SKIP_WEBSERVER;
const defaultDevCommand = `node ./node_modules/next/dist/bin/next dev --webpack --hostname ${interfaceHost} --port ${interfacePort}`;
const webServerCommand = process.env.PLAYWRIGHT_UI_SERVER_COMMAND ?? defaultDevCommand;

/**
 * Playwright E2E configuration for Mycelis Interface.
 *
 * Direct Playwright runs can own the Next.js dev server lifecycle, while
 * `uv run inv interface.e2e` may pre-start a known-good local server and
 * opt out via PLAYWRIGHT_SKIP_WEBSERVER when Windows shell startup is flaky.
 */
export default defineConfig({
    testDir: './e2e/specs',
    timeout: 45_000,
    expect: { timeout: 10_000 },
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : 4,
    reporter: process.env.CI ? 'github' : 'list',

    use: {
        baseURL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        video: 'retain-on-failure',
    },

    webServer: shouldManageWebServer
        ? {
              command: webServerCommand,
              cwd: configDir,
              url: baseURL,
              reuseExistingServer: !process.env.CI,
              stdout: 'ignore',
              stderr: 'pipe',
              timeout: 120_000,
          }
        : undefined,

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
