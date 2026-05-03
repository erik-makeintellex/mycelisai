"use client";

import { useEffect, useState, type ReactNode } from "react";
import Link from "next/link";
import {
    ArrowRight,
    BookOpen,
    Building2,
    CheckCircle2,
    Database,
    ExternalLink,
    FileText,
    ShieldCheck,
    SlidersHorizontal,
    Workflow,
    Wrench,
} from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import {
    defaultHomepageConfig,
    normalizeHomepageConfig,
    type HomepageConfig,
    type HomepageCTA,
    type HomepageLink,
} from "@/lib/homepageConfig";

const outcomes = [
    { title: "Soma orchestration", body: "One primary interface for intent, coordination, review, and output delivery.", icon: Workflow },
    { title: "AI Organizations", body: "Structured teams and roles can be shaped for planning, research, creation, review, and operations.", icon: Building2 },
    { title: "Connected tools", body: "MCP, files, search, host data, and other capabilities remain governed and reviewable.", icon: Wrench },
    { title: "Learning and context", body: "Memory, deployment context, and retained artifacts keep useful work available.", icon: Database },
    { title: "Governance and audit", body: "Approvals, capability boundaries, and auditable activity keep execution controlled.", icon: ShieldCheck },
    { title: "Self-hosted control", body: "Run locally, in Compose, or in Kubernetes with deployment-owned configuration.", icon: SlidersHorizontal },
];

export default function LandingPage() {
    const [config, setConfig] = useState<HomepageConfig>(defaultHomepageConfig);
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange(setLastOrganization);
    }, []);

    useEffect(() => {
        let cancelled = false;
        fetch("/api/v1/homepage", { cache: "no-store" })
            .then((res) => (res.ok ? res.json() : null))
            .then((body) => {
                if (!cancelled && body?.ok && body.data) setConfig(normalizeHomepageConfig(body.data));
            })
            .catch(() => {
                if (!cancelled) setConfig(defaultHomepageConfig);
            });
        return () => {
            cancelled = true;
        };
    }, []);

    const primaryCTA = config.hero.primary_cta;
    const secondaryCTA = config.hero.secondary_cta;

    return (
        <main className="min-h-screen bg-cortex-bg text-cortex-text-main">
            <nav className="sticky top-0 z-50 border-b border-cortex-border bg-cortex-bg/95 backdrop-blur">
                <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-6">
                    <Link href="/" className="flex items-center gap-3">
                        <BrandMark logoUrl={config.brand.logo_url} productName={config.brand.product_name} />
                        <span className="text-sm font-semibold uppercase tracking-[0.22em]">{config.brand.product_name}</span>
                    </Link>
                    <div className="flex items-center gap-2">
                        <ActionLink cta={secondaryCTA} quiet icon={<BookOpen className="h-4 w-4" />} />
                        <ActionLink cta={primaryCTA} icon={<ArrowRight className="h-4 w-4" />} />
                    </div>
                </div>
            </nav>

            <section className="mx-auto grid max-w-7xl gap-10 px-6 py-14 lg:grid-cols-[1fr_0.86fr] lg:items-start">
                <div className="space-y-8">
                    {config.announcement?.enabled && config.announcement.text ? (
                        <p className="inline-flex rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-4 py-1.5 text-sm font-medium text-cortex-primary">
                            {config.announcement.text}
                        </p>
                    ) : null}
                    <div className="space-y-4">
                        <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">{config.brand.tagline}</p>
                        <h1 className="max-w-4xl text-4xl font-semibold leading-tight tracking-tight sm:text-6xl">
                            {config.hero.headline}
                        </h1>
                        <p className="max-w-2xl text-lg leading-8 text-cortex-text-muted">{config.hero.subheadline}</p>
                    </div>
                    <div className="flex flex-col gap-3 sm:flex-row">
                        <ActionLink cta={primaryCTA} large icon={<ArrowRight className="h-4 w-4" />} />
                        {lastOrganization ? (
                            <Link className="inline-flex items-center justify-center rounded-full border border-cortex-border bg-cortex-surface px-6 py-3 text-base font-medium hover:border-cortex-primary/40" href={`/organizations/${lastOrganization.id}`}>
                                Resume {lastOrganization.name}
                            </Link>
                        ) : (
                            <ActionLink cta={secondaryCTA} large quiet />
                        )}
                    </div>
                </div>

                <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5 shadow-[0_18px_50px_rgba(63,67,100,0.12)]">
                    <p className="text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">Example workspace preview</p>
                    <h2 className="mt-3 text-2xl font-semibold">Intent, progress, outcome.</h2>
                    <div className="mt-5 space-y-3">
                        {["User intent", "Soma coordination", "Team/tool execution", "Review and audit"].map((step) => (
                            <div key={step} className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
                                <CheckCircle2 className="h-4 w-4 text-cortex-success" />
                                <span className="text-sm font-medium">{step}</span>
                            </div>
                        ))}
                    </div>
                    <p className="mt-5 text-sm leading-6 text-cortex-text-muted">
                        Illustrative flow only. Live activity, teams, memory, and audit records appear after real governed work runs.
                    </p>
                </section>
            </section>

            <section className="border-y border-cortex-border bg-cortex-surface/45 py-12">
                <div className="mx-auto max-w-7xl px-6">
                    <SectionHeader eyebrow="How it works" title="Intent becomes governed execution." />
                    <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                        {config.sections.map((section, index) => (
                            <StepCard key={`${section.title}-${index}`} number={index + 1} title={section.title} body={section.body} />
                        ))}
                    </div>
                </div>
            </section>

            <section className="mx-auto max-w-7xl px-6 py-12">
                <SectionHeader eyebrow="What you get" title="A self-hostable operating surface for AI work." />
                <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                    {outcomes.map((item) => {
                        const Icon = item.icon;
                        return (
                            <SupportCard key={item.title} icon={<Icon className="h-5 w-5" />} title={item.title} detail={item.body} />
                        );
                    })}
                </div>
            </section>

            <section className="border-t border-cortex-border bg-cortex-surface/45 py-12">
                <div className="mx-auto max-w-7xl px-6">
                    <SectionHeader eyebrow="Deployment resources" title="Route operators to the right deployment surfaces." />
                    <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                        {config.links.map((link) => <ResourceLink key={`${link.label}-${link.href}`} link={link} />)}
                    </div>
                    <p className="mt-8 text-sm text-cortex-text-muted">{config.footer_text}</p>
                    {config.config_issue ? <p className="mt-2 text-xs text-cortex-warning">{config.config_issue}</p> : null}
                </div>
            </section>
        </main>
    );
}

