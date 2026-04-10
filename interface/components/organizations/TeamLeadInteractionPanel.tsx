"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { ArrowRight, Loader2, RefreshCcw } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type { TeamLeadExecutionContract, TeamLeadGuidanceRequest, TeamLeadGuidanceResponse, TeamLeadGuidedAction, TeamLeadWorkflowGroupDraft } from "@/lib/organizations";

type RequestState = "idle" | "loading" | "ready" | "error";
type PersistedWorkspaceState = {
    draftPrompt?: string;
    selectedAction?: TeamLeadGuidedAction | null;
    requestState?: Extract<RequestState, "idle" | "ready">;
    requestContext?: string | null;
    guidance?: TeamLeadGuidanceResponse | null;
};

type LaunchedGroupState = {
    groupId: string;
    name: string;
};

export type SomaGuidanceUpdate = {
    requestLabel: string;
    summary: string;
    teamsEngaged: string[];
    outputs: string[];
    timestamp: string;
};

export type TeamLeadPromptSuggestion = {
    label: string;
    prompt: string;
};

function storageKeyForOrganization(organizationId: string) {
    return `mycelis-soma-workspace:${organizationId}`;
}

function readPersistedWorkspaceState(organizationId: string): PersistedWorkspaceState | null {
    if (typeof window === "undefined") {
        return null;
    }
    try {
        const raw = window.localStorage.getItem(storageKeyForOrganization(organizationId));
        if (!raw) {
            return null;
        }
        const parsed = JSON.parse(raw) as PersistedWorkspaceState;
        return typeof parsed === "object" && parsed !== null ? parsed : null;
    } catch {
        return null;
    }
}

function persistWorkspaceState(organizationId: string, state: PersistedWorkspaceState) {
    if (typeof window === "undefined") {
        return;
    }
    try {
        window.localStorage.setItem(storageKeyForOrganization(organizationId), JSON.stringify(state));
    } catch {
        // best-effort continuity only
    }
}

