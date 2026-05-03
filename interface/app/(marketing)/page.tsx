"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import Link from "next/link";
import {
    ArrowRight,
    BookOpen,
    CheckCircle2,
    Compass,
    FileText,
    Search,
    Settings,
    Sparkles,
    Wrench,
} from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";

const intentCards = [
    {
        title: "Plan",
        detail: "Turn a goal into next steps, owners, and visible checkpoints.",
        icon: Compass,
        prompt: "Help me plan the next useful step.",
    },
    {
        title: "Research",
        detail: "Ask Soma to search or review sources, then summarize what matters.",
        icon: Search,
        prompt: "Research this and show me the sources.",
    },
    {
        title: "Create",
        detail: "Draft content, artifacts, teams, or workflows with review before risky changes.",
        icon: Sparkles,
        prompt: "Create a first version and tell me where it was stored.",
    },
    {
        title: "Review",
        detail: "Review work, decisions, risks, and recent activity before acting.",
        icon: CheckCircle2,
        prompt: "Review this and recommend what I should approve.",
    },
    {
        title: "Configure tools",
        detail: "Connect search, files, MCP servers, and other tools when they are needed.",
        icon: Wrench,
        prompt: "Check which tools are available and walk me through enabling missing ones.",
    },
];

const feedbackSteps = [
    "What Soma understood",
    "What Soma is doing",
    "What changed",
    "Where the output was stored",
];

