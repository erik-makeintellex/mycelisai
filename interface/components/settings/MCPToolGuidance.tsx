"use client";

import React from "react";
import { Activity, AlertTriangle, BookOpen, CheckCircle2, Globe, Radio, Search, Users, Wrench } from "lucide-react";
import type { SearchCapabilityStatus } from "@/store/useCortexStore";
import { SearchSourceList } from "./SearchSourceList";

const somaToolPrompts = [
    {
        label: "Search web",
        prompt: 'Search the web for "latest AI agent platform releases", summarize the strongest sources, and cite them.',
        icon: Globe,
    },
    {
        label: "Create team",
        prompt: "Propose a temporary release-review team, keep it small, list target tools and outputs, and ask me to approve before launch.",
        icon: Users,
    },
    {
        label: "Ask teams",
        prompt: "Ask the active teams for blockers, compare their responses, and recommend the next action.",
        icon: Users,
    },
    {
        label: "Read host data",
        prompt: "Review host/private data boundaries, name the files or source scope, and ask me to confirm before using or retaining outputs.",
        icon: Search,
    },
    {
        label: "Review MCP",
        prompt: "Review current MCP servers, tools, and recent use, then tell me which agents should have which tools.",
        icon: Wrench,
    },
] as const;

export function SomaToolPromptCard() {
    return (
        <details className="rounded-xl border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-3">
            <summary className="cursor-pointer text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                Concrete Soma Commands
            </summary>
            <p className="mt-2 text-xs leading-5 text-cortex-text-main">
                Optional examples for talking to Soma. Private data, private services, tool assignment, and team launch should ask for confirmation or approval first.
            </p>
            <div className="mt-3 grid gap-2 md:grid-cols-5">
                {somaToolPrompts.map((item) => {
                    const Icon = item.icon;
                    return (
                        <button
                            key={item.label}
                            type="button"
                            onClick={() => void navigator.clipboard?.writeText(item.prompt)}
                            title="Copy Soma command"
                            className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-left transition hover:border-cortex-primary/30"
                        >
                            <span className="flex items-center gap-1.5 text-[11px] font-semibold text-cortex-text-main">
                                <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                                {item.label}
                            </span>
                            <span className="mt-1 line-clamp-3 block text-[10px] leading-4 text-cortex-text-muted">
                                {item.prompt}
                            </span>
                        </button>
                    );
                })}
            </div>
        </details>
    );
}

export function ConnectedToolsWorkflowCard({ isStreamConnected }: { isStreamConnected: boolean }) {
    return (
        <details className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <summary className="flex cursor-pointer items-center gap-2">
                <Activity className="w-4 h-4 text-cortex-primary" />
                <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                    Capability Workflow
                </span>
            </summary>
            <div className="mt-3 grid gap-3 md:grid-cols-3">
                <WorkflowStep title="1. Add">
                    Install a curated MCP server from Library instead of wiring raw config by hand.
                </WorkflowStep>
                <WorkflowStep title="2. Review/edit">
                    Expand an installed server to inspect command, args, env/header references, discovered tools, and recent use.
                </WorkflowStep>
                <WorkflowStep title="3. Use">
                    Ask Soma directly to search, read host data, create teams, or talk with teams, then watch MCP activity here.
                </WorkflowStep>
            </div>
            <div className="mt-3 flex items-center gap-2 text-[11px] text-cortex-text-muted">
                <Radio className={`w-3.5 h-3.5 ${isStreamConnected ? "text-cortex-success" : "text-cortex-warning"}`} />
                {isStreamConnected
                    ? "Live activity stream connected."
                    : "Live activity stream is reconnecting. Recent MCP use will appear once the stream is online."}
            </div>
        </details>
    );
}

export function WebAccessSetupCard({
    status,
    isLoading,
    error,
    onAddWebCapability,
}: {
    status: SearchCapabilityStatus | null;
    isLoading: boolean;
    error: string | null;
    onAddWebCapability: () => void;
}) {
    const hasLocalSearch = Boolean(status?.enabled && status?.configured && status?.supports_local_sources);
    const hasPublicWeb = Boolean(status?.enabled && status?.configured && status?.supports_public_web);
    const canAskSoma = Boolean(status?.direct_soma_interaction);
    const heading = error
        ? "Web access could not be checked"
        : isLoading && !status
        ? "Checking web access"
        : hasPublicWeb
        ? "Public web access is available"
        : hasLocalSearch
        ? "Local-source search is available"
        : "Web access needs setup";
    const explanation = error
        ? error
        : hasPublicWeb
        ? "Soma can search the public web now. Add fetch only when Soma or a team needs to open a specific URL you provide."
        : hasLocalSearch
        ? "Soma can search retained Mycelis sources. Add a public web search provider when you need fresh internet research, or fetch when you need explicit URL reading."
        : status?.blocker?.message ?? status?.next_actions?.[0] ?? "Add or repair a web capability before asking Soma for fresh web research.";
    const nextStep = hasPublicWeb
        ? "Ask Soma naturally for research. Use the button only when you want explicit URL reading or another optional web tool."
        : hasLocalSearch
        ? "Add or configure a web search provider for public internet research. Keep local-source search for retained Mycelis context and approved mounts."
        : "Set up web search for public research, or add local/mounted sources when Soma should search approved private data.";
    const actionLabel = hasPublicWeb
        ? "Add URL reader"
        : hasLocalSearch
        ? "Add web search provider"
        : "Set up web search";

    return (
        <section
            id="web-access"
            aria-label="Web access setup"
            className="rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 px-4 py-4 scroll-mt-20"
        >
            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                <div className="flex items-start gap-3">
                    <div className={`mt-0.5 rounded-lg border p-2 ${hasPublicWeb ? "border-cortex-success/30 bg-cortex-success/10" : "border-cortex-warning/30 bg-cortex-warning/10"}`}>
                        <Globe className={`h-4 w-4 ${hasPublicWeb ? "text-cortex-success" : "text-cortex-warning"}`} />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                            Web access setup
                        </p>
                        <h3 className="mt-1 text-sm font-semibold text-cortex-text-main">{heading}</h3>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{explanation}</p>
                    </div>
                </div>
                <button
                    type="button"
                    onClick={onAddWebCapability}
                    className="inline-flex shrink-0 items-center justify-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/10"
                >
                    <BookOpen className="h-3.5 w-3.5" />
                    {actionLabel}
                </button>
            </div>
            <div className="mt-4 grid gap-2 md:grid-cols-3">
                <CapabilityPill active={canAskSoma} label="Ask Soma: research or search" />
                <CapabilityPill active={hasLocalSearch} label="Approved local data" />
                <CapabilityPill active={hasPublicWeb} label="Public web provider" />
            </div>
            <SearchSourceList sources={status?.sources ?? []} compact />
            <p className="mt-3 text-xs leading-5 text-cortex-text-main">
                {nextStep}
            </p>
        </section>
    );
}

