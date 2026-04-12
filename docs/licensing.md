# Licensing & Editions
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

This document explains the current Mycelis product-edition and deployment licensing posture.

It is intentionally about:
- product editions
- self-hosted vs hosted layering
- which capabilities are expected to remain modular or paid
- how shared Soma governance relates to those layers

It is not a substitute for:
- an executed commercial agreement
- third-party provider terms
- future formal legal packaging that may add a root license file or separate commercial terms

## TOC

- [Why This Exists](#why-this-exists)
- [Core Position](#core-position)
- [Edition Model](#edition-model)
- [Self-Hosted Release](#self-hosted-release)
- [Self-Hosted Enterprise](#self-hosted-enterprise)
- [Hosted Control Plane](#hosted-control-plane)
- [Identity And User Management Layering](#identity-and-user-management-layering)
- [Shared Soma Governance Boundary](#shared-soma-governance-boundary)
- [MCP, Web Access, And Context Security](#mcp-web-access-and-context-security)
- [Third-Party Terms And Runtime Costs](#third-party-terms-and-runtime-costs)
- [Current Repository Posture](#current-repository-posture)

## Why This Exists

Mycelis now has a clear product-layering story:
- a self-hosted base release
- a self-hosted enterprise growth layer
- a hosted control-plane/service layer

That layering has already started to appear in product and architecture surfaces such as `product_edition`, `identity_mode`, and `shared_agent_specificity_owner`.

This document makes the licensing and edition intent explicit so:
- investors can understand the monetization and packaging direction
- operators can understand what belongs in base self-host
- contributors do not accidentally collapse paid layers into assumed free defaults

## Core Position

Mycelis should stay usable as a self-hosted product without forcing a hosted control plane.

At the same time, some higher-order operational layers are intended to remain modular and commercially packageable:
- enterprise identity adapters
- advanced user-management control layers
- hosted administration and directory services
- expanded compliance, audit, and policy-management layers

Product rule:
- paid variants should add governed control layers and enterprise operating features
- paid variants should not weaken governance, audit, or security boundaries
- edition changes must not redefine the shared-Soma operating model

## Edition Model

The intended edition model is:

1. Self-hosted release
   - the base product a customer can run directly
   - local environment ownership
   - local operator governance
   - one shared organization-owned Soma persona

2. Self-hosted enterprise
   - an add-on layer for enterprise identity, richer administration, and larger-organization governance
   - still self-hosted at the runtime layer
   - intended to preserve local control and local break-glass recovery

3. Hosted control plane
   - a paid hosted management/service layer
   - can provide centralized identity, directory sync, admin tooling, policy posture, and fleet-style management
   - should remain modular instead of becoming a mandatory dependency for self-hosted runtime operation

## Self-Hosted Release

The self-hosted release is the expected base edition.

It should include:
- local runtime operation
- local primary admin ownership
- optional local break-glass recovery principal
- governed Soma-first workflow
- approvals, audit, and capability-aware policy
- local or customer-managed MCP connectivity
- local or customer-managed provider connectivity
- governed private user context, deployment context, and RAG stores
- root-admin ownership of durable shared Soma shaping and output specificity

It should not assume:
- hosted identity plane
- mandatory vendor-managed directory services
- forced SaaS control dependencies just to run the product

## Self-Hosted Enterprise

The self-hosted enterprise layer is the intended paid expansion for organizations that need stronger identity and administration without surrendering runtime ownership.

Expected enterprise-layer capabilities:
- SAML and/or OIDC federation
- optional SCIM lifecycle sync
- richer role and access administration
- stronger shared-Soma governance controls
- broader audit/compliance administration
- delegated environment ownership and approval chains

Important boundary:
- enterprise self-host should still preserve local administrative recovery
- federated identity must not remove the break-glass path for self-hosted recovery

## Hosted Control Plane

The hosted control plane is the intended paid hosted layer for customers who want managed user administration and broader operational services.

Expected hosted-layer capabilities:
- hosted user-management plane
- managed directory and identity integration
- centralized environment administration
- optional cross-environment management and reporting
- commercial support surfaces around identity and policy operations

Hosted control plane rule:
- it is an additive layer
- it must not be the only way to run Mycelis in self-hosted mode

## Identity And User Management Layering

User management is intentionally modular.

That means:
- base self-host can run with local principals
- self-hosted enterprise can add federation and enterprise user lifecycle
- hosted control plane can supply the management plane as a paid service

This is the intended licensing/packaging split:
- runtime governance and core Soma operation belong to the base product
- advanced directory and control-plane services are eligible paid layers

## Shared Soma Governance Boundary

Editions do not change the fact that Mycelis operates through one shared organization-owned Soma persona.

The licensing model must preserve these rules:
- Soma is configured by the environment or organization owner
- the root admin or explicitly delegated owner controls durable shared output specificity
- admin-shaped `soma_operating_context` is organization-owned and governed
- ordinary user chats may request temporary local preferences but may not silently rewrite shared Soma behavior

Edition boundary:
- paid identity or control-plane features may change how users are authenticated and administered
- they must not turn Soma into a separate uncontrolled persona per user

## MCP, Web Access, And Context Security

MCP onboarding, tool usage visibility, web access, and deployment-context loading are governance features first, not edition bypasses.

That means:
- MCP can be present in base self-host, subject to governance and audit
- web access should remain securable, reviewable, and policy-bound
- context security posture should remain configurable by trust, sensitivity, source kind, and approval policy
- paid editions may add richer administration around those controls, but should not replace the need for them

In practical terms:
- base self-host should support governed MCP and external research usage
- enterprise or hosted layers may add stronger administration, directory integration, and policy management around those surfaces

## Third-Party Terms And Runtime Costs

This product-edition document does not change third-party terms.

Operators still need to account for:
- model-provider pricing
- external SaaS/provider terms
- third-party open-source licenses inside dependencies
- infrastructure costs for self-hosted runtime, storage, and model serving

Important distinction:
- a Mycelis paid edition does not automatically grant rights to third-party model or SaaS services
- those remain separately governed by the provider's own terms and pricing

## Current Repository Posture

Current repo truth:
- the product already exposes the edition story for review through `product_edition`, `identity_mode`, and `shared_agent_specificity_owner`
- the current runtime now supports explicit local admin vs break-glass admin posture
- the current docs define self-hosted release, self-hosted enterprise, and hosted control plane as the intended layering
- the full enterprise adapters and hosted control plane are still implementation work, not fully delivered runtime

Current documentation rule:
- treat this document as the canonical licensing-and-editions posture
- treat architecture and governance docs as the technical implementation contract behind it

Related references:
- [Governance System](./governance.md)
- [Governance & Trust](./user/governance-trust.md)
- [V8 Multi-User Identity And Soma Tenancy](./architecture-library/V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md)
- [API Reference](./API_REFERENCE.md)
