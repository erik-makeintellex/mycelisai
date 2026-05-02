# Authentication Modes
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this guide when enabling or reviewing local auth, break-glass recovery, or enterprise authentication posture.

## Current Boundary

Current self-hosted release authentication is deploy-owned:

- `.env` is the local secret store
- `MYCELIS_API_KEY` is the normal local admin credential
- `MYCELIS_LOCAL_ADMIN_USERNAME` and `MYCELIS_LOCAL_ADMIN_USER_ID` name that local principal
- Settings shows access posture, but user chat cannot rewrite it

Enterprise provider rows in Auth Providers are currently a review/setup scaffold unless the deployment explicitly enables the matching adapter. External providers may prove identity, but Mycelis still owns roles, organization membership, Soma access, approvals, capabilities, and audit visibility.

## Deployment Posture

Set these in `.env` or through the deployment secret/config layer:

```env
MYCELIS_ACCESS_MANAGEMENT_TIER=release
MYCELIS_PRODUCT_EDITION=self_hosted_release
MYCELIS_IDENTITY_MODE=local_only
MYCELIS_SHARED_AGENT_SPECIFICITY_OWNER=root_admin
```

Enterprise-like posture uses:

```env
MYCELIS_ACCESS_MANAGEMENT_TIER=enterprise
MYCELIS_PRODUCT_EDITION=self_hosted_enterprise
MYCELIS_IDENTITY_MODE=hybrid
MYCELIS_SHARED_AGENT_SPECIFICITY_OWNER=root_admin
```

Use `federated` only when normal users are expected to authenticate through an external identity provider. Any `hybrid` or `federated` posture needs a separate break-glass principal.

## Local Owner Mode

Use this for Free Node and simple self-hosted release.

1. Generate or set `MYCELIS_API_KEY`.
2. Set `MYCELIS_LOCAL_ADMIN_USERNAME`.
3. Set `MYCELIS_LOCAL_ADMIN_USER_ID`.
4. Keep `MYCELIS_IDENTITY_MODE=local_only`.
5. Run `uv run inv auth.posture` or `uv run inv auth.posture --compose`.
6. Verify Settings -> People & Access shows local-only/self-hosted release posture.

Do not reuse the same value for local admin and break-glass credentials.

## Break-Glass Recovery

Use break-glass only for explicit recovery when enterprise, hybrid, or federated identity is unavailable.

```env
MYCELIS_BREAK_GLASS_API_KEY=change-this-recovery-key
MYCELIS_BREAK_GLASS_USERNAME=recovery-admin
MYCELIS_BREAK_GLASS_USER_ID=00000000-0000-0000-0000-000000000001
```

Rules:

- set all three break-glass values together
- keep the API key separate from `MYCELIS_API_KEY`
- keep the user id separate from the primary local admin id
- audit recovery use when audit-backed auth is active

## OIDC Or OAuth

Use OIDC/OAuth for standard enterprise SSO providers and custom identity platforms.

1. Create an application in the provider.
2. Register the Mycelis redirect URI shown in Auth Providers.
3. Store client secret material in `.env` or the deployment secret backend.
4. Configure issuer, client id, scopes, and secret reference in Auth Providers.
5. Require issuer, audience, email, and domain validation.
6. Map claims to internal Mycelis users and roles.
7. Test login, logout, invalid-token rejection, and disabled-user rejection.

Secret values should never appear in UI, logs, docs, or state files.

## Entra ID

Use Entra ID through OIDC first.

1. Register an enterprise app in Entra ID.
2. Record tenant id, issuer, client id, and redirect URI.
3. Store the client secret in `.env` or the enterprise secret backend.
4. Configure group or app-role claims.
5. Map Entra groups/app roles to internal Mycelis roles.
6. Keep Mycelis approval authority internal even when Entra supplies groups.

## Google Workspace

Use Google Workspace through OIDC/OAuth.

1. Create an OAuth client in Google Cloud.
2. Restrict allowed domains where required.
3. Store the client secret outside committed config.
4. Configure issuer, client id, scopes, and domain restrictions.
5. Map verified email/domain to a Mycelis user and tenant.
6. Keep organization access and approval rights in Mycelis roles.

## GitHub

Use GitHub for team login only when the organization accepts GitHub identity as a login proof.

1. Create a GitHub OAuth app or GitHub App.
2. Set callback/redirect URI.
3. Store client secret or app private key by secret reference.
4. Restrict allowed GitHub organizations or teams when needed.
5. Map GitHub identity to Mycelis users and roles.
6. Do not confuse GitHub login with GitHub MCP/tool access; tool credentials remain separate Connected Tools configuration.

## SAML

Use SAML for enterprise identity providers that require metadata exchange.

1. Create the SAML app in the IdP.
2. Configure entity ID and ACS URL from Auth Providers.
3. Provide IdP metadata URL or uploaded metadata.
4. Validate signing certificate and signatures.
5. Map NameID/email, groups, and roles into Mycelis principals.
6. Test invalid signature, missing claim, disabled user, and logout/recovery behavior.

SAML assertions and private key material must never be logged or displayed.

## SCIM Provisioning

SCIM is future enterprise-only lifecycle sync.

Expected enablement flow:

1. Create a SCIM token through the enterprise admin surface.
2. Store token material as a secret reference.
3. Configure create, update, deactivate, and group-sync behavior.
4. Keep authorization decisions in Mycelis role mappings.
5. Audit provisioning changes.

SCIM should not be required for Free Node or ordinary self-hosted release.

## Verification Checklist

Before accepting an auth-mode change:

1. Settings -> People & Access shows the expected edition and identity posture.
2. Auth Providers shows only secret references or redacted values.
3. `uv run inv auth.posture --compose` passes for Compose deployments.
4. Invalid credentials and invalid tokens fail closed.
5. A user cannot access an organization without internal Mycelis membership.
6. Approval authority follows internal role, not external group alone.
7. Access-denied recovery points to Settings, System Status, and owner/admin action.
