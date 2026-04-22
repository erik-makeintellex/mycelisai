from __future__ import annotations

import hashlib
import json
import re
from pathlib import Path
from typing import Any


def resolve_repo_path(raw_path: str, *, root_dir: Path) -> Path:
    candidate = Path(raw_path).expanduser()
    if candidate.is_absolute():
        return candidate.resolve()
    return (root_dir / candidate).resolve()


def relative_to_root(path: Path, *, root_dir: Path) -> str:
    resolved = path.resolve()
    try:
        return resolved.relative_to(root_dir).as_posix()
    except ValueError:
        return resolved.as_posix()


def slugify_label(value: str) -> str:
    slug = re.sub(r"[^A-Za-z0-9._-]+", "-", value.strip())
    return slug.strip("-") or "manual"


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def write_checksum_file(path: Path) -> Path:
    checksum_path = path.parent / f"{path.name}.sha256"
    checksum = sha256_file(path)
    checksum_path.write_text(f"{checksum}  {path.name}\n", encoding="utf-8")
    return checksum_path


def write_json(path: Path, payload: dict[str, Any]) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return path
