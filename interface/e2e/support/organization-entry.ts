import { expect, type Page, type TestInfo } from "@playwright/test";

export const starterTemplate = {
    id: "engineering-starter",
    name: "Engineering Starter",
    description: "Guided AI Organization for engineering work",
    organization_type: "AI Organization starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_profile_id: "starter_defaults",
    ai_engine_settings_summary: "Starter defaults included",
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
};

export const createdTemplateOrganization = {
    id: "org-123",
    name: "Northstar Labs",
    purpose: "Ship a focused AI engineering organization for product delivery.",
    start_mode: "template",
    template_id: starterTemplate.id,
    template_name: starterTemplate.name,
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_profile_id: "starter_defaults",
    ai_engine_settings_summary: "Starter defaults included",
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
    status: "ready",
    departments: [
        {
            id: "platform",
            name: "Platform Department",
            specialist_count: 2,
            ai_engine_effective_profile_id: "starter_defaults",
            ai_engine_effective_summary: "Starter defaults included",
            inherits_organization_ai_engine: true,
            agent_type_profiles: [
                {
                    id: "planner",
                    name: "Planner",
                    helps_with: "Turns organization goals into practical next steps, delivery sequencing, and clear priorities.",
                    ai_engine_binding_profile_id: "high_reasoning",
                    ai_engine_effective_profile_id: "high_reasoning",
                    ai_engine_effective_summary: "High Reasoning",
                    inherits_department_ai_engine: false,
                    response_contract_binding_profile_id: "structured_analytical",
                    response_contract_effective_profile_id: "structured_analytical",
                    response_contract_effective_summary: "Structured & Analytical",
                    inherits_default_response_contract: false,
                },
                {
                    id: "delivery-specialist",
                    name: "Delivery Specialist",
                    helps_with: "Carries the work from plan into execution and keeps the main delivery lane moving.",
                    ai_engine_effective_profile_id: "starter_defaults",
                    ai_engine_effective_summary: "Starter defaults included",
                    inherits_department_ai_engine: true,
                    response_contract_effective_profile_id: "clear_balanced",
                    response_contract_effective_summary: "Clear & Balanced",
                    inherits_default_response_contract: true,
                },
            ],
        },
    ],
};

export const organizationChatPlaceholder = /Tell Soma what you want to plan, review, create, or execute/i;

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
    const guidedStartButton = page.getByRole("button", { name: "Open team design lane" });
    if (await guidedStartButton.isVisible().catch(() => false)) {
        await guidedStartButton.click();
        await expect(page.getByText("Choose a guided team-design action")).toBeVisible();
        return;
    }

    const workspaceToggle = page.getByRole("button", { name: "Create teams with Soma" });
    await expect(workspaceToggle).toBeVisible();
    await workspaceToggle.click();
    await expect(page.getByText("Choose a guided team-design action")).toBeVisible();
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
    expect(workspaceText).not.toMatch(/contract/i);
    expect(workspaceText).not.toMatch(/vector/i);
    expect(workspaceText).not.toMatch(/pgvector/i);
    expect(workspaceText).not.toMatch(/memory promotion/i);
}

type MockOrganizationEntryOptions = {
    templates?: unknown[];
    organizations?: unknown[];
    templateStatus?: number;
    templateError?: string;
    organizationsStatus?: number;
    organizationsError?: string;
    organizationsSummaryResponses?: Array<{ status: number; body: unknown }>;
    createHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
    actionHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
    chatHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
    homeResponsesById?: Record<string, unknown>;
};

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

    await page.route("**/api/v1/user/me", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
                ok: true,
                data: {
                    id: "operator-1",
                    name: "Operator",
                    email: "operator@example.test",
                },
            }),
        });
    });

    await page.route("**/api/v1/services/status", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
                ok: true,
                data: [
                    { name: "core", status: "ready" },
                    { name: "frontend", status: "ready" },
                ],
            }),
        });
    });

    await page.route(/\/api\/v1\/templates(?:\?.*)?$/, async (route) => {
        const url = new URL(route.request().url());
        if (route.request().method() !== "GET" || url.searchParams.get("view") !== "organization-starters") {
            await route.fallback();
            return;
        }

        await route.fulfill({
            status: templateStatus,
            contentType: "application/json",
            body: JSON.stringify(
                templateStatus >= 400
                    ? { ok: false, error: templateError }
                    : { ok: true, data: templates },
            ),
        });
    });

    await page.route(/\/api\/v1\/organizations(?:\?.*)?$/, async (route) => {
        const url = new URL(route.request().url());
        if (route.request().method() !== "GET" || url.searchParams.get("view") !== "summary") {
            await route.fallback();
            return;
        }

        const summaryResponse = mutableOrganizationsSummaryResponses.shift();
        if (summaryResponse) {
            await route.fulfill({
                status: summaryResponse.status,
                contentType: "application/json",
                body: JSON.stringify(summaryResponse.body),
            });
            return;
        }

        await route.fulfill({
            status: organizationsStatus,
            contentType: "application/json",
            body: JSON.stringify(
                organizationsStatus >= 400
                    ? { ok: false, error: organizationsError }
                    : { ok: true, data: organizations },
            ),
        });
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

        await route.fulfill({
            status: response.status,
            contentType: "application/json",
            body: JSON.stringify(response.body),
        });
    });

    await page.route("**/api/v1/organizations/*/workspace/actions", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const response = actionHandler
            ? actionHandler(requestBody)
            : {
                  status: 200,
                  body: {
                      ok: true,
                      data: {
                          action: requestBody.action ?? "plan_next_steps",
                          request_label: "Run a quick strategy check",
                          headline: "Team Lead plan for Northstar Labs",
                          summary: "Team Lead recommends a clear next move for Northstar Labs.",
                          priority_steps: [
                              "Align the first outcome with the AI Organization purpose.",
                              "Use the first Department as the routing layer for work.",
                          ],
                          suggested_follow_ups: [
                              "Review your organization setup",
                              "Choose the first priority",
                          ],
                      },
                  },
              };

        await route.fulfill({
            status: response.status,
            contentType: "application/json",
            body: JSON.stringify(response.body),
        });
    });

    await page.route("**/api/v1/chat", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const response = chatHandler
            ? chatHandler(requestBody)
            : {
                  status: 200,
                  body: {
                      ok: true,
                      data: {
                          meta: { source_node: "admin", timestamp: "2026-03-19T18:00:00Z" },
                          signal_type: "chat_response",
                          trust_score: 0.9,
                          template_id: "chat-to-answer",
                          mode: "answer",
                          payload: {
                              text: "Soma is ready to help inside this AI Organization.",
                              ask_class: "direct_answer",
                              consultations: [],
                              tools_used: [],
                              artifacts: [],
                          },
                      },
                  },
              };

        await route.fulfill({
            status: response.status,
            contentType: "application/json",
            body: JSON.stringify(response.body),
        });
    });

    await page.route("**/api/v1/organizations/*/home", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/home$/);
        const organizationId = match?.[1];
        const payload = organizationId ? mutableHomeResponsesById[organizationId] : undefined;

        await route.fulfill({
            status: payload ? 200 : 404,
            contentType: "application/json",
            body: JSON.stringify(
                payload
                    ? { ok: true, data: payload }
                    : { ok: false, error: "AI Organization unavailable" },
            ),
        });
    });
}
