import type { ReactNode } from "react";
import Link from "next/link";
import { Activity, ArrowRight, Blocks, Bot, BrainCircuit, Compass, ShieldCheck, Sparkles, Users } from "lucide-react";

const truthCards = [
    {
        title: "Structure",
        eyebrow: "Teams and specialists",
        description:
            "Design an AI Organization with Soma, supporting Team Leads, Advisors, focused Departments, and Specialists that stay visible from the start.",
        bullets: [
            "Create a clear organization shape before work begins",
            "Keep Soma, Team Leads, Advisors, Departments, and Specialists easy to understand",
            "Open with an organization workspace, not a blank assistant session",
        ],
        icon: <Blocks className="h-5 w-5" />,
    },
    {
        title: "Control",
        eyebrow: "Guided tuning",
        description:
            "Adjust AI Engine Settings and Response Style through safe guided choices so output stays intentional, reviewable, and easy to manage.",
        bullets: [
            "Use curated AI Engine options for planning depth and pace",
            "Choose a Response Style that matches tone, structure, and detail",
            "Keep advanced controls hidden until they are truly needed",
        ],
        icon: <ShieldCheck className="h-5 w-5" />,
    },
    {
        title: "Continuous Operation",
        eyebrow: "Reviews, checks, and updates",
        description:
            "See recent reviews and checks so the organization feels active between operator actions without turning the workspace into a control panel.",
        bullets: [
            "Keep Recent Activity visible inside the Soma workspace",
            "Surface updates in user-facing language instead of technical noise",
            "Reinforce that your AI Organization keeps watch on ongoing work",
        ],
        icon: <Activity className="h-5 w-5" />,
    },
];

const postCreationSteps = [
    {
        title: "Enter the Soma workspace",
        description:
            "Open into an AI Organization home where Soma is the primary working counterpart and the organization context stays visible.",
        icon: <Bot className="h-5 w-5" />,
    },
    {
        title: "See Advisors and Departments",
        description:
            "Understand who supports Soma, what each Department handles, and how Specialists fit into the organization.",
        icon: <Users className="h-5 w-5" />,
    },
    {
        title: "Watch Recent Activity",
        description:
            "Track recent reviews, checks, and updates so the organization feels active and supervised without extra setup work.",
        icon: <Activity className="h-5 w-5" />,
    },
    {
        title: "Follow guided next steps",
        description:
            "Use Soma guidance to plan the next move, review setup choices, and keep the organization moving in a structured way.",
        icon: <Compass className="h-5 w-5" />,
    },
];

const workspaceHighlights = [
    {
        label: "Soma",
        value: "Primary counterpart",
    },
    {
        label: "AI Engine Settings",
        value: "Guided, safe choices",
    },
    {
        label: "Response Style",
        value: "Consistent output behavior",
    },
];

