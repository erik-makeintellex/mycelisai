"use client";

import { useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { ScrollText, Cable, Users, ShieldCheck, Clock } from "lucide-react";
import DegradedState from "@/components/shared/DegradedState";
import { useCortexStore } from "@/store/useCortexStore";
import AutomationHub from "@/components/automations/AutomationHub";

const ApprovalsContent = dynamic(() => import("@/components/automations/ApprovalsTab"), {
    ssr: false,
    loading: () => <TabLoading label="approvals" />,
});

const WiringContent = dynamic(() => import("@/components/workspace/Workspace"), {
    ssr: false,
    loading: () => <TabLoading label="neural wiring" />,
});

const TeamsContent = dynamic(() => import("@/components/teams/TeamsPage"), {
    ssr: false,
    loading: () => <TabLoading label="teams" />,
});

const TriggersContent = dynamic(() => import("@/components/automations/TriggerRulesTab"), {
    ssr: false,
    loading: () => <TabLoading label="trigger rules" />,
});

type TabId = "active" | "drafts" | "triggers" | "approvals" | "wiring" | "teams";
const VALID_TABS: TabId[] = ["active", "drafts", "triggers", "approvals", "wiring", "teams"];

export default function AutomationsPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <AutomationsContent />
        </Suspense>
    );
}

function AutomationsContent() {
    const searchParams = useSearchParams();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const tabParam = searchParams.get("tab") as TabId | null;
    const initialTab = tabParam && VALID_TABS.includes(tabParam) ? tabParam : "active";
    const [activeTab, setActiveTab] = useState<TabId>(
        initialTab === "wiring" && !advancedMode ? "active" : initialTab
    );

    const effectiveTab = activeTab === "wiring" && !advancedMode ? "active" : activeTab;

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Automations
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Scheduled missions, trigger rules, drafts, and approvals
                        </p>
                    </div>
                </div>

                <div className="flex gap-1 border-b border-cortex-border overflow-x-auto">
                    <TabButton active={effectiveTab === "active"} onClick={() => setActiveTab("active")} icon={<Clock size={14} />} label="Active Automations" />
                    <TabButton active={effectiveTab === "drafts"} onClick={() => setActiveTab("drafts")} icon={<Cable size={14} />} label="Draft Blueprints" />
                    <TabButton active={effectiveTab === "triggers"} onClick={() => setActiveTab("triggers")} icon={<ScrollText size={14} />} label="Trigger Rules" />
                    <TabButton active={effectiveTab === "approvals"} onClick={() => setActiveTab("approvals")} icon={<ShieldCheck size={14} />} label="Approvals" />
                    <TabButton active={effectiveTab === "teams"} onClick={() => setActiveTab("teams")} icon={<Users size={14} />} label="Teams" />
                    {advancedMode && (
                        <TabButton active={effectiveTab === "wiring"} onClick={() => setActiveTab("wiring")} icon={<Cable size={14} />} label="Neural Wiring" />
                    )}
                </div>
            </header>

            <div className="flex-1 overflow-hidden">
                {effectiveTab === "active" && (
                    <AutomationHub
                        advancedMode={advancedMode}
                        openTab={(tab) => setActiveTab(tab)}
                    />
                )}
                {effectiveTab === "drafts" && (
                    <div className="h-full p-6">
                        <DegradedState
                            title="Draft Blueprints"
                            reason="Blueprint drafts are managed through Neural Wiring."
                            available={["Open Teams to inspect available execution targets", "Use Trigger Rules to chain existing workflows", "Negotiate intent from Workspace to generate new drafts"]}
                            action="Next step: create a trigger rule, route to a team, then return here for scheduler rollout."
                        />
                    </div>
                )}
                {effectiveTab === "triggers" && <TriggersContent />}
                {effectiveTab === "approvals" && <ApprovalsContent />}
                {effectiveTab === "teams" && <TeamsContent />}
                {effectiveTab === "wiring" && <WiringContent />}
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
