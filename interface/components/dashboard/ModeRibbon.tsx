"use client";

import React from "react";
import { Brain, Shield, User, Zap, Globe, AlertTriangle } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { brainDisplayName, brainLocationLabel, MODE_LABELS, GOV_POSTURE_LABELS } from "@/lib/labels";

export default function ModeRibbon() {
    const activeBrain = useCortexStore((s) => s.activeBrain);
    const activeMode = useCortexStore((s) => s.activeMode);
    const activeRole = useCortexStore((s) => s.activeRole);
    const governanceMode = useCortexStore((s) => s.governanceMode);
    const isConnected = useCortexStore((s) => s.isStreamConnected);
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);

    const modeInfo = MODE_LABELS[activeMode] || MODE_LABELS.answer;
    const govInfo = GOV_POSTURE_LABELS[governanceMode] || GOV_POSTURE_LABELS.passive;

    // Format role display
    const roleDisplay = activeRole
        ? activeRole.replace("council-", "").charAt(0).toUpperCase() + activeRole.replace("council-", "").slice(1)
        : "—";

    // Brain display
    const brainName = activeBrain ? brainDisplayName(activeBrain.provider_id) : "—";
    const brainLoc = activeBrain ? brainLocationLabel(activeBrain.location) : "";
    const isRemote = activeBrain?.location === "remote";

    return (
        <button
            onClick={() => setStatusDrawerOpen(true)}
            className="h-8 border-b border-cortex-border bg-cortex-surface/30 flex items-center px-4 gap-4 flex-shrink-0 font-mono text-[11px] w-full text-left hover:bg-cortex-surface/50 transition-colors"
            title="Open Status Drawer"
        >
            {/* Mode */}
            <RibbonChip
                icon={Zap}
                label="MODE"
                value={modeInfo.label}
                valueClass={modeInfo.color}
            />

            <Divider />

            {/* Role */}
            <RibbonChip
                icon={User}
                label="ROLE"
                value={roleDisplay}
                valueClass="text-cortex-text-main"
            />

            <Divider />

            {/* Brain */}
            <RibbonChip
                icon={isRemote ? Globe : Brain}
                label="BRAIN"
                value={activeBrain ? `${brainName} (${brainLoc})` : "—"}
                valueClass={isRemote ? "text-amber-400" : "text-cortex-text-main"}
            />
            {isRemote && (
                <span className="flex items-center gap-1 text-amber-400">
                    <AlertTriangle className="w-3 h-3" />
                    <span className="text-[10px]">REMOTE</span>
                </span>
            )}

            <Divider />

            {/* Governance */}
            <RibbonChip
                icon={Shield}
                label="GOV"
                value={govInfo.label}
                valueClass={govInfo.color}
            />

            {/* Spacer */}
            <div className="flex-1" />

            {/* Connection indicator (compact) */}
            <div className="flex items-center gap-1.5">
                <div className={`w-1.5 h-1.5 rounded-full ${isConnected ? "bg-cortex-success" : "bg-cortex-danger"}`} />
                <span className="text-cortex-text-muted text-[10px]">
                    {isConnected ? "LIVE" : "OFFLINE"}
                </span>
            </div>
        </button>
    );
}

function RibbonChip({ icon: Icon, label, value, valueClass }: {
    icon: React.ElementType;
    label: string;
    value: string;
    valueClass: string;
}) {
    return (
        <div className="flex items-center gap-1.5">
            <Icon className="w-3 h-3 text-cortex-text-muted" />
            <span className="text-cortex-text-muted">{label}:</span>
            <span className={valueClass}>{value}</span>
        </div>
    );
}

function Divider() {
    return <div className="h-3 w-px bg-cortex-border" />;
}
