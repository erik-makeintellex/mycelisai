import { expect, type Page, type Route } from "@playwright/test";

export const organizationId = "org-workflow-variants";
export const chatPlaceholder = /Tell .* what you want to plan, review, create, or execute/i;

export type GroupRecord = {
    group_id: string;
    name: string;
    goal_statement: string;
    status: string;
    work_mode: string;
    member_user_ids: string[];
    team_ids: string[];
    coordinator_profile: string;
    approval_policy_ref: string;
    expiry: string;
    created_by: string;
    created_at: string;
};

export type ArtifactRecord = {
    id: string;
    title: string;
    artifact_type: string;
    content_type: string;
    content?: string;
    file_path?: string;
    team_id: string;
    agent_id: string;
    metadata: Record<string, unknown>;
    status: string;
    created_at: string;
};

export async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

export async function gotoWithColdStartRetry(page: Page, path: string) {
    const target = resolveAppTarget(page, path);
    for (let attempt = 0; attempt < 3; attempt += 1) {
        try {
            await page.goto(target, { waitUntil: "domcontentloaded" });
            await page.waitForLoadState("domcontentloaded");
            return;
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            const canRetry =
                message.includes("net::ERR_ABORTED") ||
                message.includes("frame was detached") ||
                message.includes("net::ERR_NETWORK_CHANGED") ||
                message.includes("interrupted by another navigation") ||
                message.includes("chrome-error://chromewebdata/");
            if (!canRetry || attempt === 2) {
                throw error;
            }
            await page.waitForTimeout(500);
        }
    }
}

export async function installWorkflowOutputShell(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem("mycelis-last-organization-id", "org-workflow-variants");
        window.localStorage.setItem("mycelis-last-organization-name", "Northstar Labs");
    });

    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
                id: "operator-1",
                name: "Operator",
                email: "operator@example.test",
            },
        });
    });

    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: [
                { name: "nats", status: "online" },
                { name: "postgres", status: "online" },
                { name: "reactive", status: "online" },
            ],
        });
    });

    await page.route(`**/api/v1/organizations/${organizationId}/home`, async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
                id: organizationId,
                name: "Northstar Labs",
                purpose: "Run a product-ready Soma-first workflow with durable retained outputs.",
                start_mode: "template",
                team_lead_label: "Team Lead",
                advisor_count: 1,
                department_count: 2,
                specialist_count: 4,
                ai_engine_settings_summary: "Balanced",
                response_contract_summary: "Clear and structured",
                memory_personality_summary: "Stable continuity",
                status: "ready",
            },
        });
    });
}

export async function openOrganization(page: Page) {
    await gotoWithColdStartRetry(page, `/organizations/${organizationId}`);
    await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
    await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
}

export async function sendWorkspaceMessage(page: Page, content: string) {
    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(content);
    await input.press("Enter");
}

function resolveAppTarget(page: Page, path: string) {
    const currentUrl = page.url();
    if (!currentUrl.startsWith("http://") && !currentUrl.startsWith("https://")) {
        return path;
    }
    return new URL(path, currentUrl).toString();
}
