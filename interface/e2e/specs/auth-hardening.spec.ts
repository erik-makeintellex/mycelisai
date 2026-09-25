import { createHmac } from 'node:crypto';
import { expect, request as playwrightRequest, test, type TestInfo } from '@playwright/test';
import { liveAPIHeaders, liveAPIURL } from '../support/live-api-auth';

// Slice A1 access-hardening proof. Browser-facing cases run against any
// managed or external Interface; Core-direct cases need the live stack.

const COOKIE = 'mycelis_web_session';

function base64url(value: string | Buffer): string {
    return Buffer.from(value).toString('base64url');
}

function signedSessionCookie(role: 'admin' | 'standard', secret: string): string {
    const now = Math.floor(Date.now() / 1000);
    const payload = base64url(JSON.stringify({
        sub: `a1-proof-${role}`,
        email: `a1-proof-${role}@local.mycelis`,
        name: `A1 proof ${role}`,
        role,
        provider: 'local',
        iat: now,
        exp: now + 600,
    }));
    const signature = createHmac('sha256', secret).update(payload).digest('base64url');
    return `${payload}.${signature}`;
}

function baseURL(testInfo: TestInfo): string {
    return String(testInfo.project.use.baseURL);
}

async function anonymousPage(browser: import('@playwright/test').Browser, testInfo: TestInfo) {
    const context = await browser.newContext({ baseURL: baseURL(testInfo), storageState: { cookies: [], origins: [] } });
    return { context, page: await context.newPage() };
}

test.describe('A1 access hardening', () => {
    test('anonymous dashboard redirects to login', async ({ browser }, testInfo) => {
        const { context, page } = await anonymousPage(browser, testInfo);
        await page.goto('/dashboard');
        await expect(page).toHaveURL(/\/login\?next=/);
        await context.close();
    });

    test('forged and API-key-signed session cookies are rejected', async ({ browser }, testInfo) => {
        const { context, page } = await anonymousPage(browser, testInfo);
        const url = new URL(baseURL(testInfo));
        const valid = signedSessionCookie('admin', 'not-the-deployment-secret-0123456789abcdef');
        const forgeries = [valid, `${valid.split('.')[0]}.`, `.${valid.split('.')[1]}`];
        const apiKey = process.env.MYCELIS_API_KEY;
        if (apiKey) forgeries.push(signedSessionCookie('admin', apiKey));
        for (const value of forgeries) {
            await context.clearCookies();
            await context.addCookies([{ name: COOKIE, value, domain: url.hostname, path: '/' }]);
            await page.goto('/dashboard');
            await expect(page).toHaveURL(/\/login\?next=/);
        }
        await context.close();
    });

    test('cross-site local login POST is rejected without a session cookie', async ({}, testInfo) => {
        const api = await playwrightRequest.newContext({ baseURL: baseURL(testInfo) });
        const response = await api.post('/auth/local?next=%2Fdashboard', {
            form: { username: process.env.MYCELIS_LOCAL_ADMIN_USERNAME || 'admin', password: process.env.MYCELIS_LOCAL_ADMIN_PASSWORD || '' },
            headers: { Origin: 'https://evil.example' },
            maxRedirects: 0,
        });
        expect(response.status()).toBe(403);
        expect(response.headers()['set-cookie'] ?? '').not.toContain(`${COOKIE}=`);
        expect(response.headers()['location'] ?? '').not.toContain('evil.example');
        await api.dispose();
    });

    test('local login and logout round-trip', async ({ browser }, testInfo) => {
        test.skip(!process.env.MYCELIS_LOCAL_ADMIN_PASSWORD, 'requires MYCELIS_LOCAL_ADMIN_PASSWORD');
        const { context, page } = await anonymousPage(browser, testInfo);
        await page.goto('/login?next=/dashboard');
        await page.getByLabel(/Local admin username/i).fill(process.env.MYCELIS_LOCAL_ADMIN_USERNAME || 'admin');
        await page.getByLabel(/Local admin password/i).fill(process.env.MYCELIS_LOCAL_ADMIN_PASSWORD ?? '');
        await Promise.all([
            page.waitForURL(/\/dashboard(?:\?|$)/),
            page.getByRole('button', { name: /Sign in as local admin/i }).click(),
        ]);
        const logout = await page.request.post('/auth/logout', {
            headers: { Origin: new URL(baseURL(testInfo)).origin },
            maxRedirects: 0,
        });
        expect(logout.status()).toBe(303);
        expect(new URL(logout.headers()['location'] ?? '', baseURL(testInfo)).pathname).toBe('/login');
        await context.close();
    });
});

test.describe('A1 access hardening (live Core)', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires the live local Core/Interface stack');

    test('Core rejects ?token= and accepts the Authorization header', async () => {
        const headers = liveAPIHeaders();
        test.skip(!headers, 'requires MYCELIS_API_KEY');
        const apiKey = headers!.Authorization.replace(/^Bearer /, '');
        const api = await playwrightRequest.newContext();
        const viaQuery = await api.get(liveAPIURL(`/api/v1/user/me?token=${encodeURIComponent(apiKey)}`));
        expect(viaQuery.status()).toBe(401);
        const viaHeader = await api.get(liveAPIURL('/api/v1/user/me'), { headers });
        expect(viaHeader.status()).toBe(200);
        await api.dispose();
    });

    test('audit log is 403 for a standard session, 200 for admin, 401 anonymous', async ({}, testInfo) => {
        const secret = process.env.MYCELIS_WEB_SESSION_SECRET;
        test.skip(!secret, 'requires MYCELIS_WEB_SESSION_SECRET shared with the running Interface');
        const url = baseURL(testInfo);
        const asRole = async (cookie?: string) => {
            const api = await playwrightRequest.newContext({
                baseURL: url,
                extraHTTPHeaders: { accept: 'application/json', ...(cookie ? { cookie: `${COOKIE}=${cookie}` } : {}) },
            });
            const response = await api.get('/api/v1/audit?limit=1');
            const status = response.status();
            await api.dispose();
            return status;
        };
        expect(await asRole(signedSessionCookie('standard', secret!))).toBe(403);
        expect(await asRole(signedSessionCookie('admin', secret!))).toBe(200);
        expect(await asRole()).toBe(401);
    });
});
