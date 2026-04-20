import { test, expect, type Page, type TestInfo } from "@playwright/test";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

const starterTemplate = {
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

const createdTemplateOrganization = {
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
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Set up later in Advanced mode",
    status: "ready",
    departments: [],
};

const recentActivityByOrganizationId: Record<string, Array<{
    id: string;
    name: string;
    last_run_at: string;
    status: "success" | "warning" | "failed";
    summary: string;
}>> = {
    "org-123": [
        {
            id: "activity-1",
            name: "Department check",
            last_run_at: "2026-03-19T17:58:00Z",
            status: "success",
            summary: "No issues detected",
        },
        {
            id: "activity-2",
            name: "Specialist review",
            last_run_at: "2026-03-19T17:55:00Z",
            status: "warning",
            summary: "2 items flagged",
        },
    ],
    "org-456": [],
};

const automationsByOrganizationId: Record<string, Array<{
    id: string;
    name: string;
    purpose: string;
    trigger_type: "scheduled" | "event_driven";
    owner_label: string;
    status: "success" | "warning" | "failed";
    watches: string;
    trigger_summary: string;
    recent_outcomes: Array<{
        summary: string;
        occurred_at: string;
    }>;
}>> = {
    "org-123": [
        {
            id: "department-readiness-review",
            name: "Department readiness review",
            purpose: "Reviews the current Department structure and operating readiness without taking action.",
            trigger_type: "scheduled",
            owner_label: "Team: Platform Department",
            status: "success",
            watches: "Watches Platform Department structure, specialist coverage, and current organization defaults inside Northstar Labs.",
            trigger_summary: "Runs every minute and also after organization setup, Team Lead guidance, AI Engine changes, or Response Style changes.",
            recent_outcomes: [
                {
                    summary: "No issues detected",
                    occurred_at: "2026-03-19T17:58:00Z",
                },
            ],
        },
        {
            id: "agent-type-readiness-review",
            name: "Agent type readiness review",
            purpose: "Reviews a specialist profile and its inherited defaults without taking action.",
            trigger_type: "event_driven",
            owner_label: "Specialist role: Planner",
            status: "warning",
            watches: "Watches the Planner specialist role, its working focus, and the defaults it inherits inside Northstar Labs.",
            trigger_summary: "Runs after organization setup, AI Engine changes, or Response Style changes.",
            recent_outcomes: [
                {
                    summary: "2 items flagged",
                    occurred_at: "2026-03-19T17:55:00Z",
                },
            ],
        },
    ],
    "org-456": [],
};

const learningInsightsByOrganizationId: Record<string, Array<{
    id: string;
    summary: string;
    source: string;
    observed_at: string;
    strength: "emerging" | "consistent" | "strong";
}>> = {
    "org-123": [
        {
            id: "insight-1",
            summary: "Platform Department is building a steadier execution lane for the organization.",
            source: "Team: Platform Department",
            observed_at: "2026-03-19T17:58:00Z",
            strength: "strong",
        },
        {
            id: "insight-2",
            summary: "Planner specialists are identifying recurring gaps while turning organization goals into practical next steps, delivery sequencing, and clear priorities.",
            source: "Specialist role: Planner",
            observed_at: "2026-03-19T17:55:00Z",
            strength: "emerging",
        },
    ],
    "org-456": [],
};

const organizationChatPlaceholder = /Tell Soma what you want to plan, review, create, or execute/i;

type ChatRequestBody = {
    messages?: Array<{
        role?: string;
        content?: string;
    }>;
};

type RouteResponse = {
    status: number;
    body: unknown;
};

function answerEnvelope(
    text: string,
    options?: {
        askClass?: string;
        consultations?: Array<{ member: string; summary: string }>;
        artifacts?: Array<Record<string, unknown>>;
    },
): RouteResponse {
    return {
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
                    text,
                    ask_class: options?.askClass,
                    consultations: options?.consultations ?? [],
                    tools_used: [],
                    artifacts: options?.artifacts ?? [],
                },
            },
        },
    };
}

function lastUserMessage(requestBody: ChatRequestBody): string {
    const messages = Array.isArray(requestBody.messages) ? requestBody.messages : [];
    return messages[messages.length - 1]?.content ?? "";
}

function applyOrganizationAIEngineToDepartments(home: typeof createdTemplateOrganization, profileId: string | undefined, summary: string) {
    return {
        ...home,
        ai_engine_profile_id: profileId,
        ai_engine_settings_summary: summary,
        departments: home.departments.map((department) =>
            department.inherits_organization_ai_engine
                ? {
                      ...department,
                      ai_engine_effective_profile_id: profileId,
                      ai_engine_effective_summary: summary,
                      agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                          profile.inherits_department_ai_engine
                              ? {
                                    ...profile,
                                    ai_engine_effective_profile_id: profileId,
                                    ai_engine_effective_summary: summary,
                                }
                              : profile,
                      ),
                  }
                : department,
        ),
    };
}