export function SearchCapabilityCard({
    status,
    isLoading,
    error,
}: {
    status: SearchCapabilityStatus | null;
    isLoading: boolean;
    error: string | null;
}) {
    const provider = status?.provider ?? "unknown";
    const ready = Boolean(status?.enabled && status?.configured);
    const headline = error
        ? "Search capability status unavailable"
        : isLoading && !status
        ? "Checking search capability"
        : ready
        ? "Soma search is ready"
        : "Soma search needs configuration";
    const detail = error
        ? error
        : status?.blocker?.message
        ?? status?.next_actions?.[0]
        ?? "Soma can route governed search through the configured Mycelis Search provider.";
    const tokenText = status?.requires_hosted_api_token
        ? "Brave MCP requires BRAVE_API_KEY."
        : "Brave tokens are not required for built-in web, approved local data, mounted sources, local API, or self-hosted SearXNG.";

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex items-start gap-3">
                    <div className={`mt-0.5 rounded-lg border p-2 ${ready ? "border-cortex-success/30 bg-cortex-success/10" : "border-cortex-warning/30 bg-cortex-warning/10"}`}>
                        <Search className={`h-4 w-4 ${ready ? "text-cortex-success" : "text-cortex-warning"}`} />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Mycelis Search Capability
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">{headline}</p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{detail}</p>
                    </div>
                </div>
                <span className="self-start rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono uppercase text-cortex-text-muted">
                    {searchProviderLabel(provider)}
                </span>
            </div>
            <div className="mt-4 grid gap-2 md:grid-cols-3">
                <CapabilityPill active={Boolean(status?.direct_soma_interaction)} label="Ask Soma to search" />
                <CapabilityPill active={Boolean(status?.supports_local_sources)} label="Approved local data" />
                <CapabilityPill active={Boolean(status?.supports_public_web)} label="Public web" />
            </div>
            <div className="mt-3 flex items-center gap-2 text-[11px] text-cortex-text-muted">
                {status?.requires_hosted_api_token ? (
                    <AlertTriangle className="h-3.5 w-3.5 text-cortex-warning" />
                ) : (
                    <CheckCircle2 className="h-3.5 w-3.5 text-cortex-success" />
                )}
                {tokenText}
            </div>
            <details className="mt-3 rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-2">
                <summary className="cursor-pointer text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                    Inspect provider refs
                </summary>
                <div className="mt-2 grid gap-2 sm:grid-cols-2">
                    <SourceDetail label="Provider ref" value={provider} />
                    <SourceDetail label="Tool ref" value={status?.soma_tool_name ?? "web_search"} />
                    <SourceDetail label="Result limit" value={`${status?.max_results ?? 0}`} />
                </div>
            </details>
        </div>
    );
}

function CapabilityPill({ active, label }: { active: boolean; label: string }) {
    return (
        <div className={`rounded-lg border px-3 py-2 text-[11px] font-mono ${active ? "border-cortex-success/25 bg-cortex-success/10 text-cortex-text-main" : "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted"}`}>
            {label}
        </div>
    );
}

function SourceDetail({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2">
            <p className="text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">{label}</p>
            <p className="mt-1 break-words text-[11px] leading-4 text-cortex-text-main">{value}</p>
        </div>
    );
}

function searchProviderLabel(provider: string): string {
    if (provider === "builtin_web") return "Built-in web";
    if (provider === "local_sources") return "Local sources";
    if (provider === "local_api") return "Local API";
    if (provider === "searxng") return "Self-hosted web";
    if (provider === "brave") return "Brave MCP";
    if (provider === "disabled") return "Disabled";
    return provider;
}

function WorkflowStep({ title, children }: { title: string; children: React.ReactNode }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">{title}</p>
            <p className="mt-2 text-xs leading-5 text-cortex-text-main">{children}</p>
        </div>
    );
}
