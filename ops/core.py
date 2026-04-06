import shutil
import tarfile
import platform
import zipfile
from pathlib import Path

from invoke import task, Collection
from .config import CORE_DIR, ROOT_DIR, ensure_managed_cache_dirs, is_windows, managed_cache_env


def _task_env(extra=None):
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _binary_output_path() -> str:
    return "bin/server.exe" if is_windows() else "bin/server"


def _target_binary_name(target_os: str) -> str:
    return "server.exe" if target_os == "windows" else "server"


def _default_target_os() -> str:
    system_name = platform.system().lower()
    if system_name.startswith("win"):
        return "windows"
    if system_name == "darwin":
        return "darwin"
    return "linux"


def _release_staging_dir(version_tag: str, target_os: str, target_arch: str) -> Path:
    return ROOT_DIR / "dist" / f"mycelis-core-{version_tag}-{target_os}-{target_arch}"


def _release_archive_path(version_tag: str, target_os: str, target_arch: str) -> Path:
    suffix = "zip" if target_os == "windows" else "tar.gz"
    return ROOT_DIR / "dist" / f"mycelis-core-{version_tag}-{target_os}-{target_arch}.{suffix}"


def _package_release_archive(staging_dir: Path, archive_path: Path, target_os: str) -> None:
    archive_path.parent.mkdir(parents=True, exist_ok=True)
    if archive_path.exists():
        archive_path.unlink()

    if target_os == "windows":
        with zipfile.ZipFile(archive_path, "w", compression=zipfile.ZIP_DEFLATED) as bundle:
            for path in staging_dir.rglob("*"):
                if path.is_file():
                    bundle.write(path, path.relative_to(staging_dir.parent))
        return

    with tarfile.open(archive_path, "w:gz") as bundle:
        bundle.add(staging_dir, arcname=staging_dir.name)

@task
def test(c):
    """Run Go Core Unit Tests."""
    with c.cd(str(CORE_DIR)):
        c.run("go test ./...", env=_task_env())


@task
def compile(c):
    """Compile the Go Core binary without building a container image."""
    print("Compiling Core binary...")
    with c.cd(str(CORE_DIR)):
        c.run(
            f"go build -v -o {_binary_output_path()} ./cmd/server",
            env=_task_env(),
        )


@task(
    help={
        "target_os": "Target OS for the packaged binary (linux, darwin, windows). Defaults to the local OS.",
        "target_arch": "Target CPU architecture (amd64, arm64). Defaults to amd64.",
        "version_tag": "Version label to embed in the release artifact name. Defaults to the computed repo version.",
    }
)
def package(c, target_os="", target_arch="amd64", version_tag=""):
    """
    Cross-compile and package a versioned Core binary archive under dist/.
    """
    effective_target_os = target_os or _default_target_os()
    effective_version_tag = version_tag or get_version(c)
    binary_name = _target_binary_name(effective_target_os)
    staging_dir = _release_staging_dir(effective_version_tag, effective_target_os, target_arch)
    archive_path = _release_archive_path(effective_version_tag, effective_target_os, target_arch)
    binary_path = staging_dir / binary_name

    print(f"Packaging Core binary release for {effective_target_os}/{target_arch} as {effective_version_tag}...")

    if staging_dir.exists():
        shutil.rmtree(staging_dir)
    staging_dir.mkdir(parents=True, exist_ok=True)

    with c.cd(str(CORE_DIR)):
        env = _task_env(
            {
                "GOOS": effective_target_os,
                "GOARCH": target_arch,
                "CGO_ENABLED": "0",
            }
        )
        c.run(f"go build -v -o {binary_path} ./cmd/server", env=env)

    readme = staging_dir / "README.txt"
    readme.write_text(
        "\n".join(
            [
                "Mycelis Core Binary",
                f"Version: {effective_version_tag}",
                f"Target: {effective_target_os}/{target_arch}",
                "",
                "Quick start:",
                "1. Copy .env.example to .env and set required values.",
                "2. Ensure PostgreSQL, NATS, and an AI provider endpoint are reachable.",
                f"3. Run {binary_name} from this folder or configure your service manager to point at it.",
            ]
        )
        + "\n",
        encoding="utf-8",
    )

    _package_release_archive(staging_dir, archive_path, effective_target_os)
    print(f"Packaged release archive: {archive_path}")

@task
def clean(c):
    """Clean Go Build Artifacts."""
    print("Cleaning Core...")
    with c.cd(str(CORE_DIR)):
        c.run("go clean", env=_task_env())
        if (CORE_DIR / "bin").exists():
           import shutil
           shutil.rmtree(str(CORE_DIR / "bin"))

from .version import get_version

@task
def build(c):
    """
    Build Go Core Binary + Docker Image.
    Returns the calculated TAG.
    """
    tag = get_version(c)
    print(f"Building Artifact: mycelis/core:{tag}")

    # 1. Build Go Binary
    compile.body(c)

    # 2. Build Docker Image
    print(f"   Building Container...")
    # Assumes Dockerfile is in core/Dockerfile and context is root
    c.run(f"docker build -t mycelis/core:{tag} -f core/Dockerfile .")

    return tag

def _load_env():
    """Load .env into the process environment for local execution.
    Uses override=True so .env values win over system env vars
    (e.g. Windows User OLLAMA_HOST=0.0.0.0 bind address).
    """
    from dotenv import load_dotenv
    load_dotenv(str(ROOT_DIR / ".env"), override=True)

@task
def run(c):
    """
    Run the Core Service locally (Native).
    Stops any existing instance first to avoid port conflicts.
    """
    stop(c)
    _load_env()
    import os, sys
    os.environ["PYTHONIOENCODING"] = "utf-8"
    # Reconfigure stdout/stderr to UTF-8 so Go server emoji logs don't crash invoke
    if is_windows() and hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    print("Starting Mycelis Core (Native)...")
    with c.cd(str(CORE_DIR)):
        bin_name = "server.exe" if is_windows() else "server"
        env = _task_env({"PYTHONIOENCODING": "utf-8"})
        if is_windows():
            c.run(f"bin\\{bin_name}", pty=False, in_stream=False, env=env)
        else:
            c.run(f"./bin/{bin_name}", pty=True, env=env)

@task
def stop(c):
    """
    Stop the Core Service (Kill).
    """
    print("Stopping Core...")
    if is_windows():
        c.run("taskkill /F /IM server.exe", warn=True)
        c.run("taskkill /F /IM core.exe", warn=True) # Handle legacy naming
    else:
        c.run("pkill server", warn=True)
        c.run("pkill core", warn=True)

@task
def restart(c):
    """
    Restart the Core Service (Kill + Run).
    """
    print("Restarting Core...")
    stop(c)
    run(c)

@task
def smoke(c):
    """
    Run Governance Smoke Tests (Go).
    """
    print("Running Smoke Tests...")
    with c.cd(str(CORE_DIR)):
        c.run("go run ./cmd/smoke/main.go", env=_task_env())

ns = Collection("core")
ns.add_task(test)
ns.add_task(compile)
ns.add_task(package)
ns.add_task(clean)
ns.add_task(build)
ns.add_task(run)
ns.add_task(stop)
ns.add_task(restart)
ns.add_task(smoke)
