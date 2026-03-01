# Agent Source Instantiation Template V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-03-01`
Scope: Standardized provider onboarding template for Ollama (default), Claude, Gemini, ChatGPT/OpenAI, vLLM, and LM Studio.

This document defines a single, repeatable way to instantiate and govern agent model sources in Mycelis.

---

## Table of Contents

1. Purpose
2. Canonical Provider Contract
3. Provider-Type Mapping
4. Standard YAML Templates
5. Standard API Templates
6. Validation Checklist
7. Routing Template by Role
8. Security and Governance Baseline
9. Primary Documentation Sources

---

## 1. Purpose

Use one contract for all providers so agent source setup is:
1. repeatable
2. testable
3. policy-governed
4. easy to route by role/profile

No provider-specific ad hoc schema is allowed in runtime configuration.

---

## 2. Canonical Provider Contract

Mycelis `ProviderConfig` normalized fields:

```yaml
providers:
  <provider_id>:
    type: "<openai_compatible|openai|anthropic|google>"
    endpoint: "<base URL or provider endpoint>"
    model_id: "<model name>"
    api_key_env: "<ENV_VAR_NAME>"   # preferred for remote providers
    # api_key: "<inline key>"        # local/dev only; never commit secrets
    location: "<local|remote>"
    data_boundary: "<local_only|leaves_org>"
    usage_policy: "<local_first|require_approval>"
    roles_allowed: ["all" | "<role>", "..."]
    enabled: true
```

Required by policy:
1. `provider_id` lowercase alphanumeric + `-`/`_`
2. explicit `location`, `data_boundary`, `usage_policy`
3. no secret values committed to git

---

## 3. Provider-Type Mapping

| Provider Family | Mycelis `type` | Auth Pattern | Endpoint Pattern |
| :--- | :--- | :--- | :--- |
| Ollama | `openai_compatible` | OpenAI-compatible key/header style (dummy/local key acceptable) | `http://<host>:11434/v1` |
| ChatGPT/OpenAI API | `openai` | `Authorization: Bearer <key>` | `https://api.openai.com/v1` |
| Claude (Anthropic) | `anthropic` | `x-api-key: <key>` + `anthropic-version` header | `https://api.anthropic.com/v1/messages` |
| Gemini (Google AI) | `google` | API key (query param in current adapter) | `https://generativelanguage.googleapis.com/v1beta/models` |
| vLLM server | `openai_compatible` | OpenAI-compatible key/header style | `http://<host>:8000/v1` (typical) |
| LM Studio server | `openai_compatible` | OpenAI-compatible key/header style | `http://<host>:1234/v1` (typical) |

Note:
- Mycelis currently uses OpenAI Chat Completions style for `openai_compatible`.
- OpenAI/Anthropic/Google adapters are implemented in:
  - `core/internal/cognitive/openai.go`
  - `core/internal/cognitive/anthropic.go`
  - `core/internal/cognitive/google.go`

---

## 4. Standard YAML Templates

### 4.1 ChatGPT/OpenAI

```yaml
providers:
  openai_chatgpt:
    type: "openai"
    endpoint: "https://api.openai.com/v1"
    model_id: "gpt-4o-mini"
    api_key_env: "OPENAI_API_KEY"
    location: "remote"
    data_boundary: "leaves_org"
    usage_policy: "require_approval"
    roles_allowed: ["architect", "coder", "admin"]
    enabled: false
```

### 4.2 Ollama (Default Standard)

```yaml
providers:
  ollama:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:11434/v1"
    model_id: "qwen2.5-coder:7b"
    api_key: "ollama"
    location: "local"
    data_boundary: "local_only"
    usage_policy: "local_first"
    roles_allowed: ["all"]
    enabled: true
```

Default policy:
1. Ollama is the primary provider for all roles.
2. Non-Ollama providers are opt-in and explicitly configured per role/profile.
3. Remote providers default to disabled until approved.

### 4.3 Claude (Anthropic)

```yaml
providers:
  anthropic_claude:
    type: "anthropic"
    endpoint: "https://api.anthropic.com/v1/messages"
    model_id: "claude-3-7-sonnet-latest"
    api_key_env: "ANTHROPIC_API_KEY"
    location: "remote"
    data_boundary: "leaves_org"
    usage_policy: "require_approval"
    roles_allowed: ["architect", "sentry", "admin"]
    enabled: false
```

