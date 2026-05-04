"use client";

import { Loader2, RefreshCcw } from "lucide-react";
import type {
    TeamLeadGuidanceResponse,
    TeamLeadGuidedAction,
    TeamLeadWorkflowGroupDraft,
} from "@/lib/organizations";
import {
    CompactTeamGuidanceCard,
    ExecutionContractCard,
    TemporaryWorkflowLaunchCard,
    type LaunchedGroupState,
} from "@/components/organizations/teamLeadExecutionCards";

type RequestState = "idle" | "loading" | "ready" | "error";

export function TeamLeadGuidanceResult({
    requestState,
    error,
    isLoading,
    selectedAction,
    requestContext,
    requestScope,
    guidance,
    launchingGroup,
    launchedGroup,
    launchError,
    somaName,
    organizationName,
    onRetryAction,
    onLaunchWorkflowGroup,
}: {
    requestState: RequestState;
    error: string | null;
    isLoading: boolean;
    selectedAction: TeamLeadGuidedAction | null;
    requestContext: string | null;
    requestScope: "compact" | "broad" | null;
    guidance: TeamLeadGuidanceResponse | null;
    launchingGroup: boolean;
    launchedGroup: LaunchedGroupState | null;
    launchError: string | null;
    somaName: string;
    organizationName: string;
    onRetryAction: (action: TeamLeadGuidedAction) => void;
    onLaunchWorkflowGroup: (draft: TeamLeadWorkflowGroupDraft) => void;
}) {
    return (
        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            {requestState === "idle" && <IdleState />}
            {requestState === "loading" && <LoadingState />}
            {requestState === "error" && error && (
                <ErrorState
                    error={error}
                    isLoading={isLoading}
                    organizationName={organizationName}
                    selectedAction={selectedAction}
                    somaName={somaName}
                    onRetryAction={onRetryAction}
                />
            )}
            {requestState === "ready" && guidance && (
                <ReadyState
                    guidance={guidance}
                    requestContext={requestContext}
                    requestScope={requestScope}
                    launchingGroup={launchingGroup}
                    launchedGroup={launchedGroup}
                    launchError={launchError}
                    onLaunchWorkflowGroup={onLaunchWorkflowGroup}
                />
            )}
        </div>
    );
}

function IdleState() {
    return (
        <div>
            <p className="text-sm font-semibold text-cortex-text-main">Choose a guided team-design action</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                These starting options help Soma move from conversation into organization design. Each one should produce a clearer team-creation direction, delivery focus, or setup review without leaving the workspace.
            </p>
        </div>
    );
}

function LoadingState() {
    return (
        <div className="flex items-center gap-3 text-sm text-cortex-text-muted">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Soma is preparing guidance for this AI Organization...</span>
        </div>
    );
}

function ErrorState({
    error,
    isLoading,
    selectedAction,
    somaName,
    organizationName,
    onRetryAction,
}: {
    error: string;
    isLoading: boolean;
    selectedAction: TeamLeadGuidedAction | null;
    somaName: string;
    organizationName: string;
    onRetryAction: (action: TeamLeadGuidedAction) => void;
}) {
    return (
        <div>
            <p className="text-sm font-semibold text-cortex-text-main">Soma guidance is unavailable</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{error}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{somaName} and the AI Organization context are still here. Retry the same action when you are ready.</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">You can also choose another guided Soma action below without leaving {organizationName}.</p>
            {selectedAction && (
                <button
                    onClick={() => onRetryAction(selectedAction)}
                    disabled={isLoading}
                    className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-70"
                >
                    <RefreshCcw className="h-4 w-4" />
                    Retry Soma action
                </button>
            )}
        </div>
    );
}

function ReadyState({
    guidance,
    requestContext,
    requestScope,
    launchingGroup,
    launchedGroup,
    launchError,
    onLaunchWorkflowGroup,
}: {
    guidance: TeamLeadGuidanceResponse;
    requestContext: string | null;
    requestScope: "compact" | "broad" | null;
    launchingGroup: boolean;
    launchedGroup: LaunchedGroupState | null;
    launchError: string | null;
    onLaunchWorkflowGroup: (draft: TeamLeadWorkflowGroupDraft) => void;
}) {
    return (
        <div className="space-y-5">
            {requestContext && (
                <div>
                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">You asked Soma to help with</p>
                    <p className="mt-2 text-sm leading-6 text-cortex-text-main">{requestContext}</p>
                </div>
            )}
            <CompactTeamGuidanceCard requestScope={requestScope} />
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
                        onLaunch={() => onLaunchWorkflowGroup(guidance.execution_contract!.workflow_group!)}
                    />
                ) : null}
            </div>
            <GuidanceList title="Priority steps" items={guidance.priority_steps} />
            <GuidanceList title="Keep moving with" items={guidance.suggested_follow_ups} pill />
        </div>
    );
}

function GuidanceList({ title, items, pill = false }: { title: string; items: string[]; pill?: boolean }) {
    return (
        <div>
            <p className="text-sm font-medium text-cortex-text-main">{title}</p>
            <div className={pill ? "mt-3 flex flex-wrap gap-2" : "mt-3 space-y-2"}>
                {items.map((item) =>
                    pill ? (
                        <div key={item} className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main">{item}</div>
                    ) : (
                        <div key={item} className="flex items-start gap-3 rounded-2xl border border-cortex-border bg-cortex-surface/60 px-3 py-3 text-sm text-cortex-text-muted">
                            <span className="mt-1 h-2 w-2 rounded-full bg-cortex-primary" />
                            <span className="leading-6">{item}</span>
                        </div>
                    ),
                )}
            </div>
        </div>
    );
}
