"""ComfyUI adapter for the Mycelis local media gateway."""

from __future__ import annotations

import base64
import json
import os
import time
import urllib.error
import urllib.parse
import uuid
from typing import Any

from fastapi import HTTPException

from . import media_gateway


def _comfy_prefix() -> str:
    raw = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_API_PREFIX", "").strip().strip("/")
    return f"/{raw}" if raw else ""


def _comfy_url(upstream: str, path: str) -> str:
    return f"{upstream}{_comfy_prefix()}{path}"


def comfy_health_path() -> str:
    return f"{_comfy_prefix()}/system_stats"


def _load_comfy_workflow() -> dict[str, Any]:
    raw_workflow = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_JSON", "").strip()
    workflow_file = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE", "").strip()
    if raw_workflow:
        try:
            workflow = json.loads(raw_workflow)
        except json.JSONDecodeError as exc:
            raise HTTPException(status_code=500, detail="ComfyUI workflow JSON is invalid") from exc
    elif workflow_file:
        try:
            with open(workflow_file, "r", encoding="utf-8") as f:
                workflow = json.load(f)
        except OSError as exc:
            raise HTTPException(status_code=500, detail="ComfyUI workflow file could not be read") from exc
        except json.JSONDecodeError as exc:
            raise HTTPException(status_code=500, detail="ComfyUI workflow file is invalid JSON") from exc
    else:
        raise HTTPException(
            status_code=500,
            detail="ComfyUI backend requires MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE or MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_JSON",
        )
    if not isinstance(workflow, dict) or not workflow:
        raise HTTPException(status_code=500, detail="ComfyUI workflow must be a non-empty API-format object")
    return workflow


def _set_comfy_input(workflow: dict[str, Any], node_id: str, input_name: str, value: Any) -> None:
    node = workflow.get(node_id)
    if not isinstance(node, dict):
        raise HTTPException(status_code=500, detail=f"ComfyUI workflow node {node_id!r} was not found")
    inputs = node.setdefault("inputs", {})
    if not isinstance(inputs, dict):
        raise HTTPException(status_code=500, detail=f"ComfyUI workflow node {node_id!r} has invalid inputs")
    inputs[input_name] = value


def _prepare_comfy_workflow(req: media_gateway.ImageGenerationRequest) -> dict[str, Any]:
    workflow = _load_comfy_workflow()
    prompt_node = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_NODE_ID", "").strip()
    prompt_input = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_INPUT", "text").strip() or "text"
    if not prompt_node:
        raise HTTPException(status_code=500, detail="ComfyUI backend requires MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_NODE_ID")
    _set_comfy_input(workflow, prompt_node, prompt_input, req.prompt)

    width, height = media_gateway._parse_size(req.size)
    size_node = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_SIZE_NODE_ID", "").strip()
    if size_node:
        width_input = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_WIDTH_INPUT", "width").strip() or "width"
        height_input = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_HEIGHT_INPUT", "height").strip() or "height"
        _set_comfy_input(workflow, size_node, width_input, width)
        _set_comfy_input(workflow, size_node, height_input, height)

    batch_node = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_BATCH_NODE_ID", "").strip()
    if batch_node:
        batch_input = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_BATCH_INPUT", "batch_size").strip() or "batch_size"
        _set_comfy_input(workflow, batch_node, batch_input, req.n)

    return workflow


def _extract_comfy_images(history: dict[str, Any], prompt_id: str) -> list[dict[str, str]]:
    entry = history.get(prompt_id)
    if not isinstance(entry, dict):
        return []
    outputs = entry.get("outputs")
    if not isinstance(outputs, dict):
        return []
    images: list[dict[str, str]] = []
    for output in outputs.values():
        if not isinstance(output, dict):
            continue
        for image in output.get("images", []):
            if not isinstance(image, dict):
                continue
            filename = str(image.get("filename", "")).strip()
            if not filename:
                continue
            images.append(
                {
                    "filename": filename,
                    "subfolder": str(image.get("subfolder", "")).strip(),
                    "type": str(image.get("type", "output")).strip() or "output",
                }
            )
    return images


def comfy_generate(req: media_gateway.ImageGenerationRequest) -> media_gateway.ImageGenerationResponse:
    media_gateway._validate_response_format(req)
    upstream_url = media_gateway.gateway_upstream()
    media_gateway._validate_private_upstream(upstream_url)
    timeout = media_gateway.gateway_timeout()
    workflow = _prepare_comfy_workflow(req)
    client_id = os.getenv("MYCELIS_MEDIA_GATEWAY_COMFY_CLIENT_ID", "").strip() or f"mycelis-{uuid.uuid4()}"

    try:
        submitted = media_gateway._json_post(
            _comfy_url(upstream_url, "/prompt"),
            {"prompt": workflow, "client_id": client_id},
            timeout,
        )
    except urllib.error.URLError as exc:
        raise HTTPException(status_code=503, detail="local/private ComfyUI engine unreachable at configured upstream") from exc

    prompt_id = str(submitted.get("prompt_id", "")).strip()
    if not prompt_id:
        raise HTTPException(status_code=502, detail="ComfyUI did not return a prompt_id")

    deadline = time.time() + timeout
    poll_seconds = media_gateway._env_float("MYCELIS_MEDIA_GATEWAY_COMFY_POLL_SECONDS", 1.0, 0.1)
    images: list[dict[str, str]] = []
    while time.time() < deadline:
        try:
            history = media_gateway._json_get(
                _comfy_url(upstream_url, f"/history/{urllib.parse.quote(prompt_id)}"),
                min(timeout, 30),
            )
        except urllib.error.URLError as exc:
            raise HTTPException(status_code=503, detail="ComfyUI history endpoint became unreachable") from exc
        images = _extract_comfy_images(history, prompt_id)
        if images:
            break
        time.sleep(poll_seconds)
    if not images:
        raise HTTPException(status_code=504, detail="ComfyUI generation timed out before output images were available")

    data: list[media_gateway.ImageData] = []
    for image in images[: req.n]:
        query = urllib.parse.urlencode(image)
        try:
            image_bytes = media_gateway._bytes_get(_comfy_url(upstream_url, f"/view?{query}"), timeout)
        except urllib.error.URLError as exc:
            raise HTTPException(status_code=503, detail="ComfyUI output file could not be retrieved") from exc
        if not image_bytes:
            raise HTTPException(status_code=502, detail="ComfyUI output file was empty")
        data.append(
            media_gateway.ImageData(
                b64_json=base64.b64encode(image_bytes).decode("ascii"),
                revised_prompt=req.prompt,
            )
        )
    return media_gateway.ImageGenerationResponse(created=int(time.time()), data=data)
