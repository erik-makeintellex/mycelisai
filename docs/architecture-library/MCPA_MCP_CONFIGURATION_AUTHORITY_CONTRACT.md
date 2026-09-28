# MCPA Operator MCP Configuration Authority Contract

Owner: Core authority and MCP runtime. Product authority remains the [canonical PRD](MYCELIS_CANONICAL_PRD.md).

> Draft contract for gating MCP *configuration* routes (architect draft 2026-09-28; owner answers pending). Slice id `MCPA`, the follow-up named by MCPS D10. Packet base `dev` @f16e32b7; every citation was read at `dev` 111f62f4 (f16e32b7 + HYG gofmt, which changed no cited file). Sibling: [MCPS contract](MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md) (call authority, W1 worktree `mcps-authority` has no edits yet).

## Verified today (source facts)
- **Routes** (server/admin_routes.go): servers list :186, delete :187, library list :191, inspect :192, install :193, apply :194, toolsets list/create/update/delete :270-273. Raw install is a hard 403 (:181-185). No other HTTP path writes `mcp_servers` or `mcp_tool_sets`: `config_documents*.go`, `organization*.go` and `deployment*.go` never touch MCP, and `templates_execution*.go` only *calls* tools (MCPS scope). Non-HTTP writers are startup only: `BootstrapDefaults` installs `filesystem` and `fetch` and seeds the `workspace`/`research` toolsets (mcp/service_bootstrap.go:11, :78-81), and `EnsureRuntimeDefaults` runs at cmd/server/startup_product.go:255.
- **Auth gate: authentication only, on every route.** No handler reads a role or scope: mcp_library.go:31-132, mcp.go:66-93, mcp_toolsets.go:23-159. Principals are the same as in MCPS. A standard web user is `operator` with `soma:work, runs:read, outputs:read` (auth.go:210-230). A web admin gets `*`, and the API key and break-glass key are `admin`/`*` (auth.go:85-105). There are no scoped admins yet.
- **Install/apply is process execution.**
  - Install and apply share `installMCPLibraryEntry` (mcp_library.go:159-190): `ToServerConfig(req.Env)` → `ApplyRuntimeDefaults` → `Service.Install` → `ClientPool.Connect`.
  - Every curated entry is `transport: stdio`, `command: "npx"` (core/config/mcp-library.yaml, 16 entries). Connect spawns `transport.NewStdio(cmd, env, args...)` (mcp/pool.go:48-55) inside Core.
  - **`env` overrides accept any key.** `ToServerConfig` copies every request key over the entry defaults (mcp/library.go:112-115), with no check against `DeclaredEnvKeys` (:128). So `NODE_OPTIONS`, `LD_PRELOAD` or `PATH` reach the npx child: code execution in Core.
  - `Install` is `ON CONFLICT (name) DO UPDATE` (mcp/service.go:42-55), and Connect disconnects first (pool.go:43). A caller can therefore *replace* the env of an existing server (for example `filesystem`) and relaunch it.
  - `postgres`/`comfyui`/`stable-diffusion` take caller URLs (yaml :57, :165, :251), so a caller can point Core at arbitrary hosts.
- **Governance posture is advisory and partly spoofable.**
  - `buildMCPLibraryGovernanceDecision` (mcp_governance_decisions.go:5-40) returns `require_approval` for external_saas/remote entries. Install/apply then answer 202 (mcp_library.go:71-78, :109-117).
  - **No approval object exists and nothing resumes the install.** The interface treats any non-`allow` as failure (store/cortexStoreMcpSlice.ts:255-261), so github/slack/brave/flux/elevenlabs/replicate/dall-e cannot be installed by anyone over HTTP. The 202 is a dead end.
  - `normalizeMCPGovernanceContext` prefers body `actor_role`/`owner_user_id` over the identity (mcp_governance.go:31-47). A body can claim `owner`. Today this changes only the decision *label*, because `high` risk implies remote and remote is matched first (:165-185; decisions :22-31).
  - Inspection advertises `required_scopes: ["mcp:write"]` (mcp_governance.go:219). Nothing enforces it.
  - **`mcp:write` is also a valid agent tool ref:** `ParseToolRef("mcp:write")` = server `write`, tool `*` (mcp/toolref.go:14-23).
