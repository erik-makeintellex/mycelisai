from __future__ import annotations

import json

import pytest
from invoke import Context
from invoke.exceptions import Exit

from ops import cognitive, cognitive_profiles, cognitive_root, compose_probe


def _route(**overrides):
    route = {"provider_id": "vllm", "source": "root", "available": True, "code": "available", "reachable": True}
    route.update(overrides)
    return route


def _payload(overlay_error=False, **profile_overrides):
    profiles = {name: _route() for name in cognitive_root.EXECUTION_PROFILES}
    for name, route in profile_overrides.items():
        profiles[name] = route
    return {
        "root_provider": "vllm",
        "text": {"status": "online"},
        "profiles": profiles,
        "overlay_error": overlay_error,
    }


def _severities(payload):
    _, findings = cognitive_profiles.profile_route_findings(payload)
    return {severity for severity, _ in findings}


def test_all_root_profiles_are_ok(capsys):
    assert cognitive_profiles.print_profile_routes(_payload()) == []
    out = capsys.readouterr().out
    assert "coder      vllm                 [root] OK" in out
    assert "[WARN]" not in out


@pytest.mark.parametrize(
    "route,state",
    [
        (_route(provider_id="local-ollama-dev", source="override", override_origin="db", available=False,
                code="provider_disabled", reachable=None,
                recommended_action="Reset the coder override to root (DELETE ...) or enable provider local-ollama-dev."),
         "DISABLED"),
        (_route(reachable=False), "UNREACHABLE"),
        (_route(provider_id="", source="unbound", available=False, code="no_provider_available", reachable=None), "UNBOUND"),
    ],
)
def test_broken_execution_profile_fails(route, state, capsys):
    failures = cognitive_profiles.print_profile_routes(_payload(coder=route))
    out = capsys.readouterr().out
    assert len(failures) == 1 and failures[0].startswith("profile coder")
    assert f"] {state}" in out


def test_disabled_override_line_names_origin(capsys):
    route = _route(provider_id="local-ollama-dev", source="override", override_origin="db", available=False,
                   code="provider_disabled", reachable=None)
    cognitive_profiles.print_profile_routes(_payload(coder=route))
    assert "local-ollama-dev     [override:db] DISABLED" in capsys.readouterr().out


@pytest.mark.parametrize(
    "payload",
    [
        _payload(coder=_route(source="fallback", provider_id="ollama")),
        _payload(overlay_error=True),
        _payload(coder=_route(source="override", override_origin="env", db_row_present=True)),
        _payload(embed=_route(available=False, code="provider_disabled", reachable=None)),
    ],
    ids=["fallback", "overlay_error", "db_row_under_env", "non_execution_disabled"],
)
def test_warn_conditions_do_not_fail(payload):
    assert cognitive_profiles.print_profile_routes(payload) == []
    assert "WARN" in _severities(payload)


def test_hosted_not_probed_is_a_note():
    payload = _payload(creative=_route(reachable=None))
    assert cognitive_profiles.print_profile_routes(payload) == []
    assert _severities(payload) == {"NOTE"}


@pytest.mark.parametrize(
    "payload",
    [
        {"root_provider": "vllm"},
        {"profiles": []},
        {"profiles": {}},
        {"profiles": {"chat": _route()}},
        _payload(coder={"provider_id": "vllm", "source": "root"}),
        _payload(coder="vllm"),
    ],
    ids=["missing", "list", "empty", "missing_execution_profiles", "no_availability", "not_a_dict"],
)
def test_malformed_or_missing_profiles_fail(payload):
    assert cognitive_profiles.print_profile_routes(payload) != []


def test_core_reported_failure_is_never_ok():
    payload = {**_payload(), "profile_route_health": "failed"}
    assert cognitive_profiles.print_profile_routes(payload) == ["Core reports profile_route_health=failed"]


def test_compose_health_fails_on_disabled_execution_profile(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    failures: list[str] = []
    payload = _payload(coder=_route(provider_id="local-ollama-dev", source="override", override_origin="db",
                                    available=False, code="provider_disabled", reachable=None))
    compose_probe._append_cognitive_health_failures("url", json.dumps(payload), failures, {"MYCELIS_ROOT_PROVIDER": "vllm"})
    assert any(f.startswith("Profile Route: profile coder -> local-ollama-dev [override:db] DISABLED") for f in failures)


def test_compose_health_passes_with_warnings_only(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    failures: list[str] = []
    payload = _payload(overlay_error=True, coder=_route(source="fallback", provider_id="ollama"))
    compose_probe._append_cognitive_health_failures("url", json.dumps(payload), failures, {"MYCELIS_ROOT_PROVIDER": "vllm"})
    assert failures == []


def test_cognitive_status_exits_nonzero_on_failed_profile(monkeypatch, capsys, tmp_path):
    monkeypatch.setattr(cognitive, "is_windows", lambda: False)
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    monkeypatch.setenv("MYCELIS_ROOT_PROVIDER", "vllm")
    payload = _payload(coder=_route(reachable=False))
    monkeypatch.setattr(cognitive_root, "fetch_core_status", lambda values: (payload, ""))

    with pytest.raises(Exit, match="profile coder"):
        cognitive.status.body(Context())
    assert "UNREACHABLE" in capsys.readouterr().out


def test_cognitive_status_warnings_keep_exit_zero(monkeypatch, capsys, tmp_path):
    monkeypatch.setattr(cognitive, "is_windows", lambda: False)
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    monkeypatch.setenv("MYCELIS_ROOT_PROVIDER", "vllm")
    monkeypatch.setattr(cognitive_root, "fetch_core_status", lambda values: (_payload(overlay_error=True), ""))
    monkeypatch.setattr(cognitive, "_require_supported_local_engine_host", lambda: None)
    monkeypatch.setattr(cognitive, "_load_engine_config", lambda: {"text": {"port": 1}, "media": {"port": 1}})
    monkeypatch.setattr(cognitive.urllib.request, "urlopen", lambda *a, **k: (_ for _ in ()).throw(OSError("down")))

    cognitive.status.body(Context())
    assert "[WARN] Core could not read the DB overlay" in capsys.readouterr().out
