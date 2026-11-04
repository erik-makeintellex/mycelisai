"use client";

import { useState } from "react";
import { AlertCircle, Box, Globe, Key, Settings2, X } from "lucide-react";

export interface ProviderInfo {
    type: string;
    model_id: string;
    endpoint?: string;
}

export interface MatrixConfig {
    profiles: Record<string, string>;
    providers: Record<string, ProviderInfo>;
    media?: {
        endpoint: string;
        model_id: string;
    };
}

// Providers that need API keys (commercial/external)
export const NEEDS_API_KEY: Record<string, string> = {
    production_gpt4: "OPENAI_API_KEY",
    production_claude: "ANTHROPIC_API_KEY",
    production_gemini: "GEMINI_API_KEY",
};

export function ProviderConfigModal({
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
