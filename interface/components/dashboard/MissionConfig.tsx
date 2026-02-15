"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { X, Rocket, Shield, Zap } from "lucide-react";

interface MissionConfigProps {
    onClose: () => void;
}

export default function MissionConfig({ onClose }: MissionConfigProps) {
    const router = useRouter();
    const [name, setName] = useState("");
    const [type, setType] = useState("action");

    const handleEngage = () => {
        if (!name) return;
        onClose();
        router.push("/wiring");
    };

    return (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
            <div className="bg-cortex-bg border border-cortex-border rounded-xl w-full max-w-md shadow-2xl overflow-hidden relative">
                {/* Header */}
                <div className="p-4 border-b border-cortex-border flex justify-between items-center bg-cortex-surface/50">
                    <h2 className="text-cortex-text-main font-mono font-bold flex items-center gap-2">
                        <Rocket className="w-4 h-4 text-cortex-primary" />
                        INITIATE MISSION
                    </h2>
                    <button
                        onClick={onClose}
                        className="text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 space-y-6">
                    {/* Identity */}
                    <div className="space-y-2">
                        <label className="text-xs font-mono text-cortex-text-muted uppercase tracking-widest">
                            Operation Name
                        </label>
                        <input
                            type="text"
                            className="w-full bg-cortex-surface border border-cortex-border rounded p-2 text-cortex-text-main focus:border-cortex-primary focus:outline-none placeholder-cortex-text-muted font-mono text-sm"
                            placeholder="e.g. ALPHA OVERWATCH"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                        />
                    </div>

                    {/* Type Selection */}
                    <div className="space-y-2">
                        <label className="text-xs font-mono text-cortex-text-muted uppercase tracking-widest">
                            Protocol Type
                        </label>
                        <div className="grid grid-cols-2 gap-3">
                            <button
                                onClick={() => setType("action")}
                                className={`p-3 rounded-lg border text-left transition-all ${type === "action" ? "bg-cortex-primary/10 border-cortex-primary/50" : "bg-cortex-surface border-cortex-border hover:border-cortex-text-muted"}`}
                            >
                                <Zap
                                    className={`w-5 h-5 mb-2 ${type === "action" ? "text-cortex-primary" : "text-cortex-text-muted"}`}
                                />
                                <div
                                    className={`text-sm font-bold ${type === "action" ? "text-cortex-text-main" : "text-cortex-text-muted"}`}
                                >
                                    ACTION
                                </div>
                                <div className="text-[10px] text-cortex-text-muted mt-1 leading-tight">
                                    Logic & Execution focus. High throughput.
                                </div>
                            </button>

                            <button
                                onClick={() => setType("expression")}
                                className={`p-3 rounded-lg border text-left transition-all ${type === "expression" ? "bg-indigo-950/30 border-indigo-500/50" : "bg-cortex-surface border-cortex-border hover:border-cortex-text-muted"}`}
                            >
                                <Shield
                                    className={`w-5 h-5 mb-2 ${type === "expression" ? "text-indigo-400" : "text-cortex-text-muted"}`}
                                />
                                <div
                                    className={`text-sm font-bold ${type === "expression" ? "text-cortex-text-main" : "text-cortex-text-muted"}`}
                                >
                                    EXPRESSION
                                </div>
                                <div className="text-[10px] text-cortex-text-muted mt-1 leading-tight">
                                    Output & Presentation focus. Creative tasks.
                                </div>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Footer */}
                <div className="p-4 border-t border-cortex-border bg-cortex-surface/30 flex justify-end">
                    <button
                        onClick={handleEngage}
                        disabled={!name}
                        className={`px-4 py-2 rounded-lg text-sm font-mono font-bold flex items-center gap-2 transition-all ${!name ? "bg-cortex-border text-cortex-text-muted cursor-not-allowed" : "bg-cortex-primary hover:bg-cortex-primary/80 text-white shadow-[0_0_15px_rgba(115,103,240,0.4)]"}`}
                    >
                        ENGAGE SYSTEM
                        <Rocket className="w-4 h-4" />
                    </button>
                </div>
            </div>
        </div>
    );
}
