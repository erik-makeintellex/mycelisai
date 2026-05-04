# V8.2 User Management And Enterprise Auth Module
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: V8.2 target module.

This index owns the enterprise-ready identity/auth module. Implementation detail now lives in [V8 Enterprise Auth Details](V8_ENTERPRISE_AUTH_DETAILS.md).

## Purpose

Define the target posture for user management, pluggable auth providers, enterprise SSO, internal RBAC, approval authority, recovery, and future lifecycle sync.

## Scope

Includes:
- local-admin self-hosted posture
- break-glass recovery posture
- pluggable authentication providers
- OIDC/OAuth
- SAML
- Entra ID / Google Workspace / GitHub provider paths
- Mycelis-owned authorization and RBAC
- approval authority mapping
- user context visibility
- future SCIM/lifecycle sync target

Excludes:
- storing raw provider secrets in docs or UI
- making hosted enterprise IAM mandatory for personal-owner Compose
- bypassing Mycelis authorization because an external identity provider authenticated the user

## Product Rules

- Authentication proves who the user is.
- Mycelis authorization decides what the user can do.
- Approval authority is a Mycelis policy decision.
- Break-glass access is explicit recovery posture, not everyday access.
- User-facing errors must be normalized.

## Module Boundaries

Identity/auth work touches:
- backend auth middleware and user records
- UI settings/access surfaces
- governance proposal authority
- audit/activity context
- deployment secret posture
- tests and docs

## Edition Strategy

Personal/self-hosted lanes keep local-admin and optional break-glass flows. Enterprise lanes add SSO/SAML/OIDC, provider mapping, delegated admin, and future lifecycle sync.

## Validation Plan

Minimum proof for auth module changes:
- backend auth/user tests
- UI access/settings tests
- proposal authority tests when approval behavior changes
- no raw secrets in docs/logs/UI
- remote or live proof for delivered enterprise topology when relevant

## References

- [V8 Enterprise Auth Details](V8_ENTERPRISE_AUTH_DETAILS.md)
- [V8 Multi-User Identity And Soma Tenancy](V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md)
- [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [Licensing](../licensing.md)
