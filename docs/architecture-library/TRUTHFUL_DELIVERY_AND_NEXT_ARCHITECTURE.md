# Truthful Delivery And Next Architecture

Status: reference and live scoreboard pointer. Authority: [canonical PRD](MYCELIS_CANONICAL_PRD.md), especially Trust Recovery And Confidence. Evidence: [scoreboard](../../.state/V8_DEV_STATE.md).
Source: owner directive 2026-09-27; mycelis-architect placeholder audit (`audit/placeholder-inventory.md`); live journey probe (`probe_results.md`); architect packets M1, H1, B1-outline (owner-decided 2026-09-27); [market position and workflow targets](MARKET_POSITION_AND_WORKFLOW_TARGETS.md).

## The truthfulness invariant

No placeholder or fake functionality ships, ever (owner, 2026-09-27). Mycelis never presents templated or synthesized content standing in for real model or tool output, hard-coded demo data, a stub that returns success, a "not implemented" path that reports OK, a validation that always passes, or a UI state that claims completion without real evidence. If the real behavior cannot happen, the system returns an honest, normalized blocker instead. Proof is not a claim: any confirm-action or team-work "verified" result must come from Core reading its own output back off disk (or resolving the artifact row) and matching it against what was requested — a checksum of the requested content is not proof of what was written. A tool failure (an MCP `IsError` result, an internal-tool error, an inference failure) is always a failure, never silently absorbed into a success path. Writers build the real backend path or stop and report a blocker; reviewers (security-qa, ux, architect, docs) treat any placeholder or fake-success path found in a diff as a merge blocker, with no exception for "we'll fix it later."

## Placeholder audit status

The architect's audit (`audit/placeholder-inventory.md`, dev @244af327) found 5 P0 (false success on a primary journey), 11 P1 (name-only or failure-reported-as-success), and 7 P2 (cosmetic/latent) findings.

**Fixed and merged:**
- **D2 / PH-A** — chat proposal honesty. Removed the deterministic keyword-matched proposal path that skipped Soma and the model entirely; removed server-templated "research and council" documents and the hard-coded marketing domain template.
- **D3** — real draft preview in the proposal when content comes from agent output.
- **PH-B** — proof truth. `execution_output_proof.go` and `team_work_signal_proof.go` no longer hard-code `verified`; each output path is read back and hashed off disk, with `Degraded`/`output_missing` when a referenced file is absent.
- **PH-C** — tool failure truth. MCP `IsError` results now surface as failures; the dormant runtime-owned HTML fallback package and the PROOF.md/README.md synthesis in `write_file` are removed; internal tools that previously returned success with a nil error on failure now return errors.
- **UX1** — interface-side fixes: the false "All Clear" approvals state, the chat-card `token_wrong_purpose` regression on the team-launch path, and related UX copy.
- **PH-E** (fixed in worktree `ph-e-docs`, pending merge) — review-loop and cosmetic findings. `review_loops.go`'s structure check (department/specialist/response-style counts against fixed thresholds) no longer claims a review happened: findings/activity wording now says "structure check" and drops the "active review owner" framing (JSON field names unchanged). `swarm/axon.go` no longer silently defaults an unrouted signal to a nonexistent "genesis" team; it checks `Soma.HasTeam` and returns an honest error when no team is running (no ask-routing/intent classifier exists yet to do better). `organization_normalization.go` no longer invents department names/IDs from a bare `DepartmentCount`; the list stays empty and the count is exposed as-is. `services.go`'s Ollama row now probes the adapter (`Probe`, 2s timeout) instead of reporting "online" on adapter-exists alone. `cognitive/mock.go`'s `MockAdapter` moved to a `_test.go` file (memory package tests now carry their own local mock since Go doesn't export test-file symbols across packages). `memory/archivist.go`'s SitRep JSON parse failure now returns an honest error instead of storing the raw model text as the summary.

**Remaining, not yet started:**
- **PH-D** — name-only surfaces: the connector-install "simulate success" path (`registry/service.go`), the hard-coded `/api/v1/sensors` feed list, the heartbeat-only Symbiotic Sensors seed (which also commits without a confirm token — flagged to security-qa), and mission-commit reporting `active` when nothing activated.
- **The missions row stored "active"** — `server/mission_commit.go` returns `Status:"active"` even when `activateCommittedMission` returned nil; part of PH-D.
- **`read_file` scoping** — currently workspace-wide with no team boundary; deferred as a named follow-up slice out of H1 (owner-confirmed Q3).
- **Confirm-authority persistence** and **cancel-action ownership** — narrower authority-edge findings tracked on the scoreboard, not yet re-verified live.
- **The `webAuth` test flake** — fixed: the "tampered last character" forged-token case in `webAuth.test.ts` toggled the signature's last base64url character between fixed `"A"`/`"B"` values, but that character's low 2 bits are padding discarded on decode, so ~1/16 of the time (whenever the timestamp-dependent HMAC signature happened to end in `A`/`B`/`C`/`D`) the "tamper" decoded to the same signature bytes and the forged token verified successfully, failing the test. `webAuth.ts` itself was already correct (constant-time HMAC verify, constant-time password compare). Test now flips a significant top-4 bit of that character so the decoded byte always changes.

