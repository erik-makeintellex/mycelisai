"""
Process-name and command-line hints for lifecycle cleanup.
"""

from .config import CORE_DIR

WINDOWS_COMPILED_GO_PROCESS_NAMES = (
    "go.exe",
    "server.exe",
    "probe.exe",
    "signal_gen.exe",
    "smoke.exe",
)

COMPILED_GO_PROCESS_HINTS = tuple(
    hint.lower()
    for hint in {
        "go run ./cmd/server",
        "go run .\\cmd\\server",
        "go run ./cmd/probe",
        "go run .\\cmd\\probe",
        "go run ./cmd/signal_gen",
        "go run .\\cmd\\signal_gen",
        "go run ./cmd/smoke/main.go",
        "go run .\\cmd\\smoke\\main.go",
        "cmd/server",
        "cmd\\server",
        "cmd/probe",
        "cmd\\probe",
        "cmd/signal_gen",
        "cmd\\signal_gen",
        "cmd/smoke",
        "cmd\\smoke",
        "bin/server",
        "bin\\server",
        "bin/probe",
        "bin\\probe",
        "bin/signal_gen",
        "bin\\signal_gen",
        "bin/smoke",
        "bin\\smoke",
        str((CORE_DIR / "bin" / "server").resolve()),
        str((CORE_DIR / "bin" / "server.exe").resolve()),
        str((CORE_DIR / "bin" / "probe").resolve()),
        str((CORE_DIR / "bin" / "probe.exe").resolve()),
        str((CORE_DIR / "bin" / "signal_gen").resolve()),
        str((CORE_DIR / "bin" / "signal_gen.exe").resolve()),
        str((CORE_DIR / "cmd" / "server").resolve()),
        str((CORE_DIR / "cmd" / "probe").resolve()),
        str((CORE_DIR / "cmd" / "signal_gen").resolve()),
        str((CORE_DIR / "cmd" / "smoke").resolve()),
    }
)
