"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, Blocks, Bot, BrainCircuit, Building2, Loader2, RefreshCcw, Sparkles, Users } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import { rememberLastOrganization } from "@/lib/lastOrganization";
import type {
    AgentTypeAIEngineUpdateRequest,
    AgentTypeResponseContractUpdateRequest,
    DepartmentAIEngineUpdateRequest,
    OrganizationAutomationItem,
    OrganizationAgentTypeProfileSummary,
    OrganizationAIEngineProfileId,
    OrganizationAIEngineUpdateRequest,
    OrganizationDepartmentSummary,
    OrganizationHomePayload,
    OrganizationOutputModelBinding,
    OrganizationOutputModelCatalogEntry,
    OrganizationOutputModelRoutingMode,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputModelRoutingUpdateRequest,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
    ResponseContractUpdateRequest,
} from "@/lib/organizations";
import TeamLeadInteractionPanel, { type SomaGuidanceUpdate } from "@/components/organizations/TeamLeadInteractionPanel";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import SystemQuickChecks from "@/components/system/SystemQuickChecks";
import LaunchCrewModal from "@/components/workspace/LaunchCrewModal";
import { useCortexStore } from "@/store/useCortexStore";
import {
    CausalFact,
    LearningVisibilityPanel,
    RecentActivityPanel,
    SomaCausalStrip,
    type CausalStripState,
} from "@/components/organizations/organizationActivityPanels";
import { AutomationDetailPanel } from "@/components/organizations/organizationAutomationPanel";
import {
    extractOutputsFromConversation,
    findLatestConversationOutcome,
} from "@/components/organizations/conversationOutcomeHelpers";
import {
    advisorDetailItems,
    agentTypeAIEngineSourceLabel,
    agentTypeOutputModelSourceLabel,
    agentTypeResponseStyleSourceLabel,
    agentTypeSelectionKey,
    aiEngineDetailItems,
    departmentDetailItems,
} from "@/components/organizations/organizationDetailHelpers";
import {
    advisorSummary,
    advisorSupportItems,
    aiEngineSummary,
    aiEngineSupportItems,
    automationStatusLabel,
    automationSummary,
    automationSupportItems,
    departmentSummary,
    departmentSupportItems,
    formatAutomationCount,
    formatConfiguredCount,
    learningContextSummary,
    learningContextSupportItems,
    panelsUpdatedSince,
    parseKnownTimestamp,
    responseContractSummary,
    responseContractSupportItems,
    toTitleCase,
} from "@/components/organizations/organizationSummaryHelpers";
import { useOrganizationLivePanelData } from "@/components/organizations/useOrganizationLivePanelData";

export {
    extractOutputsFromConversation,
    findLatestConversationOutcome,
};

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

type SomaWorkspaceMode = "conversation" | "team_design";

type GuidedWorkspaceCardDefinition = {
    eyebrow: string;
    title: string;
    summary: string;
    buttonLabel: string;
    onClick: () => void;
};

const AI_ENGINE_OPTIONS: Array<{
    id: OrganizationAIEngineProfileId;
    label: string;
    description: string;
    goodFor: string;
}> = [
    {
        id: "starter_defaults",
        label: "Starter Defaults",
        description: "Keeps the guided starter profile that already came with this AI Organization.",
        goodFor: "Best for keeping the original setup intact while Soma settles the first workflow.",
    },
    {
        id: "balanced",
        label: "Balanced",
        description: "Steady planning depth and response quality for everyday Soma work.",
        goodFor: "Best for most organizations that want dependable guidance across planning and execution.",
    },
    {
        id: "high_reasoning",
        label: "High Reasoning",
        description: "Adds more careful thinking when planning and tradeoffs need extra attention.",
        goodFor: "Best for complex decisions, deeper review, and more deliberate next-step planning.",
    },
    {
        id: "fast_lightweight",
        label: "Fast & Lightweight",
        description: "Keeps responses quick and planning lighter for rapid iteration.",
        goodFor: "Best for quick reviews, check-ins, and lighter day-to-day coordination.",
    },
    {
        id: "deep_planning",
        label: "Deep Planning",
        description: "Leans into longer multi-step planning and more deliberate organization shaping.",
        goodFor: "Best for designing larger workstreams and sequencing bigger efforts.",
    },
];

const RESPONSE_CONTRACT_OPTIONS: Array<{
    id: ResponseContractProfileId;
    label: string;
    toneStyle: string;
    structure: string;
    verbosity: string;
    bestFor: string;
}> = [
    {
        id: "clear_balanced",
        label: "Clear & Balanced",
        toneStyle: "Straightforward and steady without sounding cold.",
        structure: "Uses clear sections and practical takeaways when helpful.",
        verbosity: "Balanced detail with enough context to act confidently.",
        bestFor: "Best for everyday Soma guidance, reviews, and general coordination.",
    },
    {
        id: "structured_analytical",
        label: "Structured & Analytical",
        toneStyle: "Measured, methodical, and reasoning-forward.",
        structure: "Organizes answers into clear steps, comparisons, or frameworks.",
        verbosity: "Moderate-to-detailed when structure improves decision-making.",
        bestFor: "Best for planning, tradeoffs, diagnosis, and deeper review work.",
    },
    {
        id: "concise_direct",
        label: "Concise & Direct",
        toneStyle: "Focused, efficient, and low-friction.",
        structure: "Keeps responses short and action-led unless more detail is needed.",
        verbosity: "Intentionally brief with only the highest-signal details.",
        bestFor: "Best for quick decisions, status checks, and fast-moving execution work.",
    },
    {
        id: "warm_supportive",
        label: "Warm & Supportive",
        toneStyle: "Encouraging, collaborative, and reassuring.",
        structure: "Still organized, but written to feel more human and supportive.",
        verbosity: "Balanced detail with a little more guidance and framing.",
        bestFor: "Best for onboarding, operator guidance, and people-facing support work.",
    },
];

const OUTPUT_TYPE_OPTIONS: Array<{ id: OrganizationOutputTypeId; label: string; description: string }> = [
    { id: "general_text", label: "General text", description: "Default chat, writing, and broad team communication outputs." },
    { id: "research_reasoning", label: "Research & reasoning", description: "Planning, review, synthesis, and deeper reasoning-heavy outputs." },
    { id: "code_generation", label: "Code generation", description: "Implementation, code repair, and execution-heavy output lanes." },
    { id: "vision_analysis", label: "Vision analysis", description: "Image understanding, OCR, and visual review support." },
];

