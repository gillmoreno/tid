#!/usr/bin/env python3
"""Extract clean article text + title from a URL.

Primary path uses trafilatura (best quality). Falls back to a urllib + very
small HTML tag stripper when trafilatura is unavailable, so the loop still
produces *something* in dev environments.

Usage:
    python3 extract_article.py --url URL --out drafts/{id}
Writes:
    {out}/article.txt   plain text
    {out}/article.json  {"title": ..., "text": ...}
Prints the extracted title to stdout.
"""

import argparse
import json
import os
import sys


def extract_with_trafilatura(url: str):
    import trafilatura  # type: ignore

    downloaded = trafilatura.fetch_url(url)
    if not downloaded:
        return None, None
    text = trafilatura.extract(
        downloaded,
        include_comments=False,
        include_tables=False,
        favor_precision=True,
    )
    title = None
    try:
        meta = trafilatura.extract_metadata(downloaded)
        if meta and getattr(meta, "title", None):
            title = meta.title
    except Exception:
        title = None
    return title, text


def extract_fallback(url: str):
    import re
    import urllib.request
    from html import unescape

    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0 (TID Article Factory)"})
    with urllib.request.urlopen(req, timeout=30) as resp:
        raw = resp.read().decode("utf-8", errors="ignore")

    title_match = re.search(r"<title[^>]*>(.*?)</title>", raw, re.I | re.S)
    title = unescape(title_match.group(1)).strip() if title_match else None

    # Drop script/style/nav/header/footer blocks, then strip remaining tags.
    body = re.sub(r"(?is)<(script|style|noscript|nav|header|footer|aside)[^>]*>.*?</\1>", " ", raw)
    body = re.sub(r"(?is)<br\s*/?>", "\n", body)
    body = re.sub(r"(?is)</(p|div|li|h[1-6])>", "\n", body)
    body = re.sub(r"(?s)<[^>]+>", " ", body)
    body = unescape(body)
    lines = [ln.strip() for ln in body.splitlines()]
    lines = [ln for ln in lines if len(ln) > 40]
    text = "\n\n".join(lines)
    return title, text


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--url", required=True)
    parser.add_argument("--out", required=True, help="output directory")
    args = parser.parse_args()

    title, text = None, None
    try:
        title, text = extract_with_trafilatura(args.url)
    except ImportError:
        print("trafilatura not installed — using fallback extractor", file=sys.stderr)
    except Exception as exc:  # noqa: BLE001
        print(f"trafilatura failed ({exc}) — using fallback extractor", file=sys.stderr)

    if not text:
        try:
            f_title, text = extract_fallback(args.url)
            title = title or f_title
        except Exception as exc:  # noqa: BLE001
            print(f"fallback extractor failed: {exc}", file=sys.stderr)
            return 1

    if not text or not text.strip():
        print("no article text extracted", file=sys.stderr)
        return 1

    os.makedirs(args.out, exist_ok=True)
    title = (title or "").strip()
    with open(os.path.join(args.out, "article.txt"), "w", encoding="utf-8") as fh:
        fh.write(text.strip() + "\n")
    with open(os.path.join(args.out, "article.json"), "w", encoding="utf-8") as fh:
        json.dump({"title": title, "text": text.strip()}, fh, indent=2)

    print(title)
    return 0


if __name__ == "__main__":
    sys.exit(main())
