import { expect, type Page } from '@playwright/test';

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return { raw, body: JSON.parse(raw) as T };
    } catch {
        return { raw, body: null as T | null };
    }
}

export async function storeLiveArtifactWithRetry<T>(
    page: Page,
    payload: Record<string, unknown>,
    context: string,
): Promise<T> {
    let last: { status: number; raw: string; body: T | null } | null = null;
    let lastError = '';
    for (let attempt = 1; attempt <= 3; attempt += 1) {
        let response;
        try {
            response = await page.request.post('/api/v1/artifacts', { data: payload });
        } catch (error) {
            lastError = error instanceof Error ? error.message : String(error);
            if (!isTransientRequestError(lastError) || attempt === 3) break;
            await new Promise((resolve) => setTimeout(resolve, 1500));
            continue;
        }
        const parsed = await parseJSONIfPossible<T>(response);
        if (response.ok()) {
            expect((parsed.body as { id?: string } | null)?.id, parsed.body ? JSON.stringify(parsed.body) : parsed.raw)
                .toBeTruthy();
            return parsed.body as T;
        }
        last = { status: response.status(), raw: parsed.raw, body: parsed.body };
        if (response.status() < 500 || attempt === 3) break;
        const stored = await findStoredArtifact<T>(page, payload);
        if (stored) return stored;
        await new Promise((resolve) => setTimeout(resolve, 1500));
    }
    const stored = await findStoredArtifact<T>(page, payload);
    if (stored) return stored;
    expect(
        false,
        `${context} artifact store failed after retries: ${lastError || `${last?.status} ${last?.body ? JSON.stringify(last.body) : last?.raw}`}`,
    ).toBeTruthy();
    throw new Error(`${context} artifact store failed`);
}

function isTransientRequestError(message: string) {
    return ['EADDRINUSE', 'ECONNRESET', 'ECONNREFUSED', 'ERR_CONNECTION_FAILED'].some((fragment) => message.includes(fragment));
}

async function findStoredArtifact<T>(page: Page, payload: Record<string, unknown>): Promise<T | null> {
    const teamID = typeof payload.team_id === 'string' ? payload.team_id : '';
    const title = typeof payload.title === 'string' ? payload.title : '';
    if (!teamID || !title) return null;
    const response = await page.request.get(`/api/v1/artifacts?team_id=${encodeURIComponent(teamID)}&limit=50`);
    if (!response.ok()) return null;
    const parsed = await parseJSONIfPossible<Array<{ id?: string; title?: string }>>(response);
    const match = Array.isArray(parsed.body)
        ? parsed.body.find((artifact) => artifact.title === title && artifact.id)
        : null;
    return match ? match as T : null;
}