### 4.4 Gemini (Google)

```yaml
providers:
  google_gemini:
    type: "google"
    endpoint: "https://generativelanguage.googleapis.com/v1beta/models"
    model_id: "gemini-1.5-pro"
    api_key_env: "GEMINI_API_KEY"
    location: "remote"
    data_boundary: "leaves_org"
    usage_policy: "require_approval"
    roles_allowed: ["architect", "creative"]
    enabled: false
```

### 4.5 vLLM (OpenAI-Compatible)

```yaml
providers:
  vllm_local:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:8000/v1"
    model_id: "qwen2.5-coder"
    api_key: "mycelis-local"
    location: "local"
    data_boundary: "local_only"
    usage_policy: "local_first"
    roles_allowed: ["all"]
    enabled: true
```

### 4.6 LM Studio (OpenAI-Compatible)

```yaml
providers:
  lmstudio_local:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:1234/v1"
    model_id: "local-model"
    api_key: "lm-studio"
    location: "local"
    data_boundary: "local_only"
    usage_policy: "local_first"
    roles_allowed: ["all"]
    enabled: true
```

---

## 5. Standard API Templates

Use `POST /api/v1/brains` for add and `PUT /api/v1/brains/{id}` for updates.

### 5.1 Add Provider Request

```json
{
  "id": "openai_chatgpt",
  "type": "openai",
  "endpoint": "https://api.openai.com/v1",
  "model_id": "gpt-4o-mini",
  "api_key": "",
  "location": "remote",
  "data_boundary": "leaves_org",
  "usage_policy": "require_approval",
  "roles_allowed": ["architect", "coder", "admin"],
  "enabled": false
}
```

### 5.2 Update Provider Request

```json
{
  "type": "openai_compatible",
  "endpoint": "http://127.0.0.1:8000/v1",
  "model_id": "qwen2.5-coder",
  "api_key": "mycelis-local",
  "location": "local",
  "data_boundary": "local_only",
  "usage_policy": "local_first",
  "roles_allowed": ["all"],
  "enabled": true
}
```

---

## 6. Validation Checklist

Before enabling a provider:
1. endpoint reachable from core runtime
2. model exists and can respond to probe
3. `location/data_boundary/usage_policy` set correctly
4. role scope constrained (`roles_allowed`)
5. provider probe passes (`POST /api/v1/brains/{id}/probe`)
6. cognitive status reflects expected health (`GET /api/v1/cognitive/status`)

---

## 7. Routing Template by Role

Recommended safe baseline (Ollama-first):

```yaml
profiles:
  admin: "ollama"
  architect: "ollama"
  coder: "ollama"
  creative: "ollama"
  sentry: "ollama"
  chat: "ollama"
```

Override rule:
1. keep all roles on `ollama` by default.
2. explicitly reroute only selected roles when capability/quality justification exists.
3. document every override in mission profile or change log.

---

## 8. Security and Governance Baseline

1. local-first defaults for self-hosted providers
2. Ollama remains default source unless configured otherwise
3. remote providers default `enabled: false` until approved
4. no production secret values committed in YAML
5. high-risk roles/actions require approval when remote provider is used
6. maintain Ollama fallback for degraded remote conditions

---

## 9. Primary Documentation Sources

- OpenAI API docs (Responses, auth, model API):
  - https://platform.openai.com/docs/api-reference/responses
  - https://platform.openai.com/docs/quickstart
  - https://platform.openai.com/docs/guides/production-best-practices/api-keys
- Anthropic API docs:
  - https://docs.anthropic.com/en/api/messages
  - https://docs.anthropic.com/en/api/client-sdks
- Gemini API docs:
  - https://ai.google.dev/gemini-api/docs
  - https://ai.google.dev/gemini-api/docs/quickstart
- vLLM OpenAI-compatible serving:
  - https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html
- LM Studio API / OpenAI compatibility:
  - https://lmstudio.ai/docs/app/api
  - https://lmstudio.ai/docs/app/api/endpoints/openai

---

End of document.
