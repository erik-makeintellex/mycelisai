from __future__ import annotations

import base64

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


def test_media_gateway_reports_unsupported_backend(monkeypatch):
    monkeypatch.setenv("MYCELIS_MEDIA_GATEWAY_BACKEND", "comfyui")

    client = TestClient(media_gateway.app)
    response = client.post(
        "/v1/images/generations",
        json={"prompt": "private concept", "n": 1, "size": "512x512"},
    )

    assert response.status_code == 501
    assert "Forge/AUTOMATIC1111" in response.json()["detail"]
