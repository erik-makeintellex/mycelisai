from pathlib import Path

import pytest

from ops import compose_env


ROOT = Path(__file__).resolve().parents[1]


def _compose_text() -> str:
    return (ROOT / "docker-compose.yml").read_text(encoding="utf-8")


def test_core_mounts_a_read_only_code_context_root():
    compose_text = _compose_text()

    assert (
        "${MYCELIS_CODE_CONTEXT_HOST_ROOT:-./ops/code-context/empty-root}:/repo:ro"
        in compose_text
    )


def test_core_sets_code_context_roots_env_var():
    compose_text = _compose_text()

    assert "MYCELIS_CODE_CONTEXT_ROOTS: /repo" in compose_text


def test_default_code_context_root_is_safe_and_git_tracked():
    default_root = ROOT / "ops" / "code-context" / "empty-root"

    assert (default_root / ".gitkeep").exists()
    assert not any(default_root.iterdir()) or list(default_root.iterdir()) == [
        default_root / ".gitkeep"
    ]


def test_default_code_context_root_is_not_the_repo_root_holding_env():
    compose_text = _compose_text()

    # The default mount source must never be the main checkout ("." or "./"),
    # which is where .env lives; it must be a dedicated, empty subdirectory.
    assert ".:/repo:ro" not in compose_text
    assert "./:/repo:ro" not in compose_text
    assert "empty-root" in compose_text


def test_env_compose_example_documents_the_host_root_override():
    example_text = (ROOT / ".env.compose.example").read_text(encoding="utf-8")

    assert "MYCELIS_CODE_CONTEXT_HOST_ROOT=" in example_text
    assert "code-context-readonly" in example_text
    assert ".env" in example_text


def test_code_context_host_root_unset_passes():
    compose_env.validate_code_context_host_root({})


def test_code_context_host_root_default_empty_root_passes():
    default_root = ROOT / "ops" / "code-context" / "empty-root"

    compose_env.validate_code_context_host_root(
        {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(default_root)}
    )


def _make_worktree(tmp_path, name="code-context-readonly"):
    worktree = tmp_path / name
    worktree.mkdir()
    # A linked worktree's .git is a FILE (gitdir pointer), never a directory.
    (worktree / ".git").write_text(f"gitdir: /elsewhere/.git/worktrees/{name}\n", encoding="utf-8")
    return worktree


def test_code_context_host_root_clean_worktree_passes(tmp_path):
    worktree = _make_worktree(tmp_path)
    (worktree / "README.md").write_text("fixture\n", encoding="utf-8")

    compose_env.validate_code_context_host_root(
        {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(worktree)}
    )


def test_code_context_host_root_allows_committed_env_templates(tmp_path):
    # A real clean checkout tracks .env.example and .env.compose.example; those
    # are templates, not secrets, and must not trip the guard (regression for
    # the merged guard rejecting every clean checkout).
    worktree = _make_worktree(tmp_path)
    (worktree / ".env.example").write_text("MYCELIS_API_KEY=\n", encoding="utf-8")
    (worktree / ".env.compose.example").write_text("MYCELIS_BOOTSTRAP_TEMPLATE_ID=\n", encoding="utf-8")
    (worktree / ".env.sample").write_text("PLACEHOLDER=\n", encoding="utf-8")
    (worktree / ".env.template").write_text("PLACEHOLDER=\n", encoding="utf-8")

    compose_env.validate_code_context_host_root(
        {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(worktree)}
    )


def test_code_context_host_root_rejects_main_checkout(tmp_path):
    main_checkout = tmp_path / "mycelisai"
    main_checkout.mkdir()
    # A primary checkout's .git is a real directory.
    (main_checkout / ".git").mkdir()

    with pytest.raises(SystemExit) as excinfo:
        compose_env.validate_code_context_host_root(
            {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(main_checkout)}
        )

    message = str(excinfo.value)
    assert "primary git checkout" in message
    assert str(main_checkout) not in message


@pytest.mark.parametrize("real_env_name", [".env", ".env.compose", ".env.local"])
def test_code_context_host_root_rejects_real_env_files(tmp_path, real_env_name):
    unsafe_root = _make_worktree(tmp_path, name=f"checkout-with-{real_env_name.lstrip('.')}")
    (unsafe_root / real_env_name).write_text("MYCELIS_API_KEY=should-never-be-read\n", encoding="utf-8")

    with pytest.raises(SystemExit) as excinfo:
        compose_env.validate_code_context_host_root(
            {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(unsafe_root)}
        )

    message = str(excinfo.value)
    assert ".env" in message
    assert str(unsafe_root) not in message
    assert "should-never-be-read" not in message


def test_code_context_host_root_rejects_env_alongside_its_own_template(tmp_path):
    # A template next to a real file must not mask the real secret file.
    mixed_root = _make_worktree(tmp_path, name="checkout-with-mixed-env")
    (mixed_root / ".env.example").write_text("MYCELIS_API_KEY=\n", encoding="utf-8")
    (mixed_root / ".env").write_text("MYCELIS_API_KEY=should-never-be-read\n", encoding="utf-8")

    with pytest.raises(SystemExit) as excinfo:
        compose_env.validate_code_context_host_root(
            {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(mixed_root)}
        )

    message = str(excinfo.value)
    assert ".env" in message
    assert str(mixed_root) not in message
    assert "should-never-be-read" not in message
