from __future__ import annotations

import json
from hashlib import sha256
from typing import Any

from fastapi import FastAPI, HTTPException, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from starlette.exceptions import HTTPException as StarletteHTTPException


def install_error_handlers(app: FastAPI) -> None:
    @app.exception_handler(RequestValidationError)
    async def validation_error(
        _request: Request, _exc: RequestValidationError
    ) -> JSONResponse:
        return error_response(422, "invalid_request", "Request validation failed.")

    @app.exception_handler(StarletteHTTPException)
    async def http_error(
        _request: Request, exc: StarletteHTTPException
    ) -> JSONResponse:
        detail = exc.detail if isinstance(exc.detail, dict) else {
            "code": "http_error",
            "message": str(exc.detail),
            "recoverable": False,
        }
        return JSONResponse(status_code=exc.status_code, content={"error": detail})


def fail(status_code: int, code: str, message: str, recoverable: bool = False) -> None:
    raise HTTPException(
        status_code=status_code,
        detail={"code": code, "message": message, "recoverable": recoverable},
    )


def error_response(
    status_code: int, code: str, message: str, recoverable: bool = False
) -> JSONResponse:
    return JSONResponse(
        status_code=status_code,
        content={
            "error": {
                "code": code,
                "message": message,
                "recoverable": recoverable,
            }
        },
    )


def request_fingerprint(payload: BaseModel, *, discriminator: str = "") -> str:
    normalized: dict[str, Any] = payload.model_dump(mode="json")
    if discriminator:
        normalized["_contract_discriminator"] = discriminator
    encoded = json.dumps(normalized, sort_keys=True, separators=(",", ":")).encode()
    return sha256(encoded).hexdigest()


def parse_last_event_id(value: str | None) -> int:
    if value is None or value == "":
        return 0
    if not value.isascii() or not value.isdecimal():
        fail(422, "invalid_cursor", "Last-Event-ID must be a decimal sequence.")
    return int(value)
