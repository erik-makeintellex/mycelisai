"use client";

import React, { useEffect, useMemo } from "react";
import { Radio, Wifi, WifiOff, CloudSun, Mail, Database, Eye, EyeOff, ChevronRight, MessageSquare, Brain } from "lucide-react";
import { useCortexStore, type SensorNode } from "@/store/useCortexStore";

// ── Sensor Icon Registry ─────────────────────────────────────

const SENSOR_ICONS: Record<string, React.ElementType> = {
    weather: CloudSun,
    email: Mail,
    database: Database,
    messaging: MessageSquare,
    llm: Brain,
    default: Radio,
};

function getSensorIcon(type: string): React.ElementType {
    return SENSOR_ICONS[type.toLowerCase()] ?? SENSOR_ICONS.default;
}

// ── Grouped Sensor Data ──────────────────────────────────────

interface SensorGroup {
    type: string;
    sensors: SensorNode[];
    onlineCount: number;
}

function groupSensors(sensors: SensorNode[]): SensorGroup[] {
    const grouped = new Map<string, SensorNode[]>();
    for (const sensor of sensors) {
        const key = sensor.type.toLowerCase();
        if (!grouped.has(key)) grouped.set(key, []);
        grouped.get(key)!.push(sensor);
    }

    return Array.from(grouped.entries())
        .map(([type, sensors]) => ({
            type,
            sensors,
            onlineCount: sensors.filter((s) => s.status === "online").length,
        }))
        .sort((a, b) => a.type.localeCompare(b.type));
}

// ── Group Card ───────────────────────────────────────────────

function SensorGroupCard({ group }: { group: SensorGroup }) {
    const subscribed = useCortexStore((s) => s.subscribedSensorGroups);
    const toggle = useCortexStore((s) => s.toggleSensorGroup);
    const isSubscribed = subscribed.includes(group.type);
    const Icon = getSensorIcon(group.type);

    return (
        <div className={`rounded-lg border transition-all ${
            isSubscribed
                ? "bg-cortex-bg border-cyan-500/30"
                : "bg-cortex-bg border-cortex-border opacity-70"
        }`}>
            {/* Group Header */}
            <button
                onClick={() => toggle(group.type)}
                className="w-full flex items-center gap-2.5 p-2.5 hover:bg-cortex-surface/50 transition-colors rounded-lg"
            >
                <div className={`p-1.5 rounded-md ${
                    isSubscribed ? "bg-cyan-500/10 text-cyan-400" : "bg-cortex-border/50 text-cortex-text-muted"
                }`}>
                    <Icon className="w-3.5 h-3.5" />
                </div>
                <div className="flex-1 text-left min-w-0">
                    <div className="text-[11px] font-mono font-bold text-cortex-text-main uppercase">
                        {group.type}
                    </div>
                    <div className="text-[9px] font-mono text-cortex-text-muted">
                        {group.onlineCount}/{group.sensors.length} online
                    </div>
                </div>
                <div className="flex items-center gap-1.5">
                    {isSubscribed ? (
                        <Eye className="w-3.5 h-3.5 text-cyan-400" />
                    ) : (
                        <EyeOff className="w-3.5 h-3.5 text-cortex-text-muted" />
                    )}
                    <ChevronRight className={`w-3 h-3 text-cortex-text-muted transition-transform ${isSubscribed ? "rotate-90" : ""}`} />
                </div>
            </button>

            {/* Expanded Sensor List (only when subscribed) */}
            {isSubscribed && (
                <div className="px-2.5 pb-2 space-y-1">
                    {group.sensors.map((sensor) => (
                        <div
                            key={sensor.id}
                            className="flex items-center gap-2 px-2 py-1.5 rounded-md bg-cortex-surface/40"
                        >
                            <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${
                                sensor.status === "online"
                                    ? "bg-cortex-success"
                                    : sensor.status === "degraded"
                                        ? "bg-yellow-500"
                                        : "bg-red-500"
                            } ${sensor.status === "online" ? "animate-pulse" : ""}`} />
                            <span className="text-[10px] font-mono text-cortex-text-main truncate flex-1">
                                {sensor.label || sensor.id}
                            </span>
                            <Wifi className={`w-3 h-3 flex-shrink-0 ${
                                sensor.status === "online" ? "text-cortex-success" : "text-cortex-text-muted opacity-30"
                            }`} />
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

// ── Main Component ───────────────────────────────────────────

export default function SensorLibrary() {
    const sensors = useCortexStore((s) => s.sensorFeeds);
    const isFetching = useCortexStore((s) => s.isFetchingSensors);
    const fetchSensors = useCortexStore((s) => s.fetchSensors);
    const subscribedGroups = useCortexStore((s) => s.subscribedSensorGroups);

    // Fetch sensor registry once (catalog, not stream)
    useEffect(() => {
        fetchSensors();
        // Refresh catalog every 60s (not 15s — library, not stream)
        const interval = setInterval(fetchSensors, 60000);
        return () => clearInterval(interval);
    }, [fetchSensors]);

    const groups = useMemo(() => groupSensors(sensors), [sensors]);
    const totalOnline = sensors.filter((s) => s.status === "online").length;

    return (
        <div className="h-full flex flex-col bg-cortex-surface" data-testid="sensor-library">
            {/* Header */}
            <div className="px-3 py-2 border-b border-cortex-border flex items-center gap-2 flex-shrink-0">
                <Radio className="w-3.5 h-3.5 text-cyan-400" />
                <span className="text-xs font-mono font-bold text-cortex-text-muted">SENSOR LIBRARY</span>
                <span className="ml-auto flex items-center gap-1.5">
                    <span className="text-[9px] font-mono text-cortex-text-muted bg-cortex-bg px-1.5 py-0.5 rounded">
                        {totalOnline}/{sensors.length}
                    </span>
                    {subscribedGroups.length > 0 && (
                        <span className="text-[9px] font-mono text-cyan-400 bg-cyan-500/10 px-1.5 py-0.5 rounded">
                            {subscribedGroups.length} SUB
                        </span>
                    )}
                </span>
            </div>

            {/* Group List */}
            <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                {isFetching && sensors.length === 0 && (
                    <div className="space-y-2">
                        {[1, 2, 3].map((i) => (
                            <div key={i} className="h-14 rounded-lg bg-cortex-bg animate-pulse" />
                        ))}
                    </div>
                )}

                {!isFetching && sensors.length === 0 && (
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        <WifiOff className="w-8 h-8 mb-2 opacity-30" />
                        <p className="text-xs font-mono">No sensor groups</p>
                        <p className="text-[10px] mt-1 opacity-50">Periphery offline</p>
                    </div>
                )}

                {groups.map((group) => (
                    <SensorGroupCard key={group.type} group={group} />
                ))}
            </div>
        </div>
    );
}
