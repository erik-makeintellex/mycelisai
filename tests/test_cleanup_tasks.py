from invoke import Context

from ops import misc
from ops.misc_support import remove_repo_targets


def test_clean_generated_only_removes_allowlisted_paths(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)

    generated = (
        tmp_path / ".venv",
        tmp_path / "interface" / "node_modules",
        tmp_path / "interface" / ".next",
        tmp_path / "workspace" / "tool-cache",
        tmp_path / "interface" / "test-results",
        tmp_path / "interface" / "playwright-report",
        tmp_path / ".pytest_cache",
        tmp_path / "core" / "bin",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", generated)

    for target in generated:
        target.mkdir(parents=True, exist_ok=True)
        (target / "marker.txt").write_text("x", encoding="utf-8")

    runtime_data = tmp_path / "workspace" / "docker-compose" / "data"
    runtime_data.mkdir(parents=True, exist_ok=True)
    (runtime_data / "keep.txt").write_text("keep", encoding="utf-8")

    misc.clean_generated.body(Context())

    output = capsys.readouterr().out
    assert "workspace/docker-compose/data is intentionally untouched." in output
    for target in generated:
        assert not target.exists()
    assert runtime_data.exists()


def test_clean_generated_skips_active_python_environment(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    active_venv = tmp_path / ".venv"
    disposable_cache = tmp_path / "interface" / ".next"
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", (active_venv, disposable_cache))
    monkeypatch.setattr(
        "ops.cleanup_support.sys.executable",
        str(active_venv / "Scripts" / "python.exe"),
    )

    active_venv.mkdir(parents=True)
    (active_venv / "marker.txt").write_text("keep", encoding="utf-8")
    disposable_cache.mkdir(parents=True)
    (disposable_cache / "marker.txt").write_text("remove", encoding="utf-8")

    misc.clean_generated.body(Context())

    output = capsys.readouterr().out
    assert "Skipped active runtime:" in output
    assert ".venv" in output
    assert active_venv.exists()
    assert not disposable_cache.exists()


def test_remove_repo_targets_retries_readonly_cache_files(tmp_path):
    cache_dir = tmp_path / "workspace" / "tool-cache"
    cache_dir.mkdir(parents=True)
    readonly_file = cache_dir / "cached.txt"
    readonly_file.write_text("cache", encoding="utf-8")
    readonly_file.chmod(0o400)

    removed, missing = remove_repo_targets((cache_dir,), tmp_path)

    assert removed == ["workspace/tool-cache"]
    assert missing == []
    assert not cache_dir.exists()


def test_clean_windows_dev_residue_requires_windows(monkeypatch):
    monkeypatch.setattr(misc, "is_windows", lambda: False)

    try:
        misc.clean_windows_dev_residue.body(Context())
    except SystemExit as exc:
        assert "Windows-only" in str(exc)
    else:
        raise AssertionError("expected SystemExit for non-Windows host")


def test_clean_windows_dev_residue_reports_source_only_guidance(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(misc, "is_windows", lambda: True)
    targets = (
        tmp_path / ".venv",
        tmp_path / "interface" / "node_modules",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", targets)

    for target in targets:
        target.mkdir(parents=True, exist_ok=True)
        (target / "marker.txt").write_text("x", encoding="utf-8")

    misc.clean_windows_dev_residue.body(Context())

    output = capsys.readouterr().out
    assert "Windows source-only reminder:" in output
    assert "run install/build/test/compose from the WSL checkout" in output
    for target in targets:
        assert not target.exists()


def test_clean_wsl_handoff_skips_active_python_environment(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    active_venv = tmp_path / ".venv"
    next_cache = tmp_path / "interface" / ".next"
    monkeypatch.setattr(misc, "WSL_HANDOFF_TARGETS", (active_venv, next_cache))
    monkeypatch.setattr(
        "ops.cleanup_support.sys.executable",
        str(active_venv / "Scripts" / "python.exe"),
    )

    active_venv.mkdir(parents=True)
    next_cache.mkdir(parents=True)

    misc.clean_wsl_handoff.body(Context())

    output = capsys.readouterr().out
    assert "Skipped active runtime:" in output
    assert active_venv.exists()
    assert not next_cache.exists()


def test_clean_disk_status_reports_repo_total_and_vhd_guidance(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    targets = (
        tmp_path / ".venv",
        tmp_path / "workspace" / "tool-cache",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", targets)

    first = targets[0]
    first.mkdir(parents=True, exist_ok=True)
    (first / "one.bin").write_bytes(b"a" * 1024)

    second = targets[1]
    second.mkdir(parents=True, exist_ok=True)
    (second / "two.bin").write_bytes(b"b" * 2048)

    misc.clean_disk_status.body(Context())

    output = capsys.readouterr().out
    assert ".venv: present" in output
    assert "workspace/tool-cache: present" in output
    assert "Repo-local generated total:" in output
    assert "Windows should stay source-only" in output
    assert "WSL VHD slack space are outside repo cleanup" in output
    assert "wsl --shutdown" in output
