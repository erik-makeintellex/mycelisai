"""
Mycelis Media Engine — Diffusers-backed image generation server.

Provides an OpenAI-compatible /v1/images/generations endpoint so agents
can call it the same way they'd call DALL-E. Uses Stable Diffusion XL
by default with automatic GPU/CPU selection.

Run via: uvx inv cognitive.media
"""

import base64
import io
import logging
import time
import uuid
from contextlib import asynccontextmanager
from pathlib import Path

import torch
import yaml
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger("mycelis.media")

# ── Configuration ────────────────────────────────────────────

CONFIG_PATH = Path(__file__).parent.parent / "config" / "engine.yaml"


def load_config() -> dict:
    with open(CONFIG_PATH) as f:
        return yaml.safe_load(f)["media"]


# ── Models ───────────────────────────────────────────────────


class ImageGenerationRequest(BaseModel):
    """OpenAI-compatible image generation request."""
    prompt: str
    n: int = Field(default=1, ge=1, le=4)
    size: str = Field(default="1024x1024")
    response_format: str = Field(default="b64_json")  # "b64_json" or "url"
    model: str = Field(default="stable-diffusion-xl")
    quality: str = Field(default="standard")  # ignored, for compat
    style: str = Field(default="natural")  # ignored, for compat


class ImageData(BaseModel):
    b64_json: str | None = None
    url: str | None = None
    revised_prompt: str | None = None


class ImageGenerationResponse(BaseModel):
    created: int
    data: list[ImageData]


# ── Pipeline Singleton ───────────────────────────────────────

_pipeline = None
_config = None


def get_device(cfg: dict) -> str:
    device = cfg.get("device", "auto")
    if device == "auto":
        return "cuda" if torch.cuda.is_available() else "cpu"
    return device


def load_pipeline(cfg: dict):
    """Load the Diffusers pipeline on first request."""
    global _pipeline

    if _pipeline is not None:
        return _pipeline

    from diffusers import StableDiffusionXLPipeline, DPMSolverMultistepScheduler

    model_id = cfg["model"]
    device = get_device(cfg)

    logger.info(f"Loading pipeline: {model_id} on {device}")

    pipe = StableDiffusionXLPipeline.from_pretrained(
        model_id,
        torch_dtype=torch.float16 if device == "cuda" else torch.float32,
        use_safetensors=True,
        variant="fp16" if device == "cuda" else None,
    )

    # Use DPM++ 2M Karras scheduler for faster inference
    pipe.scheduler = DPMSolverMultistepScheduler.from_config(
        pipe.scheduler.config, use_karras_sigmas=True
    )

    if device == "cuda":
        pipe = pipe.to(device)
        # Enable memory-efficient attention if available
        try:
            pipe.enable_xformers_memory_efficient_attention()
        except Exception:
            logger.info("xformers not available, using default attention")
    else:
        # CPU mode — enable sequential offload equivalent
        pipe.enable_model_cpu_offload()

    if cfg.get("compile", False) and device == "cuda":
        logger.info("Compiling pipeline with torch.compile...")
        pipe.unet = torch.compile(pipe.unet, mode="reduce-overhead", fullgraph=True)

    _pipeline = pipe
    logger.info(f"Pipeline ready on {device}")
    return pipe


# ── FastAPI App ──────────────────────────────────────────────


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Pre-load config; pipeline lazy-loads on first request."""
    global _config
    _config = load_config()
    logger.info(f"Media engine configured: {_config['model']}")
    yield
    # Cleanup
    global _pipeline
    if _pipeline is not None:
        del _pipeline
        if torch.cuda.is_available():
            torch.cuda.empty_cache()


app = FastAPI(
    title="Mycelis Media Engine",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    return {
        "status": "ok",
        "model": _config["model"] if _config else "not loaded",
        "pipeline_loaded": _pipeline is not None,
        "device": get_device(_config) if _config else "unknown",
        "cuda_available": torch.cuda.is_available(),
    }


@app.get("/v1/models")
async def list_models():
    """OpenAI-compatible model listing."""
    return {
        "object": "list",
        "data": [
            {
                "id": "stable-diffusion-xl",
                "object": "model",
                "created": int(time.time()),
                "owned_by": "mycelis",
            }
        ],
    }


@app.post("/v1/images/generations", response_model=ImageGenerationResponse)
async def generate_images(req: ImageGenerationRequest):
    """OpenAI-compatible image generation endpoint."""
    if _config is None:
        raise HTTPException(status_code=503, detail="Media engine not configured")

    pipe = load_pipeline(_config)

    # Parse size
    try:
        width, height = map(int, req.size.split("x"))
    except ValueError:
        width = _config.get("default_width", 1024)
        height = _config.get("default_height", 1024)

    steps = _config.get("default_steps", 30)
    guidance = _config.get("default_guidance_scale", 7.5)

    logger.info(f"Generating {req.n} image(s): '{req.prompt[:80]}...' [{width}x{height}]")
    start = time.time()

    images = []
    for _ in range(req.n):
        with torch.inference_mode():
            result = pipe(
                prompt=req.prompt,
                width=width,
                height=height,
                num_inference_steps=steps,
                guidance_scale=guidance,
            )
            img = result.images[0]

        # Encode to base64
        buf = io.BytesIO()
        img.save(buf, format="PNG")
        b64 = base64.b64encode(buf.getvalue()).decode("utf-8")

        images.append(ImageData(b64_json=b64, revised_prompt=req.prompt))

    elapsed = time.time() - start
    logger.info(f"Generated {req.n} image(s) in {elapsed:.1f}s")

    return ImageGenerationResponse(
        created=int(time.time()),
        data=images,
    )
