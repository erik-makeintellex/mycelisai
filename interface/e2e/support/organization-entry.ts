import { expect, type Page, type TestInfo } from "@playwright/test";
import {
    buildDefaultChatResponse,
    buildDefaultWorkspaceActionResponse,
    createdTemplateOrganization,
    fulfillJson,
    mockBaseStatusRoutes,
    type MockOrganizationEntryOptions,
    starterTemplate,
} from "./organization-entry-fixtures";

export * from "./organization-entry-fixtures";

export async function saveScreenshot(page: Page, testInfo: TestInfo, name: string) {
    await page.screenshot({
        path: testInfo.outputPath(name),
        fullPage: true,
    });
}

export async function clickStartMode(page: Page, label: RegExp | string) {
    const button = page.getByRole("button", { name: label });
    await expect(button).toBeVisible();
    await button.scrollIntoViewIfNeeded();
    await button.evaluate((element) => {
        (element as HTMLButtonElement).click();
    });
}

export async function openOrganizationSetup(page: Page) {
    const startEmptyButton = page.getByRole("button", { name: "Start Empty", exact: true });
    if (await startEmptyButton.isVisible().catch(() => false)) {
        return;
    }

    const setupTrigger = page.getByRole("button", { name: "Create or open AI Organizations" });
    await expect(setupTrigger).toBeVisible();
    await setupTrigger.click();
    if (await startEmptyButton.isVisible().catch(() => false)) {
        return;
    }
    await setupTrigger.evaluate((element) => {
        (element as HTMLButtonElement).click();
    });
    if (await startEmptyButton.isVisible().catch(() => false)) {
        return;
    }
    await page.evaluate(() => {
        const details = document.getElementById("dashboard-organization-setup");
        if (!(details instanceof HTMLDetailsElement)) {
            return;
        }
        details.open = true;
        if (typeof details.scrollIntoView === "function") {
            details.scrollIntoView({ block: "start" });
        }
    });
    await expect(startEmptyButton).toBeVisible();
}

export async function openDashboard(page: Page) {
    const dashboardTarget = resolveAppTarget(page, "/dashboard");
    for (let attempt = 0; attempt < 3; attempt += 1) {
        try {
            await page.goto(dashboardTarget, { waitUntil: "domcontentloaded" });
            await page.waitForLoadState("domcontentloaded");
            return;
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            const canRetry =
                message.includes("ERR_ABORTED")
                || message.includes("frame was detached")
                || message.includes("ERR_NETWORK_CHANGED")
                || message.includes("interrupted by another navigation")
                || message.includes("chrome-error://chromewebdata/");
            if (!canRetry || attempt === 2) {
                throw error;
            }
            await page.waitForTimeout(500);
        }
    }
}

export async function openCreatedOrganization(page: Page, organizationId: string) {
    const organizationPath = `/organizations/${organizationId}`;
    const organizationUrl = new RegExp(`${organizationPath}$`);
    const organizationTarget = resolveAppTarget(page, organizationPath);

    try {
        await page.waitForURL(organizationUrl, {
            timeout: 5000,
            waitUntil: "domcontentloaded",
        });
    } catch {
        for (let attempt = 0; attempt < 2; attempt += 1) {
            try {
                await page.goto(organizationTarget, { waitUntil: "domcontentloaded" });
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                if (
                    !message.includes("ERR_ABORTED")
                    && !message.includes("frame was detached")
                    && !message.includes("ERR_NETWORK_CHANGED")
                    && !message.includes("interrupted by another navigation")
                    && !message.includes("chrome-error://chromewebdata/")
                ) {
                    throw error;
                }
                await page.waitForTimeout(500);
            }

            if (page.url().endsWith(organizationPath)) {
                break;
            }
        }
    }

    await expect(page).toHaveURL(organizationUrl, { timeout: 15000 });
    await page.waitForLoadState("domcontentloaded");
}

export function recentOrganizationLink(page: Page, organizationName: string) {
    return page.getByRole("link", { name: `Return to ${organizationName}` });
}

function resolveAppTarget(page: Page, path: string) {
    const currentUrl = page.url();
    if (!currentUrl.startsWith("http://") && !currentUrl.startsWith("https://")) {
        return path;
    }
    return new URL(path, currentUrl).toString();
}

