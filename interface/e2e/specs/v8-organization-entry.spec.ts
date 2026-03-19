import { test, expect, type Page, type TestInfo } from "@playwright/test";

const starterTemplate = {
    id: "engineering-starter",
    name: "Engineering Starter",
    description: "Guided AI Organization for engineering work",
    organization_type: "AI Organization starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 2,
    specialist_count: 4,
    ai_engine_settings_summary: "Starter defaults included",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
};

const createdTemplateOrganization = {
    id: "org-123",
    name: "Northstar Labs",
    purpose: "Ship a focused AI engineering organization for product delivery.",
    start_mode: "template",
    template_id: starterTemplate.id,
    template_name: starterTemplate.name,
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 2,
    specialist_count: 4,
    ai_engine_settings_summary: "Starter defaults included",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
    status: "ready",
};

const createdEmptyOrganization = {
    id: "org-456",
    name: "Skylight Works",
    purpose: "Create an AI operations group from a clean starting point.",
    start_mode: "empty",
    team_lead_label: "Team Lead",
    advisor_count: 0,
    department_count: 0,
    specialist_count: 0,
    ai_engine_settings_summary: "Set up later in Advanced mode",
    memory_personality_summary: "Set up later in Advanced mode",
    status: "ready",
};

async function saveScreenshot(page: Page, testInfo: TestInfo, name: string) {
    await page.screenshot({
        path: testInfo.outputPath(name),
        fullPage: true,
    });
}

async function expectNoForbiddenCopy(page: Page) {
    await expect(page.getByText("V8 Entry Flow")).toHaveCount(0);
    await expect(page.getByText(/bounded slice/i)).toHaveCount(0);
    await expect(page.getByText(/implementation slice/i)).toHaveCount(0);
    await expect(page.getByText(/context shell/i)).toHaveCount(0);
    await expect(page.getByText(/raw architecture controls/i)).toHaveCount(0);
    await expect(page.getByText(/contract/i)).toHaveCount(0);
}

