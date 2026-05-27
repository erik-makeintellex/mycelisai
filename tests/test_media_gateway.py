from __future__ import annotations

import base64
import json

from fastapi.testclient import TestClient

from cognitive.src import media_gateway


def test_media_gateway_adapts_auto1111_txt2img(monkeypatch):
    seen: dict[str, object] = {}
    image = base64.b64encode(b"png-bytes").decode("ascii")

    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "forge")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://127.0.0.1:7860")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_STEPS", "18")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_CFG_SCALE", "6.5")

    def fake_post(url, body, timeout):
        seen["url"] = url
        seen["body"] = body
        seen["timeout"] = timeout
        return {"images": [image], "parameters": body}

    monkeypatch.setattr(media_gateway, "_json_post", fake_post)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private launch concept", "n": 1, "size": "768x512", "model": "local-media"},
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["data"][0]["b64_json"] == image
    assert seen["url"] == "http://127.0.0.1:7860/sdapi/v1/txt2img"
    assert seen["body"] == {
        "prompt": "private launch concept",
        "width": 768,
        "height": 512,
        "steps": 18,
        "cfg_scale": 6.5,
        "batch_size": 1,
        "n_iter": 1,
        "send_images": True,
        "save_images": False,
    }


def test_media_gateway_adapts_comfyui_workflow_prompt_history_and_view(monkeypatch, tmp_path):
    seen: dict[str, object] = {}
    workflow = {
        "3": {"class_type": "CLIPTextEncode", "inputs": {"text": ""}},
        "5": {"class_type": "EmptyLatentImage", "inputs": {"width": 512, "height": 512, "batch_size": 1}},
        "9": {"class_type": "SaveImage", "inputs": {"images": ["8", 0]}},
    }
    workflow_file = tmp_path / "workflow.json"
    workflow_file.write_text(json.dumps(workflow), encoding="utf-8")

    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "comfyui")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://127.0.0.1:8188")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE", str(workflow_file))
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_COMFY_PROMPT_NODE_ID", "3")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_COMFY_SIZE_NODE_ID", "5")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_COMFY_BATCH_NODE_ID", "5")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_COMFY_POLL_SECONDS", "0.1")

    def fake_post(url, body, timeout):
        seen["post_url"] = url
        seen["post_body"] = body
        return {"prompt_id": "prompt-123", "number": 1, "node_errors": {}}

    def fake_json_get(url, timeout):
        seen["history_url"] = url
        return {
            "prompt-123": {
                "outputs": {
                    "9": {
                        "images": [
                            {"filename": "ComfyUI_00001_.png", "subfolder": "", "type": "output"}
                        ]
                    }
                }
            }
        }

    def fake_bytes_get(url, timeout):
        seen["view_url"] = url
        return b"png-bytes"

    monkeypatch.setattr(media_gateway, "_json_post", fake_post)
    monkeypatch.setattr(media_gateway, "_json_get", fake_json_get)
    monkeypatch.setattr(media_gateway, "_bytes_get", fake_bytes_get)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private node graph concept", "n": 1, "size": "768x512"},
    )

    assert response.status_code == 200
    assert response.json()["data"][0]["b64_json"] == base64.b64encode(b"png-bytes").decode("ascii")
    assert seen["post_url"] == "http://127.0.0.1:8188/prompt"
    assert seen["history_url"] == "http://127.0.0.1:8188/history/prompt-123"
    assert seen["view_url"] == "http://127.0.0.1:8188/view?filename=ComfyUI_00001_.png&subfolder=&type=output"
    submitted_workflow = seen["post_body"]["prompt"]
    assert submitted_workflow["3"]["inputs"]["text"] == "private node graph concept"
    assert submitted_workflow["5"]["inputs"]["width"] == 768
    assert submitted_workflow["5"]["inputs"]["height"] == 512
    assert submitted_workflow["5"]["inputs"]["batch_size"] == 1


def test_media_gateway_comfyui_requires_workflow_mapping(monkeypatch):
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "comfyui")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://127.0.0.1:8188")
    monkeypatch.delenv("MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE", raising=False)
    monkeypatch.delenv("MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_JSON", raising=False)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512"},
    )

    assert response.status_code == 500
    assert "MYCELIS_MEDIA_GATEWAY_COMFY_WORKFLOW_FILE" in response.json()["detail"]


def test_media_gateway_rejects_url_response_format(monkeypatch):
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "forge")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://127.0.0.1:7860")

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512", "response_format": "url"},
    )

    assert response.status_code == 400
    assert "response_format='url' is not supported" in response.json()["detail"]
    assert "private API response" in response.json()["detail"]


def test_media_gateway_blocks_public_upstream_by_default(monkeypatch):
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "forge")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://8.8.8.8:7860")
    monkeypatch.delenv("MYCELIS_MEDIA_GATEWAY_ALLOW_PUBLIC_UPSTREAM", raising=False)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512"},
    )

    assert response.status_code == 500
    assert "refused a public upstream host" in response.json()["detail"]


def test_media_gateway_allows_explicit_public_upstream_override(monkeypatch):
    seen: dict[str, object] = {}
    image = base64.b64encode(b"png-bytes").decode("ascii")

    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "forge")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_UPSTREAM", "http://8.8.8.8:7860")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_ALLOW_PUBLIC_UPSTREAM", "1")

    def fake_post(url, body, timeout):
        seen["url"] = url
        return {"images": [image]}

    monkeypatch.setattr(media_gateway, "_json_post", fake_post)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512"},
    )

    assert response.status_code == 200
    assert seen["url"] == "http://8.8.8.8:7860/sdapi/v1/txt2img"


def test_media_gateway_invalid_numeric_env_falls_back(monkeypatch):
    seen: dict[str, object] = {}
    image = base64.b64encode(b"png-bytes").decode("ascii")

    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "forge")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_STEPS", "not-an-int")
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_CFG_SCALE", "not-a-float")

    def fake_post(url, body, timeout):
        seen["body"] = body
        return {"images": [image]}

    monkeypatch.setattr(media_gateway, "_json_post", fake_post)

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512"},
    )

    assert response.status_code == 200
    assert seen["body"]["steps"] == 24
    assert seen["body"]["cfg_scale"] == 7.0
