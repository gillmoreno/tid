#!/usr/bin/env python3
"""End-to-end test for the two-member local-first room flow.

Requires the meta app on :5200 and signaling API v2 on :8081.
Uses isolated browser contexts to model separate devices.
"""

from __future__ import annotations

import base64
import json
import os
import time
import urllib.error
import urllib.request
from typing import Any

from playwright.sync_api import BrowserContext, Page, sync_playwright


APP_URL = os.environ.get("META_APP_URL", "http://localhost:5200")
SIGNALING_URL = os.environ.get("SIGNALING_URL", "http://localhost:8081")
CREATOR_PERMIT = os.environ.get("ROOMWORKS_CREATOR_PERMIT", "")
CHROME_PATH = os.environ.get(
    "PLAYWRIGHT_CHROME_PATH",
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
)


def api_request(
    method: str,
    path: str,
    body: dict[str, Any] | None = None,
    headers: dict[str, str] | None = None,
) -> tuple[int, dict[str, Any]]:
    data = None if body is None else json.dumps(body).encode()
    request_headers = {
        "Accept": "application/json",
        "Origin": APP_URL,
        **({"Content-Type": "application/json"} if data is not None else {}),
        **(headers or {}),
    }
    request = urllib.request.Request(
        f"{SIGNALING_URL}{path}",
        data=data,
        headers=request_headers,
        method=method,
    )
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as error:
        raw = error.read()
        return error.code, json.loads(raw) if raw else {}


def decode_package(value: str) -> dict[str, Any]:
    assert value.startswith("roompkg1.")
    payload = value.removeprefix("roompkg1.")
    payload += "=" * (-len(payload) % 4)
    return json.loads(base64.urlsafe_b64decode(payload))


def encode_package(value: dict[str, Any]) -> str:
    payload = base64.urlsafe_b64encode(
        json.dumps(value, separators=(",", ":")).encode()
    ).decode().rstrip("=")
    return f"roompkg1.{payload}"


def vault_room(page: Page, room_id: str) -> dict[str, Any]:
    return page.evaluate(
        """async (roomId) => {
          const database = await new Promise((resolve, reject) => {
            const request = indexedDB.open('meta-room-vault');
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
          });
          return await new Promise((resolve, reject) => {
            const request = database.transaction('rooms', 'readonly')
              .objectStore('rooms').get(roomId);
            request.onsuccess = () => {
              const room = request.result;
              resolve(room ? {
                id: room.id,
                memberCredential: room.memberCredential,
                ownerCapability: room.ownerCapability,
                deviceId: room.deviceId,
              } : null);
            };
            request.onerror = () => reject(request.error);
          });
        }""",
        room_id,
    )


def room_count(page: Page) -> str:
    frame = page.frame_locator("iframe.room-frame")
    frame.locator("#count").wait_for(state="visible")
    return frame.locator("#count").inner_text()


def wait_for_count(page: Page, expected: int, timeout_ms: int = 20_000) -> None:
    frame = page.frame_locator("iframe.room-frame")
    frame.locator("#count").wait_for(state="visible")
    deadline = time.monotonic() + timeout_ms / 1000
    while time.monotonic() < deadline:
        if frame.locator("#count").inner_text() == str(expected):
            return
        page.wait_for_timeout(250)
    status = frame.locator("#status").inner_text()
    raise AssertionError(
        f"expected counter {expected}, saw {frame.locator('#count').inner_text()}; "
        f"bridge status: {status}"
    )


def click_increment(page: Page) -> None:
    page.frame_locator("iframe.room-frame").get_by_role(
        "button", name="Add one"
    ).click()


def wait_for_p2p(page: Page) -> None:
    page.get_by_text("Live peer channel", exact=True).wait_for(timeout=25_000)


def new_context(browser: Any) -> BrowserContext:
    context = browser.new_context()
    context.grant_permissions(
        ["clipboard-read", "clipboard-write"], origin=APP_URL
    )
    return context


