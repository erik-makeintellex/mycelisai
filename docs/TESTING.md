# Verification & Testing Protocol

Mycelis employs a rigorous **2-Tier Testing Strategy** to ensure cognitive reliability without sacrificing development velocity.

## Tier 1: Cognitive Unit Tests (Mocked)

**Goal:** Verify internal logic (CQA Contract enforcement) immediately.
**Speed:** < 100ms.

These tests use **Mock Providers** to simulate various LLM failure modes (timeouts, hallucinations, network errors) to ensure the `Middleware` and `Router` handle them correctly.

### 📍 Location

- `core/internal/cognitive/middleware_test.go`

### 🧪 What is Tested?

- **Retries:** Does the system retry exactly `MaxRetries` times?
- **Schema Validation:** Does it reject invalid JSON and accept valid JSON?
- **Timeouts:** Does it respect the `TimeoutMs` budget?
- **Self-Correction:** (Future) Does it inject error context into the next prompt?

### ⚡ Running Unit Tests

```bash
# Run from repository root
cd core
go test -v ./internal/cognitive/...
```

---

## Tier 2: Agent Interaction Tests (Integration)

**Goal:** Verify the **Real Model** (Ollama/OpenAI) understands our prompts and schemas.
**Speed:** 1s - 30s (depends on model).

These tests connect to the **Primary System Agent Service Source Model** (e.g., `qwen2.5-coder` running on Ollama). They verify that the actual Intelligence Layer is functional and reachable.

### 📍 Location

- `core/tests/agent_interaction_test.go`
- **Build Tag:** `//go:build integration` (Ignored by default `go test`)

### 🧪 What is Tested?

- **Connectivity:** Can we reach `http://localhost:11434`?
- **Model Capabilities:** Does `qwen2.5-coder` actually output valid JSON when asked?
- **Output Schema:** Do real responses pass the `types.go` validation logic?

### ⚡ Running Integration Tests

**Requirements:**

- Local Ollama running (`inv k8s.bridge` or local install).
- Model Pulled: `ollama pull qwen2.5-coder:7b-instruct`

```bash
# Run from repository root
# Note: Ensure OLLAMA_HOST is valid (default: http://localhost:11434)
# Windows (PowerShell)
$env:OLLAMA_HOST="http://localhost:11434"; go test -v -tags=integration ./tests/...

# Linux/Mac
OLLAMA_HOST=http://localhost:11434 go test -v -tags=integration ./tests/...
```

---

## 🛠️ Adding New Tests

### Adding a Unit Test

1. Open `middleware_test.go`.
2. Create a `MockProvider` scenario:

   ```go
   mock := &MockProvider{
       ShouldFailCount: 1, 
       OutputSequence: []string{"Valid Response"}
   }
   ```

3. Register it with `r.Providers["mock"] = mock`.
4. Run assertion.

### Adding an Integration Test

1. Open `core/tests/agent_interaction_test.go`.
2. Define a function `TestAgentInteraction_<Scenario>(t *testing.T)`.
3. Load the real config: `cognitive.NewRouter("../config/brain.yaml")`.
4. Define an `InferRequest` with a specific `Profile`.
5. Assert the `resp.Text` matches expectations.
78:
79: ---
80:
81: ## Tier 3: Governance Smoke Tests (System)
82: **Goal:** Verify that the **Gatekeeper** is actively blocking dangerous actions.
83: **Context:** Uses the CLI or NATS to inject "poison pill" messages.
84:
85: ### 📍 Protocol
86: 1.  **Start the Core:** `inv core.run` (or separate shell).
87: 2.  **Inject Poison:** Send a message with intent `k8s.delete.pod`.
88: 3.  **Verify Block:** Check logs for "⛔ Gatekeeper DENIED".
89: 4.  **Inject Require Approval:** Send `payment.create` with amount `100`.
90: 5.  **Verify Inbox:** Check `/admin/approvals` for the pending request.
91:
92: ### ⚡ Running Smoke Tests
93: ```bash
94: # Run python script
95: uv run python scripts/verify_governance.py
96:```
