import { expect, type Page, type Route } from "@playwright/test";

export const organizationId = "org-ui-testing-agent";
export const chatPlaceholder = /Tell .* what you want to plan, review, create, or execute/i;

export const organizationHome = {
    id: organizationId,
    name: "Northstar Labs",
    purpose: "Run a steady Soma-first AI Organization workflow.",
    start_mode: "template",
    template_name: "Engineering Starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_profile_id: "balanced",
    ai_engine_settings_summary: "Balanced",
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Steady working memory",
    status: "ready",
    departments: [
        {
            id: "platform",
            name: "Platform Department",
            specialist_count: 2,
            ai_engine_effective_profile_id: "balanced",
            ai_engine_effective_summary: "Balanced",
            inherits_organization_ai_engine: true,
            agent_type_profiles: [
                {
                    id: "planner",
                    name: "Planner",
                    helps_with: "Turns organization goals into practical next steps.",
                    ai_engine_effective_profile_id: "balanced",
                    ai_engine_effective_summary: "Balanced",
                    inherits_department_ai_engine: true,
                    response_contract_effective_profile_id: "clear_balanced",
                    response_contract_effective_summary: "Clear & Balanced",
                    inherits_default_response_contract: true,
                },
            ],
        },
    ],
};

const activityItems = [
    {
        id: "activity-1",
        name: "Strategy review",
        last_run_at: "2026-03-26T15:58:00Z",
        status: "success",
        summary: "No issues detected",
    },
];

const automations = [
    {
        id: "department-readiness-review",
        name: "Department readiness review",
        purpose: "Reviews the current department structure without taking action.",
        trigger_type: "scheduled",
        owner_label: "Team: Platform Department",
        status: "success",
        watches: "Watches Platform Department structure and current defaults.",
        trigger_summary: "Runs every minute and after organization changes.",
        recent_outcomes: [
            {
                summary: "No issues detected",
                occurred_at: "2026-03-26T15:58:00Z",
            },
        ],
    },
];

const learningInsights = [
    {
        id: "insight-1",
        summary: "Platform Department is building a steadier execution lane.",
        source: "Team: Platform Department",
        observed_at: "2026-03-26T15:57:00Z",
        strength: "strong",
    },
];

const auditLog = [
    {
        id: "audit-1",
        template_id: "chat-to-proposal",
        actor: "Soma",
        user: "owner",
        action: "proposal_generated",
        timestamp: "2026-03-26T15:58:00Z",
        capability_used: "write_file",
        result_status: "pending",
        approval_status: "approval_required",
        approval_reason: "capability_risk",
        run_id: "run-qa-1",
        intent_proof_id: "proof-qa-1",
        resource: "workspace/logs/hello_world.py",
    },
];

export type ChatRequestBody = {
    messages?: Array<{
        role?: string;
        content?: string;
    }>;
};

export type RouteResponse = {
    status: number;
    body: unknown;
};

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

export function lastUserMessage(requestBody: ChatRequestBody): string {
    const messages = Array.isArray(requestBody.messages) ? requestBody.messages : [];
    return messages[messages.length - 1]?.content ?? "";
}

export function answerEnvelope(
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
                meta: { source_node: "admin", timestamp: "2026-03-26T15:58:00Z" },
                signal_type: "chat.reply",
                trust_score: 0.92,
                template_id: "chat-to-answer",
                mode: "answer",
                payload: {
                    text,
                    tools_used: [],
                    ask_class: options?.askClass,
                    consultations: options?.consultations ?? [],
                    artifacts: options?.artifacts ?? [],
                },
            },
        },
    };
}

export function proposalEnvelope(): RouteResponse {
    return {
        status: 200,
        body: {
            ok: true,
            data: {
                meta: { source_node: "admin", timestamp: "2026-03-26T16:00:00Z" },
                signal_type: "chat.reply",
                trust_score: 0.88,
                template_id: "chat-to-proposal",
                mode: "proposal",
                payload: {
                    text: "I can create that file, but I need your confirmation before execution.",
                    tools_used: ["write_file"],
                    consultations: [],
                    artifacts: [],
                    proposal: {
                        intent: "create_workspace_file",
                        operator_summary: "create a hello_world.py file in your workspace.",
                        expected_result: "A new Python file will be saved to workspace/logs/hello_world.py after approval.",
                        affected_resources: ["workspace/logs/hello_world.py"],
                        teams: 1,
                        agents: 1,
                        tools: ["write_file"],
                        risk_level: "medium",
                        confirm_token: "confirm-qa-1",
                        intent_proof_id: "proof-qa-1",
                        approval: {
                            approval_required: false,
                            approval_reason: "capability_risk",
                            approval_mode: "optional",
                            capability_risk: "medium",
                            capability_ids: ["write_file"],
                            external_data_use: false,
                            estimated_cost: 0,
                        },
                    },
                },
            },
        },
    };
}

async function mockOperatorShell(page: Page) {
    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: { id: "operator-1", name: "Operator", email: "operator@example.test" },
        });
    });

    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: [
                { name: "nats", status: "online" },
                { name: "postgres", status: "online" },
                { name: "reactive", status: "degraded" },
            ],
        });
    });
}

export async function mockOrganizationWorkspace(
    page: Page,
    chatHandler: (requestBody: ChatRequestBody) => RouteResponse,
): Promise<{ cancelCalls: () => number }> {
    let cancelActionCalls = 0;

    await mockOperatorShell(page);

    await page.route(`**/api/v1/organizations/${organizationId}/home`, async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: organizationHome });
    });

    await page.route(`**/api/v1/organizations/${organizationId}/loop-activity`, async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: activityItems });
    });

    await page.route(`**/api/v1/organizations/${organizationId}/automations`, async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: automations });
    });

    await page.route(`**/api/v1/organizations/${organizationId}/learning-insights`, async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: learningInsights });
    });

    await page.route("**/api/v1/chat", async (route) => {
        const requestBody = (route.request().postDataJSON() ?? {}) as ChatRequestBody;
        const response = chatHandler(requestBody);
        await fulfillJSON(route, response.status, response.body);
    });

    await page.route("**/api/v1/intent/cancel-action", async (route) => {
        cancelActionCalls += 1;
        await fulfillJSON(route, 200, { ok: true, data: { cancelled: true } });
    });

    return {
        cancelCalls: () => cancelActionCalls,
    };
}

export async function mockApprovalsAudit(page: Page) {
    await mockOperatorShell(page);

    await page.route("**/api/v1/governance/pending", async (route) => {
        await fulfillJSON(route, 200, []);
    });

    await page.route(/\/api\/v1\/audit(?:\?.*)?$/, async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: auditLog,
        });
    });
}

export async function openOrganization(page: Page) {
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState("domcontentloaded");
    await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
    await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
}

export async function sendWorkspaceMessage(page: Page, content: string) {
    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(content);
    await input.press("Enter");
}
