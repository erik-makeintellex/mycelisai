"""Per-profile route health for `cognitive.status` and `compose.health`.

Core reports every profile's effective route on GET /api/v1/cognitive/status
(`available`, `code`, `source`, `override_origin`, `db_row_present`,
`reachable`). Both tasks apply the same severity table:

- FAIL: an execution profile that is unavailable (unbound, missing, disabled,
  model missing, uninitialized) or whose provider did not answer its probe;
  a missing or malformed `profiles` payload (never reported as OK).
- WARN: the same for any other bound profile; an explicit fallback in use;
  `overlay_error` (DB overrides may be missing); a DB role.* row hidden under
  an env override.
- NOTE: a hosted provider that Core does not probe (`reachable: null`).

This module never mutates config and never prints endpoints or secrets.
"""

from .cognitive_root import EXECUTION_PROFILES

_CODE_STATES = {
    "provider_disabled": "DISABLED",
    "provider_missing": "MISSING",
    "model_missing": "MODEL_MISSING",
    "provider_uninitialized": "UNINITIALIZED",
    "no_provider_available": "UNBOUND",
    "profile_unbound": "UNBOUND",
}


def _label(route):
    source = str(route.get("source") or "unbound")
    origin = route.get("override_origin")
    return f"override:{origin}" if source == "override" and origin else source


def _route_findings(name, route):
    """Return (state, [(severity, message)]) for one profile route."""
    execution = name in EXECUTION_PROFILES
    hard = "FAIL" if execution else "WARN"
    findings = []
    if not isinstance(route, dict) or not isinstance(route.get("available"), bool):
        return "MALFORMED", [(hard, f"profile {name}: malformed route entry (no availability reported)")]
    provider = route.get("provider_id") or "-"
    label = _label(route)
    reachable = route.get("reachable")
    hint = str(route.get("recommended_action") or "").strip()
    state = "OK"
    if not route["available"]:
        code = str(route.get("code") or "unavailable")
        state = _CODE_STATES.get(code, code.upper())
        findings.append((hard, f"profile {name} -> {provider} [{label}] {state}" + (f": {hint}" if hint else "")))
    elif reachable is False:
        state = "UNREACHABLE"
        findings.append((hard, f"profile {name} -> {provider} [{label}] did not answer its health probe"))
    elif reachable is None:
        state = "OK (not probed)"
        findings.append(("NOTE", f"profile {name} -> {provider} is hosted or a gateway and is not probed"))
    if route.get("source") == "fallback":
        findings.append(("WARN", f"profile {name} runs on explicit fallback provider {provider}"))
    if route.get("db_row_present") and route.get("override_origin") == "env":
        findings.append(("WARN", f"profile {name} has a stale DB role.{name} row hidden under an env override"))
    return state, findings


def profile_route_findings(payload):
    """Return (rows, findings): rows are (name, provider, label, state)."""
    profiles = payload.get("profiles") if isinstance(payload, dict) else None
    if not isinstance(profiles, dict) or not profiles:
        return [], [("FAIL", "Core status has no per-profile routes; profile routing is not verified")]
    rows, findings = [], []
    for name in EXECUTION_PROFILES:
        if name not in profiles:
            findings.append(("FAIL", f"profile {name} is missing from Core status"))
    for name, route in sorted(profiles.items()):
        state, route_findings = _route_findings(name, route)
        provider = (route.get("provider_id") if isinstance(route, dict) else None) or "-"
        label = _label(route) if isinstance(route, dict) else "?"
        rows.append((name, provider, label, state))
        findings.extend(route_findings)
    if payload.get("overlay_error") is True:
        findings.append(("WARN", "Core could not read the DB overlay at startup; role.* overrides may be missing"))
    if payload.get("profile_route_health") == "failed" and not any(sev == "FAIL" for sev, _ in findings):
        findings.append(("FAIL", "Core reports profile_route_health=failed"))
    return rows, findings


def print_profile_routes(payload, prefix="    "):
    """Print one line per profile plus WARN/NOTE lines; return FAIL messages."""
    rows, findings = profile_route_findings(payload)
    for name, provider, label, state in rows:
        print(f"{prefix}{name:<10} {provider:<20} [{label}] {state}")
    for severity, message in findings:
        if severity != "FAIL":
            print(f"{prefix}[{severity}] {message}")
    return [message for severity, message in findings if severity == "FAIL"]
