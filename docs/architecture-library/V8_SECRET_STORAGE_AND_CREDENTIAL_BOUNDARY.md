# V8 Secret Storage And Credential Boundary
> Navigation: [Project README](../../README.md) | [Architecture Docs Index](ARCHITECTURE_LIBRARY_INDEX.md) | [API Reference](../API_REFERENCE.md)

> Status: Canonical
> Last Updated: 2026-05-21
> Purpose: Define how Mycelis stores, references, resolves, audits, and rotates credentials.

## Purpose

Secrets are runtime-bound trust infrastructure. They are not product data, operator output, memory, proof content, UI state, or documentation content.

Mycelis should let an operator understand which capabilities are configured without exposing the credential values that make those capabilities work.

## Current Storage Contract

Local source development:
- raw secret values live only in repo-local `.env`
- `.env` is ignored and must never be committed
- `.env.example` may name required variables, but only with placeholders
- `.env.compose` and `.env.compose.example` are topology/config-shape files, not secret stores

Container and Kubernetes development:
- Compose reads secret values from `.env`
- Helm/Kubernetes deployments must inject secrets through Kubernetes Secret values or external secret-manager refs
- automation must avoid passing secret values as command-line args when a file/env/secret-backend path is practical
- rendered manifests, release artifacts, and proof bundles must not contain raw secret values

Enterprise target:
- external secret managers may own the physical values
- Mycelis stores only a secret reference, allowed consumer scope, readiness state, and audit metadata
- UI writes that create or rotate a secret must return only the reference and configured/missing posture

## SecretRef Shape

The durable target object is `SecretRef`. It is a reference to a secret, not the secret itself.

Required fields:
- `secret_ref`: stable reference id such as `env:OPENAI_API_KEY`, `k8s:mycelis-core/core-api-key`, or `vault:path/key`
- `backend`: `env`, `kubernetes_secret`, `external_secret_manager`, or `local_dev`
- `scope`: deployment, organization, capability, provider, auth provider, or MCP server
- `owner`: admin, system, deployment, or managed provider
- `risk_class`: low, medium, high, or critical
- `allowed_consumers`: runtime components allowed to resolve the value
- `configured`: whether the reference currently resolves
- `last_checked_at`: last readiness probe timestamp
- `rotation_due_at`: optional rotation target
- `status`: configured, missing, degraded, expired, rotating, or revoked

Raw values are never stored in Mycelis DB rows, docs, `.state`, API responses, logs, browser state, artifacts, proof payloads, or managed exchange items.

## Runtime Resolution

Core, Interface, and worker runtimes may resolve a secret into memory only when executing the capability or auth flow that requires it.

Resolution rules:
- API reads return `configured`, `status`, and safe reference metadata only
- provider config uses `api_key_env` now and should add `secret_ref` where persistence needs backend neutrality
- raw `api_key` values are rejected by provider-management APIs
- MCP manifests declare required env vars and secret requirements; install/review surfaces show requirements, not values
- capability manifests expose `secret_ref_policy` and readiness posture, not credentials
- outputs and proof artifacts may link to the capability and secret reference class, but never to the value

## UI Contract

Operator surfaces should answer:
- which capability or provider needs a secret
- whether the secret is configured
- which backend/reference is used
- what role may change it
- what recovery or rotation action is available

Operator surfaces should not:
- display raw secret values
- preserve typed secret values in browser state longer than a submit interaction
- show secrets in copied curl commands, logs, proof packages, or screenshots
- imply a provider is healthy when its secret reference is missing or failing readiness

If a future UI accepts direct secret entry, it must submit to a secret-write endpoint that stores the value in the configured backend and returns only a `SecretRef`.

## Audit And Proof

Audit events for secret-adjacent operations should record:
- actor identity
- action: configure, rotate, revoke, probe, or use
- safe `secret_ref`
- capability/provider/auth-provider id
- approval id when governed
- success/failure state
- degraded posture and next recovery action

Proof artifacts should record that credential readiness was validated when relevant, but they must not embed values or provider tokens.

## Rotation And Incident Response

Rotate immediately when:
- a secret was pasted into chat, docs, logs, tickets, screenshots, or committed files
- a provider account changes ownership
- a staff, operator, or service-account access boundary changes
- a capability moves from local development to shared or enterprise deployment

Near-term rotation behavior:
- local `.env` rotation requires updating `.env` and restarting the affected runtime
- Kubernetes rotation should update the Secret/external secret and restart or reload dependent pods
- audits should record the rotation event and the previous reference status, not old values

Future behavior should support dual-key overlap, runtime reload, expiry warnings, and confidence/proof degradation when a credential goes missing mid-run.

## Required Implementation Direction

`REQUIRED` provider and auth config objects should accept `secret_ref` in addition to current env-var references.

`REQUIRED` Settings, Resources, Auth Providers, MCP, and AI Engines should use a consistent secret-readiness component.

`REQUIRED` release and CI proof should include a secret scan gate for committed files and rendered deployment artifacts.

`NEXT` Kubernetes deployment automation should continue removing secret values from process args and rendered logs.

`NEXT` Mycelis API self-use should create a low-cost hosted-provider proof run only when the local `.env` reference resolves, and should mark missing secrets as degraded setup truth rather than runtime failure.
