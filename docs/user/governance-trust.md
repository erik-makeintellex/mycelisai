# Governance & Trust

> How Mycelis ensures human control over agent actions — trust scoring, approval flows, and policy rules.

---

## The Core Principle

Agents in Mycelis do not execute autonomously by default. Every output is tagged with a **trust score**. When that score falls below the configured threshold, the action is **paused for human review** rather than executed automatically.

This is not a safety net for emergencies — it is the normal operating mode. The governance system is designed to be in your loop on decisions that matter.

---

## Trust Scores

Every `CTSEnvelope` (Cognitive Transport Signal) carries a trust score from 0.0 to 1.0:

| Score Range | Source | Meaning |
|-------------|--------|---------|
| `1.0` | Commercial providers (OpenAI, Anthropic, Google) | Fully trusted — auto-execute eligible |
| `0.5` | Local/self-hosted models | Moderate confidence — review recommended |
| Custom | Policy overrides | Per-role or per-action overrides |

The default **auto-execute threshold** is `0.7`. Outputs at or above this score proceed automatically; below it, they are routed to the Approvals queue.

---

## Governance Halt

When an agent produces output with `TrustScore < AutoExecuteThreshold`:

1. The Overseer intercepts the `CTSEnvelope`
2. A `governance_halt` event is emitted via SSE
3. The proposal appears in **Automations → Approvals**
4. Nothing executes until you review it

You'll see a notification in the Workspace chat if a halt occurs during an active session.

---

## Reviewing in Approvals

Navigate to **Automations → Approvals** to see pending proposals.

Each proposal shows:
- **Action description** — plain language summary
- **Agent source** — which agent produced it
- **Trust score** — the score that triggered the halt
- **Full payload** — structured action details (expandable)
- **Risk level** — Low / Medium / High

### Approve

Confirms the proposal and activates it. The action proceeds exactly as specified in the payload. A `mission.started` event is emitted.

### Reject

Discards the proposal. The action is logged as rejected. The requesting agent receives a rejection signal and may generate an alternative proposal.

---

## Policy Configuration

Fine-grained control over what requires approval and what can auto-execute.

Access via **Automations → Approvals → Policy** tab.

### Trust Thresholds

Set per-role thresholds:

```
admin       →  0.5   (local model, requires approval for mutations)
architect   →  0.5
coder       →  0.7   (higher trust for code writes)
creative    →  0.5
sentry      →  0.9   (security decisions always reviewed)
```

Lower threshold = more auto-execution (less oversight).
Higher threshold = more approvals required (more oversight).

### Action Type Overrides

Override the global threshold for specific action types:

| Action Type | Default | Common Override |
|-------------|---------|----------------|
| `read_file` | auto-execute | always auto |
| `write_file` | threshold check | auto if coder ≥ 0.7 |
| `semantic_search` | auto-execute | always auto |
| `store_fact` | threshold check | — |
| Trigger rule creation | always approve | — |
| MCP server install | always approve | — |
| Remote provider enable | always approve | — |

### Risk Groups

Group agents or teams by risk profile and apply policies at the group level:

```yaml
groups:
  - name: "High Risk"
    targets: ["team:finance", "agent:sre-bot"]
    rules:
      - intent: "transfer_funds"
        condition: "amount > 1000"
        action: REQUIRE_APPROVAL
```

---

## Propose vs Execute Modes

This applies specifically to **Trigger Rules** (Automations → Triggers):

| Mode | Behavior |
|------|---------|
| **Propose** | Trigger match creates an approval proposal — you confirm before the mission runs |
| **Execute** | Trigger match fires the mission immediately, no approval step |

The default is always `propose`. Switch to `execute` only for fully automated pipelines where you've verified the trigger logic is correct and the risk is acceptable.

---

## Audit Log

Every governance decision is logged:
- Approval: who approved, timestamp, proposal ID
- Rejection: who rejected, timestamp, reason (if provided)
- Policy change: which rule changed, before/after values

The audit log is accessible from System → Database (Advanced Mode) or via the API at `GET /api/v1/audit`.
