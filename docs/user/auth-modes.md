# Authentication Modes
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this guide when enabling or reviewing login, local auth, break-glass recovery, or enterprise authentication posture.

## Current Boundary

Current self-hosted release authentication is deploy-owned:

- `.env` is the local secret store
- `MYCELIS_API_KEY` is the normal local admin credential
- `MYCELIS_WEB_SESSION_SECRET` signs browser sessions for the Interface login boundary
- `MYCELIS_WEB_IDENTITY_FORWARD_SECRET` optionally separates Interface-to-Core identity forwarding from the browser session secret
- `MYCELIS_LOCAL_ADMIN_USERNAME` and `MYCELIS_LOCAL_ADMIN_USER_ID` name that local principal
- Settings shows access posture, but user chat cannot rewrite it

The Interface always requires login, including Free Node deployments. Enterprise provider entries in Auth Providers are configuration contracts unless the deployment enables the matching adapter. Google Workspace OIDC is the first enabled SSO path: Google proves identity, while Mycelis still owns roles, organization membership, Soma access, approvals, capabilities, and audit visibility.

Role boundary:

- `admin`: configures identity, providers, people/access, AI engines, advanced system surfaces, deployment trust, and recovery.
- `standard`: works with Soma, teams, outputs, runs, proof, docs, and assigned organization workflows without changing provider or system configuration.

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
2. Set `MYCELIS_WEB_SESSION_SECRET`.
3. Set `MYCELIS_LOCAL_ADMIN_USERNAME`.
4. Set either `MYCELIS_LOCAL_ADMIN_PASSWORD` or `MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256`; if neither is set, the local login falls back to `MYCELIS_API_KEY`.
5. Set `MYCELIS_LOCAL_ADMIN_USER_ID`.
6. Keep `MYCELIS_IDENTITY_MODE=local_only`.
7. Set `MYCELIS_PUBLIC_ORIGIN=https://...` or `MYCELIS_WEB_COOKIE_SECURE=true` for HTTPS production deployments; local HTTP proof leaves both unset.
8. Optionally set `MYCELIS_WEB_IDENTITY_FORWARD_SECRET` when Core and Interface should use a dedicated HMAC secret for audit identity propagation; otherwise both use `MYCELIS_WEB_SESSION_SECRET`.
9. Run `uv run inv auth.posture` or `uv run inv auth.posture --compose`.
10. Verify `/login` accepts the local owner and Settings -> People & Access shows local-only/self-hosted release posture.

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

Use Google Workspace through OIDC/OAuth for enterprise SSO.

1. Create an OAuth client in Google Cloud.
2. Register the Mycelis redirect URI as `MYCELIS_AUTH_GOOGLE_REDIRECT_URI`.
3. Store `MYCELIS_AUTH_GOOGLE_CLIENT_ID` and `MYCELIS_AUTH_GOOGLE_CLIENT_SECRET` in the deployment secret layer.
4. Set `MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN` and/or `MYCELIS_AUTH_ALLOWED_DOMAINS`.
5. Set `MYCELIS_AUTH_ADMIN_EMAILS` for Mycelis admins; other allowed-domain users enter as standard users.
6. Keep organization access and approval rights in Mycelis roles.
7. Confirm `/dashboard` shows the signed-in Soma operating environment and Core audit/proof records use the signed web identity rather than the generic local API-key owner.

The login page shows the allowed Workspace domains when Google is configured. If Google account selection returns a domain error, choose an account from `MYCELIS_AUTH_ALLOWED_DOMAINS` or use the local owner login while correcting the deployment domain list.

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
8. `/login` explains the enabled path clearly: local owner login for self-hosted nodes, Google Workspace SSO only when configured, and allowed-domain guidance when domain restrictions are active.
9. After login, `/dashboard` shows the signed-in Soma operating environment with Access, Identity, and Scope before the operator starts work.
