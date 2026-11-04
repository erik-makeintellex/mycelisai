"use client";

import { useEffect, useState } from "react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type {
    AgentTypeAIEngineUpdateRequest,
    AgentTypeResponseContractUpdateRequest,
    DepartmentAIEngineUpdateRequest,
    OrganizationAgentTypeProfileSummary,
    OrganizationAIEngineProfileId,
    OrganizationDepartmentSummary,
    OrganizationHomePayload,
    ResponseContractProfileId,
} from "@/lib/organizations";
import { agentTypeSelectionKey } from "@/components/organizations/organizationDetailHelpers";
import type { OrganizationDetailViewId } from "@/components/organizations/useOrganizationAIEngineAndRouting";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export function useDepartmentAndAgentTypeOverrides({
    organization,
    setOrganization,
    activeDetailView,
    setActiveDetailView,
}: {
    organization: OrganizationHomePayload | null;
    setOrganization: React.Dispatch<React.SetStateAction<OrganizationHomePayload | null>>;
    activeDetailView: OrganizationDetailViewId | null;
    setActiveDetailView: React.Dispatch<React.SetStateAction<OrganizationDetailViewId | null>>;
}) {
    const [activeDepartmentAIEngineId, setActiveDepartmentAIEngineId] = useState<string | null>(null);
    const [selectedDepartmentAIEngineProfile, setSelectedDepartmentAIEngineProfile] = useState<OrganizationAIEngineProfileId | null>(null);
    const [departmentAIEngineUpdatePendingId, setDepartmentAIEngineUpdatePendingId] = useState<string | null>(null);
    const [departmentAIEngineUpdateError, setDepartmentAIEngineUpdateError] = useState<{ departmentId: string; message: string } | null>(null);
    const [activeAgentTypeAIEngineKey, setActiveAgentTypeAIEngineKey] = useState<string | null>(null);
    const [selectedAgentTypeAIEngineProfile, setSelectedAgentTypeAIEngineProfile] = useState<OrganizationAIEngineProfileId | null>(null);
    const [agentTypeAIEngineUpdatePendingKey, setAgentTypeAIEngineUpdatePendingKey] = useState<string | null>(null);
    const [agentTypeAIEngineUpdateError, setAgentTypeAIEngineUpdateError] = useState<{ departmentId: string; agentTypeId: string; message: string } | null>(null);
    const [activeAgentTypeResponseContractKey, setActiveAgentTypeResponseContractKey] = useState<string | null>(null);
    const [selectedAgentTypeResponseContractProfile, setSelectedAgentTypeResponseContractProfile] = useState<ResponseContractProfileId | null>(null);
    const [agentTypeResponseContractUpdatePendingKey, setAgentTypeResponseContractUpdatePendingKey] = useState<string | null>(null);
    const [agentTypeResponseContractUpdateError, setAgentTypeResponseContractUpdateError] = useState<{ departmentId: string; agentTypeId: string; message: string } | null>(null);

    useEffect(() => {
        if (activeDetailView !== "departments") {
            setActiveDepartmentAIEngineId(null);
            setSelectedDepartmentAIEngineProfile(null);
            setDepartmentAIEngineUpdatePendingId(null);
            setDepartmentAIEngineUpdateError(null);
            setActiveAgentTypeAIEngineKey(null);
            setSelectedAgentTypeAIEngineProfile(null);
            setAgentTypeAIEngineUpdatePendingKey(null);
            setAgentTypeAIEngineUpdateError(null);
            setActiveAgentTypeResponseContractKey(null);
            setSelectedAgentTypeResponseContractProfile(null);
            setAgentTypeResponseContractUpdatePendingKey(null);
            setAgentTypeResponseContractUpdateError(null);
        }
    }, [activeDetailView]);

    const openDepartmentAIEngineSelector = (department: OrganizationDepartmentSummary) => {
        setActiveDetailView("departments");
        setActiveDepartmentAIEngineId(department.id);
        setSelectedDepartmentAIEngineProfile((department.ai_engine_effective_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setDepartmentAIEngineUpdateError(null);
    };

    const patchDepartmentAIEngine = async (departmentId: string, body: DepartmentAIEngineUpdateRequest, fallbackError: string) => {
        if (!organization || departmentAIEngineUpdatePendingId) return;
        setDepartmentAIEngineUpdatePendingId(departmentId);
        setDepartmentAIEngineUpdateError(null);
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || fallbackError);
            setOrganization(extractApiData<OrganizationHomePayload>(payload));
            setActiveDepartmentAIEngineId(null);
            setSelectedDepartmentAIEngineProfile(null);
        } catch (err) {
            setDepartmentAIEngineUpdateError({ departmentId, message: err instanceof Error ? err.message : fallbackError });
        } finally {
            setDepartmentAIEngineUpdatePendingId(null);
        }
    };

    const submitDepartmentAIEngineSelection = async (departmentId: string) => {
        if (!selectedDepartmentAIEngineProfile) return;
        await patchDepartmentAIEngine(departmentId, { profile_id: selectedDepartmentAIEngineProfile }, "Unable to update this Department AI Engine.");
    };

    const revertDepartmentAIEngineSelection = async (departmentId: string) => {
        await patchDepartmentAIEngine(departmentId, { revert_to_organization_default: true }, "Unable to return this Department to the organization default.");
    };

    const openAgentTypeAIEngineSelector = (departmentId: string, profile: OrganizationAgentTypeProfileSummary) => {
        setActiveDetailView("departments");
        setActiveAgentTypeAIEngineKey(agentTypeSelectionKey(departmentId, profile.id));
        setSelectedAgentTypeAIEngineProfile((profile.ai_engine_effective_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setAgentTypeAIEngineUpdateError(null);
        setActiveAgentTypeResponseContractKey(null);
        setSelectedAgentTypeResponseContractProfile(null);
        setAgentTypeResponseContractUpdateError(null);
    };

    const patchAgentTypeAIEngine = async (departmentId: string, agentTypeId: string, body: AgentTypeAIEngineUpdateRequest, fallbackError: string) => {
        if (!organization || agentTypeAIEngineUpdatePendingKey) return;
        setAgentTypeAIEngineUpdatePendingKey(agentTypeSelectionKey(departmentId, agentTypeId));
        setAgentTypeAIEngineUpdateError(null);
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || fallbackError);
            setOrganization(extractApiData<OrganizationHomePayload>(payload));
            setActiveAgentTypeAIEngineKey(null);
            setSelectedAgentTypeAIEngineProfile(null);
        } catch (err) {
            setAgentTypeAIEngineUpdateError({ departmentId, agentTypeId, message: err instanceof Error ? err.message : fallbackError });
        } finally {
            setAgentTypeAIEngineUpdatePendingKey(null);
        }
    };

    const submitAgentTypeAIEngineSelection = async (departmentId: string, agentTypeId: string) => {
        if (!selectedAgentTypeAIEngineProfile) return;
        await patchAgentTypeAIEngine(departmentId, agentTypeId, { profile_id: selectedAgentTypeAIEngineProfile }, "Unable to update this Agent Type AI Engine.");
    };

    const revertAgentTypeAIEngineSelection = async (departmentId: string, agentTypeId: string) => {
        await patchAgentTypeAIEngine(departmentId, agentTypeId, { use_team_default: true }, "Unable to return this Agent Type to the Team default.");
    };

    const openAgentTypeResponseContractSelector = (departmentId: string, profile: OrganizationAgentTypeProfileSummary) => {
        setActiveDetailView("departments");
        setActiveAgentTypeResponseContractKey(agentTypeSelectionKey(departmentId, profile.id));
        setSelectedAgentTypeResponseContractProfile((profile.response_contract_effective_profile_id as ResponseContractProfileId | undefined) ?? "clear_balanced");
        setAgentTypeResponseContractUpdateError(null);
        setActiveAgentTypeAIEngineKey(null);
        setSelectedAgentTypeAIEngineProfile(null);
        setAgentTypeAIEngineUpdateError(null);
    };

    const patchAgentTypeResponseContract = async (departmentId: string, agentTypeId: string, body: AgentTypeResponseContractUpdateRequest, fallbackError: string) => {
        if (!organization || agentTypeResponseContractUpdatePendingKey) return;
        setAgentTypeResponseContractUpdatePendingKey(agentTypeSelectionKey(departmentId, agentTypeId));
        setAgentTypeResponseContractUpdateError(null);
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/response-contract`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || fallbackError);
            setOrganization(extractApiData<OrganizationHomePayload>(payload));
            setActiveAgentTypeResponseContractKey(null);
            setSelectedAgentTypeResponseContractProfile(null);
        } catch (err) {
            setAgentTypeResponseContractUpdateError({ departmentId, agentTypeId, message: err instanceof Error ? err.message : fallbackError });
        } finally {
            setAgentTypeResponseContractUpdatePendingKey(null);
        }
    };

    const submitAgentTypeResponseContractSelection = async (departmentId: string, agentTypeId: string) => {
        if (!selectedAgentTypeResponseContractProfile) return;
        await patchAgentTypeResponseContract(departmentId, agentTypeId, { profile_id: selectedAgentTypeResponseContractProfile }, "Unable to update this Agent Type Response Style.");
    };

    const revertAgentTypeResponseContractSelection = async (departmentId: string, agentTypeId: string) => {
        await patchAgentTypeResponseContract(departmentId, agentTypeId, { use_organization_or_team_default: true }, "Unable to return this Agent Type to the Organization / Team default.");
    };

    return {
        activeDepartmentAIEngineId,
        selectedDepartmentAIEngineProfile,
        departmentAIEngineUpdatePendingId,
        departmentAIEngineUpdateError,
        onDepartmentAIEngineProfileSelect: setSelectedDepartmentAIEngineProfile,
        onOpenDepartmentAIEngineSelector: openDepartmentAIEngineSelector,
        onCloseDepartmentAIEngineSelector: () => {
            setActiveDepartmentAIEngineId(null);
            setSelectedDepartmentAIEngineProfile(null);
            setDepartmentAIEngineUpdateError(null);
        },
        onSubmitDepartmentAIEngineSelection: submitDepartmentAIEngineSelection,
        onRevertDepartmentAIEngineSelection: revertDepartmentAIEngineSelection,
        activeAgentTypeAIEngineKey,
        selectedAgentTypeAIEngineProfile,
        agentTypeAIEngineUpdatePendingKey,
        agentTypeAIEngineUpdateError,
        onAgentTypeAIEngineProfileSelect: setSelectedAgentTypeAIEngineProfile,
        onOpenAgentTypeAIEngineSelector: openAgentTypeAIEngineSelector,
        onCloseAgentTypeAIEngineSelector: () => {
            setActiveAgentTypeAIEngineKey(null);
            setSelectedAgentTypeAIEngineProfile(null);
            setAgentTypeAIEngineUpdateError(null);
        },
        onSubmitAgentTypeAIEngineSelection: submitAgentTypeAIEngineSelection,
        onRevertAgentTypeAIEngineSelection: revertAgentTypeAIEngineSelection,
        activeAgentTypeResponseContractKey,
        selectedAgentTypeResponseContractProfile,
        agentTypeResponseContractUpdatePendingKey,
        agentTypeResponseContractUpdateError,
        onAgentTypeResponseContractProfileSelect: setSelectedAgentTypeResponseContractProfile,
        onOpenAgentTypeResponseContractSelector: openAgentTypeResponseContractSelector,
        onCloseAgentTypeResponseContractSelector: () => {
            setActiveAgentTypeResponseContractKey(null);
            setSelectedAgentTypeResponseContractProfile(null);
            setAgentTypeResponseContractUpdateError(null);
        },
        onSubmitAgentTypeResponseContractSelection: submitAgentTypeResponseContractSelection,
        onRevertAgentTypeResponseContractSelection: revertAgentTypeResponseContractSelection,
    };
}

export type DepartmentAndAgentTypeOverrideControls = ReturnType<typeof useDepartmentAndAgentTypeOverrides>;