## The live journey probe as the delivery metric

`scratchpad/live_probe.py` runs 11 end-to-end journeys (J1-J5, with sub-steps) against a live dev instance and is the delivery metric for truthful delivery, not a proxy count of merged slices. Latest run, dev `01590f14` (D2 + PH-C + PH-B merged), 2026-09-27: **6/11 PASS.**

- PASS J1 (ask a question). Minor: Soma still describes Mycelis as a "neural organism" (prompt wording, not a truthfulness defect).
- FAIL J2a: no draft preview when content comes from agent output — fixed by D3 (in progress at probe time).
- PASS J2b-e: real content on disk, correct content type, honest verified status.
- FAIL J3a-c: memory save returns a 400 "embedding failed"; recall is generic with no citation — targeted by M1 (below).
- FAIL J4: `/api/v1/sensors` hard-codes "gmail inbox: online" plus nine other feeds — targeted by PH-D.
- PASS J5: web search returns real results with a verification notice.

The probe must be rerun after each of D3, M1, and PH-D land, and the pass count is what goes on the scoreboard — not "the slice merged."

## Planned architecture (owner-decided, 2026-09-27)

### M1 — memory without embeddings

Deployment-context save is not atomic today and lies on failure (an orphan artifact can claim a false vector count); every recall path is silently dead when the embedding provider is unavailable, which it is in the current deployment. M1 makes lexical recall the honest floor:
- **Atomic, honest save.** One transaction writes the artifact and chunk rows with `embedding=NULL` and `embedding_status="pending"`; a best-effort embed follows and updates each row on success. A failed transaction leaves no orphan.
- **An embedding-availability probe** with a negative cache, so a dead embedding stack costs zero round trips per turn.
- **One governed lexical-recall function** (PostgreSQL full-text search, `ts_rank_cd`, stop-word aware) used by every recall site, returning a mode of `semantic`, `keyword`, or `hybrid`.
- **Opportunistic backfill** plus an admin endpoint; no daemon.
- **Honest status** surfaced to the UI: "Saved. Soma can recall this by keywords; semantic search needs an embedding engine."
- **Scope on recall**: non-lead agents excluded from restricted/operating-context classes; goal-scoped rows excluded without an intersecting goal set.
- **Citation.** Recall carries `ContextSourceRef` (artifact id, title, class, retrieval mode, `used`) through to `ChatResponsePayload.context_sources`; `used=true` only on genuine shingle overlap with the answer, never from template text.
- **Work log renames.** `diary_entry`/`diary` are relabeled `worklog_entry`/`worklog` ("Work log entry" / "Work log"); old values remain accepted write aliases and are normalized. Other taxonomy values are relabeled only, with stored values kept, per the owner's accepted defaults (Q1-Q6 in the M1 packet).

### H1 — team handoffs

Today, a created artifact's provenance is hard-coded to `agent_id='internal'`, the Exchange actor role comes from unauthenticated model-supplied args (`role:"admin"` grants admin), there is no handoff channel, and delegation is fire-and-forget with no work item — so a receiving team's signals are silently dropped as uncorrelated. H1 builds the smallest real flow:
- **Provenance at creation** comes from the invocation context (agent/team/run), never from tool args.
- **Exchange actor from identity, not args** — closes the forged-admin path.
- **A new `hand_off` tool** (lead-only by default): verifies artifact ownership and the target team in one transaction, publishes on a new `organization.team.handoffs` channel, creates a target-team work item, and stages outbox dispatch. It returns `queued`, never `delivered`.
- **Governance tier.** A handoff runs at `medium` (no new approval) only when the target team already has a work item on the same run; otherwise it returns `handoff_needs_approval` and Soma proposes it. Restricted-sensitivity content always needs approval (owner-confirmed).
- **A new `read_handoff_input` tool** scoped to the target team and the handoff's artifact refs; anything else is `handoff_input_not_in_scope`.
- **Use is provable**, not inferred from a read: `handoff_consumed` fires only when the target work item completes with output refs.
- Owner-confirmed: handoffs outside the approved run are blocked with approval required; only team leads get the handoff tools; `read_file` scoping to a team's group is a separate follow-up slice; the Work-lane chain link is added by the U1 owner after U1 merges.

### B1 — token budgets (planned, not started)

Closes the Paperclip budgets gap. Every agent execution (Soma turn, team run, specialist/council consult, draft turn) runs under a token budget resolved `agent -> team -> profile -> model/provider default -> global default`, with defaults keyed to model class (local small, local large, hosted premium) and the context window as a ceiling. Enforcement sits at the cognitive router, covering prompt plus completion across ReAct iterations; a budget hit stops the run with an honest `token_budget_exhausted` blocker and never presents a partial result as complete, with an 80%-threshold warning. Accounting uses real provider-reported usage. Only root admin plus `cognitive:write` may change budgets, and every change is audited. Sequencing: starts after M1, PH-D, and H1 merge; needs a short architect packet, then a Go writer, an interface writer, security QA, and a live probe journey J6 (a budget hit gives an honest stop).
