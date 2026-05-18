import { request, type FullConfig } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';

export const STORAGE_STATE = path.join(process.cwd(), 'test-results', '.auth', 'admin.json');

export default async function globalSetup(config: FullConfig) {
    if (process.env.PLAYWRIGHT_SKIP_AUTH_SETUP === '1') return;
    const baseURL = config.projects[0]?.use.baseURL;
    if (!baseURL) return;
    await fs.mkdir(path.dirname(STORAGE_STATE), { recursive: true });
    const context = await request.newContext({ baseURL: String(baseURL) });
    const response = await context.post('/auth/local?next=/dashboard', {
        form: {
            username: process.env.MYCELIS_LOCAL_ADMIN_USERNAME || 'admin',
            password: process.env.MYCELIS_LOCAL_ADMIN_PASSWORD || process.env.MYCELIS_API_KEY || 'playwright-admin',
        },
    });
    if (!response.ok()) throw new Error(`Playwright auth setup failed with ${response.status()}`);
    await context.storageState({ path: STORAGE_STATE });
    await context.dispose();
}
