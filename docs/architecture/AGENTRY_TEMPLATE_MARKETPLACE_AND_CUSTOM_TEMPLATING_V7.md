# Agentry Template Marketplace + Custom Templating V7

Version: `1.0`
Status: `Authoritative Planning`
Last Updated: `2026-03-01`
Scope: API-first support for external template marketplaces (ClawHub-style) and tenant/user-owned custom templates

---

## Table of Contents

1. Why This Exists
2. Template Source Model
3. Canonical Template Package Contract
4. Marketplace API Contracts
5. Custom Template API Contracts
6. Purchase, Licensing, and Entitlements
7. Governance + Security Controls
8. NATS Event Model
9. UI/Workflow Requirements
10. Testing and Release Gates
11. Rollout Plan

---

## 1. Why This Exists

Mycelis already supports local catalogue templates.
Enterprise deployment now requires:

1. external marketplace ingestion (for example, ClawHub-style template feeds)
2. controlled purchase/install lifecycle
3. tenant-specific custom template authoring and private distribution
4. one consistent API contract across marketplace and custom sources

Goal:
- teams can instantiate proven templates quickly
- operators can buy/import templates safely
- organizations can version and govern their own templates

---

## 2. Template Source Model

Every template in Mycelis must declare a source class:

- `builtin`: shipped with Mycelis
- `marketplace`: external provider feed (for example, clawhub)
- `private_hub`: org-hosted template registry
- `tenant_custom`: user/org-authored template

Required source metadata:
- `source_id`
- `source_type`
- `publisher`
- `namespace`
- `template_ref`
- `version`
- `signature_status` (`verified|unverified|failed`)

---

## 3. Canonical Template Package Contract

All templates normalize to a package contract:

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

---

## 4. Marketplace API Contracts

### 4.1 Source Registration

- `GET /api/v1/template-market/sources`
  - list configured marketplace/private sources
- `POST /api/v1/template-market/sources`
  - register a source (`clawhub`, private hub, or custom URL)
- `PATCH /api/v1/template-market/sources/{source_id}`
  - rotate credentials, enable/disable
- `POST /api/v1/template-market/sources/{source_id}/probe`
  - probe feed health and auth

### 4.2 Discovery

- `GET /api/v1/template-market/templates`
  - search + filter by source, category, tag, pricing, compatibility
- `GET /api/v1/template-market/templates/{template_id}`
  - detailed package view, requirements, pricing, license

### 4.3 Acquire and Install

- `POST /api/v1/template-market/templates/{template_id}/purchase-intent`
  - create governed purchase proposal
- `POST /api/v1/template-market/purchases/{purchase_id}/confirm`
  - execute purchase after approval
- `POST /api/v1/template-market/templates/{template_id}/install`
  - install into tenant catalog (requires entitlement if paid)
- `POST /api/v1/template-market/templates/{template_id}/sync`
  - pull latest version metadata

### 4.4 Lifecycle

- `GET /api/v1/template-market/installs`
  - list installed marketplace templates + versions
- `POST /api/v1/template-market/installs/{install_id}/upgrade`
  - controlled upgrade (with compatibility checks)
- `DELETE /api/v1/template-market/installs/{install_id}`
  - uninstall

---

## 5. Custom Template API Contracts

Custom templates are first-class and tenant scoped.

- `GET /api/v1/templates/custom`
  - list custom templates for tenant/user
- `POST /api/v1/templates/custom`
  - create a new custom template
- `GET /api/v1/templates/custom/{template_id}`
  - template detail + version history
- `PUT /api/v1/templates/custom/{template_id}`
  - update draft metadata/spec
- `POST /api/v1/templates/custom/{template_id}/publish`
  - publish a versioned immutable release
- `POST /api/v1/templates/custom/{template_id}/fork`
  - fork from marketplace/builtin/custom template
- `DELETE /api/v1/templates/custom/{template_id}`
  - archive template

Forking rules:
- paid marketplace template forks must retain source and license attribution
- disallowed edits (license restricted) are blocked at publish step

---

## 6. Purchase, Licensing, and Entitlements

### 6.1 Purchase States

- `draft`
- `pending_approval`
- `approved`
- `completed`
- `failed`
- `revoked`

### 6.2 Entitlement Contract

Each paid template install binds an entitlement record:
- `tenant_id`
- `template_id`
- `license_id`
- `entitlement_scope` (`tenant|seat|environment`)
- `status`
- `expires_at` (if subscription)

### 6.3 Enforcement

- install blocked without active entitlement (for paid templates)
- upgrade blocked if license compatibility check fails
- runtime warnings for expiring subscriptions

---

## 7. Governance + Security Controls

Mandatory controls:

1. source allowlist by default
2. signed package verification for marketplace sources
3. purchase/install/upgrade are governed mutations
4. secret fields never returned in API payloads
5. deterministic audit trail for every purchase/install/upgrade/delete
6. template sandbox checks before activation
7. explicit compatibility validation:
   - provider availability
   - MCP/action dependencies
   - policy constraints

ClawHub-style support requirement:
- source connector supports token auth, catalog pull, and package verification
- all imported packages are normalized to canonical package contract before exposure

---

## 8. NATS Event Model

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

Every event must include:
- `tenant_id`
- `user_id`
- `run_id` (if workflow-driven)
- `template_id`
- `source_type`

---

## 9. UI/Workflow Requirements

### 9.1 Resources → Marketplace

Must support:
- source filters (`builtin|marketplace|private_hub|tenant_custom`)
- pricing and entitlement badges
- compatibility check result before install
- proposal flow for purchases (not silent buy)

### 9.2 Template Builder

For tenant custom templates:
- schema-aware template editor
- draft/publish workflow
- version history and rollback
- fork-from-installed action

### 9.3 Instantiation Flow

When launching teams from template:
- show source + license info
- show required capabilities and missing dependencies
- offer auto-remediation path where possible

---

## 10. Testing and Release Gates

### Unit
- package normalization
- license/entitlement validation
- source auth + signature verification

### Integration
- source registration/probe/sync
- purchase intent -> approval -> confirm
- install/upgrade/uninstall lifecycle
- custom template draft/publish/fork

### E2E
- buy/install from marketplace (governed)
- create/publish custom template and instantiate
- entitlement expiry behavior
- degraded marketplace source recovery

Release gate:
- no template-market production enablement until purchase/install paths are fully audited and reversible

---

## 11. Rollout Plan

### Wave 1
- source registration + marketplace discovery APIs
- read-only listing UI

### Wave 2
- governed purchase intent + entitlement model
- install/uninstall path

### Wave 3
- custom template builder + publish/fork
- upgrade path + compatibility checks

### Wave 4
- multi-source federation (`clawhub` + private hub + tenant custom)
- full observability and policy hardening

---

End of document.