const GUIDED_ACTIONS: Array<{ action: TeamLeadGuidedAction; label: string; detail: string }> = [
    {
        action: "plan_next_steps",
        label: "Run a quick strategy check",
        detail: "See the next priorities Soma would set so the workspace starts moving in a visible direction.",
    },
    {
        action: "focus_first",
        label: "Choose the first priority",
        detail: "Ask Soma to identify the first move most likely to create visible progress across the workspace.",
    },
    {
        action: "review_setup",
        label: "Review your organization setup",
        detail: "Check whether Advisors, Departments, and Specialists are ready and see what Soma recommends inspecting next.",
    },
];

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export default function TeamLeadInteractionPanel({
    organizationId,
    organizationName,
    somaName,
    teamLeadName,
    autoFocusOnLoad = false,
    embedded = false,
    onGuidanceStateChange,
    promptSuggestions = [],
}: {
    organizationId: string;
    organizationName: string;
    somaName: string;
    teamLeadName: string;
    autoFocusOnLoad?: boolean;
    embedded?: boolean;
    onGuidanceStateChange?: (update: SomaGuidanceUpdate) => void;
    promptSuggestions?: TeamLeadPromptSuggestion[];
}) {
    const panelRef = useRef<HTMLElement | null>(null);
    const promptRef = useRef<HTMLTextAreaElement | null>(null);
    const [selectedAction, setSelectedAction] = useState<TeamLeadGuidedAction | null>(null);
    const [requestState, setRequestState] = useState<RequestState>("idle");
    const [error, setError] = useState<string | null>(null);
    const [guidance, setGuidance] = useState<TeamLeadGuidanceResponse | null>(null);
    const [draftPrompt, setDraftPrompt] = useState("");
    const [requestContext, setRequestContext] = useState<string | null>(null);
    const [hasHydratedState, setHasHydratedState] = useState(false);
    const [launchingGroup, setLaunchingGroup] = useState(false);
    const [launchError, setLaunchError] = useState<string | null>(null);
    const [launchedGroup, setLaunchedGroup] = useState<LaunchedGroupState | null>(null);
    const isLoading = requestState === "loading";

    useEffect(() => {
        const persisted = readPersistedWorkspaceState(organizationId);
        if (persisted) {
            setDraftPrompt(typeof persisted.draftPrompt === "string" ? persisted.draftPrompt : "");
            setSelectedAction(persisted.selectedAction ?? null);
            setRequestState(persisted.requestState === "ready" ? "ready" : "idle");
            setRequestContext(typeof persisted.requestContext === "string" ? persisted.requestContext : null);
            setGuidance(persisted.guidance ?? null);
        }
        setHasHydratedState(true);
    }, [organizationId]);

    useEffect(() => {
        if (!hasHydratedState) {
            return;
        }
        persistWorkspaceState(organizationId, {
            draftPrompt,
            selectedAction,
            requestState: requestState === "ready" ? "ready" : "idle",
            requestContext,
            guidance: requestState === "ready" ? guidance : null,
        });
    }, [organizationId, draftPrompt, selectedAction, requestState, requestContext, guidance, hasHydratedState]);

    useEffect(() => {
        if (!autoFocusOnLoad) {
            return;
        }
        panelRef.current?.scrollIntoView?.({ behavior: "smooth", block: "start" });
        window.setTimeout(() => {
            promptRef.current?.focus();
        }, 80);
    }, [autoFocusOnLoad]);

    const triggerAction = async (action: TeamLeadGuidedAction, contextLabel?: string | null) => {
        if (isLoading) {
            return;
        }
        setSelectedAction(action);
        setRequestState("loading");
        setError(null);
        setRequestContext(contextLabel ?? null);
        setLaunchError(null);
        setLaunchedGroup(null);

        const payload: TeamLeadGuidanceRequest = { action };
        if (contextLabel && contextLabel.trim().length > 0) {
            payload.request_context = contextLabel.trim();
        }
        try {
            const response = await fetch(`/api/v1/organizations/${organizationId}/workspace/actions`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });
            const responsePayload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(responsePayload) || "Team Lead guidance is unavailable right now.");
            }
            const rawGuidance = extractApiData<Partial<TeamLeadGuidanceResponse> | null>(responsePayload);
            const normalized = normalizeGuidanceResponse(rawGuidance, action, organizationName, somaName, teamLeadName);
            setGuidance(normalized);
            setRequestState("ready");
            onGuidanceStateChange?.({
                requestLabel: normalized.request_label,
                summary: normalized.summary,
                teamsEngaged: [teamLeadName],
                outputs: normalized.execution_contract?.target_outputs?.length
                    ? normalized.execution_contract.target_outputs
                    : ["Team design guidance"],
                timestamp: new Date().toISOString(),
            });
        } catch (err) {
            const message = err instanceof Error ? err.message : "Soma guidance is unavailable right now.";
            setError(rewriteGuidanceText(message, somaName, teamLeadName));
            setRequestState("error");
        }
    };

    const handlePromptSubmit = async () => {
        const normalizedPrompt = draftPrompt.trim();
        if (isLoading) {
            return;
        }
        await triggerAction(
            resolvePromptAction(normalizedPrompt),
            normalizedPrompt.length > 0 ? normalizedPrompt : null,
        );
    };

    const launchWorkflowGroup = async (draft: TeamLeadWorkflowGroupDraft) => {
        if (launchingGroup) {
            return;
        }
        setLaunchingGroup(true);
        setLaunchError(null);
        setLaunchedGroup(null);
        try {
            const expiry = typeof draft.expiry_hours === "number" && draft.expiry_hours > 0
                ? new Date(Date.now() + draft.expiry_hours * 60 * 60 * 1000).toISOString()
                : null;
            const response = await fetch("/api/v1/groups", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name: draft.name,
                    goal_statement: draft.goal_statement,
                    work_mode: draft.work_mode,
                    coordinator_profile: draft.coordinator_profile,
                    allowed_capabilities: draft.allowed_capabilities ?? [],
                    expiry,
                }),
            });
            const responsePayload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(responsePayload) || "Could not create the temporary workflow group.");
            }
            const created = extractApiData<{ group_id?: string; name?: string } | null>(responsePayload);
            if (!created?.group_id) {
                throw new Error("Could not create the temporary workflow group.");
            }
            setLaunchedGroup({
                groupId: created.group_id,
                name: created.name || draft.name,
            });
        } catch (groupError) {
            setLaunchError(groupError instanceof Error ? groupError.message : "Could not create the temporary workflow group.");
        } finally {
            setLaunchingGroup(false);
        }
    };

    return (
        <section
            id="soma-panel"
            ref={panelRef}
            className={embedded ? "space-y-5" : "rounded-3xl border border-cortex-border bg-cortex-surface p-6"}
        >
            <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                <div>
                    <h2 className="text-xl font-semibold text-cortex-text-main">{embedded ? "Create teams with Soma" : "Create teams with Soma"}</h2>
                    <p className="mt-1 text-sm text-cortex-text-muted">
                        Use this lane when you want Soma to shape a team, delivery lane, or first execution structure for {organizationName}.
                    </p>
                </div>
                {!embedded ? (
                    <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                        <p className="font-medium text-cortex-text-main">Primary counterpart</p>
                        <p className="mt-1">{somaName}</p>
                        <p className="mt-3 font-medium text-cortex-text-main">Operational lead</p>
                        <p className="mt-1">{teamLeadName}</p>
                    </div>
                ) : (
                    <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                        <p className="font-medium text-cortex-text-main">How this mode works</p>
                        <p className="mt-1 leading-6">
                            Stay in the same Soma workspace while focusing this mode on team design, delivery lanes, and execution structure.
                        </p>
                    </div>
                )}
            </div>

            <div className="mt-5 rounded-2xl border border-cortex-primary/30 bg-cortex-primary/10 p-4">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                    <div className="space-y-2">
                        <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Team creation lane</p>
                        <p className="text-sm leading-6 text-cortex-text-main">
                            Tell Soma what kind of team, delivery lane, or execution structure you want to create. This panel keeps that request focused on team design instead of general organization conversation.
                        </p>
                    </div>
                    <button
                        type="button"
                        onClick={handlePromptSubmit}
                        disabled={isLoading}
                        className="inline-flex items-center justify-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                        Start team design
                    </button>
                </div>
                <div className="mt-4 space-y-2">
                    <label htmlFor="soma-guided-prompt" className="text-sm font-medium text-cortex-text-main">
                        Tell Soma what team or delivery lane you want to create
                    </label>
                    <textarea
                        id="soma-guided-prompt"
                        ref={promptRef}
                        value={draftPrompt}
                        onChange={(event) => setDraftPrompt(event.target.value)}
                        onKeyDown={(event) => {
                            if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
                                event.preventDefault();
                                void handlePromptSubmit();
                            }
                        }}
                        placeholder="Tell Soma what team or delivery lane you want to create"
                        className="min-h-[104px] w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                    />
                    <p className="text-xs leading-6 text-cortex-text-muted">
                        Soma will use this request to choose the right first team-design move. Leave it blank to start with a quick strategy check right away.
                    </p>
                    {promptSuggestions.length > 0 ? (
                        <div className="flex flex-wrap gap-2 pt-1">
                            {promptSuggestions.map((suggestion) => (
                                <button
                                    key={suggestion.label}
                                    type="button"
                                    onClick={() => setDraftPrompt(suggestion.prompt)}
                                    className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
                                >
                                    {suggestion.label}
                                </button>
                            ))}
                        </div>
                    ) : null}
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {GUIDED_ACTIONS.map((item) => (
                    <button
                        key={item.action}
                        onClick={() => triggerAction(item.action)}
                        disabled={isLoading}
                        className={`w-full rounded-2xl border p-4 text-left transition-colors ${
                            selectedAction === item.action
                                ? "border-cortex-primary/40 bg-cortex-primary/10"
                                : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                        } disabled:cursor-not-allowed disabled:opacity-70`}
                        aria-busy={selectedAction === item.action && isLoading}
                    >
                        <div className="flex items-start justify-between gap-3">
                            <div>
                                <p className="text-sm font-semibold text-cortex-text-main">{item.label}</p>
                                <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{item.detail}</p>
                            </div>
                            <ArrowRight className={`mt-0.5 h-4 w-4 ${selectedAction === item.action ? "text-cortex-primary" : "text-cortex-text-muted"}`} />
                        </div>
                    </button>
                ))}
            </div>

            <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                {requestState === "idle" && (
                    <div>
                        <p className="text-sm font-semibold text-cortex-text-main">Choose a guided team-design action</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            These starting options help Soma move from conversation into organization design. Each one should produce a clearer team-creation direction, delivery focus, or setup review without leaving the workspace.
                        </p>
                    </div>
                )}

                {requestState === "loading" && (
                    <div className="flex items-center gap-3 text-sm text-cortex-text-muted">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        <span>Soma is preparing guidance for this AI Organization...</span>
                    </div>
                )}

                {requestState === "error" && error && (
                    <div>
                        <p className="text-sm font-semibold text-cortex-text-main">Soma guidance is unavailable</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{error}</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            {somaName} and the AI Organization context are still here. Retry the same action when you are ready.
                        </p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            You can also choose another guided Soma action below without leaving {organizationName}.
                        </p>
                        {selectedAction && (
                            <button
                                onClick={() => triggerAction(selectedAction)}
                                disabled={isLoading}
                                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-70"
                            >
                                <RefreshCcw className="h-4 w-4" />
                                Retry Soma action
                            </button>
                        )}
                    </div>
                )}

                {requestState === "ready" && guidance && (
                    <div className="space-y-5">
                        {requestContext && (
                            <div>
                                <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">You asked Soma to help with</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-main">{requestContext}</p>
                            </div>
                        )}

                        <div>
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Current request</p>
                            <p className="mt-2 text-base font-semibold text-cortex-text-main">{guidance.request_label}</p>
                        </div>

                        <div>
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Soma guidance</p>
                            <p className="mt-2 text-base font-semibold text-cortex-text-main">{guidance.headline}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{guidance.summary}</p>
                        </div>

                        <div className="space-y-4">
                            {guidance.execution_contract && <ExecutionContractCard contract={guidance.execution_contract} />}
                            {guidance.execution_contract?.workflow_group ? (
                                <TemporaryWorkflowLaunchCard
                                    draft={guidance.execution_contract.workflow_group}
                                    launching={launchingGroup}
                                    launchedGroup={launchedGroup}
                                    error={launchError}
                                    onLaunch={() => void launchWorkflowGroup(guidance.execution_contract!.workflow_group!)}
                                />
                            ) : null}
                        </div>

                        <div>
                            <p className="text-sm font-medium text-cortex-text-main">Priority steps</p>
                            <div className="mt-3 space-y-2">
                                {guidance.priority_steps.map((step) => (
                                    <GuidanceRow key={step}>{step}</GuidanceRow>
                                ))}
                            </div>
                        </div>

                        <div>
                            <p className="text-sm font-medium text-cortex-text-main">Keep moving with</p>
                            <div className="mt-3 flex flex-wrap gap-2">
                                {guidance.suggested_follow_ups.map((suggestion) => (
                                    <div
                                        key={suggestion}
                                        className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main"
                                    >
                                        {suggestion}
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </section>
    );
}

function normalizeGuidanceResponse(
    payload: Partial<TeamLeadGuidanceResponse> | null,
    action: TeamLeadGuidedAction,
    organizationName: string,
    somaName: string,
    teamLeadName: string,
): TeamLeadGuidanceResponse {
    const requestLabel = rewriteGuidanceText(sanitizeGuidanceText(payload?.request_label, defaultRequestLabel(action)), somaName, teamLeadName);
    const headline = rewriteGuidanceText(sanitizeGuidanceText(payload?.headline, `${somaName} guidance for ${organizationName}`), somaName, teamLeadName);
    const summary = rewriteGuidanceText(sanitizeGuidanceText(
        payload?.summary,
        `${somaName} has guidance ready for ${organizationName}.`,
    ), somaName, teamLeadName);
    const prioritySteps = normalizeGuidanceList(payload?.priority_steps, defaultPrioritySteps(action, organizationName), somaName, teamLeadName);
    const suggestedFollowUps = normalizeGuidanceList(
        payload?.suggested_follow_ups,
        defaultFollowUps(action),
        somaName,
        teamLeadName,
    );
    const executionContract = normalizeExecutionContract(payload?.execution_contract, somaName, teamLeadName);

    return {
        action,
        request_label: requestLabel,
        headline,
        summary,
        priority_steps: prioritySteps,
        suggested_follow_ups: suggestedFollowUps,
        execution_contract: executionContract,
    };
}

function defaultRequestLabel(action: TeamLeadGuidedAction) {
    return GUIDED_ACTIONS.find((item) => item.action === action)?.label ?? "Work with Soma";
}

function resolvePromptAction(prompt: string): TeamLeadGuidedAction {
    const normalized = prompt.trim().toLowerCase();
    if (/(review|check|audit|inspect|understand|setup|structure)/i.test(normalized)) {
        return "review_setup";
    }
    if (/(first|priority|focus|start with|begin with|most important)/i.test(normalized)) {
        return "focus_first";
    }
    return "plan_next_steps";
}

function defaultPrioritySteps(action: TeamLeadGuidedAction, organizationName: string) {
    switch (action) {
        case "focus_first":
            return [
                `Confirm the first priority for ${organizationName}.`,
                "Use that priority to create the first visible movement across the workspace.",
            ];
        case "review_setup":
            return [
                "Check whether Advisors, Departments, and Specialists are ready for the next move.",
                "Open the parts of the organization that need attention first.",
            ];
        case "plan_next_steps":
        default:
            return [
                `Turn ${organizationName} into a clear next move.`,
                "Use Soma guidance to choose the next action that will show up across the workspace.",
            ];
    }
}

function defaultFollowUps(action: TeamLeadGuidedAction) {
    const fallbacks = [
        "Run a quick strategy check",
        "Choose the first priority",
        "Review your organization setup",
    ];
    return fallbacks.filter((label) => label !== defaultRequestLabel(action));
}

function normalizeGuidanceList(value: unknown, fallback: string[], somaName: string, teamLeadName: string) {
    if (!Array.isArray(value)) {
        return fallback;
    }
    const normalized = value
        .map((entry) => rewriteGuidanceText(sanitizeGuidanceText(entry, ""), somaName, teamLeadName))
        .filter((entry) => entry.length > 0);
    return normalized.length > 0 ? normalized : fallback;
}

function normalizeExecutionContract(
    value: unknown,
    somaName: string,
    teamLeadName: string,
): TeamLeadGuidanceResponse["execution_contract"] {
    if (!value || typeof value !== "object") {
        return undefined;
    }

    const contract = value as Partial<TeamLeadExecutionContract>;
    const executionMode = contract.execution_mode;
    if (
        executionMode !== "guided_review" &&
        executionMode !== "native_team" &&
        executionMode !== "external_workflow_contract"
    ) {
        return undefined;
    }

    const targetOutputs = Array.isArray(contract.target_outputs)
        ? contract.target_outputs
            .map((entry) => rewriteGuidanceText(sanitizeExecutionContractText(entry, ""), somaName, teamLeadName))
            .filter((entry) => entry.length > 0)
        : [];

    return {
        execution_mode: executionMode,
        owner_label: rewriteGuidanceText(sanitizeExecutionContractText(contract.owner_label, "Execution path"), somaName, teamLeadName),
        summary: rewriteGuidanceText(sanitizeExecutionContractText(contract.summary, "Soma has an execution path ready."), somaName, teamLeadName),
        team_name: sanitizeExecutionContractText(contract.team_name, ""),
        external_target: sanitizeExecutionContractText(contract.external_target, ""),
        target_outputs: targetOutputs,
        workflow_group: normalizeWorkflowGroupDraft(contract.workflow_group, somaName, teamLeadName),
    };
}

function normalizeWorkflowGroupDraft(
    value: unknown,
    somaName: string,
    teamLeadName: string,
): TeamLeadWorkflowGroupDraft | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }

    const draft = value as Partial<TeamLeadWorkflowGroupDraft>;
    const workMode = draft.work_mode;
    if (
        workMode !== "read_only" &&
        workMode !== "propose_only" &&
        workMode !== "execute_with_approval" &&
        workMode !== "execute_bounded"
    ) {
        return undefined;
    }

    return {
        name: rewriteGuidanceText(sanitizeExecutionContractText(draft.name, "Temporary workflow group"), somaName, teamLeadName),
        goal_statement: rewriteGuidanceText(sanitizeExecutionContractText(draft.goal_statement, "Coordinate a focused workflow."), somaName, teamLeadName),
        work_mode: workMode,
        coordinator_profile: rewriteGuidanceText(sanitizeExecutionContractText(draft.coordinator_profile, "Workflow lead"), somaName, teamLeadName),
        allowed_capabilities: Array.isArray(draft.allowed_capabilities)
            ? draft.allowed_capabilities
                .map((entry) => sanitizeExecutionContractText(entry, ""))
                .filter((entry) => entry.length > 0)
            : [],
        expiry_hours: typeof draft.expiry_hours === "number" ? draft.expiry_hours : undefined,
        summary: rewriteGuidanceText(sanitizeExecutionContractText(draft.summary, "Launch a temporary workflow group from this Soma plan."), somaName, teamLeadName),
    };
}

function rewriteGuidanceText(value: string, somaName: string, teamLeadName: string) {
    return value
        .replaceAll(teamLeadName, somaName)
        .replace(/\bTeam Lead\b/g, "Soma");
}

function sanitizeGuidanceText(value: unknown, fallback: string) {
    if (typeof value !== "string") {
        return fallback;
    }

    const normalized = value
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line.length > 0)
        .filter((line) => !/^(system|debug|trace|tool|agent_id)\s*:/i.test(line))
        .filter((line) => !containsForbiddenGuidanceCopy(line))
        .join(" ")
        .replace(/\s+/g, " ")
        .trim();

    if (!normalized || normalized.startsWith("{") || normalized.startsWith("[")) {
        return fallback;
    }

    return normalized;
}

