import { mockFetch } from "../../setup";
import { useCortexStore } from "@/store/useCortexStore";
import type {
    OrganizationAIEngineProfileId,
    OrganizationAutomationItem,
    OrganizationHomePayload,
    OrganizationLearningInsightItem,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
} from "@/lib/organizations";

export const organizationHome: OrganizationHomePayload = {
    id: "org-123",
    name: "Northstar Labs",
    purpose: "Ship a focused AI engineering organization for product delivery.",
    start_mode: "template",
    template_id: "engineering-starter",
    template_name: "Engineering Starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_profile_id: "starter_defaults",
    ai_engine_settings_summary: "Starter defaults included",
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
    output_model_routing_mode: "single_model",
    default_output_model_id: "qwen2.5-coder:7b-instruct",
    default_output_model_summary: "Qwen2.5 Coder 7B",
    status: "ready",
    description: "Guided AI Organization for engineering work",
    output_model_bindings: [
        { output_type_id: "general_text", output_type_label: "General text", model_id: "qwen3:8b", model_summary: "Qwen3 8B" },
        { output_type_id: "research_reasoning", output_type_label: "Research & reasoning", model_id: "llama3.1:8b", model_summary: "Llama 3.1 8B" },
        { output_type_id: "code_generation", output_type_label: "Code generation", model_id: "qwen2.5-coder:7b", model_summary: "Qwen2.5 Coder 7B" },
        { output_type_id: "vision_analysis", output_type_label: "Vision analysis", model_id: "llava:7b", model_summary: "LLaVA 7B" },
    ],
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
                    output_type_id: "research_reasoning",
                    output_type_label: "Research & reasoning",
                    output_model_effective_id: "qwen2.5-coder:7b-instruct",
                    output_model_effective_summary: "Qwen2.5 Coder 7B",
                    inherits_default_output_model: true,
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
                    output_type_id: "code_generation",
                    output_type_label: "Code generation",
                    output_model_effective_id: "qwen2.5-coder:7b-instruct",
                    output_model_effective_summary: "Qwen2.5 Coder 7B",
                    inherits_default_output_model: true,
                },
            ],
        },
    ],
};

export const outputModelRouting: OrganizationOutputModelRoutingPayload = {
    routing_mode: "single_model",
    default_model_id: "qwen2.5-coder:7b-instruct",
    default_model_summary: "Qwen2.5 Coder 7B",
    hardware_summary: "Local-first self-hosted posture tuned for the current Ollama inventory and a 16GB-class GPU host.",
    review_permission_prompt: "Ask the owner/admin before Soma reviews potential model behavior for a requested output or changes saved routing.",
    automatic_selection_criteria: [
        "Prefer an installed self-hosted model that declares fit for the detected output type before suggesting a pull or remote provider.",
        "Keep the operator in control: ask for owner approval before running a model-behavior review or changing the organization's saved routing policy.",
    ],
    bindings: organizationHome.output_model_bindings,
    recommended_models: [
        {
            model_id: "qwen3:8b",
            label: "Qwen3 8B",
            summary: "Strong local-first default for general text, agent planning, and multi-step reasoning.",
            installed: true,
            popular: true,
            self_hostable: true,
            hosting_fit: "Fits well on the current self-hosted GPU class and is already a common local-first general model.",
        },
        {
            model_id: "llama3.1:8b",
            label: "Llama 3.1 8B",
            summary: "Popular local general model with long context and strong multilingual/research-oriented posture.",
            installed: true,
            popular: true,
            self_hostable: true,
            hosting_fit: "Fits well on the current self-hosted GPU class and gives a strong second general-purpose local option.",
        },
    ],
    available_models: [
        {
            model_id: "qwen3:8b",
            label: "Qwen3 8B",
            summary: "Strong local-first default for general text, agent planning, and multi-step reasoning.",
            installed: true,
            popular: true,
            self_hostable: true,
        },
        {
            model_id: "llama3.1:8b",
            label: "Llama 3.1 8B",
            summary: "Popular local general model with long context and strong multilingual/research-oriented posture.",
            installed: true,
            popular: true,
            self_hostable: true,
        },
        {
            model_id: "qwen2.5-coder:7b",
            label: "Qwen2.5 Coder 7B",
            summary: "Focused local model for code generation, code repair, and implementation-heavy team lanes.",
            installed: true,
            popular: false,
            self_hostable: true,
        },
        {
            model_id: "llava:7b",
            label: "LLaVA 7B",
            summary: "Local multimodal model for image understanding, OCR, and visual review work.",
            installed: true,
            popular: false,
            self_hostable: true,
        },
    ],
    review_candidates: [
        {
            output_type_id: "general_text",
            output_type_label: "General text",
            model_id: "qwen3:8b",
            model_summary: "Qwen3 8B",
            installed: true,
            review_criteria: ["prioritize readable direct answers, broad instruction following, and low-friction drafting"],
        },
        {
            output_type_id: "research_reasoning",
            output_type_label: "Research & reasoning",
            model_id: "qwen3:8b",
            model_summary: "Qwen3 8B",
            installed: true,
            review_criteria: ["prioritize planning depth, synthesis quality, and long-context behavior"],
        },
        {
            output_type_id: "code_generation",
            output_type_label: "Code generation",
            model_id: "qwen2.5-coder:7b",
            model_summary: "Qwen2.5 Coder 7B",
            installed: true,
            review_criteria: ["prioritize implementation accuracy, test repair, and structured code output"],
        },
        {
            output_type_id: "vision_analysis",
            output_type_label: "Vision analysis",
            model_id: "llava:7b",
            model_summary: "LLaVA 7B",
            installed: true,
            review_criteria: ["prioritize multimodal image understanding, OCR, and visual review reliability"],
        },
    ],
};

