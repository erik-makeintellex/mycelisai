from __future__ import annotations

import ipaddress
import json
import os
import urllib.error
import urllib.parse
import urllib.request

from invoke.exceptions import Exit

from .config import ROOT_DIR


CLIENT_KEY_ENV = "LITELLM_PROXY_API_KEY"
MAX_RESPONSE_BYTES = 1024 * 1024


class _RejectRedirects(urllib.request.HTTPRedirectHandler):
    """Keep the scoped proxy key on the exact operator-approved origin."""

    def redirect_request(self, request, fp, code, msg, headers, new_url):
        return None


_NO_REDIRECT_OPENER = urllib.request.build_opener(_RejectRedirects())


def _open_without_redirects(request, timeout):
    return _NO_REDIRECT_OPENER.open(request, timeout=timeout)


def _resolve_client_key():
    value = os.environ.get(CLIENT_KEY_ENV, "").strip()
    if not value:
        try:
            from dotenv import dotenv_values
        except ImportError as exc:
            raise Exit(
                "python-dotenv is required to resolve the LiteLLM client key from .env"
            ) from exc
        value = str(dotenv_values(ROOT_DIR / ".env").get(CLIENT_KEY_ENV) or "").strip()
    if not value:
        raise Exit(f"{CLIENT_KEY_ENV} must be set in the shell or repo-local .env")
    return value


def _is_private_http_host(hostname):
    normalized = (hostname or "").strip().lower().rstrip(".")
    if normalized == "localhost" or "." not in normalized:
        return True
    if normalized.endswith((".local", ".internal", ".svc", ".svc.cluster.local")):
        return True
    try:
        address = ipaddress.ip_address(normalized)
    except ValueError:
        return False
    return address.is_private or address.is_loopback or address.is_link_local


def _normalize_endpoint(raw_endpoint):
    endpoint = str(raw_endpoint or "").strip()
    if not endpoint:
        raise Exit("--litellm-endpoint is required for the LiteLLM external proxy preflight")

    try:
        parsed = urllib.parse.urlsplit(endpoint)
        hostname = parsed.hostname
        parsed.port
    except ValueError as exc:
        raise Exit("--litellm-endpoint must be a valid absolute HTTP(S) URL") from exc

    if parsed.scheme not in {"http", "https"} or not parsed.netloc or not hostname:
        raise Exit("--litellm-endpoint must be a valid absolute HTTP(S) URL")
    if parsed.username or parsed.password:
        raise Exit("--litellm-endpoint must not contain credentials")
    if parsed.query or parsed.fragment or ";" in parsed.path:
        raise Exit(
            "--litellm-endpoint must not contain query parameters, fragments, or path parameters"
        )

    api_path = parsed.path.rstrip("/")
    if not api_path.endswith("/v1"):
        raise Exit("--litellm-endpoint must end with /v1")
    if parsed.scheme == "http" and not _is_private_http_host(hostname):
        raise Exit("public LiteLLM endpoints must use HTTPS")

    api_base = urllib.parse.urlunsplit((parsed.scheme, parsed.netloc, api_path, "", ""))
    service_path = api_path[:-3].rstrip("/")
    service_base = urllib.parse.urlunsplit((parsed.scheme, parsed.netloc, service_path, "", ""))
    display_target = urllib.parse.urlunsplit((parsed.scheme, parsed.netloc, "", "", ""))
    return api_base, service_base, display_target


def _probe_json(url, *, headers, timeout, label):
    request = urllib.request.Request(url, headers=headers, method="GET")
    try:
        with _open_without_redirects(request, timeout) as response:
            body = response.read(MAX_RESPONSE_BYTES + 1)
    except urllib.error.HTTPError as exc:
        if exc.code in {401, 403}:
            raise Exit(f"LiteLLM {label} rejected authentication (HTTP {exc.code})") from None
        raise Exit(f"LiteLLM {label} failed (HTTP {exc.code})") from None
    except (urllib.error.URLError, TimeoutError, OSError):
        raise Exit(f"LiteLLM {label} could not reach the configured proxy") from None
    except Exception:
        raise Exit(f"LiteLLM {label} failed with a normalized transport error") from None

    if len(body) > MAX_RESPONSE_BYTES:
        raise Exit(f"LiteLLM {label} response exceeded the safe size limit")
    try:
        return json.loads(body)
    except (json.JSONDecodeError, UnicodeDecodeError):
        raise Exit(f"LiteLLM {label} returned invalid JSON") from None


def _probe_status(url, *, timeout, label):
    request = urllib.request.Request(url, headers={"Accept": "application/json"}, method="GET")
    try:
        with _open_without_redirects(request, timeout):
            return
    except urllib.error.HTTPError as exc:
        raise Exit(f"LiteLLM {label} failed (HTTP {exc.code})") from None
    except (urllib.error.URLError, TimeoutError, OSError):
        raise Exit(f"LiteLLM {label} could not reach the configured proxy") from None
    except Exception:
        raise Exit(f"LiteLLM {label} failed with a normalized transport error") from None


def preflight(endpoint, api_key_env, model, timeout):
    if str(api_key_env or "").strip() != CLIENT_KEY_ENV:
        raise Exit(
            f"--litellm-api-key-env must be {CLIENT_KEY_ENV}; "
            "proxy administration keys are not accepted"
        )
    expected_model = str(model or "").strip()
    if not expected_model:
        raise Exit("--litellm-model is required for the LiteLLM external proxy preflight")
    try:
        timeout_seconds = float(timeout)
    except (TypeError, ValueError) as exc:
        raise Exit("--timeout must be a number from 1 through 30 seconds") from exc
    if timeout_seconds < 1 or timeout_seconds > 30:
        raise Exit("--timeout must be a number from 1 through 30 seconds")

    api_base, service_base, display_target = _normalize_endpoint(endpoint)
    client_key = _resolve_client_key()

    print("LiteLLM external proxy preflight")
    print(f"  Target      : {display_target}")
    print(f"  Credential  : env:{CLIENT_KEY_ENV} (scoped client/virtual key)")
    print("  Completion  : SKIPPED (no hosted inference request is sent)")

    _probe_status(
        f"{service_base}/health/liveliness",
        timeout=timeout_seconds,
        label="liveness probe",
    )
    print("  [OK] Liveness")
    _probe_status(
        f"{service_base}/health/readiness",
        timeout=timeout_seconds,
        label="readiness probe",
    )
    print("  [OK] Readiness")

    payload = _probe_json(
        f"{api_base}/models",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {client_key}",
        },
        timeout=timeout_seconds,
        label="authenticated model discovery",
    )
    data = payload.get("data") if isinstance(payload, dict) else None
    if not isinstance(data, list):
        raise Exit("LiteLLM authenticated model discovery returned an invalid response shape")
    model_ids = {
        str(item.get("id", "")).strip()
        for item in data
        if isinstance(item, dict) and str(item.get("id", "")).strip()
    }
    if expected_model not in model_ids:
        raise Exit(
            "LiteLLM authenticated model discovery did not expose the required model alias "
            f"{expected_model!r} ({len(model_ids)} model(s) visible)"
        )
    print(f"  [OK] Authenticated model alias: {expected_model}")
    print(
        "Result: CORRELATION-CAPABLE transport posture. Production enablement remains open; "
        "the provider stays disabled and non-swarm inference is not scope-correlated."
    )
