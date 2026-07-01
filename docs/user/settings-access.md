# Settings And Access
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this guide when you need to understand profile, access, auth-provider, and connected-tool management without reading implementation docs.

## What Settings Is For

Settings is the administrative surface for account posture and operator preferences. It should help an owner answer:

- who is operating this node
- whether the current web session is admin or standard
- which edition/access posture is active
- whether external auth is configured or only planned
- which AI engine and MCP/tool surfaces are available
- what recovery action is needed when access is denied

For a new admin, the guided setup panel should make four checks obvious:
SSO/auth, AI provider, workspace/output roots, and Soma capability setup.
Workspace/output roots are confirmed from `System -> Deployments`, while
capability install, web-access setup, and filesystem workspace browsing remain
in `Resources`. Settings includes a direct **Open web access setup** shortcut so
an operator who asked Soma how to enable web access does not have to discover
the Resources path by trial and error.

Settings is not the default Soma workspace. Use Soma for normal work, groups, and execution requests; use Settings when you need account, provider, or access posture.

## Profile

Profile controls operator-facing identity preferences such as preferred name, timezone, and assistant naming.

Current expectations:

- profile preferences may change how the interface names the operator or assistant
- the authenticated Soma shell defaults to the dark `midnight-cortex` theme before hydration; saved `aero-light`, `midnight-cortex`, or `system` preferences are applied before the shell paints so login does not flash between themes
- deploy-owned edition and access posture are read-only here
- profile edits must not override `.env` or deployment-owned auth configuration

## People And Access

`Settings -> People & Access` shows the current principal and account access posture.

In the free/self-hosted node release this is deliberately simple:

- the Interface still requires login; local owner login is the default free-node path
- the local owner or bootstrap principal remains the primary operator
- access-management tier, product edition, identity mode, and shared-Soma ownership are deployment-owned
- role and status fields are visibility surfaces unless the current edition enables management actions
- Google Workspace OIDC can create standard/admin web sessions when configured; broader directory, SAML/OIDC, and SCIM flows remain gated by edition and provider readiness

Mycelis authorization remains internal. External identity providers may later prove who a user is, but they do not become the source of runtime authority for Soma, groups, approvals, capabilities, or audit visibility.

## Auth Providers

Auth Providers is the setup and inspection surface for local auth, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM posture. The page uses a compact provider menu with one focused detail panel so operators can compare options without reading every provider contract at once. Use [Authentication Modes](auth-modes.md) for the full enablement checklist.

Current expectations:

- provider menu entries may show planned or configured posture before full adapter activation
- secrets are referenced, not displayed
- raw client secrets, tokens, and SAML private material must stay in `.env` or the configured secret backend
- provider changes should be audited when mutation support is enabled

For the current release, Google Workspace is the enabled enterprise SSO path. Other Enterprise SSO entries remain target contracts unless the deployment explicitly enables the matching provider flow.

Mode summary:

- local owner mode: set `MYCELIS_API_KEY`, `MYCELIS_WEB_SESSION_SECRET`, local admin name/id, optional local login password/hash, and `MYCELIS_IDENTITY_MODE=local_only`
- break-glass recovery: set a separate `MYCELIS_BREAK_GLASS_API_KEY`, username, and user id for hybrid/federated recovery
- OIDC/OAuth: configure issuer, client id, redirect URI, scopes, and secret reference; validate issuer/audience/email/domain
- Entra ID: use OIDC first, add tenant id and group/app-role claims, then map them to internal Mycelis roles
- Google Workspace: use OIDC/OAuth with domain restrictions, verified email mapping, and `MYCELIS_AUTH_ADMIN_EMAILS` for admin assignment
- GitHub: use OAuth/App identity for login proof only; keep GitHub tool/MCP credentials separate
- SAML: configure metadata, entity ID, ACS URL, certificate/signature validation, NameID/email, and group claims
- SCIM: future enterprise-only provisioning/deprovisioning, not required for Free Node or base self-hosted release

## Capabilities And MCP

Capability and MCP management lives in `Resources -> Capabilities`, while Settings exposes related deep links for setup tasks.

Use Capabilities to inspect:

- installed MCP servers
- capability permission groups for Everyone, Group, or Host boundaries
- each server's transport, command or endpoint, args, env/header references, and status
- available tools
- recent persisted MCP activity
- active search posture, including `local_sources`, self-hosted `searxng`, operator-owned `local_api`, optional `brave`, and disabled blockers
- configured search sources, including public web providers, local/shared sources, explicit URL retrieval, and private or client-owned authenticated data sources

Soma should use the Mycelis-owned `web_search` path when search is configured. Brave is optional; self-hosted SearXNG and local API search do not require Brave tokens. The `fetch` MCP is also optional: add or repair it when Soma needs to retrieve a specific supplied URL, not as a prerequisite for built-in search.

Private or client-owned data sources should be configured as governed search sources rather than pasted into chat. Open `Resources -> Capabilities`, use **Add search source**, name the source, provide the endpoint/base URL when it is external or API-backed, describe the allowed source boundary, choose Everyone/Group/Host visibility, and select a saved secret reference when authentication is needed. Operator-managed sources can be edited or removed from that same Capabilities lane. Use standard token-based web auth where possible: bearer token or API-token reference first, service-required query/header placement only when necessary, and OAuth2/client-credential metadata as a follow-on extension when supported. Store the actual secret in `.env` or the configured secret backend, then select the secret reference in the UI. Scope the source to Everyone, one Group, or one Host so Soma searches it only where the operator intended. When a `source_id` is selected, the runtime enforces source status and scope before use; authenticated sources are registered for governance but return a repairable blocker until a safe adapter exists for the selected auth shape.

Use Capability permissions when you need to decide where Soma can use connected tools. Use Library when you need to edit or reapply a curated MCP structure. Secrets stay in `.env` or the configured secret backend; the UI should show only references or redacted values.

Useful Soma prompts from this surface:

- `Search the web for "<topic>", summarize the strongest sources, and cite them.`
- `Use host data under workspace/shared-sources and list the files that shaped the answer.`
- `Search the approved customer portal and company docs for the current onboarding policy, then tell me which source each claim came from.`
- `Add an authenticated search source for this internal docs URL using the configured token reference, limited to this group, and ask me to approve before Soma uses it.`
- `Review current MCP servers, tools, and recent use, then tell me which agents should have which tools.`
- `Review my latest team request, match it to prior action context, name the MCP servers needed, and ask me to confirm before enabling or assigning tools.`

When Soma infers that a team or agent needs a missing MCP server, it should guide rather than guess:

1. identify whether the server is already registered
2. name the library entry to install or reapply
3. name required `.env` variables without showing secret values
4. ask for confirmation before assigning the tool refs to a team/member template
5. return to the requesting Soma workflow after the server is enabled

Protected wording matters here. Requests that mention private services, credentials, customer/private data, production systems, or recurring behavior should map to protected interaction templates. Soma can review the request and explain the likely path, but it should ask for confirmation before using the service/data, binding a tool, or storing the behavior as a reusable template.

## Access Denied

If you land on Access Denied:

1. Check the account role and edition posture in Settings.
2. Open System Status to confirm the backend is reachable.
3. Confirm the deployment has the expected `.env` values for break-glass or enterprise-like recovery.
4. Ask an owner/admin to grant access when the edition supports user management.

Access-denied recovery should be explicit. The UI should not silently downgrade role checks or let Soma bypass policy.
