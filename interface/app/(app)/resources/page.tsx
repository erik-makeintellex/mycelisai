"use client";

import { useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { Brain, Wrench, BookOpen, FolderOpen } from "lucide-react";
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

type TabId = "brains" | "tools" | "workspace" | "catalogue";
const VALID_TABS: TabId[] = ["brains", "tools", "workspace", "catalogue"];

export default function ResourcesPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <ResourcesContent />
        </Suspense>
    );
}

function ResourcesContent() {
    const searchParams = useSearchParams();
    const tabParam = searchParams.get("tab") as TabId | null;
    const [activeTab, setActiveTab] = useState<TabId>(
        tabParam && VALID_TABS.includes(tabParam) ? tabParam : "brains"
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Resources
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Brains, tools, workspace, and agent capabilities
                        </p>
                    </div>
                </div>

                <div className="flex gap-1 border-b border-cortex-border">
                    <TabButton active={activeTab === "brains"} onClick={() => setActiveTab("brains")} icon={<Brain size={14} />} label="Brains" />
                    <TabButton active={activeTab === "tools"} onClick={() => setActiveTab("tools")} icon={<Wrench size={14} />} label="MCP Tools" />
                    <TabButton active={activeTab === "workspace"} onClick={() => setActiveTab("workspace")} icon={<FolderOpen size={14} />} label="Workspace Explorer" />
                    <TabButton active={activeTab === "catalogue"} onClick={() => setActiveTab("catalogue")} icon={<BookOpen size={14} />} label="Capabilities" />
                </div>
            </header>

            <div className="flex-1 overflow-hidden">
                {activeTab === "brains" && (
                    <div className="h-full overflow-y-auto">
                        <div className="max-w-5xl mx-auto w-full px-6 py-6">
                            <BrainsPage />
                        </div>
                    </div>
                )}
                {activeTab === "tools" && <MCPToolRegistry />}
                {activeTab === "workspace" && (
                    <WorkspaceExplorer onOpenToolsTab={() => setActiveTab("tools")} />
                )}
                {activeTab === "catalogue" && <CataloguePage />}
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
