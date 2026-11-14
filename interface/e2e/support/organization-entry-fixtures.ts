import { type Page, type Route } from "@playwright/test";

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

export type MockOrganizationEntryOptions = {
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

export const defaultOperatorUserPayload = {
    ok: true,
    data: {
        id: "operator-1",
        name: "Operator",
        email: "operator@example.test",
    },
};

export const defaultServicesStatusPayload = {
    ok: true,
    data: [
        { name: "core", status: "ready" },
        { name: "frontend", status: "ready" },
    ],
};

export function buildDefaultWorkspaceActionResponse(requestBody: Record<string, unknown>) {
    return {
        status: 200,
        body: {
            ok: true,
            data: {
                action: requestBody.action ?? "plan_next_steps",
                request_label: "Run a quick strategy check",
                headline: "Soma plan for Northstar Labs",
                summary: "Soma recommends a clear next move for Northstar Labs.",
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
}

export function buildDefaultChatResponse() {
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
                    text: "Soma is ready to help inside this AI Organization.",
                    ask_class: "direct_answer",
                    consultations: [],
                    tools_used: [],
                    artifacts: [],
                },
            },
        },
    };
}

export async function fulfillJson(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

export async function mockBaseStatusRoutes(page: Page) {
    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJson(route, 200, defaultOperatorUserPayload);
    });

    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJson(route, 200, defaultServicesStatusPayload);
    });
}
