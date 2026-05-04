"use client";

import { useEffect, useRef, useState } from "react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type {
    TeamLeadGuidanceRequest,
    TeamLeadGuidanceResponse,
    TeamLeadGuidedAction,
    TeamLeadWorkflowGroupDraft,
} from "@/lib/organizations";
import {
    classifyRequestScope,
    defaultRequestLabel,
    normalizeGuidanceResponse,
    rewriteGuidanceText,
    resolvePromptAction,
} from "@/components/organizations/teamLeadGuidanceNormalization";
import {
    persistWorkspaceState,
    readPersistedWorkspaceState,
    type RequestState,
} from "@/components/organizations/teamLeadWorkspaceStorage";
import { TeamLeadGuidanceResult } from "@/components/organizations/teamLeadGuidanceResult";
import type { LaunchedGroupState } from "@/components/organizations/teamLeadExecutionCards";
import {
    TeamLeadActionList,
    TeamLeadPanelHeader,
    TeamLeadRequestComposer,
} from "@/components/organizations/teamLeadRequestComposer";

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
    const requestScope = requestContext ? classifyRequestScope(requestContext) : null;

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
            <TeamLeadPanelHeader
                organizationName={organizationName}
                somaName={somaName}
                teamLeadName={teamLeadName}
                embedded={embedded}
            />
            <TeamLeadRequestComposer
                promptRef={promptRef}
                draftPrompt={draftPrompt}
                promptSuggestions={promptSuggestions}
                isLoading={isLoading}
                onPromptChange={setDraftPrompt}
                onSubmit={() => void handlePromptSubmit()}
            />
            <TeamLeadActionList
                selectedAction={selectedAction}
                isLoading={isLoading}
                onAction={(action) => void triggerAction(action)}
            />

            <TeamLeadGuidanceResult
                requestState={requestState}
                error={error}
                isLoading={isLoading}
                selectedAction={selectedAction}
                requestContext={requestContext}
                requestScope={requestScope}
                guidance={guidance}
                launchingGroup={launchingGroup}
                launchedGroup={launchedGroup}
                launchError={launchError}
                somaName={somaName}
                organizationName={organizationName}
                onRetryAction={(action) => void triggerAction(action)}
                onLaunchWorkflowGroup={(draft) => void launchWorkflowGroup(draft)}
            />
        </section>
    );
}
