import { mockFetch } from "../../setup";
import { useCortexStore } from "@/store/useCortexStore";
import type {
    OrganizationAIEngineProfileId,
    OrganizationHomePayload,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
} from "@/lib/organizations";
import { automations, jsonResponse, learningInsights, organizationHome, outputModelRouting, recentActivity } from "./OrganizationPage.fixtures";
import {
    applyAgentTypeAIEngine,
    applyAgentTypeResponseContract,
    applyOrganizationAIEngineToDepartments,
    applyOutputModelRouting,
    applyResponseContract,
} from "./OrganizationPage.mutations";

export type OrganizationPageTestFetchOptions = {
    homeHandler?: () => Promise<Response>;
    automationsHandler?: () => Promise<Response>;
    loopActivityHandler?: () => Promise<Response>;
    learningInsightsHandler?: () => Promise<Response>;
    actionHandler?: (body: Record<string, unknown>) => Promise<Response>;
    aiEngineUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    departmentAIEngineUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    agentTypeAIEngineUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    agentTypeResponseContractUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    responseContractUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    outputModelRoutingHandler?: (body: Record<string, unknown>) => Promise<Response>;
};

export function resetOrganizationPageStoreState() {
    if (typeof window !== "undefined") {
        window.localStorage.clear();
    }
    useCortexStore.setState({ missionChat: [], isMissionChatting: false, missionChatError: null, missionChatFailure: null, assistantName: "Soma", councilTarget: "admin", councilMembers: [], servicesStatus: [], isFetchingServicesStatus: false, streamConnectionState: "online", isStreamConnected: true });
}

