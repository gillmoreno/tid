# Signaling API v2

Production-shaped SQLite-backed signaling and encrypted-mailbox service. The server treats every
`envelope` as opaque ciphertext and never accepts, stores, logs, or returns a room data key.

Initialize a local Ed25519 signer, export only its public key, and run the API:

```sh
go run ./cmd/roomworks-token init --seed-file /tmp/roomworks-creator-signing.seed
export ROOMWORKS_CREATOR_VERIFY_KEY="$(go run ./cmd/roomworks-token public-key \
  --seed-file /tmp/roomworks-creator-signing.seed)"
go run .
```

Minting remains offline and copies the permit directly to the macOS clipboard:

```sh
go run ./cmd/roomworks-token mint --capacity 3 --ttl 24h \
  --seed-file /tmp/roomworks-creator-signing.seed
```

The public pilot runs in the repository's dedicated unified Docker image. See
`docs_and_changelog/2026-07-12-public-room-pilot-deployment.md`; do not add this service to the
existing TID Compose service.

`SIGNALING_ADDR`, `SIGNALING_DB_PATH`, and comma-separated
`SIGNALING_ALLOWED_ORIGINS` are supported. `ROOMWORKS_CREATOR_VERIFY_KEY` is required and must be a
base64url Ed25519 public key. The private seed never enters the Go service. The default bind is
`127.0.0.1:8081` and the default browser origin is `http://localhost:5200`. Requests without an
`Origin` header (native clients and tests) are allowed.

## Credentials and admission

- `POST /v2/rooms` — requires a one-use `X-Room-Creator-Permit` whose signed capacity N exactly
  matches `{ "maxMembers": N }`. Capacity includes the owner and can be 2 through 50. Returns a
  public random `roomId`, an `ownerCapability`, and the owner's member/device IDs and member
  credential. Provisioning credentials are returned once and stored only as SHA-256 hashes. Permit
  consumption and room creation share one transaction; only the permit's token-ID hash is stored.
- `GET /v2/rooms/{roomId}` — bearer member authentication. Returns capacity and durable member
  count.
- `GET /v2/rooms/{roomId}/devices` — bearer member authentication. Returns admitted public
  device IDs plus `isOwner` and caller-relative `isSelf` flags for WebRTC rendezvous. It never
  returns member credentials, identity/idempotency hashes, owner capabilities, or invite data.
- `POST /v2/rooms/{roomId}/invites` — owner authentication with
  `X-Owner-Capability`. Accepts `{ "expiresInSeconds": 86400 }`; returns separate random
  `inviteId` (locator) and one-time `inviteSecret`.
- `DELETE /v2/rooms/{roomId}/invites/{inviteId}` — owner-authenticated revocation.
- `POST /v2/invites/{inviteId}/redeem` — accepts:

```json
{
  "inviteSecret": "secret-from-the-invite",
  "deviceId": "public_stable_device_id",
  "deviceIdentity": "client-generated-random-secret-at-least-16-chars",
  "memberCredential": "client-generated-random-secret-at-least-16-chars",
  "idempotencyKey": "stable-retry-key"
}
```

The client-generated member credential lets a lost redemption response be retried without the
server retaining plaintext credentials. Redemption is transactional. A retry by the same device
returns the same member; another device cannot replay the one-time invite. Only durable member
rows consume room capacity.

Member-authenticated requests use `Authorization: Bearer {memberCredential}`. Owner authority is
separate and used only for invite administration.

## Session signaling

- `POST /v2/rooms/{roomId}/sessions/{sessionId}/signals`
- `GET /v2/rooms/{roomId}/sessions/{sessionId}/signals?after={signalId}`

POST body:

```json
{
  "kind": "offer",
  "fromDeviceId": "authenticated-device-id",
  "toDeviceId": "admitted-recipient-device-id",
  "envelope": "opaque-encrypted-or-SDP-string",
  "expiresInSeconds": 600
}
```

`kind` is `offer`, `answer`, or `candidate`. Reads return only signals addressed to the
authenticated device and only from the named session. Signals expire (10-minute default,
one-hour maximum), are limited to 64 KiB each and 512 live records per room/session, and never
affect room capacity.

## Encrypted mailbox

- `PUT|GET /v2/rooms/{roomId}/mailbox/checkpoint` — canonical room bootstrap checkpoint. Any
  admitted member can replace it and any admitted member can read the latest non-expired value.
  PUT body is `{ "envelope": "...", "expiresInSeconds": 604800 }`. GET returns the opaque
  envelope, `uploaderDeviceId`, `updatedAt`, and `expiresAt`. This room-wide scope lets a
  brand-new invitee bootstrap while every pre-existing member is offline.
- `PUT|GET /v2/rooms/{roomId}/mailbox/{deviceId}/checkpoint` — compatibility endpoint for
  device-private checkpoints; only that authenticated device can replace/read it. New clients
  should use the room-wide checkpoint for shared room bootstrap.
- `POST|GET /v2/rooms/{roomId}/mailbox/{deviceId}/operations?after={operationId}` — any admitted
  member can POST an opaque operation to another admitted device; only the recipient can read it.

Room and per-device checkpoints are limited to 256 KiB and stored only as opaque envelopes with
uploader/device and timestamp metadata. Operations are limited to 64 KiB and 1,000 live records
per recipient. Mailbox TTL defaults to seven days and is capped at 30 days.

All errors are JSON:

```json
{ "error": { "code": "room_full", "message": "room has reached its member capacity" } }
```

The migrations create separate `v2_*` tables (including the v3 room-checkpoint addition) and do
not alter or drop prototype tables in an existing `signaling.db`. Prototype `/admin`, `/dev/*`,
`/register`, `/get`, and `/room/*` endpoints are not registered.
