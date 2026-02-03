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

    useEffect(() => {
        fetch("/api/v1/cognitive/matrix")
            .then(res => res.json())
            .then(setConfig)
            .catch(console.error);
    }, []);

    if (!config) return <div className="p-4 text-xs text-slate-500">Loading Matrix...</div>;

    return (
        <div className="p-4 bg-[#0b101a] border border-slate-800 rounded-xl h-full flex flex-col">
            <h3 className="text-xs font-bold text-slate-400 mb-4 flex items-center gap-2">
                <Brain size={12} className="text-purple-500" />
                COGNITIVE_MATRIX
            </h3>

            <div className="space-y-4">
                {Object.entries(config.profiles).map(([profile, providerID]) => {
                    const provider = config.providers[providerID];
                    if (!provider) return null;

                    return (
                        <div key={profile} className="bg-slate-900 border border-slate-800 rounded-lg p-3 flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-purple-500/10 text-purple-400 rounded-md">
                                    <Zap size={14} />
                                </div>
                                <div>
                                    <div className="text-xs font-bold text-slate-200 capitalize">{profile}</div>
                                    <div className="text-[10px] text-slate-500">Profile</div>
                                </div>
                            </div>

                            <div className="h-px bg-slate-800 flex-1 mx-4 relative">
                                <div className="absolute right-0 top-1/2 -translate-y-1/2 w-1.5 h-1.5 bg-slate-600 rounded-full" />
                            </div>

                            <div className="flex items-center gap-3">
                                <div className="text-right">
                                    <div className="text-xs font-bold text-slate-200">{provider.model_id}</div>
                                    <div className="text-[10px] text-slate-500">{provider.type} ({providerID})</div>
                                </div>
                                <div className="p-2 bg-blue-500/10 text-blue-400 rounded-md">
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
