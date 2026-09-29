from __future__ import annotations

import subprocess

import pytest

from ops import compose, compose_images
from ops import test as test_tasks

REMEDY = "uv run inv compose.down"


class FakeDocker:
    """Fake docker runner keyed by the inspect target, mutable so a recreate can fix state."""

    def __init__(self, containers: dict[str, str | None], tags: dict[str, str | None]):
        self.containers = containers  # container name -> image id it runs (None = not found)
        self.tags = tags  # image tag -> current id (None = image missing)
        self.calls: list[list[str]] = []

    def __call__(self, args: list[str]) -> subprocess.CompletedProcess[str]:
        self.calls.append(list(args))
        assert args[0] == "inspect" or args[:2] == ["image", "inspect"], args
        if args[0] == "inspect":
            value = self.containers.get(args[1])
        else:
            value = self.tags.get(args[2])
        if value is None:
            return subprocess.CompletedProcess(args, 1, stdout="", stderr="No such object")
        return subprocess.CompletedProcess(args, 0, stdout=value + "\n", stderr="")


def _fake(core="sha256:new", interface="sha256:new") -> FakeDocker:
    return FakeDocker(
        {"mycelis-home-core-1": core, "mycelis-home-interface-1": interface},
        {"mycelis-home-core": "sha256:new", "mycelis-home-interface": "sha256:new"},
    )


def test_check_reports_current_when_ids_match(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake())
    results = compose_images.check_app_images()
    assert [(r.service, r.state) for r in results] == [("core", "current"), ("interface", "current")]


def test_check_reports_stale_when_container_runs_old_image(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake(interface="sha256:old"))
    states = {r.service: r.state for r in compose_images.check_app_images()}
    assert states == {"core": "current", "interface": "stale"}


def test_check_reports_missing_image_and_missing_container(monkeypatch):
    fake = _fake()
    fake.tags["mycelis-home-core"] = None
    fake.containers["mycelis-home-interface-1"] = None
    monkeypatch.setattr(compose_images, "_docker", fake)
    states = {r.service: r.state for r in compose_images.check_app_images()}
    assert states == {"core": "missing-image", "interface": "missing-container"}


def test_require_current_passes_and_names_each_container(monkeypatch, capsys):
    monkeypatch.setattr(compose_images, "_docker", _fake())
    compose_images.require_current("compose.health")
    out = capsys.readouterr().out
    assert "mycelis-home-core-1" in out and "mycelis-home-interface-1" in out
    assert "[OK]" in out


def test_require_current_fails_with_remedy_for_stale(monkeypatch, capsys):
    monkeypatch.setattr(compose_images, "_docker", _fake(core="sha256:old"))
    with pytest.raises(SystemExit) as exc:
        compose_images.require_current("compose.health")
    assert exc.value.code != 0
    out = capsys.readouterr().out
    assert "[FAIL]" in out and "mycelis-home-core-1" in out
    assert REMEDY in str(exc.value) and "uv run inv compose.up --build" in str(exc.value)


def test_require_current_fails_honestly_on_missing_image(monkeypatch, capsys):
    fake = _fake()
    fake.tags["mycelis-home-interface"] = None
    monkeypatch.setattr(compose_images, "_docker", fake)
    with pytest.raises(SystemExit):
        compose_images.require_current("compose.health")
    assert "image mycelis-home-interface not found" in capsys.readouterr().out


def test_require_current_allow_stale_warns_but_passes(monkeypatch, capsys):
    monkeypatch.setattr(compose_images, "_docker", _fake(core="sha256:old"))
    compose_images.require_current("test.probe", allow_stale=True)
    out = capsys.readouterr().out
    assert "[STALE]" in out and "--allow-stale" in out


def test_ensure_current_after_up_recreates_only_stale_service_then_passes(monkeypatch):
    fake = _fake(interface="sha256:old")
    monkeypatch.setattr(compose_images, "_docker", fake)
    recreated: list[list[str]] = []

    def recreate(services: list[str]):
        recreated.append(services)
        fake.containers["mycelis-home-interface-1"] = "sha256:new"

    compose_images.ensure_current_after_up(recreate)
    assert recreated == [["interface"]]


def test_ensure_current_after_up_fails_when_still_stale(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake(core="sha256:old"))
    with pytest.raises(SystemExit) as exc:
        compose_images.ensure_current_after_up(lambda services: None)
    assert REMEDY in str(exc.value)


def test_ensure_current_after_up_does_not_recreate_when_current(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake())
    compose_images.ensure_current_after_up(lambda services: pytest.fail("must not recreate"))


def test_ensure_current_after_up_fails_on_missing_image_without_recreate(monkeypatch):
    fake = _fake()
    fake.tags["mycelis-home-core"] = None
    monkeypatch.setattr(compose_images, "_docker", fake)
    with pytest.raises(SystemExit):
        compose_images.ensure_current_after_up(lambda services: pytest.fail("cannot fix a missing image"))


def test_compose_up_recreates_stale_service_with_force_recreate(monkeypatch):
    fake = _fake(core="sha256:old")
    monkeypatch.setattr(compose_images, "_docker", fake)
    commands: list[list[str]] = []

    def run_compose(args, check=True, env=None):
        commands.append(args)
        if "--force-recreate" in args:
            fake.containers["mycelis-home-core-1"] = "sha256:new"

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda values: values)
    monkeypatch.setattr(compose, "_run_compose", run_compose)
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: None)
    monkeypatch.setattr(compose, "_wait_for_port", lambda *a, **k: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda **k: True)
    monkeypatch.setattr(compose, "_wait_for_http_ok", lambda *a, **k: True)
    monkeypatch.setattr(compose.status, "body", lambda _c=None: None)

    compose.up.body(None, build=True)

    assert commands[-1] == compose._compose_command("up", "-d", "--force-recreate", "--no-deps", "core")
    assert not any("volumes" in part for cmd in commands for part in cmd)


def test_compose_health_fails_on_stale_container_before_endpoint_probes(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake(core="sha256:old"))
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda values: values)
    monkeypatch.setattr(compose, "_http_get", lambda *a, **k: pytest.fail("probes must not run on stale app"))
    with pytest.raises(SystemExit) as exc:
        compose.health.body(None)
    assert REMEDY in str(exc.value)


class _FakeContext:
    def __init__(self):
        self.commands: list[str] = []

    def run(self, command, **kwargs):
        self.commands.append(command)


def test_probe_refuses_on_stale_containers(monkeypatch, capsys):
    monkeypatch.setattr(compose_images, "_docker", _fake(interface="sha256:old"))
    ctx = _FakeContext()
    with pytest.raises(SystemExit) as exc:
        test_tasks.probe.body(ctx)
    assert exc.value.code != 0
    assert ctx.commands == []
    assert "mycelis-home-interface-1" in capsys.readouterr().out


def test_probe_runs_when_current(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _fake())
    ctx = _FakeContext()
    test_tasks.probe.body(ctx)
    assert len(ctx.commands) == 1 and "live_journey_probe.py" in ctx.commands[0]


def test_probe_allow_stale_runs_with_warning(monkeypatch, capsys):
    monkeypatch.setattr(compose_images, "_docker", _fake(core="sha256:old"))
    ctx = _FakeContext()
    test_tasks.probe.body(ctx, allow_stale=True)
    assert len(ctx.commands) == 1
    assert "[STALE]" in capsys.readouterr().out
