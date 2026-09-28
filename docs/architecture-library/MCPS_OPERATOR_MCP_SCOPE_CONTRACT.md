# MCPS Operator MCP Scope Contract

Owner: Core authority and MCP runtime. Product authority remains the [canonical PRD](MYCELIS_CANONICAL_PRD.md).

> Frozen (owner defaults applied 2026-09-28). Contract for scoping the operator-direct MCP tool call route (architect draft 2026-09-27; Q1-Q3 answered with the recommended defaults). Slice id `MCPS`. Base `dev` @b9d1e172 (verified unchanged for every cited file at `dev` 9cdb84ac). Origin: S7b follow-up (`af56af15`), "The operator MCP call route is gated by route auth only, not by agent scope."

## Verified today (source facts)
- **One direct route exists:** `POST /api/v1/mcp/servers/{id}/tools/{tool}/call` (server/admin_routes.go:188, handler server/mcp.go:97-132). There is no SSE, streaming, BFF `app/api` or second Core variant. `GET /api/v1/mcp/tools` (:189) and `GET /api/v1/mcp/servers` (:186) only list tools. Raw install is disabled (:181-185). All other `CallTool` callers are agent or approved-plan paths that are already scoped: swarm/tool_executor.go:169 and server/templates_execution.go:247 behind `newApprovedPlanToolGuard`.
- **Auth gate: authentication only.** Core wraps the whole mux in `AuthMiddleware` (cmd/server/main.go:72, server/auth.go:255-312), which resolves an identity and nothing more. The handler reads no identity, role or scope (mcp.go:97-132). The BFF proxy strips inbound authority headers and forwards the session (interface/proxy.ts:6-11, 44-60). Its only role gate is on page URLs (`/system` and some `/settings` tabs, proxy.ts:38, 86-91), never on `/api/*`.
- **Principals that reach it:**
  - A standard web user is forwarded as role `operator` with scopes `soma:work, runs:read, outputs:read` (auth.go:210-230).
  - A web admin is forwarded as `admin` with `*`.
  - The local API key and the break-glass key resolve to `admin`/`owner` with `*` (auth.go:85-105).
  - There are no scoped admins today: every admin carries `*`. The scope checks below still matter for future scoped identities.
- **What it reaches:**
  - Any server connected in `ClientPool`, and any tool name, because the name goes straight to the server (mcp/pool.go:188-210).
  - The tool is not checked against the discovered tool cache (`Service.ListTools`, mcp/service_lookup.go:11).
  - Servers have no owner or group field (mcp/service_types.go:11-24).
  - The pool holds config-bootstrapped servers plus library installs. `handleMCPLibraryInstall` and `handleMCPLibraryApply` have no role gate either (server/mcp_library.go:56, :93).
  - Curated servers include filesystem, github, postgres, sqlite, fetch, slack, puppeteer and paid media APIs (core/config/mcp-library.yaml).
  - The internal tool registry (`InternalServerID`, swarm/tool_executor.go:12) is not in the pool.
- **Risk, approval and cost:** the handler does not classify risk, ask for confirmation or pass through A2b. The fail-closed classifier `capabilityRiskForTool` (server/governance_tool_risk.go:35-41) rates every `mcp:*` name `high`, and A2b maps high to tier 2 (root admin + `approvals:decide`, `approverTier` governance_approver_tier.go:74-91; `isApprover` governance_authority.go:181-183). `approverCostCeiling = 5.0` (:69) applies to inference-cost estimates. A direct MCP call has no inference cost, but it can spend on external paid APIs. `chatToolRisk` (server/cognitive_tool_risk.go:4-17) is display-only and does not apply here.
- **Audit:** no audit or mission event is written.
  - The only record is an Exchange item (mcp.go:147-157) whose `Result.arguments` holds the raw arguments (mcp.go:181), so there is no redaction.
  - The handler never calls `exchangeContext` (server/exchange.go:12-21), so the actor defaults to role `mcp` (exchange/normalization.go:119-121) with no human attribution.
  - Error text is redacted and capped by `redactToolErrorText` (mcp/error_redaction.go:24).
  - Precedent for a fail-closed audit is B1: `createAuditEvent` (templates_audit.go:15), with a 503 when there is no audit id (token_budgets.go:272-278).
- **Interface callers:**
  - `WorkspaceExplorer.tsx:63-83` (the Resources → Workspace tab) calls filesystem `list_directory` (:89), `read_text_file` (:115), `create_directory` (:206) and `write_file` (:222). It is enabled whenever the server is connected (:53), with no role check.
  - `ToolsPalette.tsx:38-55` calls any listed MCP tool with `{arguments:{}}`.
  - Four e2e specs stub the route (resources-workspace-files, soma-media-retained-output-live, first-demo-success, support/trusted-outcome-journey).
- **Read precedent:** `HandleWorkspaceFileView` (server/workspace_files.go:121) serves any workspace file to any authenticated caller.

