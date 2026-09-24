# Post-G4 Delivery and GUI Verification Plan

Status: `COMPLETE` for bounded GUI verification and B2 private Compose deployment; live Kubernetes NetworkPolicy enforcement remains REQUIRED. Evidence is in the [scoreboard](../../.state/V8_DEV_STATE.md).
Authority: [canonical PRD](MYCELIS_CANONICAL_PRD.md), especially P0.10; [bounded G4/E10 contract](G4_E10_DURABLE_INVOCATION.md).
This plan sequences accepted work and proof; it does not expand product authority or certify deferred paths.

## Delivered gates

The Compose GUI slice passed bounded real-session and deterministic presentation checks. It does not certify a live generated Outcome, all roles, full accessibility or release promotion.

B2 packages a disabled-by-default private Runs controller in Compose and the canonical Helm chart. The isolated full-Compose proof passed authenticated TLS/readiness, separate PostgreSQL persistence, executorless create rejection, restart/recreation, database failure/recovery, private-network denial and fixture cleanup. `production_ready=false`: no production executor or Core dispatch is enabled. The operator-approved Compose-only private control peer permits worker connectivity to authenticated Core but grants no Core credential or API authority. Kubernetes retains stricter directional policy; Helm render tests do not prove live cluster enforcement. Public internet was not contacted, and B2 network evidence is not C DNS/SSRF certification.

Use `MYCELIS_COMPOSE_FRAMEWORK_RUNS=1` consistently for opt-in Compose up/health/status/logs/down. The first-boot proof is `uv run inv lifecycle.first-boot-proof --isolated --framework-runs --build`; it owns only its disposable project. Scoped token, dedicated database and TLS host files come from `.env` references, never committed values. No candidate-store credential is mounted before an uploader exists.

## SSO transition

Compose now projects the configured Google client and auth policy from root `.env`; callback state and session behavior remain bounded by the existing auth route. This does not activate an external worker or close the live generated Outcome gate.

## Next boundaries

C requires a frozen subject/grant/invocation mapping, post-commit create, pinned origin/trust/address binding, exact response-loss reconciliation and no central fallback after possible acceptance. D replay/projection, E controls/output and F adapter certification remain separate. Live generated Outcome and release gates remain open.
