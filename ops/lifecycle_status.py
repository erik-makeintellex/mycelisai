"""
Shared lifecycle status probe helpers.
"""

from collections.abc import Callable


def loopback_probe_urls(port: int, path: str, host: str = "localhost") -> list[str]:
    candidates = [host, "localhost", "127.0.0.1", "::1"]
    urls: list[str] = []
    seen: set[str] = set()
    normalized_path = path if path.startswith("/") else f"/{path}"
    for candidate in candidates:
        if not candidate:
            continue
        url_host = candidate
        if ":" in candidate and not candidate.startswith("["):
            url_host = f"[{candidate}]"
        url = f"http://{url_host}:{port}{normalized_path}"
        if url in seen:
            continue
        seen.add(url)
        urls.append(url)
    return urls


def first_healthy_http_probe(
    urls: list[str],
    http_get: Callable[[str, float], tuple[int, str]],
    timeout: float = 2.0,
) -> tuple[int, str, str]:
    last_code = 0
    last_url = urls[0] if urls else ""
    last_body = ""
    for url in urls:
        last_code, last_body = http_get(url, timeout)
        last_url = url
        if last_code == 200:
            return last_code, url, last_body
    return last_code, last_url, last_body


def status_service_alive(
    key: str,
    port: int,
    http_get: Callable[[str, float], tuple[int, str]],
    port_open: Callable[[int], bool],
    api_host: str,
) -> bool:
    if key == "core":
        code, _url, _body = first_healthy_http_probe(
            loopback_probe_urls(port, "/healthz", api_host),
            http_get,
        )
        return code == 200 or port_open(port)
    if key == "ollama":
        code, _url, _body = first_healthy_http_probe(
            loopback_probe_urls(port, "/api/tags", "localhost"),
            http_get,
        )
        return code == 200 or port_open(port)
    return port_open(port)
