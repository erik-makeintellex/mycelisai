from invoke import task, Collection
from pathlib import Path
from .config import ROOT_DIR, is_windows

COGNITIVE_DIR = ROOT_DIR / "cognitive"
ENGINE_CONFIG = COGNITIVE_DIR / "config" / "engine.yaml"


def _load_engine_config():
    """Load the cognitive engine YAML configuration."""
    import yaml
    with open(str(ENGINE_CONFIG)) as f:
        return yaml.safe_load(f)


@task
def install(c):
    """Install cognitive engine dependencies via uv (vLLM + Diffusers)."""
    print("Installing Mycelis Cognitive Engine dependencies...")
    with c.cd(str(COGNITIVE_DIR)):
        c.run("uv sync", pty=not is_windows())
    print("Cognitive engine dependencies installed.")


@task
def llm(c):
    """
    Start the vLLM text inference server.

    Serves the configured model with OpenAI-compatible API on the configured port.
    Uses AWQ quantization and respects GPU memory partitioning from engine.yaml.
    """
    cfg = _load_engine_config()
    text = cfg["text"]

    model = text["model"]
    host = text["host"]
    port = text["port"]
    gpu_mem = text.get("gpu_memory_utilization", 0.60)
    max_len = text.get("max_model_len", 8192)
    quant = text.get("quantization", "awq")
    tp = text.get("tensor_parallel_size", 1)
    api_key = text.get("api_key", "mycelis-local")

    print(f"Starting vLLM text engine: {model}")
    print(f"  Host: {host}:{port}")
    print(f"  GPU Memory: {gpu_mem*100:.0f}%")
    print(f"  Context: {max_len} tokens")
    print(f"  Quantization: {quant}")

    cmd = (
        f"uv run python -m vllm.entrypoints.openai.api_server "
        f"--model {model} "
        f"--host {host} "
        f"--port {port} "
        f"--gpu-memory-utilization {gpu_mem} "
        f"--max-model-len {max_len} "
        f"--quantization {quant} "
        f"--tensor-parallel-size {tp} "
        f"--api-key {api_key} "
        f"--served-model-name qwen2.5-coder"
    )

    with c.cd(str(COGNITIVE_DIR)):
        c.run(cmd, pty=not is_windows(), in_stream=False)


@task
def media(c):
    """
    Start the Diffusers media generation server.

    Serves an OpenAI-compatible /v1/images/generations endpoint.
    """
    cfg = _load_engine_config()
    media_cfg = cfg["media"]

    host = media_cfg["host"]
    port = media_cfg["port"]
    model = media_cfg["model"]

    print(f"Starting Media Engine: {model}")
    print(f"  Host: {host}:{port}")

    cmd = (
        f"uv run uvicorn src.media_server:app "
        f"--host {host} --port {port} --log-level info"
    )

    with c.cd(str(COGNITIVE_DIR)):
        c.run(cmd, pty=not is_windows(), in_stream=False)


@task
def up(c):
    """
    Start the full cognitive stack (vLLM + Media Server).

    Launches both services. Use Ctrl+C to stop.
    For production, run `cognitive.llm` and `cognitive.media` in separate terminals.
    """
    import subprocess
    import sys
    import signal

    cfg = _load_engine_config()
    text = cfg["text"]
    media_cfg = cfg["media"]

    procs = []

    # Build vLLM command
    llm_cmd = [
        sys.executable, "-m", "vllm.entrypoints.openai.api_server",
        "--model", text["model"],
        "--host", text["host"],
        "--port", str(text["port"]),
        "--gpu-memory-utilization", str(text.get("gpu_memory_utilization", 0.60)),
        "--max-model-len", str(text.get("max_model_len", 8192)),
        "--quantization", text.get("quantization", "awq"),
        "--tensor-parallel-size", str(text.get("tensor_parallel_size", 1)),
        "--api-key", text.get("api_key", "mycelis-local"),
        "--served-model-name", "qwen2.5-coder",
    ]

    # Build media server command
    media_cmd = [
        sys.executable, "-m", "uvicorn", "src.media_server:app",
        "--host", media_cfg["host"],
        "--port", str(media_cfg["port"]),
        "--log-level", "info",
    ]

    print("Starting Mycelis Cognitive Stack...")
    print(f"  Text Engine (vLLM): port {text['port']}")
    print(f"  Media Engine (Diffusers): port {media_cfg['port']}")

    try:
        llm_proc = subprocess.Popen(llm_cmd, cwd=str(COGNITIVE_DIR))
        procs.append(llm_proc)
        print(f"  vLLM PID: {llm_proc.pid}")

        media_proc = subprocess.Popen(media_cmd, cwd=str(COGNITIVE_DIR))
        procs.append(media_proc)
        print(f"  Media PID: {media_proc.pid}")

        print("\nCognitive stack running. Press Ctrl+C to stop.")
        # Wait for either process to exit
        for p in procs:
            p.wait()

    except KeyboardInterrupt:
        print("\nShutting down cognitive stack...")
        for p in procs:
            p.terminate()
        for p in procs:
            p.wait(timeout=10)
        print("Cognitive stack stopped.")


@task
def stop(c):
    """Stop all cognitive engine processes (vLLM + Media Server)."""
    print("Stopping Cognitive Engine...")
    if is_windows():
        # vLLM runs as python process
        c.run('taskkill /F /FI "WINDOWTITLE eq vllm*"', warn=True)
        # Also kill by port
        c.run(
            'for /f "tokens=5" %a in (\'netstat -aon ^| findstr :8000 ^| findstr LISTENING\') do taskkill /F /PID %a',
            warn=True,
        )
        c.run(
            'for /f "tokens=5" %a in (\'netstat -aon ^| findstr :8001 ^| findstr LISTENING\') do taskkill /F /PID %a',
            warn=True,
        )
    else:
        c.run("pkill -f 'vllm.entrypoints'", warn=True)
        c.run("pkill -f 'media_server'", warn=True)
    print("Cognitive engine stopped.")


@task
def status(c):
    """Check the status of cognitive engine services."""
    import urllib.request
    import json

    cfg = _load_engine_config()

    # Check vLLM
    text_port = cfg["text"]["port"]
    try:
        req = urllib.request.urlopen(f"http://localhost:{text_port}/health", timeout=3)
        print(f"  Text Engine (vLLM) : UP (port {text_port})")
    except Exception:
        print(f"  Text Engine (vLLM) : DOWN (port {text_port})")

    # Check Media Server
    media_port = cfg["media"]["port"]
    try:
        req = urllib.request.urlopen(f"http://localhost:{media_port}/health", timeout=3)
        print(f"  Media Engine       : UP (port {media_port})")
    except Exception:
        print(f"  Media Engine       : DOWN (port {media_port})")


ns = Collection("cognitive")
ns.add_task(install)
ns.add_task(llm)
ns.add_task(media)
ns.add_task(up)
ns.add_task(stop)
ns.add_task(status)
