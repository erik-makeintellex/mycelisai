# AI Provider Runtime
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [Canonical PRD](architecture-library/MYCELIS_CANONICAL_PRD.md) | [API Reference](API_REFERENCE.md)

This is the scoped implementation contract for model providers, routing, embeddings, and local media gateways. Product and UX authority remains in the Canonical PRD.

## TOC

- [Provider Registry](#provider-registry)
- [Provider Auth Contract](#provider-auth-contract)
- [Profile Routing](#profile-routing)
- [Optional LiteLLM Model Gateway](#optional-litellm-model-gateway)
- [AI Engines UI](#ai-engines-ui)
- [Live Health Probing](#live-health-probing)
- [Configuration File](#configuration-file)
- [Local Model Switching](#local-model-switching)
- [Embedding](#embedding)
- [Hardware Grading](#hardware-grading)

Mycelis supports **multiple self-hosted and commercial inference engines** — configure any combination of vLLM, Ollama, LM Studio, OpenAI, Anthropic, and Google via `cognitive.yaml` or the AI Engines settings surface. An existing LiteLLM proxy may be configured as one optional OpenAI-compatible provider; Mycelis does not embed the LiteLLM Python SDK or ship a proxy service in the current slice.

Media generation stays local-first. Core supports direct Forge/AUTOMATIC1111 generation with `MYCELIS_MEDIA_TYPE=forge` and the native `/sdapi/v1/txt2img` contract, so a Pinokio-hosted Forge instance does not require a second gateway process. Forge must start with API mode enabled. OpenAI-compatible and hosted providers continue to use `/v1/images/generations`; ComfyUI continues through the optional local gateway because its workflow submission and output retrieval require an adapter. Before Soma offers approval for image work, Core checks configured Forge readiness. A running UI with API mode off produces one setup instruction and no doomed proposal/run.

## Provider Registry

| Provider ID | Type | Default Endpoint | Description |
| :--- | :--- | :--- | :--- |
| `vllm` | `openai_compatible` | `http://127.0.0.1:8000/v1` | vLLM inference server — high throughput, GPU-optimized |
| `ollama` | `openai_compatible` | `http://127.0.0.1:11434/v1` | Ollama — local model runner, easy setup |
| `lmstudio` | `openai_compatible` | `http://127.0.0.1:1234/v1` | LM Studio — GUI-based local inference |
| `litellm` | `openai_compatible` | `http://127.0.0.1:4000/v1` | Optional operator-managed LiteLLM proxy; disabled by default |
| `production_gpt4` | `openai` | `https://api.openai.com/v1` | Hosted OpenAI provider; model is configurable and credentials come from `OPENAI_API_KEY` |
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
| LiteLLM proxy | `openai_compatible` | `Authorization: Bearer $LITELLM_PROXY_API_KEY` | Optional gateway only; upstream provider credentials remain gateway-owned |
| OpenAI | `openai` | `Authorization: Bearer $OPENAI_API_KEY` | Remote hosted provider |
| Anthropic | `anthropic` | `x-api-key: $ANTHROPIC_API_KEY` plus `anthropic-version` | Remote hosted provider |
| Google Gemini | `google` | `x-goog-api-key: $GEMINI_API_KEY` | Remote hosted provider |

Secret-handling rules:
- use `api_key_env` whenever a provider enforces a credential so secrets stay in env or deployment secret stores
- the vLLM launcher reads `text.api_key_secret_ref: env:MYCELIS_TEXT_ENGINE_API_KEY` from `cognitive/config/engine.yaml`, resolves the named value from the shell and then the repo-local `.env`, and fails when it is unavailable; it has no committed credential or hardcoded fallback
- local compatibility engines that ignore authentication may retain an explicitly non-secret placeholder client value
- LiteLLM uses `api_key_env: LITELLM_PROXY_API_KEY`; never place a proxy master key, virtual key, or upstream provider credential in committed provider YAML
- provider reads and browser inventory views never return stored secrets

Official provider references:
- OpenAI API auth: <https://platform.openai.com/docs/api-reference/authentication>
- Anthropic Messages API: <https://docs.anthropic.com/en/api/messages-examples>
- Google Gemini API: <https://ai.google.dev/api/generate-content>
- Ollama OpenAI compatibility: <https://docs.ollama.com/api/openai-compatibility>
- vLLM OpenAI-compatible server: <https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html>
- LiteLLM gateway and SDK boundary: <https://docs.litellm.ai/> and <https://github.com/BerriAI/litellm>

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

## Optional LiteLLM Model Gateway

LiteLLM belongs between Core's cognitive router and model providers. It is not an agent orchestrator and is unrelated to the `framework_runs` execution selector:

```text
Soma / Mycelis Core
  -> provider policy, profile/team eligibility, data boundary, approval, Outcome budget
  -> optional LiteLLM proxy
  -> approved local or hosted model deployment
```

Ownership is deliberately split:

- Mycelis Core owns provider and profile selection, role eligibility, `local_only` versus `leaves_org`, semantic and Outcome budgets, approval, run/team correlation, audit, and proof.
- LiteLLM may own provider-protocol translation, deployment-held upstream credentials, operational RPM/TPM and spend ceilings, and retries or load balancing among equivalent deployments already admitted to the same boundary.
- A LiteLLM alias or fallback pool must never cross from local-only to remote inference. Core selects a policy-bounded alias; the gateway may vary only inside that alias's approved boundary.
- Gateway usage and cost records are evidence for reconciliation. They do not approve work, complete an Outcome, or replace Core mission events.
- Prompt/response logging, callbacks, caching, and persistence remain disabled until retention, redaction, tenancy, and recovery are certified. Browser clients and framework workers never receive gateway administration or upstream secrets.

The first slice is conformance-only: the disabled provider proves OpenAI-compatible requests, tool-call normalization, configured provider identity, actual gateway-reported model identity, and exact reported token usage. It does not install or start LiteLLM. Production enablement additionally requires an authenticated proxy deployment, dedicated persistence and distributed rate-limit posture where used, scoped correlation, cost reconciliation, failure-mode proof, and operations certification.

## AI Engines UI

Navigate to `/settings` → **AI Engines** (Advanced mode):

- Click a **profile** to change which provider it routes to
- Click a **provider** to configure endpoint, model ID, and API keys
- Changes persist to `cognitive.yaml` via `PUT /api/v1/cognitive/profiles` and `PUT /api/v1/cognitive/providers/{id}`

Compatibility organization contexts also expose an output-model routing layer. That layer does not replace provider configuration; it selects which configured local models are used for delivery types such as general text, research and reasoning, code generation, and vision analysis.

Current self-hosted starting points surfaced in product:
- `Qwen3 8B`
- `Llama 3.1 8B`

## Live Health Probing

`GET /api/v1/cognitive/status` returns real-time health for all providers:

- **Text engines** (vLLM, Ollama, LM Studio): enabled providers are probed via `LLMProvider.Probe()` to check endpoint reachability; disabled providers remain in configuration without contributing latency or online status
- **Media engines**: local/self-hosted endpoints are probed via HTTP GET to the configured endpoint and returned with typed provider metadata (`provider_id`, `type`, `location`, `data_boundary`, `usage_policy`, `enabled`) so local/self-hosted and hosted providers stay explicit in status output
- Hosted media providers report `configured` once endpoint/model/credentials posture is expressible; live upstream errors are handled during generation so APIs without a local-style `/health` endpoint are not mislabeled as offline
- Status: `online` / `offline` / `configured` / `disabled` / `error`

Frontend cognitive-status surfaces can poll this endpoint on a short interval to keep operator-visible engine health current.

## Configuration File

`core/config/cognitive.yaml`:

```yaml
providers:
  vllm:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:8000/v1"
    model_id: "qwen2.5-coder"
    api_key: ""
    api_key_env: "MYCELIS_TEXT_ENGINE_API_KEY"
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

  litellm:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:4000/v1"
    model_id: "mycelis-default"
    api_key: ""
    api_key_env: "LITELLM_PROXY_API_KEY"
    location: "remote"
    data_boundary: "leaves_org"
    usage_policy: "require_approval"
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
  provider:
    provider_id: "media-local"
    type: "openai_compatible"
    location: "local"
    data_boundary: "local_only"
    usage_policy: "local_first"
    api_key_env: ""
    enabled: true
  endpoint: "http://127.0.0.1:8001/v1"
  model_id: "local-media"
```

Set `MYCELIS_TEXT_ENGINE_API_KEY` in the shell or repo-local `.env` before starting vLLM through `uv run inv cognitive.llm` or `uv run inv cognitive.up`. Core's `vllm` provider resolves the same variable named by `api_key_env`, so the launcher and client cannot drift to separate committed credentials.

The media block keeps a legacy `endpoint` and `model_id` surface for compatibility, but the typed `media.provider` record is the preferred contract when you want local-hosted or hosted media behavior to be explicit.

Pinokio local media profile:

```env
MYCELIS_MEDIA_PROVIDER_ID=pinokio-local
MYCELIS_MEDIA_TYPE=forge
MYCELIS_MEDIA_ENDPOINT=http://127.0.0.1:7860
MYCELIS_MEDIA_MODEL_ID=forge-local
MYCELIS_MEDIA_LOCATION=local
MYCELIS_MEDIA_DATA_BOUNDARY=local_only
MYCELIS_MEDIA_USAGE_POLICY=local_first
MYCELIS_MEDIA_ENABLED=true
```

Start Forge through Pinokio with API mode enabled and verify `http://127.0.0.1:7860/sdapi/v1/options` returns success. The ordinary Forge UI at `/` is not sufficient proof; when the UI is open but this route returns `404`, Soma explains the missing API mode before asking for approval.

ComfyUI local media profile:

```env
MYCELIS_MEDIA_PROVIDER_ID=pinokio-comfyui-local
MYCELIS_MEDIA_TYPE=openai_compatible
MYCELIS_MEDIA_ENDPOINT=http://127.0.0.1:8001/v1
MYCELIS_MEDIA_MODEL_ID=local-media
MYCELIS_MEDIA_LOCATION=local
MYCELIS_MEDIA_DATA_BOUNDARY=local_only
MYCELIS_MEDIA_USAGE_POLICY=local_first
MYCELIS_MEDIA_ENABLED=true
MYCELIS_MEDIA_GATEWAY_BACKEND=comfyui
MYCELIS_MEDIA_GATEWAY_UPSTREAM=http://127.0.0.1:8188
MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE=./workspace/media/comfyui-workflow.json
MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_NODE_ID=6
MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_INPUT=text
MYCELIS_MEDIA_GATEWAY_COMFY_SIZE_NODE_ID=5
MYCELIS_MEDIA_GATEWAY_COMFY_WIDTH_INPUT=width
MYCELIS_MEDIA_GATEWAY_COMFY_HEIGHT_INPUT=height
MYCELIS_MEDIA_GATEWAY_COMFY_BATCH_NODE_ID=5
MYCELIS_MEDIA_GATEWAY_COMFY_BATCH_INPUT=batch_size
MYCELIS_MEDIA_GATEWAY_COMFY_POLL_SECONDS=1
MYCELIS_MEDIA_GATEWAY_ALLOW_PUBLIC_UPSTREAM=0
```

ComfyUI setup steps:

1. Start ComfyUI locally through Pinokio and confirm the local API is reachable at `http://127.0.0.1:8188`.
2. In ComfyUI, build the desired image workflow and export/save it in API format.
3. Store that workflow under the governed workspace, for example `core/workspace/media/comfyui-workflow.json`.
4. Set `MYCELIS_MEDIA_GATEWAY_BACKEND=comfyui`, point `MYCELIS_MEDIA_GATEWAY_UPSTREAM` at the local ComfyUI endpoint, and map the prompt node through `MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_NODE_ID`.
5. If the workflow has a latent or image-size node, set `MYCELIS_MEDIA_GATEWAY_COMFY_SIZE_NODE_ID` plus width/height input names so Soma's requested image size transmits into the workflow.
6. Start `uv run inv cognitive.media-gateway`; Core continues to call `http://127.0.0.1:8001/v1/images/generations`.
7. Ask Soma to generate an image. The gateway submits the workflow to ComfyUI `/prompt`, polls `/history/{prompt_id}`, fetches generated files through `/view`, and returns `b64_json` to Core so outputs can be retained as Mycelis artifacts.

Use `MYCELIS_MEDIA_GATEWAY_ALLOW_PUBLIC_UPSTREAM=1` only when intentionally routing the private media gateway to a reviewed non-private endpoint. Normal local Pinokio proof should use `localhost`, `host.docker.internal`, loopback, or private LAN IP upstreams.

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
- **Memory tools:** `remember` writes scoped durable memory records and recall metadata, while semantic search surfaces candidate memories across Soma/team/governed lanes using scope-aware retrieval
- **Trusted recall boundary:** semantic relevance is not final authority by itself; governed doctrine and deterministic evidence outrank lower-order recalled memory when they conflict
- **Promotion boundary:** automatic planning continuity stays in temporary continuity channels; durable lessons and reflection should cross the `LearningCandidate` boundary before governed promotion into long-term recall
- **Fallback chain:** `Router.Embed()` tries each provider that implements `EmbedProvider`

## Hardware Grading

| Tier | RAM | Supported Models | Use Case |
| :--- | :--- | :--- | :--- |
| **Tier 1 (Min)** | 16 GB | 7B Models (Q4) | Basic Coding, CLI |
| **Tier 2 (Rec)** | 32 GB | 14B - 32B Models | Complex Architecture, Deep Reasoning |
| **Tier 3 (Ultra)** | 64 GB+ | 70B+ or Multi-Model | **Enterprise Core** (Current Dev Host) |

The system auto-detects resources but defaults to the 7B model for speed.
