"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { ScrollText, Cable, ShieldCheck, Clock, CalendarClock } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import AutomationHub from "@/components/automations/AutomationHub";
import { useBrowserSearch, useClientReady } from "@/lib/browserLocation";

const ApprovalsContent = dynamic(() => import("@/components/automations/ApprovalsTab"), {
    ssr: false,
    loading: () => <TabLoading label="approvals" />,
});

const WiringContent = dynamic(() => import("@/components/workspace/Workspace"), {
    ssr: false,
    loading: () => <TabLoading label="neural wiring" />,
});

const TriggersContent = dynamic(() => import("@/components/automations/TriggerRulesTab"), {
    ssr: false,
    loading: () => <TabLoading label="trigger rules" />,
});

const SchedulesContent = dynamic(() => import("@/components/automations/ScheduleRulesTab"), {
    ssr: false,
    loading: () => <TabLoading label="schedule rules" />,
});

type TabId = "active" | "triggers" | "schedules" | "approvals" | "wiring";
const VALID_TABS: TabId[] = ["active", "triggers", "schedules", "approvals", "wiring"];

export default function AutomationsPage() {
    return <AutomationsContent />;
}

function AutomationsContent() {
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const isHydrated = useClientReady();
    const search = useBrowserSearch();
    const queryTab = new URLSearchParams(search).get("tab") as TabId | null;
    const requestedTab = queryTab && VALID_TABS.includes(queryTab) ? queryTab : null;
    const effectiveAdvancedMode = isHydrated ? advancedMode : false;
    const [selectedTab, setSelectedTab] = useState<TabId | null>(null);
    const activeTab = selectedTab ?? requestedTab ?? "active";

    const effectiveTab =
        !effectiveAdvancedMode && activeTab === "wiring"
            ? "active"
            : activeTab;

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Automations
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Review active automations, configure event rules, and handle approvals around Soma/team workstreams.
                        </p>
                    </div>
                </div>

                <div className="flex gap-1 border-b border-cortex-border overflow-x-auto">
                    <TabButton active={effectiveTab === "active"} onClick={() => setSelectedTab("active")} icon={<Clock size={14} />} label="Active Automations" />
                    <TabButton active={effectiveTab === "triggers"} onClick={() => setSelectedTab("triggers")} icon={<ScrollText size={14} />} label="Trigger Rules" />
                    <TabButton active={effectiveTab === "schedules"} onClick={() => setSelectedTab("schedules")} icon={<CalendarClock size={14} />} label="Schedule Rules" />
                    <TabButton active={effectiveTab === "approvals"} onClick={() => setSelectedTab("approvals")} icon={<ShieldCheck size={14} />} label="Approvals" />
                    {effectiveAdvancedMode && (
                        <TabButton active={effectiveTab === "wiring"} onClick={() => setSelectedTab("wiring")} icon={<Cable size={14} />} label="Workflow Builder" />
                    )}
                </div>
            </header>

            <div className="flex-1 overflow-hidden">
                {effectiveTab === "active" && (
                    <AutomationHub
                        advancedMode={effectiveAdvancedMode}
                        openTab={setSelectedTab}
                    />
                )}
                {effectiveTab === "triggers" && <TriggersContent />}
                {effectiveTab === "schedules" && <SchedulesContent />}
                {effectiveTab === "approvals" && <ApprovalsContent />}
                {effectiveTab === "wiring" && <WiringContent />}
            </div>
        </div>
    );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
    return (
        <button
            type="button"
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