export default function OrganizationContextShell({ organizationId }: { organizationId: string }) {
    const [organization, setOrganization] = useState<OrganizationHomePayload | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [retryToken, setRetryToken] = useState(0);
    const [activeDetailView, setActiveDetailView] = useState<"advisors" | "departments" | "automations" | "aiEngine" | "responseContract" | null>(null);
    const [isAIEngineSelectorOpen, setIsAIEngineSelectorOpen] = useState(false);
    const [selectedAIEngineProfile, setSelectedAIEngineProfile] = useState<OrganizationAIEngineProfileId | null>(null);
    const [aiEngineUpdatePending, setAIEngineUpdatePending] = useState(false);
    const [aiEngineUpdateError, setAIEngineUpdateError] = useState<string | null>(null);
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
    const {
        recentActivity,
        activityLoading,
        activityError,
        retryActivity,
        automations,
        automationsLoading,
        automationsError,
        retryAutomations,
        learningInsights,
        learningInsightsLoading,
        learningInsightsError,
        retryLearningInsights,
    } = useOrganizationLivePanelData(organizationId);
    const [isLaunchCrewOpen, setIsLaunchCrewOpen] = useState(false);
    const [somaWorkspaceMode, setSomaWorkspaceMode] = useState<SomaWorkspaceMode>("conversation");
    const [lastGuidanceUpdate, setLastGuidanceUpdate] = useState<SomaGuidanceUpdate | null>(null);
    const missionChat = useCortexStore((s) => s.missionChat);

    useEffect(() => {
        let cancelled = false;

        const load = async () => {
            setLoading(true);
            setError(null);
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/home`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Unable to load AI Organization.");
                }
                if (cancelled) {
                    return;
                }
                setOrganization(extractApiData<OrganizationHomePayload>(payload));
            } catch (err) {
                if (cancelled) {
                    return;
                }
                setError(err instanceof Error ? err.message : "Unable to load AI Organization.");
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        };

        void load();
        return () => {
            cancelled = true;
        };
    }, [organizationId, retryToken]);

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
                if (cancelled) {
                    return;
                }
                const data = extractApiData<OrganizationOutputModelRoutingPayload>(payload);
                setOutputModelRouting(data);
                setSelectedOutputModelRoutingMode((data.routing_mode ?? "single_model") as OrganizationOutputModelRoutingMode);
                setSelectedDefaultOutputModelId(data.default_model_id ?? "");
                setSelectedOutputModelBindings(
                    Object.fromEntries((data.bindings ?? []).map((binding) => [binding.output_type_id, binding.model_id ?? data.default_model_id ?? ""])),
                );
            } catch (err) {
                if (cancelled) {
                    return;
                }
                setOutputModelRoutingError(err instanceof Error ? err.message : "Unable to load output model routing.");
            } finally {
                if (!cancelled) {
                    setOutputModelRoutingLoading(false);
                }
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
        }
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
        if (activeDetailView !== "responseContract") {
            setIsResponseContractSelectorOpen(false);
            setSelectedResponseContractProfile(null);
            setResponseContractUpdatePending(false);
            setResponseContractUpdateError(null);
        }
        if (activeDetailView !== "aiEngine") {
            setOutputModelRoutingUpdatePending(false);
            setOutputModelRoutingUpdateError(null);
        }
    }, [activeDetailView]);

    const openAIEngineSelector = () => {
        if (!organization) {
            return;
        }
        setActiveDetailView("aiEngine");
        setSelectedAIEngineProfile((organization.ai_engine_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setAIEngineUpdateError(null);
        if (outputModelRouting) {
            setSelectedOutputModelRoutingMode(outputModelRouting.routing_mode);
            setSelectedDefaultOutputModelId(outputModelRouting.default_model_id ?? "");
            setSelectedOutputModelBindings(
                Object.fromEntries((outputModelRouting.bindings ?? []).map((binding) => [binding.output_type_id, binding.model_id ?? outputModelRouting.default_model_id ?? ""])),
            );
        }
        setOutputModelRoutingUpdateError(null);
        setIsAIEngineSelectorOpen(true);
    };

    const submitAIEngineSelection = async () => {
        if (!organization || !selectedAIEngineProfile || aiEngineUpdatePending) {
            return;
        }

        setAIEngineUpdatePending(true);
        setAIEngineUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedAIEngineProfile } satisfies OrganizationAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update AI Engine Settings.");
            }

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
        if (!organization || !selectedDefaultOutputModelId || outputModelRoutingUpdatePending) {
            return;
        }

        setOutputModelRoutingUpdatePending(true);
        setOutputModelRoutingUpdateError(null);

        const bindings = OUTPUT_TYPE_OPTIONS.map((option) => {
            const modelId = selectedOutputModelBindings[option.id] || selectedDefaultOutputModelId;
            return {
                output_type_id: option.id,
                model_id: modelId,
                use_organization_default: modelId === selectedDefaultOutputModelId,
            };
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
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update output model routing.");
            }

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
        if (!organization) {
            return;
        }
        setActiveDetailView("responseContract");
        setSelectedResponseContractProfile((organization.response_contract_profile_id as ResponseContractProfileId | undefined) ?? "clear_balanced");
        setResponseContractUpdateError(null);
        setIsResponseContractSelectorOpen(true);
    };

    const submitResponseContractSelection = async () => {
        if (!organization || !selectedResponseContractProfile || responseContractUpdatePending) {
            return;
        }

        setResponseContractUpdatePending(true);
        setResponseContractUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/response-contract`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedResponseContractProfile } satisfies ResponseContractUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update Response Style.");
            }

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

    const openDepartmentAIEngineSelector = (department: OrganizationDepartmentSummary) => {
        setActiveDetailView("departments");
        setActiveDepartmentAIEngineId(department.id);
        setSelectedDepartmentAIEngineProfile((department.ai_engine_effective_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setDepartmentAIEngineUpdateError(null);
    };

    const submitDepartmentAIEngineSelection = async (departmentId: string) => {
        if (!organization || !selectedDepartmentAIEngineProfile || departmentAIEngineUpdatePendingId) {
            return;
        }

        setDepartmentAIEngineUpdatePendingId(departmentId);
        setDepartmentAIEngineUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedDepartmentAIEngineProfile } satisfies DepartmentAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update this Department AI Engine.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveDepartmentAIEngineId(null);
            setSelectedDepartmentAIEngineProfile(null);
        } catch (err) {
            setDepartmentAIEngineUpdateError({
                departmentId,
                message: err instanceof Error ? err.message : "Unable to update this Department AI Engine.",
            });
        } finally {
            setDepartmentAIEngineUpdatePendingId(null);
        }
    };

    const revertDepartmentAIEngineSelection = async (departmentId: string) => {
        if (!organization || departmentAIEngineUpdatePendingId) {
            return;
        }

        setDepartmentAIEngineUpdatePendingId(departmentId);
        setDepartmentAIEngineUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ revert_to_organization_default: true } satisfies DepartmentAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to return this Department to the organization default.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveDepartmentAIEngineId(null);
            setSelectedDepartmentAIEngineProfile(null);
        } catch (err) {
            setDepartmentAIEngineUpdateError({
                departmentId,
                message: err instanceof Error ? err.message : "Unable to return this Department to the organization default.",
            });
        } finally {
            setDepartmentAIEngineUpdatePendingId(null);
        }
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

    const submitAgentTypeAIEngineSelection = async (departmentId: string, agentTypeId: string) => {
        if (!organization || !selectedAgentTypeAIEngineProfile || agentTypeAIEngineUpdatePendingKey) {
            return;
        }

        const selectionKey = agentTypeSelectionKey(departmentId, agentTypeId);
        setAgentTypeAIEngineUpdatePendingKey(selectionKey);
        setAgentTypeAIEngineUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedAgentTypeAIEngineProfile } satisfies AgentTypeAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update this Agent Type AI Engine.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveAgentTypeAIEngineKey(null);
            setSelectedAgentTypeAIEngineProfile(null);
        } catch (err) {
            setAgentTypeAIEngineUpdateError({
                departmentId,
                agentTypeId,
                message: err instanceof Error ? err.message : "Unable to update this Agent Type AI Engine.",
            });
        } finally {
            setAgentTypeAIEngineUpdatePendingKey(null);
        }
    };

    const revertAgentTypeAIEngineSelection = async (departmentId: string, agentTypeId: string) => {
        if (!organization || agentTypeAIEngineUpdatePendingKey) {
            return;
        }

        const selectionKey = agentTypeSelectionKey(departmentId, agentTypeId);
        setAgentTypeAIEngineUpdatePendingKey(selectionKey);
        setAgentTypeAIEngineUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/ai-engine`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ use_team_default: true } satisfies AgentTypeAIEngineUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to return this Agent Type to the Team default.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveAgentTypeAIEngineKey(null);
            setSelectedAgentTypeAIEngineProfile(null);
        } catch (err) {
            setAgentTypeAIEngineUpdateError({
                departmentId,
                agentTypeId,
                message: err instanceof Error ? err.message : "Unable to return this Agent Type to the Team default.",
            });
        } finally {
            setAgentTypeAIEngineUpdatePendingKey(null);
        }
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

    const submitAgentTypeResponseContractSelection = async (departmentId: string, agentTypeId: string) => {
        if (!organization || !selectedAgentTypeResponseContractProfile || agentTypeResponseContractUpdatePendingKey) {
            return;
        }

        const selectionKey = agentTypeSelectionKey(departmentId, agentTypeId);
        setAgentTypeResponseContractUpdatePendingKey(selectionKey);
        setAgentTypeResponseContractUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/response-contract`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profile_id: selectedAgentTypeResponseContractProfile } satisfies AgentTypeResponseContractUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to update this Agent Type Response Style.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveAgentTypeResponseContractKey(null);
            setSelectedAgentTypeResponseContractProfile(null);
        } catch (err) {
            setAgentTypeResponseContractUpdateError({
                departmentId,
                agentTypeId,
                message: err instanceof Error ? err.message : "Unable to update this Agent Type Response Style.",
            });
        } finally {
            setAgentTypeResponseContractUpdatePendingKey(null);
        }
    };

    const revertAgentTypeResponseContractSelection = async (departmentId: string, agentTypeId: string) => {
        if (!organization || agentTypeResponseContractUpdatePendingKey) {
            return;
        }

        const selectionKey = agentTypeSelectionKey(departmentId, agentTypeId);
        setAgentTypeResponseContractUpdatePendingKey(selectionKey);
        setAgentTypeResponseContractUpdateError(null);

        try {
            const response = await fetch(`/api/v1/organizations/${organization.id}/departments/${departmentId}/agent-types/${agentTypeId}/response-contract`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ use_organization_or_team_default: true } satisfies AgentTypeResponseContractUpdateRequest),
            });
            const payload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(payload) || "Unable to return this Agent Type to the Organization / Team default.");
            }

            const updated = extractApiData<OrganizationHomePayload>(payload);
            setOrganization(updated);
            setActiveAgentTypeResponseContractKey(null);
            setSelectedAgentTypeResponseContractProfile(null);
        } catch (err) {
            setAgentTypeResponseContractUpdateError({
                departmentId,
                agentTypeId,
                message: err instanceof Error ? err.message : "Unable to return this Agent Type to the Organization / Team default.",
            });
        } finally {
            setAgentTypeResponseContractUpdatePendingKey(null);
        }
    };

    useEffect(() => {
        if (typeof window === "undefined" || !organization) {
            return;
        }
        rememberLastOrganization({ id: organization.id, name: organization.name });
    }, [organization]);

    if (loading) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg">
                <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-surface px-5 py-4 text-sm text-cortex-text-muted">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Loading AI Organization...
                </div>
            </div>
        );
    }

    if (error || !organization) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg px-6">
                <div className="max-w-lg rounded-3xl border border-cortex-danger/30 bg-cortex-surface p-6">
                    <p className="text-lg font-semibold text-cortex-text-main">AI Organization unavailable</p>
                    <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{error || "This AI Organization could not be loaded."}</p>
                    <p className="mt-3 text-sm leading-7 text-cortex-text-muted">
                        Try loading the organization again, or return to the setup screen to create a new AI Organization.
                    </p>
                    <div className="mt-5 flex flex-wrap gap-3">
                        <button
                            onClick={() => setRetryToken((value) => value + 1)}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                        >
                            <RefreshCcw className="h-4 w-4" />
                            Retry
                        </button>
                        <Link href="/dashboard" className="inline-flex items-center gap-2 rounded-xl border border-transparent px-3 py-2 text-sm font-medium text-cortex-primary hover:bg-cortex-primary/10">
                            <ArrowLeft className="h-4 w-4" />
                            Return to Create AI Organization
                        </Link>
                    </div>
                </div>
            </div>
        );
    }

    const somaName = `Soma for ${organization.name}`;
    const teamLeadName = `${organization.team_lead_label} for ${organization.name}`;
    const lastConversationAction = [...missionChat].reverse().find((message) => message.role === "user" && !message.content.startsWith("[BROADCAST]"));
    const lastConversationOutcome = findLatestConversationOutcome(missionChat, teamLeadName);
    const latestConversationTimestamp = lastConversationOutcome?.timestamp ?? 0;
    const latestGuidanceTimestamp = parseKnownTimestamp(lastGuidanceUpdate?.timestamp);
    const latestSomaTimestamp = Math.max(latestConversationTimestamp, latestGuidanceTimestamp);
    const overviewItems = [
        { label: "Started from", value: organization.start_mode === "template" ? (organization.template_name || "Template") : "Empty" },
        { label: "Advisors", value: formatConfiguredCount(organization.advisor_count, "Advisor") },
        { label: "Departments", value: formatConfiguredCount(organization.department_count, "Department") },
        { label: "Specialists", value: formatConfiguredCount(organization.specialist_count, "Specialist") },
        { label: "AI Organization", value: toTitleCase(organization.status) },
    ];
    const panelUpdates = panelsUpdatedSince(latestSomaTimestamp, recentActivity, automations, learningInsights);
    const causalStrip: CausalStripState =
        latestGuidanceTimestamp > latestConversationTimestamp && lastGuidanceUpdate
            ? {
                  action: lastGuidanceUpdate.requestLabel,
                  teamsEngaged: lastGuidanceUpdate.teamsEngaged,
                  outputsGenerated: lastGuidanceUpdate.outputs,
                  panelsUpdated: panelUpdates,
              }
            : lastConversationOutcome
              ? {
                    action: lastConversationOutcome.actionLabel,
                    teamsEngaged: lastConversationOutcome.teamsEngaged,
                    outputsGenerated: lastConversationOutcome.outputsGenerated,
                    panelsUpdated: panelUpdates,
                }
              : lastConversationAction
                ? {
                    action: lastConversationAction.content,
                    teamsEngaged: ["Soma"],
                    outputsGenerated: ["Awaiting a trustworthy workspace outcome"],
                    panelsUpdated: panelUpdates.length > 0 ? panelUpdates : ["Workspace chat"],
                }
              : {
                    action: "Ready for your first Soma request",
                    teamsEngaged: ["Soma"],
                    outputsGenerated: ["Conversation guidance"],
                    panelsUpdated: panelUpdates.length > 0 ? panelUpdates : ["Quick Checks"],
                };

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <section className="rounded-2xl border border-cortex-border bg-cortex-surface px-5 py-4">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                        <nav aria-label="Organization breadcrumb" className="flex flex-wrap items-center gap-2 text-sm text-cortex-text-muted">
                            <Link href="/dashboard" className="font-medium text-cortex-primary transition-colors hover:text-cortex-primary/80">
                                AI Organizations
                            </Link>
                            <span>/</span>
                            <span className="font-medium text-cortex-text-main">{organization.name}</span>
                        </nav>
                        <Link
                            href="/dashboard"
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                        >
                            <ArrowLeft className="h-4 w-4" />
                            Return to Organization list
                        </Link>
                    </div>
                </section>

                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                    <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
                        <div className="space-y-3">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">
                                <Blocks className="h-3.5 w-3.5" />
                                AI Organization Home
                            </div>
                            <div>
                                <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">{organization.name}</h1>
                                <p className="mt-2 max-w-3xl text-base leading-7 text-cortex-text-muted">{organization.purpose}</p>
                            </div>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                            <p className="font-medium text-cortex-text-main">Soma ready</p>
                            <p className="mt-1 leading-6">
                                {somaName} is ready to guide planning, structure review, and organization setup decisions while working through the right Team Lead support when needed.
                            </p>
                            <a
                                href="#soma-panel"
                                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-2 font-medium text-cortex-primary transition-colors hover:bg-cortex-primary/15"
                            >
                                Start with Soma
                                <Sparkles className="h-4 w-4" />
                            </a>
                        </div>
                    </div>
                </section>

                <section className="space-y-4">
                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                <div className="space-y-3">
                                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                                        <Sparkles className="h-3.5 w-3.5" />
                                        Soma
                                    </div>
                                    <div>
                                        <h2 className="text-2xl font-semibold text-cortex-text-main">{somaName}</h2>
                                        <p className="mt-2 max-w-2xl text-sm leading-7 text-cortex-text-muted">
                                            Soma is the primary counterpart for {organization.name}, coordinating the right Team Leads, Advisors, Departments, and Specialists around the organization purpose.
                                        </p>
                                    </div>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-xs">
                                    <p className="font-medium text-cortex-text-main">Operational layer</p>
                                    <p className="mt-1 leading-6">
                                        {teamLeadName} leads day-to-day operational follow-through while Soma keeps the organization focused and surfaces the next best actions.
                                    </p>
                                </div>
                            </div>

                            <div className="mt-6">
                                <p className="text-sm font-medium text-cortex-text-main">Start here in this organization</p>
                                <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">
                                    Choose the first move that matches what you need right now, then use the deeper inspect surfaces only when you want more structure, operating detail, or guided tuning.
                                </p>
                                <div className="mt-4 grid gap-3 lg:grid-cols-3">
                                    {[
                                        {
                                            eyebrow: "Primary conversation",
                                            title: "Plan, review, or create with Soma",
                                            summary: "Stay in the main Soma conversation for plans, drafts, imagery, governed changes, and broad delivery shaping.",
                                            buttonLabel: "Open Soma conversation",
                                            onClick: () => setSomaWorkspaceMode("conversation"),
                                        },
                                        {
                                            eyebrow: "Team design lane",
                                            title: "Shape the first team or lane",
                                            summary: "Move into the focused team-design mode when you want roles, delivery lanes, or execution structure to become explicit.",
                                            buttonLabel: "Open team design lane",
                                            onClick: () => setSomaWorkspaceMode("team_design"),
                                        },
                                        {
                                            eyebrow: "Setup review",
                                            title: "Review current organization setup",
                                            summary: "Inspect the current working structure when you want a clearer read on departments, specialists, and what is ready next.",
                                            buttonLabel: "Review organization setup",
                                            onClick: () => setActiveDetailView("departments"),
                                        },
                                    ].map((card) => (
                                        <GuidedWorkspaceCard key={card.title} {...card} />
                                    ))}
                                </div>
                            </div>

                            <div className="mt-6">
                                <p className="text-sm font-medium text-cortex-text-main">Inspect the current organization</p>
                                <div className="mt-3 flex flex-wrap gap-2">
                                    <HelpPill label="Run a quick strategy check" />
                                    <ActionPill
                                        label="Review Advisors"
                                        isActive={activeDetailView === "advisors"}
                                        onClick={() => setActiveDetailView("advisors")}
                                    />
                                    <ActionPill
                                        label="Open Departments"
                                        isActive={activeDetailView === "departments"}
                                        onClick={() => setActiveDetailView("departments")}
                                    />
                                    <ActionPill
                                        label="Review Automations"
                                        isActive={activeDetailView === "automations"}
                                        onClick={() => setActiveDetailView("automations")}
                                    />
                                    <ActionPill
                                        label="Review AI Engine Settings"
                                        isActive={activeDetailView === "aiEngine"}
                                        onClick={() => setActiveDetailView("aiEngine")}
                                    />
                                    <ActionPill
                                        label="Review Response Style"
                                        isActive={activeDetailView === "responseContract"}
                                        onClick={() => setActiveDetailView("responseContract")}
                                    />
                                </div>
                            </div>
                        </div>

                        {activeDetailView && (
                            <WorkspaceDetailView
                                view={activeDetailView}
                                organization={organization}
                                teamLeadName={teamLeadName}
                                automations={automations}
                                automationsLoading={automationsLoading}
                                automationsError={automationsError}
                                onRetryAutomations={retryAutomations}
                                isAIEngineSelectorOpen={isAIEngineSelectorOpen}
                                selectedAIEngineProfile={selectedAIEngineProfile}
                                aiEngineUpdatePending={aiEngineUpdatePending}
                                aiEngineUpdateError={aiEngineUpdateError}
                                outputModelRouting={outputModelRouting}
                                outputModelRoutingLoading={outputModelRoutingLoading}
                                outputModelRoutingError={outputModelRoutingError}
                                selectedOutputModelRoutingMode={selectedOutputModelRoutingMode}
                                selectedDefaultOutputModelId={selectedDefaultOutputModelId}
                                selectedOutputModelBindings={selectedOutputModelBindings}
                                outputModelRoutingUpdatePending={outputModelRoutingUpdatePending}
                                outputModelRoutingUpdateError={outputModelRoutingUpdateError}
                                onAIEngineProfileSelect={setSelectedAIEngineProfile}
                                onOutputModelRoutingModeChange={setSelectedOutputModelRoutingMode}
                                onDefaultOutputModelChange={handleDefaultOutputModelChange}
                                onOutputModelBindingChange={(outputTypeId, modelId) =>
                                    setSelectedOutputModelBindings((current) => ({ ...current, [outputTypeId]: modelId }))
                                }
                                onOpenAIEngineSelector={openAIEngineSelector}
                                onCloseAIEngineSelector={() => {
                                    setIsAIEngineSelectorOpen(false);
                                    setAIEngineUpdateError(null);
                                }}
                                onSubmitAIEngineSelection={submitAIEngineSelection}
                                onSubmitOutputModelRouting={submitOutputModelRouting}
                                isResponseContractSelectorOpen={isResponseContractSelectorOpen}
                                selectedResponseContractProfile={selectedResponseContractProfile}
                                responseContractUpdatePending={responseContractUpdatePending}
                                responseContractUpdateError={responseContractUpdateError}
                                onResponseContractProfileSelect={setSelectedResponseContractProfile}
                                onOpenResponseContractSelector={openResponseContractSelector}
                                onCloseResponseContractSelector={() => {
                                    setIsResponseContractSelectorOpen(false);
                                    setResponseContractUpdateError(null);
                                }}
                                onSubmitResponseContractSelection={submitResponseContractSelection}
                                activeDepartmentAIEngineId={activeDepartmentAIEngineId}
                                selectedDepartmentAIEngineProfile={selectedDepartmentAIEngineProfile}
                                departmentAIEngineUpdatePendingId={departmentAIEngineUpdatePendingId}
                                departmentAIEngineUpdateError={departmentAIEngineUpdateError}
                                onDepartmentAIEngineProfileSelect={setSelectedDepartmentAIEngineProfile}
                                onOpenDepartmentAIEngineSelector={openDepartmentAIEngineSelector}
                                onCloseDepartmentAIEngineSelector={() => {
                                    setActiveDepartmentAIEngineId(null);
                                    setSelectedDepartmentAIEngineProfile(null);
                                    setDepartmentAIEngineUpdateError(null);
                                }}
                                onSubmitDepartmentAIEngineSelection={submitDepartmentAIEngineSelection}
                                onRevertDepartmentAIEngineSelection={revertDepartmentAIEngineSelection}
                                activeAgentTypeAIEngineKey={activeAgentTypeAIEngineKey}
                                selectedAgentTypeAIEngineProfile={selectedAgentTypeAIEngineProfile}
                                agentTypeAIEngineUpdatePendingKey={agentTypeAIEngineUpdatePendingKey}
                                agentTypeAIEngineUpdateError={agentTypeAIEngineUpdateError}
                                onAgentTypeAIEngineProfileSelect={setSelectedAgentTypeAIEngineProfile}
                                onOpenAgentTypeAIEngineSelector={openAgentTypeAIEngineSelector}
                                onCloseAgentTypeAIEngineSelector={() => {
                                    setActiveAgentTypeAIEngineKey(null);
                                    setSelectedAgentTypeAIEngineProfile(null);
                                    setAgentTypeAIEngineUpdateError(null);
                                }}
                                onSubmitAgentTypeAIEngineSelection={submitAgentTypeAIEngineSelection}
                                onRevertAgentTypeAIEngineSelection={revertAgentTypeAIEngineSelection}
                                activeAgentTypeResponseContractKey={activeAgentTypeResponseContractKey}
                                selectedAgentTypeResponseContractProfile={selectedAgentTypeResponseContractProfile}
                                agentTypeResponseContractUpdatePendingKey={agentTypeResponseContractUpdatePendingKey}
                                agentTypeResponseContractUpdateError={agentTypeResponseContractUpdateError}
                                onAgentTypeResponseContractProfileSelect={setSelectedAgentTypeResponseContractProfile}
                                onOpenAgentTypeResponseContractSelector={openAgentTypeResponseContractSelector}
                                onCloseAgentTypeResponseContractSelector={() => {
                                    setActiveAgentTypeResponseContractKey(null);
                                    setSelectedAgentTypeResponseContractProfile(null);
                                    setAgentTypeResponseContractUpdateError(null);
                                }}
                                onSubmitAgentTypeResponseContractSelection={submitAgentTypeResponseContractSelection}
                                onRevertAgentTypeResponseContractSelection={revertAgentTypeResponseContractSelection}
                                onBack={() => setActiveDetailView(null)}
                            />
                        )}

                        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-6 shadow-[0_24px_60px_rgba(29,42,53,0.08)]">
                            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                <div className="space-y-3">
                                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                                        <Bot className="h-3.5 w-3.5" />
                                        Soma conversation
                                    </div>
                                    <div>
                                        <h2 className="text-2xl font-semibold text-cortex-text-main">Talk with Soma</h2>
                                        <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                                            Use Soma as the primary root workspace for plans, concepts, imagery, drafts, and delivery shaping. Admins can ask Soma to create teams, structure new lanes, and coordinate the right advisor support before work is handed to focused Team Leads, Departments, and Specialists.
                                        </p>
                                    </div>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                    <p className="font-medium text-cortex-text-main">How to read this workspace</p>
                                    <p className="mt-1 leading-6">
                                        Use Soma as the main interface. The overview and quick checks alongside it explain what is healthy, what is active, and what may need attention while you shape teams with Soma and then move into a focused Team Lead workspace when a specific lane is selected.
                                    </p>
                                </div>
                            </div>
                            <SomaCausalStrip
                                action={causalStrip.action}
                                teamsEngaged={causalStrip.teamsEngaged}
                                outputsGenerated={causalStrip.outputsGenerated}
                                panelsUpdated={causalStrip.panelsUpdated}
                            />
                            <div className="mt-5 grid gap-3 md:grid-cols-4">
                                {overviewItems.map((item) => (
                                    <Metric key={item.label} label={item.label} value={item.value} />
                                ))}
                            </div>
                            <div className="mt-5 flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                                <div className="flex flex-wrap gap-3">
                                    <button
                                        type="button"
                                        onClick={() => setSomaWorkspaceMode("conversation")}
                                        className={`inline-flex items-center gap-2 rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors ${
                                            somaWorkspaceMode === "conversation"
                                                ? "border-cortex-primary/35 bg-cortex-primary text-cortex-bg"
                                                : "border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
                                        }`}
                                    >
                                        Talk with Soma
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setSomaWorkspaceMode("team_design")}
                                        className={`inline-flex items-center gap-2 rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors ${
                                            somaWorkspaceMode === "team_design"
                                                ? "border-cortex-primary/35 bg-cortex-primary text-cortex-bg"
                                                : "border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
                                        }`}
                                    >
                                        Create teams with Soma
                                    </button>
                                    {somaWorkspaceMode === "team_design" ? (
                                        <button
                                            type="button"
                                            onClick={() => setIsLaunchCrewOpen(true)}
                                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                                        >
                                            Open crew launcher
                                        </button>
                                    ) : null}
                                </div>
                                <span className="inline-flex items-center rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm text-cortex-text-muted">
                                    {somaWorkspaceMode === "conversation"
                                        ? "Stay in Soma's root workspace for planning, team creation, concepts, imagery, and broad delivery shaping."
                                        : "Use team design mode when you want Soma to turn the current conversation into roles, lanes, and execution structure."}
                                </span>
                            </div>
                            <div className="mt-6 overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg">
                                {somaWorkspaceMode === "conversation" ? (
                                    <div className="h-[46rem] min-h-0 lg:h-[52rem]">
                                        <MissionControlChat simpleMode autoFocus organizationId={organizationId} />
                                    </div>
                                ) : (
                                    <div className="p-5 lg:p-6">
                                        <TeamLeadInteractionPanel
                                            organizationId={organization.id}
                                            organizationName={organization.name}
                                            somaName={somaName}
                                            teamLeadName={teamLeadName}
                                            autoFocusOnLoad
                                            embedded
                                            onGuidanceStateChange={setLastGuidanceUpdate}
                                        />
                                    </div>
                                )}
                            </div>
                        </section>
                    </div>

                    <section className="grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
                        <div className="space-y-4">
                            <RecentActivityPanel
                                items={recentActivity}
                                loading={activityLoading}
                                error={activityError}
                                onRetry={retryActivity}
                                causalAction={causalStrip.action}
                            />

                            <LearningVisibilityPanel
                                items={learningInsights}
                                loading={learningInsightsLoading}
                                error={learningInsightsError}
                                onRetry={retryLearningInsights}
                                causalAction={causalStrip.action}
                            />
                        </div>
                        <div className="space-y-4">
                            <SystemQuickChecks />

                            <InspectOnlySummary
                                icon={<Users className="h-4 w-4" />}
                                title="Advisors"
                                countLabel={formatConfiguredCount(organization.advisor_count, "Advisor")}
                                summary={advisorSummary(organization.advisor_count, teamLeadName)}
                                supportLabel="Advisor support"
                                items={advisorSupportItems(organization.advisor_count)}
                                changeSummary={organization.advisor_count > 0 ? `${formatConfiguredCount(organization.advisor_count, "Advisor")} visible right now` : "No advisor support configured yet"}
                                changeReason="Advisor support expands when Soma needs more review coverage around current work."
                                somaConnection="Soma uses advisor support for decision review, priority checks, and quality assurance."
                                inspectActionLabel="Review Advisors"
                                onInspect={() => setActiveDetailView("advisors")}
                            />

                            <InspectOnlySummary
                                icon={<Building2 className="h-4 w-4" />}
                                title="Departments"
                                countLabel={formatConfiguredCount(organization.department_count, "Department")}
                                summary={departmentSummary(organization.department_count, organization.specialist_count, teamLeadName)}
                                supportLabel="Visible specialist roles"
                                items={departmentSupportItems(organization)}
                                changeSummary={organization.department_count > 0 ? `${formatConfiguredCount(organization.department_count, "Department")} shaping delivery` : "No working lanes defined yet"}
                                changeReason="Departments and specialist roles show the execution structure Soma can work through next."
                                somaConnection="Soma uses Departments and Specialists to turn intent into concrete delivery lanes."
                                inspectActionLabel="Open Departments"
                                onInspect={() => setActiveDetailView("departments")}
                            />

                            <InspectOnlySummary
                                icon={<Blocks className="h-4 w-4" />}
                                title="Automations"
                                countLabel={formatAutomationCount(automations.length, automationsLoading, automationsError)}
                                statusLabel={automationStatusLabel(automationsLoading, automationsError)}
                                summary={automationSummary(automations.length, teamLeadName)}
                                supportLabel="What these cover"
                                items={automationSupportItems(automations, automationsLoading, automationsError)}
                                changeSummary={automationsError ? "Automation details unavailable right now" : formatAutomationCount(automations.length, automationsLoading, automationsError)}
                                changeReason="Automations refresh as ongoing reviews and checks become visible around the organization."
                                somaConnection="Soma uses Automations to explain what is watching, reviewing, and checking work in the background."
                                inspectActionLabel="Review Automations"
                                onInspect={() => setActiveDetailView("automations")}
                            />

                            <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                                <div className="grid gap-4">
                                    <InspectOnlySummary
                                        icon={<Bot className="h-4 w-4" />}
                                        title="AI Engine Settings"
                                        countLabel="Guided tuning"
                                        statusLabel="Organization level"
                                        summary={aiEngineSummary(organization.ai_engine_settings_summary)}
                                        supportLabel="What this affects"
                                        items={aiEngineSupportItems(organization.ai_engine_settings_summary)}
                                        changeSummary={`Current profile: ${organization.ai_engine_settings_summary || "Starter Defaults"}`}
                                        changeReason="This guided choice shapes how deeply Soma plans, responds, and carries work forward."
                                        somaConnection="Soma uses these AI Engine settings as the organization-wide operating posture."
                                        inspectActionLabel="Review AI Engine Settings"
                                        onInspect={() => setActiveDetailView("aiEngine")}
                                    />
                                    <InspectOnlySummary
                                        icon={<BrainCircuit className="h-4 w-4" />}
                                        title="Response Style"
                                        countLabel="Guided tuning"
                                        statusLabel="Organization level"
                                        summary={responseContractSummary(organization.response_contract_summary)}
                                        supportLabel="What this shapes"
                                        items={responseContractSupportItems(organization.response_contract_summary)}
                                        changeSummary={`Current profile: ${organization.response_contract_summary || "Clear & Balanced"}`}
                                        changeReason="This guided choice shapes how Soma sounds, structures responses, and presents detail."
                                        somaConnection="Soma uses this Response Style to keep the organization consistent and readable."
                                        inspectActionLabel="Review Response Style"
                                        onInspect={() => setActiveDetailView("responseContract")}
                                    />
                                    <InspectOnlySummary
                                        icon={<BrainCircuit className="h-4 w-4" />}
                                        title="Memory & Continuity"
                                        countLabel="Inspect only"
                                        summary={learningContextSummary(organization.memory_personality_summary)}
                                        supportLabel="What this affects"
                                        items={learningContextSupportItems(organization.memory_personality_summary)}
                                        changeSummary={learningInsights.length > 0 ? `${learningInsights.length} retained pattern${learningInsights.length === 1 ? "" : "s"} visible` : "No retained patterns visible yet"}
                                        changeReason="Reusable patterns and continuity cues become visible as Soma and supporting reviews build repeatable signals."
                                        somaConnection="Soma uses durable memory for reusable recall and temporary continuity to stay oriented across return visits and later work."
                                    />
                                </div>
                                <div className="mt-5">
                                    <Link href="/dashboard" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                                        <ArrowLeft className="h-4 w-4" />
                                        Create another AI Organization
                                    </Link>
                                </div>
                            </div>
                        </div>
                    </section>
                </section>
            </div>
            {isLaunchCrewOpen && <LaunchCrewModal onClose={() => setIsLaunchCrewOpen(false)} />}
        </div>
    );
}