function applyResponseContract(home: typeof createdTemplateOrganization, profileId: string | undefined, summary: string) {
    return {
        ...home,
        response_contract_profile_id: profileId,
        response_contract_summary: summary,
        departments: home.departments.map((department) => ({
            ...department,
            agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                profile.inherits_default_response_contract
                    ? {
                          ...profile,
                          response_contract_effective_profile_id: profileId,
                          response_contract_effective_summary: summary,
                      }
                    : profile,
            ),
        })),
    };
}

function applyAgentTypeAIEngine(
    home: typeof createdTemplateOrganization,
    departmentId: string,
    agentTypeId: string,
    profileId: string | undefined,
    summary: string,
) {
    return {
        ...home,
        departments: home.departments.map((department) =>
            department.id === departmentId
                ? {
                      ...department,
                      agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                          profile.id === agentTypeId
                              ? {
                                    ...profile,
                                    ai_engine_binding_profile_id: profileId,
                                    ai_engine_effective_profile_id: profileId ?? department.ai_engine_effective_profile_id,
                                    ai_engine_effective_summary: profileId ? summary : department.ai_engine_effective_summary,
                                    inherits_department_ai_engine: !profileId,
                                }
                              : profile,
                      ),
                  }
                : department,
        ),
    };
}

function applyAgentTypeResponseContract(
    home: typeof createdTemplateOrganization,
    departmentId: string,
    agentTypeId: string,
    profileId: string | undefined,
    summary: string,
) {
    return {
        ...home,
        departments: home.departments.map((department) =>
            department.id === departmentId
                ? {
                      ...department,
                      agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                          profile.id === agentTypeId
                              ? {
                                    ...profile,
                                    response_contract_binding_profile_id: profileId,
                                    response_contract_effective_profile_id: profileId ?? home.response_contract_profile_id,
                                    response_contract_effective_summary: profileId ? summary : home.response_contract_summary,
                                    inherits_default_response_contract: !profileId,
                                }
                              : profile,
                      ),
                  }
                : department,
        ),
    };
}

async function saveScreenshot(page: Page, testInfo: TestInfo, name: string) {
    await page.screenshot({
        path: testInfo.outputPath(name),
        fullPage: true,
    });
}

async function clickStartMode(page: Page, label: RegExp | string) {
    const button = page.getByRole("button", { name: label });
    await expect(button).toBeVisible();
    await button.scrollIntoViewIfNeeded();
    await button.evaluate((element) => {
        (element as HTMLButtonElement).click();
    });
}

async function openOrganizationSetup(page: Page) {
    const startEmptyButton = page.getByRole("button", { name: "Start Empty", exact: true });
    if (await startEmptyButton.isVisible().catch(() => false)) {
        return;
    }

    const setupTrigger = page.getByRole("button", { name: "Create or open AI Organizations" });
    await expect(setupTrigger).toBeVisible();
    await setupTrigger.click();
    await expect(startEmptyButton).toBeVisible();
}

async function openCreatedOrganization(page: Page, organizationId: string) {
    const organizationPath = `/organizations/${organizationId}`;
    const organizationUrl = new RegExp(`${organizationPath}$`);

    try {
        await page.waitForURL(organizationUrl, {
            timeout: 5000,
            waitUntil: "domcontentloaded",
        });
    } catch {
        for (let attempt = 0; attempt < 2; attempt += 1) {
            try {
                await page.goto(organizationPath, { waitUntil: "domcontentloaded" });
            } catch (error) {
                if (!(error instanceof Error) || !error.message.includes("ERR_ABORTED")) {
                    throw error;
                }
            }

            if (page.url().endsWith(organizationPath)) {
                break;
            }
        }
    }

    await expect(page).toHaveURL(organizationUrl, { timeout: 15000 });
    await page.waitForLoadState("domcontentloaded");
}

function recentOrganizationLink(page: Page, organizationName: string) {
    return page.getByRole("link", { name: `Return to ${organizationName}` });
}

function recentOrganizationOpenButton(page: Page, organizationName: string) {
    return page.getByRole("button", {
        name: new RegExp(`${organizationName}.*Open AI Organization`, "i"),
    });
}

async function sendOrganizationMessage(page: Page, content: string) {
    const input = page.getByPlaceholder(organizationChatPlaceholder);
    await input.fill(content);
    await input.press("Enter");
}

