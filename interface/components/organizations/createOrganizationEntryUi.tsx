"use client";

import Link from "next/link";
import { ArrowRight, Loader2, RefreshCcw } from "lucide-react";
import type { OrganizationStarterTemplate } from "@/lib/organizations";
import { StarterConfigViewerAction } from "@/components/organizations/StarterConfigViewer";

export function StartModeCard({
    active,
    title,
    description,
    icon,
    onClick,
}: {
    active: boolean;
    title: string;
    description: string;
    icon: React.ReactNode;
    onClick: () => void;
}) {
    return (
        <button onClick={onClick} className={`rounded-2xl border p-4 text-left transition-colors ${active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"}`}>
            <div className="flex items-start gap-3">
                <div className="rounded-xl border border-cortex-border bg-cortex-surface p-2 text-cortex-primary">{icon}</div>
                <div>
                    <p className="text-base font-semibold text-cortex-text-main">{title}</p>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{description}</p>
                </div>
            </div>
        </button>
    );
}

export function StarterCard({
    template,
    active,
    onSelect,
}: {
    template: OrganizationStarterTemplate;
    active: boolean;
    onSelect: () => void;
}) {
    const unavailable = template.available === false;
    const isCustom = template.source === "operator";
    const title = template.starter?.label || template.name;
    const workModeLabel = template.delivery_summary?.work_mode_label || "Governed delivery";
    return (
        <div
            data-testid={`starter-card-${template.id}`}
            role="button"
            tabIndex={unavailable ? -1 : 0}
            aria-disabled={unavailable}
            onClick={() => {
                if (!unavailable) onSelect();
            }}
            onKeyDown={(event) => {
                if (unavailable) return;
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onSelect();
                }
            }}
            className={`rounded-2xl border p-3.5 text-left transition-colors ${
                unavailable ? "cursor-not-allowed opacity-60" : "cursor-pointer"
            } ${
                active && !unavailable
                    ? "border-cortex-primary/50 bg-cortex-primary/10 shadow-sm"
                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/25"
            }`}
        >
            <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-cortex-text-main">{title}</p>
                    <p className="mt-1 line-clamp-2 text-xs leading-5 text-cortex-text-muted">{template.description}</p>
                </div>
                <div className="flex flex-shrink-0 items-center gap-1">
                    {isCustom ? (
                        <span
                            data-testid={`starter-card-${template.id}-custom-badge`}
                            className="rounded-full border border-cortex-primary/35 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary"
                        >
                            Custom
                        </span>
                    ) : null}
                    <span className="rounded-full border border-cortex-border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
                        {workModeLabel}
                    </span>
                    <StarterConfigViewerAction templateId={template.id} />
                </div>
            </div>
            <div className="mt-2.5 flex flex-wrap items-center gap-1.5 text-[11px] font-mono text-cortex-text-muted">
                <span className="rounded-md border border-cortex-border bg-cortex-surface/70 px-2 py-0.5 text-cortex-text-main">
                    {template.team_lead_label}
                </span>
                <span className="rounded-md border border-cortex-border bg-cortex-surface/70 px-2 py-0.5">
                    {template.department_count} {template.department_count === 1 ? "Dept" : "Depts"}
                </span>
                <span className="rounded-md border border-cortex-border bg-cortex-surface/70 px-2 py-0.5">
                    {template.specialist_count} {template.specialist_count === 1 ? "Specialist" : "Specialists"}
                </span>
                <span className="rounded-md border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-cortex-primary">
                    + Ideal Reviewer Gate
                </span>
            </div>
            {unavailable && (
                <p className="mt-2 text-xs leading-5 text-cortex-danger">
                    {template.unavailable_reason || "This starter is unavailable right now."}
                </p>
            )}
        </div>
    );
}

