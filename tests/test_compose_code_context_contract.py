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


def test_code_context_host_root_clean_worktree_passes(tmp_path):
    worktree = tmp_path / "code-context-readonly"
    worktree.mkdir()
    # A linked worktree's .git is a FILE (gitdir pointer), never a directory.
    (worktree / ".git").write_text("gitdir: /elsewhere/.git/worktrees/code-context-readonly\n", encoding="utf-8")
    (worktree / "README.md").write_text("fixture\n", encoding="utf-8")

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


def test_code_context_host_root_rejects_directory_with_dotenv(tmp_path):
    unsafe_root = tmp_path / "checkout-with-env"
    unsafe_root.mkdir()
    (unsafe_root / ".env").write_text("MYCELIS_API_KEY=should-never-be-read\n", encoding="utf-8")

    with pytest.raises(SystemExit) as excinfo:
        compose_env.validate_code_context_host_root(
            {"MYCELIS_CODE_CONTEXT_HOST_ROOT": str(unsafe_root)}
        )

    message = str(excinfo.value)
    assert ".env" in message
    assert str(unsafe_root) not in message
    assert "should-never-be-read" not in message
