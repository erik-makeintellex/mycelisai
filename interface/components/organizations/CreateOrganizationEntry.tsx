"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { ArrowRight, Blocks, Bot, BrainCircuit, Building2, Layers3, Loader2, RefreshCcw, Sparkles } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type {
    OrganizationCreateRequest,
    OrganizationStartMode,
    OrganizationSummary,
    OrganizationTemplateSummary,
} from "@/lib/organizations";

type ResourceState = "loading" | "ready" | "error";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export default function CreateOrganizationEntry() {
    const router = useRouter();
    const [templates, setTemplates] = useState<OrganizationTemplateSummary[]>([]);
    const [organizations, setOrganizations] = useState<OrganizationSummary[]>([]);
    const [templatesState, setTemplatesState] = useState<ResourceState>("loading");
    const [organizationsState, setOrganizationsState] = useState<ResourceState>("loading");
    const [templatesError, setTemplatesError] = useState<string | null>(null);
    const [organizationsError, setOrganizationsError] = useState<string | null>(null);
    const [templatesReloadToken, setTemplatesReloadToken] = useState(0);
    const [organizationsReloadToken, setOrganizationsReloadToken] = useState(0);
    const [selectedMode, setSelectedMode] = useState<OrganizationStartMode | null>(null);
    const [selectedTemplateId, setSelectedTemplateId] = useState<string>("");
    const [name, setName] = useState("");
    const [purpose, setPurpose] = useState("");
    const [submitError, setSubmitError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [openingOrganizationId, setOpeningOrganizationId] = useState<string | null>(null);
    const [, startTransition] = useTransition();

    const rememberOrganization = (organization: Pick<OrganizationSummary, "id" | "name">) => {
        if (typeof window === "undefined") {
            return;
        }
        window.localStorage.setItem("mycelis-last-organization-id", organization.id);
        window.localStorage.setItem("mycelis-last-organization-name", organization.name);
    };

    const openOrganization = (organization: Pick<OrganizationSummary, "id" | "name">) => {
        if (openingOrganizationId === organization.id || isSubmitting) {
            return;
        }
        setOpeningOrganizationId(organization.id);
        rememberOrganization(organization);
        startTransition(() => {
            router.push(`/organizations/${organization.id}`);
        });
    };

    useEffect(() => {
        let cancelled = false;

        const loadTemplates = async () => {
            setTemplatesState("loading");
            setTemplatesError(null);
            try {
                const response = await fetch("/api/v1/templates?view=organization-starters", { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Starter templates are unavailable right now.");
                }
                if (cancelled) {
                    return;
                }

                const nextTemplates = extractApiData<OrganizationTemplateSummary[]>(payload) || [];
                setTemplates(nextTemplates);
                setSelectedTemplateId((current) => {
                    if (current && nextTemplates.some((template) => template.id === current)) {
                        return current;
                    }
                    return nextTemplates[0]?.id ?? "";
                });
                setTemplatesState("ready");
            } catch (err) {
                if (cancelled) {
                    return;
                }
                setTemplatesError(err instanceof Error ? err.message : "Starter templates are unavailable right now.");
                setTemplatesState("error");
            }
        };

        void loadTemplates();
        return () => {
            cancelled = true;
        };
    }, [templatesReloadToken]);

    useEffect(() => {
        let cancelled = false;

        const loadOrganizations = async () => {
            setOrganizationsState("loading");
            setOrganizationsError(null);
            try {
                const response = await fetch("/api/v1/organizations?view=summary", { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Recent AI Organizations are unavailable right now.");
                }
                if (cancelled) {
                    return;
                }

                setOrganizations(extractApiData<OrganizationSummary[]>(payload) || []);
                setOrganizationsState("ready");
            } catch (err) {
                if (cancelled) {
                    return;
                }
                setOrganizationsError(err instanceof Error ? err.message : "Recent AI Organizations are unavailable right now.");
                setOrganizationsState("error");
            }
        };

        void loadOrganizations();
        return () => {
            cancelled = true;
        };
    }, [organizationsReloadToken]);

    const selectedTemplate = templates.find((template) => template.id === selectedTemplateId) ?? null;
    const isTemplateMode = selectedMode === "template";
    const canSubmit =
        name.trim().length > 1 &&
        purpose.trim().length > 8 &&
        selectedMode !== null &&
        (!isTemplateMode || (templatesState === "ready" && !!selectedTemplate));

    const handleCreate = async () => {
        if (!canSubmit || isSubmitting || selectedMode === null) {
            return;
        }

        const payload: OrganizationCreateRequest = {
            name: name.trim(),
            purpose: purpose.trim(),
            start_mode: selectedMode,
            template_id: isTemplateMode ? selectedTemplate?.id : undefined,
        };

        setIsSubmitting(true);
        setSubmitError(null);
        try {
            const response = await fetch("/api/v1/organizations", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });
            const responsePayload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(responsePayload) || "Unable to create AI Organization.");
            }
            const created = extractApiData<OrganizationSummary>(responsePayload);
            if (!created?.id) {
                throw new Error("AI Organization created, but the workspace route is unavailable right now.");
            }
            rememberOrganization(created);
            startTransition(() => {
                router.push(`/organizations/${created.id}`);
            });
        } catch (err) {
            setSubmitError(err instanceof Error ? err.message : "Unable to create AI Organization.");
        } finally {
            setIsSubmitting(false);
        }
    };

    const showRecentOrganizations = organizationsState !== "ready" || organizations.length > 0;

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_24px_60px_rgba(0,0,0,0.18)]">
                    <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
                        <div className="max-w-3xl space-y-4">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
                                <Sparkles className="h-3.5 w-3.5" />
                                AI Organization Setup
                            </div>
                            <div className="space-y-3">
                                <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">
                                    Create AI Organization
                                </h1>
                                <p className="max-w-2xl text-base leading-7 text-cortex-text-muted">
                                    Start Mycelis by creating an AI Organization, not a blank assistant session. Choose a starting point, define the purpose, and open the new organization home.
                                </p>
                            </div>
                            <div className="flex flex-wrap gap-3">
                                <ActionButton onClick={() => setSelectedMode("template")} kind="primary">
                                    Explore Templates
                                </ActionButton>
                                <ActionButton onClick={() => setSelectedMode("empty")} kind="secondary">
                                    Start Empty
                                </ActionButton>
                            </div>
                        </div>
                            <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                <p className="font-medium text-cortex-text-main">Hidden until Advanced mode</p>
                                <p className="mt-1 leading-6">
                                    AI Engine Settings and Learning &amp; Context stay out of the default flow until the operator intentionally opens advanced controls.
                                </p>
                            </div>
                    </div>
                </section>

                {showRecentOrganizations && (
                    <section className="space-y-3">
                        <div>
                            <h2 className="text-lg font-semibold text-cortex-text-main">Recent AI Organizations</h2>
                            <p className="text-sm text-cortex-text-muted">Reopen an existing organization when you want to continue instead of starting over.</p>
                        </div>

                        {organizationsState === "loading" && (
                            <LoadingState label="Loading recent AI Organizations..." />
                        )}

                        {organizationsState === "error" && organizationsError && (
                            <ActionableState
                                title="Recent AI Organizations are unavailable"
                                message={organizationsError}
                                guidance="You can still create a new AI Organization below while we retry your recent organizations."
                                actions={
                                    <ActionButton onClick={() => setOrganizationsReloadToken((value) => value + 1)} kind="secondary" showRetryIcon>
                                        Retry recent AI Organizations
                                    </ActionButton>
                                }
                            />
                        )}

                        {organizationsState === "ready" && organizations.length > 0 && (
                            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                                {organizations.map((organization) => (
                                    <button
                                        key={organization.id}
                                        type="button"
                                        onClick={() => openOrganization(organization)}
                                        disabled={openingOrganizationId === organization.id || isSubmitting}
                                        className="rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-left transition-colors hover:border-cortex-primary/25 disabled:cursor-not-allowed disabled:opacity-70"
                                        aria-busy={openingOrganizationId === organization.id}
                                    >
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <p className="text-base font-semibold text-cortex-text-main">{organization.name}</p>
                                                <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{organization.purpose}</p>
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
                                            {openingOrganizationId === organization.id ? (
                                                <Loader2 className="h-4 w-4 animate-spin" />
                                            ) : null}
                                            Open AI Organization
                                            <ArrowRight className="h-4 w-4" />
                                        </div>
                                    </button>
                                ))}
                            </div>
                        )}
                    </section>
                )}

                <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
                    <div className="space-y-6 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">Choose how to start</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">
                                Shape an AI Organization first so Mycelis feels like an operating environment, not a generic chat.
                            </p>
                        </div>

                        <div className="grid gap-3 md:grid-cols-2">
                            <StartModeCard
                                active={selectedMode === "template"}
                                title="Start from template"
                                description="Use a starter that already defines a Team Lead, initial Departments, and Specialist structure."
                                icon={<Layers3 className="h-5 w-5" />}
                                onClick={() => setSelectedMode("template")}
                            />
                            <StartModeCard
                                active={selectedMode === "empty"}
                                title="Start empty"
                                description="Begin with a clean AI Organization and shape the structure after the organization exists."
                                icon={<Building2 className="h-5 w-5" />}
                                onClick={() => setSelectedMode("empty")}
                            />
                        </div>

                        {templatesState === "loading" && (
                            <LoadingState label="Loading starter templates..." />
                        )}

                        {templatesState === "error" && templatesError && (
                            <ActionableState
                                title="Starter templates are unavailable"
                                message={templatesError}
                                guidance="You can retry, start empty now, or reopen a recent AI Organization if one is available above."
                                actions={
                                    <>
                                        <ActionButton onClick={() => setTemplatesReloadToken((value) => value + 1)} kind="secondary" showRetryIcon>
                                            Retry starters
                                        </ActionButton>
                                        <ActionButton onClick={() => setSelectedMode("empty")} kind="ghost">
                                            Start empty instead
                                        </ActionButton>
                                    </>
                                }
                            />
                        )}

                        {templatesState === "ready" && isTemplateMode && templates.length === 0 && (
                            <EmptyState
                                title="No starter templates available"
                                detail="You can still create an AI Organization with Start empty while starter templates are unavailable."
                                actions={
                                    <ActionButton onClick={() => setSelectedMode("empty")} kind="secondary">
                                        Start empty instead
                                    </ActionButton>
                                }
                            />
                        )}

                        {templatesState === "ready" && isTemplateMode && templates.length > 0 && (
                            <div className="space-y-3">
                                <div>
                                    <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-cortex-text-muted">Starter templates</h3>
                                    <p className="mt-1 text-sm text-cortex-text-muted">Choose an AI Organization starter using operator-facing terms only.</p>
                                </div>
                                <div className="grid gap-3">
                                    {templates.map((template) => (
                                        <button
                                            key={template.id}
                                            onClick={() => setSelectedTemplateId(template.id)}
                                            className={`rounded-2xl border p-4 text-left transition-colors ${
                                                selectedTemplateId === template.id
                                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                                            }`}
                                        >
                                            <div className="flex items-start justify-between gap-3">
                                                <div>
                                                    <p className="text-base font-semibold text-cortex-text-main">{template.name}</p>
                                                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{template.description}</p>
                                                </div>
                                                <span className="rounded-full border border-cortex-border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                                                    {template.organization_type}
                                                </span>
                                            </div>
                                            <div className="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
                                                <Metric label="Team Lead" value={template.team_lead_label} />
                                                <Metric label="Advisors" value={String(template.advisor_count)} />
                                                <Metric label="Departments" value={String(template.department_count)} />
                                                <Metric label="Specialists" value={String(template.specialist_count)} />
                                                <Metric label="AI Engine Settings" value={template.ai_engine_settings_summary} />
                                                <Metric label="Learning & Context" value={template.memory_personality_summary} />
                                            </div>
                                        </button>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>

                    <div className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">Define the AI Organization</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">
                                Focus on the organization name, purpose, and starting point. Advanced details stay tucked away until you choose to open them.
                            </p>
                        </div>

                        <label className="block space-y-2">
                            <span className="text-sm font-medium text-cortex-text-main">AI Organization name</span>
                            <input
                                value={name}
                                onChange={(event) => setName(event.target.value)}
                                placeholder="Northstar Labs"
                                className="w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                            />
                        </label>

                        <label className="block space-y-2">
                            <span className="text-sm font-medium text-cortex-text-main">Purpose</span>
                            <textarea
                                value={purpose}
                                onChange={(event) => setPurpose(event.target.value)}
                                placeholder="Ship a focused AI engineering organization for product delivery."
                                className="min-h-[120px] w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                            />
                        </label>

                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">Selected start</p>
                            <p className="mt-1 leading-6">
                                {selectedMode === "template" && selectedTemplate
                                    ? `${selectedTemplate.name} will shape the first Team Lead, Departments, and Specialists for this AI Organization.`
                                    : selectedMode === "empty"
                                      ? "Start empty creates the AI Organization first so you can shape Advisors, Departments, and Specialists afterward."
                                      : "Choose Start from template or Start empty to continue."}
                            </p>
                        </div>

                        {submitError && (
                            <ActionableState
                                title="Unable to create AI Organization"
                                message={submitError}
                                guidance="Check the organization name, purpose, and chosen starting point, then try again."
                                actions={
                                    <>
                                        <ActionButton onClick={handleCreate} kind="secondary" disabled={!canSubmit || isSubmitting} showRetryIcon>
                                            Try again
                                        </ActionButton>
                                        {isTemplateMode && (
                                            <ActionButton onClick={() => setSelectedMode("empty")} kind="ghost">
                                                Start empty instead
                                            </ActionButton>
                                        )}
                                    </>
                                }
                            />
                        )}

                        <button
                            onClick={handleCreate}
                            disabled={!canSubmit || isSubmitting}
                            className="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-cortex-primary/35 bg-cortex-primary/10 px-4 py-3 text-sm font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Blocks className="h-4 w-4" />}
                            Create AI Organization
                        </button>
                        {isSubmitting && (
                            <p className="text-sm text-cortex-text-muted">
                                Opening your Soma workspace as soon as the organization is ready.
                            </p>
                        )}

                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">What stays hidden for now</p>
                            <div className="mt-3 grid gap-2 sm:grid-cols-2">
                                <HiddenLater label="AI Engine Settings" icon={<Bot className="h-4 w-4" />} />
                                <HiddenLater label="Learning & Context" icon={<BrainCircuit className="h-4 w-4" />} />
                            </div>
                            <p className="mt-3 text-xs leading-6 text-cortex-text-muted">
                                Advanced setup opens later so you can focus on the organization name, purpose, and starting point first.
                            </p>
                        </div>
                    </div>
                </section>

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
            </div>
        </div>
    );
}

function StartModeCard({
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
        <button
            onClick={onClick}
            className={`rounded-2xl border p-4 text-left transition-colors ${
                active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
            }`}
        >
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

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface/60 px-3 py-2">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

function HiddenLater({ label, icon }: { label: string; icon: React.ReactNode }) {
    return (
        <div className="flex items-center gap-2 rounded-xl border border-cortex-border px-3 py-2">
            <span className="text-cortex-primary">{icon}</span>
            <span>{label}</span>
        </div>
    );
}

function LoadingState({ label }: { label: string }) {
    return (
        <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>{label}</span>
        </div>
    );
}

function ActionableState({
    title,
    message,
    guidance,
    actions,
}: {
    title: string;
    message: string;
    guidance: string;
    actions?: React.ReactNode;
}) {
    return (
        <div className="rounded-2xl border border-cortex-danger/30 bg-cortex-danger/10 px-4 py-4 text-sm text-cortex-text-main">
            <p className="font-semibold">{title}</p>
            <p className="mt-2 leading-6">{message}</p>
            <p className="mt-2 text-cortex-text-muted">{guidance}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

function EmptyState({
    title,
    detail,
    actions,
}: {
    title: string;
    detail: string;
    actions?: React.ReactNode;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

function ActionButton({
    children,
    onClick,
    kind,
    disabled = false,
    showRetryIcon = false,
}: {
    children: React.ReactNode;
    onClick: () => void;
    kind: "primary" | "secondary" | "ghost";
    disabled?: boolean;
    showRetryIcon?: boolean;
}) {
    const classes =
        kind === "primary"
            ? "border border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/15"
            : kind === "secondary"
              ? "border border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
              : "border border-transparent bg-transparent text-cortex-primary hover:bg-cortex-primary/10";

    return (
        <button
            onClick={onClick}
            disabled={disabled}
            className={`inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${classes}`}
        >
            {showRetryIcon && <RefreshCcw className="h-4 w-4" />}
            {children}
        </button>
    );
}
