"use client";

import { useState } from "react";
import { User, Users, BrainCircuit, Settings, Shield } from "lucide-react";
import MatrixGrid from "@/components/matrix/MatrixGrid";

export default function SettingsPage() {
    const [activeTab, setActiveTab] = useState("profile");

    return (
        <div className="max-w-5xl mx-auto space-y-8 p-6">
            {/* Header */}
            <div>
                <h1 className="text-2xl font-bold tracking-tight text-cortex-text-main flex items-center gap-2">
                    <Settings className="w-6 h-6 text-cortex-text-muted" />
                    Settings
                </h1>
                <p className="text-cortex-text-muted">Manage your neural interface preferences.</p>
            </div>

            {/* Tabs */}
            <div className="flex items-center gap-1 border-b border-cortex-border">
                <Tab
                    id="profile"
                    label="Profile"
                    icon={User}
                    active={activeTab === "profile"}
                    onClick={() => setActiveTab("profile")}
                />
                <Tab
                    id="teams"
                    label="Teams"
                    icon={Users}
                    active={activeTab === "teams"}
                    onClick={() => setActiveTab("teams")}
                />
                <Tab
                    id="matrix"
                    label="Cognitive Matrix"
                    icon={BrainCircuit}
                    active={activeTab === "matrix"}
                    onClick={() => setActiveTab("matrix")}
                />
            </div>

            {/* Content */}
            <div className="min-h-[400px]">
                {activeTab === "profile" && <ProfileSettings />}
                {activeTab === "teams" && <TeamSettings />}
                {activeTab === "matrix" && <div className="p-1"><MatrixGrid /></div>}
            </div>
        </div>
    );
}

function Tab({ id, label, icon: Icon, active, onClick }: any) {
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
    )
}

function ProfileSettings() {
    return (
        <div className="space-y-6 max-w-lg">
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Appearance</h3>

                <div className="flex items-center justify-between">
                    <span className="text-cortex-text-main text-sm">Theme</span>
                    <select className="bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary">
                        <option>Vuexy Dark</option>
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
    )
}

function TeamSettings() {
    return (
        <div className="space-y-6">
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider mb-4">Operations Team</h3>

                <div className="space-y-2">
                    <div className="flex items-center justify-between p-2 hover:bg-cortex-bg rounded border border-transparent hover:border-cortex-border transition-colors">
                        <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-full bg-cortex-primary/10 flex items-center justify-center text-cortex-primary font-bold text-xs border border-cortex-primary/30">AD</div>
                            <div>
                                <p className="text-sm font-medium text-cortex-text-main">Admin User</p>
                                <p className="text-xs text-cortex-text-muted">Tier 0 (Owner)</p>
                            </div>
                        </div>
                        <Shield className="w-4 h-4 text-cortex-success" />
                    </div>
                </div>
            </div>
        </div>
    )
}