export function recentOrganizationOpenButton(page: Page, organizationName: string) {
    return page.getByRole("button", {
        name: new RegExp(`${organizationName}.*Open AI Organization`, "i"),
    });
}

export async function openTeamDesignLane(page: Page) {
    await page.goto(resolveAppTarget(page, "/teams/create"));
    await page.waitForLoadState("domcontentloaded");
    await expect(page.getByRole("heading", { name: "Create a team through Soma" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Focused team design with Soma" })).toBeVisible();
}

export async function expectNoForbiddenCopy(page: Page) {
    const workspaceText = await page.locator("body").innerText();
    expect(workspaceText).not.toContain("V8 Entry Flow");
    expect(workspaceText).not.toMatch(/bounded slice/i);
    expect(workspaceText).not.toMatch(/implementation slice/i);
    expect(workspaceText).not.toMatch(/context shell/i);
    expect(workspaceText).not.toMatch(/loop profile/i);
    expect(workspaceText).not.toMatch(/scheduler/i);
    expect(workspaceText).not.toMatch(/Memory & Personality/i);
    expect(workspaceText).not.toMatch(/central council/i);
    expect(workspaceText).not.toMatch(/raw architecture controls/i);
    expect(workspaceText).not.toMatch(/vector/i);
    expect(workspaceText).not.toMatch(/pgvector/i);
    expect(workspaceText).not.toMatch(/memory promotion/i);
}

export async function mockOrganizationEntryApis(
    page: Page,
    options?: MockOrganizationEntryOptions,
) {
    const {
        templates = [starterTemplate],
        organizations = [],
        templateStatus = 200,
        templateError = "Starter templates are unavailable right now.",
        organizationsStatus = 200,
        organizationsError = "Recent AI Organizations are unavailable right now.",
        organizationsSummaryResponses = [],
        createHandler,
        actionHandler,
        chatHandler,
        homeResponsesById = {
            [createdTemplateOrganization.id]: createdTemplateOrganization,
        },
    } = options ?? {};

    const mutableHomeResponsesById = structuredClone(homeResponsesById);
    const mutableOrganizationsSummaryResponses = structuredClone(organizationsSummaryResponses);

    await mockBaseStatusRoutes(page);

    await page.route(/\/api\/v1\/templates(?:\?.*)?$/, async (route) => {
        const url = new URL(route.request().url());
        if (route.request().method() !== "GET" || url.searchParams.get("view") !== "organization-starters") {
            await route.fallback();
            return;
        }
        await fulfillJson(
            route,
            templateStatus,
            templateStatus >= 400 ? { ok: false, error: templateError } : { ok: true, data: templates },
        );
    });

    await page.route(/\/api\/v1\/organizations(?:\?.*)?$/, async (route) => {
        const url = new URL(route.request().url());
        if (route.request().method() !== "GET" || url.searchParams.get("view") !== "summary") {
            await route.fallback();
            return;
        }
        const summaryResponse = mutableOrganizationsSummaryResponses.shift();
        if (summaryResponse) {
            await fulfillJson(route, summaryResponse.status, summaryResponse.body);
            return;
        }
        await fulfillJson(
            route,
            organizationsStatus,
            organizationsStatus >= 400 ? { ok: false, error: organizationsError } : { ok: true, data: organizations },
        );
    });

    await page.route("**/api/v1/organizations", async (route) => {
        if (route.request().method() !== "POST") {
            await route.fallback();
            return;
        }
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const response = createHandler
            ? createHandler(requestBody)
            : { status: 201, body: { ok: true, data: createdTemplateOrganization } };
        await fulfillJson(route, response.status, response.body);
    });

    await page.route("**/api/v1/organizations/*/workspace/actions", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const response = actionHandler
            ? actionHandler(requestBody)
            : buildDefaultWorkspaceActionResponse(requestBody);
        await fulfillJson(route, response.status, response.body);
    });

    await page.route("**/api/v1/chat", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const response = chatHandler ? chatHandler(requestBody) : buildDefaultChatResponse();
        await fulfillJson(route, response.status, response.body);
    });

    await page.route("**/api/v1/organizations/*/home", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/home$/);
        const organizationId = match?.[1];
        const payload = organizationId ? mutableHomeResponsesById[organizationId] : undefined;
        await fulfillJson(
            route,
            payload ? 200 : 404,
            payload ? { ok: true, data: payload } : { ok: false, error: "AI Organization unavailable" },
        );
    });
}
