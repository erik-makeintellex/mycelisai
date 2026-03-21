# Secure Gateway and Remote Actuation Profile V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Security controls for OpenClaw-inspired gateway patterns in self-hosted and remote actuation deployments
Hardware companion: `docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md`.

This profile defines the minimum controls required before enabling remote actuation services.
Default mode is self-hosted local execution with remote actuation disabled.

---

## Table of Contents

1. Security Objectives
2. Threat Model
3. Deployment Security Modes
4. Identity, Authentication, and Session Controls
5. Authorization and Scope Model
6. Command Safety Controls for Actuation
7. Transport and Network Controls
8. Secrets and Key Management
9. Audit, Detection, and Recovery
10. API and Envelope Requirements
11. Test and Verification Requirements
12. External Research References
13. Inception Harsh-Truth Security Addendum

---

## 1. Security Objectives

1. Prevent unauthorized remote actuation.
2. Prevent replay, duplication, and stale-command execution.
3. Preserve governance authority for all side effects.
4. Keep self-hosted deployments secure by default.
5. Maintain full forensics (who, what, when, why, outcome).

---

## 2. Threat Model

Primary threats:
- stolen token/session replay
- unauthorized client pairing
- scope escalation at connection time
- command tampering in transit
- remote endpoint impersonation
- duplicated side effects via retry storms
- bypass of governance approval path
- brute-force probing against local control endpoints
- accidental credential discovery from over-broad audits

Actuation-specific threats:
- unsafe command parameterization
- actuator drift due to stale command execution
- control-plane compromise leading to mass command fan-out

---

## 3. Deployment Security Modes

### Mode A: Local-Only (Default)

- Gateway bound to loopback/private network.
- Remote actuation endpoints disabled.
- Only local adapters (`mcp|openapi|python`) with local locality.

### Mode B: Private Remote Mesh (Approved)

- Remote actuation allowed only through private-network or VPN overlays.
- Mutual TLS required between gateway and actuation servers.
- Device registration and pairing approval required.

### Mode C: Public Remote Exposure (Restricted)

- Not allowed by default.
- Requires explicit risk acceptance and hardened boundary controls.
- Additional rate-limits, WAF/edge policy, and formal review required.

---

## 4. Identity, Authentication, and Session Controls

1. Every client and remote actuator has a unique identity (`client_id` / `node_id`).
2. Pairing is explicit and admin-approved before command capability is granted.
3. Session tokens are short-lived and bound to scope + audience.
4. Handshake must include protocol version, role, scopes, and nonce proof.
5. Re-authentication required on scope changes and long-lived sessions.

Mandatory handshake fields:
- `protocol_version`
- `role`
- `requested_scopes[]`
- `nonce`
- `timestamp`
- `client_signature` (or equivalent proof)

---

## 5. Authorization and Scope Model

1. Scopes are least-privilege and action-specific.
2. Remote actuation scopes are split from read/diagnostic scopes.
3. Scope grants are policy-checked at execution time, not only connect time.
4. Unknown scopes are denied by default.

Example scope families:
- `actions.read`
- `actions.invoke.low_risk`
- `actions.invoke.medium_risk`
- `actions.invoke.high_risk` (approval-gated)
- `actuation.remote.invoke`
- `actuation.remote.admin`

---

## 6. Command Safety Controls for Actuation

1. Every side-effecting command carries:
   - `run_id`, `action_id`, `idempotency_key`, `issued_at`, `expires_at`, `nonce`
2. Commands are rejected if expired, replayed, or out-of-sequence.
3. High-risk actions require two-phase flow:
   - proposal/approval
   - execute with approval artifact reference
4. Remote actuator action sets are allowlisted.
5. Dangerous generic execution surfaces (arbitrary shell) remain disabled by default.
6. Self-update operations are staging-only and never permitted as direct production runtime mutations.

Execution policy controls:
- default deny for unknown command classes
- max parameter bounds validation (schema + policy)
- mandatory dry-run support where feasible
- emergency halt / kill-switch path