function responseContractDetailItems(summary: string) {
    const option = RESPONSE_CONTRACT_OPTIONS.find((item) => item.label === summary.trim()) ?? RESPONSE_CONTRACT_OPTIONS[0];
    return [
        {
            name: "Current response style",
            purpose: `Current profile: ${option.label}.`,
            supportCue: option.bestFor,
        },
        {
            name: "Tone and style",
            purpose: option.toneStyle,
            supportCue: "Guides how supportive, direct, or analytical Soma should sound by default.",
        },
        {
            name: "Structure and detail",
            purpose: `${option.structure} ${option.verbosity}`,
            supportCue: "Guides how organized and how detailed responses should feel across the AI Organization.",
        },
    ];
}

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

function HelpPill({ label }: { label: string }) {
    return (
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </div>
    );
}

function ActionPill({
    label,
    isActive,
    onClick,
}: {
    label: string;
    isActive?: boolean;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm transition-colors ${
                isActive
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
            }`}
        >
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </button>
    );
}

function InspectOnlySummary({
    icon,
    title,
    countLabel,
    summary,
    supportLabel,
    items,
    changeSummary,
    changeReason,
    somaConnection,
    inspectActionLabel,
    onInspect,
    statusLabel = "Inspect only",
}: {
    icon: React.ReactNode;
    title: string;
    countLabel: string;
    summary: string;
    supportLabel: string;
    items: Array<string | { label: string; detail?: string }>;
    changeSummary: string;
    changeReason: string;
    somaConnection: string;
    inspectActionLabel?: string;
    onInspect?: () => void;
    statusLabel?: string;
}) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex items-start justify-between gap-4">
                <div>
                    <div className="flex items-center gap-2 text-cortex-primary">
                        {icon}
                        <h2 className="text-xl font-semibold text-cortex-text-main">{title}</h2>
                    </div>
                    <p className="mt-1 text-sm text-cortex-text-muted">{summary}</p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">{countLabel}</p>
                    <p className="mt-1">{statusLabel}</p>
                </div>
            </div>
            <div className="mt-5">
                <div className="grid gap-3 lg:grid-cols-3">
                    <CausalFact label="What changed" value={changeSummary} />
                    <CausalFact label="Why it changed" value={changeReason} />
                    <CausalFact label="How Soma uses it" value={somaConnection} />
                </div>
            </div>
            <div className="mt-5">
                <p className="text-sm font-medium text-cortex-text-main">{supportLabel}</p>
                <div className="mt-3 grid gap-2">
                    {items.map((item) => (
                        <div
                            key={typeof item === "string" ? item : `${item.label}-${item.detail ?? ""}`}
                            className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-3 text-sm text-cortex-text-main"
                        >
                            {typeof item === "string" ? (
                                item
                            ) : (
                                <>
                                    <p className="font-medium text-cortex-text-main">{item.label}</p>
                                    {item.detail ? <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{item.detail}</p> : null}
                                </>
                            )}
                        </div>
                    ))}
                </div>
            </div>
            {onInspect && inspectActionLabel && (
                <div className="mt-5">
                    <button
                        type="button"
                        onClick={onInspect}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <Sparkles className="h-4 w-4 text-cortex-primary" />
                        {inspectActionLabel}
                    </button>
                </div>
            )}
        </div>
    );
}

function WorkspaceDetailView({
    view,
    organization,
    teamLeadName,
    automations,
    automationsLoading,
    automationsError,
    onRetryAutomations,
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
    onAIEngineProfileSelect,
    onOutputModelRoutingModeChange,
    onDefaultOutputModelChange,
    onOutputModelBindingChange,
    onOpenAIEngineSelector,
    onCloseAIEngineSelector,
    onSubmitAIEngineSelection,
    onSubmitOutputModelRouting,
    isResponseContractSelectorOpen,
    selectedResponseContractProfile,
    responseContractUpdatePending,
    responseContractUpdateError,
    onResponseContractProfileSelect,
    onOpenResponseContractSelector,
    onCloseResponseContractSelector,
    onSubmitResponseContractSelection,
    activeDepartmentAIEngineId,
    selectedDepartmentAIEngineProfile,
    departmentAIEngineUpdatePendingId,
    departmentAIEngineUpdateError,
    onDepartmentAIEngineProfileSelect,
    onOpenDepartmentAIEngineSelector,
    onCloseDepartmentAIEngineSelector,
    onSubmitDepartmentAIEngineSelection,
    onRevertDepartmentAIEngineSelection,
    activeAgentTypeAIEngineKey,
    selectedAgentTypeAIEngineProfile,
    agentTypeAIEngineUpdatePendingKey,
    agentTypeAIEngineUpdateError,
    onAgentTypeAIEngineProfileSelect,
    onOpenAgentTypeAIEngineSelector,
    onCloseAgentTypeAIEngineSelector,
    onSubmitAgentTypeAIEngineSelection,
    onRevertAgentTypeAIEngineSelection,
    activeAgentTypeResponseContractKey,
    selectedAgentTypeResponseContractProfile,
    agentTypeResponseContractUpdatePendingKey,
    agentTypeResponseContractUpdateError,
    onAgentTypeResponseContractProfileSelect,
    onOpenAgentTypeResponseContractSelector,
    onCloseAgentTypeResponseContractSelector,
    onSubmitAgentTypeResponseContractSelection,
    onRevertAgentTypeResponseContractSelection,
    onBack,
}: {
    view: "advisors" | "departments" | "automations" | "aiEngine" | "responseContract";
    organization: OrganizationHomePayload;
    teamLeadName: string;
    automations: OrganizationAutomationItem[];
    automationsLoading: boolean;
    automationsError: string | null;
    onRetryAutomations: () => void;
    isAIEngineSelectorOpen: boolean;
    selectedAIEngineProfile: OrganizationAIEngineProfileId | null;
    aiEngineUpdatePending: boolean;
    aiEngineUpdateError: string | null;
    outputModelRouting: OrganizationOutputModelRoutingPayload | null;
    outputModelRoutingLoading: boolean;
    outputModelRoutingError: string | null;
    selectedOutputModelRoutingMode: OrganizationOutputModelRoutingMode;
    selectedDefaultOutputModelId: string;
    selectedOutputModelBindings: Record<string, string>;
    outputModelRoutingUpdatePending: boolean;
    outputModelRoutingUpdateError: string | null;
    onAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onOutputModelRoutingModeChange: (mode: OrganizationOutputModelRoutingMode) => void;
    onDefaultOutputModelChange: (modelId: string) => void;
    onOutputModelBindingChange: (outputTypeId: OrganizationOutputTypeId, modelId: string) => void;
    onOpenAIEngineSelector: () => void;
    onCloseAIEngineSelector: () => void;
    onSubmitAIEngineSelection: () => void;
    onSubmitOutputModelRouting: () => void;
    isResponseContractSelectorOpen: boolean;
    selectedResponseContractProfile: ResponseContractProfileId | null;
    responseContractUpdatePending: boolean;
    responseContractUpdateError: string | null;
    onResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onOpenResponseContractSelector: () => void;
    onCloseResponseContractSelector: () => void;
    onSubmitResponseContractSelection: () => void;
    activeDepartmentAIEngineId: string | null;
    selectedDepartmentAIEngineProfile: OrganizationAIEngineProfileId | null;
    departmentAIEngineUpdatePendingId: string | null;
    departmentAIEngineUpdateError: { departmentId: string; message: string } | null;
    onDepartmentAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onOpenDepartmentAIEngineSelector: (department: OrganizationDepartmentSummary) => void;
    onCloseDepartmentAIEngineSelector: () => void;
    onSubmitDepartmentAIEngineSelection: (departmentId: string) => void;
    onRevertDepartmentAIEngineSelection: (departmentId: string) => void;
    activeAgentTypeAIEngineKey: string | null;
    selectedAgentTypeAIEngineProfile: OrganizationAIEngineProfileId | null;
    agentTypeAIEngineUpdatePendingKey: string | null;
    agentTypeAIEngineUpdateError: { departmentId: string; agentTypeId: string; message: string } | null;
    onAgentTypeAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onOpenAgentTypeAIEngineSelector: (departmentId: string, profile: OrganizationAgentTypeProfileSummary) => void;
    onCloseAgentTypeAIEngineSelector: () => void;
    onSubmitAgentTypeAIEngineSelection: (departmentId: string, agentTypeId: string) => void;
    onRevertAgentTypeAIEngineSelection: (departmentId: string, agentTypeId: string) => void;
    activeAgentTypeResponseContractKey: string | null;
    selectedAgentTypeResponseContractProfile: ResponseContractProfileId | null;
    agentTypeResponseContractUpdatePendingKey: string | null;
    agentTypeResponseContractUpdateError: { departmentId: string; agentTypeId: string; message: string } | null;
    onAgentTypeResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onOpenAgentTypeResponseContractSelector: (departmentId: string, profile: OrganizationAgentTypeProfileSummary) => void;
    onCloseAgentTypeResponseContractSelector: () => void;
    onSubmitAgentTypeResponseContractSelection: (departmentId: string, agentTypeId: string) => void;
    onRevertAgentTypeResponseContractSelection: (departmentId: string, agentTypeId: string) => void;
    onBack: () => void;
}) {
    const title =
        view === "advisors"
            ? "Advisor details"
            : view === "departments"
              ? "Department details"
              : view === "automations"
                ? "Automation details"
              : view === "aiEngine"
                ? "AI Engine Settings details"
                : "Response Style details";
    const summary =
        view === "advisors"
            ? `Soma can review the current Advisor support in ${organization.name} here without leaving the workspace.`
            : view === "departments"
              ? `Soma can inspect the current Department structure in ${organization.name} here without leaving the workspace while keeping ${teamLeadName} in view.`
              : view === "automations"
                ? `Soma can inspect the ongoing reviews and checks that support ${organization.name} here without leaving the workspace.`
              : view === "aiEngine"
                ? `Soma can inspect the current AI Engine Settings in ${organization.name} here without leaving the workspace.`
                : `Soma can inspect the current Response Style in ${organization.name} here without leaving the workspace.`;
    const items =
        view === "advisors"
            ? advisorDetailItems(organization.advisor_count).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support: item.supportCue,
              }))
            : view === "automations"
              ? []
            : (view === "aiEngine" ? aiEngineDetailItems(organization) : responseContractDetailItems(organization.response_contract_summary)).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support: item.supportCue,
              }));
    const departmentItems = view === "departments" ? departmentDetailItems(organization) : [];
    const automationItems = view === "automations" ? automations : [];
    const emptyStateMessage =
        view === "advisors"
            ? "Advisor support will appear here when Soma starts working with additional review help. Try reviewing your organization setup to decide where that support is needed first."
            : view === "departments"
              ? "Departments are the working lanes for this organization. They will appear here once the first lane is defined. Try reviewing your organization setup or running a quick strategy check."
              : view === "automations"
                ? "Automations will appear here as ongoing reviews and checks become available for this AI Organization. Run a quick strategy check or review your organization setup to start creating visible signals."
              : view === "aiEngine"
                ? `Soma is still using the shared AI engine profile until more scoped settings are surfaced here.`
                : `Soma is still using the organization-wide Response Style until a different guided profile is selected here.`;

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h3 className="text-xl font-semibold text-cortex-text-main">{title}</h3>
                    <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{summary}</p>
                </div>
                <div className="flex flex-wrap gap-3">
                    {view === "aiEngine" && (
                        <button
                            type="button"
                            onClick={isAIEngineSelectorOpen ? onCloseAIEngineSelector : onOpenAIEngineSelector}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40"
                        >
                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                            {isAIEngineSelectorOpen ? "Close AI Engine choices" : "Change AI Engine"}
                        </button>
                    )}
                    {view === "responseContract" && (
                        <button
                            type="button"
                            onClick={isResponseContractSelectorOpen ? onCloseResponseContractSelector : onOpenResponseContractSelector}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40"
                        >
                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                            {isResponseContractSelectorOpen ? "Close Response Style choices" : "Change Response Style"}
                        </button>
                    )}
                    <button
                        type="button"
                        onClick={onBack}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <ArrowLeft className="h-4 w-4" />
                        Back to Soma
                    </button>
                </div>
            </div>

            {view === "departments" ? (
                departmentItems.length > 0 ? (
                    <div className="mt-5 grid gap-4">
                        {departmentItems.map((item) => {
                            const isEditing = activeDepartmentAIEngineId === item.id;
                            const isPending = departmentAIEngineUpdatePendingId === item.id;
                            const errorMessage = departmentAIEngineUpdateError?.departmentId === item.id ? departmentAIEngineUpdateError.message : null;

                            return (
                                <div key={item.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                        <div>
                                            <p className="text-sm font-semibold text-cortex-text-main">{item.name}</p>
                                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.purpose}</p>
                                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                {item.specialist_count > 0
                                                    ? `${item.specialist_count} ${item.specialist_count === 1 ? "Specialist" : "Specialists"} visible here.`
                                                    : "No Specialists assigned yet."}
                                            </p>
                                        </div>
                                        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                            <p className="font-medium text-cortex-text-main">AI Engine status</p>
                                            <p className="mt-1 leading-6">{item.aiEngineStateLabel}</p>
                                            <p className="mt-2 leading-6">
                                                {item.inherits_organization_ai_engine
                                                    ? "This Department follows the organization-wide AI Engine unless you set a Department-specific choice."
                                                    : "This Department is using its own guided AI Engine choice and will keep it until you return it to the organization default."}
                                            </p>
                                        </div>
                                    </div>

                                    <div className="mt-4 flex flex-wrap gap-3">
                                        <button
                                            type="button"
                                            onClick={() => onOpenDepartmentAIEngineSelector(item)}
                                            disabled={Boolean(departmentAIEngineUpdatePendingId && departmentAIEngineUpdatePendingId !== item.id)}
                                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                                        >
                                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                                            Change for this Team
                                        </button>
                                        {!item.inherits_organization_ai_engine && (
                                            <button
                                                type="button"
                                                onClick={() => onRevertDepartmentAIEngineSelection(item.id)}
                                                disabled={isPending}
                                                className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                            >
                                                {isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                                Revert to Organization Default
                                            </button>
                                        )}
                                    </div>

                                    {isEditing && (
                                        <DepartmentAIEngineSelectionPanel
                                            selectedDepartmentAIEngineProfile={selectedDepartmentAIEngineProfile}
                                            departmentAIEngineUpdatePending={isPending}
                                            departmentAIEngineUpdateError={errorMessage}
                                            onDepartmentAIEngineProfileSelect={onDepartmentAIEngineProfileSelect}
                                            onClose={onCloseDepartmentAIEngineSelector}
                                            onSubmit={() => onSubmitDepartmentAIEngineSelection(item.id)}
                                        />
                                    )}

                                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                        <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                            <div>
                                                <p className="text-sm font-semibold text-cortex-text-main">Agent Type Profiles</p>
                                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                    These guided profiles show how this Team's role types inherit the Team default or follow a type-specific AI Engine or Response Style.
                                                </p>
                                            </div>
                                            <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                                <p className="font-medium text-cortex-text-main">Guided role tuning</p>
                                                <p className="mt-1 leading-6">Use these profiles to review or safely tune role-type defaults while individual Specialist instances stay out of scope.</p>
                                            </div>
                                        </div>

                                        {item.agentTypeProfiles.length > 0 ? (
                                            <div className="mt-4 grid gap-3">
                                                {item.agentTypeProfiles.map((profile) => {
                                                    const agentTypeKey = agentTypeSelectionKey(item.id, profile.id);
                                                    const isAgentTypeEditing = activeAgentTypeAIEngineKey === agentTypeKey;
                                                    const isAgentTypePending = agentTypeAIEngineUpdatePendingKey === agentTypeKey;
                                                    const isAgentTypeResponseEditing = activeAgentTypeResponseContractKey === agentTypeKey;
                                                    const isAgentTypeResponsePending = agentTypeResponseContractUpdatePendingKey === agentTypeKey;
                                                    const agentTypeError =
                                                        agentTypeAIEngineUpdateError?.departmentId === item.id && agentTypeAIEngineUpdateError.agentTypeId === profile.id
                                                            ? agentTypeAIEngineUpdateError.message
                                                            : null;
                                                    const agentTypeResponseError =
                                                        agentTypeResponseContractUpdateError?.departmentId === item.id &&
                                                        agentTypeResponseContractUpdateError.agentTypeId === profile.id
                                                            ? agentTypeResponseContractUpdateError.message
                                                            : null;

                                                    return (
                                                        <div key={profile.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                                                            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                                                <div>
                                                                    <p className="text-sm font-semibold text-cortex-text-main">{profile.name}</p>
                                                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{profile.helps_with}</p>
                                                                </div>
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                                                    <p className="font-medium text-cortex-text-main">Inheritance clarity</p>
                                                                    <p className="mt-1 leading-6">
                                                                        {profile.inherits_department_ai_engine
                                                                            ? "This agent type follows the Team AI Engine unless a type-specific binding is shown here."
                                                                            : "This agent type keeps its own AI Engine binding instead of following the Team default."}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                            <div className="mt-4 grid gap-3 lg:grid-cols-3">
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">AI Engine</p>
                                                                    <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeAIEngineSourceLabel(profile)}</p>
                                                                </div>
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Response Style</p>
                                                                    <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeResponseStyleSourceLabel(profile)}</p>
                                                                </div>
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Detected output model</p>
                                                                    <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeOutputModelSourceLabel(profile)}</p>
                                                                    {profile.output_type_label && (
                                                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                                            Output type: <span className="font-medium text-cortex-text-main">{profile.output_type_label}</span>
                                                                        </p>
                                                                    )}
                                                                </div>
                                                            </div>
                                                            <div className="mt-4 flex flex-wrap gap-3">
                                                                <button
                                                                    type="button"
                                                                    onClick={() => onOpenAgentTypeAIEngineSelector(item.id, profile)}
                                                                    disabled={
                                                                        Boolean(agentTypeAIEngineUpdatePendingKey && agentTypeAIEngineUpdatePendingKey !== agentTypeKey) ||
                                                                        Boolean(agentTypeResponseContractUpdatePendingKey && agentTypeResponseContractUpdatePendingKey !== agentTypeKey)
                                                                    }
                                                                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                                                                >
                                                                    <Sparkles className="h-4 w-4 text-cortex-primary" />
                                                                    Change for this Agent Type
                                                                </button>
                                                                {!profile.inherits_department_ai_engine && (
                                                                    <button
                                                                        type="button"
                                                                        onClick={() => onRevertAgentTypeAIEngineSelection(item.id, profile.id)}
                                                                        disabled={isAgentTypePending}
                                                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                                                    >
                                                                        {isAgentTypePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                                                        Use Team Default
                                                                    </button>
                                                                )}
                                                                <button
                                                                    type="button"
                                                                    onClick={() => onOpenAgentTypeResponseContractSelector(item.id, profile)}
                                                                    disabled={
                                                                        Boolean(agentTypeAIEngineUpdatePendingKey && agentTypeAIEngineUpdatePendingKey !== agentTypeKey) ||
                                                                        Boolean(agentTypeResponseContractUpdatePendingKey && agentTypeResponseContractUpdatePendingKey !== agentTypeKey)
                                                                    }
                                                                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                                                                >
                                                                    <Sparkles className="h-4 w-4 text-cortex-primary" />
                                                                    Change Response Style for this Agent Type
                                                                </button>
                                                                {!profile.inherits_default_response_contract && (
                                                                    <button
                                                                        type="button"
                                                                        onClick={() => onRevertAgentTypeResponseContractSelection(item.id, profile.id)}
                                                                        disabled={isAgentTypeResponsePending}
                                                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                                                    >
                                                                        {isAgentTypeResponsePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                                                        Use Organization / Team Default
                                                                    </button>
                                                                )}
                                                            </div>

                                                            {isAgentTypeEditing && (
                                                                <AgentTypeAIEngineSelectionPanel
                                                                    selectedAgentTypeAIEngineProfile={selectedAgentTypeAIEngineProfile}
                                                                    agentTypeAIEngineUpdatePending={isAgentTypePending}
                                                                    agentTypeAIEngineUpdateError={agentTypeError}
                                                                    onAgentTypeAIEngineProfileSelect={onAgentTypeAIEngineProfileSelect}
                                                                    onClose={onCloseAgentTypeAIEngineSelector}
                                                                    onSubmit={() => onSubmitAgentTypeAIEngineSelection(item.id, profile.id)}
                                                                />
                                                            )}

                                                            {isAgentTypeResponseEditing && (
                                                                <AgentTypeResponseContractSelectionPanel
                                                                    selectedAgentTypeResponseContractProfile={selectedAgentTypeResponseContractProfile}
                                                                    agentTypeResponseContractUpdatePending={isAgentTypeResponsePending}
                                                                    agentTypeResponseContractUpdateError={agentTypeResponseError}
                                                                    onAgentTypeResponseContractProfileSelect={onAgentTypeResponseContractProfileSelect}
                                                                    onClose={onCloseAgentTypeResponseContractSelector}
                                                                    onSubmit={() => onSubmitAgentTypeResponseContractSelection(item.id, profile.id)}
                                                                />
                                                            )}
                                                        </div>
                                                    );
                                                })}
                                            </div>
                                        ) : (
                                            <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                                                Agent Type Profiles will appear here as this Team starts assigning specialist role types.
                                            </div>
                                        )}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                ) : (
                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                        {`${teamLeadName} can still shape the first operating lane before Departments are added.`}
                    </div>
                )
            ) : view === "automations" ? (
                <AutomationDetailPanel
                    items={automationItems}
                    loading={automationsLoading}
                    error={automationsError}
                    onRetry={onRetryAutomations}
                />
            ) : items.length > 0 ? (
                <div className="mt-5 grid gap-3">
                    {items.map((item) => (
                        <div key={item.key} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                            <p className="text-sm font-semibold text-cortex-text-main">{item.title}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.detail}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.support}</p>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                    {emptyStateMessage}
                </div>
            )}

            {view === "aiEngine" && isAIEngineSelectorOpen && (
                <AIEngineSelectionPanel
                    selectedAIEngineProfile={selectedAIEngineProfile}
                    aiEngineUpdatePending={aiEngineUpdatePending}
                    aiEngineUpdateError={aiEngineUpdateError}
                    onAIEngineProfileSelect={onAIEngineProfileSelect}
                    onClose={onCloseAIEngineSelector}
                    onSubmit={onSubmitAIEngineSelection}
                />
            )}

            {view === "aiEngine" && (
                <OutputModelRoutingPanel
                    routing={outputModelRouting}
                    loading={outputModelRoutingLoading}
                    error={outputModelRoutingError}
                    selectedRoutingMode={selectedOutputModelRoutingMode}
                    selectedDefaultModelId={selectedDefaultOutputModelId}
                    selectedBindings={selectedOutputModelBindings}
                    updatePending={outputModelRoutingUpdatePending}
                    updateError={outputModelRoutingUpdateError}
                    onRoutingModeChange={onOutputModelRoutingModeChange}
                    onDefaultModelChange={onDefaultOutputModelChange}
                    onBindingChange={onOutputModelBindingChange}
                    onSubmit={onSubmitOutputModelRouting}
                />
            )}

            {view === "responseContract" && isResponseContractSelectorOpen && (
                <ResponseContractSelectionPanel
                    selectedResponseContractProfile={selectedResponseContractProfile}
                    responseContractUpdatePending={responseContractUpdatePending}
                    responseContractUpdateError={responseContractUpdateError}
                    onResponseContractProfileSelect={onResponseContractProfileSelect}
                    onClose={onCloseResponseContractSelector}
                    onSubmit={onSubmitResponseContractSelection}
                />
            )}
        </div>
    );
}

function outputModelOptions(routing: OrganizationOutputModelRoutingPayload | null) {
    const options = routing?.available_models ?? [];
    return options.length > 0 ? options : [];
}

function OutputModelRoutingPanel({
    routing,
    loading,
    error,
    selectedRoutingMode,
    selectedDefaultModelId,
    selectedBindings,
    updatePending,
    updateError,
    onRoutingModeChange,
    onDefaultModelChange,
    onBindingChange,
    onSubmit,
}: {
    routing: OrganizationOutputModelRoutingPayload | null;
    loading: boolean;
    error: string | null;
    selectedRoutingMode: OrganizationOutputModelRoutingMode;
    selectedDefaultModelId: string;
    selectedBindings: Record<string, string>;
    updatePending: boolean;
    updateError: string | null;
    onRoutingModeChange: (mode: OrganizationOutputModelRoutingMode) => void;
    onDefaultModelChange: (modelId: string) => void;
    onBindingChange: (outputTypeId: OrganizationOutputTypeId, modelId: string) => void;
    onSubmit: () => void;
}) {
    const modelOptions = outputModelOptions(routing);

    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Output model routing</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Admins can keep every Team Lead and Specialist on one concrete model by default, or let Soma route detected output types like research, code, and vision to more appropriate self-hosted models.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Local-first routing</p>
                    <p className="mt-1">{routing?.hardware_summary ?? "Model routing data is still loading."}</p>
                </div>
            </div>

            {loading ? (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    Loading local model inventory...
                </div>
            ) : error ? (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to load output model routing</p>
                    <p className="mt-2 leading-6">{error}</p>
                </div>
            ) : (
                <>
                    <div className="mt-5 grid gap-3 lg:grid-cols-2">
                        <button
                            type="button"
                            onClick={() => onRoutingModeChange("single_model")}
                            disabled={updatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                selectedRoutingMode === "single_model"
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            }`}
                        >
                            <p className="text-sm font-semibold text-cortex-text-main">Use one model for everyone</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                Keeps every Team Lead and Specialist on the same concrete model until you intentionally split by detected output type.
                            </p>
                        </button>
                        <button
                            type="button"
                            onClick={() => onRoutingModeChange("detected_output_types")}
                            disabled={updatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                selectedRoutingMode === "detected_output_types"
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            }`}
                        >
                            <p className="text-sm font-semibold text-cortex-text-main">Use detected models by output type</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                Lets Soma keep the team simple while routing research, code, and vision-heavy work to more appropriate local models.
                            </p>
                        </button>
                    </div>

                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <label className="text-sm font-semibold text-cortex-text-main" htmlFor="organization-default-output-model">
                            Organization default model
                        </label>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            This is the fallback model all roles use by default. In detected mode, output types only diverge where you explicitly bind a different model.
                        </p>
                        <select
                            id="organization-default-output-model"
                            value={selectedDefaultModelId}
                            onChange={(event) => onDefaultModelChange(event.target.value)}
                            disabled={updatePending}
                            className="mt-3 w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main"
                        >
                            {modelOptions.map((option) => (
                                <option key={option.model_id} value={option.model_id}>
                                    {option.label} ({option.model_id})
                                </option>
                            ))}
                        </select>
                    </div>

                    {selectedRoutingMode === "detected_output_types" && (
                        <div className="mt-5 grid gap-3">
                            {OUTPUT_TYPE_OPTIONS.map((option) => (
                                <div key={option.id} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                    <select
                                        value={selectedBindings[option.id] ?? selectedDefaultModelId}
                                        onChange={(event) => onBindingChange(option.id, event.target.value)}
                                        disabled={updatePending}
                                        className="mt-3 w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main"
                                    >
                                        {modelOptions.map((entry) => (
                                            <option key={`${option.id}-${entry.model_id}`} value={entry.model_id}>
                                                {entry.label} ({entry.model_id})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            ))}
                        </div>
                    )}

                    {(routing?.recommended_models?.length ?? 0) > 0 && (
                        <div className="mt-5">
                            <p className="text-sm font-semibold text-cortex-text-main">Popular self-hosted starting points</p>
                            <div className="mt-3 grid gap-3 lg:grid-cols-2">
                                {(routing?.recommended_models ?? []).map((entry) => (
                                    <div key={entry.model_id} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <p className="text-sm font-semibold text-cortex-text-main">{entry.label}</p>
                                            {entry.installed && (
                                                <span className="inline-flex rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-cortex-primary">
                                                    Installed
                                                </span>
                                            )}
                                        </div>
                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{entry.summary}</p>
                                        {entry.hosting_fit && <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{entry.hosting_fit}</p>}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <p className="text-sm font-semibold text-cortex-text-main">Soma model review guardrail</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            {routing?.review_permission_prompt ??
                                "Ask the owner/admin before Soma reviews potential model behavior for a requested output or changes saved routing."}
                        </p>
                        {(routing?.automatic_selection_criteria?.length ?? 0) > 0 && (
                            <div className="mt-3 grid gap-2">
                                {(routing?.automatic_selection_criteria ?? []).map((criterion) => (
                                    <p key={criterion} className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm leading-6 text-cortex-text-muted">
                                        {criterion}
                                    </p>
                                ))}
                            </div>
                        )}
                    </div>

                    {(routing?.review_candidates?.length ?? 0) > 0 && (
                        <div className="mt-5">
                            <p className="text-sm font-semibold text-cortex-text-main">Behavior review candidates</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                These are Soma's first-pass local candidates when the admin has not pinned a model for a specific output type yet.
                            </p>
                            <div className="mt-3 grid gap-3 lg:grid-cols-2">
                                {(routing?.review_candidates ?? []).map((candidate) => (
                                    <div key={`${candidate.output_type_id}-${candidate.model_id}`} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <p className="text-sm font-semibold text-cortex-text-main">
                                                {candidate.output_type_label}: {candidate.model_summary}
                                            </p>
                                            {candidate.installed && (
                                                <span className="inline-flex rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-cortex-primary">
                                                    Installed
                                                </span>
                                            )}
                                        </div>
                                        <p className="mt-2 text-xs font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{candidate.model_id}</p>
                                        <div className="mt-3 grid gap-2">
                                            {candidate.review_criteria.map((criterion) => (
                                                <p key={criterion} className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm leading-6 text-cortex-text-muted">
                                                    {criterion}
                                                </p>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {updateError && (
                        <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">Unable to update output model routing</p>
                            <p className="mt-2 leading-6">{updateError}</p>
                        </div>
                    )}

                    <div className="mt-5 flex flex-wrap gap-3">
                        <button
                            type="button"
                            onClick={onSubmit}
                            disabled={!selectedDefaultModelId || updatePending}
                            className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                        >
                            {updatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                            {updatePending ? "Updating output model routing..." : "Use output model routing"}
                        </button>
                    </div>
                </>
            )}
        </div>
    );
}

function AIEngineSelectionPanel({
    selectedAIEngineProfile,
    aiEngineUpdatePending,
    aiEngineUpdateError,
    onAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAIEngineProfile: OrganizationAIEngineProfileId | null;
    aiEngineUpdatePending: boolean;
    aiEngineUpdateError: string | null;
    onAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine profile</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one guided AI Engine profile for the whole AI Organization. This tunes how Soma plans, responds, and carries work forward.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Organization level only</p>
                    <p className="mt-1">Department and role tuning stay in Department details so this organization-wide choice stays clear.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAIEngineProfileSelect(option.id)}
                            disabled={aiEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            } ${aiEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Current choice
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Good for:</span> {option.goodFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {aiEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update AI Engine Settings</p>
                    <p className="mt-2 leading-6">{aiEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to apply the selected AI Engine profile while keeping this Soma workspace in place.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAIEngineProfile || aiEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {aiEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {aiEngineUpdatePending ? "Updating AI Engine..." : "Use selected AI Engine"}
                </button>
                {aiEngineUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedAIEngineProfile || aiEngineUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry AI Engine change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={aiEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

function ResponseContractSelectionPanel({
    selectedResponseContractProfile,
    responseContractUpdatePending,
    responseContractUpdateError,
    onResponseContractProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedResponseContractProfile: ResponseContractProfileId | null;
    responseContractUpdatePending: boolean;
    responseContractUpdateError: string | null;
    onResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose a Response Style</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one safe Response Style for the whole AI Organization. This shapes tone, structure, and detail without exposing raw prompt or policy text.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Organization level only</p>
                    <p className="mt-1">Department and role Response Style tuning stays in Department details so this organization-wide default stays easy to understand.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {RESPONSE_CONTRACT_OPTIONS.map((option) => {
                    const isSelected = selectedResponseContractProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onResponseContractProfileSelect(option.id)}
                            disabled={responseContractUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            } ${responseContractUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        <span className="font-medium text-cortex-text-main">Tone:</span> {option.toneStyle}
                                    </p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Current choice
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Structure:</span> {option.structure}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Detail:</span> {option.verbosity}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Best for:</span> {option.bestFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {responseContractUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update Response Style</p>
                    <p className="mt-2 leading-6">{responseContractUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep the Soma workspace visible while the organization-wide response style is updated.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedResponseContractProfile || responseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {responseContractUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {responseContractUpdatePending ? "Updating Response Style..." : "Use selected Response Style"}
                </button>
                {responseContractUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedResponseContractProfile || responseContractUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Response Style change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={responseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

function DepartmentAIEngineSelectionPanel({
    selectedDepartmentAIEngineProfile,
    departmentAIEngineUpdatePending,
    departmentAIEngineUpdateError,
    onDepartmentAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedDepartmentAIEngineProfile: OrganizationAIEngineProfileId | null;
    departmentAIEngineUpdatePending: boolean;
    departmentAIEngineUpdateError: string | null;
    onDepartmentAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine for this Team</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick a guided AI Engine only for this Team when it needs a different planning or response style than the rest of the AI Organization.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Department-level only</p>
                    <p className="mt-1">Soma and the other Departments keep their current AI Engine behavior.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedDepartmentAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onDepartmentAIEngineProfileSelect(option.id)}
                            disabled={departmentAIEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${departmentAIEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Selected
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Good for:</span> {option.goodFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {departmentAIEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Team AI Engine</p>
                    <p className="mt-2 leading-6">{departmentAIEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this Department aligned with the Soma workspace while the rest of the organization stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedDepartmentAIEngineProfile || departmentAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {departmentAIEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {departmentAIEngineUpdatePending ? "Updating Team AI Engine..." : "Use selected AI Engine"}
                </button>
                <button
                    type="button"
                    onClick={onClose}
                    disabled={departmentAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

function AgentTypeAIEngineSelectionPanel({
    selectedAgentTypeAIEngineProfile,
    agentTypeAIEngineUpdatePending,
    agentTypeAIEngineUpdateError,
    onAgentTypeAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAgentTypeAIEngineProfile: OrganizationAIEngineProfileId | null;
    agentTypeAIEngineUpdatePending: boolean;
    agentTypeAIEngineUpdateError: string | null;
    onAgentTypeAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine for this Agent Type</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick a guided AI Engine only for this Agent Type when this role needs a consistent specialist behavior that should not drift with the Team default.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Agent Type level only</p>
                    <p className="mt-1">The Soma workspace stays visible, and Team-default inheritance remains in place for every other role type.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedAgentTypeAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAgentTypeAIEngineProfileSelect(option.id)}
                            disabled={agentTypeAIEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${agentTypeAIEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Selected
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Good for:</span> {option.goodFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {agentTypeAIEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Agent Type AI Engine</p>
                    <p className="mt-2 leading-6">{agentTypeAIEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this specialist role aligned with the Department while the Soma workspace stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAgentTypeAIEngineProfile || agentTypeAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {agentTypeAIEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {agentTypeAIEngineUpdatePending ? "Updating Agent Type AI Engine..." : "Use selected AI Engine"}
                </button>
                <button
                    type="button"
                    onClick={onClose}
                    disabled={agentTypeAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

function AgentTypeResponseContractSelectionPanel({
    selectedAgentTypeResponseContractProfile,
    agentTypeResponseContractUpdatePending,
    agentTypeResponseContractUpdateError,
    onAgentTypeResponseContractProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAgentTypeResponseContractProfile: ResponseContractProfileId | null;
    agentTypeResponseContractUpdatePending: boolean;
    agentTypeResponseContractUpdateError: string | null;
    onAgentTypeResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose a Response Style for this Agent Type</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one safe Response Style only for this Agent Type when this role needs a consistent specialist voice that should stay steady even if the wider default changes.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Agent Type level only</p>
                    <p className="mt-1">The Soma workspace stays visible, and Organization / Team default behavior remains in place for every other role type.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {RESPONSE_CONTRACT_OPTIONS.map((option) => {
                    const isSelected = selectedAgentTypeResponseContractProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAgentTypeResponseContractProfileSelect(option.id)}
                            disabled={agentTypeResponseContractUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${agentTypeResponseContractUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        <span className="font-medium text-cortex-text-main">Tone:</span> {option.toneStyle}
                                    </p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Selected
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Structure:</span> {option.structure}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Detail:</span> {option.verbosity}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Best for:</span> {option.bestFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {agentTypeResponseContractUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Agent Type Response Style</p>
                    <p className="mt-2 leading-6">{agentTypeResponseContractUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this specialist role aligned with the Soma workspace while the wider organization stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAgentTypeResponseContractProfile || agentTypeResponseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {agentTypeResponseContractUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {agentTypeResponseContractUpdatePending ? "Updating Agent Type Response Style..." : "Use selected Response Style"}
                </button>
                {agentTypeResponseContractUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedAgentTypeResponseContractProfile || agentTypeResponseContractUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Response Style change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={agentTypeResponseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

function GuidedWorkspaceCard({
    eyebrow,
    title,
    summary,
    buttonLabel,
    onClick,
}: GuidedWorkspaceCardDefinition) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">{eyebrow}</p>
            <p className="mt-2 text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
            <button
                type="button"
                onClick={onClick}
                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 hover:text-cortex-primary"
            >
                {buttonLabel}
                <ArrowRight className="h-4 w-4" />
            </button>
        </div>
    );
}

