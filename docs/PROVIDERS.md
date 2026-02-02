# 🧠 Cognitive Providers & Adaptive Selection

Mycelis Core uses a **Universal Adapter** system to connect to AI providers.
Crucially, it implements **Adaptive Selection**, where the system grades available resources and matches them to the precise needs of each agent profile.

## The Grading System

The Cortex checks available hosts (Local & Cloud) and grades models into Tiers.
This ensures that if a specific model is unavailable, the system adapts by selecting the next best candidate that meets the *capability requirements*.

| Tier | Class | Generic Capability | Typical Use Case |
| :--- | :--- | :--- | :--- |
| **S** | **Supreme** | High-Reasoning, Large Context, Complex Logic | **The Architect**: System Design, Root Cause Analysis, Refactoring. |
| **A** | **Advanced** | Strong Logic, Code Generation, Medium Speed | **The Coder**: Routine implementation, Unit Tests, API glue. |
| **B** | **Basic** | Balanced Speed/Smarts, Instruction Following | **The Sentry**: Security Analysis, Log parsing, Intent Routing. |
| **C** | **Edge** | Ultra-Fast, Low Latency, Small Footprint | **IoT/Sensor**: High-frequency filtering, Wake-word detection. |

## Gradience Test Values (Tier Assignment)

The system uses the following pattern matching (Regex) to assign Tiers to models discovered at runtime.

| Tier | Pattern (Partial Match) | Examples |
| :--- | :--- | :--- |
| **S** | `gpt-4`, `claude-3-opus`, `gemini-1.5-pro`, `o1` | `gpt-4-turbo`, `claude-3-opus-20240229` |
| **A** | `claude-3-5-sonnet`, `qwen2.5-32b`, `llama-3-70b` | `anthropic/claude-3.5-sonnet` |
| **B** | `qwen`, `llama`, `mistral`, `gemma` | `qwen2.5:7b`, `mistral-nemo` |
| **C** | `phi`, `tiny` | `phi-3-mini` |

*Note: Any unknown model defaults to **Tier B**.*

## Configuration (`brain.yaml`)

You define *Providers* (Sources) and *profiles* (Intents). The system handles the binding.

### 1. Define Providers
Providers represent the "Host" or "Hardware" layer.

```yaml
providers:
  # --- Local Host (Primary) ---
  local_primary:
    type: "openai_compatible"
    endpoint: "http://host.docker.internal:11434/v1"
    model_id: "qwen2.5-coder:7b" # Probed & Graded at runtime
    api_key: "ollama"

  # --- Cloud Utilities (Optional) ---
  cloud_reasoning:
    type: "openai"
    endpoint: "https://api.openai.com/v1"
    model_id: "gpt-4-turbo"      # Graded as Tier S
    api_key_env: "OPENAI_API_KEY"
```

### 2. Auto-Discovery & Defaults
On startup, the Cortex **Probes** these providers.
-   If `cloud_reasoning` fails auth, it is marked **Offline**.
-   The **Architect Profile** (requiring Class S/A) will attempt to find an alternative.
-   If `local_primary` (Tier B) is the only Survivor, the Architect will adapt (Warns user, but proceeds with Tier B).

### 3. Profile Mapping
Profiles express **Need**, not Hardware.

```yaml
profiles:
  architect: 
    preferred: "cloud_reasoning"
    min_tier: "A"
    fallback: "local_primary"
    
  sentry: 
    preferred: "local_primary"
    min_tier: "C"
```

## Precise Adaptability
The goal is **Efficiency**.
-   Do not waste Tier S tokens on log filtering.
-   Do not trust Tier C models with system architecture.
-   **Default to Minimum Viable Host**: The system defaults to the most efficient model that satisfies the Tier requirement.
