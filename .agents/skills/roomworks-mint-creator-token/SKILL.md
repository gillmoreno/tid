---
name: roomworks-mint-creator-token
description: Mint a one-use, capacity-bound Roomworks room creator token locally and copy it to the macOS clipboard without calling the API or printing secrets. Use when Gil asks to create, mint, or get a Rooms/Roomworks token for a specific number of people, such as 3, 10, or 11 people.
---

# Mint Roomworks Creator Token

Interpret the requested number as the room's total unique-member capacity, including the creator.
Require an integer from 2 through 50. Ask for the capacity only when it is missing or ambiguous.

Use a 24-hour lifetime unless Gil requests another duration. The supported lifetime is 5 minutes
through 7 days.

Run the bundled script from this skill directory:

```sh
scripts/mint-token.sh --capacity CAPACITY --ttl 24h
```

The script signs locally with `~/.config/roomworks/creator-signing.seed` and copies the permit
directly to the clipboard. It does not call the production API.

Never print or log the permit, private seed, environment contents, or clipboard contents. Never
generate, replace, or rotate the signing seed implicitly. On success, report only that the token
was copied, its capacity, expiration, and that it is single-use. Tell Gil to paste it into the
**Room creator token** field.

On failure, report the exact missing path or permission problem. Do not weaken file permissions or
fall back to the old deployment-wide pilot token.
