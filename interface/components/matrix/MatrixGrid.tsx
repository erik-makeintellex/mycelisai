"use client";

import { useEffect, useState } from "react";
import {
    Brain, Cpu, Zap, Save, Check, AlertCircle,
    Settings2, Key,
} from "lucide-react";
import {
    NEEDS_API_KEY,
    ProviderConfigModal,
    type MatrixConfig,
    type ProviderInfo,
} from "./MatrixGridProviderModal";

export { NEEDS_API_KEY, ProviderConfigModal, type MatrixConfig, type ProviderInfo };

export default function MatrixGrid() {
    const [config, setConfig] = useState<MatrixConfig | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [editedProfiles, setEditedProfiles] = useState<Record<string, string>>({});
    const [saving, setSaving] = useState(false);
    const [saved, setSaved] = useState(false);
    const [configuring, setConfiguring] = useState<string | null>(null);

    useEffect(() => {
        fetch("/api/v1/cognitive/config")
            .then((res) => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return res.json();
            })
            .then((data) => {
                setConfig(data);
                setEditedProfiles({ ...data.profiles });
            })
            .catch((err) => setError(err.message));
    }, []);

    const hasChanges =
        config &&
        Object.entries(editedProfiles).some(
            ([k, v]) => config.profiles[k] !== v
        );

    async function handleSave() {
        if (!hasChanges) return;
        setSaving(true);
        setSaved(false);
        try {
            const res = await fetch("/api/v1/cognitive/profiles", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ profiles: editedProfiles }),
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(text);
            }
            const updated = await res.json();
            setConfig(updated);
            setEditedProfiles({ ...updated.profiles });
            setSaved(true);
            setTimeout(() => setSaved(false), 2000);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    }

    if (error)
        return (
            <div className="p-4 text-xs text-cortex-danger font-mono flex items-center gap-2">
                <AlertCircle size={14} />
                MATRIX OFFLINE — {error}
            </div>
        );

    if (!config)
        return (
            <div className="p-4 text-xs text-cortex-text-muted font-mono animate-pulse">
                Loading Matrix...
            </div>
        );

    const providerIDs = Object.keys(config.providers);

    return (
        <div className="space-y-6">
            {/* Profile → Provider Mapping */}
            <div className="bg-cortex-bg border border-cortex-border rounded-xl p-4">
                <div className="flex items-center justify-between mb-4">
                    <h3 className="text-xs font-bold text-cortex-text-muted flex items-center gap-2 font-mono uppercase tracking-wider">
                        <Brain size={12} className="text-cortex-primary" />
                        Profile Routing
                    </h3>
                    <button
                        onClick={handleSave}
                        disabled={!hasChanges || saving}
                        className={`flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-mono font-bold transition-all ${
                            saved
                                ? "bg-cortex-success/20 text-cortex-success border border-cortex-success/30"
                                : hasChanges
                                  ? "bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/20"
                                  : "bg-cortex-border/30 text-cortex-text-muted border border-cortex-border cursor-not-allowed opacity-50"
                        }`}
                    >
                        {saved ? <Check size={12} /> : <Save size={12} />}
                        {saving ? "SAVING..." : saved ? "SAVED" : "SYNC MATRIX"}
                    </button>
                </div>

                <div className="space-y-2">
                    {Object.entries(editedProfiles).map(([profile, providerID]) => {
                        const changed = config.profiles[profile] !== providerID;

                        return (
                            <div
                                key={profile}
                                className={`bg-cortex-surface border rounded-lg p-3 flex items-center justify-between gap-4 transition-colors ${
                                    changed ? "border-cortex-primary/50" : "border-cortex-border"
                                }`}
                            >
                                <div className="flex items-center gap-3 min-w-0">
                                    <div className="p-2 bg-cortex-primary/10 text-cortex-primary rounded-md flex-shrink-0">
                                        <Zap size={14} />
                                    </div>
                                    <div className="min-w-0">
                                        <div className="text-xs font-bold text-cortex-text-main capitalize">
                                            {profile}
                                        </div>
                                        <div className="text-[10px] text-cortex-text-muted">Profile</div>
                                    </div>
                                </div>

                                <div className="h-px bg-cortex-border flex-1 mx-2 relative hidden sm:block">
                                    <div className="absolute right-0 top-1/2 -translate-y-1/2 w-1.5 h-1.5 bg-cortex-text-muted rounded-full" />
                                </div>

                                <div className="flex items-center gap-3 flex-shrink-0">
                                    <select
                                        value={providerID}
                                        onChange={(e) =>
                                            setEditedProfiles((prev) => ({
                                                ...prev,
                                                [profile]: e.target.value,
                                            }))
                                        }
                                        className="bg-cortex-bg border border-cortex-border rounded px-2 py-1.5 text-xs font-mono text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                                    >
                                        {providerIDs.map((pid) => (
                                            <option key={pid} value={pid}>
                                                {config.providers[pid].model_id} ({pid})
                                            </option>
                                        ))}
                                    </select>
                                    <div className="p-2 bg-cortex-primary/10 text-cortex-primary rounded-md">
                                        <Cpu size={14} />
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>

            {/* Provider Registry (editable) */}
            <div className="bg-cortex-bg border border-cortex-border rounded-xl p-4">
                <h3 className="text-xs font-bold text-cortex-text-muted flex items-center gap-2 font-mono uppercase tracking-wider mb-3">
                    <Cpu size={12} className="text-cortex-text-muted" />
                    Registered Providers
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    {Object.entries(config.providers).map(([id, prov]) => {
                        const needsKey = id in NEEDS_API_KEY;
                        return (
                            <div
                                key={id}
                                className="bg-cortex-surface border border-cortex-border rounded-lg p-3 group hover:border-cortex-text-muted transition-colors cursor-pointer"
                                onClick={() => setConfiguring(id)}
                            >
                                <div className="flex items-center justify-between mb-1">
                                    <span className="text-xs font-bold text-cortex-text-main font-mono">
                                        {id}
                                    </span>
                                    <div className="flex items-center gap-1">
                                        {needsKey && (
                                            <span title="API key required"><Key size={10} className="text-cortex-warning" /></span>
                                        )}
                                        <span className="text-[10px] text-cortex-text-muted font-mono px-1.5 py-0.5 bg-cortex-bg rounded border border-cortex-border">
                                            {prov.type}
                                        </span>
                                        <Settings2 size={10} className="text-cortex-text-muted opacity-0 group-hover:opacity-100 transition-opacity" />
                                    </div>
                                </div>
                                <div className="text-[10px] text-cortex-text-muted font-mono truncate">
                                    {prov.model_id}
                                </div>
                                {prov.endpoint && (
                                    <div className="text-[10px] text-cortex-text-muted font-mono truncate opacity-60">
                                        {prov.endpoint}
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>

            {/* Provider Config Modal */}
            {configuring && (
                <ProviderConfigModal
                    providerId={configuring}
                    provider={config.providers[configuring]}
                    onClose={() => setConfiguring(null)}
                    onSaved={(updated) => {
                        setConfig((prev) =>
                            prev
                                ? {
                                      ...prev,
                                      providers: {
                                          ...prev.providers,
                                          [configuring]: {
                                              ...prev.providers[configuring],
                                              ...(updated.endpoint && { endpoint: updated.endpoint }),
                                              ...(updated.model_id && { model_id: updated.model_id }),
                                              ...(updated.type && { type: updated.type }),
                                          },
                                      },
                                  }
                                : prev
                        );
                        setConfiguring(null);
                    }}
                />
            )}
        </div>
    );
}
