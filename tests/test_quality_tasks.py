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
interface/store/useCortexStore.ts=2396
""".strip()
        + "\n",
        encoding="utf-8",
    )

    caps = quality._load_legacy_caps(caps_file)

    assert caps["core/internal/swarm/agent.go"] == 1025
    assert caps["interface/store/useCortexStore.ts"] == 2396


def test_load_legacy_caps_rejects_malformed_entries(tmp_path: Path):
    caps_file = tmp_path / "caps.txt"
    caps_file.write_text("bad-entry\n", encoding="utf-8")

    with pytest.raises(SystemExit, match="malformed legacy cap entry"):
        quality._load_legacy_caps(caps_file)


def test_load_legacy_caps_rejects_duplicate_entries(tmp_path: Path):
    caps_file = tmp_path / "caps.txt"
    caps_file.write_text("a.py=301\na.py=302\n", encoding="utf-8")

    with pytest.raises(SystemExit, match="duplicate legacy cap entry"):
        quality._load_legacy_caps(caps_file)


def test_default_quality_paths_cover_main_source_tree():
    paths = set(quality.DEFAULT_SOURCE_PATHS.split(","))

    assert {"core", "interface", "ops", "tests"}.issubset(paths)
    assert {"docs", "charts", "k8s", "architecture", "README.md"}.issubset(paths)


def test_generated_sources_are_skipped():
    assert quality._should_skip(Path("core/pkg/pb/swarm/swarm.pb.go"))
    assert quality._should_skip(Path("sdk/python/src/relay/proto/swarm_pb2.py"))
    assert quality._should_skip(Path("sdk/python/src/relay/proto/swarm_pb2_grpc.py"))
    assert quality._should_skip(Path("interface/package-lock.json"))
    assert quality._should_skip(Path("uv.lock"))


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


def test_max_lines_fails_when_legacy_cap_is_loose(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 11)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    with pytest.raises(SystemExit, match="caps must match current line counts"):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_fails_when_capped_file_is_back_under_limit(monkeypatch, tmp_path: Path):
    src = tmp_path / "small.py"
    _write_lines(src, 9)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    with pytest.raises(SystemExit, match="caps for files under the limit must be removed"):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_fails_when_exceeding_legacy_cap(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 13)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    with pytest.raises(SystemExit):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_fails_when_legacy_cap_points_to_missing_scanned_file(monkeypatch, tmp_path: Path):
    src = tmp_path / "small.py"
    _write_lines(src, 4)
    missing = tmp_path / "deleted.py"

    monkeypatch.setattr(quality, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {missing.relative_to(tmp_path).as_posix(): 12})

    with pytest.raises(SystemExit, match="stale legacy max-line caps"):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=False)


def test_max_lines_strict_ignores_legacy_caps(monkeypatch, tmp_path: Path):
    src = tmp_path / "big.py"
    _write_lines(src, 12)

    monkeypatch.setattr(quality, "_parse_paths", lambda _paths: [tmp_path])
    monkeypatch.setattr(quality, "_iter_source_files", lambda _roots: [src])
    monkeypatch.setattr(quality, "_load_legacy_caps", lambda: {src.as_posix(): 12})

    with pytest.raises(SystemExit, match="file length violations"):
        quality.max_lines.body(None, limit=10, paths="ignored", strict=True)
