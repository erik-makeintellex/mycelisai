import { chromium, type FullConfig } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';

// Mirrors WEB_SESSION_COOKIE in lib/webAuth.ts (kept literal so global setup
// does not import server-side auth code).
const WEB_SESSION_COOKIE = 'mycelis_web_session';

export const STORAGE_STATE =path.join(process.cwd(), '.playwright', '.auth', 'admin.json');

export default async function globalSetup(config: FullConfig) {
    if (process.env.PLAYWRIGHT_SKIP_AUTH_SETUP === '1') return;
    const baseURL = config.projects[0]?.use.baseURL;
    if (!baseURL) return;
    await fs.mkdir(path.dirname(STORAGE_STATE), { recursive: true });
    const browser = await chromium.launch();
    const context = await browser.newContext({ baseURL: String(baseURL) });
    const page = await context.newPage();
    await page.goto('/login?next=/dashboard');
    await page.getByLabel(/Local admin username/i).fill(process.env.MYCELIS_LOCAL_ADMIN_USERNAME || 'admin');
    const password = process.env.MYCELIS_LOCAL_ADMIN_PASSWORD;
    if (!password) {
        throw new Error('Set MYCELIS_LOCAL_ADMIN_PASSWORD for Playwright; MYCELIS_API_KEY is not a login password.');
    }
    await page.getByLabel(/Local admin password/i).fill(password);
    // The login URL itself (`/login?next=/dashboard`) ends in "/dashboard", and
    // waitForURL resolves at once when the current URL already matches. So wait
    // on the pathname, then prove the session cookie exists before saving state:
    // a failed or slow sign-in must fail here, not as 100+ tests on the login page.
    await page.getByRole('button', { name: /Sign in as local admin/i }).click();
    await page.waitForURL(
        (url) => url.pathname === '/dashboard' || (url.pathname === '/login' && url.searchParams.has('error')),
        { timeout: 30_000 },
    );
    const landed = new URL(page.url());
    if (landed.pathname !== '/dashboard') {
        throw new Error(`Local admin sign-in failed (error=${landed.searchParams.get('error') ?? 'unknown'}); check MYCELIS_LOCAL_ADMIN_USERNAME and MYCELIS_LOCAL_ADMIN_PASSWORD.`);
    }
    const cookies = await context.cookies();
    if (!cookies.some((cookie) => cookie.name === WEB_SESSION_COOKIE)) {
        throw new Error(`Local admin sign-in reached /dashboard but no ${WEB_SESSION_COOKIE} cookie was set; e2e state would be signed out.`);
    }
    await context.storageState({ path: STORAGE_STATE });
    await browser.close();
}