async function expectNoForbiddenCopy(page: Page) {
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

async function mockOrganizationEntryApis(
    page: Page,
    options?: {
        templates?: unknown[];
        organizations?: unknown[];
        templateStatus?: number;
        templateError?: string;
        organizationsStatus?: number;
        organizationsError?: string;
        organizationsSummaryResponses?: Array<{ status: number; body: unknown }>;
        createHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        actionHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        aiEngineUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        departmentAIEngineUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        agentTypeAIEngineUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        agentTypeResponseContractUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        responseContractUpdateHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
        homeResponsesById?: Record<string, unknown>;
        automationsById?: Record<string, unknown>;
        loopActivityById?: Record<string, unknown>;
        learningInsightsById?: Record<string, unknown>;
        chatHandler?: (requestBody: Record<string, unknown>) => { status: number; body: unknown };
    },
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
        aiEngineUpdateHandler,
        departmentAIEngineUpdateHandler,
        agentTypeAIEngineUpdateHandler,
        agentTypeResponseContractUpdateHandler,
        responseContractUpdateHandler,
        homeResponsesById = {
            [createdTemplateOrganization.id]: createdTemplateOrganization,
            [createdEmptyOrganization.id]: createdEmptyOrganization,
        },
        automationsById = automationsByOrganizationId,
        loopActivityById = recentActivityByOrganizationId,
        learningInsightsById = learningInsightsByOrganizationId,
        chatHandler,
    } = options ?? {};

    const mutableHomeResponsesById = structuredClone(homeResponsesById);
    const mutableAutomationsById = structuredClone(automationsById);
    const mutableLoopActivityById = structuredClone(loopActivityById);
    const mutableLearningInsightsById = structuredClone(learningInsightsById);
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

    await page.route("**/api/v1/organizations/*/loop-activity", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/loop-activity$/);
        const organizationId = match?.[1];
        const payload = organizationId ? mutableLoopActivityById[organizationId] : undefined;

        await route.fulfill({
            status: payload !== undefined ? 200 : 404,
            contentType: "application/json",
            body: JSON.stringify(
                payload !== undefined
                    ? { ok: true, data: payload }
                    : { ok: false, error: "Activity unavailable" },
            ),
        });
    });

    await page.route("**/api/v1/organizations/*/learning-insights", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/learning-insights$/);
        const organizationId = match?.[1];
        const payload = organizationId ? mutableLearningInsightsById[organizationId] : undefined;

        await route.fulfill({
            status: payload !== undefined ? 200 : 404,
            contentType: "application/json",
            body: JSON.stringify(
                payload !== undefined
                    ? { ok: true, data: payload }
                    : { ok: false, error: "Memory & Continuity updates unavailable" },
            ),
        });
    });

    await page.route("**/api/v1/organizations/*/automations", async (route) => {
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/automations$/);
        const organizationId = match?.[1];
        const payload = organizationId ? mutableAutomationsById[organizationId] : undefined;

        await route.fulfill({
            status: payload !== undefined ? 200 : 404,
            contentType: "application/json",
            body: JSON.stringify(
                payload !== undefined
                    ? { ok: true, data: payload }
                    : { ok: false, error: "Automations unavailable" },
            ),
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
            mutableHomeResponsesById[organizationId] = applyOrganizationAIEngineToDepartments(
                mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization,
                String(requestBody.profile_id ?? ""),
                summaries[String(requestBody.profile_id ?? "")] ?? "Balanced",
            );
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

    await page.route("**/api/v1/organizations/*/response-contract", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/response-contract$/);
        const organizationId = match?.[1];

        if (responseContractUpdateHandler) {
            const response = responseContractUpdateHandler(requestBody);
            await route.fulfill({
                status: response.status,
                contentType: "application/json",
                body: JSON.stringify(response.body),
            });
            return;
        }

        const summaries: Record<string, string> = {
            clear_balanced: "Clear & Balanced",
            structured_analytical: "Structured & Analytical",
            concise_direct: "Concise & Direct",
            warm_supportive: "Warm & Supportive",
        };

        if (organizationId && mutableHomeResponsesById[organizationId]) {
            mutableHomeResponsesById[organizationId] = applyResponseContract(
                mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization,
                String(requestBody.profile_id ?? ""),
                summaries[String(requestBody.profile_id ?? "")] ?? "Clear & Balanced",
            );
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

    await page.route("**/api/v1/organizations/*/departments/*/ai-engine", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/departments\/([^/]+)\/ai-engine$/);
        const organizationId = match?.[1];
        const departmentId = match?.[2];

        if (departmentAIEngineUpdateHandler) {
            const response = departmentAIEngineUpdateHandler(requestBody);
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

        if (organizationId && departmentId && mutableHomeResponsesById[organizationId]) {
            const home = mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization;
            mutableHomeResponsesById[organizationId] = {
                ...home,
                departments: home.departments.map((department) => {
                    if (department.id !== departmentId) {
                        return department;
                    }
                    if (requestBody.revert_to_organization_default) {
                        return {
                            ...department,
                            ai_engine_override_profile_id: undefined,
                            ai_engine_override_summary: undefined,
                            ai_engine_effective_profile_id: home.ai_engine_profile_id,
                            ai_engine_effective_summary: home.ai_engine_settings_summary,
                            inherits_organization_ai_engine: true,
                            agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                                profile.inherits_department_ai_engine
                                    ? {
                                          ...profile,
                                          ai_engine_effective_profile_id: home.ai_engine_profile_id,
                                          ai_engine_effective_summary: home.ai_engine_settings_summary,
                                      }
                                    : profile,
                            ),
                        };
                    }
                    const profileId = String(requestBody.profile_id ?? "");
                    return {
                        ...department,
                        ai_engine_override_profile_id: profileId,
                        ai_engine_override_summary: summaries[profileId],
                        ai_engine_effective_profile_id: profileId,
                        ai_engine_effective_summary: summaries[profileId] ?? department.ai_engine_effective_summary,
                        inherits_organization_ai_engine: false,
                        agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                            profile.inherits_department_ai_engine
                                ? {
                                      ...profile,
                                      ai_engine_effective_profile_id: profileId,
                                      ai_engine_effective_summary: summaries[profileId] ?? department.ai_engine_effective_summary,
                                  }
                                : profile,
                        ),
                    };
                }),
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

    await page.route("**/api/v1/organizations/*/departments/*/agent-types/*/ai-engine", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/departments\/([^/]+)\/agent-types\/([^/]+)\/ai-engine$/);
        const organizationId = match?.[1];
        const departmentId = match?.[2];
        const agentTypeId = match?.[3];

        if (agentTypeAIEngineUpdateHandler) {
            const response = agentTypeAIEngineUpdateHandler(requestBody);
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

        if (organizationId && departmentId && agentTypeId && mutableHomeResponsesById[organizationId]) {
            const home = mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization;
            mutableHomeResponsesById[organizationId] = requestBody.use_team_default
                ? applyAgentTypeAIEngine(home, departmentId, agentTypeId, undefined, "")
                : applyAgentTypeAIEngine(
                      home,
                      departmentId,
                      agentTypeId,
                      String(requestBody.profile_id ?? ""),
                      summaries[String(requestBody.profile_id ?? "")] ?? "Balanced",
                  );
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

    await page.route("**/api/v1/organizations/*/departments/*/agent-types/*/response-contract", async (route) => {
        const requestBody = route.request().postDataJSON() as Record<string, unknown>;
        const url = new URL(route.request().url());
        const match = url.pathname.match(/\/api\/v1\/organizations\/([^/]+)\/departments\/([^/]+)\/agent-types\/([^/]+)\/response-contract$/);
        const organizationId = match?.[1];
        const departmentId = match?.[2];
        const agentTypeId = match?.[3];

        if (agentTypeResponseContractUpdateHandler) {
            const response = agentTypeResponseContractUpdateHandler(requestBody);
            await route.fulfill({
                status: response.status,
                contentType: "application/json",
                body: JSON.stringify(response.body),
            });
            return;
        }

        const summaries: Record<string, string> = {
            clear_balanced: "Clear & Balanced",
            structured_analytical: "Structured & Analytical",
            concise_direct: "Concise & Direct",
            warm_supportive: "Warm & Supportive",
        };

        if (organizationId && departmentId && agentTypeId && mutableHomeResponsesById[organizationId]) {
            if (requestBody.use_organization_or_team_default) {
                mutableHomeResponsesById[organizationId] = applyAgentTypeResponseContract(
                    mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization,
                    departmentId,
                    agentTypeId,
                    undefined,
                    "",
                );
            } else {
                mutableHomeResponsesById[organizationId] = applyAgentTypeResponseContract(
                    mutableHomeResponsesById[organizationId] as typeof createdTemplateOrganization,
                    departmentId,
                    agentTypeId,
                    String(requestBody.profile_id ?? ""),
                    summaries[String(requestBody.profile_id ?? "")] ?? "Clear & Balanced",
                );
            }
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
    test("creates an AI Organization from a template and starts a guided Soma workflow", async ({ page }, testInfo) => {
        test.slow();
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
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start from template/i);
        await page.getByLabel("AI Organization name").fill(createdTemplateOrganization.name);
        await page.getByLabel("Purpose").fill(createdTemplateOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdTemplateOrganization.id);
        expect(capturedRequestBody).toEqual({
            name: createdTemplateOrganization.name,
            purpose: createdTemplateOrganization.purpose,
            start_mode: "template",
            template_id: starterTemplate.id,
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdTemplateOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByText("Start here in this organization")).toBeVisible();
        await expect(page.getByRole("button", { name: "Open Soma conversation" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Open team design lane" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Review organization setup" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await expect(page.getByText("Create teams with Soma")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Quick Checks" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Automations" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "What the Organization Is Retaining" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expect(page.getByText("Advisor support", { exact: true })).toBeVisible();
        await expect(page.getByText("Visible specialist roles", { exact: true })).toBeVisible();
        await expect(page.getByText("Department readiness review • Scheduled")).toBeVisible();
        await expect(page.getByText("Agent type readiness review • Event-driven")).toBeVisible();
        await expect(page.getByText("Your AI Organization is actively working through recent reviews, checks, and updates in the background.")).toBeVisible();
        await expect(page.getByText("Department check")).toBeVisible();
        await expect(page.getByText("Specialist review")).toBeVisible();
        await expect(page.getByText("No issues detected").first()).toBeVisible();
        await expect(page.getByText("2 items flagged").first()).toBeVisible();
        await expect(page.getByText("Platform Department is building a steadier execution lane for the organization.").first()).toBeVisible();
        await expect(page.getByText("Team: Platform Department").first()).toBeVisible();
        await expect(page.getByText("Strong", { exact: true })).toBeVisible();
        await expect(page.getByText("Planning review")).toBeVisible();
        await expect(page.getByText("Started from", { exact: true })).toBeVisible();
        await expect(page.getByText("Engineering Starter", { exact: true })).toBeVisible();
        await expect(page.getByText("What this affects")).toHaveCount(2);
        await expect(page.getByText("What this shapes")).toBeVisible();
        await expect(page.getByText("Response style", { exact: true })).toBeVisible();
        await expect(page.getByText("Planning depth", { exact: true })).toBeVisible();
        await expect(page.getByText("Tone", { exact: true })).toBeVisible();
        await expect(page.getByText("Structure", { exact: true })).toBeVisible();
        await expect(page.getByText("Verbosity", { exact: true })).toBeVisible();
        await expect(page.getByText("Durable memory recall")).toBeVisible();
        await expect(page.getByText("Temporary planning continuity")).toBeVisible();
        await expect(page.getByText("Plan the next move")).toBeVisible();
        await expect(page.getByText("Review current state")).toBeVisible();
        await expect(page.getByText("Run a governed change")).toBeVisible();
        await expect(page.getByText("Create an artifact")).toBeVisible();
        await expect(page.getByRole("button", { name: "Review Advisors" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Open Departments" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Review Automations" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Review AI Engine Settings" }).first()).toBeVisible();
        await expect(page.getByRole("button", { name: "Review Response Style" }).first()).toBeVisible();

        await page.getByRole("button", { name: "Open team design lane" }).click();
        await expect(page.getByText("Choose a guided team-design action")).toBeVisible();
        await page.getByRole("button", { name: "Review organization setup" }).click();
        await expect(page.getByRole("heading", { name: "Department details" })).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Review Automations" }).first().click();
        await expect(page.getByRole("heading", { name: "Automation details" })).toBeVisible();
        await expect(page.getByText("Department readiness review", { exact: true })).toBeVisible();
        await expect(page.getByText("Team: Platform Department").first()).toBeVisible();
        await expect(page.getByText("What it watches")).toHaveCount(2);
        await expect(page.getByText("How it runs")).toHaveCount(2);
        await expect(page.getByText("Recent outcomes")).toHaveCount(2);
        await expect(page.getByText("Runs every minute and also after organization setup, Team Lead guidance, AI Engine changes, or Response Style changes.")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Review Advisors" }).first().click();
        await expect(page.getByRole("heading", { name: "Advisor details" })).toBeVisible();
        await expect(page.getByText("Planning Advisor")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Open Departments" }).last().click();
        await expect(page.getByRole("heading", { name: "Department details" })).toBeVisible();
        await expect(page.getByText("Platform Department", { exact: true })).toBeVisible();
        await expect(page.getByText("2 Specialists visible here.").first()).toBeVisible();
        await expect(page.getByText("Using Organization Default: Starter defaults included")).toBeVisible();
        await expect(page.getByText("Agent Type Profiles")).toBeVisible();
        await expect(page.getByText("Planner", { exact: true }).first()).toBeVisible();
        await expect(page.getByText("Delivery Specialist", { exact: true }).first()).toBeVisible();
        await expect(page.getByText("Type-specific Engine: High Reasoning")).toBeVisible();
        await expect(page.getByText("Using Team Default: Starter defaults included")).toBeVisible();
        await expect(page.getByText("Type-specific Response Style: Structured & Analytical")).toBeVisible();
        await expect(page.getByText("Using Organization or Team Default: Clear & Balanced")).toBeVisible();
        await page.getByRole("button", { name: "Change for this Team" }).click();
        await expect(page.getByRole("heading", { name: "Choose an AI Engine for this Team" })).toBeVisible();
        await page.getByRole("button", { name: /Balanced/i }).click();
        await page.getByRole("button", { name: "Use selected AI Engine" }).click();
        await expect(page.getByText("Overridden: Balanced")).toBeVisible();
        await expect(page.getByText("Using Team Default: Balanced")).toBeVisible();
        await page.getByRole("button", { name: "Change for this Agent Type" }).last().click();
        await expect(page.getByRole("heading", { name: "Choose an AI Engine for this Agent Type" })).toBeVisible();
        await page.getByRole("button", { name: /Fast & Lightweight/i }).click();
        await page.getByRole("button", { name: "Use selected AI Engine" }).click();
        await expect(page.getByText("Type-specific Engine: Fast & Lightweight")).toBeVisible();
        await expect(page.getByRole("button", { name: "Use Team Default" }).last()).toBeVisible();
        await page.getByRole("button", { name: "Use Team Default" }).last().click();
        await expect(page.getByText("Using Team Default: Balanced")).toBeVisible();
        await page.getByRole("button", { name: "Change Response Style for this Agent Type" }).last().click();
        await expect(page.getByRole("heading", { name: "Choose a Response Style for this Agent Type" })).toBeVisible();
        await page.getByRole("button", { name: /Warm & Supportive/i }).click();
        await page.getByRole("button", { name: "Use selected Response Style" }).click();
        await expect(page.getByText("Type-specific Response Style: Warm & Supportive")).toBeVisible();
        await expect(page.getByRole("button", { name: "Use Organization / Team Default" }).last()).toBeVisible();
        await page.getByRole("button", { name: "Use Organization / Team Default" }).last().click();
        await expect(page.getByText("Using Organization or Team Default: Clear & Balanced")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Review AI Engine Settings" }).last().click();
        await expect(page.getByRole("heading", { name: "AI Engine Settings details" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Change AI Engine" })).toBeVisible();
        await expect(page.getByText("Organization-wide AI engine", { exact: true })).toBeVisible();
        await expect(page.getByText("Team defaults", { exact: true })).toBeVisible();
        await expect(page.getByText("Specific role overrides", { exact: true })).toBeVisible();
        await expect(page.getByText("Current profile: Starter defaults included.")).toBeVisible();
        await page.getByRole("button", { name: "Change AI Engine" }).click();
        await expect(page.getByRole("heading", { name: "Choose an AI Engine profile" })).toBeVisible();
        await expect(page.getByRole("button", { name: /^Balanced/ })).toBeVisible();
        await expect(page.getByRole("button", { name: /^High Reasoning/ })).toBeVisible();
        await page.getByRole("button", { name: /High Reasoning/i }).click();
        await page.getByRole("button", { name: "Use selected AI Engine" }).click();
        await expect(page.getByText("Current profile: High Reasoning.")).toBeVisible();
        await expect(page.getByText("The current AI Engine Settings profile is high reasoning and shapes how the organization responds, plans, and carries work forward.")).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Review Response Style" }).last().click();
        await expect(page.getByRole("heading", { name: "Response Style details" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Change Response Style" })).toBeVisible();
        await expect(page.getByText("Current response style", { exact: true })).toBeVisible();
        await page.getByRole("button", { name: "Change Response Style" }).click();
        await expect(page.getByRole("heading", { name: "Choose a Response Style" })).toBeVisible();
        await page.getByRole("button", { name: /Warm & Supportive/i }).click();
        await page.getByRole("button", { name: "Use selected Response Style" }).click();
        await expect(page.getByText("Current profile: Warm & Supportive.")).toBeVisible();
        await expect(page.getByText("The current Response Style is warm & supportive, which shapes how Soma presents tone, structure, and detail.")).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await page.getByRole("button", { name: "Open Departments" }).last().click();
        await expect(page.getByText("Overridden: Balanced")).toBeVisible();
        await page.getByRole("button", { name: "Revert to Organization Default" }).click();
        await expect(page.getByText("Using Organization Default: High Reasoning")).toBeVisible();
        await expect(page.getByText("Using Team Default: High Reasoning")).toBeVisible();
        await expect(page.getByText("Using Organization or Team Default: Warm & Supportive")).toBeVisible();
        await page.getByRole("button", { name: "Back to Soma" }).click();

        await expect(page.getByRole("link", { name: "Start with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        await expect(page.getByLabel("Tell Soma what team or delivery lane you want to create")).toBeVisible();
        await page.getByLabel("Tell Soma what team or delivery lane you want to create").fill("Help me choose the first priority for this launch.");
        await page.getByRole("button", { name: "Start team design" }).click();
        expect(capturedActionBody).toEqual({
            action: "focus_first",
            request_context: "Help me choose the first priority for this launch.",
        });
        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("Help me choose the first priority for this launch.").last()).toBeVisible();
        await expect(page.getByText("Priority steps")).toBeVisible();
        await expect(page.getByText("Keep moving with")).toBeVisible();
        await expect(page.getByText("Mission Control")).toHaveCount(0);
        await expect(page.getByText("New Chat")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-guided-workflow.png");
    });

    test("creates an empty-start AI Organization and keeps the organization frame after success", async ({ page }, testInfo) => {
        test.slow();
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
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start Empty$/i);
        await page.getByLabel("AI Organization name").fill(createdEmptyOrganization.name);
        await page.getByLabel("Purpose").fill(createdEmptyOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdEmptyOrganization.id);
        expect(capturedRequestBody).toEqual({
            name: createdEmptyOrganization.name,
            purpose: createdEmptyOrganization.purpose,
            start_mode: "empty",
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdEmptyOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await expect(page.getByRole("link", { name: "Start with Soma" })).toBeVisible();
        await expect(page.getByPlaceholder("Tell Soma what you want to plan, review, create, or execute")).toBeFocused();
        await expect(page.getByText("Started from", { exact: true })).toBeVisible();
        await expect(page.getByText("Empty", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Automations" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "What the Organization Is Retaining" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expect(page.getByText("Review support appears here")).toBeVisible();
        await expect(page.getByText("Try reviewing your organization setup").first()).toBeVisible();
        await expect(page.getByText("Reviews appear here")).toBeVisible();
        await expect(page.getByText("No recent activity yet")).toBeVisible();
        await expect(page.getByText("No retained patterns yet")).toBeVisible();
        await expect(page.getByText("The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.")).toBeVisible();
        await expect(page.getByText("The current Response Style is clear & balanced, which shapes how Soma presents tone, structure, and detail.")).toBeVisible();
        await expect(page.getByText("Memory & Continuity stay on a simple starter posture so Soma can keep working continuity without turning every conversation into durable memory.")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "empty-success-home.png");
    });

    test("keeps the Soma draft and last guidance visible after leaving and returning to the workspace", async ({ page }) => {
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        const prompt = page.getByLabel("Tell Soma what team or delivery lane you want to create");
        await prompt.fill("Help me choose the first priority for this launch.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("You asked Soma to help with")).toBeVisible();

        await page.goto("/dashboard");
        await expect(page).toHaveURL(/\/dashboard$/);
        await recentOrganizationLink(page, "Northstar Labs").click();

        await expect(page).toHaveURL(/\/organizations\/org-123$/);
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        await expect(page.getByLabel("Tell Soma what team or delivery lane you want to create")).toHaveValue("Help me choose the first priority for this launch.");
        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("You asked Soma to help with")).toBeVisible();
    });

    test("keeps native team output and external workflow contract paths visibly separated in team design", async ({ page }) => {
        const capturedActionBodies: Record<string, unknown>[] = [];

        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
            actionHandler: (requestBody) => {
                capturedActionBodies.push(requestBody);
                const requestContext = String(requestBody.request_context ?? "");

                if (/n8n/i.test(requestContext)) {
                    return {
                        status: 200,
                        body: {
                            ok: true,
                            data: {
                                action: requestBody.action ?? "plan_next_steps",
                                request_label: "Plan next steps for this organization",
                                headline: "Soma plan for Northstar Labs",
                                summary: "Soma is ready to keep the external workflow path clear for this organization.",
                                priority_steps: [
                                    "Confirm the workflow trigger and expected return shape.",
                                    "Keep the external automation separate from native team execution.",
                                ],
                                suggested_follow_ups: [
                                    "Review your organization setup",
                                    "Choose the first priority",
                                ],
                                execution_contract: {
                                    execution_mode: "external_workflow_contract",
                                    owner_label: "External workflow contract",
                                    external_target: "n8n workflow contract",
                                    summary: "Use an external workflow contract here so Mycelis can invoke it cleanly and return a normalized result without treating it like a native team.",
                                    target_outputs: [
                                        "Normalized workflow result",
                                        "Linked artifact or execution note",
                                    ],
                                },
                            },
                        },
                    };
                }

                return {
                    status: 200,
                    body: {
                        ok: true,
                        data: {
                            action: requestBody.action ?? "plan_next_steps",
                            request_label: "Plan next steps for this organization",
                            headline: "Soma plan for Northstar Labs",
                            summary: "Soma is ready to shape a native creative delivery path for this organization.",
                            priority_steps: [
                                "Align the image goal with the organization purpose.",
                                "Stand up the creative team inside the organization before generating output.",
                            ],
                            suggested_follow_ups: [
                                "Review your organization setup",
                                "Choose the first priority",
                            ],
                            execution_contract: {
                                execution_mode: "native_team",
                                owner_label: "Native Mycelis team",
                                team_name: "Creative Delivery Team",
                                summary: "Use a native creative team here so Soma can coordinate the work and return a reviewable image artifact inside the organization.",
                                target_outputs: [
                                    "Reviewable image artifact",
                                    "Short concept note",
                                ],
                            },
                        },
                    },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        const prompt = page.getByLabel("Tell Soma what team or delivery lane you want to create");

        await prompt.fill("Create a creative team to generate a launch hero image.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Execution path")).toBeVisible();
        await expect(page.getByText("Native Mycelis team").first()).toBeVisible();
        await expect(page.getByText("Creative Delivery Team")).toBeVisible();
        await expect(page.getByText("Reviewable image artifact", { exact: true })).toBeVisible();

        await prompt.fill("Create an n8n workflow contract for inbound leads.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("External workflow contract").first()).toBeVisible();
        await expect(page.getByText("n8n workflow contract", { exact: true })).toBeVisible();
        await expect(page.getByText("Normalized workflow result", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();

        expect(capturedActionBodies).toEqual([
            {
                action: "plan_next_steps",
                request_context: "Create a creative team to generate a launch hero image.",
            },
            {
                action: "plan_next_steps",
                request_context: "Create an n8n workflow contract for inbound leads.",
            },
        ]);
    });

    test("keeps ask-class output cues visible inside the organization workspace chat", async ({ page }) => {
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
            chatHandler: (requestBody) => {
                const content = lastUserMessage(requestBody as ChatRequestBody);
                if (/artifact/i.test(content)) {
                    return answerEnvelope("I prepared a launch brief for review inside Northstar Labs.", {
                        askClass: "governed_artifact",
                        artifacts: [
                            {
                                id: "artifact-launch-brief",
                                type: "document",
                                title: "Launch brief",
                                content_type: "text/markdown",
                                content: "# Launch brief",
                            },
                        ],
                    });
                }

                return answerEnvelope("The architect reviewed the tradeoffs and recommends the safer rollout path.", {
                    askClass: "specialist_consultation",
                    consultations: [
                        {
                            member: "council-architect",
                            summary: "Recommend the safer rollout path.",
                        },
                    ],
                });
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByPlaceholder(organizationChatPlaceholder)).toBeVisible();

        await sendOrganizationMessage(page, "Create an artifact brief for this launch.");
        await expect(page.getByTestId("mission-chat").getByText("Artifact result")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma prepared 1 artifact for review: Launch brief.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByTestId("mission-chat").getByText("Launch brief").first()).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();

        await sendOrganizationMessage(page, "Get specialist advice on the architecture tradeoffs.");
        await expect(page.getByTestId("mission-chat").getByText("Specialist support")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma checked with Architect while shaping this answer: Recommend the safer rollout path.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Soma consulted/i)).toBeVisible();
        await expect(page.getByTestId("mission-chat").getByText("Architect", { exact: true }).last()).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expectNoForbiddenCopy(page);
    });

    test.skip("preserves organization context when a guided Soma action fails and then succeeds on retry", async ({ page }, testInfo) => {
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
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start from template/i);
        await page.getByLabel("AI Organization name").fill(createdTemplateOrganization.name);
        await page.getByLabel("Purpose").fill(createdTemplateOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdTemplateOrganization.id);
        await page.getByRole("button", { name: /Run a quick strategy check/i }).click();

        await expect(page.getByText("Soma guidance is unavailable", { exact: true })).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Retry Soma action" })).toBeVisible();
        await expectNoForbiddenCopy(page);

        await page.getByRole("button", { name: "Retry Soma action" }).click();

        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("Priority steps")).toBeVisible();
        await expect(page.getByText("Keep moving with")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "guided-retry-recovery.png");
    });

});
