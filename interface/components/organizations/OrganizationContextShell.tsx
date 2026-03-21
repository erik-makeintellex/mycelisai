"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Activity, ArrowLeft, Blocks, Bot, BrainCircuit, Building2, Loader2, RefreshCcw, Sparkles, Users } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
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
    OrganizationLearningInsightItem,
    OrganizationLoopActivityItem,
    ResponseContractProfileId,
    ResponseContractUpdateRequest,
} from "@/lib/organizations";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

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
    const [recentActivity, setRecentActivity] = useState<OrganizationLoopActivityItem[]>([]);
    const [activityLoading, setActivityLoading] = useState(true);
    const [activityError, setActivityError] = useState<string | null>(null);
    const [activityReloadToken, setActivityReloadToken] = useState(0);
    const [automations, setAutomations] = useState<OrganizationAutomationItem[]>([]);
    const [automationsLoading, setAutomationsLoading] = useState(true);
    const [automationsError, setAutomationsError] = useState<string | null>(null);
    const [automationsReloadToken, setAutomationsReloadToken] = useState(0);
    const [learningInsights, setLearningInsights] = useState<OrganizationLearningInsightItem[]>([]);
    const [learningInsightsLoading, setLearningInsightsLoading] = useState(true);
    const [learningInsightsError, setLearningInsightsError] = useState<string | null>(null);
    const [learningInsightsReloadToken, setLearningInsightsReloadToken] = useState(0);

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

        const loadActivity = async (background: boolean) => {
            if (!background && !cancelled) {
                setActivityLoading(true);
            }

            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/loop-activity`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Activity unavailable");
                }
                if (cancelled) {
                    return;
                }
                setRecentActivity(extractApiData<OrganizationLoopActivityItem[]>(payload) ?? []);
                setActivityError(null);
            } catch {
                if (cancelled) {
                    return;
                }
                setActivityError("Activity unavailable");
            } finally {
                if (!cancelled) {
                    setActivityLoading(false);
                }
            }
        };

        void loadActivity(false);
        const intervalId = window.setInterval(() => {
            void loadActivity(true);
        }, 15000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, activityReloadToken]);

    useEffect(() => {
        let cancelled = false;

        const loadAutomations = async (background: boolean) => {
            if (!background && !cancelled) {
                setAutomationsLoading(true);
            }

            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/automations`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Automations unavailable");
                }
                if (cancelled) {
                    return;
                }
                setAutomations(extractApiData<OrganizationAutomationItem[]>(payload) ?? []);
                setAutomationsError(null);
            } catch {
                if (cancelled) {
                    return;
                }
                setAutomationsError("Automations unavailable");
            } finally {
                if (!cancelled) {
                    setAutomationsLoading(false);
                }
            }
        };

        void loadAutomations(false);
        const intervalId = window.setInterval(() => {
            void loadAutomations(true);
        }, 20000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, automationsReloadToken]);

    useEffect(() => {
        let cancelled = false;

        const loadLearningInsights = async (background: boolean) => {
            if (!background && !cancelled) {
                setLearningInsightsLoading(true);
            }

            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/learning-insights`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Learning updates unavailable");
                }
                if (cancelled) {
                    return;
                }
                setLearningInsights(extractApiData<OrganizationLearningInsightItem[]>(payload) ?? []);
                setLearningInsightsError(null);
            } catch {
                if (cancelled) {
                    return;
                }
                setLearningInsightsError("Learning updates unavailable");
            } finally {
                if (!cancelled) {
                    setLearningInsightsLoading(false);
                }
            }
        };

        void loadLearningInsights(false);
        const intervalId = window.setInterval(() => {
            void loadLearningInsights(true);
        }, 25000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, learningInsightsReloadToken]);

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
    }, [activeDetailView]);

    const openAIEngineSelector = () => {
        if (!organization) {
            return;
        }
        setActiveDetailView("aiEngine");
        setSelectedAIEngineProfile((organization.ai_engine_profile_id as OrganizationAIEngineProfileId | undefined) ?? null);
        setAIEngineUpdateError(null);
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
    const overviewItems = [
        { label: "Started from", value: organization.start_mode === "template" ? (organization.template_name || "Template") : "Empty" },
        { label: "Advisors", value: formatConfiguredCount(organization.advisor_count, "Advisor") },
        { label: "Departments", value: formatConfiguredCount(organization.department_count, "Department") },
        { label: "Specialists", value: formatConfiguredCount(organization.specialist_count, "Specialist") },
        { label: "AI Organization", value: toTitleCase(organization.status) },
    ];

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_24px_60px_rgba(0,0,0,0.18)]">
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
                        </div>
                    </div>
                </section>

                <section className="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
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
                                <p className="text-sm font-medium text-cortex-text-main">What I can help with</p>
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
                                onRetryAutomations={() => setAutomationsReloadToken((value) => value + 1)}
                                isAIEngineSelectorOpen={isAIEngineSelectorOpen}
                                selectedAIEngineProfile={selectedAIEngineProfile}
                                aiEngineUpdatePending={aiEngineUpdatePending}
                                aiEngineUpdateError={aiEngineUpdateError}
                                onAIEngineProfileSelect={setSelectedAIEngineProfile}
                                onOpenAIEngineSelector={openAIEngineSelector}
                                onCloseAIEngineSelector={() => {
                                    setIsAIEngineSelectorOpen(false);
                                    setAIEngineUpdateError(null);
                                }}
                                onSubmitAIEngineSelection={submitAIEngineSelection}
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

                        <TeamLeadInteractionPanel
                            organizationId={organization.id}
                            organizationName={organization.name}
                            somaName={somaName}
                            teamLeadName={teamLeadName}
                        />
                    </div>

                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div>
                                <h2 className="text-xl font-semibold text-cortex-text-main">Organization overview</h2>
                                <p className="mt-1 text-sm text-cortex-text-muted">See Soma, the operational structure, and the starting point for this AI Organization at a glance.</p>
                            </div>
                            <div className="mt-4 grid gap-3 md:grid-cols-2">
                                {overviewItems.map((item) => (
                                    <Metric key={item.label} label={item.label} value={item.value} />
                                ))}
                            </div>
                        </div>

                        <InspectOnlySummary
                            icon={<Users className="h-4 w-4" />}
                            title="Advisors"
                            countLabel={formatConfiguredCount(organization.advisor_count, "Advisor")}
                            summary={advisorSummary(organization.advisor_count, teamLeadName)}
                            supportLabel="Advisor support"
                            items={advisorSupportItems(organization.advisor_count)}
                            inspectActionLabel="Review Advisors"
                            onInspect={() => setActiveDetailView("advisors")}
                        />

                        <InspectOnlySummary
                            icon={<Building2 className="h-4 w-4" />}
                            title="Departments"
                            countLabel={formatConfiguredCount(organization.department_count, "Department")}
                            summary={departmentSummary(organization.department_count, organization.specialist_count, teamLeadName)}
                            supportLabel="Department view"
                            items={departmentSupportItems(organization)}
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
                            inspectActionLabel="Review Automations"
                            onInspect={() => setActiveDetailView("automations")}
                        />

                        <RecentActivityPanel
                            items={recentActivity}
                            loading={activityLoading}
                            error={activityError}
                            onRetry={() => setActivityReloadToken((value) => value + 1)}
                        />

                        <LearningVisibilityPanel
                            items={learningInsights}
                            loading={learningInsightsLoading}
                            error={learningInsightsError}
                            onRetry={() => setLearningInsightsReloadToken((value) => value + 1)}
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
                                    inspectActionLabel="Review Response Style"
                                    onInspect={() => setActiveDetailView("responseContract")}
                                />
                                <InspectOnlySummary
                                    icon={<BrainCircuit className="h-4 w-4" />}
                                    title="Learning & Context"
                                    countLabel="Inspect only"
                                    summary={learningContextSummary(organization.memory_personality_summary)}
                                    supportLabel="What this affects"
                                    items={learningContextSupportItems(organization.memory_personality_summary)}
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
            </div>
        </div>
    );
}

function formatConfiguredCount(count: number, label: string) {
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${label}${count === 1 ? "" : "s"}`;
}

function advisorSummary(count: number, teamLeadName: string) {
    if (count === 0) {
        return `Soma is handling planning and review directly for now. Advisor support will appear here when ${teamLeadName} needs a second set of eyes.`;
    }
    if (count === 1) {
        return `1 Advisor is ready to help Soma and ${teamLeadName} with review, priorities, and decision support.`;
    }
    return `${count} Advisors are ready to help Soma and ${teamLeadName} review decisions and keep the organization aligned.`;
}

function advisorSupportItems(count: number) {
    if (count === 0) {
        return [
            "Review support appears here",
            "Advisors help with decisions and checks",
            "Try reviewing your organization setup",
        ];
    }
    return ["Planning review", "Decision support", "Priority checks"].slice(0, Math.max(2, Math.min(count + 1, 3)));
}

function departmentSummary(count: number, specialistCount: number, teamLeadName: string) {
    if (count === 0) {
        return `Soma can still shape the first working lane through ${teamLeadName} before Departments are configured. Departments will appear here once the organization has a clear first focus.`;
    }
    return `${count} Departments and ${formatConfiguredCount(specialistCount, "Specialist").toLowerCase()} are visible here so Soma and ${teamLeadName} can work with a clear delivery structure.`;
}

function departmentSupportItems(organization: OrganizationHomePayload) {
    const items = [
        organization.start_mode === "template" && organization.template_name
            ? `Started from ${organization.template_name}`
            : "Started from Empty",
        formatConfiguredCount(organization.specialist_count, "Specialist"),
        organization.department_count > 0 ? "Open the current team structure" : "Try reviewing your organization setup",
    ];
    return items;
}

function formatAutomationCount(count: number, loading: boolean, error: string | null) {
    if (error) {
        return "Unavailable right now";
    }
    if (loading && count === 0) {
        return "Checking now";
    }
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${count === 1 ? "Automation" : "Automations"}`;
}

function automationStatusLabel(loading: boolean, error: string | null) {
    if (error) {
        return "Read only";
    }
    if (loading) {
        return "Refreshing";
    }
    return "Read only";
}

function automationSummary(count: number, teamLeadName: string) {
    const mentalModel = "This system runs ongoing reviews and checks to help your organization improve over time.";
    if (count === 0) {
        return `${mentalModel} Soma will show those ongoing reviews and checks here as this AI Organization becomes more active.`;
    }
    if (count === 1) {
        return `${mentalModel} 1 Automation is visible here so Soma can explain what ongoing review is supporting this AI Organization through ${teamLeadName}.`;
    }
    return `${mentalModel} ${count} Automations are visible here so Soma can explain what ongoing reviews and checks are supporting this AI Organization through ${teamLeadName}.`;
}

function automationSupportItems(items: OrganizationAutomationItem[], loading: boolean, error: string | null) {
    if (error) {
        return ["Automations unavailable", "Workspace still ready", "Read only"];
    }
    if (loading && items.length === 0) {
        return ["Checking Reviews", "Checking Watchers", "Read only"];
    }
    if (items.length === 0) {
        return ["Reviews appear here", "Watchers appear here", "Read only"];
    }

    return items.slice(0, 3).map((item) => {
        if (item.trigger_type === "scheduled") {
            return `${item.name} • Scheduled`;
        }
        return `${item.name} • Event-driven`;
    });
}

function advisorDetailItems(count: number) {
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

function departmentDetailItems(organization: OrganizationHomePayload) {
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

function departmentPurpose(name: string) {
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

function agentTypeAIEngineSourceLabel(profile: OrganizationAgentTypeProfileSummary) {
    return profile.inherits_department_ai_engine ? `Using Team Default: ${profile.ai_engine_effective_summary}` : `Type-specific Engine: ${profile.ai_engine_effective_summary}`;
}

function agentTypeResponseStyleSourceLabel(profile: OrganizationAgentTypeProfileSummary) {
    return profile.inherits_default_response_contract
        ? `Using Organization or Team Default: ${profile.response_contract_effective_summary}`
        : `Type-specific Response Style: ${profile.response_contract_effective_summary}`;
}

function agentTypeSelectionKey(departmentId: string, agentTypeId: string) {
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

function generateFallbackDepartmentSummaries(
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

function aiEngineSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.";
    }
    return `The current AI Engine Settings profile is ${normalized.toLowerCase()} and shapes how the organization responds, plans, and carries work forward.`;
}

function aiEngineSupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Response style", "Planning depth", "Inspect only for now"];
    }
    return [summary, "Response style", "Planning depth"];
}

function aiEngineDetailItems(organization: OrganizationHomePayload) {
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

function responseContractSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized) {
        return "The current Response Style keeps the organization on a clear, steady default for safe day-to-day guidance.";
    }
    return `The current Response Style is ${normalized.toLowerCase()}, which shapes how Soma presents tone, structure, and detail.`;
}

function responseContractSupportItems(summary: string) {
    if (!summary.trim()) {
        return ["Tone", "Structure", "Verbosity"];
    }
    return [summary, "Tone", "Structure", "Verbosity"];
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

function learningContextSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "Learning & Context stay on a simple starter posture so Soma keeps a steady working style while the organization gets established.";
    }
    return `Learning & Context are currently ${normalized.toLowerCase()}, which shapes how Soma carries context forward and keeps guidance grounded.`;
}

function learningContextSupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Learning visibility", "Context continuity", "Inspect only for now"];
    }
    return [summary, "Learning visibility", "Context continuity"];
}

function toTitleCase(value: string) {
    return value
        .split(/[\s_-]+/)
        .filter(Boolean)
        .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
        .join(" ");
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
    inspectActionLabel,
    onInspect,
    statusLabel = "Inspect only",
}: {
    icon: React.ReactNode;
    title: string;
    countLabel: string;
    summary: string;
    supportLabel: string;
    items: string[];
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
                <p className="text-sm font-medium text-cortex-text-main">{supportLabel}</p>
                <div className="mt-3 flex flex-wrap gap-2">
                    {items.map((item) => (
                        <div key={item} className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
                            {item}
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
    onAIEngineProfileSelect,
    onOpenAIEngineSelector,
    onCloseAIEngineSelector,
    onSubmitAIEngineSelection,
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
    onAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onOpenAIEngineSelector: () => void;
    onCloseAIEngineSelector: () => void;
    onSubmitAIEngineSelection: () => void;
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
                                                            <div className="mt-4 grid gap-3 lg:grid-cols-2">
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">AI Engine</p>
                                                                    <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeAIEngineSourceLabel(profile)}</p>
                                                                </div>
                                                                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Response Style</p>
                                                                    <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeResponseStyleSourceLabel(profile)}</p>
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

function AutomationDetailPanel({
    items,
    loading,
    error,
    onRetry,
}: {
    items: OrganizationAutomationItem[];
    loading: boolean;
    error: string | null;
    onRetry: () => void;
}) {
    const guidanceText = "This system runs ongoing reviews and checks to help your organization improve over time.";

    if (error) {
        return (
            <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                <p className="font-medium text-cortex-text-main">How Automations help</p>
                <p className="mt-2 leading-6">{guidanceText}</p>
                <p className="font-medium text-cortex-text-main">Automations unavailable</p>
                <p className="mt-2 leading-6">Reviews and checks are temporarily unavailable here. The Soma workspace is still ready.</p>
                <button
                    type="button"
                    onClick={onRetry}
                    className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                >
                    <RefreshCcw className="h-4 w-4" />
                    Retry Automations
                </button>
            </div>
        );
    }

    if (loading && items.length === 0) {
        return (
            <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                <p className="font-medium text-cortex-text-main">How Automations help</p>
                <p className="mt-2 leading-6">{guidanceText}</p>
                <p className="font-medium text-cortex-text-main">Checking active reviews</p>
                <p className="mt-2 leading-6">The latest Automations will appear here shortly.</p>
            </div>
        );
    }

    if (items.length === 0) {
        return (
            <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                <p className="font-medium text-cortex-text-main">How Automations help</p>
                <p className="mt-2 leading-6">{guidanceText}</p>
                <p className="font-medium text-cortex-text-main">No Automations visible yet</p>
                <p className="mt-2 leading-6">Reviews and checks will appear here as this AI Organization begins operating.</p>
                <p className="mt-2 leading-6">Try reviewing your organization setup or running a quick strategy check to create the first visible signals.</p>
            </div>
        );
    }

    return (
        <div className="mt-5 grid gap-4">
            <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                <p className="font-medium text-cortex-text-main">How Automations help</p>
                <p className="mt-2 leading-6">{guidanceText}</p>
            </div>
            {items.map((item) => (
                <div key={item.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                        <div>
                            <div className="flex flex-wrap items-center gap-2">
                                <p className="text-sm font-semibold text-cortex-text-main">{item.name}</p>
                                <ActivityStatusBadge status={item.status} />
                            </div>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.purpose}</p>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                            <p className="font-medium text-cortex-text-main">{automationTriggerLabel(item.trigger_type)}</p>
                            <p className="mt-1 leading-6">{item.owner_label}</p>
                        </div>
                    </div>

                    <div className="mt-4 grid gap-3 lg:grid-cols-2">
                        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">What it watches</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.watches}</p>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">How it runs</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.trigger_summary}</p>
                        </div>
                    </div>

                    <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <p className="text-sm font-semibold text-cortex-text-main">Recent outcomes</p>
                        {item.recent_outcomes && item.recent_outcomes.length > 0 ? (
                            <div className="mt-3 space-y-3">
                                {item.recent_outcomes.map((outcome) => (
                                    <div key={`${item.id}-${outcome.occurred_at}-${outcome.summary}`} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
                                        <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                                            <p className="text-sm leading-6 text-cortex-text-muted">{outcome.summary}</p>
                                            <p className="text-xs font-medium uppercase tracking-[0.14em] text-cortex-text-muted">
                                                {formatRelativeActivityTime(outcome.occurred_at)}
                                            </p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">This Automation is ready, but it has not reported a recent review yet.</p>
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
}

function automationTriggerLabel(triggerType: OrganizationAutomationItem["trigger_type"]) {
    return triggerType === "scheduled" ? "Scheduled" : "Event-driven";
}

function RecentActivityPanel({
    items,
    loading,
    error,
    onRetry,
}: {
    items: OrganizationLoopActivityItem[];
    loading: boolean;
    error: string | null;
    onRetry: () => void;
}) {
    const visibleItems = items.slice(0, 8);

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex items-start gap-3">
                <div className="mt-0.5 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 p-2 text-cortex-primary">
                    <Activity className="h-4 w-4" />
                </div>
                <div>
                    <h2 className="text-xl font-semibold text-cortex-text-main">Recent Activity</h2>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                        Your AI Organization is actively working through recent reviews, checks, and updates in the background.
                    </p>
                </div>
            </div>

            {error && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Activity unavailable</p>
                    <p className="mt-1 leading-6">Recent reviews and updates are not available right now. The Soma workspace is still ready.</p>
                    <button
                        type="button"
                        onClick={onRetry}
                        className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Recent Activity
                    </button>
                </div>
            )}

            {!error && loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Checking for recent updates</p>
                    <p className="mt-1 leading-6">The latest reviews and checks will appear here shortly.</p>
                </div>
            )}

            {!error && !loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">No recent activity yet</p>
                    <p className="mt-1 leading-6">This is where reviews, checks, and updates will appear as your AI Organization starts operating.</p>
                    <p className="mt-1 leading-6">Take a guided Soma action to start creating visible movement here.</p>
                </div>
            )}

            {visibleItems.length > 0 && (
                <div className="mt-4 space-y-3">
                    {visibleItems.map((item) => (
                        <div key={item.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                <div>
                                    <div className="flex flex-wrap items-center gap-2">
                                        <p className="text-sm font-semibold text-cortex-text-main">{item.name}</p>
                                        <ActivityStatusBadge status={item.status} />
                                    </div>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.summary}</p>
                                </div>
                                <p className="text-xs font-medium uppercase tracking-[0.14em] text-cortex-text-muted">{formatRelativeActivityTime(item.last_run_at)}</p>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

function LearningVisibilityPanel({
    items,
    loading,
    error,
    onRetry,
}: {
    items: OrganizationLearningInsightItem[];
    loading: boolean;
    error: string | null;
    onRetry: () => void;
}) {
    const visibleItems = items.slice(0, 6);

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex items-start gap-3">
                <div className="mt-0.5 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 p-2 text-cortex-primary">
                    <Sparkles className="h-4 w-4" />
                </div>
                <div>
                    <h2 className="text-xl font-semibold text-cortex-text-main">What the Organization is Learning</h2>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                        See the recurring improvements and themes your AI Organization is picking up across recent work, and why they matter for what happens next.
                    </p>
                </div>
            </div>

            {error && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Learning updates unavailable</p>
                    <p className="mt-1 leading-6">Recent learning highlights are not available right now. The Soma workspace is still ready.</p>
                    <button
                        type="button"
                        onClick={onRetry}
                        className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Learning
                    </button>
                </div>
            )}

            {!error && loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Checking recent learning</p>
                    <p className="mt-1 leading-6">The latest improvements and patterns will appear here shortly.</p>
                </div>
            )}

            {!error && !loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">No learning highlights yet</p>
                    <p className="mt-1 leading-6">This is where recurring patterns, improvements, and stronger working habits will appear in plain language.</p>
                    <p className="mt-1 leading-6">Use Soma guidance and early reviews to give the organization enough signal to learn from.</p>
                </div>
            )}

            {visibleItems.length > 0 && (
                <div className="mt-4 space-y-3">
                    {visibleItems.map((item) => (
                        <div key={item.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                <div>
                                    <div className="flex flex-wrap items-center gap-2">
                                        <p className="text-sm font-semibold text-cortex-text-main">{learningHeadline(item)}</p>
                                        <LearningStrengthBadge strength={item.strength} />
                                    </div>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.summary}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        <span className="font-medium text-cortex-text-main">Why it matters:</span> {learningWhyItMatters(item)}
                                    </p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.source}</p>
                                </div>
                                <p className="text-xs font-medium uppercase tracking-[0.14em] text-cortex-text-muted">{formatRelativeActivityTime(item.observed_at)}</p>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

function learningHeadline(item: OrganizationLearningInsightItem) {
    if (item.strength === "strong") {
        return "Consistently improving";
    }
    if (item.strength === "consistent") {
        return "Identifying recurring patterns";
    }
    return "Detecting new opportunities";
}

function learningWhyItMatters(item: OrganizationLearningInsightItem) {
    if (/^Team:/i.test(item.source)) {
        return "It gives Soma a clearer view of how this part of the organization is getting stronger over time.";
    }
    if (/role:/i.test(item.source)) {
        return "It helps Soma see where specialist support is becoming more reliable or where more guidance may be needed.";
    }
    return "It helps Soma turn repeated signals into clearer next steps for the organization.";
}

function ActivityStatusBadge({ status }: { status: OrganizationLoopActivityItem["status"] }) {
    const config =
        status === "warning"
            ? {
                  label: "Needs review",
                  className: "border-amber-500/30 bg-amber-500/10 text-amber-200",
              }
            : status === "failed"
              ? {
                    label: "Unavailable",
                    className: "border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger",
                }
              : {
                    label: "Ready",
                    className: "border-emerald-500/30 bg-emerald-500/10 text-emerald-200",
                };

    return (
        <span className={`inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.16em] ${config.className}`}>
            <span className="h-1.5 w-1.5 rounded-full bg-current" />
            {config.label}
        </span>
    );
}

function LearningStrengthBadge({ strength }: { strength: OrganizationLearningInsightItem["strength"] }) {
    const config =
        strength === "strong"
            ? {
                  label: "Strong",
                  className: "border-emerald-500/30 bg-emerald-500/10 text-emerald-200",
              }
            : strength === "consistent"
              ? {
                    label: "Consistent",
                    className: "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary",
                }
              : {
                    label: "Emerging",
                    className: "border-amber-500/30 bg-amber-500/10 text-amber-200",
                };

    return (
        <span className={`inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.16em] ${config.className}`}>
            <span className="h-1.5 w-1.5 rounded-full bg-current" />
            {config.label}
        </span>
    );
}

function formatRelativeActivityTime(timestamp: string) {
    const parsed = Date.parse(timestamp);
    if (Number.isNaN(parsed)) {
        return "Recently";
    }

    const diffMs = Math.max(0, Date.now() - parsed);
    const diffMinutes = Math.floor(diffMs / 60000);
    if (diffMinutes <= 0) {
        return "Just now";
    }
    if (diffMinutes === 1) {
        return "1 minute ago";
    }
    if (diffMinutes < 60) {
        return `${diffMinutes} minutes ago`;
    }

    const diffHours = Math.floor(diffMinutes / 60);
    if (diffHours === 1) {
        return "1 hour ago";
    }
    if (diffHours < 24) {
        return `${diffHours} hours ago`;
    }

    const diffDays = Math.floor(diffHours / 24);
    if (diffDays === 1) {
        return "1 day ago";
    }
    return `${diffDays} days ago`;
}
