import { defineConfig, devices } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { STORAGE_STATE } from './e2e/global-setup';

const configDir = typeof __dirname === 'string' ? __dirname : process.cwd();
loadRepoEnv(path.resolve(configDir, '..', '.env'));
const interfaceHost = process.env.INTERFACE_HOST ?? process.env.MYCELIS_INTERFACE_HOST ?? '127.0.0.1';
const interfacePort = process.env.PLAYWRIGHT_PORT ?? process.env.INTERFACE_PORT ?? '3100';
const baseUrlHost = interfaceHost.includes(':') && !interfaceHost.startsWith('[') ? `[${interfaceHost}]` : interfaceHost;
const baseURL = `http://${baseUrlHost}:${interfacePort}`;
const shouldManageWebServer = !process.env.PLAYWRIGHT_SKIP_WEBSERVER;
const defaultWebServerCommand = `node ./scripts/playwright-webserver.mjs`;
const webServerCommand = process.env.PLAYWRIGHT_UI_SERVER_COMMAND ?? defaultWebServerCommand;
process.env.MYCELIS_WEB_SESSION_SECRET ||= 'playwright-web-session-secret';
process.env.MYCELIS_LOCAL_ADMIN_USERNAME ||= 'admin';
process.env.MYCELIS_LOCAL_ADMIN_PASSWORD ||= process.env.MYCELIS_API_KEY || 'playwright-admin';

function loadRepoEnv(envPath: string) {
    if (!fs.existsSync(envPath)) return;
    const lines = fs.readFileSync(envPath, 'utf8').split(/\r?\n/);
    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || trimmed.startsWith('#')) continue;
        const equals = trimmed.indexOf('=');
        if (equals <= 0) continue;
        const key = trimmed.slice(0, equals).trim();
        if (!key || process.env[key] !== undefined) continue;
        let value = trimmed.slice(equals + 1).trim();
        if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
            value = value.slice(1, -1);
        }
        process.env[key] = value;
    }
}

/**
 * Playwright E2E configuration for Mycelis Interface.
 *
 * Direct Playwright runs own a built Interface server lifecycle by default,
 * while `uv run inv interface.e2e` may pre-start a known-good local server and
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
    globalSetup: './e2e/global-setup.ts',

    use: {
        baseURL,
        storageState: STORAGE_STATE,
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
