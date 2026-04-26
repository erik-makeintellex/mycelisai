import type {
    OrganizationAgentTypeProfileSummary,
    OrganizationAIEngineProfileId,
    OrganizationDepartmentSummary,
    OrganizationHomePayload,
} from "@/lib/organizations";

export function advisorDetailItems(count: number) {
    if (count === 0) {
        return [];
    }

    return [
        {
            name: "Planning Advisor",
            purpose: "Helps Soma and the Team Lead test priorities, sequence work, and keep the first plan practical.",
            supportCue: "Best when the operator wants a second look at the next move.",
        },
        {
            name: "Delivery Advisor",
            purpose: "Keeps an eye on progress, blockers, and whether work is ready to move forward.",
            supportCue: "Useful when delivery momentum needs a quick review.",
        },
        {
            name: "Decision Advisor",
            purpose: "Supports tradeoffs, review points, and operator-facing choices before work expands.",
            supportCue: "Useful when Soma needs a clear go/no-go perspective from the Team Lead layer.",
        },
    ].slice(0, Math.min(count, 3));
}

export function departmentDetailItems(organization: OrganizationHomePayload) {
    const departments = organization.departments?.length
        ? organization.departments
        : generateFallbackDepartmentSummaries(organization.department_count, organization.specialist_count, organization.ai_engine_profile_id, organization.ai_engine_settings_summary);

    return departments.map((department) => ({
        ...department,
        purpose: departmentPurpose(department.name),
        aiEngineStateLabel: department.inherits_organization_ai_engine
            ? `Using Organization Default: ${department.ai_engine_effective_summary}`
            : `Overridden: ${department.ai_engine_effective_summary}`,
        agentTypeProfiles: department.agent_type_profiles ?? [],
    }));
}

export function departmentPurpose(name: string) {
    if (name.includes("Planning")) {
        return "Shapes the first approach, breaks work into practical steps, and keeps priorities aligned.";
    }
    if (name.includes("Operations")) {
        return "Keeps follow-through organized, reduces friction, and supports steady execution.";
    }
    if (name.includes("Support")) {
        return "Handles supporting work that helps the main delivery lane stay clear and focused.";
    }
    return "Carries the main delivery lane so Soma can move from planning into execution through the operational team.";
}

export function agentTypeAIEngineSourceLabel(profile: OrganizationAgentTypeProfileSummary) {
    return profile.inherits_department_ai_engine ? `Using Team Default: ${profile.ai_engine_effective_summary}` : `Type-specific Engine: ${profile.ai_engine_effective_summary}`;
}

export function agentTypeResponseStyleSourceLabel(profile: OrganizationAgentTypeProfileSummary) {
    return profile.inherits_default_response_contract
        ? `Using Organization or Team Default: ${profile.response_contract_effective_summary}`
        : `Type-specific Response Style: ${profile.response_contract_effective_summary}`;
}

export function agentTypeOutputModelSourceLabel(profile: OrganizationAgentTypeProfileSummary) {
    const summary = profile.output_model_effective_summary?.trim() || "Set up later in Advanced mode";
    return profile.inherits_default_output_model ? `Using Organization Default: ${summary}` : `Detected for this role: ${summary}`;
}

export function agentTypeSelectionKey(departmentId: string, agentTypeId: string) {
    return `${departmentId}:${agentTypeId}`;
}

function spreadSpecialists(total: number, departmentCount: number, index: number) {
    if (departmentCount <= 0 || total <= 0) {
        return 0;
    }
    const base = Math.floor(total / departmentCount);
    const remainder = total % departmentCount;
    return base + (index < remainder ? 1 : 0);
}

export function generateFallbackDepartmentSummaries(
    departmentCount: number,
    specialistCount: number,
    aiEngineProfileId: OrganizationAIEngineProfileId | undefined,
    aiEngineSummary: string,
): OrganizationDepartmentSummary[] {
    if (departmentCount <= 0) {
        return [];
    }

    const names = departmentCount === 1 ? ["Core Delivery Department"] : ["Planning Department", "Delivery Department", "Operations Department", "Support Department"];
    return Array.from({ length: departmentCount }, (_, index) => {
        const name = names[index] ?? `Department ${index + 1}`;
        return {
            id: slugifyDepartmentId(name, index),
            name,
            specialist_count: spreadSpecialists(specialistCount, departmentCount, index),
            ai_engine_effective_profile_id: aiEngineProfileId,
            ai_engine_effective_summary: aiEngineSummary || "Set up later in Advanced mode",
            inherits_organization_ai_engine: true,
        };
    });
}

function slugifyDepartmentId(name: string, index: number) {
    const slug = name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
    return slug || `department-${index + 1}`;
}

export function aiEngineDetailItems(organization: OrganizationHomePayload) {
    const hasOrganizationLevelSelection = Boolean(organization.ai_engine_profile_id && organization.ai_engine_settings_summary.trim());

    return [
        {
            name: "Organization-wide AI engine",
            purpose:
                !hasOrganizationLevelSelection || organization.ai_engine_settings_summary.trim() === "Set up later in Advanced mode"
                    ? "No organization-wide AI engine has been chosen yet. Pick one here when you want to tune how Soma plans and responds."
                    : `Current profile: ${organization.ai_engine_settings_summary}.`,
            supportCue: "Affects the overall response style, planning depth, and how work is carried across the organization.",
        },
        {
            name: "Team defaults",
            purpose:
                hasOrganizationLevelSelection
                    ? "Departments start from the organization-wide AI engine unless a team-specific setting appears here."
                    : "Departments will follow the organization-wide AI engine after one is chosen here.",
            supportCue: "Affects how each Department begins its work before any more specific assignment takes over.",
        },
        {
            name: "Specific role overrides",
            purpose:
                organization.specialist_count > 0 || organization.advisor_count > 0
                    ? "No specific role overrides are visible in this workspace right now."
                    : "No specific role overrides are visible because the organization is still on a simple starter setup.",
            supportCue: "Affects a single Team Lead, Advisor, or Specialist only when a scoped override is present.",
        },
    ];
}
