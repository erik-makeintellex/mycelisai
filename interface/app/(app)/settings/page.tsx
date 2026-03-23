"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { User, Settings, Shield, Wrench, Brain, Layers } from "lucide-react";
import BrainsPage from "@/components/settings/BrainsPage";
import UsersPage from "@/components/settings/UsersPage";
import MissionProfilesPage from "@/components/settings/MissionProfilesPage";
import { useCortexStore } from "@/store/useCortexStore";

const MCPToolRegistry = dynamic(() => import("@/components/settings/MCPToolRegistry"), {
    ssr: false,
    loading: () => (
        <div className="h-64 flex items-center justify-center">
            <span className="text-cortex-text-muted text-xs font-mono animate-pulse">Loading tool registry...</span>
        </div>
    ),
});

export default function SettingsPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <SettingsContent />
        </Suspense>
    );
}

function SettingsContent() {
    const searchParams = useSearchParams();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    type TabId = "profile" | "profiles" | "users" | "engines" | "tools";
    const validTabs: TabId[] = ["profile", "profiles", "users", "engines", "tools"];
    const requestedTab = searchParams.get("tab") as TabId | null;
    const initialTab =
        requestedTab && validTabs.includes(requestedTab) && (advancedMode || (requestedTab !== "engines" && requestedTab !== "tools"))
            ? requestedTab
            : "profile";
    const [activeTab, setActiveTab] = useState<TabId>(initialTab);

    return (
        <div className="h-full flex flex-col">
            {/* Header */}
            <div className="max-w-5xl mx-auto w-full px-6 pt-6 pb-0">
                <h1 className="text-2xl font-bold tracking-tight text-cortex-text-main flex items-center gap-2">
                    <Settings className="w-6 h-6 text-cortex-text-muted" />
                    Settings
                </h1>
                <p className="text-cortex-text-muted mt-1">Manage workspace preferences, people access, and optional advanced setup.</p>
            </div>

            {/* Tabs */}
            <div className="max-w-5xl mx-auto w-full px-6 mt-6">
                <div className="flex items-center gap-1 border-b border-cortex-border">
                    <Tab label="Profile" icon={User} active={activeTab === "profile"} onClick={() => setActiveTab("profile")} />
                    <Tab label="Mission Profiles" icon={Layers} active={activeTab === "profiles"} onClick={() => setActiveTab("profiles")} />
                    <Tab label="People & Access" icon={Shield} active={activeTab === "users"} onClick={() => setActiveTab("users")} />
                    {advancedMode && (
                        <>
                            <Tab label="AI Engines" icon={Brain} active={activeTab === "engines"} onClick={() => setActiveTab("engines")} />
                            <Tab label="Connected Tools" icon={Wrench} active={activeTab === "tools"} onClick={() => setActiveTab("tools")} />
                        </>
                    )}
                </div>
            </div>

            {/* Content */}
            {activeTab === "tools" ? (
                <div className="flex-1 overflow-hidden mt-2">
                    <MCPToolRegistry />
                </div>
            ) : (
                <div className="max-w-5xl mx-auto w-full px-6 py-6 min-h-[400px]">
                    {activeTab === "profile" && <ProfileSettings />}
                    {activeTab === "engines" && advancedMode && <BrainsPage />}
                    {activeTab === "profiles" && <MissionProfilesPage />}
                    {activeTab === "users" && <UsersPage />}
                </div>
            )}
        </div>
    );
}

function Tab({ label, icon: Icon, active, onClick }: any) {
    return (
        <button
            onClick={onClick}
            className={`
                flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors
                ${active
                    ? "border-cortex-primary text-cortex-primary"
                    : "border-transparent text-cortex-text-muted hover:text-cortex-text-main hover:border-cortex-border"}
            `}
        >
            <Icon className="w-4 h-4" />
            {label}
        </button>
    );
}

function ProfileSettings() {
    const assistantName = useCortexStore((s) => s.assistantName);
    const updateAssistantName = useCortexStore((s) => s.updateAssistantName);
    const [nameDraft, setNameDraft] = useState(assistantName);
    const [isSavingName, setIsSavingName] = useState(false);
    const [saveMessage, setSaveMessage] = useState<string | null>(null);

    useEffect(() => {
        setNameDraft(assistantName);
    }, [assistantName]);

    const onSaveAssistantName = async () => {
        setIsSavingName(true);
        setSaveMessage(null);
        const ok = await updateAssistantName(nameDraft);
        setIsSavingName(false);
        setSaveMessage(ok ? "Saved" : "Failed to save");
    };

    return (
        <div className="space-y-6 max-w-lg">
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Identity</h3>
                <div className="space-y-2">
                    <label className="text-cortex-text-main text-sm block" htmlFor="assistant-name">
                        Assistant Name
                    </label>
                    <div className="flex items-center gap-2">
                        <input
                            id="assistant-name"
                            value={nameDraft}
                            onChange={(e) => setNameDraft(e.target.value)}
                            placeholder="Soma"
                            maxLength={48}
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                        <button
                            type="button"
                            onClick={onSaveAssistantName}
                            disabled={isSavingName || !nameDraft.trim()}
                            className="px-3 py-1 rounded border border-cortex-primary/40 text-cortex-primary text-sm font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                        >
                            {isSavingName ? "Saving..." : "Save"}
                        </button>
                    </div>
                    <p className="text-xs text-cortex-text-muted">
                        This updates how your orchestrator name is shown across chat, status, and workflow surfaces.
                    </p>
                    {saveMessage && (
                        <p className="text-xs font-mono text-cortex-text-muted">{saveMessage}</p>
                    )}
                </div>
            </div>
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Appearance</h3>
                <div className="flex items-center justify-between">
                    <label className="text-cortex-text-main text-sm" htmlFor="theme-select">Theme</label>
                    <select
                        id="theme-select"
                        className="bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    >
                        <option>Midnight Cortex</option>
                        <option>Cyber Dark</option>
                        <option>System</option>
                    </select>
                </div>
            </div>
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Notifications</h3>
                <div className="flex items-center justify-between">
                    <span className="text-cortex-text-main text-sm">Task Completion</span>
                    <div className="w-8 h-4 bg-cortex-success/20 rounded-full relative cursor-pointer border border-cortex-success/30">
                        <div className="absolute right-0 w-4 h-4 bg-cortex-success rounded-full shadow-sm"></div>
                    </div>
                </div>
            </div>
        </div>
    );
}
