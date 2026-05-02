# V8.2 User Management And Enterprise Auth Module
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.2 Production Architecture](../../architecture/v8-2.md)

> Status: Canonical
> Last Updated: 2026-04-27
> Purpose: Define the V8.2 user-management, SSO, SAML, OIDC/OAuth, enterprise identity, and internal authorization module.

## TOC

- [Core Principle](#core-principle)
- [Edition Strategy](#edition-strategy)
- [Identity Tree](#identity-tree)
- [Auth Provider Model](#auth-provider-model)
- [Supported Auth Modes](#supported-auth-modes)
- [Authorization Model](#authorization-model)
- [Approval Authority](#approval-authority)
- [User Context For Soma](#user-context-for-soma)
- [UI Requirements](#ui-requirements)
- [Security Requirements](#security-requirements)
- [Backend Foundation](#backend-foundation)
- [Validation Plan](#validation-plan)
- [Release Posture](#release-posture)
- [Related References](#related-references)

## Core Principle

Authentication provider is pluggable. Authorization remains internal and canonical.

External identity systems may prove who a user is:
- Entra ID / Azure AD
- Google Workspace
- GitHub
- OIDC/OAuth providers
- SAML identity providers

Mycelis still owns runtime authority:
- tenant/account roles
- organization membership
- Soma access rules
- capability permissions
- approval rights
- audit visibility
- role-to-permission resolution

External auth must never become the source of internal runtime authority. Every authenticated subject resolves into a stable Mycelis principal before authorization, approvals, memory visibility, or capability access are evaluated.

## Edition Strategy

Free Node:
- single-user secure mode
- local login / bootstrap owner
- optional local OAuth/OIDC later
- no required SSO

Team:
- multi-user access
- invitations
- organization memberships
- role-based access

Business:
- groups
- service accounts
- admin controls
- stronger audit visibility

Enterprise:
- SSO
- SAML
- OIDC/OAuth
- Entra ID / Azure AD
- Google Workspace
- GitHub
- SCIM provisioning
- domain verification
- group-to-role mapping
- advanced audit/export
- policy enforcement

## Identity Tree

Tenant / Account owns:
- users
- groups
- roles
- auth providers
- organization memberships
- service accounts
- audit records

AI Organization owns scoped role assignments:
- Owner
- Manager
- Operator
- Reviewer
- Viewer

Soma must know the active user context:
- current user
- effective role
- allowed organizations
- user preferences
- approval authority
- capability permissions

Soma must not bypass access policy. Soma may help explain permissions and route requests, but backend authorization remains decisive.

## Auth Provider Model

`AuthProvider` is a configuration record with secret references only:

- `id`
- `type`: `local`, `oidc`, `oauth`, `saml`, `entra_id`, `google_workspace`, `github`
- `display_name`
- `enabled`
- `issuer`
- `client_id`
- `client_secret_ref`
- `redirect_uri`
- `scopes`
- `domain_restrictions`
- `group_claim_mapping`
- `role_mapping_rules`

Secrets rule:
- store secret references, not raw secret values
- never return raw secrets from APIs
- never display raw secrets in UI
- never log tokens, client secrets, SAML assertions, OIDC ID/access tokens, or authorization codes

## Supported Auth Modes

Local Auth:
- owner bootstrap
- password hashing
- secure sessions
- logout
- MFA-ready design

OIDC / OAuth:
- provider discovery
- callback handling
- issuer and audience validation
- token validation
- user identity mapping
- email and domain verification

SAML:
- metadata URL or uploaded metadata
- entity ID
- ACS URL
- certificate validation
- signature validation
- NameID/email mapping
- group and role claims

Entra ID:
- OIDC first
- tenant ID support
- enterprise app registration guidance
- group claim mapping
- role mapping into internal Mycelis roles

SCIM future:
- create/update/deactivate users
- group sync
- enterprise-only
- not required in first implementation

## Authorization Model

Use internal RBAC first.

Tenant roles:
- Account Owner
- Platform Admin
- Billing Admin
- Security Admin
- Organization Admin
- User
- Viewer

Organization roles:
- Owner
- Manager
- Operator
- Reviewer
- Viewer

Permission categories:
- access organization
- operate Soma
- confirm proposals
- approve high-risk actions
- edit AI Engine
- edit Response Style
- view audit
- manage users
- manage auth providers
- manage capabilities
- manage automations
- access advanced runtime surfaces

## Approval Authority

User role must influence approval workflow:
- Viewer cannot operate Soma
- Operator can request work
- Reviewer can approve review-required actions
- Manager can approve medium-risk actions
- Owner/Admin can approve high-risk and capability-changing actions

Every approval decision must be audited with:
- human principal
- organization scope
- effective role
- approval risk class
- proposal/action id
- decision
- timestamp
- auth source

## User Context For Soma

Each user may have Soma-readable context:
- preferred name
- timezone
- communication style
- role/title
- business context
- decision authority
- review strictness
- cost sensitivity
- automation tolerance

Visibility scopes:
- `private`
- `visible_to_soma`
- `visible_to_org_admins`
- `org_specific`

Soma may use this context for communication and planning. Soma may not use it to bypass authorization, approvals, memory visibility, or capability policy.

## UI Requirements

Free Node:
- sign in
- profile
- local owner account

Team / Business:
- Users
- Invitations
- Roles
- Organization access

Enterprise:
- Authentication Providers
- SSO setup
- SAML/OIDC configuration
- Domain restrictions
- Group mappings
- SCIM status
- Security audit

Keep user management outside the default Soma workspace. Default Soma interaction stays focused on direct work; identity administration belongs in Settings, admin, and security surfaces gated by role and edition.

Current implementation note:
- Settings now exposes an Advanced-mode, read-only Auth Providers scaffold for Local, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM. It is a secret-reference-only planning surface and does not activate provider adapters yet.

## Security Requirements

- secure cookies and sessions
- CSRF protection where applicable
- password hashing
- OIDC issuer/audience validation
- OAuth/OIDC token validation
- SAML signature validation
- certificate validation
- secret references only
- no auth tokens in logs
- audit login/logout/provider changes
- audit role changes
- audit approval actions
- fail closed for unknown provider, invalid token, missing required claim, or disabled user

## Backend Foundation

Add tables/models:
- `tenants` / `accounts`
- `users`
- `groups`
- `roles`
- `org_memberships`
- `auth_providers`
- `sessions`
- `user_profiles` / `user_context`
- `role_permissions`
- auth/access audit events

Add service interfaces:
- `AuthProviderAdapter`
- `UserDirectoryService`
- `AuthorizationService`
- `SessionService`
- `UserContextService`

Initial implementation should keep the provider adapter boundary narrow: authenticate externally, normalize to a Mycelis principal, then resolve internal authorization through Mycelis-owned services.

## Validation Plan

Backend tests:
- local login
- session creation/rejection
- role resolution
- organization membership checks
- permission checks
- auth-provider config validation
- user-context visibility rules

Security tests:
- invalid token rejected
- unauthorized organization access denied
- insufficient role cannot approve high-risk action
- secret values are never returned

Frontend tests:
- profile renders
- access-denied state
- user management gated by role
- auth-provider setup visible only to admins with enterprise entitlement

Release validation commands:
- `cd core && go test ./... -count=1 -p 1`
- `cd interface && npm test`
- `cd interface && npx tsc --noEmit`
- `uv run inv interface.e2e`
- `uv run pytest tests/test_docs_links.py -q`

## Release Posture

This module is a V8.2 target definition. It does not mean the full enterprise auth runtime is shipped.

Current free-node and V8.1 behavior remains simple:
- local/self-hosted owner posture
- deploy-owned edition/auth metadata
- internal principal metadata exposed through existing user APIs
- enterprise adapters deferred behind this module boundary

Next implementation slices should proceed in this order:
1. internal user/tenant/membership schema
2. session service and local auth hardening
3. internal authorization service and permission checks
4. user-management UI gated by role/edition
5. auth-provider config model with secret references
6. OIDC adapter
7. SAML adapter
8. Entra ID / Google Workspace / GitHub presets
9. SCIM provisioning

## Related References

- [V8 Multi-User Identity And Soma Tenancy](V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md)
- [Authentication Modes](../user/auth-modes.md)
- [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md)
- [Licensing & Editions](../licensing.md)
- [V8.2 Production Architecture](../../architecture/v8-2.md)