- **Delete** (mcp.go:66-93): disconnects the pool client first, best-effort (:82), then deletes the row (:87). An authenticated caller can remove `filesystem` and break Resources → Workspace until a restart re-bootstraps it.
- **Toolsets widen agent authority.** Agents declaring `toolset:<name>` expand it through `ResolveRefs` (swarm/tool_scope.go:61-77, :93-119), and non-`mcp:` refs become internal tool grants (:79-82). A standard user can PUT `mcp:github/*` or internal tool names into the shared `workspace` set and widen every agent that declares it. The expansion is cached per agent scope instance (:96-97), so narrowing a toolset is not seen by live agents. The `mcp_tool_sets` schema has no owner column (001_current_schema.sql:799-807). The `governance` field in toolset responses is a label computed after the write (mcp_toolsets.go:76, :119, :157).
- **Secrets.**
  - Env values are stored plaintext in `mcp_servers.env` JSONB (001_current_schema.sql:387).
  - Responses redact env and headers (mcp_redaction.go:7-26; mcp.go:56; mcp_library.go:86, :127).
  - Install errors echo `err.Error()` (mcp_library.go:163, :169).
- **Audit:** none on any configuration route. The fail-closed precedent is `createAuditEvent` (templates_audit.go:15; nil DB returns `""`, nil, :17-18) with a 503 on an empty id (token_budgets.go:272-278).
- **Interface callers:**
  - `/resources?tab=tools` renders `MCPToolRegistry` for every role (app/(app)/resources/page.tsx:188). The BFF gates only `/settings?tab=tools` (proxy.ts:86-91).
  - Install goes through `MCPLibraryBrowser.tsx:49` → `installFromLibrary` (store :241-285). Its success copy says "Installed into your current MCP group" (:273), but servers are global, with no owner/group column (schema :381-394).
  - Delete goes through `MCPToolRegistry.tsx:211` → `deleteMCPServer`, which swallows non-2xx silently (store :90-101).
  - Toolset create goes through `MCPToolSetLayersPanel.tsx:234`.
  - `/api/v1/user/me` reports `role`, not scopes (identity.go:46-56).
  - e2e: mcp-connected-tools (legacy cap 515), mcp-connected-tools-edge, mcp-toolset-layers.

## Threat model
| Principal | Can today | Should |
|---|---|---|
| Standard user (`operator`) | Install or reinstall any local-first curated server with arbitrary env (code execution in Core through `NODE_OPTIONS`), repoint postgres/comfyui URLs, delete any server, and rewrite shared toolsets to widen every agent. Unaudited. | Read the library, servers, tools and toolsets. No configuration writes: 403 `admin_required`. |
| Scoped admin (no `mcp_config:write`) | Same as the standard user. | Reads only. 403 `admin_required` with `required_scope`. |
| Root admin / API key / break-glass (`*`) | Everything, unaudited. Credentialed SaaS entries are impossible (202 dead end). | Everything, with a fail-closed audit first and only declared env keys. `require_approval` entries need `approvals:decide`, recorded as a self-approved tier-2 action. |
Other threats:
- Env injection even by admins, because a pasted snippet can carry `NODE_OPTIONS`.
- Body-spoofed `governance_context` in the audit trail.
- The `mcp:write` string doubling as a tool grant.
- Replace-by-name silently swapping credentials on an existing server.

## Decisions (one recommended option each)
- **D1. Core enforces in one new file; no new routes, no BFF gate.**
  - Create `server/mcp_config_authority.go` with `requireMCPConfigWrite(w,r)` (it wraps `requireRootAdminScope(w,r,scopeMCPConfigWrite)`) and `auditMCPConfigChange(...)`.
  - The check runs first in install, apply, delete, and toolset create/update/delete, before the subsystem-nil checks and before body decode side effects.
- **D2. Scope: new `mcp_config:write`, root admin role.** It follows the `config_documents:write`/`groups:write` pattern (groups_auth.go:27; config_documents.go:47). **`mcp:write` is rejected**, because it parses as an agent tool grant (toolref.go:21-22). Inspection's `required_scopes` changes to `["mcp_config:write"]`, plus `approvals:decide` when the decision is `require_approval`. Web admins, the API key and break-glass keep access through `*`.
- **D3. Reads stay authentication-only.** Library, servers (already redacted), tools, toolsets list, activity and **inspect** have no side effects, and Soma/Resources render them for everyone. Inspect answers from the identity: D6 ignores body `actor_role`/`owner_user_id`.
- **D4. Env allowlist for every principal.**
  - Reject any `env` key not in `entry.DeclaredEnvKeys()`, and any empty key, with 400 `mcp_env_rejected`. The response names the keys, never the values. The check happens before Install.
  - Keys are matched exactly and case-sensitively.
  - Values are not validated (curated semantics). This closes code execution through env even for admins.
