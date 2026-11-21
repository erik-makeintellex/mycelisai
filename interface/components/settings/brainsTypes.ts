// ── Types ─────────────────────────────────────────────────────────────────────

export interface BrainEntry {
    id: string;
    type: string;
    endpoint?: string;
    model_id: string;
    api_key_env?: string;
    auth_configured?: boolean;
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
    api_key_env: string;
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
    openai: {
        type: "openai",
        endpoint: "https://api.openai.com/v1",
        model_id: "gpt-4o",
        api_key_env: "OPENAI_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    anthropic: {
        type: "anthropic",
        endpoint: "https://api.anthropic.com/v1/messages",
        model_id: "claude-3-7-sonnet-latest",
        api_key_env: "ANTHROPIC_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    google: {
        type: "google",
        endpoint: "https://generativelanguage.googleapis.com/v1beta/models",
        model_id: "gemini-2.5-pro",
        api_key_env: "GEMINI_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    openrouter: {
        type: "openai_compatible",
        endpoint: "https://openrouter.ai/api/v1",
        model_id: "anthropic/claude-3.7-sonnet",
        api_key_env: "OPENROUTER_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    groq: {
        type: "openai_compatible",
        endpoint: "https://api.groq.com/openai/v1",
        model_id: "llama-3.3-70b-versatile",
        api_key_env: "GROQ_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    deepseek: {
        type: "openai_compatible",
        endpoint: "https://api.deepseek.com/v1",
        model_id: "deepseek-chat",
        api_key_env: "DEEPSEEK_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    mistral: {
        type: "openai_compatible",
        endpoint: "https://api.mistral.ai/v1",
        model_id: "mistral-large-latest",
        api_key_env: "MISTRAL_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    xai: {
        type: "openai_compatible",
        endpoint: "https://api.x.ai/v1",
        model_id: "grok-2-latest",
        api_key_env: "XAI_API_KEY",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        token_budget_profile: "extended",
        max_output_tokens: 2048,
        roles_allowed: ["all"],
    },
    ollama: {
        type: "openai_compatible",
        endpoint: "http://localhost:11434/v1",
        model_id: "qwen2.5-coder:7b-instruct",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    vllm: {
        type: "openai_compatible",
        endpoint: "http://localhost:8000/v1",
        model_id: "qwen2.5-coder",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    lmstudio: {
        type: "openai_compatible",
        endpoint: "http://localhost:1234/v1",
        model_id: "default",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
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
    openai: "OpenAI",
    anthropic: "Anthropic Claude",
    google: "Google Gemini",
    openrouter: "OpenRouter",
    groq: "Groq Cloud",
    deepseek: "DeepSeek",
    mistral: "Mistral AI",
    xai: "xAI Grok",
    ollama: "Ollama",
    vllm: "vLLM",
    lmstudio: "LM Studio",
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
    api_key_env: "",
    location: "local",
    data_boundary: "local_only",
    usage_policy: "local_first",
    token_budget_profile: "standard",
    max_output_tokens: 1024,
    roles_allowed: ["all"],
    enabled: true,
});

export const toProviderFormData = (b: BrainEntry): ProviderFormData => ({
    id: b.id,
    type: b.type,
    endpoint: b.endpoint ?? "",
    model_id: b.model_id,
    api_key: "",
    api_key_env: b.api_key_env ?? "",
    location: b.location,
    data_boundary: b.data_boundary,
    usage_policy: b.usage_policy,
    token_budget_profile: b.token_budget_profile || "standard",
    max_output_tokens: b.max_output_tokens || 1024,
    roles_allowed: b.roles_allowed?.length ? b.roles_allowed : ["all"],
    enabled: b.enabled,
});
