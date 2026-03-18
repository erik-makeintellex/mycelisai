"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, Blocks, Loader2 } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type { OrganizationHomePayload } from "@/lib/organizations";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export default function OrganizationContextShell({ organizationId }: { organizationId: string }) {
    const [organization, setOrganization] = useState<OrganizationHomePayload | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

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

        load();
        return () => {
            cancelled = true;
        };
    }, [organizationId]);

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
                    <div className="mt-5">
                        <Link href="/dashboard" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                            <ArrowLeft className="h-4 w-4" />
                            Return to Create AI Organization
                        </Link>
                    </div>
                </div>
            </div>
        );
    }

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
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">Current context shell</p>
                            <p className="mt-1 leading-6">
                                This first UI slice lands in AI Organization context before the Team Lead workspace itself ships.
                            </p>
                        </div>
                    </div>
                </section>

                <section className="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
                    <div className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">Organization header</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">User-facing context only. Advanced configuration remains hidden for this slice.</p>
                        </div>
                        <div className="grid gap-3 md:grid-cols-2">
                            <Metric label="Team Lead" value={organization.team_lead_label} />
                            <Metric label="Started from" value={organization.start_mode === "template" ? (organization.template_name || "Template") : "Empty"} />
                            <Metric label="Advisors" value={String(organization.advisor_count)} />
                            <Metric label="Departments" value={String(organization.department_count)} />
                            <Metric label="Specialists" value={String(organization.specialist_count)} />
                            <Metric label="AI Organization" value={organization.status} />
                        </div>
                    </div>

                    <div className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">What comes next</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">The next bounded slice turns this context shell into the Team Lead-first workspace.</p>
                        </div>
                        <div className="space-y-3">
                            <Metric label="AI Engine Settings" value={organization.ai_engine_settings_summary} />
                            <Metric label="Memory & Personality" value={organization.memory_personality_summary} />
                            <button
                                disabled
                                className="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-cortex-border px-4 py-3 text-sm font-medium text-cortex-text-muted opacity-80"
                            >
                                Team Lead workspace is next
                                <ArrowRight className="h-4 w-4" />
                            </button>
                            <Link href="/dashboard" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                                <ArrowLeft className="h-4 w-4" />
                                Create another AI Organization
                            </Link>
                        </div>
                    </div>
                </section>
            </div>
        </div>
    );
}

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}
