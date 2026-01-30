# Gatekeeper Logic (Governance Middleware)

## Overview
The Gatekeeper is a middleware component in the Neural Core that intercepts messages before they reach the main Router. It enforces the rules defined in `core/policy/policy.yaml`.

## Architecture

### 1. The Intercept Pipeline
```go
func (g *Gatekeeper) Intercept(msg *pb.MsgEnvelope) bool {
    // 1. Identification
    team := msg.TeamId
    agent := msg.SourceAgentId
    intent := msg.Payload.Event.EventType // Simplified extraction

    // 2. Policy Evaluation
    action := g.PolicyEngine.Evaluate(team, agent, intent, msg.Data)

    switch action {
    case ALLOW:
        return true // Proceed to Router
    case DENY:
        g.Logger.Warn("Blocked action", "intent", intent, "agent", agent)
        return false // Drop
    case REQUIRE_APPROVAL:
        g.SuspendMessage(msg)
        return false // Drop from Router, hold in Pending
    }
}
```

### 2. Suspension & Alerting
When `REQUIRE_APPROVAL` is triggered:
1.  **Wrap**: Create an `ApprovalRequest` containing the original `MsgEnvelope` and the `reason`.
2.  **Store**: Save to `PendingBuffer` (InMemory for V1, Redis for V2) keyed by `request_id`.
3.  **Alert**: Publish `swarm.governance.request` so the "User Team" (or Admin UI) sees it.

### 3. Approval Signal Handling
The Core subscribes to `swarm.governance.signal`.
1.  **Validate**: Check `user_signature` (if enabled).
2.  **Resume**:
    *   If `approved == true`: Look up `request_id` in `PendingBuffer`. If found, Inject back into the **Router** (bypassing Gatekeeper to avoid loops, or re-evaluating with "Approved" context).
    *   If `approved == false`: Delete from Buffer.

## Data Structures

### Policy Config (YAML)
Loaded into:
```go
type PolicyRule struct {
    Intent    string
    Condition string // "amount > 50" (Parsed via simple expression engine)
    Action    string // "ALLOW", "DENY", "REQUIRE_APPROVAL"
}
```

### Pending Buffer
```go
type PendingBuffer struct {
    Reqs map[string]*pb.ApprovalRequest
    Mu   sync.RWMutex
}
```
