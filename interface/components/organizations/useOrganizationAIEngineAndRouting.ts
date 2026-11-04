"use client";

import { useEffect, useState } from "react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type {
    OrganizationAIEngineProfileId,
    OrganizationAIEngineUpdateRequest,
    OrganizationHomePayload,
    OrganizationOutputModelRoutingMode,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputModelRoutingUpdateRequest,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
    ResponseContractUpdateRequest,
} from "@/lib/organizations";
import { OUTPUT_TYPE_OPTIONS } from "@/components/organizations/organizationProfileOptions";

export type OrganizationDetailViewId = "advisors" | "departments" | "automations" | "aiEngine" | "responseContract";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export function useOrganizationAIEngineAndRouting({
    organizationId,
    retryToken,
    organization,
    setOrganization,
    activeDetailView,
    setActiveDetailView,
}: {
    organizationId: string;
    retryToken: number;
    organization: OrganizationHomePayload | null;
    setOrganization: React.Dispatch<React.SetStateAction<OrganizationHomePayload | null>>;
    activeDetailView: OrganizationDetailViewId | null;
    setActiveDetailView: React.Dispatch<React.SetStateAction<OrganizationDetailViewId | null>>;
}) {
    const [isAIEngineSelectorOpen, setIsAIEngineSelectorOpen] = useState(false);
    const [selectedAIEngineProfile, setSelectedAIEngineProfile] = useState<OrganizationAIEngineProfileId | null>(null);
    const [aiEngineUpdatePending, setAIEngineUpdatePending] = useState(false);
    const [aiEngineUpdateError, setAIEngineUpdateError] = useState<string | null>(null);
    const [isResponseContractSelectorOpen, setIsResponseContractSelectorOpen] = useState(false);
    const [selectedResponseContractProfile, setSelectedResponseContractProfile] = useState<ResponseContractProfileId | null>(null);
    const [responseContractUpdatePending, setResponseContractUpdatePending] = useState(false);
    const [responseContractUpdateError, setResponseContractUpdateError] = useState<string | null>(null);
    const [outputModelRouting, setOutputModelRouting] = useState<OrganizationOutputModelRoutingPayload | null>(null);
    const [outputModelRoutingLoading, setOutputModelRoutingLoading] = useState(true);
    const [outputModelRoutingError, setOutputModelRoutingError] = useState<string | null>(null);
    const [selectedOutputModelRoutingMode, setSelectedOutputModelRoutingMode] = useState<OrganizationOutputModelRoutingMode>("single_model");
    const [selectedDefaultOutputModelId, setSelectedDefaultOutputModelId] = useState<string>("");
    const [selectedOutputModelBindings, setSelectedOutputModelBindings] = useState<Record<string, string>>({});
    const [outputModelRoutingUpdatePending, setOutputModelRoutingUpdatePending] = useState(false);
    const [outputModelRoutingUpdateError, setOutputModelRoutingUpdateError] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;
        const loadOutputModelRouting = async () => {
            setOutputModelRoutingLoading(true);
            setOutputModelRoutingError(null);
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/output-model-routing`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Unable to load output model routing.");
                }
                if (cancelled) return;
                const data = extractApiData<OrganizationOutputModelRoutingPayload>(payload);
                setOutputModelRouting(data);
                setSelectedOutputModelRoutingMode((data.routing_mode ?? "single_model") as OrganizationOutputModelRoutingMode);
                setSelectedDefaultOutputModelId(data.default_model_id ?? "");
                setSelectedOutputModelBindings(
                    Object.fromEntries((data.bindings ?? []).map((b) => [b.output_type_id, b.model_id ?? data.default_model_id ?? ""])),
                );
            } catch (err) {
                if (!cancelled) {
                    setOutputModelRoutingError(err instanceof Error ? err.message : "Unable to load output model routing.");
                }
            } finally {
                if (!cancelled) setOutputModelRoutingLoading(false);
            }
        };
        void loadOutputModelRouting();
        return () => {
            cancelled = true;
        };
    }, [organizationId, retryToken]);

    useEffect(() => {
        if (activeDetailView !== "aiEngine") {
            setIsAIEngineSelectorOpen(false);
            setAIEngineUpdateError(null);
            setAIEngineUpdatePending(false);
            setOutputModelRoutingUpdatePending(false);
            setOutputModelRoutingUpdateError(null);
        }
        if (activeDetailView !== "responseContract") {
            setIsResponseContractSelectorOpen(false);
            setSelectedResponseContractProfile(null);
            setResponseContractUpdatePending(false);
            setResponseContractUpdateError(null);
        }
    }, [activeDetailView]);

    const openAIEngineSelector = () => {
        if (!organization) return;
        setActiveDetailView("aiEngine");
        setSelectedAIEngineProfile((organization.ai_engine_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setAIEngineUpdateError(null);
        if (outputModelRouting) {
            setSelectedOutputModelRoutingMode(outputModelRouting.routing_mode);
            setSelectedDefaultOutputModelId(outputModelRouting.default_model_id ?? "");
            setSelectedOutputModelBindings(
                Object.fromEntries((outputModelRouting.bindings ?? []).map((b) => [b.output_type_id, b.model_id ?? outputModelRouting.default_model_id ?? ""])),
            );
        }
        setOutputModelRoutingUpdateError(null);
        setIsAIEngineSelectorOpen(true);
    };

    const submitAIEngineSelection = async () => {
        if (!organization || !selectedAIEngineProfile || aiEngineUpdatePending) return;
        setAIEngineUpdatePending(true);
        setAIEngineUpdateError(null);
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedAIEngineProfile } satisfies OrganizationAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || "Unable to update AI Engine Settings.");
            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setSelectedAIEngineProfile((updated.ai_engine_profile_id as OrganizationAIEngineProfileId | undefined) ?? selectedAIEngineProfile);
            setIsAIEngineSelectorOpen(false);
        } catch (err) {
            setAIEngineUpdateError(err instanceof Error ? err.message : "Unable to update AI Engine Settings.");
        } finally {
            setAIEngineUpdatePending(false);
        }
    };

    const submitOutputModelRouting = async () => {
        if (!organization || !selectedDefaultOutputModelId || outputModelRoutingUpdatePending) return;
        setOutputModelRoutingUpdatePending(true);
        setOutputModelRoutingUpdateError(null);
        const bindings = OUTPUT_TYPE_OPTIONS.map((option) => {
            const modelId = selectedOutputModelBindings[option.id] || selectedDefaultOutputModelId;
            return { output_type_id: option.id, model_id: modelId, use_organization_default: modelId === selectedDefaultOutputModelId };
        });
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/output-model-routing`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    routing_mode: selectedOutputModelRoutingMode,
                    default_model_id: selectedDefaultOutputModelId,
                    bindings,
                } satisfies OrganizationOutputModelRoutingUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || "Unable to update output model routing.");
            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setOutputModelRouting((current) =>
                current
                    ? {
                          ...current,
                          routing_mode: selectedOutputModelRoutingMode,
                          default_model_id: updated.default_output_model_id,
                          default_model_summary: updated.default_output_model_summary ?? current.default_model_summary,
                          bindings: updated.output_model_bindings ?? current.bindings,
                      }
                    : current,
            );
        } catch (err) {
            setOutputModelRoutingUpdateError(err instanceof Error ? err.message : "Unable to update output model routing.");
        } finally {
            setOutputModelRoutingUpdatePending(false);
        }
    };

    const handleDefaultOutputModelChange = (modelId: string) => {
        setSelectedOutputModelBindings((current) => {
            const next: Record<string, string> = {};
            for (const option of OUTPUT_TYPE_OPTIONS) {
                const currentValue = current[option.id];
                next[option.id] = currentValue === selectedDefaultOutputModelId || !currentValue ? modelId : currentValue;
            }
            return next;
        });
        setSelectedDefaultOutputModelId(modelId);
    };

    const openResponseContractSelector = () => {
        if (!organization) return;
        setActiveDetailView("responseContract");
        setSelectedResponseContractProfile((organization.response_contract_profile_id as ResponseContractProfileId | undefined) ?? "clear_balanced");
        setResponseContractUpdateError(null);
        setIsResponseContractSelectorOpen(true);
    };

    const submitResponseContractSelection = async () => {
        if (!organization || !selectedResponseContractProfile || responseContractUpdatePending) return;
        setResponseContractUpdatePending(true);
        setResponseContractUpdateError(null);
        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/response-contract`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedResponseContractProfile } satisfies ResponseContractUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) throw new Error(extractApiError(payload) || "Unable to update Response Style.");
            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setSelectedResponseContractProfile((updated.response_contract_profile_id as ResponseContractProfileId | undefined) ?? selectedResponseContractProfile);
            setIsResponseContractSelectorOpen(false);
        } catch (err) {
            setResponseContractUpdateError(err instanceof Error ? err.message : "Unable to update Response Style.");
        } finally {
            setResponseContractUpdatePending(false);
        }
    };

    return {
        isAIEngineSelectorOpen,
        selectedAIEngineProfile,
        aiEngineUpdatePending,
        aiEngineUpdateError,
        outputModelRouting,
        outputModelRoutingLoading,
        outputModelRoutingError,
        selectedOutputModelRoutingMode,
        selectedDefaultOutputModelId,
        selectedOutputModelBindings,
        outputModelRoutingUpdatePending,
        outputModelRoutingUpdateError,
        onAIEngineProfileSelect: setSelectedAIEngineProfile,
        onOutputModelRoutingModeChange: setSelectedOutputModelRoutingMode,
        onDefaultOutputModelChange: handleDefaultOutputModelChange,
        onOutputModelBindingChange: (outputTypeId: OrganizationOutputTypeId, modelId: string) =>
            setSelectedOutputModelBindings((current) => ({ ...current, [outputTypeId]: modelId })),
        onOpenAIEngineSelector: openAIEngineSelector,
        onCloseAIEngineSelector: () => {
            setIsAIEngineSelectorOpen(false);
            setAIEngineUpdateError(null);
        },
        onSubmitAIEngineSelection: submitAIEngineSelection,
        onSubmitOutputModelRouting: submitOutputModelRouting,
        isResponseContractSelectorOpen,
        selectedResponseContractProfile,
        responseContractUpdatePending,
        responseContractUpdateError,
        onResponseContractProfileSelect: setSelectedResponseContractProfile,
        onOpenResponseContractSelector: openResponseContractSelector,
        onCloseResponseContractSelector: () => {
            setIsResponseContractSelectorOpen(false);
            setResponseContractUpdateError(null);
        },
        onSubmitResponseContractSelection: submitResponseContractSelection,
    };
}

export type OrganizationAIEngineAndRoutingControls = ReturnType<typeof useOrganizationAIEngineAndRouting>;
