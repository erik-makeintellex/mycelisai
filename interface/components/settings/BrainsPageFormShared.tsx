import type React from "react";
import { X } from "lucide-react";

// Split out of BrainsPage.tsx to stay within its line cap: shared types,
// presets and the Modal wrapper used by the Add/Edit provider dialogs.

export interface BrainEntry {
    id: string;
    type: string;
    endpoint?: string;
    model_id: string;
    location: string;
    data_boundary: string;
    usage_policy: string;
    token_budget_profile: string;
    max_output_tokens: number;
    roles_allowed: string[];
    enabled: boolean;
    status: string;
}

export interface ProviderFormData {
    id: string;
    type: string;
    endpoint: string;
    model_id: string;
    api_key: string;
    location: string;
    data_boundary: string;
    usage_policy: string;
    token_budget_profile: string;
    max_output_tokens: number;
    roles_allowed: string[];
    enabled: boolean;
}

// ── Type presets ──────────────────────────────────────────────────────────────

export const PROVIDER_PRESETS: Record<string, Partial<ProviderFormData>> = {
    ollama: {
        type: "openai_compatible",
        endpoint: "http://localhost:11434/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    vllm: {
        type: "openai_compatible",
        endpoint: "http://localhost:8000/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    lmstudio: {
        type: "openai_compatible",
        endpoint: "http://localhost:1234/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    openai: {
        type: "openai",
        endpoint: "https://api.openai.com/v1",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    anthropic: {
        type: "anthropic",
        endpoint: "",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    google: {
        type: "google",
        endpoint: "",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    custom: {
        type: "openai_compatible",
        endpoint: "",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
};

export const PRESET_LABELS: Record<string, string> = {
    ollama: "Ollama",
    vllm: "vLLM",
    lmstudio: "LM Studio",
    openai: "OpenAI",
    anthropic: "Anthropic",
    google: "Google",
    custom: "Custom",
};

export const COUNCIL_ROLES = ["all", "architect", "coder", "creative", "sentry", "admin"];
export const TOKEN_BUDGET_PRESETS: Record<string, { label: string; tokens: number; description: string }> = {
    conservative: { label: "Conservative", tokens: 512, description: "Short, low-cost responses." },
    standard: { label: "Standard", tokens: 1024, description: "Balanced default for everyday agentry." },
    extended: { label: "Extended", tokens: 2048, description: "Longer reasoning or coding responses." },
    deep: { label: "Deep", tokens: 4096, description: "Heavy output budget for complex work." },
};

export const blankForm = (): ProviderFormData => ({
    id: "",
    type: "openai_compatible",
    endpoint: "",
    model_id: "",
    api_key: "",
    location: "local",
    data_boundary: "local_only",
    usage_policy: "local_first",
    token_budget_profile: "standard",
    max_output_tokens: 1024,
    roles_allowed: ["all"],
    enabled: true,
});

// ── Modal wrapper ─────────────────────────────────────────────────────────────

export function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
                <div className="flex items-center justify-between px-5 py-3 border-b border-cortex-border">
                    <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                    <button onClick={onClose} className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                        <X className="w-4 h-4" />
                    </button>
                </div>
                <div className="p-5">{children}</div>
            </div>
        </div>
    );
}
