#!/usr/bin/env python3
"""Run one text-generation request through the OpenAI Responses API."""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any


DEFAULT_MODEL = "gpt-5.6-sol"
DEFAULT_REASONING_EFFORT = "medium"
DEFAULT_MAX_OUTPUT_TOKENS = 16_000
DEFAULT_TIMEOUT_SECONDS = 300


def build_payload(
    prompt: str,
    model: str,
    reasoning_effort: str,
    max_output_tokens: int,
) -> dict[str, Any]:
    return {
        "model": model,
        "reasoning": {"effort": reasoning_effort},
        "input": prompt,
        "max_output_tokens": max_output_tokens,
    }


def extract_output_text(response: dict[str, Any]) -> str:
    direct = response.get("output_text")
    if isinstance(direct, str) and direct.strip():
        return direct.strip()

    chunks: list[str] = []
    for item in response.get("output", []):
        if not isinstance(item, dict) or item.get("type") != "message":
            continue
        for content in item.get("content", []):
            if not isinstance(content, dict) or content.get("type") != "output_text":
                continue
            text = content.get("text")
            if isinstance(text, str) and text:
                chunks.append(text)
    result = "\n".join(chunks).strip()
    if not result:
        raise ValueError("OpenAI response contained no output text")
    return result


def response_url(base_url: str) -> str:
    base = base_url.rstrip("/")
    if not base.endswith("/v1"):
        base += "/v1"
    return base + "/responses"


def api_error_message(raw: bytes, status: int) -> str:
    try:
        payload = json.loads(raw)
        message = payload.get("error", {}).get("message")
        if isinstance(message, str) and message.strip():
            return f"OpenAI API error HTTP {status}: {message.strip()}"
    except (json.JSONDecodeError, AttributeError):
        pass
    return f"OpenAI API error HTTP {status}"


def create_response(
    *,
    prompt: str,
    api_key: str,
    model: str,
    reasoning_effort: str,
    max_output_tokens: int,
    timeout_seconds: int,
    base_url: str,
) -> str:
    payload = build_payload(prompt, model, reasoning_effort, max_output_tokens)
    request = urllib.request.Request(
        response_url(base_url),
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as result:
            response = json.load(result)
    except urllib.error.HTTPError as exc:
        raise RuntimeError(api_error_message(exc.read(), exc.code)) from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"OpenAI API request failed: {exc.reason}") from exc
    return extract_output_text(response)


def positive_int_env(name: str, default: int) -> int:
    raw = os.getenv(name, str(default))
    try:
        value = int(raw)
    except ValueError as exc:
        raise ValueError(f"{name} must be an integer") from exc
    if value <= 0:
        raise ValueError(f"{name} must be positive")
    return value


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--prompt-file", required=True)
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    api_key = os.getenv("OPENAI_API_KEY", "").strip()
    if not api_key:
        print("OPENAI_API_KEY is not set", file=sys.stderr)
        return 2

    try:
        text = create_response(
            prompt=Path(args.prompt_file).read_text(encoding="utf-8"),
            api_key=api_key,
            model=os.getenv("FACTORY_OPENAI_MODEL", DEFAULT_MODEL),
            reasoning_effort=os.getenv(
                "FACTORY_OPENAI_REASONING_EFFORT", DEFAULT_REASONING_EFFORT
            ),
            max_output_tokens=positive_int_env(
                "FACTORY_OPENAI_MAX_OUTPUT_TOKENS", DEFAULT_MAX_OUTPUT_TOKENS
            ),
            timeout_seconds=positive_int_env(
                "FACTORY_OPENAI_TIMEOUT_SECONDS", DEFAULT_TIMEOUT_SECONDS
            ),
            base_url=os.getenv("OPENAI_BASE_URL", "https://api.openai.com"),
        )
        Path(args.out).write_text(text + "\n", encoding="utf-8")
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        print(str(exc), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
