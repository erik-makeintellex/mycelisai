from __future__ import annotations

from pathlib import Path

import pytest

from ops import compose, compose_framework_runs


TOKEN = "runs-only-token-0123456789abcdef0123456789abcdef"
CORE_KEY = "core-only-token-0123456789abcdef0123456789abcdef"
DB_PASSWORD = "runs-fixture-password"


def write_secret_source(root: Path, values: dict[str, str]) -> None:
    (root / ".env").write_text(
        "\n".join(f"{key}={values[key]}" for key in compose_framework_runs.SECRET_INPUT_KEYS) + "\n",
        encoding="utf-8",
    )


def fixture_values(root: Path) -> dict[str, str]:
    for name in ("cert.crt", "cert.key", "ca.crt"):
        (root / name).write_text("fixture", encoding="utf-8")
    values = {
        "FRAMEWORK_RUNS_CORE_TOKEN": TOKEN,
        "FRAMEWORK_RUNS_DB_USER": "runner",
        "FRAMEWORK_RUNS_DB_PASSWORD": DB_PASSWORD,
        "FRAMEWORK_RUNS_DB_NAME": "runs",
        "FRAMEWORK_RUNS_DATABASE_URL": (
            f"postgres://runner:{DB_PASSWORD}@framework-runs-postgres:5432/runs?sslmode=disable"
        ),
        "FRAMEWORK_RUNS_TLS_CERT_HOST_FILE": str(root / "cert.crt"),
        "FRAMEWORK_RUNS_TLS_KEY_HOST_FILE": str(root / "cert.key"),
        "FRAMEWORK_RUNS_TLS_CA_HOST_FILE": str(root / "ca.crt"),
        "MYCELIS_API_KEY": CORE_KEY,
        "DB_PASSWORD": "separate-core-password",
    }
    write_secret_source(root, values)
    return values


def test_compose_overlay_is_explicit_and_default_command_is_unchanged(monkeypatch):
    monkeypatch.delenv(compose_framework_runs.OPT_IN_KEY, raising=False)
    default = compose._compose_command("ps")
    assert default[-3:] == ["-f", str(compose.COMPOSE_FILE), "ps"]
    assert "framework-runs.yml" not in " ".join(default)

    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "1")
    selected = compose._compose_command("ps")
    assert selected[-5:] == [
        "-f", str(compose.COMPOSE_FILE), "-f",
        str(compose_framework_runs.overlay_file(compose.COMPOSE_FILE)), "ps",
    ]


@pytest.mark.parametrize("raw", ["maybe", "enabled", " 1 typo"])
def test_compose_overlay_rejects_ambiguous_opt_in(monkeypatch, raw):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, raw)
    with pytest.raises(SystemExit, match="must be explicitly true or false"):
        compose._compose_command("ps")


def test_compose_overlay_requires_scoped_secret_source_and_exact_database(tmp_path, monkeypatch):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "true")
    values = fixture_values(tmp_path)
    compose_framework_runs.validate(values, root_dir=tmp_path)

    bad = dict(values, FRAMEWORK_RUNS_DATABASE_URL=values["FRAMEWORK_RUNS_DATABASE_URL"].replace(
        "framework-runs-postgres", "postgres",
    ))
    write_secret_source(tmp_path, bad)
    with pytest.raises(SystemExit, match="must match its dedicated Compose database"):
        compose_framework_runs.validate(bad, root_dir=tmp_path)

    bad = dict(values, FRAMEWORK_RUNS_DB_PASSWORD=values["DB_PASSWORD"])
    write_secret_source(tmp_path, bad)
    with pytest.raises(SystemExit, match="separate from Core PostgreSQL"):
        compose_framework_runs.validate(bad, root_dir=tmp_path)

    bad = dict(values, FRAMEWORK_RUNS_CORE_TOKEN=values["MYCELIS_API_KEY"])
    write_secret_source(tmp_path, bad)
    with pytest.raises(SystemExit, match="canonical, scoped"):
        compose_framework_runs.validate(bad, root_dir=tmp_path)

    (tmp_path / ".env").write_text("FRAMEWORK_RUNS_CORE_TOKEN=only-one-secret\n", encoding="utf-8")
    with pytest.raises(SystemExit, match="requires scoped token"):
        compose_framework_runs.validate(values, root_dir=tmp_path)


def test_compose_overlay_requires_existing_tls_files(tmp_path, monkeypatch):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "1")
    values = fixture_values(tmp_path)
    values["FRAMEWORK_RUNS_TLS_KEY_HOST_FILE"] = str(tmp_path / "missing.key")
    with pytest.raises(SystemExit, match="TLS input FRAMEWORK_RUNS_TLS_KEY_HOST_FILE"):
        compose_framework_runs.validate(values, root_dir=tmp_path)


@pytest.mark.parametrize("key", compose_framework_runs.SECRET_INPUT_KEYS)
def test_compose_overlay_rejects_quoted_scoped_secrets(tmp_path, monkeypatch, key):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "1")
    values = fixture_values(tmp_path)
    values[key] = f'"{values[key]}"'
    write_secret_source(tmp_path, values)
    with pytest.raises(SystemExit, match="unquoted canonical"):
        compose_framework_runs.validate(values, root_dir=tmp_path)


def test_compose_overlay_compares_runtime_normalized_core_secrets(tmp_path, monkeypatch):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "1")
    values = fixture_values(tmp_path)
    values["MYCELIS_API_KEY"] = f'"{TOKEN}"'
    with pytest.raises(SystemExit, match="canonical, scoped"):
        compose_framework_runs.validate(values, root_dir=tmp_path)

    values = fixture_values(tmp_path)
    values["DB_PASSWORD"] = f"'{DB_PASSWORD}'"
    with pytest.raises(SystemExit, match="separate from Core PostgreSQL"):
        compose_framework_runs.validate(values, root_dir=tmp_path)


def test_compose_secret_like_database_url_prefers_dot_env():
    from ops import compose_env

    assert compose_env.is_secret_key("FRAMEWORK_RUNS_DATABASE_URL")


def test_compose_worker_health_fails_closed_without_leaking_probe_output(monkeypatch, capsys):
    monkeypatch.setenv(compose_framework_runs.OPT_IN_KEY, "1")
    commands: list[list[str]] = []

    def command(*args):
        return list(args)

    def run(args, **_kwargs):
        commands.append(args)
        return type("Result", (), {"returncode": 0 if args[2] == "framework-runs-postgres" else 1,
                                    "stdout": "fixture-private-token", "stderr": "fixture-private-token"})()

    monkeypatch.setattr(compose_framework_runs.subprocess, "run", run)
    with pytest.raises(SystemExit, match="framework-runs is not ready"):
        compose_framework_runs.verify_health(
            {"FRAMEWORK_RUNS_DB_USER": "runner", "FRAMEWORK_RUNS_DB_NAME": "runs"},
            compose_command=command, runtime_env=lambda: {},
        )
    assert len(commands) == 2
    assert "fixture-private-token" not in capsys.readouterr().out