export const recentActivity = [
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
];

export const automations: OrganizationAutomationItem[] = [
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
];

export const learningInsights: OrganizationLearningInsightItem[] = [
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
];

export function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), {
        status,
        headers: {
            "Content-Type": "application/json",
        },
    }));
}

export function applyOrganizationAIEngineToDepartments(
    home: OrganizationHomePayload,
    profileId: OrganizationAIEngineProfileId | undefined,
    summary: string,
) {
    return {
        ...home,
        ai_engine_profile_id: profileId,
        ai_engine_settings_summary: summary,
        departments: (home.departments ?? []).map((department) => ({
            ...department,
            ...(department.inherits_organization_ai_engine
                ? {
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
                : {}),
        })),
    };
}

export function applyResponseContract(
    home: OrganizationHomePayload,
    profileId: ResponseContractProfileId | undefined,
    summary: string,
) {
    return {
        ...home,
        response_contract_profile_id: profileId,
        response_contract_summary: summary,
        departments: (home.departments ?? []).map((department) => ({
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

export function applyAgentTypeAIEngine(
    home: OrganizationHomePayload,
    departmentId: string,
    agentTypeId: string,
    profileId: OrganizationAIEngineProfileId | undefined,
    summary: string,
) {
    return {
        ...home,
        departments: (home.departments ?? []).map((department) =>
            department.id !== departmentId
                ? department
                : {
                      ...department,
                      agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                          profile.id !== agentTypeId
                              ? profile
                              : {
                                    ...profile,
                                    ai_engine_binding_profile_id: profileId,
                                    ai_engine_effective_profile_id: profileId ?? department.ai_engine_effective_profile_id,
                                    ai_engine_effective_summary: profileId ? summary : department.ai_engine_effective_summary,
                                    inherits_department_ai_engine: profileId === undefined,
                                },
                      ),
                  },
        ),
    };
}

export function applyAgentTypeResponseContract(
    home: OrganizationHomePayload,
    departmentId: string,
    agentTypeId: string,
    profileId: ResponseContractProfileId | undefined,
    summary: string,
) {
    return {
        ...home,
        departments: (home.departments ?? []).map((department) =>
            department.id !== departmentId
                ? department
                : {
                      ...department,
                      agent_type_profiles: department.agent_type_profiles?.map((profile) =>
                          profile.id !== agentTypeId
                              ? profile
                              : {
                                    ...profile,
                                    response_contract_binding_profile_id: profileId,
                                    response_contract_effective_profile_id: profileId ?? home.response_contract_profile_id,
                                    response_contract_effective_summary: profileId ? summary : home.response_contract_summary,
                                    inherits_default_response_contract: profileId === undefined,
                                },
                      ),
                  },
        ),
    };
}

export function applyOutputModelRouting(
    home: OrganizationHomePayload,
    routingMode: "single_model" | "detected_output_types",
    defaultModelId: string,
    bindings: Array<{ output_type_id: OrganizationOutputTypeId; model_id?: string }>,
) {
    const bindingMap = new Map(bindings.map((binding) => [binding.output_type_id, binding.model_id ?? defaultModelId]));
    const summaryForModel = (modelId: string) => {
        const summaries: Record<string, string> = {
            "qwen3:8b": "Qwen3 8B",
            "llama3.1:8b": "Llama 3.1 8B",
            "qwen2.5-coder:7b": "Qwen2.5 Coder 7B",
            "qwen2.5-coder:7b-instruct": "Qwen2.5 Coder 7B",
            "llava:7b": "LLaVA 7B",
        };
        return summaries[modelId] ?? modelId;
    };

    return {
        ...home,
        output_model_routing_mode: routingMode,
        default_output_model_id: defaultModelId,
        default_output_model_summary: summaryForModel(defaultModelId),
        output_model_bindings: OUTPUT_TYPE_BINDINGS.map((binding) => {
            const modelId = bindingMap.get(binding.output_type_id) ?? defaultModelId;
            return {
                ...binding,
                model_id: modelId,
                model_summary: summaryForModel(modelId),
            };
        }),
        departments: (home.departments ?? []).map((department) => ({
            ...department,
            agent_type_profiles: department.agent_type_profiles?.map((profile) => {
                const outputTypeId = profile.output_type_id ?? "general_text";
                const effectiveModelId =
                    routingMode === "detected_output_types" ? (bindingMap.get(outputTypeId) ?? defaultModelId) : defaultModelId;
                return {
                    ...profile,
                    output_model_effective_id: effectiveModelId,
                    output_model_effective_summary: summaryForModel(effectiveModelId),
                    inherits_default_output_model: effectiveModelId === defaultModelId,
                };
            }),
        })),
    };
}

export const OUTPUT_TYPE_BINDINGS: Array<{ output_type_id: OrganizationOutputTypeId; output_type_label: string }> = [
    { output_type_id: "general_text", output_type_label: "General text" },
    { output_type_id: "research_reasoning", output_type_label: "Research & reasoning" },
    { output_type_id: "code_generation", output_type_label: "Code generation" },
    { output_type_id: "vision_analysis", output_type_label: "Vision analysis" },
];

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
    useCortexStore.setState({
        missionChat: [],
        isMissionChatting: false,
        missionChatError: null,
        missionChatFailure: null,
        assistantName: "Soma",
        councilTarget: "admin",
        councilMembers: [],
        servicesStatus: [],
        isFetchingServicesStatus: false,
        streamConnectionState: "online",
        isStreamConnected: true,
    });
}

export function setupOrganizationFetch(options?: OrganizationPageTestFetchOptions) {
    let currentOrganizationHome: OrganizationHomePayload = structuredClone(organizationHome);
    const councilMembers = [
        { id: "admin", role: "admin", team: "admin-core" },
        { id: "council-architect", role: "architect", team: "council-core" },
        { id: "council-coder", role: "coder", team: "council-core" },
        { id: "council-creative", role: "creative", team: "council-core" },
        { id: "council-sentry", role: "sentry", team: "council-core" },
    ];

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

            const summaries: Record<string, string> = {
                starter_defaults: "Starter Defaults",
                balanced: "Balanced",
                high_reasoning: "High Reasoning",
                fast_lightweight: "Fast & Lightweight",
                deep_planning: "Deep Planning",
            };
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

            const summaries: Record<string, string> = {
                clear_balanced: "Clear & Balanced",
                structured_analytical: "Structured & Analytical",
                concise_direct: "Concise & Direct",
                warm_supportive: "Warm & Supportive",
            };
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

            const summaries: Record<string, string> = {
                starter_defaults: "Starter Defaults",
                balanced: "Balanced",
                high_reasoning: "High Reasoning",
                fast_lightweight: "Fast & Lightweight",
                deep_planning: "Deep Planning",
            };
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

            const summaries: Record<string, string> = {
                starter_defaults: "Starter Defaults",
                balanced: "Balanced",
                high_reasoning: "High Reasoning",
                fast_lightweight: "Fast & Lightweight",
                deep_planning: "Deep Planning",
            };

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

            const summaries: Record<string, string> = {
                clear_balanced: "Clear & Balanced",
                structured_analytical: "Structured & Analytical",
                concise_direct: "Concise & Direct",
                warm_supportive: "Warm & Supportive",
            };

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
