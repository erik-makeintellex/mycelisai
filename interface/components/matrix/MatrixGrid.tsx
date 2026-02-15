"use client";

import { useEffect, useState } from "react";
import { Brain, Cpu, Zap } from "lucide-react";

interface MatrixConfig {
    profiles: Record<string, string>; // profile -> providerID
    providers: Record<string, {
        type: string;
        model_id: string;
        endpoint?: string;
    }>;
}

export default function MatrixGrid() {
    const [config, setConfig] = useState<MatrixConfig | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetch("/api/v1/cognitive/matrix")
            .then(res => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return res.json();
            })
            .then(setConfig)
            .catch(err => setError(err.message));
    }, []);

    if (error) return (
        <div className="p-4 text-xs text-cortex-text-muted font-mono">
            MATRIX OFFLINE — {error}
        </div>
    );

    if (!config) return (
        <div className="p-4 text-xs text-cortex-text-muted font-mono animate-pulse">
            Loading Matrix...
        </div>
    );

    return (
        <div className="p-4 bg-cortex-bg border border-cortex-border rounded-xl h-full flex flex-col">
            <h3 className="text-xs font-bold text-cortex-text-muted mb-4 flex items-center gap-2 font-mono">
                <Brain size={12} className="text-cortex-primary" />
                COGNITIVE_MATRIX
            </h3>

            <div className="space-y-4">
                {Object.entries(config.profiles).map(([profile, providerID]) => {
                    const provider = config.providers[providerID];
                    if (!provider) return null;

                    return (
                        <div key={profile} className="bg-cortex-surface border border-cortex-border rounded-lg p-3 flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-cortex-primary/10 text-cortex-primary rounded-md">
                                    <Zap size={14} />
                                </div>
                                <div>
                                    <div className="text-xs font-bold text-cortex-text-main capitalize">{profile}</div>
                                    <div className="text-[10px] text-cortex-text-muted">Profile</div>
                                </div>
                            </div>

                            <div className="h-px bg-cortex-border flex-1 mx-4 relative">
                                <div className="absolute right-0 top-1/2 -translate-y-1/2 w-1.5 h-1.5 bg-cortex-text-muted rounded-full" />
                            </div>

                            <div className="flex items-center gap-3">
                                <div className="text-right">
                                    <div className="text-xs font-bold text-cortex-text-main">{provider.model_id}</div>
                                    <div className="text-[10px] text-cortex-text-muted">{provider.type} ({providerID})</div>
                                </div>
                                <div className="p-2 bg-cortex-primary/10 text-cortex-primary rounded-md">
                                    <Cpu size={14} />
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}