async function mockOrganizationEntryApis(
    page: Page,
    options?: {
        templates?: unknown[];
        organizations?: unknown[];
        templateStatus?: number;
        templateError?: string;
        organizationsStatus?: number;
        organizationsError?: string;
        createHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        homeResponsesById?: Record<string, unknown>;
    },
) {
    const {
        templates = [starterTemplate],
        organizations = [],
        templateStatus = 200,
        templateError = "Starter templates are unavailable right now.",
        organizationsStatus = 200,
        organizationsError = "Recent AI Organizations are unavailable right now.",
        createHandler,
        homeResponsesById = {
            [createdTemplateOrganization.id]: createdTemplateOrganization,
            [createdEmptyOrganization.id]: createdEmptyOrganization,
        },
    } = options ?? {};

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

    await page.route("**/api/v1/templates?view=organization-starters*", async (route) => {
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

    await page.route("**/api/v1/organizations?view=summary*", async (route) => {
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

    await page.route("**/api/v1/organizations/*/home", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/home$/);
        const organizationId = match?.[1];
        const payload = organizationId ? homeResponsesById[organizationId] : undefined;

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

test.describe("V8 AI Organization entry flow", () => {
    test("lands on the AI Organization setup flow with a dominant creation entrypoint", async ({ page }, testInfo) => {
        await mockOrganizationEntryApis(page);

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");

        await expect(page.getByRole("heading", { name: "Create AI Organization" })).toBeVisible();
        await expect(page.getByText("AI Organization Setup")).toBeVisible();
        await expect(page.getByRole("button", { name: "Explore Templates" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Start Empty", exact: true })).toBeVisible();
        await expect(page.getByText("not a blank assistant session")).toBeVisible();
        await expect(page.getByText("Why start here")).toBeVisible();
        await expect(page.getByText("Mission Control")).toHaveCount(0);
        await expect(page.getByText("New Chat")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "dominant-entrypoint.png");
    });

    test("shows the starter template path with user-facing terminology only", async ({ page }, testInfo) => {
        await mockOrganizationEntryApis(page);

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await page.getByRole("button", { name: /Start from template/i }).click();

        const starterCard = page.getByRole("button", { name: /Engineering Starter/i });
        await expect(starterCard).toBeVisible();
        await expect(starterCard).toContainText("AI Organization starter");
        await expect(starterCard).toContainText("Team Lead");
        await expect(starterCard).toContainText("Advisors");
        await expect(starterCard).toContainText("Departments");
        await expect(starterCard).toContainText("Specialists");
        await expect(starterCard).toContainText("AI Engine Settings");
        await expect(starterCard).toContainText("Memory & Personality");
        await expect(page.getByText("Hidden until Advanced mode")).toBeVisible();
        await expect(page.getByText("Learn about AI Organizations")).toBeVisible();
        await expect(page.getByText("Inception")).toHaveCount(0);
        await expect(page.getByText("Soma Kernel")).toHaveCount(0);
        await expect(page.getByText("Central Council")).toHaveCount(0);
        await expect(page.getByText("Provider Policy")).toHaveCount(0);
        await expect(page.getByText("Identity / Continuity")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-mode.png");
    });

    test("creates an AI Organization from a template and lands on the organization home", async ({ page }, testInfo) => {
        let capturedRequestBody: Record<string, unknown> | null = null;

        await mockOrganizationEntryApis(page, {
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return {
                    status: 201,
                    body: { ok: true, data: createdTemplateOrganization },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await page.getByRole("button", { name: /Start from template/i }).click();
        await page.getByLabel("AI Organization name").fill(createdTemplateOrganization.name);
        await page.getByLabel("Purpose").fill(createdTemplateOrganization.purpose);
        await page.getByRole("button", { name: "Create AI Organization" }).click();

        await expect(page).toHaveURL(/\/organizations\/org-123$/);
        expect(capturedRequestBody).toEqual({
            name: createdTemplateOrganization.name,
            purpose: createdTemplateOrganization.purpose,
            start_mode: "template",
            template_id: starterTemplate.id,
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdTemplateOrganization.name })).toBeVisible();
        await expect(page.getByText("Team Lead status")).toBeVisible();
        await expect(page.getByText("Organization overview")).toBeVisible();
        await expect(page.getByText("Team Lead workspace coming soon")).toBeVisible();
        await expect(page.getByText("Mission Control")).toHaveCount(0);
        await expect(page.getByText("New Chat")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-success-home.png");
    });

    test("creates an empty-start AI Organization and keeps the organization frame after success", async ({ page }, testInfo) => {
        let capturedRequestBody: Record<string, unknown> | null = null;

        await mockOrganizationEntryApis(page, {
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return {
                    status: 201,
                    body: { ok: true, data: createdEmptyOrganization },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await page.getByRole("button", { name: /Start Empty$/i }).click();
        await page.getByLabel("AI Organization name").fill(createdEmptyOrganization.name);
        await page.getByLabel("Purpose").fill(createdEmptyOrganization.purpose);
        await page.getByRole("button", { name: "Create AI Organization" }).click();

        await expect(page).toHaveURL(/\/organizations\/org-456$/);
        expect(capturedRequestBody).toEqual({
            name: createdEmptyOrganization.name,
            purpose: createdEmptyOrganization.purpose,
            start_mode: "empty",
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdEmptyOrganization.name })).toBeVisible();
        await expect(page.getByText("Team Lead status")).toBeVisible();
        await expect(page.getByText("Started from")).toBeVisible();
        await expect(page.getByText("Empty")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "empty-success-home.png");
    });

    test("keeps creation available when recent organizations fail and shows retry guidance", async ({ page }, testInfo) => {
        let organizationsAttempt = 0;

        await mockOrganizationEntryApis(page, {
            organizationsStatus: 500,
            organizationsError: "Recent AI Organizations are unavailable right now.",
        });

        await page.unroute("**/api/v1/organizations?view=summary*");
        await page.route("**/api/v1/organizations?view=summary*", async (route) => {
            organizationsAttempt += 1;
            const isFailure = organizationsAttempt === 1;
            await route.fulfill({
                status: isFailure ? 500 : 200,
                contentType: "application/json",
                body: JSON.stringify(
                    isFailure
                        ? { ok: false, error: "Recent AI Organizations are unavailable right now." }
                        : {
                              ok: true,
                              data: [
                                  {
                                      id: "org-999",
                                      name: "Atlas",
                                      purpose: "Resume me later",
                                      start_mode: "empty",
                                      team_lead_label: "Team Lead",
                                      advisor_count: 0,
                                      department_count: 0,
                                      specialist_count: 0,
                                      status: "ready",
                                  },
                              ],
                          },
                ),
            });
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");

        await expect(page.getByText("Recent AI Organizations are unavailable", { exact: true })).toBeVisible();
        await expect(page.getByText("You can still create a new AI Organization below while we retry your recent organizations.")).toBeVisible();
        await expect(page.getByRole("button", { name: "Retry recent AI Organizations" })).toBeVisible();
        await page.getByRole("button", { name: /Start from template/i }).click();
        await expect(page.getByRole("button", { name: /Engineering Starter/i })).toBeVisible();

        await page.getByRole("button", { name: "Retry recent AI Organizations" }).click();
        await expect(page.getByText("Atlas")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "retry-recovery.png");
    });
});
