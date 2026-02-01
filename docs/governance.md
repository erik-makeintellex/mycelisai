# Governance & Policy System

> 🔙 **Navigation**: [Back to README](../README.md) | [Architecture Overview](../architecture.md)

The Mycelis Core includes a **Governance Gatekeeper** that intercepts every message passing through the Swarm Router. It enforces a logic-based policy defined in `core/policy/policy.yaml`.

## 1. How It Works

1.  **Intercept**: The Router passes every `MsgEnvelope` to the Gatekeeper before routing.
2.  **Evaluate**: The Gatekeeper checks the message against the loaded Policy rules.
    -   **Matches Team/Agent**: Does the sender belong to a targeted group?
    -   **Matches Intent**: Does the message `intent` (or event type) match a rule?
    -   **Unlocks Condition**: Do the context variables satisfy the logic (e.g., `amount > 50`)?
3.  **Decide**:
    -   `ALLOW`: The message proceeds normally.
    -   `DENY`: The message is dropped and logged.
    -   `REQUIRE_APPROVAL`: The message is **Parked** in memory. An `ApprovalRequest` is generated. It will not be delivered until an admin signal approves it.

## 2. Policy Configuration (`policy.yaml`)

Policies are grouped by "Risk Profile".

```yaml
groups:
  - name: "High Risk"
    targets: ["team:finance", "agent:sre-bot"]
    rules:
      - intent: "transfer_funds"
        condition: "amount > 1000"
        action: "REQUIRE_APPROVAL"
```

### Available Actions
- `ALLOW`
- `DENY`
- `REQUIRE_APPROVAL`

### Conditions
Simple comparison logic is supported for context variables (extracted from `swarm_context` or event data).
- Operators: `>`, `<`, `==`, `>=` `...`
- Example: `temperature > 100`, `role == "admin"`

---

## 3. Ideal Configurations

### Scenario A: High-Velocity IoT (Sensor Networks)
**Goal**: Maximize throughput, minimize latency. Only block dangerous actuators.

**Recommended Policy**:
- **Defaults**: `ALLOW` (Assume telemetry is safe).
- **Restrictions**: Only guard physical control signals.

```yaml
groups:
  - name: "IoT Actuators"
    description: "Prevent hardware damage"
    targets: ["team:robotics"]
    rules:
      - intent: "motor.set_speed"
        condition: "rpm > 8000"
        action: "DENY" 
      - intent: "firmware.update"
        action: "REQUIRE_APPROVAL"

defaults:
  default_action: "ALLOW"
```

### Scenario B: High-Stakes Agents (Message Parsing & Finance)
**Goal**: Safety & Human-in-the-loop. Prevent hallucinations from executing expensive actions.

**Recommended Policy**:
- **Defaults**: `ALLOW` (for harmless "thinking" logs), but strictly guard "tool usage".
- **Restrictions**: Gate interactions with external APIs or databases.

```yaml
groups:
  - name: "Autonomous Agents"
    description: "Guard against hallucinated tool calls"
    targets: ["team:finance", "team:devops"]
    rules:
      - intent: "stripe.charge"
        condition: "amount > 50"
        action: "REQUIRE_APPROVAL"
      - intent: "k8s.delete_pod"
        action: "REQUIRE_APPROVAL"
      - intent: "k8s.scale"
        condition: "replicas > 10"
        action: "REQUIRE_APPROVAL"

# Allow chatter, block action by default if unknown intent?
# For development, usually ALLOW default.
defaults:
  default_action: "ALLOW"
```