function sanitizeExecutionContractText(value: unknown, fallback: string) {
    if (typeof value !== "string") {
        return fallback;
    }

    const normalized = value
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line.length > 0)
        .filter((line) => !/^(system|debug|trace|tool|agent_id)\s*:/i.test(line))
        .join(" ")
        .replace(/\s+/g, " ")
        .trim();

    if (!normalized || normalized.startsWith("{") || normalized.startsWith("[")) {
        return fallback;
    }

    return normalized;
}

function containsForbiddenGuidanceCopy(value: string) {
    return [
        /v8 entry flow/i,
        /bounded slice/i,
        /implementation slice/i,
        /context shell/i,
        /raw architecture controls/i,
        /\bcontract\b/i,
        /\bloop\b/i,
        /scheduler/i,
        /inception/i,
        /soma kernel/i,
        /\bcouncil\b/i,
        /provider policy/i,
        /identity\s*\/\s*continuity/i,
        /memory promotion/i,
        /pgvector/i,
    ].some((pattern) => pattern.test(value));
}

function GuidanceRow({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex items-start gap-3 rounded-2xl border border-cortex-border bg-cortex-surface/60 px-3 py-3 text-sm text-cortex-text-muted">
            <span className="mt-1 h-2 w-2 rounded-full bg-cortex-primary" />
            <span className="leading-6">{children}</span>
        </div>
    );
}