- **D5. `require_approval` becomes a real tier-2 path, not a dead end.**
  - The install/apply caller must pass D2 *and* `isApprover` (governance_authority.go:181). Otherwise the response is 403 `admin_required` with `required_scope=approvals:decide`.
  - An approver proceeds, recorded as tier 2, authority `approvals:decide`, `self_approved=true` (MCPS D5 pattern). No new token purpose, proposal kind or schema.
  - The 202 response is removed, because it claims a pending approval that cannot exist (no-placeholder rule).
- **D6. Truthful governance context.** `normalizeMCPGovernanceContext` takes `OwnerUserID`/`ActorRole` from the identity only. Body values are ignored, but `source_surface`/`config_scope` stay as labels. The "your current MCP group" copy goes, because servers are global.
- **D7. Fail-closed audit before every write.**
  - Call `createAuditEvent(protocol.TemplateChatToProposal, mcpConfigAuditSource="mcp-config", ...)` with `attachActorIdentity`.
  - Actions: `mcp_server_installed` (name, transport, deployment/credential boundary, sorted env **keys** via `sortedMCPEnvKeys`, `replaced_existing`, tier, self_approved), `mcp_server_deleted` (id, name), `mcp_toolset_created|updated|deleted` (id, name, scope_kind/ref, `tool_refs` before/after).
  - If the audit id is empty or there is an error, return 503 `service_unavailable` with zero Install/Connect/Disconnect/Delete/toolset calls. This includes a nil DB, which covers the templates_audit.go:17-18 graceful-nil case.
  - Failures after the audit get a best-effort `..._failed` record.
- **D8. Secrets hygiene without schema.**
  - Install and delete errors return a fixed message, and the raw `err` goes only to a redacted log via `mcp.RedactToolText` (the MCPS W1 export). No env value appears in responses, audit or logs.
  - Plaintext `mcp_servers.env` at rest is **out of scope** (owner Q3).
- **D9. Delete ordering.** Authority, then audit, then `MCP.Get` (404 `mcp_server_not_found` when absent, no disconnect), then Disconnect, then Delete. An unknown id no longer disconnects anything.
- **D10. Interface.**
  - Resources → Tools hides install, delete and new-layer controls unless `/user/me` role is `admin`, and shows `blockerCopy` on any 403.
  - `deleteMCPServer` and `createMCPToolSet` surface the blocker envelope instead of failing silently or showing raw text.
  - `blockerCopy.ts` gains `mcp_env_rejected`. `mcp_server_not_found` is optional copy.
- **D11. Out of scope:**
  - Per-group MCP servers and owner columns.
  - Secret references for env.
  - Live-agent toolset revocation (cached expansion, tool_scope.go:96).
  - The MCPS call route.

## Test-first acceptance (-race; fake pool and fake services record calls; helper prefix `mcpa`)
1. Positive:
   - An admin `*` installs `filesystem` and `fetch` with no env: 200, with the audit written before `Install` (ordering asserted) and env keys only in the audit.
   - The API key behaves the same.
   - An admin deletes an existing server: Get, then Disconnect, then Delete.
   - Toolset create, update and delete are audited with refs before and after.
   - An approver installs `github` with `GITHUB_PERSONAL_ACCESS_TOKEN`: 200 with `self_approved=true`, and the token value appears in neither the response nor the audit.
2. Negative:
   - A standard identity on install, apply, delete, and toolset POST/PUT/DELETE gets 403 `admin_required` with zero service or pool calls.
   - An admin with `outputs:read` but without `mcp_config:write` gets 403 with `required_scope=mcp_config:write`.
   - An admin with `mcp_config:write` but without `approvals:decide` installing `github` gets 403 with `required_scope=approvals:decide`.
   - No identity gets 401.
   - Reads (library, servers, tools, toolsets GET, inspect) as standard still return 200.
3. Env: an admin sending `{"name":"fetch","env":{"NODE_OPTIONS":"--require /tmp/x"}}`, `LD_PRELOAD`, `PATH`, a case-variant `github_personal_access_token` on github, or an empty key gets 400 `mcp_env_rejected` with zero Install calls, and the body contains no value.
4. Adversarial:
   - Body `governance_context.actor_role="owner"` and `owner_user_id` from a standard user still gets 403. From an admin, the audit and the decision carry the identity's user id, not the body's.
   - Reinstalling an existing name records `replaced_existing=true`.
   - `mcp_config:write` in an agent Tools list grants nothing: `IsMCPRef` is false (a regression test in swarm is read-only; assert only in server).
   - Forged web identity returns 401 (existing test, rerun).
   - An unknown delete id returns 404 with zero Disconnect calls.
