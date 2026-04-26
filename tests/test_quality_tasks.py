from __future__ import annotations

from pathlib import Path

import pytest

from ops import quality


def _write_lines(path: Path, count: int):
    path.write_text("\n".join(f"line {i}" for i in range(count)) + "\n", encoding="utf-8")


def test_load_legacy_caps_parses_valid_entries(tmp_path: Path):
    caps_file = tmp_path / "caps.txt"
    caps_file.write_text(
        """
# comment
core/internal/swarm/agent.go=1025
bad-entry
interface/store/useCortexStore.ts=2396
""".strip()
        + "\n",
        encoding="utf-8",
    )

    caps = quality._load_legacy_caps(caps_file)

    assert caps["core/internal/swarm/agent.go"] == 1025
    assert caps["interface/store/useCortexStore.ts"] == 2396


def test_default_quality_paths_cover_main_source_tree():
    paths = set(quality.DEFAULT_SOURCE_PATHS.split(","))

    assert {"core", "interface", "ops", "tests"}.issubset(paths)


def test_generated_sources_are_skipped():
    assert quality._should_skip(Path("core/pkg/pb/swarm/swarm.pb.go"))
    assert quality._should_skip(Path("sdk/python/src/relay/proto/swarm_pb2.py"))
    assert quality._should_skip(Path("sdk/python/src/relay/proto/swarm_pb2_grpc.py"))


def test_max_lines_fails_without_cap(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 12)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {})

    with pytest.raises(SystemExit):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_passes_with_legacy_cap(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 12)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_fails_when_exceeding_legacy_cap(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 13)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    with pytest.raises(SystemExit):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)
