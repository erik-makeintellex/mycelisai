# Agentry Template Marketplace + Custom Templating V7
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Overview](OVERVIEW.md) | [API Reference](../API_REFERENCE.md)

Version: `1.0`
Status: `Authoritative Planning`
Last Updated: `2026-03-01`
Scope: API-first support for external template marketplaces, private hubs, and tenant/user-owned custom templates.

## Purpose

Mycelis supports local catalogue templates. Enterprise deployment also needs governed marketplace ingestion, purchase/install lifecycle, tenant-specific custom authoring, and one normalized template contract across all sources.

Goals:
- teams instantiate proven templates quickly
- operators buy, import, fork, and publish templates safely
- organizations can govern private and marketplace templates through the same API posture

## Source Model

Every template declares a source class:
- `builtin`: shipped with Mycelis
- `marketplace`: external provider feed, such as ClawHub-style feeds
- `private_hub`: org-hosted registry
- `tenant_custom`: user/org-authored template

Required source metadata:
- `source_id`
- `source_type`
- `publisher`
- `namespace`
- `template_ref`
- `version`
- `signature_status` (`verified|unverified|failed`)

## Canonical Package Contract

All sources normalize into one package shape before exposure:

```json
{
  "template_id": "tpl_research_swarm_v2",
  "source": {
    "source_type": "marketplace",
    "source_id": "clawhub",
    "publisher": "trusted-labs",
    "namespace": "operations",
    "template_ref": "ops/research-swarm"
  },
  "version": "2.1.0",
  "display_name": "Research Swarm",
  "description": "Multi-agent research and synthesis workflow",
  "category": "research",
  "tags": ["research", "swarm", "analysis"],
  "runtime_requirements": {
    "providers": ["ollama-local"],
    "mcp_services": ["filesystem", "fetch"],
    "action_services": ["web.search"]
  },
  "parameters_schema": {},
  "blueprint_spec": {},
  "pricing": {
    "model": "free|one_time|subscription",
    "currency": "USD",
    "amount": 49.0
  },
  "license": {
    "license_id": "lic_standard_v1",
    "allow_edit": true,
    "allow_redistribute": false
  },
  "integrity": {
    "digest": "sha256:...",
    "signature": "...",
    "verified": true
  }
}
```

Rules:
- `parameters_schema` and `blueprint_spec` are required.
- Marketplace packages must include integrity metadata.
- Installed templates are immutable by default; edits create a forked custom template.
- Paid marketplace forks retain source and license attribution, and license-restricted edits are blocked at publish.

## API Contracts

The canonical endpoint list is mirrored in [API Reference](../API_REFERENCE.md). This document owns the governance meaning.

Marketplace source management:
- `GET /api/v1/template-market/sources`
- `POST /api/v1/template-market/sources`
- `PATCH /api/v1/template-market/sources/{source_id}`
- `POST /api/v1/template-market/sources/{source_id}/probe`

Marketplace discovery, acquisition, and install:
- `GET /api/v1/template-market/templates`
- `GET /api/v1/template-market/templates/{template_id}`
- `POST /api/v1/template-market/templates/{template_id}/purchase-intent`
- `POST /api/v1/template-market/purchases/{purchase_id}/confirm`
- `POST /api/v1/template-market/templates/{template_id}/install`
- `POST /api/v1/template-market/templates/{template_id}/sync`
- `GET /api/v1/template-market/installs`
- `POST /api/v1/template-market/installs/{install_id}/upgrade`
- `DELETE /api/v1/template-market/installs/{install_id}`

Tenant custom templates:
- `GET /api/v1/templates/custom`
- `POST /api/v1/templates/custom`
- `GET /api/v1/templates/custom/{template_id}`
- `PUT /api/v1/templates/custom/{template_id}`
- `POST /api/v1/templates/custom/{template_id}/publish`
- `POST /api/v1/templates/custom/{template_id}/fork`
- `DELETE /api/v1/templates/custom/{template_id}`

## Purchase, Licensing, And Entitlements

Purchase states:
- `draft`
- `pending_approval`
- `approved`
- `completed`
- `failed`
- `revoked`

Each paid install binds an entitlement record with:
- `tenant_id`
- `template_id`
- `license_id`
- `entitlement_scope` (`tenant|seat|environment`)
- `status`
- `expires_at` for subscriptions

Enforcement:
- paid template install requires an active entitlement
- upgrade requires license compatibility
- expiring subscriptions surface runtime warnings
- purchase, install, upgrade, and delete are governed mutations with deterministic audit trail

## Governance And Security

Mandatory controls:
- source allowlist by default
- signed package verification for marketplace sources
- secret fields never returned in API payloads
- template sandbox checks before activation
- compatibility validation for providers, MCP/action dependencies, and policy constraints
- ClawHub-style connectors support token auth, catalog pull, package verification, and normalization before exposure

## Event Model

Required event types:
- `template.source.registered`
- `template.market.listed`
- `template.purchase.proposed`
- `template.purchase.completed`
- `template.installed`
- `template.upgraded`
- `template.custom.published`
- `template.entitlement.expiring`

Recommended subjects:
- `swarm.marketplace.template.*`
- `swarm.catalog.template.*`
- `swarm.audit.template.*`

Every event includes:
- `tenant_id`
- `user_id`
- `run_id` when workflow-driven
- `template_id`
- `source_type`

## UI And Workflow Requirements

Resources -> Marketplace supports source filters, pricing and entitlement badges, compatibility results before install, and governed purchase proposals.

Template Builder supports schema-aware editing, draft/publish workflow, version history, rollback, and fork-from-installed.

Team instantiation shows source, license, required capabilities, missing dependencies, and safe remediation paths where available.

## Testing And Release Gates

Required tests:
- package normalization
- license and entitlement validation
- source auth and signature verification
- source registration, probe, sync, purchase, approval, install, upgrade, uninstall, draft, publish, and fork paths
- E2E buy/install, custom publish/instantiate, entitlement expiry, and degraded marketplace recovery

Release gate:
- no template-market production enablement until purchase/install paths are fully audited and reversible

## Rollout

1. Source registration and marketplace discovery APIs with read-only listing UI.
2. Governed purchase intent, entitlement model, install, and uninstall.
3. Custom template builder, publish/fork, upgrade path, and compatibility checks.
4. Multi-source federation with full observability and policy hardening.
