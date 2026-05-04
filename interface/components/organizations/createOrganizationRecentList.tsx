"use client";

import { ArrowRight, Loader2 } from "lucide-react";
import type { OrganizationSummary } from "@/lib/organizations";
import {
    ActionableState,
    ActionButton,
    EmptyState,
    LoadingState,
    Metric,
} from "@/components/organizations/createOrganizationEntryUi";

type ResourceState = "loading" | "ready" | "error";

export function CreateOrganizationRecentList({
    showRecentOrganizations,
    diagnosticOrganizations,
    showDiagnosticOrganizations,
    visibleDiagnosticOrganizations,
    hiddenDiagnosticCount,
    testingSetupCountLabel,
    organizationsState,
    organizationsError,
    visibleOrganizations,
    openingOrganizationId,
    isSubmitting,
    onToggleDiagnostics,
    onReloadOrganizations,
    onOpenOrganization,
}: {
    showRecentOrganizations: boolean;
    diagnosticOrganizations: OrganizationSummary[];
    showDiagnosticOrganizations: boolean;
    visibleDiagnosticOrganizations: OrganizationSummary[];
    hiddenDiagnosticCount: number;
    testingSetupCountLabel: string;
    organizationsState: ResourceState;
    organizationsError: string | null;
    visibleOrganizations: OrganizationSummary[];
    openingOrganizationId: string | null;
    isSubmitting: boolean;
    onToggleDiagnostics: () => void;
    onReloadOrganizations: () => void;
    onOpenOrganization: (organization: OrganizationSummary) => void;
}) {
    if (!showRecentOrganizations) {
        return null;
    }
    return (
        <section className="space-y-3">
            <div className="flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
                <div>
                    <h2 className="text-lg font-semibold text-cortex-text-main">Recent AI Organizations</h2>
                    <p className="text-sm text-cortex-text-muted">Reopen an existing organization when you want to continue instead of starting over.</p>
                </div>
                {diagnosticOrganizations.length > 0 && (
                    <div className="flex flex-col items-start gap-2 lg:items-end">
                        <button type="button" onClick={onToggleDiagnostics} className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-muted transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary">
                            {showDiagnosticOrganizations ? "Hide" : "Show"} {visibleDiagnosticOrganizations.length} {testingSetupCountLabel}
                        </button>
                        {showDiagnosticOrganizations && hiddenDiagnosticCount > 0 && (
                            <p className="text-xs text-cortex-text-muted">Keeping {hiddenDiagnosticCount} older testing {hiddenDiagnosticCount === 1 ? "setup" : "setups"} out of the default product view.</p>
                        )}
                    </div>
                )}
            </div>

            {organizationsState === "loading" && <LoadingState label="Loading recent AI Organizations..." />}
            {organizationsState === "error" && organizationsError && (
                <ActionableState
                    title="Recent AI Organizations are unavailable"
                    message={organizationsError}
                    guidance="You can still create a new AI Organization above while we retry your recent organizations."
                    actions={<ActionButton onClick={onReloadOrganizations} kind="secondary" showRetryIcon>Retry recent AI Organizations</ActionButton>}
                />
            )}
            {organizationsState === "ready" && visibleOrganizations.length > 0 && (
                <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                    {visibleOrganizations.map((organization) => (
                        <button
                            key={organization.id}
                            type="button"
                            onClick={() => onOpenOrganization(organization)}
                            disabled={openingOrganizationId === organization.id || isSubmitting}
                            className="rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-left transition-colors hover:border-cortex-primary/25 disabled:cursor-not-allowed disabled:opacity-70"
                            aria-busy={openingOrganizationId === organization.id}
                        >
                            <div className="flex items-start justify-between gap-3">
                                <div>
                                    <p className="text-base font-semibold text-cortex-text-main">{organization.name}</p>
                                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{organization.purpose || "Reopen this AI Organization and continue where you left off."}</p>
                                </div>
                                <span className="rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                                    {organization.start_mode === "template" ? "Template" : "Empty"}
                                </span>
                            </div>
                            <div className="mt-4 grid grid-cols-2 gap-2 text-xs text-cortex-text-muted">
                                <Metric label="Team Lead" value={organization.team_lead_label} />
                                <Metric label="Departments" value={String(organization.department_count)} />
                                <Metric label="Specialists" value={String(organization.specialist_count)} />
                                <Metric label="AI Organization" value={organization.status} />
                            </div>
                            <div className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 px-3 py-2 text-sm font-medium text-cortex-primary transition-colors hover:bg-cortex-primary/10">
                                {openingOrganizationId === organization.id ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                                Open AI Organization
                                <ArrowRight className="h-4 w-4" />
                            </div>
                        </button>
                    ))}
                </div>
            )}
            {organizationsState === "ready" && visibleOrganizations.length === 0 && (
                <EmptyState title="No recent AI Organizations yet" detail="Create your first AI Organization above to give Soma a real workspace, team structure, and continuity." />
            )}
        </section>
    );
}