const recentActivityPreview = [
    {
        name: "Strategy review",
        time: "2 minutes ago",
        outcome: "No issues detected",
    },
    {
        name: "Delivery check",
        time: "10 minutes ago",
        outcome: "2 items flagged",
    },
    {
        name: "Readiness update",
        time: "18 minutes ago",
        outcome: "Soma guidance refreshed",
    },
];

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-cortex-bg text-cortex-text-main">
            <div className="absolute inset-x-0 top-0 -z-10 h-[36rem] bg-[radial-gradient(circle_at_top,rgba(34,211,238,0.16),transparent_56%)]" />

            <nav className="sticky top-0 z-50 border-b border-cortex-border bg-cortex-bg/90 backdrop-blur">
                <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
                    <Link href="/" className="flex items-center gap-3">
                        <span className="h-3 w-3 rounded-full bg-cortex-primary shadow-[0_0_16px_rgba(34,211,238,0.7)]" />
                        <span className="text-sm font-semibold uppercase tracking-[0.22em] text-cortex-text-main">Mycelis</span>
                    </Link>
                    <div className="hidden items-center gap-6 text-sm text-cortex-text-muted md:flex">
                        <Link href="#structure" className="transition-colors hover:text-cortex-text-main">
                            Structure
                        </Link>
                        <Link href="#control" className="transition-colors hover:text-cortex-text-main">
                            Control
                        </Link>
                        <Link href="#activity" className="transition-colors hover:text-cortex-text-main">
                            Activity
                        </Link>
                        <Link href="#after-creation" className="transition-colors hover:text-cortex-text-main">
                            After Creation
                        </Link>
                        <Link href="/docs" className="transition-colors hover:text-cortex-text-main">
                            Docs
                        </Link>
                    </div>
                    <Link
                        href="/dashboard"
                        className="inline-flex items-center rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-4 py-2 text-sm font-medium text-cortex-primary transition-colors hover:bg-cortex-primary hover:text-cortex-bg"
                    >
                        Create AI Organization
                    </Link>
                </div>
            </nav>

            <main>
                <section className="mx-auto max-w-7xl px-6 pb-20 pt-20 md:pb-28 md:pt-28">
                    <div className="grid gap-14 lg:grid-cols-[1.1fr_0.9fr] lg:items-center">
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-1.5 text-sm text-cortex-primary">
                                <Sparkles className="h-4 w-4" />
                                Living AI Organization platform
                            </div>

                            <div className="space-y-5">
                                <h1 className="max-w-4xl text-5xl font-semibold leading-tight tracking-tight text-cortex-text-main md:text-7xl">
                                    Build AI Organizations that think, review, and evolve.
                                </h1>
                                <p className="max-w-2xl text-lg leading-8 text-cortex-text-muted md:text-xl">
                                    Mycelis starts with an AI Organization, keeps Soma at the center of the workspace,
                                    and makes continuous reviews, checks, and updates visible without turning the product into a
                                    generic assistant experience.
                                </p>
                            </div>

                            <div className="flex flex-col gap-3 sm:flex-row">
                                <Link
                                    href="/dashboard"
                                    className="inline-flex items-center justify-center gap-2 rounded-full bg-cortex-primary px-6 py-3 text-base font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                                >
                                    Create AI Organization
                                    <ArrowRight className="h-4 w-4" />
                                </Link>
                                <Link
                                    href="/dashboard"
                                    className="inline-flex items-center justify-center rounded-full border border-cortex-border bg-cortex-surface px-6 py-3 text-base font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 hover:text-cortex-primary"
                                >
                                    Explore Templates
                                </Link>
                            </div>

                            <div className="grid gap-4 md:grid-cols-3">
                                {workspaceHighlights.map((item) => (
                                    <div key={item.label} className="rounded-3xl border border-cortex-border bg-cortex-surface/80 px-5 py-4">
                                        <p className="text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">{item.label}</p>
                                        <p className="mt-2 text-sm font-semibold text-cortex-text-main">{item.value}</p>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className="relative">
                            <div className="absolute inset-0 rounded-[2rem] bg-[radial-gradient(circle_at_top,rgba(34,211,238,0.18),transparent_60%)] blur-2xl" />
                            <div className="relative space-y-5 rounded-[2rem] border border-cortex-border bg-cortex-surface/95 p-6 shadow-[0_28px_80px_rgba(29,42,53,0.10)]">
                                <div className="inline-flex items-center gap-2 rounded-full border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-warning">
                                    Illustrative preview
                                </div>
                                <div className="rounded-3xl border border-cortex-border bg-cortex-bg px-5 py-5">
                                    <div className="flex items-start justify-between gap-4">
                                        <div>
                                            <p className="text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">Organization home</p>
                                            <h2 className="mt-2 text-2xl font-semibold text-cortex-text-main">Northstar Labs</h2>
                                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                Soma-guided delivery organization for product planning, specialist review, and steady execution.
                                            </p>
                                        </div>
                                        <span className="rounded-full border border-cortex-success/30 bg-cortex-success/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-success">
                                            Example state
                                        </span>
                                    </div>
                                </div>

                                <div className="grid gap-4 md:grid-cols-[1.2fr_0.8fr]">
                                    <div className="rounded-3xl border border-cortex-border bg-cortex-bg px-5 py-5">
                                        <p className="text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">Work with Soma</p>
                                        <p className="mt-3 text-lg font-semibold text-cortex-text-main">Soma for Northstar Labs</p>
                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                            Guides planning, reviews structure, and recommends the next move while keeping the wider AI Organization visible.
                                        </p>
                                        <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-main">
                                            What I can help with: organize priorities, review setup choices, and turn the next step into a practical plan.
                                        </div>
                                    </div>

                                    <div className="space-y-4">
                                        <PreviewCard title="Advisors" description="Decision support and review coverage" />
                                        <PreviewCard title="Departments" description="Focused lanes for coordinated execution" />
                                        <PreviewCard title="Specialists" description="Repeatable role behavior inside each team" />
                                    </div>
                                </div>

                                <div className="rounded-3xl border border-cortex-border bg-cortex-bg px-5 py-5">
                                    <div className="flex items-center gap-3">
                                        <Activity className="h-5 w-5 text-cortex-primary" />
                                        <div>
                                            <p className="text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">Recent Activity</p>
                                            <p className="mt-1 text-sm text-cortex-text-muted">
                                                Your AI Organization stays visibly active through recent reviews, checks, and updates.
                                            </p>
                                        </div>
                                    </div>
                                    <div className="mt-4 space-y-3">
                                        {recentActivityPreview.map((item) => (
                                            <div key={item.name} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                <div className="flex items-center justify-between gap-4">
                                                    <p className="text-sm font-semibold text-cortex-text-main">{item.name}</p>
                                                    <p className="text-xs font-medium uppercase tracking-[0.16em] text-cortex-text-muted">{item.time}</p>
                                                </div>
                                                <p className="mt-2 text-sm text-cortex-text-muted">{item.outcome}</p>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <section className="border-y border-cortex-border bg-cortex-surface/40 py-20">
                    <div className="mx-auto max-w-7xl px-6">
                        <div className="max-w-3xl">
                            <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">Three truths</p>
                            <h2 className="mt-4 text-4xl font-semibold tracking-tight text-cortex-text-main">
                                Mycelis is built around structure, control, and continuous operation.
                            </h2>
                            <p className="mt-4 text-lg leading-8 text-cortex-text-muted">
                                The product starts from an AI Organization design, keeps tuning safe and guided, and makes ongoing
                                activity visible without dragging operators into technical complexity.
                            </p>
                        </div>

                        <div className="mt-12 grid gap-6 lg:grid-cols-3">
                            {truthCards.map((card) => (
                                <section
                                    key={card.title}
                                    id={card.title === "Structure" ? "structure" : card.title === "Control" ? "control" : "activity"}
                                    className="rounded-[1.75rem] border border-cortex-border bg-cortex-bg p-7"
                                >
                                    <div className="inline-flex rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 p-3 text-cortex-primary">
                                        {card.icon}
                                    </div>
                                    <p className="mt-5 text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">{card.eyebrow}</p>
                                    <h3 className="mt-2 text-2xl font-semibold text-cortex-text-main">{card.title}</h3>
                                    <p className="mt-4 text-sm leading-7 text-cortex-text-muted">{card.description}</p>
                                    <ul className="mt-6 space-y-3">
                                        {card.bullets.map((bullet) => (
                                            <li key={bullet} className="flex gap-3 text-sm leading-6 text-cortex-text-main">
                                                <span className="mt-1.5 h-2 w-2 rounded-full bg-cortex-primary" />
                                                <span>{bullet}</span>
                                            </li>
                                        ))}
                                    </ul>
                                </section>
                            ))}
                        </div>
                    </div>
                </section>

                <section id="after-creation" className="mx-auto max-w-7xl px-6 py-20">
                    <div className="grid gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:items-start">
                        <div className="space-y-4">
                            <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">What happens after creation</p>
                            <h2 className="text-4xl font-semibold tracking-tight text-cortex-text-main">
                                The first screen after creation is a real operating workspace.
                            </h2>
                            <p className="text-lg leading-8 text-cortex-text-muted">
                                You land inside the AI Organization with Soma, supporting structure, recent activity,
                                and guided next steps already in view.
                            </p>
                            <div className="rounded-[1.75rem] border border-cortex-border bg-cortex-surface/80 p-6">
                                <p className="text-sm font-semibold text-cortex-text-main">Default operator experience</p>
                                <p className="mt-3 text-sm leading-7 text-cortex-text-muted">
                                    Create the AI Organization first, then work through Soma with visible structure,
                                    safe tuning controls, and recent checks that reinforce continuity.
                                </p>
                            </div>
                        </div>

                        <div className="grid gap-5 md:grid-cols-2">
                            {postCreationSteps.map((step) => (
                                <SurfaceCard key={step.title} icon={step.icon} title={step.title} description={step.description} />
                            ))}
                        </div>
                    </div>
                </section>

                <section className="border-t border-cortex-border bg-cortex-surface/40 py-20">
                    <div className="mx-auto max-w-5xl px-6 text-center">
                        <p className="text-sm font-medium uppercase tracking-[0.2em] text-cortex-primary">Start here</p>
                        <h2 className="mt-4 text-4xl font-semibold tracking-tight text-cortex-text-main">
                            Create an AI Organization with clear structure, guided control, and visible activity.
                        </h2>
                        <p className="mx-auto mt-4 max-w-3xl text-lg leading-8 text-cortex-text-muted">
                            Mycelis opens with a Soma-primary workspace, not a one-off assistant prompt. Build from a
                            template or begin empty, then shape the organization with confidence.
                        </p>
                        <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
                            <Link
                                href="/dashboard"
                                className="inline-flex items-center justify-center gap-2 rounded-full bg-cortex-primary px-6 py-3 text-base font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                            >
                                Create AI Organization
                                <ArrowRight className="h-4 w-4" />
                            </Link>
                            <Link
                                href="/dashboard"
                                className="inline-flex items-center justify-center rounded-full border border-cortex-border bg-cortex-bg px-6 py-3 text-base font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 hover:text-cortex-primary"
                            >
                                Explore Templates
                            </Link>
                        </div>
                    </div>
                </section>
            </main>
        </div>
    );
}

function PreviewCard({ title, description }: { title: string; description: string }) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-cortex-text-muted">{title}</p>
            <p className="mt-2 text-sm font-semibold text-cortex-text-main">{description}</p>
        </div>
    );
}

function SurfaceCard({ icon, title, description }: { icon: ReactNode; title: string; description: string }) {
    return (
        <div className="rounded-[1.75rem] border border-cortex-border bg-cortex-surface p-6">
            <div className="inline-flex rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 p-3 text-cortex-primary">
                {icon}
            </div>
            <h3 className="mt-5 text-xl font-semibold text-cortex-text-main">{title}</h3>
            <p className="mt-3 text-sm leading-7 text-cortex-text-muted">{description}</p>
        </div>
    );
}
