import { expect, type Page } from '@playwright/test';

export type APIEnvelope<T> = {
    ok?: boolean;
    data?: T;
    error?: string;
};

export type OrganizationEnvelope = {
    data?: {
        id?: string;
        name?: string;
    };
};

export type ChatEnvelope = {
    ok?: boolean;
    data?: {
        mode?: string;
        payload?: {
            text?: string;
            ask_class?: string;
        };
    };
};

export type TeamLeadWorkflowGroupDraft = {
    name: string;
    goal_statement: string;
    work_mode: 'read_only' | 'propose_only' | 'execute_with_approval' | 'execute_bounded';
    coordinator_profile: string;
    allowed_capabilities?: string[];
    recommended_member_limit?: number;
    expiry_hours?: number;
    summary: string;
};

export type TeamLeadExecutionContract = {
    execution_mode?: string;
    team_name?: string;
    summary?: string;
    coordination_model?: string;
    recommended_team_count?: number;
    recommended_team_member_limit?: number;
    target_outputs?: string[];
    workflow_group?: TeamLeadWorkflowGroupDraft;
};

export type TeamLeadGuidanceResponse = {
    headline?: string;
    summary?: string;
    execution_contract?: TeamLeadExecutionContract;
};

export type GroupRecord = {
    group_id: string;
    name: string;
    goal_statement: string;
    status: string;
};

export type ArtifactRecord = {
    id?: string;
    title?: string;
};

export type SeedResponse = {
    mission_id?: string;
};

export type MissionDetail = {
    id?: string;
    teams?: Array<{
        id?: string;
        name?: string;
    }>;
};

export async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return {
            raw,
            body: JSON.parse(raw) as T,
        };
    } catch {
        return {
            raw,
            body: null as T | null,
        };
    }
}

export async function expectExecutionContractOutputs(page: Page, outputs: string[]) {
    const targetOutputsSection = page.getByText('Target outputs', { exact: true }).locator('..');
    for (const output of outputs) {
        await expect(targetOutputsSection.getByText(output, { exact: true }).first()).toBeVisible();
    }
}

export async function gotoWithColdStartRetry(page: Page, path: string) {
    try {
        await page.goto(path, { waitUntil: 'domcontentloaded' });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!message.includes('net::ERR_ABORTED') && !message.includes('frame was detached')) {
            throw error;
        }
        await page.goto(path, { waitUntil: 'domcontentloaded' });
    }
}

export async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: {
            name,
            purpose: 'Live workflow variant verification',
            start_mode: 'empty',
        },
    });
    const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.data?.id).toBeTruthy();
    return {
        id: parsed.body!.data!.id!,
        name: parsed.body!.data!.name || name,
    };
}

export async function openWorkspace(page: Page, organizationId: string) {
    await gotoWithColdStartRetry(page, `/organizations/${organizationId}`);
    await page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i).waitFor({
        timeout: 30_000,
    });
}

export async function submitWorkspaceChat(page: Page, content: string) {
    const input = page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i);
    await input.fill(content);
    const responsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return response.request().method() === 'POST' && url.pathname === '/api/v1/chat';
    }, { timeout: 120_000 });
    await input.press('Enter');
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

export async function openTeamCreation(page: Page, organizationId: string) {
    await gotoWithColdStartRetry(page, `/teams/create?organization_id=${encodeURIComponent(organizationId)}`);
    await expect(page.getByRole('heading', { name: 'Create a team through Soma' })).toBeVisible({ timeout: 30_000 });
}

export async function submitTeamDesign(page: Page, organizationId: string, prompt: string) {
    const responsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return response.request().method() === 'POST'
            && url.pathname === `/api/v1/organizations/${organizationId}/workspace/actions`;
    }, { timeout: 60_000 });
    await page.getByLabel('Tell Soma what team or delivery lane you want to create').fill(prompt);
    await page.getByRole('button', { name: 'Start team design' }).click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<APIEnvelope<TeamLeadGuidanceResponse>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.ok, parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(true);
    expect(parsed.body?.data?.execution_contract).toBeTruthy();
    return parsed.body!.data!;
}