export function StarterDeliverySummary({ template }: { template: OrganizationStarterTemplate }) {
    const summary = template.delivery_summary;
    const label = template.starter?.label || template.name;
    const departments = template.departments ?? [];
    const loops = template.loop_profiles ?? [];
    if (!summary && departments.length === 0) {
        return null;
    }
    return (
        <div
            data-testid="starter-target-delivery-breakdown"
            className="space-y-3 rounded-2xl border border-cortex-primary/30 bg-cortex-bg p-4 text-sm text-cortex-text-muted"
        >
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-cortex-border pb-2.5">
                <div>
                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                        Target Delivery &amp; Execution Breakdown
                    </p>
                    <p className="mt-0.5 text-base font-semibold text-cortex-text-main">
                        What you get with {label}
                    </p>
                </div>
                {summary?.work_mode_label ? (
                    <span className="rounded-full border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 py-0.5 text-xs font-mono font-semibold text-cortex-primary">
                        {summary.work_mode_label}
                    </span>
                ) : null}
            </div>
            {template.starter?.soma_intro ? (
                <p className="rounded-xl border border-cortex-border bg-cortex-surface/70 px-3 py-2 text-xs leading-5 text-cortex-text-main">
                    <span className="font-semibold text-cortex-primary">Soma Lead Approach: </span>
                    {template.starter.soma_intro}
                </p>
            ) : null}
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                <Metric label="Team Lead & Advisors" value={`${template.team_lead_label} + ${template.advisor_count} Advisor(s)`} />
                <Metric label="Planned Structure" value={`${template.department_count} Dept(s) · ${template.specialist_count} Specialist(s)`} />
                <Metric label="AI Engine & Memory" value={`${template.ai_engine_settings_summary} · ${template.memory_personality_summary}`} />
            </div>
            {summary && (
                <div className="rounded-xl border border-cortex-border bg-cortex-surface/50 p-3">
                    <p className="text-xs font-semibold uppercase tracking-wider text-cortex-text-main">
                        Execution &amp; Governance Posture
                    </p>
                    <ul className="mt-1.5 space-y-1 text-xs leading-5">
                        <li>{summary.approval_posture}</li>
                        <li>{summary.output_proof}</li>
                        <li>{summary.team_shape}</li>
                        <li>Departments start as planned; Soma proposes launching teams and you approve.</li>
                        <li>
                            <span className="font-semibold text-cortex-primary">Delivery Verification Gate: </span>
                            Every instantiated team guarantees an in-team Ideal Delivery Reviewer (<span className="font-mono">council-sentry</span>) plus a companion Review Team for complex multi-file deliveries.
                        </li>
                    </ul>
                </div>
            )}
            {departments.length > 0 && (
                <div className="space-y-2" data-testid="starter-department-execution-lanes">
                    <p className="text-xs font-semibold uppercase tracking-wider text-cortex-text-main">
                        Planned Departments &amp; Specialist Execution Lanes
                    </p>
                    <div className="grid gap-2 sm:grid-cols-2">
                        {departments.map((dept) => (
                            <div key={dept.id} className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2.5 text-xs">
                                <div className="flex items-center justify-between gap-2">
                                    <span className="font-semibold text-cortex-text-main">{dept.name}</span>
                                    <span className="font-mono text-[11px] text-cortex-primary">
                                        {dept.specialist_count} Specialist{dept.specialist_count === 1 ? "" : "s"}
                                    </span>
                                </div>
                                {dept.agent_type_profiles && dept.agent_type_profiles.length > 0 ? (
                                    <ul className="mt-1.5 space-y-1 border-t border-cortex-border/60 pt-1.5 text-[11px] leading-4">
                                        {dept.agent_type_profiles.map((agent) => (
                                            <li key={agent.id} className="flex items-start justify-between gap-2">
                                                <span className="font-medium text-cortex-text-main">{agent.name}: <span className="font-normal text-cortex-text-muted">{agent.helps_with}</span></span>
                                            </li>
                                        ))}
                                    </ul>
                                ) : null}
                            </div>
                        ))}
                    </div>
                </div>
            )}
            {loops.length > 0 && (
                <div className="rounded-xl border border-cortex-border bg-cortex-surface/50 px-3 py-2 text-xs">
                    <span className="font-semibold text-cortex-text-main">Automated Readiness &amp; Review Cadence: </span>
                    {loops.map((loop) => loop.name).join(" · ")}
                </div>
            )}
        </div>
    );
}

export function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface/60 px-3 py-2">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

export function HiddenLater({ label, icon }: { label: string; icon: React.ReactNode }) {
    return (
        <div className="flex items-center gap-2 rounded-xl border border-cortex-border px-3 py-2">
            <span className="text-cortex-primary">{icon}</span>
            <span>{label}</span>
        </div>
    );
}

export function LoadingState({ label }: { label: string }) {
    return (
        <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>{label}</span>
        </div>
    );
}

export function ActionableState({ title, message, guidance, actions }: { title: string; message: string; guidance: string; actions?: React.ReactNode }) {
    return (
        <div className="rounded-2xl border border-cortex-danger/30 bg-cortex-danger/10 px-4 py-4 text-sm text-cortex-text-main">
            <p className="font-semibold">{title}</p>
            <p className="mt-2 leading-6">{message}</p>
            <p className="mt-2 text-cortex-text-muted">{guidance}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

export function EmptyState({ title, detail, actions }: { title: string; detail: string; actions?: React.ReactNode }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

export function ActionButton({ children, onClick, kind, disabled = false, showRetryIcon = false }: { children: React.ReactNode; onClick: () => void; kind: "primary" | "secondary" | "ghost"; disabled?: boolean; showRetryIcon?: boolean }) {
    const classes =
        kind === "primary"
            ? "border border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/15"
            : kind === "secondary"
              ? "border border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
              : "border border-transparent bg-transparent text-cortex-primary hover:bg-cortex-primary/10";
    return (
        <button onClick={onClick} disabled={disabled} className={`inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${classes}`}>
            {showRetryIcon && <RefreshCcw className="h-4 w-4" />}
            {children}
        </button>
    );
}

export function WhyStartHere() {
    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-6 text-sm text-cortex-text-muted">
            <p className="font-medium text-cortex-text-main">Why start here</p>
            <p className="mt-2 max-w-4xl leading-7">
                Create the AI Organization first so Mycelis opens with structure, not a one-off assistant session. The organization home keeps the Team Lead, Advisors, Departments, and Specialists visible from the beginning.
            </p>
            <div className="mt-4">
                <Link href="/docs?doc=v8-ui-api-operator-experience-contract" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                    Learn about AI Organizations
                    <ArrowRight className="h-4 w-4" />
                </Link>
            </div>
        </section>
    );
}
