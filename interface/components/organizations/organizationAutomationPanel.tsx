import { RefreshCcw } from "lucide-react";
import type { OrganizationAutomationItem } from "@/lib/organizations";
import {
    ActivityStatusBadge,
    formatRelativeActivityTime,
} from "@/components/organizations/organizationActivityPanels";

export function AutomationDetailPanel({
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
