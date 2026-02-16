"use client";

import { useEffect } from "react";
import { Cpu, Image, Circle } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

export default function CognitiveStatusPanel() {
    const cognitiveStatus = useCortexStore((s) => s.cognitiveStatus);
    const fetchCognitiveStatus = useCortexStore((s) => s.fetchCognitiveStatus);

    useEffect(() => {
        fetchCognitiveStatus();
        const interval = setInterval(fetchCognitiveStatus, 15000);
        return () => clearInterval(interval);
    }, [fetchCognitiveStatus]);

    const text = cognitiveStatus?.text;
    const media = cognitiveStatus?.media;

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
            <h3 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-3">
                Cognitive Engines
            </h3>
            <div className="flex flex-col gap-2.5">
                {/* Text Engine (vLLM) */}
                <EngineRow
                    icon={<Cpu className="w-3.5 h-3.5" />}
                    label="Text Engine"
                    sublabel={text?.model ?? "vLLM"}
                    status={text?.status ?? "offline"}
                    endpoint={text?.endpoint}
                />
                {/* Media Engine (Diffusers) */}
                <EngineRow
                    icon={<Image className="w-3.5 h-3.5" />}
                    label="Media Engine"
                    sublabel={media?.model ?? "Diffusers"}
                    status={media?.status ?? "offline"}
                    endpoint={media?.endpoint}
                />
            </div>
        </div>
    );
}

function EngineRow({
    icon,
    label,
    sublabel,
    status,
    endpoint,
}: {
    icon: React.ReactNode;
    label: string;
    sublabel: string;
    status: "online" | "offline";
    endpoint?: string;
}) {
    const isOnline = status === "online";

    return (
        <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2.5 min-w-0">
                <div className="text-cortex-text-muted flex-shrink-0">{icon}</div>
                <div className="min-w-0">
                    <p className="text-xs font-mono font-semibold text-cortex-text-main truncate">
                        {label}
                    </p>
                    <p className="text-[9px] font-mono text-cortex-text-muted truncate">
                        {sublabel}
                        {endpoint && (
                            <span className="ml-1 opacity-50">{endpoint}</span>
                        )}
                    </p>
                </div>
            </div>
            <div className="flex items-center gap-1.5 flex-shrink-0">
                <Circle
                    className={`w-2 h-2 ${
                        isOnline
                            ? "fill-cortex-success text-cortex-success"
                            : "fill-cortex-danger text-cortex-danger"
                    }`}
                />
                <span
                    className={`text-[9px] font-mono font-bold uppercase ${
                        isOnline
                            ? "text-cortex-success"
                            : "text-cortex-danger"
                    }`}
                >
                    {status}
                </span>
            </div>
        </div>
    );
}
