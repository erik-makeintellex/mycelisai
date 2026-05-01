from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def test_state_and_architecture_entrypoints_are_not_loose_at_root():
    forbidden_root_files = [
        "V8_DEV_STATE.md",
        "V7_DEV_STATE.md",
        "v8-2.md",
        "mycelis-architecture-v7.md",
    ]

    loose_files = [name for name in forbidden_root_files if (ROOT / name).exists()]
    assert not loose_files, f"State and architecture files must not be loose at repo root: {loose_files}"

    required_locations = [
        ROOT / ".state" / "V8_DEV_STATE.md",
        ROOT / ".state" / "V7_DEV_STATE.md",
        ROOT / "architecture" / "v8-2.md",
        ROOT / "architecture" / "mycelis-architecture-v7.md",
    ]
    missing = [str(path.relative_to(ROOT)) for path in required_locations if not path.exists()]
    assert not missing, f"State and architecture files are missing required locations: {missing}"