## Threat model
| Principal | Can today | Should |
|---|---|---|
| Standard user (`operator`) | Call any tool on any connected server: write or move workspace files, run github/slack/postgres writes, trigger paid media APIs, and reach anything puppeteer or fetch can reach. Nothing is attributed. | Read the workspace through the filesystem read allowlist only. Anything else goes through a Soma proposal (S7b scoped plan with the A2b approver). |
| Scoped admin (no `approvals:decide`) | Same as the standard user. | Same as the standard user. Direct high-risk calls need tier-2 authority. |
| Root admin / API key / break-glass (`*`) | Everything, unaudited. | Everything, with a fail-closed audit record first and redacted retention. It counts as a self-approved tier-2 action. |
Other threats:
- Secrets in arguments (a github PAT, `POSTGRES_URL`, bearer headers) are kept verbatim in a `team_scoped` Exchange item.
- An unknown or case-variant tool name reaches the server undeclared.
- The route bypasses the A2b tier because it never mints a scope.
- The cost ceiling is bypassed for paid APIs. D4 keeps that bypass to approver-tier principals only.

## Decisions (one recommended option each)
- **D1. Keep the route. Core enforces, not the BFF.** No new route and no BFF gate. One authority check runs in a new `server/mcp_call_authority.go`, called from `handleMCPToolCall` before `normalizeMCPToolCallArgumentsForServer`. mcp.go stays at or under 385 lines.
- **D2. Resolve before authorizing.**
  - The server must exist (`s.MCP.Get`).
  - `{tool}` must match a discovered tool name for that server exactly and case-sensitively (`s.MCP.ListTools`), with no trimming.
  - Otherwise return 404 `mcp_tool_not_found`. The internal server ID gets the same 404.
  - A pool that is missing or disconnected returns 503 `service_unavailable`.
- **D3. Risk class: one code-owned read allowlist, and everything else high.** `directMCPReadTools` keyed by server name: `filesystem` → `list_directory, list_directory_with_sizes, directory_tree, read_text_file, read_file, read_media_file, read_multiple_files, get_file_info, search_files, list_allowed_directories` (class `low`). Every other tool gets `capabilityRiskForTool("mcp:"+server+"/"+tool, args)` = high. This sits next to `registeredInternalToolRisk`, whose registry-equality test stays untouched. No MCP annotation or model can lower a class.
- **D4. Who may call.**
  - A low call requires `hasScope(identity, "outputs:read")`, which covers standard users, admins and the API key. Missing it returns 403 `mcp_call_forbidden`.
  - A high call requires `isApprover(identity)`. Otherwise `requireRootAdminScope(w, r, scopeApprovalsDecide)` returns 403 `admin_required` with `required_scope`, and the recommended action is "ask Soma to propose it".
  - A nil identity returns 401.
- **D5. No confirm-token round trip on the direct route.** The only principals allowed a high call are the A2b tier-2 approvers, so a proposal would be approved by the same principal. The call is recorded as tier 2, authority `approvals:decide`, `self_approved=true`. A new token purpose or proposal kind would add a second approval path and needs schema, so it is rejected. Non-approvers keep the existing path: a Soma proposal whose `mcp:` refs S7b checks against the agent's declared scope and whose tier is enforced at confirm.
- **D6. Audit and events.**
  - Before every high call, write `createAuditEvent` with `action:"mcp_tool_called"`, server, tool, risk, tier, authority, self_approved, sorted argument **keys** only and `attachActorIdentity`. If the audit id is empty or there is an error, return 503 `service_unavailable` and make no call (the B1 precedent).
  - The Exchange item carries `audit_event_id` and the outcome.
  - Low calls write no audit record because they are reads. The Exchange item (D7) is their record.
  - A high-call refusal writes a best-effort `mcp_tool_call_refused` audit record: non-blocking, with keys only.
- **D7. Retention hygiene.**
  - Publish with `exchangeContext(r)` so the item carries the human actor.
  - Before retention, replace `Result.arguments` with a redacted copy. Sensitive keys (`api_key, authorization, credential, password, access_token, refresh_token, secret, token`, and any key ending `_key/_token/_secret`) become `[REDACTED]`. String values go through the credential patterns in `redactToolErrorText`, exported as `mcp.RedactToolText`.
  - The tool still receives the raw arguments. Responses never echo arguments.
- **D8. Declared scope is rejected as the direct-call model.** Operators are not agents. A standard user who can create a team could declare `mcp:x/*` and so grant themselves the call, which is the escalation S7b closed for plans. Per-server allowlisting (D3) plus the approver tier (D4) is the single rule.
- **D9. Interface.**
  - `WorkspaceExplorer` shows create/write controls only when `/api/v1/user/me` reports an approver, and renders blocker copy otherwise.
  - `ToolsPalette` renders the blocker copy.
  - `blockerCopy.ts` gains `mcp_tool_not_found` and `mcp_call_forbidden`.
- **D10. Out of scope (the follow-up slice MCPA):** role gates on library install/apply, server delete and toolset CRUD (mcp_library.go:56/:93, mcp.go:66, mcp_toolsets.go). They are in the same class of gap, but they are configuration authority rather than call authority.

