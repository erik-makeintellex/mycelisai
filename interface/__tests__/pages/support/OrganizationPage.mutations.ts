import type {
    OrganizationAIEngineProfileId,
    OrganizationHomePayload,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
} from "@/lib/organizations";
import { OUTPUT_TYPE_BINDINGS } from "./OrganizationPage.fixtures";

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

