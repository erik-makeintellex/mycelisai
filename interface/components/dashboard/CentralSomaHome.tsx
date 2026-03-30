"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowRight, Blocks, Bot, Building2, Clock3, ShieldCheck, Sparkles } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";

type LastOrganization = {
    id: string;
    name: string;
};

const CENTRAL_PROMPTS = [
    "Review what changed across the active organization",
    "Compare priorities before acting in a deployment",
    "Create a new AI Organization when a fresh context is needed",
];

export default function CentralSomaHome() {
    const [lastOrganization, setLastOrganization] = useState<LastOrganization | null>(null);

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange((organization) => {
            setLastOrganization(organization);
        });
    }, []);

    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-7 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
            <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
                <div className="space-y-5">
                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
                        <Sparkles className="h-3.5 w-3.5" />
                        Central Soma
                    </div>
                    <div className="space-y-3">
                        <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">
                            Work with one Soma across every AI Organization.
                        </h1>
                        <p className="max-w-3xl text-base leading-7 text-cortex-text-muted">
                            Soma and Council stay persistent across organizations, deployments, logs, and retained continuity. AI Organizations stay as governed working contexts where execution, approvals, and artifacts remain explicitly scoped.
                        </p>
                    </div>
                    <div className="flex flex-wrap gap-3">
                        <a
                            href="#create-ai-organization"
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                        >
                            Create AI Organization
                            <Blocks className="h-4 w-4" />
                        </a>
                        {lastOrganization ? (
                            <Link
                                href={`/organizations/${lastOrganization.id}`}
                                className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
                            >
                                Return to {lastOrganization.name}
                                <ArrowRight className="h-4 w-4" />
                            </Link>
                        ) : (
                            <a
                                href="#create-ai-organization"
                                className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
                            >
                                Open setup flow
                                <ArrowRight className="h-4 w-4" />
                            </a>
                        )}
                        <Link
                            href="/docs?doc=v8-universal-soma-context-model"
                            className="inline-flex items-center gap-2 rounded-xl border border-transparent bg-transparent px-2 py-2.5 text-sm font-medium text-cortex-primary transition-colors hover:bg-cortex-primary/10"
                        >
                            Review the Soma context model
                            <ArrowRight className="h-4 w-4" />
                        </Link>
                    </div>
                    <div className="grid gap-3 md:grid-cols-3">
                        {CENTRAL_PROMPTS.map((prompt) => (
                            <div key={prompt} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
                                <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Soma can help with</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-main">{prompt}</p>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-bg p-5">
                    <div>
                        <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">How the product works</p>
                        <h2 className="mt-2 text-xl font-semibold text-cortex-text-main">Universal counterpart, scoped action.</h2>
                    </div>
                    <PrincipleRow
                        icon={<Bot className="h-4 w-4" />}
                        title="One persistent Soma"
                        detail="Soma does not restart its identity for every project. The operator returns to one counterpart."
                    />
                    <PrincipleRow
                        icon={<Building2 className="h-4 w-4" />}
                        title="Organizations are contexts"
                        detail="Each AI Organization is a governed place for teams, specialists, continuity, and delivery decisions."
                    />
                    <PrincipleRow
                        icon={<ShieldCheck className="h-4 w-4" />}
                        title="Execution stays scoped"
                        detail="Answers may span contexts, but mutations, approvals, artifacts, and audit remain tied to the chosen context."
                    />
                    <PrincipleRow
                        icon={<Clock3 className="h-4 w-4" />}
                        title="Continuity stays visible"
                        detail="Recent activity, retained knowledge, and current context remain part of the same Soma-led operating loop."
                    />
                </div>
            </div>
        </section>
    );
}

function PrincipleRow({
    icon,
    title,
    detail,
}: {
    icon: React.ReactNode;
    title: string;
    detail: string;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <div className="flex items-center gap-2 text-cortex-primary">
                {icon}
                <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            </div>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{detail}</p>
        </div>
    );
}