function ExecutionContractCard({ contract }: { contract: TeamLeadExecutionContract }) {
    const isNativeTeam = contract.execution_mode === "native_team";
    const title = isNativeTeam ? "Native Mycelis team" : contract.execution_mode === "external_workflow_contract"
        ? "External workflow contract"
        : "Guided execution path";

    return (
        <div className="rounded-2xl border border-cortex-primary/25 bg-cortex-primary/5 px-4 py-4">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Execution path</p>
            <div className="mt-3 flex flex-wrap items-center gap-2">
                <span className="rounded-full border border-cortex-primary/25 bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-primary">
                    {title}
                </span>
                <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                    {contract.owner_label}
                </span>
                {contract.team_name ? (
                    <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                        {contract.team_name}
                    </span>
                ) : null}
                {contract.external_target ? (
                    <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                        {contract.external_target}
                    </span>
                ) : null}
            </div>
            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">{contract.summary}</p>
            {contract.target_outputs.length > 0 ? (
                <div className="mt-4">
                    <p className="text-sm font-medium text-cortex-text-main">Target outputs</p>
                    <div className="mt-2 flex flex-wrap gap-2">
                        {contract.target_outputs.map((output) => (
                            <span
                                key={output}
                                className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main"
                            >
                                {output}
                            </span>
                        ))}
                    </div>
                </div>
            ) : null}
        </div>
    );
}