function BrandMark({ logoUrl, productName }: { logoUrl?: string; productName: string }) {
    if (logoUrl) return <img src={logoUrl} alt={`${productName} logo`} className="h-8 w-8 rounded-xl object-contain" />;
    return <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-cortex-primary text-cortex-bg"><Workflow className="h-4 w-4" /></span>;
}

function ActionLink({ cta, icon, quiet = false, large = false }: { cta: HomepageCTA; icon?: ReactNode; quiet?: boolean; large?: boolean }) {
    const className = quiet
        ? `inline-flex items-center justify-center gap-2 rounded-full border border-cortex-border bg-cortex-surface ${large ? "px-6 py-3 text-base" : "px-4 py-2 text-sm"} font-medium hover:border-cortex-primary/40`
        : `inline-flex items-center justify-center gap-2 rounded-full bg-cortex-primary ${large ? "px-6 py-3 text-base" : "px-4 py-2 text-sm"} font-semibold text-cortex-bg hover:bg-cortex-primary/90`;
    if (cta.external) return <a className={className} href={cta.href} target="_blank" rel="noopener noreferrer">{cta.label}{icon}</a>;
    return <Link className={className} href={cta.href}>{cta.label}{icon}</Link>;
}

function SectionHeader({ eyebrow, title }: { eyebrow: string; title: string }) {
    return (
        <div>
            <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">{eyebrow}</p>
            <h2 className="mt-3 text-3xl font-semibold">{title}</h2>
        </div>
    );
}

function StepCard({ number, title, body }: { number: number; title: string; body: string }) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-bg p-5">
            <span className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-2.5 py-1 text-xs font-semibold text-cortex-primary">{number}</span>
            <h3 className="mt-4 text-lg font-semibold">{title}</h3>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{body}</p>
        </div>
    );
}

function SupportCard({ icon, title, detail }: { icon: ReactNode; title: string; detail: string }) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="inline-flex rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 p-3 text-cortex-primary">{icon}</div>
            <h3 className="mt-5 text-xl font-semibold">{title}</h3>
            <p className="mt-3 text-sm leading-7 text-cortex-text-muted">{detail}</p>
        </div>
    );
}

function ResourceLink({ link }: { link: HomepageLink }) {
    const content = (
        <>
            <div className="flex items-center gap-2 text-base font-semibold">
                {link.external ? <ExternalLink className="h-4 w-4 text-cortex-primary" /> : <FileText className="h-4 w-4 text-cortex-primary" />}
                {link.label}
            </div>
            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">{link.description}</p>
        </>
    );
    const className = "block rounded-3xl border border-cortex-border bg-cortex-bg p-5 transition hover:border-cortex-primary/35 hover:bg-cortex-surface";
    if (link.external) return <a className={className} href={link.href} target="_blank" rel="noopener noreferrer">{content}</a>;
    return <Link className={className} href={link.href}>{content}</Link>;
}
