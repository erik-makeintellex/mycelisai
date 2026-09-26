from __future__ import annotations

import ast
import re
import sys
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

# First-party packages that live in this repo rather than on PyPI.
LOCAL_MODULES = {"ops", "tests", "tasks", "cognitive", "framework_runs"}

# Imported but genuinely optional: gated behind a runtime dependency check and
# faked via sys.modules in tests (see tests/test_framework_runs_langgraph.py),
# so it is not a hard project dependency.
OPTIONAL_MODULES = {"langgraph"}

# Import name -> PyPI distribution name, where the two differ.
IMPORT_TO_DISTRIBUTION = {
    "yaml": "pyyaml",
    "dotenv": "python-dotenv",
    "nats": "nats-py",
}

_DEP_NAME_RE = re.compile(r"^[A-Za-z0-9_.-]+")


def _distribution_name(module: str) -> str:
    return IMPORT_TO_DISTRIBUTION.get(module, module).lower().replace("_", "-")


def _declared_distributions() -> set[str]:
    project = tomllib.loads((ROOT / "pyproject.toml").read_text(encoding="utf-8"))
    entries = list(project["project"].get("dependencies", []))
    for group_deps in project.get("dependency-groups", {}).values():
        entries.extend(group_deps)
    declared = set()
    for entry in entries:
        match = _DEP_NAME_RE.match(entry)
        if match:
            declared.add(match.group(0).lower().replace("_", "-"))
    return declared


def _module_imports(*root_names: str) -> set[str]:
    modules: set[str] = set()
    for root_name in root_names:
        for path in (ROOT / root_name).rglob("*.py"):
            tree = ast.parse(path.read_text(encoding="utf-8"), filename=str(path))
            for node in ast.walk(tree):
                if isinstance(node, ast.Import):
                    for alias in node.names:
                        modules.add(alias.name.split(".")[0])
                elif isinstance(node, ast.ImportFrom):
                    if node.level:  # relative import (`from . import x`)
                        continue
                    if node.module:
                        modules.add(node.module.split(".")[0])
    return modules


def _missing_declarations(imports: set[str], declared: set[str]) -> list[str]:
    missing = []
    for module in sorted(imports):
        if module in LOCAL_MODULES or module in OPTIONAL_MODULES:
            continue
        distribution = _distribution_name(module)
        if distribution not in declared:
            missing.append(f"{module} (expected distribution '{distribution}')")
    return missing


def test_every_third_party_import_under_tests_and_ops_is_declared():
    stdlib = set(sys.stdlib_module_names)
    imports = _module_imports("tests", "ops") - stdlib
    declared = _declared_distributions()
    missing = _missing_declarations(imports, declared)
    assert not missing, (
        "Undeclared third-party imports under tests/ or ops/: "
        + ", ".join(missing)
        + ". Add the dependency to pyproject.toml (dependencies or dependency-groups.dev)."
    )


def test_missing_declarations_flags_an_undeclared_import():
    missing = _missing_declarations({"pytest", "totally_fake_pkg"}, declared={"pytest"})
    assert missing == ["totally_fake_pkg (expected distribution 'totally-fake-pkg')"]


def test_missing_declarations_resolves_known_import_aliases():
    missing = _missing_declarations({"yaml", "dotenv"}, declared={"pyyaml", "python-dotenv"})
    assert missing == []
