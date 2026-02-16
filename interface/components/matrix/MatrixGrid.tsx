"use client";

import { useEffect, useState } from "react";
import {
    Brain, Cpu, Zap, Save, Check, AlertCircle,
    Settings2, X, Key, Globe, Box,
} from "lucide-react";

interface ProviderInfo {
    type: string;
    model_id: string;
    endpoint?: string;
}

interface MatrixConfig {
    profiles: Record<string, string>;
    providers: Record<string, ProviderInfo>;
    media?: {
        endpoint: string;
        model_id: string;
    };
}

// Providers that need API keys (commercial/external)
const NEEDS_API_KEY: Record<string, string> = {
    production_gpt4: "OPENAI_API_KEY",
    production_claude: "ANTHROPIC_API_KEY",
    production_gemini: "GEMINI_API_KEY",
};

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
                                            <Key size={10} className="text-cortex-warning" title="API key required" />
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

function ProviderConfigModal({
    providerId,
    provider,
    onClose,
    onSaved,
}: {
    providerId: string;
    provider: ProviderInfo;
    onClose: () => void;
    onSaved: (data: any) => void;
}) {
    const [endpoint, setEndpoint] = useState(provider.endpoint ?? "");
    const [modelId, setModelId] = useState(provider.model_id ?? "");
    const [apiKeyEnv, setApiKeyEnv] = useState(NEEDS_API_KEY[providerId] ?? "");
    const [apiKey, setApiKey] = useState("");
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    async function handleSave() {
        setSaving(true);
        setError(null);
        try {
            const body: Record<string, string> = {};
            if (endpoint !== (provider.endpoint ?? "")) body.endpoint = endpoint;
            if (modelId !== provider.model_id) body.model_id = modelId;
            if (apiKeyEnv) body.api_key_env = apiKeyEnv;
            if (apiKey) body.api_key = apiKey;

            const res = await fetch(`/api/v1/cognitive/providers/${providerId}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(text);
            }
            const data = await res.json();
            onSaved(data);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    }

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl shadow-2xl w-full max-w-md">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-cortex-border">
                    <div className="flex items-center gap-2">
                        <Settings2 size={16} className="text-cortex-primary" />
                        <span className="text-sm font-mono font-bold text-cortex-text-main">
                            Configure: {providerId}
                        </span>
                    </div>
                    <button onClick={onClose} className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted">
                        <X size={16} />
                    </button>
                </div>

                {/* Form */}
                <div className="p-4 space-y-4">
                    <div>
                        <label className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider block mb-1">
                            <Globe size={10} className="inline mr-1" />
                            Endpoint URL
                        </label>
                        <input
                            type="text"
                            value={endpoint}
                            onChange={(e) => setEndpoint(e.target.value)}
                            placeholder="https://api.openai.com/v1"
                            className="w-full bg-cortex-bg border border-cortex-border rounded px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                    </div>

                    <div>
                        <label className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider block mb-1">
                            <Box size={10} className="inline mr-1" />
                            Model ID
                        </label>
                        <input
                            type="text"
                            value={modelId}
                            onChange={(e) => setModelId(e.target.value)}
                            placeholder="gpt-4-turbo"
                            className="w-full bg-cortex-bg border border-cortex-border rounded px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                    </div>

                    <div>
                        <label className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider block mb-1">
                            <Key size={10} className="inline mr-1" />
                            API Key (stored in-memory only)
                        </label>
                        <input
                            type="password"
                            value={apiKey}
                            onChange={(e) => setApiKey(e.target.value)}
                            placeholder="sk-..."
                            className="w-full bg-cortex-bg border border-cortex-border rounded px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                        <p className="text-[9px] text-cortex-text-muted mt-1">
                            Direct key — only stored in server memory, never written to disk.
                        </p>
                    </div>

                    <div>
                        <label className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider block mb-1">
                            <Key size={10} className="inline mr-1" />
                            API Key Env Variable (persisted to config)
                        </label>
                        <input
                            type="text"
                            value={apiKeyEnv}
                            onChange={(e) => setApiKeyEnv(e.target.value)}
                            placeholder="OPENAI_API_KEY"
                            className="w-full bg-cortex-bg border border-cortex-border rounded px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                        <p className="text-[9px] text-cortex-text-muted mt-1">
                            Environment variable name — the server reads the key at runtime.
                        </p>
                    </div>

                    {error && (
                        <div className="text-[10px] text-cortex-danger font-mono flex items-center gap-1">
                            <AlertCircle size={10} />
                            {error}
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="flex justify-end gap-2 px-4 py-3 border-t border-cortex-border">
                    <button
                        onClick={onClose}
                        className="px-3 py-1.5 rounded text-xs font-mono text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border transition-colors"
                    >
                        CANCEL
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="px-4 py-1.5 rounded text-xs font-mono font-bold bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/20 transition-colors disabled:opacity-50"
                    >
                        {saving ? "SAVING..." : "SAVE CONFIG"}
                    </button>
                </div>
            </div>
        </div>
    );
}
