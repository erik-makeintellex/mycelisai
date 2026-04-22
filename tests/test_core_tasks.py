from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
import json
import tarfile
import zipfile

from invoke import Context

from ops import core


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""


class FakeContext(Context):
    def __init__(self):
        super().__init__()
        self.commands: list[str] = []
        self.cd_paths: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return FakeResult()

    @contextmanager
    def cd(self, path: str):
        self.cd_paths.append(path)
        yield


def test_compile_builds_repo_local_binary(monkeypatch):
    ctx = FakeContext()
    monkeypatch.setattr(core, "is_windows", lambda: False)

    core.compile.body(ctx)

    assert ctx.cd_paths == [str(core.CORE_DIR)]
    assert ctx.commands == ["go build -v -o bin/server ./cmd/server"]


def test_build_uses_compile_and_never_tags_latest(monkeypatch):
    ctx = FakeContext()
    compile_calls: list[str] = []

    monkeypatch.setattr(core.compile, "body", lambda _c: compile_calls.append("compile"))
    monkeypatch.setattr(core, "get_version", lambda _c: "v0.1.0-deadbee")

    core.build.body(ctx)

    assert compile_calls == ["compile"]
    assert ctx.commands == ["docker build -t mycelis/core:v0.1.0-deadbee -f core/Dockerfile ."]
    assert not any("latest" in command for command in ctx.commands)


def test_package_builds_versioned_linux_archive(monkeypatch, tmp_path):
    ctx = FakeContext()

    monkeypatch.setattr(core, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(core, "CORE_DIR", tmp_path / "core")
    monkeypatch.setattr(core, "get_version", lambda _c: "v0.1.0-deadbee")

    core.package.body(ctx, target_os="linux", target_arch="amd64")

    expected_staging = tmp_path / "dist" / "mycelis-core-v0.1.0-deadbee-linux-amd64"
    expected_archive = tmp_path / "dist" / "mycelis-core-v0.1.0-deadbee-linux-amd64.tar.gz"
    expected_manifest = tmp_path / "dist" / "mycelis-core-v0.1.0-deadbee-linux-amd64.manifest.json"
    expected_checksum = tmp_path / "dist" / "mycelis-core-v0.1.0-deadbee-linux-amd64.tar.gz.sha256"

    assert ctx.cd_paths == [str(tmp_path / "core")]
    assert ctx.commands == [f"go build -v -o {expected_staging / 'server'} ./cmd/server"]
    assert expected_archive.exists()
    assert expected_manifest.exists()
    assert expected_checksum.exists()
    with tarfile.open(expected_archive, "r:gz") as bundle:
        names = bundle.getnames()
    assert f"{expected_staging.name}/README.txt" in names
    assert f"{expected_staging.name}/release-manifest.json" in names
    manifest = json.loads(expected_manifest.read_text(encoding="utf-8"))
    assert manifest["artifact_kind"] == "self_hosted_core_binary"
    assert manifest["status"] == "scaffold"
    assert manifest["archive_path"] == "dist/mycelis-core-v0.1.0-deadbee-linux-amd64.tar.gz"


def test_package_builds_versioned_windows_zip(monkeypatch, tmp_path):
    ctx = FakeContext()

    monkeypatch.setattr(core, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(core, "CORE_DIR", tmp_path / "core")

    core.package.body(ctx, target_os="windows", target_arch="amd64", version_tag="v9.9.9-test")

    expected_staging = tmp_path / "dist" / "mycelis-core-v9.9.9-test-windows-amd64"
    expected_archive = tmp_path / "dist" / "mycelis-core-v9.9.9-test-windows-amd64.zip"
    expected_manifest = tmp_path / "dist" / "mycelis-core-v9.9.9-test-windows-amd64.manifest.json"

    assert ctx.commands == [f"go build -v -o {expected_staging / 'server.exe'} ./cmd/server"]
    assert expected_archive.exists()
    assert expected_manifest.exists()
    with zipfile.ZipFile(expected_archive) as bundle:
        names = bundle.namelist()
    assert f"{expected_staging.name}/README.txt" in names
    assert f"{expected_staging.name}/release-manifest.json" in names


def test_default_target_os_maps_platform_names(monkeypatch):
    monkeypatch.setattr(core.platform, "system", lambda: "Darwin")
    assert core._default_target_os() == "darwin"
