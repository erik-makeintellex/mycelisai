from pathlib import Path

from ops import config


def test_ensure_windows_tool_path_prepends_existing_standard_bins(monkeypatch, tmp_path: Path):
    rancher_bin = tmp_path / "Rancher Desktop" / "resources" / "resources" / "win32" / "bin"
    choco_bin = tmp_path / "chocolatey" / "bin"
    existing_bin = tmp_path / "existing"
    rancher_bin.mkdir(parents=True)
    choco_bin.mkdir(parents=True)
    existing_bin.mkdir()

    monkeypatch.setattr(config, "is_windows", lambda: True)
    monkeypatch.setattr(config, "WINDOWS_TOOL_PATHS", (rancher_bin, choco_bin))
    monkeypatch.setenv("PATH", str(existing_bin))

    added = config.ensure_windows_tool_path()

    parts = config.os.environ["PATH"].split(config.os.pathsep)
    assert added == [rancher_bin, choco_bin]
    assert parts[:2] == [str(rancher_bin), str(choco_bin)]
    assert parts[-1] == str(existing_bin)


def test_shell_command_quotes_windows_paths(monkeypatch):
    monkeypatch.setattr(config, "is_windows", lambda: True)

    command = config.shell_command([r"C:\Program Files\Rancher Desktop\docker.exe", "build", "-t", "image:tag"])

    assert command.startswith('"C:\\Program Files\\Rancher Desktop\\docker.exe" build')


def test_default_project_cache_root_explicit_env_always_wins(monkeypatch, tmp_path: Path):
    explicit = tmp_path / "custom-cache"
    monkeypatch.setenv("MYCELIS_PROJECT_CACHE_ROOT", str(explicit))
    monkeypatch.setattr(config, "main_checkout_root", lambda checkout_root=None: tmp_path / "main")

    resolved = config.default_project_cache_root(checkout_root=tmp_path / "worktree")

    assert resolved == explicit


def test_default_project_cache_root_shares_main_checkout_for_linked_worktree(monkeypatch, tmp_path: Path):
    monkeypatch.delenv("MYCELIS_PROJECT_CACHE_ROOT", raising=False)
    main_root = tmp_path / "main-checkout"
    main_cache = main_root / "workspace" / "tool-cache"
    main_cache.mkdir(parents=True)
    worktree_root = tmp_path / "worktree"
    worktree_root.mkdir()
    monkeypatch.setattr(config, "main_checkout_root", lambda checkout_root=None: main_root)

    resolved = config.default_project_cache_root(checkout_root=worktree_root)

    assert resolved == main_cache


def test_default_project_cache_root_falls_back_when_main_cache_missing(monkeypatch, tmp_path: Path):
    monkeypatch.delenv("MYCELIS_PROJECT_CACHE_ROOT", raising=False)
    main_root = tmp_path / "main-checkout"
    main_root.mkdir()
    worktree_root = tmp_path / "worktree"
    worktree_root.mkdir()
    monkeypatch.setattr(config, "main_checkout_root", lambda checkout_root=None: main_root)

    resolved = config.default_project_cache_root(checkout_root=worktree_root)

    assert resolved == worktree_root / "workspace" / "tool-cache"


def test_default_project_cache_root_unchanged_for_main_checkout(monkeypatch, tmp_path: Path):
    monkeypatch.delenv("MYCELIS_PROJECT_CACHE_ROOT", raising=False)
    checkout_root = tmp_path / "main-checkout"
    checkout_root.mkdir()
    monkeypatch.setattr(config, "main_checkout_root", lambda checkout_root=None: None)

    resolved = config.default_project_cache_root(checkout_root=checkout_root)

    assert resolved == checkout_root / "workspace" / "tool-cache"
