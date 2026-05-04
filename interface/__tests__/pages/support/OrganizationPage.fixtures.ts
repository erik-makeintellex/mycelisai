import type {
    OrganizationAutomationItem,
    OrganizationHomePayload,
    OrganizationLearningInsightItem,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputTypeId,
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

export const OUTPUT_TYPE_BINDINGS: Array<{ output_type_id: OrganizationOutputTypeId; output_type_label: string }> = [
    { output_type_id: "general_text", output_type_label: "General text" },
    { output_type_id: "research_reasoning", output_type_label: "Research & reasoning" },
    { output_type_id: "code_generation", output_type_label: "Code generation" },
    { output_type_id: "vision_analysis", output_type_label: "Vision analysis" },
];
