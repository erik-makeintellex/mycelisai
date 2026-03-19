"use client";

import { useState } from "react";
import { ArrowRight, Loader2, RefreshCcw } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type { TeamLeadGuidanceRequest, TeamLeadGuidanceResponse, TeamLeadGuidedAction } from "@/lib/organizations";

type RequestState = "idle" | "loading" | "ready" | "error";

const GUIDED_ACTIONS: Array<{ action: TeamLeadGuidedAction; label: string; detail: string }> = [
    {
        action: "plan_next_steps",
        label: "Plan next steps for this organization",
        detail: "Ask the Team Lead for a focused first plan tied to the current AI Organization.",
    },
    {
        action: "focus_first",
        label: "What should I focus on first?",
        detail: "Let the Team Lead surface the first operating priority before work fans out.",
    },
    {
        action: "review_setup",
        label: "Review my organization setup",
        detail: "Check whether Advisors, Departments, and Specialists are ready for the next move.",
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
    teamLeadName,
}: {
    organizationId: string;
    organizationName: string;
    teamLeadName: string;
}) {
    const [selectedAction, setSelectedAction] = useState<TeamLeadGuidedAction | null>(null);
    const [requestState, setRequestState] = useState<RequestState>("idle");
    const [error, setError] = useState<string | null>(null);
    const [guidance, setGuidance] = useState<TeamLeadGuidanceResponse | null>(null);
    const isLoading = requestState === "loading";

    const triggerAction = async (action: TeamLeadGuidedAction) => {
        if (isLoading) {
            return;
        }
        setSelectedAction(action);
        setRequestState("loading");
        setError(null);

        const payload: TeamLeadGuidanceRequest = { action };
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
            setGuidance(normalizeGuidanceResponse(rawGuidance, action, organizationName, teamLeadName));
            setRequestState("ready");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Team Lead guidance is unavailable right now.");
            setRequestState("error");
        }
    };

    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                <div>
                    <h2 className="text-xl font-semibold text-cortex-text-main">Work with the Team Lead</h2>
                    <p className="mt-1 text-sm text-cortex-text-muted">
                        Start inside {organizationName} with a guided Team Lead action instead of a blank assistant thread.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Current Team Lead</p>
                    <p className="mt-1">{teamLeadName}</p>
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
                        <p className="text-sm font-semibold text-cortex-text-main">Choose a guided Team Lead action</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            The Team Lead will return structured guidance for {organizationName} with clear steps and follow-through suggestions.
                        </p>
                    </div>
                )}

                {requestState === "loading" && (
                    <div className="flex items-center gap-3 text-sm text-cortex-text-muted">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        <span>Team Lead is preparing guidance for this AI Organization...</span>
                    </div>
                )}

                {requestState === "error" && error && (
                    <div>
                        <p className="text-sm font-semibold text-cortex-text-main">Team Lead guidance is unavailable</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{error}</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            {teamLeadName} and the AI Organization context are still here. Retry the same action when you are ready.
                        </p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            You can also choose another guided Team Lead action below without leaving {organizationName}.
                        </p>
                        {selectedAction && (
                            <button
                                onClick={() => triggerAction(selectedAction)}
                                disabled={isLoading}
                                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-70"
                            >
                                <RefreshCcw className="h-4 w-4" />
                                Retry Team Lead action
                            </button>
                        )}
                    </div>
                )}

                {requestState === "ready" && guidance && (
                    <div className="space-y-5">
                        <div>
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Current request</p>
                            <p className="mt-2 text-base font-semibold text-cortex-text-main">{guidance.request_label}</p>
                        </div>

                        <div>
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Team Lead guidance</p>
                            <p className="mt-2 text-base font-semibold text-cortex-text-main">{guidance.headline}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{guidance.summary}</p>
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
    teamLeadName: string,
): TeamLeadGuidanceResponse {
    const requestLabel = sanitizeGuidanceText(payload?.request_label, defaultRequestLabel(action));
    const headline = sanitizeGuidanceText(payload?.headline, `${teamLeadName} guidance for ${organizationName}`);
    const summary = sanitizeGuidanceText(
        payload?.summary,
        `${teamLeadName} has guidance ready for ${organizationName}.`,
    );
    const prioritySteps = normalizeGuidanceList(payload?.priority_steps, defaultPrioritySteps(action, organizationName));
    const suggestedFollowUps = normalizeGuidanceList(
        payload?.suggested_follow_ups,
        defaultFollowUps(action),
    );

    return {
        action,
        request_label: requestLabel,
        headline,
        summary,
        priority_steps: prioritySteps,
        suggested_follow_ups: suggestedFollowUps,
    };
}

function defaultRequestLabel(action: TeamLeadGuidedAction) {
    return GUIDED_ACTIONS.find((item) => item.action === action)?.label ?? "Work with the Team Lead";
}

function defaultPrioritySteps(action: TeamLeadGuidedAction, organizationName: string) {
    switch (action) {
        case "focus_first":
            return [
                `Confirm the first priority for ${organizationName}.`,
                "Keep the Team Lead as the working counterpart while the next step is clarified.",
            ];
        case "review_setup":
            return [
                "Check whether Advisors, Departments, and Specialists are ready for the next move.",
                "Confirm the organization shape before expanding the workflow.",
            ];
        case "plan_next_steps":
        default:
            return [
                `Turn ${organizationName} into a focused first delivery loop.`,
                "Use the Team Lead guidance to choose the next concrete move.",
            ];
    }
}

function defaultFollowUps(action: TeamLeadGuidedAction) {
    const fallbacks = [
        "Plan next steps for this organization",
        "What should I focus on first?",
        "Review my organization setup",
    ];
    return fallbacks.filter((label) => label !== defaultRequestLabel(action));
}

function normalizeGuidanceList(value: unknown, fallback: string[]) {
    if (!Array.isArray(value)) {
        return fallback;
    }
    const normalized = value
        .map((entry) => sanitizeGuidanceText(entry, ""))
        .filter((entry) => entry.length > 0);
    return normalized.length > 0 ? normalized : fallback;
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

function containsForbiddenGuidanceCopy(value: string) {
    return [
        /v8 entry flow/i,
        /bounded slice/i,
        /implementation slice/i,
        /context shell/i,
        /raw architecture controls/i,
        /\bcontract\b/i,
        /inception/i,
        /soma kernel/i,
        /central council/i,
        /provider policy/i,
        /identity\s*\/\s*continuity/i,
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
