# Cortex Memory (Logging) Schema

> 🔙 **Navigation**: [Back to README](../README.md) | [Architecture Overview](../architecture.md)

The "Cortex" is the centralized memory stream of the Mycelis Swarm. All actions, thoughts, and events are normalized into a strict `LogEntry` format and published to the `cortex.logs` NATS subject.

## 1. The LogEntry Protocol

**Protobuf Definition**: `proto/swarm/v1/swarm.proto`

Every entry in the log stream follows this structure. This ensures that agents (like the Verifier or SRE-Bot) can deterministically parse history without guessing formats.

| Field | Type | Description |
| :--- | :--- | :--- |
| **`timestamp`** | `Timestamp` | Precise UTC time of the event. |
| **`trace_id`** | `string` | The distributed trace ID (Correlation ID). Critical for debugging huge workflows. |
| **`span_id`** | `string` | The specific step/span within the trace. |
| **`agent_id`** | `string` | The specific ACTOR (e.g., `process-01`, `gpt-4-01`). |
| **`team_id`** | `string` | The logic GROUP (e.g., `marketing`, `sensors`). |
| **`source_uri`** | `string` | The CAPABILITY signature (e.g., `swarm:model:llama3`, `swarm:driver:rpi`). |
| **`intent`** | `string` | The high-level GOAL or classification (e.g., `process_image`, `error`, `startup`). |
| **`context_snapshot`** | `Struct` | A copy of the Global State (`swarm_context`) at the exact moment of the event. |
| **`status`** | `string` | `SUCCESS`, `FAILURE`, or `PENDING`. |
| **`error_message`** | `string` | Human-readable details if status is FAILURE. |
| **`raw_payload`** | `Any` | The full original message or tool result (for replayability). |

## 2. Parsing Guide (for Agents)

When an agent needs to "read memory" (e.g., "What did we do last week?"):

1.  **Filter by `trace_id`** to reconstruct a single thought process.
2.  **Filter by `intent`** to find specific actions (e.g., `intent="tool_call"`) without noise.
3.  **Inspect `context_snapshot`** to understand *why* a decision was made (the state of the world at that time).

## 3. JSON Example

When consumed via the API or CLI, the Protobuf is serialized to JSON:

```json
{
  "timestamp": "2026-02-01T20:00:00Z",
  "trace_id": "corr_12345",
  "agent_id": "payment-bot-01",
  "team_id": "finance",
  "intent": "charge_card",
  "status": "SUCCESS",
  "context_snapshot": {
    "user_tier": "gold",
    "risk_score": 0.1
  },
  "raw_payload": { ... }
}
```

## 4. Best Practices

-   **Don't log noise.** If it's a debug print, keep it local. If it affects state, send it to Cortex.
-   **Always include `trace_id`.** An orphan log is a lost memory.
-   **Use broad `intents`.** Don't use `started_processing_file_a`, use `intent="processing" file="a"`.
