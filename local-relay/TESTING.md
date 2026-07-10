# Thin P2P + Signaling Testing Guide (Stage 2)

## Local direct TCP test (still works)

Terminal 1 (listener):
```bash
go run main.go -listen=:4501 -room=test-expenses
```

Terminal 2 (connector):
```bash
go run main.go -peer=localhost:4501 -room=test-expenses
```

Type `send` + Enter after you see the connected message. You should see the update received on the other side.

## Real test with signaling (desktop + phone)

### 1. Run the signaling server

```bash
cd ../signaling
go run main.go -addr=:8081
```

### 2. Expose it (required for phone)

Fastest:
```bash
ngrok http 8081
```
Copy the https://...ngrok-free.app URL.

### 3. Test with two browsers (recommended for phone test)

Open the client on **two different devices** (e.g. your computer + phone), both pointing at the same ngrok signaling URL.

Exact sequence:

1. On **both** devices:
   - Paste the ngrok https URL
   - Use the same Room ID
   - Click **Register**

2. On **one** device only:
   - Click **Create Offer & Send**

3. On the **other** device:
   - Click **Poll for Answer**

4. Wait until one side says "P2P Connected" (green status).

5. Click **Send Test Update** on either side.

You should see the message appear in the log on the receiving side.

**Debug tips if it doesn't work:**
- "Csignal: interrupt" or "signal: interrupt" = you pressed Ctrl+C (Unix signal) to stop the previous run. Normal, ignore it.
- When you see "Tiny signaling server on :8081" the server is running and waiting. It stays quiet until browsers actually talk to it.
- You will only see "CORS set ..." and "offer/answer received" lines *after* the browsers make requests.
- Port 8005 (or 8000) is your *local HTML server* (python -m http.server). The signaling log shows the browser's Origin (where the page was loaded from), not the signaling port.
- To see activity from port 8005: make sure one browser tab is served from that port and uses the ngrok URL for signaling, then click the buttons.
- Always use the *ngrok https URL* for Signaling base URL on every client (computer and phone).
- Hard refresh clients after changes.

### 4. Test with Go (desktop) + browser (phone)

The current Go thin layer still works best with direct `--peer` for local.

For real cross-network with the browser client:
- Run the Go side with signaling (the signaling mode is partially wired for discovery).
- The browser side does full WebRTC.
- For full Go <-> browser WebRTC we will add pion/webrtc in a later thin iteration if needed.

For now the browser <-> browser test via signaling is the quickest way to validate the "small centralized signaling + real P2P" model on your phone.

### What the signaling does
- Only helps with the very first connection (WebRTC offer/answer + ICE candidates, or simple addr exchange).
- Once the direct P2P DataChannel is up, everything (including future "framed updates") flows directly between the two devices.
- This is the minimal gate we will use for the monetization path.

Let me know what you see when you try the browser-to-browser test with ngrok! Then we can wire more of the real framing / queues / crypto.