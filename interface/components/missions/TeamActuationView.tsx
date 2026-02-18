"use client";

import React, { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Users, Package, BarChart3 } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import TeamRoster from "./TeamRoster";
import AgentActivityFeed from "./AgentActivityFeed";
import ArtifactViewer from "./ArtifactViewer";
import MissionSummaryTab from "./MissionSummaryTab";

type ActiveTab = "teams" | "artifacts" | "summary";

interface TabDef {
    key: ActiveTab;
    label: string;
    icon: React.ReactNode;
}

const TABS: TabDef[] = [
    { key: "teams", label: "Teams", icon: <Users className="w-3 h-3" /> },
    { key: "artifacts", label: "Artifacts", icon: <Package className="w-3 h-3" /> },
    { key: "summary", label: "Summary", icon: <BarChart3 className="w-3 h-3" /> },
];

export default function TeamActuationView({
    paramsPromise,
}: {
    paramsPromise: Promise<{ id: string }>;
}) {
    const { id: missionId } = React.use(paramsPromise);
    const router = useRouter();
    const [activeTab, setActiveTab] = useState<ActiveTab>("teams");

    const missions = useCortexStore((s) => s.missions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);
    const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);

    useEffect(() => {
        fetchMissions();
        fetchArtifacts({ mission_id: missionId });
    }, [missionId, fetchMissions, fetchArtifacts]);

    const mission = useMemo(
        () => missions.find((m) => m.id === missionId),
        [missions, missionId]
    );

    if (!mission) {
        return (
            <div className="h-full flex flex-col items-center justify-center bg-cortex-bg text-cortex-text-main">
                <Package className="w-12 h-12 opacity-20 mb-4" />
                <p className="font-mono text-sm text-cortex-text-muted">
                    Mission not found
                </p>
                <p className="font-mono text-xs text-cortex-text-muted/60 mt-1">
                    ID: {missionId}
                </p>
                <button
                    onClick={() => router.push("/")}
                    className="mt-6 px-4 py-2 rounded bg-cortex-primary/10 border border-cortex-primary/30 hover:bg-cortex-primary/20 text-cortex-primary text-xs font-mono transition-all"
                >
                    Back to Mission Control
                </button>
            </div>
        );
    }

    const renderRightPanel = () => {
        switch (activeTab) {
            case "teams":
                return <AgentActivityFeed missionId={missionId} />;
            case "artifacts":
                return <ArtifactViewer missionId={missionId} />;
            case "summary":
                return (
                    <MissionSummaryTab
                        missionId={missionId}
                        teamCount={mission.teams}
                        agentCount={mission.agents}
                    />
                );
        }
    };

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            {/* Header */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-3">
                    <button
                        onClick={() => router.push("/")}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Back to Mission Control"
                    >
                        <ArrowLeft className="w-4 h-4" />
                    </button>
                    <div className="h-4 w-px bg-cortex-border" />
                    <h1 className="font-mono font-bold text-sm text-cortex-text-main tracking-wide">
                        MISSION:{" "}
                        <span className="text-cortex-primary">
                            {missionId.slice(0, 12).toUpperCase()}
                        </span>
                    </h1>
                    <span
                        className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded ${
                            mission.status === "active"
                                ? "bg-cortex-success/15 text-cortex-success border border-cortex-success/30"
                                : mission.status === "completed"
                                ? "bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                                : "bg-cortex-danger/15 text-cortex-danger border border-cortex-danger/30"
                        }`}
                    >
                        {mission.status}
                    </span>
                </div>

                {/* Tab buttons */}
                <div className="flex items-center gap-1">
                    {TABS.map((tab) => (
                        <button
                            key={tab.key}
                            onClick={() => setActiveTab(tab.key)}
                            className={`flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-mono transition-all ${
                                activeTab === tab.key
                                    ? "bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                                    : "text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border/50"
                            }`}
                        >
                            {tab.icon}
                            {tab.label}
                        </button>
                    ))}
                </div>
            </header>

            {/* Body: Two-column split */}
            <div
                className="flex-1 overflow-hidden"
                style={{
                    display: "grid",
                    gridTemplateColumns: "minmax(260px, 2fr) minmax(400px, 3fr)",
                    minHeight: 0,
                }}
            >
                {/* Left: TeamRoster */}
                <div className="h-full overflow-hidden border-r border-cortex-border">
                    <TeamRoster
                        missionId={missionId}
                        teamCount={mission.teams}
                        agentCount={mission.agents}
                    />
                </div>

                {/* Right: Tab content */}
                <div className="h-full overflow-hidden relative">
                    {renderRightPanel()}
                </div>
            </div>
        </div>
    );
}
