# Local P2P Relay (Starter)

This is the beginning of the local relay that will run on user devices.

**Goals for this component (from PRD):**
- Speak the exact same framing as the original rooms relay (MSG_UPDATE, MSG_CHECKPOINT, etc.)
- Manage per-room, per-peer queues
- Perform authenticated handshake using room secrets
- Handle P2P connections (start with TCP, move to WebRTC)
- Work together with the mandatory signaling service for initial bootstrap

**Current state:** Very basic TCP listener + connector + framing skeleton. For early testing of two local instances talking.

Run:
```
go run main.go -listen=:4501 -room=my-expenses
go run main.go -peer=localhost:4501 -room=my-expenses
```

Next steps (see PRD Phase 1):
- Add real queue + cursor logic
- Add crypto handshake (reuse rooms deriveRoom / deriveChannelKey)
- Integrate logical clocks
- Wire to local persistence

The signaling service (separate) will be used to obtain initial connection candidates.