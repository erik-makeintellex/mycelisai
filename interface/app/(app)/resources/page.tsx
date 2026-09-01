"use client";

import { Suspense } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { Brain, BookOpen, FolderOpen, GitBranch, BookMarked, ShieldCheck, type LucideIcon } from "lucide-react";
import BrainsPage from "@/components/settings/BrainsPage";

const MCPToolRegistry = dynamic(() => import("@/components/settings/MCPToolRegistry"), {
    ssr: false,
    loading: () => <TabLoading label="tools" />,
});

const CataloguePage = dynamic(() => import("@/components/catalogue/CataloguePage"), {
    ssr: false,
    loading: () => <TabLoading label="catalogue" />,
});

const WorkspaceExplorer = dynamic(() => import("@/components/resources/WorkspaceExplorer"), {
    ssr: false,
    loading: () => <TabLoading label="workspace" />,
});

const ExchangeInspector = dynamic(() => import("@/components/resources/ExchangeInspector"), {
    ssr: false,
    loading: () => <TabLoading label="exchange" />,
});

const DeploymentContextPanel = dynamic(() => import("@/components/resources/DeploymentContextPanel"), {
    ssr: false,
    loading: () => <TabLoading label="deployment context" />,
});

type TabId = "engines" | "tools" | "workspace" | "roles" | "exchange" | "deployment-context";
const VALID_TABS: TabId[] = ["engines", "tools", "workspace", "roles", "exchange", "deployment-context"];
type ResourceTab = {
    id: TabId;
    label: string;
    summary: string;
    detail: string;
    icon: LucideIcon;
};

type ResourceGroup = {
    label: string;
    tabs: ResourceTab[];
};

const RESOURCE_TABS: ResourceTab[] = [
    {
        id: "workspace",
        label: "Deliverables",
        summary: "Open the files, packages, media, and other results Soma delivered for you.",
        detail: "Your retained results",
        icon: FolderOpen,
    },
    {
        id: "tools",
        label: "Capabilities",
        summary: "What Soma can use, what needs repair, and what can be requested.",
        detail: "What Soma can use",
        icon: ShieldCheck,
    },
    {
        id: "exchange",
        label: "Exchange",
        summary: "Review normalized handoffs and evidence moving between teams and tools.",
        detail: "Team handoffs",
        icon: GitBranch,
    },
    {
        id: "deployment-context",
        label: "Deployment Context",
        summary: "Save files or notes Soma should reuse as long-lived, scoped source context.",
        detail: "Long-term context",
        icon: BookMarked,
    },
    {
        id: "engines",
        label: "AI Engines",
        summary: "Provider configuration, model routing, and health for advanced operators.",
        detail: "Model providers",
        icon: Brain,
    },
    {
        id: "roles",
        label: "Worker Profiles",
        summary: "Inspect ready-made teammates or continue with Soma to create an activated scoped profile.",
        detail: "Reusable teammates",
        icon: BookOpen,
    },
];

const RESOURCE_GROUPS: ResourceGroup[] = [
    { label: "Results", tabs: RESOURCE_TABS.slice(0, 1) },
    { label: "Advanced resources", tabs: RESOURCE_TABS.slice(1) },
];

export default function ResourcesPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <ResourcesContent />
        </Suspense>
    );
}