5. Audit fail-closed: with a nil DB or a failing insert, every write returns 503 `service_unavailable` with zero side-effect calls, while reads are unaffected.
6. No dead end: no install or apply path returns 202, and inspect still returns `require_approval` with the new `required_scopes`.
7. Existing mcp_library_apply_test, mcp_redaction_test, mcp_toolsets_test, mcp_test and mcp_library_inspect_test pass after moving them to admin identities. The mutation proof removes the D1 hook, and tests 2, 3 and 5 fail.
8. Interface:
   - A standard role sees no install, delete or new-layer controls.
   - A 403 renders blocker copy, and delete no longer fails silently.
   - A `blockerCopy` snapshot covers `mcp_env_rejected` for user and admin.
   - The e2e stubs that assert install success run as admin, and mcp-connected-tools.spec.ts does not grow.
Gates: targeted `go test -race -count=1 ./internal/server/ -run 'MCP|ToolSet|Mcpa'`, `go vet`, `uv run inv core.test`, `interface.test`, `interface.typecheck`, max-lines, `diff --check`, and the docs-links test.

## Writers, owned files, proof gate
- **Schema: none.** Audit goes to `log_entries`, and there is no token purpose.
- **Sequencing:** W1 branches from `dev` *after MCPS W1 merges*. Both touch `mcp.go` (MCPS :97-132 versus MCPA :66-93) and `blocker_copy.go` (181 lines + MCPS codes, cap 385). W1 reuses the MCPS `mcp.RedactToolText`.
- **W1 Go (Opus), about 400 lines including tests:**
  - new `server/mcp_config_authority.go` and `server/mcp_config_authority_test.go`
  - hooks in `mcp_library.go` (190 lines; drop the 202 branches), `mcp.go` `handleMCPDelete` only, and `mcp_toolsets.go` (159)
  - `mcp_governance.go` :31-47 and :219 (258 lines)
  - the `mcp_env_rejected` constant and copy in `blocker_copy.go`
  - identity updates in the five existing test files
- **W2 interface (Sonnet), about 150 lines,** after W1 freezes the codes:
  - `store/cortexStoreMcpSlice.ts` (287)
  - `components/settings/MCPToolRegistry.tsx` (272), `MCPLibraryBrowser.tsx` (294) and `MCPToolSetLayersPanel.tsx` (300)
  - `lib/blockerCopy.ts` (308)
  - their unit tests, `__tests__/store/useCortexStore.resource-registry.test.ts`, and the three e2e specs
- **Docs in the same slice:**
  - `docs/API_REFERENCE.md` rows 165 and 170-176: authority, 401/403/400/404/503, no 202, and env allowlist.
  - `docs/user/resources.md`, `docs/user/governance-trust.md`, and `docs/architecture/FRONTEND.md` (MCP settings copy).
  - The `.state/V8_DEV_STATE.md` row-13 HIGH risk is narrowed to closed for configuration.
- Do not touch: the MCPS files (`mcp_call_authority.go`, `handleMCPToolCall`), A2b tier code, `swarm/tool_scope*.go`, the `mcp/` package, the library yaml, or `.github/workflows`.
- **Live probe (the lead runs it; needs the stack):**
  - As a standard user, Resources → Tools shows no install or delete control, and a direct POST returns 403.
  - As admin, a `fetch` install succeeds and the audit shows `mcp_server_installed` with env keys only.
- Verdict: **CONDITIONAL**. GO once MCPS W1 has merged and the owner answers Q1-Q4.

## Owner questions (recommended default in bold)
- Q1. Scope name: a new `mcp_config:write`, or reuse the advertised `mcp:write`? **`mcp_config:write`. `mcp:write` is a valid agent tool grant string.**
- Q2. Should an approver be able to install credentialed/external entries directly (self-approved tier 2), replacing the dead-end 202? **Yes. Otherwise github/slack/media connectors stay uninstallable, and the 202 misrepresents a pending approval.**
- Q3. Should MCP env secrets move to secret references or encryption at rest? **Not in MCPA. Queue it as a separate schema slice. MCPA only stops leaks in responses, logs and audit.**
- Q4. Should toolset writes need `approvals:decide` because they widen agent grants? **No. `mcp_config:write` plus audit is enough, because widened grants still hit the A2b tier at plan confirm (MCPS D5/S7b).**