def run() -> None:
    assert CREATOR_PERMIT, "ROOMWORKS_CREATOR_PERMIT is required"
    status, health = api_request("GET", "/healthz")
    assert status == 200 and health == {"status": "ok"}

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(
            executable_path=CHROME_PATH,
            headless=True,
        )
        owner = new_context(browser)
        invalid_joiner = new_context(browser)
        invitee = new_context(browser)
        third_joiner = new_context(browser)

        owner_page = owner.new_page()
        owner_page.goto(APP_URL, wait_until="networkidle")
        owner_page.get_by_role("button", name="Create a room").click()
        owner_page.get_by_label("Room creator token").fill(CREATOR_PERMIT)
        owner_page.get_by_role("button", name="Unlock creation").click()
        owner_page.get_by_label("Room purpose").fill("E2E shared counter")
        capacity_input = owner_page.get_by_label("Unique member capacity")
        assert capacity_input.input_value() == "2"
        assert capacity_input.get_attribute("readonly") is not None
        owner_page.get_by_role(
            "button", name="Create room & invitation"
        ).click()

        share_card = owner_page.get_by_label("Room invitation")
        share_card.wait_for(state="visible")
        invitation_url = share_card.locator(".copy-block code").nth(0).inner_text()
        invitation_package = share_card.locator(".copy-block code").nth(1).inner_text()
        package_data = decode_package(invitation_package)
        room_id = package_data["roomId"]
        owner_room_url = f"{APP_URL}/rooms/{room_id}"
        owner_data = vault_room(owner_page, room_id)
        assert owner_data and owner_data["ownerCapability"]
        owner_page.goto(owner_room_url, wait_until="networkidle")
        wait_for_count(owner_page, 0)

        # A syntactically valid package with the wrong secret must not create a room.
        wrong_data = dict(package_data)
        wrong_data["inviteSecret"] = "wrong-secret-" + "x" * 32
        wrong_page = invalid_joiner.new_page()
        wrong_page.goto(invitation_url, wait_until="networkidle")
        wrong_page.get_by_label("Invitation package").fill(
            encode_package(wrong_data)
        )
        wrong_page.get_by_role("button", name="Join encrypted room").click()
        wrong_page.get_by_role("alert").wait_for(state="visible")
        wrong_page.goto(APP_URL, wait_until="networkidle")
        assert wrong_page.locator(".room-row").count() == 0
        invalid_joiner.close()

        # Close every owner page: the invitee must bootstrap from the encrypted mailbox.
        owner_page.close()
        invitee_page = invitee.new_page()
        invitee_page.goto(invitation_url, wait_until="networkidle")
        invitee_page.get_by_label("Invitation package").fill(
            invitation_package
        )
        invitee_page.get_by_role("button", name="Join encrypted room").click()
        invitee_page.wait_for_url(f"**/rooms/{room_id}")
        assert invitee_page.url == f"{APP_URL}/rooms/{room_id}"
        wait_for_count(invitee_page, 0)

        click_increment(invitee_page)
        click_increment(invitee_page)
        wait_for_count(invitee_page, 2)
        invitee_page.wait_for_timeout(5_000)

        # Reopen the owner from the same vault and converge from mailbox operations.
        reopened_owner = owner.new_page()
        reopened_owner.goto(owner_room_url, wait_until="networkidle")
        wait_for_count(reopened_owner, 2)
        wait_for_p2p(reopened_owner)
        wait_for_p2p(invitee_page)

        # Re-redeeming on the same admitted device is idempotent.
        retry_page = invitee.new_page()
        retry_page.goto(invitation_url, wait_until="networkidle")
        retry_page.get_by_label("Invitation package").fill(invitation_package)
        retry_page.get_by_role("button", name="Join encrypted room").click()
        retry_page.wait_for_url(f"**/rooms/{room_id}")
        retry_page.close()

        status, room_info = api_request(
            "GET",
            f"/v2/rooms/{room_id}",
            headers={"Authorization": f"Bearer {owner_data['memberCredential']}"},
        )
        assert status == 200
        assert room_info["memberCount"] == 2

        # A second valid invite cannot admit a third unique device.
        status, second_invite = api_request(
            "POST",
            f"/v2/rooms/{room_id}/invites",
            {"expiresInSeconds": 3600},
            {"X-Owner-Capability": owner_data["ownerCapability"]},
        )
        assert status == 201
        full_package_data = dict(package_data)
        full_package_data["inviteId"] = second_invite["inviteId"]
        full_package_data["inviteSecret"] = second_invite["inviteSecret"]
        full_url = f"{APP_URL}/join/{second_invite['inviteId']}"
        third_page = third_joiner.new_page()
        third_page.goto(full_url, wait_until="networkidle")
        third_page.get_by_label("Invitation package").fill(
            encode_package(full_package_data)
        )
        third_page.get_by_role("button", name="Join encrypted room").click()
        third_page.get_by_role("alert").wait_for(state="visible")
        assert "capacity" in third_page.get_by_role("alert").inner_text().lower()
        third_page.goto(APP_URL, wait_until="networkidle")
        assert third_page.locator(".room-row").count() == 0

        # Concurrent increments converge in both isolated browser vaults.
        click_increment(reopened_owner)
        click_increment(invitee_page)
        wait_for_count(reopened_owner, 4)
        wait_for_count(invitee_page, 4)

        # The obsolete prototype/admin surfaces stay disabled.
        status, _ = api_request("GET", "/admin")
        assert status == 404

        owner.close()
        invitee.close()
        third_joiner.close()
        browser.close()

    print(
        "PASS: offline join, idempotent reconnect, durable capacity, "
        "same-origin routing, and convergent counter sync"
    )


if __name__ == "__main__":
    run()
