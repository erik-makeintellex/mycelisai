"use client";

import { useState } from "react";
import type { MCPLibraryEntry } from "@/store/useCortexStore";

interface EnvModalProps {
    entry: MCPLibraryEntry;
    onInstall: (env: Record<string, string>) => void;
    onClose: () => void;
}

type EnvFieldSpec = {
    name: string;
    description?: string;
    required?: boolean;
    secret?: boolean;
    default_value?: string;
};

export function EnvConfigModal({ entry, onInstall, onClose }: EnvModalProps) {
    const declaredEnv: EnvFieldSpec[] = entry.environment_variables && entry.environment_variables.length > 0
        ? entry.environment_variables
        : Object.keys(entry.env ?? {}).map((name) => ({
            name,
            default_value: entry.env?.[name] ?? "",
        }));
    const [envValues, setEnvValues] = useState<Record<string, string>>(() => {
        const init: Record<string, string> = {};
        declaredEnv.forEach((spec) => {
            init[spec.name] = spec.default_value ?? "";
        });
        return init;
    });

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-md p-6 shadow-2xl">
                <h3 className="text-sm font-mono font-bold text-cortex-text-main mb-1">Configure {entry.name}</h3>
                <p className="text-[10px] font-mono text-cortex-text-muted mb-4">Set required environment variables before installing.</p>

                <div className="flex flex-col gap-3 mb-5">
                    {declaredEnv.map((spec) => (
                        <label key={spec.name} className="flex flex-col gap-1">
                            <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                                {spec.name}{spec.required ? " *" : ""}
                            </span>
                            {spec.description && (
                                <span className="text-[10px] font-mono text-cortex-text-muted leading-relaxed">
                                    {spec.description}
                                </span>
                            )}
                            <input
                                type={spec.secret ? "password" : "text"}
                                value={envValues[spec.name]}
                                onChange={(e) => setEnvValues((v) => ({ ...v, [spec.name]: e.target.value }))}
                                placeholder={spec.default_value ? `${spec.name} (${spec.default_value})` : `Enter ${spec.name}`}
                                className="bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary/50"
                            />
                        </label>
                    ))}
                </div>

                <div className="flex justify-end gap-2">
                    <button
                        onClick={onClose}
                        className="px-3 py-1.5 rounded-lg text-xs font-mono text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => onInstall(envValues)}
                        className="px-4 py-1.5 rounded-lg bg-cortex-success/10 border border-cortex-success/30 text-xs font-mono font-bold text-cortex-success hover:bg-cortex-success/20 transition-colors"
                    >
                        Install
                    </button>
                </div>
            </div>
        </div>
    );
}
