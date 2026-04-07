# Cognitive Architecture
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> Back to [README](../README.md) | See also: [Swarm Operations](SWARM_OPERATIONS.md) | [API Reference](API_REFERENCE.md) | [Agent Source Instantiation Template](architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md)

## TOC

- [Provider Registry](#provider-registry)
- [Provider Auth Contract](#provider-auth-contract)
- [Profile Routing](#profile-routing)
- [AI Engines UI](#ai-engines-ui)
- [Live Health Probing](#live-health-probing)
- [Configuration File](#configuration-file)
- [Local Model Switching](#local-model-switching)
- [Embedding](#embedding)
- [Hardware Grading](#hardware-grading)

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

## Provider Auth Contract

Current supported auth patterns:

| Provider | Runtime type | Auth used by Mycelis | Notes |
| :--- | :--- | :--- | :--- |
| Ollama | `openai_compatible` | Bearer-style client key is sent, but Ollama ignores the placeholder value | Default local engine, `/v1` endpoint on `11434` |
| vLLM | `openai_compatible` | Bearer-style client key is sent; vLLM can enforce it when started with `--api-key` | Optional local engine, `/v1` endpoint on `8000` |
| LM Studio | `openai_compatible` | Bearer-style client key is sent; LM Studio compatibility mode may ignore it | Optional local engine, `/v1` endpoint on `1234` |
| OpenAI | `openai` | `Authorization: Bearer $OPENAI_API_KEY` | Remote hosted provider |
| Anthropic | `anthropic` | `x-api-key: $ANTHROPIC_API_KEY` plus `anthropic-version` | Remote hosted provider |
| Google Gemini | `google` | `x-goog-api-key: $GEMINI_API_KEY` | Remote hosted provider |

Secret-handling rules:
- prefer `api_key_env` for hosted providers so secrets stay in env or deployment secret stores
- local engines can use `api_key` directly when a local compatibility server expects a simple static token
- provider reads and browser inventory views never return stored secrets

Official provider references:
- OpenAI API auth: <https://platform.openai.com/docs/api-reference/authentication>
- Anthropic Messages API: <https://docs.anthropic.com/en/api/messages-examples>
- Google Gemini API: <https://ai.google.dev/api/generate-content>
- Ollama OpenAI compatibility: <https://docs.ollama.com/api/openai-compatibility>
- vLLM OpenAI-compatible server: <https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html>

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

AI Organizations also expose an operator-facing output-model routing layer from the organization workspace. That layer does not replace provider configuration; it selects which configured local models are used for delivery types such as general text, research and reasoning, code generation, and vision analysis.

Current self-hosted starting points surfaced in product:
- `Qwen3 8B`
- `Llama 3.1 8B`

## Live Health Probing

`GET /api/v1/cognitive/status` returns real-time health for all providers:

- **Text engines** (vLLM, Ollama, LM Studio): probed via `LLMProvider.Probe()` — checks endpoint reachability
- **Media engine** (Diffusers/SDXL): probed via HTTP GET to the configured endpoint
- Status: `online` / `offline` / `error`

Frontend cognitive-status surfaces can poll this endpoint on a short interval to keep operator-visible engine health current.

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

## Local Model Switching

Default local posture:
- `ollama` is the default shipped provider and profile target
- `vllm` and `lmstudio` stay available but disabled until you opt in

To switch from Ollama to another local engine:
1. start the target engine
2. enable its provider entry in `core/config/cognitive.yaml` or `/settings` -> **AI Engines**
3. confirm the endpoint matches the real server:
   - `ollama` -> `http://127.0.0.1:11434/v1`
   - `vllm` -> `http://127.0.0.1:8000/v1`
   - `lmstudio` -> `http://127.0.0.1:1234/v1`
4. set `model_id` to the served model name
5. re-route the desired profiles from `ollama` to the new provider

For optional repo-local vLLM:
1. `uv run inv install --optional-engines`
2. `uv run inv cognitive.install`
3. `uv run inv cognitive.llm`
4. switch one or more profiles in `cognitive.yaml` from `ollama` to `vllm`

Host support note:
- repo-local `cognitive.*` helpers are intended for supported Linux GPU hosts
- on Windows, keep Ollama as the local default or point the `vllm` provider at a remote OpenAI-compatible vLLM server

For local model changes without touching YAML by hand, you can also use env overrides such as:

```bash
MYCELIS_PROVIDER_VLLM_MODEL_ID=Qwen/Qwen2.5-Coder-7B-Instruct-AWQ
MYCELIS_PROVIDER_VLLM_ENABLED=true
MYCELIS_PROFILE_CODER_PROVIDER=vllm
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
