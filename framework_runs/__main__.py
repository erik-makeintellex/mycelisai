from __future__ import annotations

import os

import uvicorn


def main() -> None:
    if os.environ.get("MYCELIS_FRAMEWORK_RUNS_ALLOW_CONFORMANCE") != "1":
        raise SystemExit(
            "Refusing to launch the non-production conformance driver. Set "
            "MYCELIS_FRAMEWORK_RUNS_ALLOW_CONFORMANCE=1 for local protocol proof."
        )
    uvicorn.run(
        "framework_runs.api:app",
        host=os.environ.get("MYCELIS_FRAMEWORK_RUNS_HOST", "127.0.0.1"),
        port=int(os.environ.get("MYCELIS_FRAMEWORK_RUNS_PORT", "8091")),
    )


if __name__ == "__main__":
    main()
