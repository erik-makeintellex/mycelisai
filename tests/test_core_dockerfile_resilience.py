from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CORE_DOCKERFILE = ROOT / "core" / "Dockerfile"
TESTING = ROOT / "docs" / "TESTING.md"
OPS_README = ROOT / "ops" / "README.md"


def test_core_dockerfile_module_download_is_bounded_resumable_and_cached():
    """P4: proxy.golang.org intermittently resets parallel `go mod download`
    fetches inside the Core image build. The download step must persist
    progress across attempts (BuildKit cache mount), retry a bounded number
    of times with backoff instead of failing on the first reset, and still
    fail loudly on a real, non-transient error.
    """
    text = CORE_DOCKERFILE.read_text(encoding="utf-8")

    # Cache mounts are native to the built-in BuildKit frontend (Docker 23+).
    # A `# syntax=` directive would add a registry pull of the frontend image
    # to every build: a new network failure point in a step meant to remove one.
    assert "# syntax=" not in text, (
        "core/Dockerfile must not pull a Dockerfile frontend image; the built-in BuildKit frontend supports cache mounts"
    )

    required_snippets = [
        "--mount=type=cache,target=/go/pkg/mod",
        "max_attempts=5",
        "until go mod download; do",
        "GOMAXPROCS=2",
    ]
    # `go mod download` sizes its fetch pool from GOMAXPROCS; GOFLAGS=-p is ignored by it.
    assert "-p=2" not in text
    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "core/Dockerfile is missing bounded/resumable module download support:\n" + "\n".join(missing)

    # A real failure must still exit non-zero: no silent success path.
    assert "|| true" not in text, "core/Dockerfile must not mask a real go mod download failure with `|| true`"
    assert "exit 1" in text, "core/Dockerfile retry loop must fail loudly after exhausting attempts"

    # The build stage that compiles the binary must reuse the same module
    # cache mount so it does not re-download everything from a cold cache.
    build_line_index = text.index("go build -o /server")
    preceding = text[:build_line_index]
    assert "--mount=type=cache,target=/go/pkg/mod" in preceding.rsplit("RUN", 1)[-1] or (
        preceding.count("--mount=type=cache,target=/go/pkg/mod") >= 2
    ), "core/Dockerfile's build RUN step must also mount the /go/pkg/mod cache"


def test_docs_mention_resilient_core_module_download():
    testing_text = TESTING.read_text(encoding="utf-8")
    ops_text = OPS_README.read_text(encoding="utf-8")

    assert "go mod download" in testing_text and "retry" in testing_text, (
        "docs/TESTING.md must document the retrying Core module download build step"
    )
    assert "go mod download" in ops_text and "retry" in ops_text, (
        "ops/README.md must document the retrying Core module download build step"
    )
