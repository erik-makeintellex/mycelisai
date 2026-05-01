# Settings And Access
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this guide when you need to understand profile, access, auth-provider, and connected-tool management without reading implementation docs.

## What Settings Is For

Settings is the administrative surface for account posture and operator preferences. It should help an owner answer:

- who is operating this node
- which edition/access posture is active
- whether external auth is configured or only planned
- which AI engine and MCP/tool surfaces are available
- what recovery action is needed when access is denied

Settings is not the default Soma workspace. Use Soma for normal work, groups, and execution requests; use Settings when you need account, provider, or access posture.

## Profile

Profile controls operator-facing identity preferences such as preferred name, timezone, and assistant naming.

Current expectations:

- profile preferences may change how the interface names the operator or assistant
- deploy-owned edition and access posture are read-only here
- profile edits must not override `.env` or deployment-owned auth configuration

## People And Access

`Settings -> People & Access` shows the current principal and account access posture.

In the free/self-hosted node release this is deliberately simple:

- the local owner or bootstrap principal remains the primary operator
- access-management tier, product edition, identity mode, and shared-Soma ownership are deployment-owned
- role and status fields are visibility surfaces unless the current edition enables management actions
- enterprise directory, SSO, SAML/OIDC, and SCIM flows remain gated by edition and provider readiness

Mycelis authorization remains internal. External identity providers may later prove who a user is, but they do not become the source of runtime authority for Soma, groups, approvals, capabilities, or audit visibility.

## Auth Providers

Auth Providers is the setup and inspection surface for local auth, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM posture.

Current expectations:

- provider rows may show planned or configured posture before full adapter activation
- secrets are referenced, not displayed
- raw client secrets, tokens, and SAML private material must stay in `.env` or the configured secret backend
- provider changes should be audited when mutation support is enabled

For the current release, treat Enterprise SSO configuration as a target contract unless the deployment explicitly enables the corresponding provider flow.

## Connected Tools And MCP

Connected-tool management lives in `Resources -> Connected Tools`, while Settings may expose related deep links.

Use Connected Tools to inspect:

- installed MCP servers
- each server's transport, command or endpoint, args, env/header references, and status
- available tools
- recent persisted MCP activity
- active search posture, including `local_sources`, self-hosted `searxng`, operator-owned `local_api`, optional `brave`, and disabled blockers

Soma should use the Mycelis-owned `web_search` path when search is configured. Brave is optional; self-hosted SearXNG and local API search do not require Brave tokens.

Use Library when you need to edit or reapply a curated MCP structure. Secrets stay in `.env` or the configured secret backend; the UI should show only references or redacted values.

Useful Soma prompts from this surface:

- `Search the web for "<topic>", summarize the strongest sources, and cite them.`
- `Use host data under workspace/shared-sources and list the files that shaped the answer.`
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