## Test-first acceptance (-race; fake pool records calls; helper prefix `mcps`)
1. Positive:
   - A standard identity can call filesystem `list_directory` and `read_text_file`: 200, one pool call, and an Exchange actor = user id.
   - An admin `*` can call filesystem `write_file` and github `create_issue`: 200, with an audit record written before the pool call (ordering asserted) and `self_approved=true`.
   - The API key behaves the same as a web admin.
2. Negative:
   - A standard user calling `write_file`, `create_directory`, `move_file`, `edit_file` or any github/slack/fetch tool gets 403 `admin_required` with `required_scope=approvals:decide`, and there are zero pool calls.
   - An admin identity with `outputs:read` but without `approvals:decide` gets the same 403.
   - An identity without `outputs:read` calling a read tool gets 403 `mcp_call_forbidden`.
   - No identity gets 401.
3. Resolution: an unknown server UUID, the internal server ID, a tool missing from the discovered cache, a case variant `Read_Text_File`, a trailing space, and a URL-encoded `/` in `{tool}` each return 404 `mcp_tool_not_found` with zero pool calls.
4. Adversarial:
   - A second server named `filesystem2`, or a non-library server with a read-tool name: the allowlist is keyed on the exact server name, so it is high.
   - Body fields `actor`, `confirm_token`, `risk` and `self_approved` are ignored.
   - Forged `X-Mycelis-Web-Identity` without a valid signature returns 401 (existing test, rerun).
   - A header that claims admin while the session is standard is stripped by the BFF (proxy test).
5. Audit fail-closed: with a nil DB or a failing audit insert, a high call returns 503 `service_unavailable` with zero pool calls. A low call still succeeds.
6. Redaction: `{"token":"t","env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"g"},"url":"https://u:p@h"}` is stored as `[REDACTED]` in the Exchange item, `u:p` is removed, and the tool receives the raw values. Neither the audit record nor the response contains any argument value.
7. Existing mcp_test.go and mcp_tool_failure_test.go still pass (the isError → 502 path is unchanged).
8. Interface:
   - `WorkspaceExplorer` hides create/write for non-approvers and shows the blocker copy on a 403.
   - A `blockerCopy` snapshot covers both new codes for user and admin.
   - The four e2e route stubs are updated where they assert write success.
Gates: targeted `go test -race -count=1 ./internal/server/ ./internal/mcp/`, `go vet`, `uv run inv core.test`, `interface.test`, `interface.typecheck`, max-lines, `diff --check`, and the docs-links test.

## Writers, owned files, proof gate
- **Schema: none.** Audit uses `log_entries` through `createAuditEvent`, and there is no new token purpose.
- **W1 Go (Opus), about 450 lines including tests:**
  - new `server/mcp_call_authority.go` and `server/mcp_call_authority_test.go`
  - a hook of at most 15 lines in `server/mcp.go`
  - argument redaction in `server/mcp_redaction.go`
  - the exported wrapper in `mcp/error_redaction.go`
  - the new code constants and copy in `server/blocker_copy.go` (181 lines)
- **W2 interface (Sonnet), about 150 lines,** after W1 freezes the codes: `lib/blockerCopy.ts`, `components/resources/WorkspaceExplorer.tsx` (298 lines), `components/workspace/ToolsPalette.tsx`, their unit tests, and the e2e stubs that name the call route. It must not grow `e2e/specs/mcp-connected-tools.spec.ts` (legacy cap 515).
- **Docs in the same slice:**
  - `docs/API_REFERENCE.md` row 168: authority, 401/403/404/503, and redaction.
  - `docs/user/resources.md` and `docs/user/governance-trust.md`.
  - The known-gap note in `docs/architecture/BACKEND.md` (net-zero).
  - The `.state/V8_DEV_STATE.md` S7b row follow-up closed.
- Do not touch: A2b tier code, `registeredInternalToolRisk`, S7b plan-scope files, or `.github/workflows`.
- **Proof gate:** a mutation check against pre-MCPS `mcp.go` must fail tests 2, 3, 5 and 6.
- **Live probe (the lead runs it; needs the stack):**
  - As a standard user, Workspace lists and reads work, while create and write show the admin-required copy and no file appears.
  - As admin, a write succeeds and the audit shows `mcp_tool_called`.
- Verdict: **CONDITIONAL**. GO once the owner answers Q1–Q3.

## Owner questions (recommended default in bold)
- Q1. Should standard users keep workspace create/write through Resources → Workspace? **No. They read directly and write through a Soma proposal. Only approvers write directly.**
- Q2. Should the direct read allowlist cover more than filesystem reads (for example fetch, or github read tools)? **No, filesystem reads only in MCPS. Widen later by adding code entries with tests.**
- Q3. Should MCPA (role gates on MCP install/apply/delete/toolsets) be queued right after MCPS? **Yes, as the next Opus slice. It is the same class of gap.**
