import { Activity, RefreshCcw, Sparkles } from "lucide-react";
import type {
    OrganizationLearningInsightItem,
    OrganizationLoopActivityItem,
} from "@/lib/organizations";

export type CausalStripState = {
    action: string;
    teamsEngaged: string[];
    outputsGenerated: string[];
    panelsUpdated: string[];
};

export function RecentActivityPanel({
    items,
    loading,
    error,
    onRetry,
    causalAction,
}: {
    items: OrganizationLoopActivityItem[];
    loading: boolean;
    error: string | null;
    onRetry: () => void;
    causalAction: string;
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
            <div className="mt-4 grid gap-3 lg:grid-cols-3">
                <CausalFact label="What changed" value={visibleItems[0]?.summary ?? "No visible activity yet"} />
                <CausalFact label="Why it changed" value="Recent Activity refreshes after Soma requests and ongoing checks create new visible signals." />
                <CausalFact label="How Soma uses it" value={`Soma uses this panel to explain what happened after "${causalAction}".`} />
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

export function LearningVisibilityPanel({
    items,
    loading,
    error,
    onRetry,
    causalAction,
}: {
    items: OrganizationLearningInsightItem[];
    loading: boolean;
    error: string | null;
    onRetry: () => void;
    causalAction: string;
}) {
    const visibleItems = items.slice(0, 6);

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex items-start gap-3">
                <div className="mt-0.5 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 p-2 text-cortex-primary">
                    <Sparkles className="h-4 w-4" />
                </div>
                <div>
                    <h2 className="text-xl font-semibold text-cortex-text-main">What the Organization Is Retaining</h2>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                        See the recurring patterns and continuity cues your AI Organization is turning into reusable guidance, and why they matter for what happens next.
                    </p>
                </div>
            </div>
            <div className="mt-4 grid gap-3 lg:grid-cols-3">
                <CausalFact label="What changed" value={visibleItems[0]?.summary ?? "No retained patterns visible yet"} />
                <CausalFact label="Why it changed" value="Retained patterns appear when repeated work and review signals become strong enough to describe in plain language." />
                <CausalFact label="How Soma uses it" value={`Soma uses this panel to separate reusable knowledge from temporary continuity after "${causalAction}".`} />
            </div>

            {error && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Memory &amp; Continuity updates unavailable</p>
                    <p className="mt-1 leading-6">Recent retained patterns are not available right now. The Soma workspace is still ready.</p>
                    <button
                        type="button"
                        onClick={onRetry}
                        className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Memory &amp; Continuity
                    </button>
                </div>
            )}

            {!error && loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Checking recent retained patterns</p>
                    <p className="mt-1 leading-6">The latest reusable patterns and continuity cues will appear here shortly.</p>
                </div>
            )}

            {!error && !loading && visibleItems.length === 0 && (
                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">No retained patterns yet</p>
                    <p className="mt-1 leading-6">This is where reusable patterns, continuity cues, and stronger working habits will appear in plain language.</p>
                    <p className="mt-1 leading-6">Ordinary planning chat stays in working continuity until a stronger reusable pattern emerges or you intentionally save something for later recall.</p>
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

export function SomaCausalStrip({
    action,
    teamsEngaged,
    outputsGenerated,
    panelsUpdated,
}: CausalStripState) {
    return (
        <div className="mt-5 rounded-2xl border border-cortex-primary/25 bg-cortex-primary/10 p-4">
            <div className="flex items-center gap-2 text-cortex-primary">
                <Sparkles className="h-4 w-4" />
                <p className="text-sm font-semibold text-cortex-text-main">Soma just did this</p>
            </div>
            <div className="mt-4 grid gap-3 lg:grid-cols-4">
                <CausalFact label="Last action" value={action} />
                <CausalFact label="Teams engaged" value={teamsEngaged.join(", ")} />
                <CausalFact label="Outputs generated" value={outputsGenerated.join(", ")} />
                <CausalFact label="Panels updated" value={panelsUpdated.join(", ")} />
            </div>
        </div>
    );
}

export function CausalFact({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-main">{value}</p>
        </div>
    );
}

export function ActivityStatusBadge({ status }: { status: OrganizationLoopActivityItem["status"] }) {
    const config =
        status === "warning"
            ? { label: "Needs review", className: "border-amber-500/30 bg-amber-500/10 text-amber-200" }
            : status === "failed"
              ? { label: "Unavailable", className: "border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger" }
              : { label: "Ready", className: "border-emerald-500/30 bg-emerald-500/10 text-emerald-200" };

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
            ? { label: "Strong", className: "border-emerald-500/30 bg-emerald-500/10 text-emerald-200" }
            : strength === "consistent"
              ? { label: "Consistent", className: "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary" }
              : { label: "Emerging", className: "border-amber-500/30 bg-amber-500/10 text-amber-200" };

    return (
        <span className={`inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.16em] ${config.className}`}>
            <span className="h-1.5 w-1.5 rounded-full bg-current" />
            {config.label}
        </span>
    );
}

export function formatRelativeActivityTime(timestamp: string) {
    const parsed = Date.parse(timestamp);
    if (Number.isNaN(parsed)) {
        return "Recently";
    }

    const diffMinutes = Math.floor(Math.max(0, Date.now() - parsed) / 60000);
    if (diffMinutes <= 0) {
        return "Just now";
    }
    if (diffMinutes < 60) {
        return diffMinutes === 1 ? "1 minute ago" : `${diffMinutes} minutes ago`;
    }

    const diffHours = Math.floor(diffMinutes / 60);
    if (diffHours < 24) {
        return diffHours === 1 ? "1 hour ago" : `${diffHours} hours ago`;
    }

    const diffDays = Math.floor(diffHours / 24);
    return diffDays === 1 ? "1 day ago" : `${diffDays} days ago`;
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