export function setupOrganizationFetch(options?: OrganizationPageTestFetchOptions) {
    let currentOrganizationHome: OrganizationHomePayload = structuredClone(organizationHome);
    const councilMembers = [{ id: "admin", role: "admin", team: "admin-core" }, { id: "council-architect", role: "architect", team: "council-core" }, { id: "council-coder", role: "coder", team: "council-core" }, { id: "council-creative", role: "creative", team: "council-core" }, { id: "council-sentry", role: "sentry", team: "council-core" }];

    mockFetch.mockImplementation((input, init) => {
        const url = typeof input === "string" ? input : input instanceof Request ? input.url : String(input);
        const method = init?.method ?? (input instanceof Request ? input.method : "GET");

        if (url.includes("/api/v1/organizations/org-123/home")) {
            return options?.homeHandler?.() ?? jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/automations")) {
            return options?.automationsHandler?.() ?? jsonResponse({ ok: true, data: automations });
        }

        if (url.includes("/api/v1/organizations/org-123/loop-activity")) {
            return options?.loopActivityHandler?.() ?? jsonResponse({ ok: true, data: recentActivity });
        }

        if (url.includes("/api/v1/organizations/org-123/learning-insights")) {
            return options?.learningInsightsHandler?.() ?? jsonResponse({ ok: true, data: learningInsights });
        }

        if (url.includes("/api/v1/organizations/org-123/output-model-routing") && method === "GET") {
            return jsonResponse({ ok: true, data: outputModelRouting });
        }

        if (url.includes("/api/v1/organizations/org-123/output-model-routing") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.outputModelRoutingHandler) {
                return options.outputModelRoutingHandler(body);
            }

            currentOrganizationHome = applyOutputModelRouting(
                currentOrganizationHome,
                String(body.routing_mode ?? "single_model") as "single_model" | "detected_output_types",
                String(body.default_model_id ?? currentOrganizationHome.default_output_model_id ?? ""),
                Array.isArray(body.bindings) ? (body.bindings as Array<{ output_type_id: OrganizationOutputTypeId; model_id?: string }>) : [],
            );
            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/ai-engine") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.aiEngineUpdateHandler) {
                return options.aiEngineUpdateHandler(body);
            }

            const summaries: Record<string, string> = { starter_defaults: "Starter Defaults", balanced: "Balanced", high_reasoning: "High Reasoning", fast_lightweight: "Fast & Lightweight", deep_planning: "Deep Planning" };
            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = applyOrganizationAIEngineToDepartments(
                currentOrganizationHome,
                profileId as OrganizationAIEngineProfileId,
                summaries[profileId] ?? currentOrganizationHome.ai_engine_settings_summary,
            );

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/response-contract") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.responseContractUpdateHandler) {
                return options.responseContractUpdateHandler(body);
            }

            const summaries: Record<string, string> = { clear_balanced: "Clear & Balanced", structured_analytical: "Structured & Analytical", concise_direct: "Concise & Direct", warm_supportive: "Warm & Supportive" };
            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = applyResponseContract(
                currentOrganizationHome,
                profileId as ResponseContractProfileId,
                summaries[profileId] ?? currentOrganizationHome.response_contract_summary,
            );

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/departments/platform/ai-engine") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.departmentAIEngineUpdateHandler) {
                return options.departmentAIEngineUpdateHandler(body);
            }

            if (body.revert_to_organization_default) {
                currentOrganizationHome = {
                    ...currentOrganizationHome,
                    departments: (currentOrganizationHome.departments ?? []).map((department) =>
                        department.id === "platform"
                            ? {
                                  ...department,
                                  ai_engine_override_profile_id: undefined,
                                  ai_engine_override_summary: undefined,
                                  ai_engine_effective_profile_id: currentOrganizationHome.ai_engine_profile_id,
                                  ai_engine_effective_summary: currentOrganizationHome.ai_engine_settings_summary,
                                  inherits_organization_ai_engine: true,
                                  agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                                      profile.inherits_department_ai_engine
                                          ? {
                                                ...profile,
                                                ai_engine_effective_profile_id: currentOrganizationHome.ai_engine_profile_id,
                                                ai_engine_effective_summary: currentOrganizationHome.ai_engine_settings_summary,
                                            }
                                          : profile,
                                  ),
                              }
                            : department,
                    ),
                };
                return jsonResponse({ ok: true, data: currentOrganizationHome });
            }

            const summaries: Record<string, string> = { starter_defaults: "Starter Defaults", balanced: "Balanced", high_reasoning: "High Reasoning", fast_lightweight: "Fast & Lightweight", deep_planning: "Deep Planning" };
            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = {
                ...currentOrganizationHome,
                departments: (currentOrganizationHome.departments ?? []).map((department) =>
                    department.id === "platform"
                        ? {
                              ...department,
                              ai_engine_override_profile_id: profileId as OrganizationAIEngineProfileId,
                              ai_engine_override_summary: summaries[profileId],
                              ai_engine_effective_profile_id: profileId as OrganizationAIEngineProfileId,
                              ai_engine_effective_summary: summaries[profileId] ?? department.ai_engine_effective_summary,
                              inherits_organization_ai_engine: false,
                              agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                                  profile.inherits_department_ai_engine
                                      ? {
                                            ...profile,
                                            ai_engine_effective_profile_id: profileId as OrganizationAIEngineProfileId,
                                            ai_engine_effective_summary: summaries[profileId] ?? department.ai_engine_effective_summary,
                                        }
                                      : profile,
                              ),
                          }
                        : department,
                ),
            };

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/departments/platform/agent-types/") && url.includes("/ai-engine") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.agentTypeAIEngineUpdateHandler) {
                return options.agentTypeAIEngineUpdateHandler(body);
            }

            const summaries: Record<string, string> = { starter_defaults: "Starter Defaults", balanced: "Balanced", high_reasoning: "High Reasoning", fast_lightweight: "Fast & Lightweight", deep_planning: "Deep Planning" };

            const agentTypeId = url.includes("/delivery-specialist/") ? "delivery-specialist" : "planner";
            if (body.use_team_default) {
                currentOrganizationHome = applyAgentTypeAIEngine(currentOrganizationHome, "platform", agentTypeId, undefined, "");
                return jsonResponse({ ok: true, data: currentOrganizationHome });
            }

            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = applyAgentTypeAIEngine(
                currentOrganizationHome,
                "platform",
                agentTypeId,
                profileId as OrganizationAIEngineProfileId,
                summaries[profileId] ?? "Balanced",
            );

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/departments/platform/agent-types/") && url.includes("/response-contract") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.agentTypeResponseContractUpdateHandler) {
                return options.agentTypeResponseContractUpdateHandler(body);
            }

            const summaries: Record<string, string> = { clear_balanced: "Clear & Balanced", structured_analytical: "Structured & Analytical", concise_direct: "Concise & Direct", warm_supportive: "Warm & Supportive" };

            const agentTypeId = url.includes("/delivery-specialist/") ? "delivery-specialist" : "planner";
            if (body.use_organization_or_team_default) {
                currentOrganizationHome = applyAgentTypeResponseContract(currentOrganizationHome, "platform", agentTypeId, undefined, "");
                return jsonResponse({ ok: true, data: currentOrganizationHome });
            }

            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = applyAgentTypeResponseContract(
                currentOrganizationHome,
                "platform",
                agentTypeId,
                profileId as ResponseContractProfileId,
                summaries[profileId] ?? "Clear & Balanced",
            );

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/workspace/actions") && method === "POST") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            return options?.actionHandler?.(body) ?? jsonResponse({
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
            });
        }

        if (url.includes("/api/v1/council/members")) {
            return jsonResponse({ ok: true, data: councilMembers });
        }

        if (url.includes("/api/v1/chat") && method === "POST") {
            return jsonResponse({
                ok: true,
                data: {
                    meta: { source_node: "admin", timestamp: "2026-03-19T18:00:00Z" },
                    signal_type: "chat_response",
                    trust_score: 0.82,
                    payload: {
                        text: "Soma is ready to shape samples, plans, and delivery guidance for Northstar Labs.",
                        consultations: null,
                        tools_used: null,
                    },
                },
            });
        }

        throw new Error(`Unexpected fetch: ${method} ${url}`);
    });
}
