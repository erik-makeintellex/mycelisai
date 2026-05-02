"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { User, Settings, Shield, Wrench, Brain, Layers, KeyRound, type LucideIcon } from "lucide-react";
import AuthProvidersPage from "@/components/settings/auth-providers/AuthProvidersPage";
import BrainsPage from "@/components/settings/BrainsPage";
import ProfileSettings from "@/components/settings/ProfileSettings";
import UsersPage from "@/components/settings/UsersPage";
import MissionProfilesPage from "@/components/settings/MissionProfilesPage";
import {
    SettingsGuidedWorkflow,
    type SettingsTabId as TabId,
} from "@/components/settings/SettingsGuidedWorkflow";
import { useCortexStore } from "@/store/useCortexStore";

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
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const validTabs: TabId[] = ["profile", "profiles", "users", "engines", "auth", "tools"];
    const requestedTab = (searchParams?.get("tab") as TabId | null) ?? null;
    const requestedAdvancedTab = requestedTab && ["engines", "auth", "tools"].includes(requestedTab) ? requestedTab : null;
    const initialTab =
        requestedTab && validTabs.includes(requestedTab) && (advancedMode || !["engines", "auth", "tools"].includes(requestedTab))
            ? requestedTab
            : "profile";
    const [activeTab, setActiveTab] = useState<TabId>(initialTab);
    const openRequestedAdvancedTab = () => {
        if (!requestedAdvancedTab) return;
        if (!advancedMode) {
            toggleAdvancedMode();
        }
        setActiveTab(requestedAdvancedTab);
    };

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

            <div className="max-w-5xl mx-auto w-full px-6 mt-6 space-y-5">
                {requestedAdvancedTab && !advancedMode ? (
                    <AdvancedDeepLinkNotice requestedTab={requestedAdvancedTab} onOpen={openRequestedAdvancedTab} />
                ) : null}
                <SettingsGuidedWorkflow advancedMode={advancedMode} activeTab={activeTab} onSelect={setActiveTab} />

                {/* Tabs */}
                <div role="tablist" aria-label="Settings sections" className="flex items-center gap-1 border-b border-cortex-border">
                    <Tab label="Profile" icon={User} active={activeTab === "profile"} onClick={() => setActiveTab("profile")} />
                    <Tab label="Mission Profiles" icon={Layers} active={activeTab === "profiles"} onClick={() => setActiveTab("profiles")} />
                    <Tab label="People & Access" icon={Shield} active={activeTab === "users"} onClick={() => setActiveTab("users")} />
                    {advancedMode && (
                        <>
                            <Tab label="AI Engines" icon={Brain} active={activeTab === "engines"} onClick={() => setActiveTab("engines")} />
                            <Tab label="Auth Providers" icon={KeyRound} active={activeTab === "auth"} onClick={() => setActiveTab("auth")} />
                            <Tab label="Connected Tools" icon={Wrench} active={activeTab === "tools"} onClick={() => setActiveTab("tools")} />
                        </>
                    )}
                </div>
            </div>

            <div className="max-w-5xl mx-auto w-full px-6 py-6 min-h-[400px]">
                {activeTab === "profile" && <ProfileSettings />}
                {activeTab === "engines" && advancedMode && <BrainsPage />}
                {activeTab === "auth" && advancedMode && <AuthProvidersPage />}
                {activeTab === "profiles" && <MissionProfilesPage />}
                {activeTab === "users" && <UsersPage />}
                {activeTab === "tools" && advancedMode && <ConnectedToolsRedirect />}
            </div>
        </div>
    );
}

function ConnectedToolsRedirect() {
    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
                Connected Tools
            </p>
            <h2 className="mt-2 text-xl font-semibold text-cortex-text-main">
                Manage MCP and search from Resources.
            </h2>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-cortex-text-muted">
                Resources is the operational home for MCP servers, web search readiness,
                recent tool activity, and workspace data boundaries. Settings keeps the
                admin setup path focused on preferences, people, engines, and auth.
            </p>
            <Link
                href="/resources?tab=tools"
                className="mt-5 inline-flex rounded-xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
            >
                Open Resources tools
            </Link>
        </section>
    );
}

function AdvancedDeepLinkNotice({ requestedTab, onOpen }: { requestedTab: TabId; onOpen: () => void }) {
    const labels: Record<TabId, string> = {
        profile: "Profile",
        profiles: "Mission Profiles",
        users: "People & Access",
        engines: "AI Engines",
        auth: "Auth Providers",
        tools: "Connected Tools",
    };

    return (
        <section className="rounded-2xl border border-cortex-primary/25 bg-cortex-primary/10 px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p className="text-sm font-semibold text-cortex-text-main">{labels[requestedTab]} lives in Advanced mode</p>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                        This link is valid, but the section is intentionally hidden until Advanced mode is on.
                    </p>
                </div>
                <button
                    type="button"
                    onClick={onOpen}
                    className="inline-flex items-center justify-center rounded-xl border border-cortex-primary/30 bg-cortex-bg px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                >
                    Open {labels[requestedTab]}
                </button>
            </div>
        </section>
    );
}

function Tab({
    label,
    icon: Icon,
    active,
    onClick,
}: {
    label: string;
    icon: LucideIcon;
    active: boolean;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            role="tab"
            aria-selected={active}
            aria-current={active ? "page" : undefined}
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

