#!/usr/bin/env python3
"""
post_to_x.py — Post a draft to X using the logged-in Chrome Default profile (Playwright)

Usage:
  python post_to_x.py --draft 20260704-naval-taiwan-competition

Requirements:
  pip install playwright
  playwright install chromium
"""

import argparse
import json
from pathlib import Path
from playwright.sync_api import sync_playwright, TimeoutError as PlaywrightTimeout

CHROME_PROFILE = Path.home() / "Library/Application Support/Google/Chrome/Default"
X_COMPOSE_URL = "https://x.com/compose/post"

def post_draft(draft_id: str):
    draft_dir = Path(__file__).parent / "drafts" / draft_id
    meta_path = draft_dir / "meta.json"
    clip_path = draft_dir / "clip.mp4"

    if not meta_path.exists():
        print(f"ERROR: {meta_path} not found")
        return

    meta = json.loads(meta_path.read_text())
    post_text = meta.get("post_text", "").strip()

    if not clip_path.exists():
        print(f"ERROR: {clip_path} not found")
        return

    print(f"Posting draft: {draft_id}")
    print(f"Text: {post_text[:80]}...")

    with sync_playwright() as p:
        browser = p.chromium.launch_persistent_context(
            user_data_dir=str(CHROME_PROFILE),
            channel="chrome",
            headless=False,
            args=[
                "--disable-blink-features=AutomationControlled",
                "--no-default-browser-check",
            ],
        )

        page = browser.new_page()
        page.goto(X_COMPOSE_URL, wait_until="domcontentloaded")

        # Wait for the compose box
        try:
            page.wait_for_selector('div[role="textbox"]', timeout=15000)
        except PlaywrightTimeout:
            print("Could not find compose textbox. Make sure you're logged in.")
            browser.close()
            return

        # Paste the text
        textbox = page.locator('div[role="textbox"]').first
        textbox.click()
        textbox.fill(post_text)

        # Attach the video
        file_input = page.locator('input[type="file"]')
        file_input.set_input_files(str(clip_path))

        # Wait a moment for the media to upload
        page.wait_for_timeout(3000)

        # Click Post button
        post_button = page.locator('button[data-testid="tweetButton"]')
        post_button.click()

        print("Post submitted. Check the browser window.")
        page.wait_for_timeout(5000)
        browser.close()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--draft", required=True, help="Draft ID (folder name)")
    args = parser.parse_args()
    post_draft(args.draft)