---

## 7. Transport and Network Controls

1. TLS required for all non-local traffic.
2. mTLS required for gateway <-> remote actuation server links.
3. Certificate pinning or trusted CA constraints required per environment.
4. Network egress allowlist enforced for remote endpoints.
5. Segregate control-plane and telemetry-plane network paths.

Operational controls:
- IP allowlists for actuator services
- rate limits per client and per action
- connection concurrency caps
- circuit-breakers on repeated actuator failures

---

## 8. Secrets and Key Management

1. No secrets in docs manifests, UI payloads, or logs.
2. Service credentials stored encrypted at rest.
3. Key rotation policy required for actuator credentials.
4. Compromised credential revocation must be immediate and auditable.

---

## 9. Audit, Detection, and Recovery

1. Persist all actuation attempts (allow/deny/failed/succeeded).
2. Emit mission events before transport publish for side effects.
3. Alert on repeated deny/replay attempts and scope escalation attempts.
4. Recovery playbooks must include:
   - token/cert revoke
   - client/node quarantine
   - actuator isolation
   - replay-window shrink and rekey

---

## 10. API and Envelope Requirements

All remote actuation requests must include context:

```json
{
  "context": {
    "run_id": "",
    "team_id": "",
    "agent_id": "",
    "origin": "workspace|trigger|schedule|api|sensor",
    "risk_level": "low|medium|high",
    "approval_ref": ""
  },
  "security": {
    "idempotency_key": "",
    "nonce": "",
    "issued_at": "",
    "expires_at": "",
    "scope": ""
  }
}
```

Enforcement order:
1. authenticate
2. authorize scope
3. validate freshness + replay protections
4. validate schema + policy bounds
5. execute and audit

---

## 11. Test and Verification Requirements

### Unit
- scope parser/validator tests
- nonce/idempotency/replay tests
- command TTL and expiry tests

### Integration
- mTLS handshake success/failure tests
- pairing + scope escalation denial tests
- two-phase approval enforcement tests

### API
- unauthorized invoke blocked
- stale command rejected
- duplicate idempotency key behavior validated

### Resilience
- forced token revoke behavior
- remote endpoint impersonation simulation
- mass retry storm and dedupe behavior

Release gate:
- remote actuation cannot be enabled unless all high-risk security tests pass.

---

## 12. External Research References

- OpenClaw architecture and protocol docs:
  - https://openclaw.ai/
  - https://docs.openclaw.ai/reference/architecture
  - https://docs.openclaw.ai/reference/protocols/acp
- A2A Protocol:
  - https://github.com/google/A2A
- OpenAPI Specification:
  - https://spec.openapis.org/oas/latest.html
- JSON Schema:
  - https://json-schema.org/

---

## 13. Inception Harsh-Truth Security Addendum

This addendum encodes early-stage operational lessons for secure autonomy rollout.

1. Sandbox-first execution:
- dynamic skills/adapters run in isolated sandbox contexts by default
- filesystem/network privileges are explicit allowlists, never inherited

2. Single-use-case rollout:
- only one production automation pipeline is permitted during initial enablement
- additional pipelines require passed security + reliability gates

3. Heartbeat autonomy limits:
- long-running loops require bounded budgets (`time`, `actions`, `spend`, `escalations`)
- budget exhaustion must force safe pause + auditable event emission

4. Skill marketplace hardening:
- external skill imports require signed manifest verification and scope review
- unverified skill packages cannot receive medium/high-risk invoke scopes

5. Self-update safety:
- runtime-generated update plans go to staging pipeline only
- production apply requires approval artifact and rollback checkpoint

6. Credential overreach prevention:
- broad credential scans are denied by default policy
- deterministic secret redaction required across logs, traces, and diagnostics exports

7. Incident response trigger:
- any brute-force/probing pattern causes immediate throttle + temporary principal quarantine

---

End of document.
