from __future__ import annotations

import time
import urllib.request
from collections.abc import Iterable


def _read_page(url: str) -> tuple[int, str]:
    last_error: Exception | None = None
    for attempt in range(1, 5):
        try:
            req = urllib.request.Request(url, headers={"Connection": "close"})
            with urllib.request.urlopen(req, timeout=10) as resp:
                return resp.status, resp.read().decode("utf-8", errors="replace")
        except Exception as exc:
            last_error = exc
            if "WinError 10048" not in str(exc) or attempt == 4:
                raise
            time.sleep(2 * attempt)
    raise last_error or RuntimeError("interface probe failed")


def probe_pages(base: str, pages: Iterable[str]) -> list[str]:
    errors: list[str] = []
    for page in pages:
        url = f"{base}{page}"
        try:
            status, body = _read_page(url)
            issues: list[str] = []
            if "NEXT_REDIRECT" in body and "404" in body:
                issues.append("404 redirect detected")
            if "Internal Server Error" in body:
                issues.append("500 Internal Server Error")
            if "__next_error__" in body:
                issues.append("Next.js error boundary triggered")
            if "Application error" in body or "Unhandled Runtime Error" in body:
                issues.append("React runtime error detected")
            if "bg-white" in body and page in ("/wiring", "/architect"):
                issues.append("Light-mode bg-white leak detected")

            ok = status == 200 and not issues
            icon = "[OK]" if ok else "[FAIL]"
            print(f"  {icon} {page} [{status}]", end="")
            if issues:
                print(f"  WARN: {', '.join(issues)}")
                errors.extend([f"{page}: {issue}" for issue in issues])
            else:
                print()
            time.sleep(0.2)
        except Exception as exc:
            print(f"  [FAIL] {page} - {exc}")
            errors.append(f"{page}: {exc}")
    return errors
