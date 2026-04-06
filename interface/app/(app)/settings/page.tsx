"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import { User, Settings, Shield, Wrench, Brain, Layers, ArrowRight, type LucideIcon } from "lucide-react";
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

type TabId = "profile" | "profiles" | "users" | "engines" | "tools";

type WorkflowCardDefinition = {
    id: TabId;
    title: string;
    summary: string;
    buttonLabel: string;
    icon: LucideIcon;
};

const DEFAULT_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
    {
        id: "profile",
        title: "Name Soma and set the workspace look",
        summary: "Start by setting your assistant identity and the product theme operators see every day.",
        buttonLabel: "Open Profile",
        icon: User,
    },
    {
        id: "profiles",
        title: "Shape reusable mission defaults",
        summary: "Keep workflow-ready mission profiles available before deeper tuning or governed execution begins.",
        buttonLabel: "Open Mission Profiles",
        icon: Layers,
    },
    {
        id: "users",
        title: "Review people and access",
        summary: "Confirm who can work in this workspace and keep visible access controls intentional.",
        buttonLabel: "Open People & Access",
        icon: Shield,
    },
];

const ADVANCED_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
    {
        id: "engines",
        title: "Inspect AI engine posture",
        summary: "Review how Soma and the wider workspace are tuned without dropping raw runtime detail into the default path.",
        buttonLabel: "Open AI Engines",
        icon: Brain,
    },
    {
        id: "tools",
        title: "Inspect connected tools",
        summary: "Open the approved tool library when you intentionally need deeper integrations or workspace access.",
        buttonLabel: "Open Connected Tools",
        icon: Wrench,
    },
];

function SettingsContent() {
    const searchParams = useSearchParams();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const validTabs: TabId[] = ["profile", "profiles", "users", "engines", "tools"];
    const requestedTab = (searchParams?.get("tab") as TabId | null) ?? null;
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

            <div className="max-w-5xl mx-auto w-full px-6 mt-6 space-y-5">
                <SettingsGuidedWorkflow advancedMode={advancedMode} activeTab={activeTab} onSelect={setActiveTab} />

                {/* Tabs */}
                <div role="tablist" aria-label="Settings sections" className="flex items-center gap-1 border-b border-cortex-border">
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

function SettingsGuidedWorkflow({
    advancedMode,
    activeTab,
    onSelect,
}: {
    advancedMode: boolean;
    activeTab: TabId;
    onSelect: (tab: TabId) => void;
}) {
    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-5 py-5 shadow-sm">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="max-w-2xl space-y-2">
                    <p className="text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">Guided setup path</p>
                    <h2 className="text-xl font-semibold text-cortex-text-main">Start with the controls most operators actually need.</h2>
                    <p className="text-sm leading-6 text-cortex-text-muted">
                        Settings should help you move from identity and workflow defaults into deeper controls without making the default path feel like a raw admin console.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                    <p className="font-medium text-cortex-text-main">Advanced setup stays intentional</p>
                    <p className="mt-1 leading-6">
                        AI Engines and Connected Tools remain available when you intentionally open Advanced mode, but they stay out of the everyday setup path by default.
                    </p>
                </div>
            </div>

            <div className="mt-5 grid gap-3 lg:grid-cols-3">
                {DEFAULT_WORKFLOW_CARDS.map((card) => (
                    <WorkflowCard key={card.id} {...card} active={activeTab === card.id} onSelect={() => onSelect(card.id)} />
                ))}
            </div>

            {advancedMode ? (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                    <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                        <div className="max-w-2xl">
                            <p className="text-sm font-semibold text-cortex-text-main">Advanced controls are open</p>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                Use these when you need deeper tuning or integrations, not as the default way into the product.
                            </p>
                        </div>
                        <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Advanced mode on</p>
                    </div>
                    <div className="mt-4 grid gap-3 lg:grid-cols-2">
                        {ADVANCED_WORKFLOW_CARDS.map((card) => (
                            <WorkflowCard key={card.id} {...card} active={activeTab === card.id} onSelect={() => onSelect(card.id)} />
                        ))}
                    </div>
                </div>
            ) : (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Advanced controls unlock when you need them</p>
                    <p className="mt-1 leading-6">
                        Turn on Advanced mode when you want to inspect AI Engines or Connected Tools. Those deeper controls stay intentionally outside the default workflow.
                    </p>
                </div>
            )}
        </section>
    );
}

function WorkflowCard({
    title,
    summary,
    buttonLabel,
    icon: Icon,
    active,
    onSelect,
}: WorkflowCardDefinition & {
    active: boolean;
    onSelect: () => void;
}) {
    return (
        <div
            className={`rounded-2xl border px-4 py-4 transition-colors ${
                active
                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                    : "border-cortex-border bg-cortex-bg"
            }`}
        >
            <div className="flex items-center gap-2 text-cortex-primary">
                <Icon className="h-4 w-4" />
                <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            </div>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
            <button
                type="button"
                onClick={onSelect}
                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
            >
                {buttonLabel}
                <ArrowRight className="h-4 w-4" />
            </button>
        </div>
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

function ProfileSettings() {
    const assistantName = useCortexStore((s) => s.assistantName);
    const updateAssistantName = useCortexStore((s) => s.updateAssistantName);
    const theme = useCortexStore((s) => s.theme);
    const updateTheme = useCortexStore((s) => s.updateTheme);
    const [nameDraft, setNameDraft] = useState(assistantName);
    const [isSavingName, setIsSavingName] = useState(false);
    const [saveMessage, setSaveMessage] = useState<string | null>(null);
    const [themeDraft, setThemeDraft] = useState(theme);
    const [isSavingTheme, setIsSavingTheme] = useState(false);
    const [themeMessage, setThemeMessage] = useState<string | null>(null);

    useEffect(() => {
        setNameDraft(assistantName);
    }, [assistantName]);

    useEffect(() => {
        setThemeDraft(theme);
    }, [theme]);

    const onSaveAssistantName = async () => {
        setIsSavingName(true);
        setSaveMessage(null);
        const ok = await updateAssistantName(nameDraft);
        setIsSavingName(false);
        setSaveMessage(ok ? "Saved" : "Failed to save");
    };

    const onSaveTheme = async () => {
        setIsSavingTheme(true);
        setThemeMessage(null);
        const ok = await updateTheme(themeDraft);
        setIsSavingTheme(false);
        setThemeMessage(ok ? "Saved" : "Failed to save");
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
                <div className="space-y-2">
                    <label className="text-cortex-text-main text-sm block" htmlFor="theme-select">Theme</label>
                    <div className="flex items-center gap-2">
                        <select
                            id="theme-select"
                            value={themeDraft}
                            onChange={(e) => setThemeDraft(e.target.value as typeof themeDraft)}
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        >
                            <option value="aero-light">Aero Light</option>
                            <option value="midnight-cortex">Midnight Cortex</option>
                            <option value="system">System</option>
                        </select>
                        <button
                            type="button"
                            onClick={onSaveTheme}
                            disabled={isSavingTheme || themeDraft === theme}
                            className="px-3 py-1 rounded border border-cortex-primary/40 text-cortex-primary text-sm font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                        >
                            {isSavingTheme ? "Saving..." : "Save"}
                        </button>
                    </div>
                    <p className="text-xs text-cortex-text-muted">
                        Pick the product surface theme for your workspace. System follows your device preference automatically.
                    </p>
                    {themeMessage && (
                        <p className="text-xs font-mono text-cortex-text-muted">{themeMessage}</p>
                    )}
                </div>
            </div>
        </div>
    );
}
