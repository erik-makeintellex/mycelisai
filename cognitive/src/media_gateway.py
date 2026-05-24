"""
Local media gateway for Pinokio-hosted image engines.

Mycelis Core speaks an OpenAI-compatible image-generation contract. Pinokio
apps such as Forge and AUTOMATIC1111 expose their own Stable Diffusion APIs.
This gateway keeps Core stable while adapting local/private generators behind
the same /v1/images/generations endpoint.
"""

from __future__ import annotations

import base64
import json
import os
import time
import urllib.error
import urllib.request
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field


DEFAULT_BACKEND = "auto1111"
DEFAULT_UPSTREAM = "http://127.0.0.1:7860"
DEFAULT_TIMEOUT_SECONDS = 300


class ImageGenerationRequest(BaseModel):
    prompt: str
    n: int = Field(default=1, ge=1, le=4)
    size: str = Field(default="1024x1024")
    response_format: str = Field(default="b64_json")
    model: str = Field(default="local-media")
    quality: str = Field(default="standard")
    style: str = Field(default="natural")


class ImageData(BaseModel):
    b64_json: str | None = None
    url: str | None = None
    revised_prompt: str | None = None


class ImageGenerationResponse(BaseModel):
    created: int
    data: list[ImageData]


def _env(name: str, default: str) -> str:
    return os.getenv(name, default).strip() or default


def gateway_backend() -> str:
    return _env("MYCELIS_MEDIA_GATEWAY_BACKEND", DEFAULT_BACKEND).lower()


def gateway_upstream() -> str:
    return _env("MYCELIS_MEDIA_GATEWAY_UPSTREAM", DEFAULT_UPSTREAM).rstrip("/")


def gateway_timeout() -> int:
    raw = _env("MYCELIS_MEDIA_GATEWAY_TIMEOUT_SECONDS", str(DEFAULT_TIMEOUT_SECONDS))
    try:
        return max(1, int(raw))
    except ValueError:
        return DEFAULT_TIMEOUT_SECONDS


def _json_get(url: str, timeout: int) -> dict[str, Any]:
    req = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = resp.read().decode("utf-8")
    return json.loads(payload) if payload else {}


def _json_post(url: str, body: dict[str, Any], timeout: int) -> dict[str, Any]:
    data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = resp.read().decode("utf-8")
    return json.loads(payload) if payload else {}


def _parse_size(size: str) -> tuple[int, int]:
    try:
        width_raw, height_raw = size.lower().split("x", 1)
        width = int(width_raw)
        height = int(height_raw)
    except (AttributeError, ValueError):
        return 1024, 1024
    if width <= 0 or height <= 0:
        return 1024, 1024
    return width, height


def _strip_data_uri(value: str) -> str:
    if "," in value and value.lower().startswith("data:image/"):
        return value.split(",", 1)[1]
    return value


def _ensure_base64_image(value: str) -> str:
    image = _strip_data_uri(value.strip())
    if not image:
        raise HTTPException(status_code=502, detail="local media engine returned an empty image")
    try:
        base64.b64decode(image, validate=True)
    except Exception as exc:
        raise HTTPException(status_code=502, detail="local media engine returned invalid base64") from exc
    return image


def _auto1111_generate(req: ImageGenerationRequest) -> ImageGenerationResponse:
    width, height = _parse_size(req.size)
    timeout = gateway_timeout()
    steps = int(_env("MYCELIS_MEDIA_GATEWAY_STEPS", "24"))
    cfg_scale = float(_env("MYCELIS_MEDIA_GATEWAY_CFG_SCALE", "7"))
    sampler = os.getenv("MYCELIS_MEDIA_GATEWAY_SAMPLER", "").strip()

    body: dict[str, Any] = {
        "prompt": req.prompt,
        "width": width,
        "height": height,
        "steps": steps,
        "cfg_scale": cfg_scale,
        "batch_size": req.n,
        "n_iter": 1,
        "send_images": True,
        "save_images": False,
    }
    if sampler:
        body["sampler_name"] = sampler

    try:
        upstream = _json_post(f"{gateway_upstream()}/sdapi/v1/txt2img", body, timeout)
    except urllib.error.URLError as exc:
        raise HTTPException(status_code=503, detail=f"local media engine unreachable: {exc}") from exc

    images = upstream.get("images")
    if not isinstance(images, list) or not images:
        raise HTTPException(status_code=502, detail="local media engine returned no images")

    data = [
        ImageData(b64_json=_ensure_base64_image(str(image)), revised_prompt=req.prompt)
        for image in images[: req.n]
    ]
    return ImageGenerationResponse(created=int(time.time()), data=data)


app = FastAPI(title="Mycelis Local Media Gateway", version="0.1.0")


@app.get("/health")
async def health() -> dict[str, Any]:
    backend = gateway_backend()
    upstream = gateway_upstream()
    timeout = min(gateway_timeout(), 5)

    if backend not in {"auto1111", "automatic1111", "forge"}:
        return {
            "status": "unsupported",
            "backend": backend,
            "upstream": upstream,
            "detail": "Only Forge/AUTOMATIC1111 txt2img is implemented in this gateway slice.",
        }

    try:
        _json_get(f"{upstream}/sdapi/v1/options", timeout)
    except Exception as exc:
        return {"status": "offline", "backend": backend, "upstream": upstream, "detail": str(exc)}

    return {"status": "ok", "backend": backend, "upstream": upstream}


@app.get("/v1/models")
async def list_models() -> dict[str, Any]:
    return {
        "object": "list",
        "data": [
            {
                "id": "local-media",
                "object": "model",
                "created": int(time.time()),
                "owned_by": "mycelis-local",
            }
        ],
    }


@app.post("/v1/images/generations", response_model=ImageGenerationResponse)
async def generate_images(req: ImageGenerationRequest) -> ImageGenerationResponse:
    backend = gateway_backend()
    if backend in {"auto1111", "automatic1111", "forge"}:
        return _auto1111_generate(req)
    raise HTTPException(
        status_code=501,
        detail="Only Forge/AUTOMATIC1111 txt2img is implemented in this gateway slice.",
    )
