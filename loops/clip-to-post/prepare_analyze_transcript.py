#!/usr/bin/env python3
"""Prepare timestamped transcript for analysis — full episode coverage, not just the opening."""

import re
import sys

MAX_CHARS = 52_000
NUM_BUCKETS = 8
PRIORITY_TERMS = (
    "ai",
    "llm",
    "compute",
    "money",
    "billion",
    "million",
    "chip",
    "nvidia",
    "openai",
    "anthropic",
    "startup",
    "market",
    "invest",
    "moore",
    "robot",
    "agent",
    "scale",
    "profit",
    "cost",
    "dollar",
)

TS_RE = re.compile(r"^\[(\d{2}):(\d{2}):(\d{2})\]")


def parse_timestamp(line: str) -> int | None:
    match = TS_RE.match(line.strip())
    if not match:
        return None
    h, m, s = map(int, match.groups())
    return h * 3600 + m * 60 + s


def parse_lines(text: str) -> list[tuple[int, str]]:
    lines: list[tuple[int, str]] = []
    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            continue
        ts = parse_timestamp(line)
        if ts is None:
            continue
        lines.append((ts, line))
    return lines


def score_line(line: str) -> int:
    lower = line.lower()
    return sum(1 for term in PRIORITY_TERMS if term in lower)


def bucket_lines(lines: list[tuple[int, str]]) -> list[list[tuple[int, str]]]:
    if not lines:
        return []

    start = lines[0][0]
    end = lines[-1][0]
    span = max(end - start, 1)
    bucket_size = span / NUM_BUCKETS

    buckets: list[list[tuple[int, str]]] = [[] for _ in range(NUM_BUCKETS)]
    for ts, line in lines:
        idx = min(int((ts - start) / bucket_size), NUM_BUCKETS - 1)
        buckets[idx].append((ts, line))
    return buckets


def select_bucket_lines(bucket: list[tuple[int, str]], budget: int) -> list[str]:
    if not bucket:
        return []
    if sum(len(line) + 1 for _, line in bucket) <= budget:
        return [line for _, line in bucket]

    ranked = sorted(
        bucket,
        key=lambda item: (score_line(item[1]), len(item[1])),
        reverse=True,
    )
    chosen: list[tuple[int, str]] = []
    used = 0
    for ts, line in ranked:
        cost = len(line) + 1
        if used + cost > budget:
            continue
        chosen.append((ts, line))
        used += cost

    if not chosen:
        ts, line = ranked[0]
        return [line[: max(budget - 1, 0)]]

    chosen.sort(key=lambda item: item[0])
    return [line for _, line in chosen]


def prepare(text: str, max_chars: int = MAX_CHARS) -> str:
    lines = parse_lines(text)
    if not lines:
        return text[:max_chars]

    plain = "\n".join(line for _, line in lines)
    if len(plain) <= max_chars:
        return plain

    buckets = bucket_lines(lines)
    non_empty = [b for b in buckets if b]
    per_bucket = max(max_chars // max(len(non_empty), 1), 1200)

    out: list[str] = []
    used = 0
    for i, bucket in enumerate(buckets):
        if not bucket:
            continue
        if i > 0:
            marker = f"\n[--- segment {i + 1}/{NUM_BUCKETS} of episode ---]\n"
            if used + len(marker) > max_chars:
                break
            out.append(marker.strip())
            used += len(marker)

        remaining = max_chars - used
        selected = select_bucket_lines(bucket, min(per_bucket, remaining))
        for line in selected:
            cost = len(line) + 1
            if used + cost > max_chars:
                break
            out.append(line)
            used += cost

    return "\n".join(out)


def main() -> None:
    text = sys.stdin.read()
    sys.stdout.write(prepare(text))


if __name__ == "__main__":
    main()