function TemporaryWorkflowLaunchCard({
    draft,
    launching,
    launchedGroup,
    error,
    onLaunch,
}: {
    draft: TeamLeadWorkflowGroupDraft;
    launching: boolean;
    launchedGroup: LaunchedGroupState | null;
    error: string | null;
    onLaunch: () => void;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Launch temporary workflow group</p>
            <p className="mt-3 text-sm font-semibold text-cortex-text-main">{draft.name}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{draft.summary}</p>
            <div className="mt-3 flex flex-wrap gap-2">
                <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                    {draft.work_mode}
                </span>
                <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                    {draft.coordinator_profile}
                </span>
                {typeof draft.expiry_hours === "number" && draft.expiry_hours > 0 ? (
                    <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                        expires in {draft.expiry_hours}h
                    </span>
                ) : null}
            </div>
            <div className="mt-4 flex flex-wrap items-center gap-3">
                <button
                    type="button"
                    onClick={onLaunch}
                    disabled={launching}
                    className="rounded-2xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:opacity-60"
                >
                    {launching ? "Creating workflow group..." : "Create temporary workflow group"}
                </button>
                {launchedGroup ? (
                    <Link
                        href={`/groups?group_id=${encodeURIComponent(launchedGroup.groupId)}`}
                        className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-2 text-sm font-medium text-cortex-text-main hover:border-cortex-primary/20"
                    >
                        Open {launchedGroup.name}
                    </Link>
                ) : null}
            </div>
            {launchedGroup ? (
                <p className="mt-3 text-sm text-cortex-primary">
                    Soma launched {launchedGroup.name}. The workflow group is ready for focused coordination and retained output review.
                </p>
            ) : null}
            {error ? (
                <p className="mt-3 text-sm text-cortex-danger">{error}</p>
            ) : null}
        </div>
    );
}
