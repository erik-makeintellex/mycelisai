"use client";

import React, { useEffect } from "react";
import { Radio, Wifi, WifiOff, CloudSun, Mail, Database } from "lucide-react";
import { useCortexStore, type SensorNode } from "@/store/useCortexStore";

const SENSOR_ICONS: Record<string, React.ElementType> = {
    weather: CloudSun,
    email: Mail,
    database: Database,
    default: Radio,
};

function getSensorIcon(type: string): React.ElementType {
    return SENSOR_ICONS[type.toLowerCase()] ?? SENSOR_ICONS.default;
}

function StatusDot({ status }: { status: SensorNode["status"] }) {
    const colors: Record<string, string> = {
        online: "bg-cortex-success",
        offline: "bg-red-500",
        degraded: "bg-yellow-500",
    };
    return (
        <span className={`w-2 h-2 rounded-full ${colors[status] ?? colors.offline} ${status === "online" ? "animate-pulse" : ""}`} />
    );
}

export default function SensoryPeriphery() {
    const sensors = useCortexStore((s) => s.sensorFeeds);
    const isFetching = useCortexStore((s) => s.isFetchingSensors);
    const fetchSensors = useCortexStore((s) => s.fetchSensors);

    useEffect(() => {
        fetchSensors();
        const interval = setInterval(fetchSensors, 15000);
        return () => clearInterval(interval);
    }, [fetchSensors]);

    return (
        <div className="h-full flex flex-col bg-cortex-surface border-r border-cortex-border" data-testid="sensory-periphery">
            {/* Header */}
            <div className="px-3 py-2 border-b border-cortex-border flex items-center gap-2">
                <Radio className="w-3.5 h-3.5 text-cyan-400" />
                <span className="text-xs font-mono font-bold text-cortex-text-muted">SENSORY PERIPHERY</span>
                <span className="ml-auto text-[10px] font-mono text-cortex-text-muted bg-cortex-bg px-1.5 py-0.5 rounded">
                    {sensors.length}
                </span>
            </div>

            {/* Sensor List */}
            <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                {isFetching && sensors.length === 0 && (
                    <div className="space-y-2">
                        {[1, 2, 3].map((i) => (
                            <div key={i} className="h-12 rounded-lg bg-cortex-bg animate-pulse" />
                        ))}
                    </div>
                )}

                {!isFetching && sensors.length === 0 && (
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        <WifiOff className="w-8 h-8 mb-2 opacity-30" />
                        <p className="text-xs font-mono">No sensors detected</p>
                        <p className="text-[10px] mt-1 opacity-50">Periphery offline</p>
                    </div>
                )}

                {sensors.map((sensor) => {
                    const Icon = getSensorIcon(sensor.type);
                    return (
                        <div
                            key={sensor.id}
                            className="flex items-center gap-2.5 p-2 rounded-lg bg-cortex-bg border border-cortex-border hover:border-cyan-500/30 transition-colors cursor-default"
                        >
                            <div className="p-1.5 rounded-md bg-cyan-500/10 text-cyan-400">
                                <Icon className="w-3.5 h-3.5" />
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="text-xs font-mono text-cortex-text-main truncate">
                                    {sensor.label || sensor.id}
                                </div>
                                <div className="text-[10px] text-cortex-text-muted font-mono">
                                    {sensor.type}
                                </div>
                            </div>
                            <div className="flex items-center gap-1.5">
                                <StatusDot status={sensor.status} />
                                <Wifi className={`w-3 h-3 ${sensor.status === "online" ? "text-cortex-success" : "text-cortex-text-muted opacity-30"}`} />
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}
