"use client";

import { useState } from "react";
import { User, Users, BrainCircuit, Settings, Shield } from "lucide-react";
import MatrixGrid from "@/components/matrix/MatrixGrid";

export default function SettingsPage() {
    const [activeTab, setActiveTab] = useState("profile");

    return (
        <div className="max-w-5xl mx-auto space-y-8">
            {/* Header */}
            <div>
                <h1 className="text-2xl font-bold tracking-tight text-zinc-900 flex items-center gap-2">
                    <Settings className="w-6 h-6 text-zinc-500" />
                    Settings
                </h1>
                <p className="text-zinc-500">Manage your neural interface preferences.</p>
            </div>

            {/* Tabs */}
            <div className="flex items-center gap-1 border-b border-zinc-200">
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
                    ? "border-emerald-500 text-emerald-600"
                    : "border-transparent text-zinc-500 hover:text-zinc-900 hover:border-zinc-300"}
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
            <div className="p-6 rounded-lg border border-zinc-200 bg-white shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider">Appearance</h3>

                <div className="flex items-center justify-between">
                    <span className="text-zinc-700 text-sm">Theme</span>
                    <select className="bg-zinc-50 border border-zinc-300 rounded px-2 py-1 text-sm text-zinc-900 focus:outline-none focus:ring-1 focus:ring-emerald-500">
                        <option>Aero Light</option>
                        <option>Cyber Dark</option>
                        <option>System</option>
                    </select>
                </div>
            </div>

            <div className="p-6 rounded-lg border border-zinc-200 bg-white shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider">Notifications</h3>
                <div className="flex items-center justify-between">
                    <span className="text-zinc-700 text-sm">Task Completion</span>
                    <div className="w-8 h-4 bg-emerald-100 rounded-full relative cursor-pointer border border-emerald-200">
                        <div className="absolute right-0 w-4 h-4 bg-emerald-500 rounded-full shadow-sm"></div>
                    </div>
                </div>
            </div>
        </div>
    )
}

function TeamSettings() {
    return (
        <div className="space-y-6">
            <div className="p-6 rounded-lg border border-zinc-200 bg-white shadow-sm">
                <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider mb-4">Operations Team</h3>

                <div className="space-y-2">
                    <div className="flex items-center justify-between p-2 hover:bg-zinc-50 rounded border border-transparent hover:border-zinc-100 transition-colors">
                        <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-xs border border-blue-200">AD</div>
                            <div>
                                <p className="text-sm font-medium text-zinc-900">Admin User</p>
                                <p className="text-xs text-zinc-500">Tier 0 (Owner)</p>
                            </div>
                        </div>
                        <Shield className="w-4 h-4 text-emerald-500" />
                    </div>
                </div>
            </div>
        </div>
    )
}
