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
    ai_engine_profile_id: "starter_defaults",
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
    ai_engine_profile_id: "starter_defaults",
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
        actionHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        aiEngineUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
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
        actionHandler,
        aiEngineUpdateHandler,
        homeResponsesById = {
            [createdTemplateOrganization.id]: createdTemplateOrganization,
            [createdEmptyOrganization.id]: createdEmptyOrganization,
        },
    } = options ?? {};

    const mutableHomeResponsesById = structuredClone(homeResponsesById);

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
                          request_label: "Plan next steps for this organization",
                          headline: "Team Lead plan for Northstar Labs",
                          summary: "Team Lead recommends a focused first delivery loop for this AI Organization.",
                          priority_steps: [
                              "Align the first outcome with the AI Organization purpose.",
                              "Use the first Department as the routing layer for work.",
                          ],
                          suggested_follow_ups: [
                              "Review my organization setup",
                              "What should I focus on first?",
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

    await page.route("**/api/v1/organizations/*/ai-engine", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/ai-engine$/);
        const organizationId = match?.[1];

        if (aiEngineUpdateHandler) {
            const response = aiEngineUpdateHandler(requestBody);
            await route.fulfill({
                status: response.status,
                contentType: "application/json",
                body: JSON.stringify(response.body),
            });
            return;
        }

        const summaries: Record<string, string> = {
            starter_defaults: "Starter Defaults",
            balanced: "Balanced",
            high_reasoning: "High Reasoning",
            fast_lightweight: "Fast & Lightweight",
            deep_planning: "Deep Planning",
        };

        if (organizationId && mutableHomeResponsesById[organizationId]) {
            mutableHomeResponsesById[organizationId] = {
                ...(mutableHomeResponsesById[organizationId] as Record<string, unknown>),
                ai_engine_profile_id: requestBody.profile_id,
                ai_engine_settings_summary: summaries[String(requestBody.profile_id ?? "")] ?? "Balanced",
            };
        }

        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
                ok: true,
                data: organizationId ? mutableHomeResponsesById[organizationId] : null,
            }),
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

    test("creates an AI Organization from a template and starts a guided Team Lead workflow", async ({ page }, testInfo) => {
        let capturedRequestBody: Record<string, unknown> | null = null;
        let capturedActionBody: Record<string, unknown> | null = null;

        await mockOrganizationEntryApis(page, {
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return {
                    status: 201,
                    body: { ok: true, data: createdTemplateOrganization },
                };
            },
            actionHandler: (requestBody) => {
                capturedActionBody = requestBody;
                return {
                    status: 200,
                    body: {
                        ok: true,
                        data: {
                            action: "plan_next_steps",
                            request_label: "Plan next steps for this organization",
                            headline: "Team Lead plan for Northstar Labs",
                            summary: "Team Lead recommends a focused first delivery loop for this AI Organization.",
                            priority_steps: [
                                "Align the first outcome with the AI Organization purpose.",
                                "Use the first Department as the routing layer for work.",
                            ],
                            suggested_follow_ups: [
                                "Review my organization setup",
                                "What should I focus on first?",
                            ],
                        },
                    },
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
        await expect(page.getByRole("heading", { name: createdTemplateOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Team Lead ready")).toBeVisible();
        await expect(page.getByText("Work with the Team Lead")).toBeVisible();
        await expect(page.getByText("Organization overview")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Personality" })).toBeVisible();
        await expect(page.getByText("Advisor support")).toBeVisible();
        await expect(page.getByText("Department view")).toBeVisible();
        await expect(page.getByText("Planning review")).toBeVisible();
        await expect(page.getByText("Started from Engineering Starter")).toBeVisible();
        await expect(page.getByText("What this affects")).toHaveCount(2);
        await expect(page.getByText("Response style")).toBeVisible();
        await expect(page.getByText("Planning depth")).toBeVisible();
        await expect(page.getByText("Working tone")).toBeVisible();
        await expect(page.getByText("Context continuity")).toBeVisible();
        await expect(page.getByRole("button", { name: /Plan next steps for this organization/i })).toBeVisible();
        await expect(page.getByRole("button", { name: "Review Advisors" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Open Departments" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Review AI Engine Settings" }).first()).toBeVisible();

        await page.getByRole("button", { name: "Review Advisors" }).first().click();
        await expect(page.getByRole("heading", { name: "Advisor details" })).toBeVisible();
        await expect(page.getByText("Planning Advisor")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Work with the Team Lead")).toBeVisible();
        await page.getByRole("button", { name: "Back to Team Lead" }).click();

        await page.getByRole("button", { name: "Open Departments" }).last().click();
        await expect(page.getByRole("heading", { name: "Department details" })).toBeVisible();
        await expect(page.getByText("Planning Department")).toBeVisible();
        await expect(page.getByText("2 Specialists visible here.").first()).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Work with the Team Lead")).toBeVisible();
        await page.getByRole("button", { name: "Back to Team Lead" }).click();

        await page.getByRole("button", { name: "Review AI Engine Settings" }).last().click();
        await expect(page.getByRole("heading", { name: "AI Engine Settings details" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Change AI Engine" })).toBeVisible();
        await expect(page.getByText("Organization-wide AI engine", { exact: true })).toBeVisible();
        await expect(page.getByText("Team defaults", { exact: true })).toBeVisible();
        await expect(page.getByText("Specific role overrides", { exact: true })).toBeVisible();
        await expect(page.getByText("Current profile: Starter defaults included.")).toBeVisible();
        await page.getByRole("button", { name: "Change AI Engine" }).click();
        await expect(page.getByRole("heading", { name: "Choose an AI Engine profile" })).toBeVisible();
        await expect(page.getByText("Balanced")).toBeVisible();
        await expect(page.getByText("High Reasoning")).toBeVisible();
        await page.getByRole("button", { name: /High Reasoning/i }).click();
        await page.getByRole("button", { name: "Use selected AI Engine" }).click();
        await expect(page.getByText("Current profile: High Reasoning.")).toBeVisible();
        await expect(page.getByText("The current AI Engine Settings profile is high reasoning and shapes how the organization responds, plans, and carries work forward.")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Work with the Team Lead")).toBeVisible();
        await page.getByRole("button", { name: "Back to Team Lead" }).click();

        await page.getByRole("button", { name: /Plan next steps for this organization/i }).click();
        expect(capturedActionBody).toEqual({ action: "plan_next_steps" });
        await expect(page.getByText("Team Lead plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("Priority steps")).toBeVisible();
        await expect(page.getByText("Keep moving with")).toBeVisible();
        await expect(page.getByText("Mission Control")).toHaveCount(0);
        await expect(page.getByText("New Chat")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-guided-workflow.png");
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
        await expect(page.getByRole("heading", { name: createdEmptyOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Team Lead ready")).toBeVisible();
        await expect(page.getByText("Work with the Team Lead")).toBeVisible();
        await expect(page.getByText("Started from", { exact: true })).toBeVisible();
        await expect(page.getByText("Empty", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Personality" })).toBeVisible();
        await expect(page.getByText("Advisor roles appear here once they are added")).toBeVisible();
        await expect(page.getByText("Add the first Department when ready")).toBeVisible();
        await expect(page.getByText("The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.")).toBeVisible();
        await expect(page.getByText("Memory & Personality stay on a simple starter posture so the Team Lead keeps a consistent tone and working style.")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "empty-success-home.png");
    });

    test("preserves organization context when a guided Team Lead action fails and then succeeds on retry", async ({ page }, testInfo) => {
        let actionAttempts = 0;

        await mockOrganizationEntryApis(page, {
            createHandler: () => ({
                status: 201,
                body: { ok: true, data: createdTemplateOrganization },
            }),
            actionHandler: (requestBody) => {
                actionAttempts += 1;
                if (actionAttempts === 1) {
                    return {
                        status: 500,
                        body: { ok: false, error: "Team Lead guidance is unavailable right now." },
                    };
                }

                return {
                    status: 200,
                    body: {
                        ok: true,
                        data: {
                            action: requestBody.action ?? "plan_next_steps",
                            request_label: "Plan next steps for this organization",
                            headline: "Team Lead plan for Northstar Labs",
                            summary: "Team Lead recommends a focused first delivery loop for this AI Organization.",
                            priority_steps: [
                                "Align the first outcome with the AI Organization purpose.",
                                "Use the first Department as the routing layer for work.",
                            ],
                            suggested_follow_ups: [
                                "Review my organization setup",
                                "What should I focus on first?",
                            ],
                        },
                    },
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
        await page.getByRole("button", { name: /Plan next steps for this organization/i }).click();

        await expect(page.getByText("Team Lead guidance is unavailable", { exact: true })).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Team Lead ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Team Lead for Northstar Labs" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Personality" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Retry Team Lead action" })).toBeVisible();
        await expectNoForbiddenCopy(page);

        await page.getByRole("button", { name: "Retry Team Lead action" }).click();

        await expect(page.getByText("Team Lead plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("Priority steps")).toBeVisible();
        await expect(page.getByText("Keep moving with")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "guided-retry-recovery.png");
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