function ResourcesContent() {
    const router = useRouter();
    const pathname = usePathname() ?? "/resources";
    const searchParams = useSearchParams();
    const tabParam = (searchParams?.get("tab") as TabId | null) ?? null;
    const pathParam = searchParams?.get("path") ?? null;
    const activeTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : "workspace";

    const tabHref = (tab: TabId) => {
        const nextParams = new URLSearchParams(searchParams?.toString() ?? "");
        if (tab === "workspace") {
            nextParams.delete("tab");
        } else {
            nextParams.set("tab", tab);
        }
        if (tab !== "workspace") {
            nextParams.delete("path");
        }
        const query = nextParams.toString();
        return query ? `${pathname}?${query}` : pathname;
    };
    const selectTab = (tab: TabId) => router.push(tabHref(tab), { scroll: false });

    return (
        <div className="flex h-full min-h-0 flex-col overflow-hidden bg-cortex-bg">
            <header className="flex-shrink-0 border-b border-cortex-border px-4 py-3 sm:px-6 sm:py-4">
                <div className="flex min-w-0 flex-col gap-1">
                    <div className="min-w-0">
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Resources
                        </h1>
                        <p className="mt-1 max-w-3xl text-sm leading-6 text-cortex-text-muted">
                            Open delivered work, connect Soma to sources, and save context for future use.
                        </p>
                    </div>
                </div>
            </header>

            <div className="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)] gap-3 p-3 lg:grid-cols-[16rem_minmax(0,1fr)] lg:grid-rows-1 lg:gap-4 lg:p-4">
                <nav
                    className="min-w-0 overflow-x-auto overflow-y-hidden rounded-lg border border-cortex-border bg-cortex-surface/70 p-1.5 lg:min-h-0 lg:overflow-x-hidden lg:overflow-y-auto lg:p-2"
                    aria-label="Resource type menu"
                    role="tablist"
                    aria-orientation="horizontal"
                    data-testid="resource-type-tabs"
                >
                    <div className="flex w-max min-w-full gap-3 lg:block lg:w-auto lg:space-y-3">
                        {RESOURCE_GROUPS.map((group) => (
                            <div key={group.label} role="presentation" className="min-w-0">
                                <div className="px-2 pb-1 pt-1 text-[10px] font-semibold uppercase text-cortex-text-muted">
                                    {group.label}
                                </div>
                                <div className="flex gap-1 lg:block lg:space-y-1">
                                    {group.tabs.map((tab) => (
                                        <ResourceMenuButton
                                            key={tab.id}
                                            tab={tab}
                                            active={activeTab === tab.id}
                                            href={tabHref(tab.id)}
                                        />
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                </nav>

                <section
                    id={`resource-panel-${activeTab}`}
                    role="tabpanel"
                    aria-labelledby={`resource-tab-${activeTab}`}
                    className="flex min-h-0 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface/60"
                >
                    <ResourcePanelHeader tab={RESOURCE_TABS.find((tab) => tab.id === activeTab) ?? RESOURCE_TABS[0]} />
                    <div className="min-h-0 flex-1 overflow-y-auto">
                        {activeTab === "engines" && (
                            <div className="mx-auto w-full max-w-5xl px-6 py-6">
                                <BrainsPage />
                            </div>
                        )}
                        {activeTab === "tools" && <MCPToolRegistry />}
                        {activeTab === "exchange" && <ExchangeInspector />}
                        {activeTab === "deployment-context" && <DeploymentContextPanel />}
                        {activeTab === "workspace" && (
                            <WorkspaceExplorer initialPath={pathParam} onOpenToolsTab={() => selectTab("tools")} />
                        )}
                        {activeTab === "roles" && <CataloguePage />}
                    </div>
                </section>
            </div>
        </div>
    );
}

function ResourceMenuButton({ active, href, tab }: { active: boolean; href: string; tab: ResourceTab }) {
    const Icon = tab.icon;

    return (
        <Link
            id={`resource-tab-${tab.id}`}
            href={href}
            scroll={false}
            role="tab"
            aria-selected={active}
            aria-controls={`resource-panel-${tab.id}`}
            aria-current={active ? "page" : undefined}
            className={`flex h-11 min-w-max items-center gap-2 rounded border px-3 text-left transition-colors lg:h-auto lg:w-full lg:min-w-0 lg:items-start lg:gap-3 lg:py-2 ${
                active
                    ? "border-cortex-primary/50 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-transparent text-cortex-text-muted hover:border-cortex-border hover:bg-cortex-bg"
            }`}
        >
            <Icon className="mt-0.5 h-4 w-4 flex-shrink-0 text-cortex-primary" aria-hidden="true" />
            <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold">{tab.label}</span>
                <span className="mt-0.5 hidden text-[11px] leading-4 lg:block">{tab.detail}</span>
            </span>
        </Link>
    );
}

function ResourcePanelHeader({ tab }: { tab: ResourceTab }) {
    const Icon = tab.icon;

    return (
        <div className="flex flex-shrink-0 flex-col gap-2 border-b border-cortex-border bg-cortex-bg px-4 py-3 md:flex-row md:items-center md:justify-between">
            <div className="flex min-w-0 items-start gap-3">
                <div className="rounded border border-cortex-border bg-cortex-surface p-2 text-cortex-primary">
                    <Icon className="h-4 w-4" aria-hidden="true" />
                </div>
                <div className="min-w-0">
                    <h2 className="text-sm font-semibold text-cortex-text-main">{tab.label}</h2>
                    <p className="mt-1 max-w-3xl text-xs leading-5 text-cortex-text-muted">{tab.summary}</p>
                </div>
            </div>
        </div>
    );
}

function TabLoading({ label }: { label: string }) {
    return (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono animate-pulse">Loading {label}...</span>
        </div>
    );
}
