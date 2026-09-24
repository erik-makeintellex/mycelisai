# Post-G4 Delivery and GUI Verification Plan

Status: `COMPLETE` for the bounded GUI slice; broader release certification remains separate. Current evidence and deployment identity are in the [scoreboard](../../.state/V8_DEV_STATE.md).
Authority: [canonical PRD](MYCELIS_CANONICAL_PRD.md), especially P0.10; [bounded G4/E10 contract](G4_E10_DURABLE_INVOCATION.md).
This plan sequences accepted work and proof; it does not expand product authority or certify deferred paths.

## Bounded GUI result

The delivered Compose UI at `127.0.0.1:3000` passed the scoped new-user, docs, Soma/Work, output, mobile and session journeys recorded in the scoreboard. Deterministic route fixtures test presentation behavior, not live generated Outcome success. Required controls and the compact composer are reachable; the repaired compact empty state keeps its guidance within the scroll viewport. Independent source and evidence QA returned GO for this bounded slice.

Unverified gates remain: novice/tenant role coverage, session expiry and replay, full accessibility, live generated Outcome, cross-browser release proof and promotion. The next deployment gate is private, disabled-by-default P0.10 B2; C through F remain separate execution/trust/adapter gates. No worker executor is active.

## Delivery order

1. Keep the proven central authority/finalization path and G4/E10 counting ledger intact.
2. Package and prove B2 private Runs startup with scoped credentials, TLS, separate persistence, disabled defaults and isolated restart.
3. Freeze C identity, grant, invocation, origin and response-loss contracts before post-commit external create. D replay/projection, E controls/output and F LangGraph certification follow their PRD order.
4. Require a separate live Ask → Approve → Execute → Open → Interact → Proof → Revisit journey before release claims.
