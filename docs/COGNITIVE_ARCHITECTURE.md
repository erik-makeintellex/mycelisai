# Cognitive Architecture

> Back to [README](../README.md) | See also: [Swarm Operations](SWARM_OPERATIONS.md) | [API Reference](API_REFERENCE.md) | [Agent Source Instantiation Template](architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md)

Mycelis supports **multiple self-hosted and commercial inference engines** — configure any combination of vLLM, Ollama, LM Studio, OpenAI, Anthropic, and Google via `cognitive.yaml` or the AI Engines settings surface.

## Provider Registry

| Provider ID | Type | Default Endpoint | Description |
| :--- | :--- | :--- | :--- |
| `vllm` | `openai_compatible` | `http://127.0.0.1:8000/v1` | vLLM inference server — high throughput, GPU-optimized |
| `ollama` | `openai_compatible` | `http://127.0.0.1:11434/v1` | Ollama — local model runner, easy setup |
| `lmstudio` | `openai_compatible` | `http://127.0.0.1:1234/v1` | LM Studio — GUI-based local inference |
| `production_gpt4` | `openai` | `https://api.openai.com/v1` | OpenAI GPT-4 (requires `OPENAI_API_KEY`) |
| `production_claude` | `anthropic` | — | Anthropic Claude (requires `ANTHROPIC_API_KEY`) |
| `production_gemini` | `google` | — | Google Gemini (requires `GEMINI_API_KEY`) |

All `openai_compatible` providers can point to **any host on the network** — they are not restricted to localhost. Configure endpoints via `/settings` → **AI Engines** (Advanced mode) or edit `core/config/cognitive.yaml` directly.

Startup behavior:
- Mycelis only performs startup connectivity calibration against default `ollama` plus providers explicitly routed by active profiles.
- Declared-but-unrouted backends are not startup-probed unless you route profiles to them.

## Profile Routing

Profiles map agent roles to providers. Each profile routes to a provider ID:

```yaml
profiles:
  admin: "ollama"       # Could be "vllm", "lmstudio", "production_gpt4", etc.
  architect: "ollama"
  coder: "vllm"
  creative: "ollama"
  sentry: "ollama"
  chat: "ollama"
```

- **Default Model:** `qwen2.5-coder:7b` (via Ollama).
- **Agent Overrides:** Each agent can specify a custom `model` field to override the profile default.

## AI Engines UI

Navigate to `/settings` → **AI Engines** (Advanced mode):

- Click a **profile** to change which provider it routes to
- Click a **provider** to configure endpoint, model ID, and API keys
- Changes persist to `cognitive.yaml` via `PUT /api/v1/cognitive/profiles` and `PUT /api/v1/cognitive/providers/{id}`

## Live Health Probing

`GET /api/v1/cognitive/status` returns real-time health for all providers:

- **Text engines** (vLLM, Ollama, LM Studio): probed via `LLMProvider.Probe()` — checks endpoint reachability
- **Media engine** (Diffusers/SDXL): probed via HTTP GET to the configured endpoint
- Status: `online` / `offline` / `error`

The frontend `CognitiveStatusPanel` polls this endpoint every 15 seconds.

## Configuration File

`core/config/cognitive.yaml`:

```yaml
providers:
  vllm:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:8000/v1"
    model_id: "qwen2.5-coder"
    api_key: "mycelis-local"
    enabled: false

  ollama:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:11434/v1"
    model_id: "qwen2.5-coder:7b"
    api_key: "ollama"
    enabled: true

  lmstudio:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:1234/v1"
    model_id: "default"
    api_key: "lm-studio"
    enabled: false

  production_gpt4:
    type: "openai"
    endpoint: "https://api.openai.com/v1"
    model_id: "gpt-4-turbo"
    api_key_env: "OPENAI_API_KEY"

profiles:
  sentry: "ollama"
  architect: "ollama"
  creative: "ollama"
  coder: "ollama"
  chat: "ollama"
  admin: "ollama"

media:
  endpoint: "http://127.0.0.1:8001/v1"
  model_id: "stable-diffusion-xl"
```

## Embedding

Mycelis uses `nomic-embed-text` (768 dimensions) for semantic vector operations:

- **Archivist auto-embed:** SitReps → `context_vectors` table (pgvector, cosine distance)
- **Memory tools:** `remember` stores both RDBMS record + vector embedding; `recall` searches both using team-aware scope when execution context provides it
- **Promotion boundary:** automatic planning continuity is kept in temporary continuity channels; only deliberate durable memory promotion should enter pgvector-backed recall
- **Fallback chain:** `Router.Embed()` tries each provider that implements `EmbedProvider`

## Hardware Grading

| Tier | RAM | Supported Models | Use Case |
| :--- | :--- | :--- | :--- |
| **Tier 1 (Min)** | 16 GB | 7B Models (Q4) | Basic Coding, CLI |
| **Tier 2 (Rec)** | 32 GB | 14B - 32B Models | Complex Architecture, Deep Reasoning |
| **Tier 3 (Ultra)** | 64 GB+ | 70B+ or Multi-Model | **Enterprise Core** (Current Dev Host) |

The system auto-detects resources but defaults to the 7B model for speed.
