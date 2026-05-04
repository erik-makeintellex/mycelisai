# V8 Enterprise Auth Details
> Navigation: [V8.2 User Management And Enterprise Auth Module](V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md)

Status: V8.2 auth detail.

## Provider Abstraction

Providers should normalize identity claims into Mycelis user records without making provider-specific claims the authorization source of truth.

Provider classes:
- local-admin API key
- break-glass recovery API key
- OIDC/OAuth
- SAML
- Entra ID
- Google Workspace
- GitHub
- future SCIM/lifecycle sync

## Authorization Model

Mycelis owns:
- roles
- groups/organizations access
- approval authority
- admin/recovery posture
- audit context
- policy checks

External IdPs may supply identity and group claims, but Mycelis decides capability and approval authority.

## Break-Glass

Break-glass credentials must be:
- optional
- explicit
- separately identified
- auditable
- documented as recovery posture only

## UI Requirements

Settings/access UI should show:
- current auth posture
- provider state
- user roles/authority
- recovery posture when configured
- normalized blockers for missing/invalid provider setup

## Security Rules

- no raw tokens in committed config
- no provider secrets in UI
- no secret values in state files or architecture docs
- auth failures produce normalized errors
- provider config is validated before claiming readiness
