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
