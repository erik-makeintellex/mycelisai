"use client";

import { useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { Brain, Wrench, BookOpen, FolderOpen, GitBranch, BookMarked, type LucideIcon } from "lucide-react";
import BrainsPage from "@/components/settings/BrainsPage";
import AdvancedModeGate from "@/components/shared/AdvancedModeGate";
import { useCortexStore } from "@/store/useCortexStore";

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

const RESOURCE_TABS: ResourceTab[] = [
    {
        id: "tools",
        label: "Connected Tools",
        summary: "MCP servers, capability readiness, search posture, and recent tool use.",
        detail: "What Soma can use",
        icon: Wrench,
    },
    {
        id: "workspace",
        label: "Output Files",
        summary: "Open generated content folders and browse retained files through the governed workspace boundary.",
        detail: "Generated content",
        icon: FolderOpen,
    },
    {
        id: "exchange",
        label: "Exchange",
        summary: "Channels, threads, and normalized outputs crossing teams and tools.",
        detail: "Managed output lanes",
        icon: GitBranch,
    },
    {
        id: "deployment-context",
        label: "Deployment Context",
        summary: "Governed intake for deployment knowledge and private context lanes.",
        detail: "Context loading",
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
        label: "Role Library",
        summary: "Reusable specialist roles and templates for advanced workflow work.",
        detail: "Reusable templates",
        icon: BookOpen,
    },
];

export default function ResourcesPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <ResourcesContent />
        </Suspense>
    );
}

function ResourcesContent() {
    const searchParams = useSearchParams();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const tabParam = (searchParams?.get("tab") as TabId | null) ?? null;
    const [activeTab, setActiveTab] = useState<TabId>(
        tabParam && VALID_TABS.includes(tabParam) ? tabParam : "tools"
    );

    if (!advancedMode) {
        return (
            <AdvancedModeGate
                title="Advanced resources stay tucked away by default"
                summary="Connected tools, workspace file access, role libraries, and global AI engine setup are available when you intentionally open Advanced mode."
            />
        );
    }

    return (
        <div className="flex h-full min-h-0 flex-col overflow-hidden bg-cortex-bg">
            <header className="flex-shrink-0 border-b border-cortex-border px-6 py-5">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                    <div className="min-w-0">
                        <p className="text-[10px] font-semibold uppercase text-cortex-text-muted">Operator support systems</p>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Advanced Resources
                        </h1>
                        <p className="mt-1 max-w-3xl text-sm leading-6 text-cortex-text-muted">
                            Select one resource type, work inside the focused panel, and keep the rest of the support
                            system out of the operator's scroll path.
                        </p>
                    </div>
                    <div className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-xs text-cortex-primary">
                        {RESOURCE_TABS.length} resource types
                    </div>
                </div>
            </header>

            <div className="grid min-h-0 flex-1 gap-4 p-4 lg:grid-cols-[19rem_minmax(0,1fr)]">
                <nav
                    className="min-h-0 overflow-y-auto rounded-lg border border-cortex-border bg-cortex-surface/70 p-2"
                    aria-label="Resource type menu"
                >
                    <div className="px-2 pb-2 pt-1 text-[10px] font-semibold uppercase text-cortex-text-muted">
                        Resource types
                    </div>
                    <div className="space-y-1">
                        {RESOURCE_TABS.map((tab) => (
                            <ResourceMenuButton
                                key={tab.id}
                                tab={tab}
                                active={activeTab === tab.id}
                                onClick={() => setActiveTab(tab.id)}
                            />
                        ))}
                    </div>
                </nav>

                <section className="flex min-h-0 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface/60">
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
                            <WorkspaceExplorer onOpenToolsTab={() => setActiveTab("tools")} />
                        )}
                        {activeTab === "roles" && <CataloguePage />}
                    </div>
                </section>
            </div>
        </div>
    );
}

function ResourceMenuButton({ active, onClick, tab }: { active: boolean; onClick: () => void; tab: ResourceTab }) {
    const Icon = tab.icon;

    return (
        <button
            type="button"
            onClick={onClick}
            aria-current={active ? "page" : undefined}
            className={`flex w-full items-start gap-3 rounded border px-3 py-2.5 text-left transition-colors ${
                active
                    ? "border-cortex-primary/50 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-transparent text-cortex-text-muted hover:border-cortex-border hover:bg-cortex-bg"
            }`}
        >
            <Icon className="mt-0.5 h-4 w-4 flex-shrink-0 text-cortex-primary" aria-hidden="true" />
            <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold">{tab.label}</span>
                <span className="mt-0.5 block text-[11px] leading-4">{tab.detail}</span>
            </span>
        </button>
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
            <span className="rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-[10px] uppercase text-cortex-text-muted">
                Inner scroll
            </span>
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
