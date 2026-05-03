"use client";

import { useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { Brain, Wrench, BookOpen, FolderOpen, GitBranch, BookMarked } from "lucide-react";
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
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Advanced Resources
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Support systems for Soma: connected tools, MCP/search readiness, workspace files, exchange lanes, and AI engine setup
                        </p>
                    </div>
                </div>

                <div className="flex gap-1 border-b border-cortex-border">
                    <TabButton active={activeTab === "tools"} onClick={() => setActiveTab("tools")} icon={<Wrench size={14} />} label="Connected Tools" />
                    <TabButton active={activeTab === "exchange"} onClick={() => setActiveTab("exchange")} icon={<GitBranch size={14} />} label="Exchange" />
                    <TabButton active={activeTab === "deployment-context"} onClick={() => setActiveTab("deployment-context")} icon={<BookMarked size={14} />} label="Deployment Context" />
                    <TabButton active={activeTab === "workspace"} onClick={() => setActiveTab("workspace")} icon={<FolderOpen size={14} />} label="Workspace Files" />
                    <TabButton active={activeTab === "engines"} onClick={() => setActiveTab("engines")} icon={<Brain size={14} />} label="AI Engines" />
                    <TabButton active={activeTab === "roles"} onClick={() => setActiveTab("roles")} icon={<BookOpen size={14} />} label="Role Library" />
                </div>
            </header>

            <div className="flex-1 overflow-hidden">
                {activeTab === "engines" && (
                    <div className="h-full overflow-y-auto">
                        <div className="max-w-5xl mx-auto w-full px-6 py-6">
                            <BrainsPage />
                        </div>
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
        </div>
    );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
    return (
        <button
            onClick={onClick}
            className={`px-4 py-2.5 text-xs font-medium flex items-center gap-2 border-b-2 transition-colors -mb-px whitespace-nowrap ${
                active
                    ? "border-cortex-primary text-cortex-primary"
                    : "border-transparent text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            {icon}
            {label}
        </button>
    );
}

function TabLoading({ label }: { label: string }) {
    return (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono animate-pulse">Loading {label}...</span>
        </div>
    );
}
