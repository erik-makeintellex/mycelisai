import { expect, test, type Page } from "@playwright/test";

const organizationId = "org-ui-testing-agent";
const chatPlaceholder = /Tell .* what you want to plan, review, create, or execute/i;

const organizationHome = {
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

function answerEnvelope(text: string): RouteResponse {
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
                    consultations: [],
                    artifacts: [],
                },
            },
        },
    };
}

function proposalEnvelope(): RouteResponse {
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

async function fulfillJSON(route: { fulfill: (options: { status: number; contentType: string; body: string }) => Promise<void> }, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

async function mockOrganizationWorkspace(
    page: Page,
    chatHandler: (requestBody: ChatRequestBody) => RouteResponse,
): Promise<{ cancelCalls: () => number }> {
    let cancelActionCalls = 0;

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
                { name: "reactive", status: "degraded" },
            ],
        });
    });

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

async function mockApprovalsAudit(page: Page) {
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
                { name: "reactive", status: "degraded" },
            ],
        });
    });

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

function lastUserMessage(requestBody: ChatRequestBody): string {
    const messages = Array.isArray(requestBody.messages) ? requestBody.messages : [];
    return messages[messages.length - 1]?.content ?? "";
}

async function openOrganization(page: Page) {
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState("domcontentloaded");
    await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
    await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
}

async function sendWorkspaceMessage(page: Page, content: string) {
    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(content);
    await input.press("Enter");
}

test.describe("V8 UI testing agentry product contract", () => {
    test("keeps Soma primary, preserves continuity on reload, and contains oversized markdown output", async ({ page }) => {
        const directAnswer =
            "Workspace V8 keeps Soma at the center of the AI Organization while recent activity, retained knowledge, and quick checks explain what changed and why.";
        const hugeTable = [
            "| C1 | C2 | C3 | C4 | C5 | C6 | C7 | C8 | C9 | C10 | C11 | C12 | C13 | C14 | C15 | C16 | C17 | C18 | C19 | C20 |",
            "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |",
            "| v1 | v2 | v3 | v4 | v5 | v6 | v7 | v8 | v9 | v10 | v11 | v12 | v13 | v14 | v15 | v16 | v17 | v18 | v19 | v20 |",
        ].join("\n");

        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/huge markdown table/i.test(content)) {
                return answerEnvelope(hugeTable);
            }
            return answerEnvelope(directAnswer);
        });

        await openOrganization(page);
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();

        await sendWorkspaceMessage(page, "Summarize the current Workspace V8 design objectives.");
        await expect(page.getByText(directAnswer)).toBeVisible({ timeout: 20_000 });

        await page.reload({ waitUntil: "domcontentloaded" });
        await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
        await expect(page.getByText("Summarize the current Workspace V8 design objectives.")).toBeVisible();
        await expect(page.getByText(directAnswer)).toBeVisible();

        await sendWorkspaceMessage(page, "Generate a huge markdown table with 20 columns.");
        const table = page.locator('[data-testid="mission-chat"] table').last();
        await expect(table).toBeVisible({ timeout: 20_000 });
        await expect
            .poll(async () => table.evaluate((node) => node.parentElement?.className ?? ""))
            .toContain("overflow-x-auto");

        const body = await page.content();
        expect(body).not.toContain("bg-white");
    });

    test("silently retries a first transient Soma failure and recovers without surfacing a blocker", async ({ page }) => {
        let attempts = 0;

        await mockOrganizationWorkspace(page, () => {
            attempts += 1;
            if (attempts === 1) {
                return {
                    status: 500,
                    body: {
                        ok: false,
                        error: "Soma chat unreachable (500)",
                    },
                };
            }
            return answerEnvelope("Recovered answer after startup wobble.");
        });

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Summarize the current Workspace V8 design objectives.");

        await expect(page.getByText("Recovered answer after startup wobble.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Soma Chat Blocked/i)).toHaveCount(0);
        await expect(page.getByTestId("mission-chat").getByRole("button", { name: /^Retry$/i })).toHaveCount(0);
        expect(attempts).toBe(2);
    });

    test("routes mutating requests through proposal mode and keeps cancel explicit", async ({ page }) => {
        const workspace = await mockOrganizationWorkspace(page, () => proposalEnvelope());

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Create a simple python file named hello_world.py in the workspace.");

        await expect(page.getByText("PROPOSED ACTION")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Approval optional for this action")).toBeVisible();
        await expect(page.getByText(/Capability MEDIUM/i)).toBeVisible();

        await page.getByRole("button", { name: /^Cancel$/i }).click();
        await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 20_000 });
        await expect.poll(() => workspace.cancelCalls()).toBe(1);
    });

    test("shows inspect-only audit activity in Automations approvals", async ({ page }) => {
        await mockApprovalsAudit(page);

        await page.goto("/automations?tab=approvals");
        await page.waitForLoadState("domcontentloaded");
        await expect(page.getByRole("button", { name: "Approvals" })).toBeVisible({ timeout: 20_000 });
        await page.getByRole("button", { name: "Audit" }).click();

        await expect(page.getByText("Activity Log")).toBeVisible();
        await expect(page.getByText(/Inspect recent approvals, execution outcomes, capability use/i)).toBeVisible();
        await expect(page.getByText("proposal generated")).toBeVisible();
        await expect(page.getByText("approval required")).toBeVisible();
        await expect(page.getByText("Capability: write_file")).toBeVisible();
        await expect(page.getByText("Resource: workspace/logs/hello_world.py")).toBeVisible();
    });
});
