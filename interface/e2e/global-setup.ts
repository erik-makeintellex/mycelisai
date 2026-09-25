import { chromium, type FullConfig } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';

export const STORAGE_STATE = path.join(process.cwd(), '.playwright', '.auth', 'admin.json');

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
    await Promise.all([
        page.waitForURL(/\/dashboard(?:\?|$)/),
        page.getByRole('button', { name: /Sign in as local admin/i }).click(),
    ]);
    await context.storageState({ path: STORAGE_STATE });
    await browser.close();
}