export default function LandingPage() {
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange(setLastOrganization);
    }, []);

    return (
        <main className="min-h-screen bg-cortex-bg text-cortex-text-main">
            <nav className="sticky top-0 z-50 border-b border-cortex-border bg-cortex-bg/95 backdrop-blur">
                <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-6">
                    <Link href="/" className="flex items-center gap-3">
                        <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-cortex-primary text-cortex-bg">
                            <Sparkles className="h-4 w-4" />
                        </span>
                        <span className="text-sm font-semibold uppercase tracking-[0.22em]">Mycelis</span>
                    </Link>
                    <div className="flex items-center gap-2">
                        <Link
                            href="/docs"
                            className="inline-flex items-center gap-2 rounded-full border border-cortex-border px-4 py-2 text-sm font-medium text-cortex-text-muted hover:text-cortex-text-main"
                        >
                            <BookOpen className="h-4 w-4" />
                            Docs
                        </Link>
                        <Link
                            href="/dashboard"
                            className="inline-flex items-center gap-2 rounded-full bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg hover:bg-cortex-primary/90"
                        >
                            Start with Soma
                            <ArrowRight className="h-4 w-4" />
                        </Link>
                    </div>
                </div>
            </nav>

            <section className="mx-auto grid max-w-7xl gap-10 px-6 py-14 lg:grid-cols-[1fr_0.85fr] lg:items-start">
                <div className="space-y-8">
                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-4 py-1.5 text-sm font-medium text-cortex-primary">
                        <Sparkles className="h-4 w-4" />
                        Soma-first AI work
                    </div>
                    <div className="space-y-4">
                        <h1 className="max-w-4xl text-4xl font-semibold leading-tight tracking-tight sm:text-6xl">
                            What do you want Soma to do?
                        </h1>
                        <p className="max-w-2xl text-lg leading-8 text-cortex-text-muted">
                            Start with an intent. Soma turns it into understandable work, shows progress,
                            connects outputs to activity and artifacts, and keeps advanced tools out of the
                            way until you need them.
                        </p>
                    </div>
                    <div className="flex flex-col gap-3 sm:flex-row">
                        <Link
                            href="/dashboard"
                            className="inline-flex items-center justify-center gap-2 rounded-full bg-cortex-primary px-6 py-3 text-base font-semibold text-cortex-bg hover:bg-cortex-primary/90"
                        >
                            Ask Soma
                            <ArrowRight className="h-4 w-4" />
                        </Link>
                        {lastOrganization ? (
                            <Link
                                href={`/organizations/${lastOrganization.id}`}
                                className="inline-flex items-center justify-center rounded-full border border-cortex-border bg-cortex-surface px-6 py-3 text-base font-medium hover:border-cortex-primary/40"
                            >
                                Resume {lastOrganization.name}
                            </Link>
                        ) : (
                            <Link
                                href="/dashboard#dashboard-organization-setup"
                                className="inline-flex items-center justify-center rounded-full border border-cortex-border bg-cortex-surface px-6 py-3 text-base font-medium hover:border-cortex-primary/40"
                            >
                                Create AI Organization
                            </Link>
                        )}
                    </div>
                </div>

                <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5 shadow-[0_18px_50px_rgba(63,67,100,0.12)]">
                    <p className="text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">
                        Default experience
                    </p>
                    <h2 className="mt-3 text-2xl font-semibold">Intent, progress, outcome.</h2>
                    <div className="mt-5 space-y-3">
                        {feedbackSteps.map((step) => (
                            <div key={step} className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
                                <CheckCircle2 className="h-4 w-4 text-cortex-success" />
                                <span className="text-sm font-medium">{step}</span>
                            </div>
                        ))}
                    </div>
                    <p className="mt-5 text-sm leading-6 text-cortex-text-muted">
                        Groups, runs, memory, MCP, auth, and audit remain available as advanced
                        support systems. The first product moment is asking Soma for an outcome.
                    </p>
                </section>
            </section>

            <section className="border-y border-cortex-border bg-cortex-surface/45 py-12">
                <div className="mx-auto max-w-7xl px-6">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
                        <div>
                            <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">
                                Start with intent
                            </p>
                            <h2 className="mt-3 text-3xl font-semibold">Choose the shape of work.</h2>
                        </div>
                        <Link href="/dashboard" className="inline-flex items-center gap-2 text-sm font-semibold text-cortex-primary">
                            Open Soma workspace
                            <ArrowRight className="h-4 w-4" />
                        </Link>
                    </div>
                    <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
                        {intentCards.map((card) => {
                            const Icon = card.icon;
                            return (
                                <Link
                                    key={card.title}
                                    href={`/dashboard?intent=${encodeURIComponent(card.title.toLowerCase())}`}
                                    className="rounded-3xl border border-cortex-border bg-cortex-bg p-5 transition hover:border-cortex-primary/35 hover:bg-cortex-surface"
                                >
                                    <Icon className="h-5 w-5 text-cortex-primary" />
                                    <h3 className="mt-4 text-lg font-semibold">{card.title}</h3>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{card.detail}</p>
                                    <p className="mt-4 text-xs font-medium text-cortex-primary">{card.prompt}</p>
                                </Link>
                            );
                        })}
                    </div>
                </div>
            </section>

            <section className="mx-auto grid max-w-7xl gap-6 px-6 py-12 lg:grid-cols-3">
                <SupportCard
                    icon={<FileText className="h-5 w-5" />}
                    title="Outputs stay visible"
                    detail="Soma should connect completed work to activity, learning, artifacts, teams, and storage locations."
                />
                <SupportCard
                    icon={<Settings className="h-5 w-5" />}
                    title="Advanced stays available"
                    detail="Admins can open tools, memory, runs, audit, auth, and MCP when they need operational depth."
                />
                <SupportCard
                    icon={<BookOpen className="h-5 w-5" />}
                    title="Docs explain the system"
                    detail="New users should not need architecture knowledge before getting value from Soma."
                />
            </section>
        </main>
    );
}

function SupportCard({
    icon,
    title,
    detail,
}: {
    icon: ReactNode;
    title: string;
    detail: string;
}) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="inline-flex rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 p-3 text-cortex-primary">
                {icon}
            </div>
            <h3 className="mt-5 text-xl font-semibold">{title}</h3>
            <p className="mt-3 text-sm leading-7 text-cortex-text-muted">{detail}</p>
        </div>
    );
}
