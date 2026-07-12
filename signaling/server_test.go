package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type testAPI struct {
	t       *testing.T
	server  *server
	handler http.Handler
}

type roomFixture struct {
	RoomID          string
	OwnerCapability string
	OwnerMemberID   string
	OwnerDeviceID   string
	OwnerCredential string
}

type inviteFixture struct {
	ID     string
	Secret string
}

type memberFixture struct {
	ID         string
	DeviceID   string
	Credential string
	Identity   string
	Idem       string
}

func newTestAPI(t *testing.T, now func() time.Time) *testAPI {
	t.Helper()
	s, err := newServer(filepath.Join(t.TempDir(), "signaling.db"), serverOptions{
		AllowedOrigins: []string{"http://localhost:5200"},
		Now:            now,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return &testAPI{t: t, server: s, handler: s.Handler()}
}

func (a *testAPI) request(method, path string, body any, headers map[string]string) (int, []byte) {
	a.t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			a.t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

func decodeMap(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	return out
}

func (a *testAPI) createRoom(max int) roomFixture {
	a.t.Helper()
	status, body := a.request(http.MethodPost, "/v2/rooms", map[string]any{"maxMembers": max}, nil)
	if status != http.StatusCreated {
		a.t.Fatalf("create room: status=%d body=%s", status, body)
	}
	out := decodeMap(a.t, body)
	return roomFixture{
		RoomID:          out["roomId"].(string),
		OwnerCapability: out["ownerCapability"].(string),
		OwnerMemberID:   out["ownerMemberId"].(string),
		OwnerDeviceID:   out["ownerDeviceId"].(string),
		OwnerCredential: out["ownerMemberCredential"].(string),
	}
}

func (a *testAPI) issueInvite(room roomFixture, expires int64) inviteFixture {
	a.t.Helper()
	status, body := a.request(http.MethodPost, "/v2/rooms/"+room.RoomID+"/invites",
		map[string]any{"expiresInSeconds": expires},
		map[string]string{"X-Owner-Capability": room.OwnerCapability})
	if status != http.StatusCreated {
		a.t.Fatalf("issue invite: status=%d body=%s", status, body)
	}
	out := decodeMap(a.t, body)
	return inviteFixture{ID: out["inviteId"].(string), Secret: out["inviteSecret"].(string)}
}

func newMemberFixture(n int) memberFixture {
	return memberFixture{
		DeviceID:   fmt.Sprintf("device_%08d", n),
		Credential: fmt.Sprintf("member-credential-%032d", n),
		Identity:   fmt.Sprintf("device-identity-%032d", n),
		Idem:       fmt.Sprintf("idempotency-%016d", n),
	}
}

func (a *testAPI) redeem(invite inviteFixture, m memberFixture) (int, []byte) {
	a.t.Helper()
	return a.request(http.MethodPost, "/v2/invites/"+invite.ID+"/redeem", map[string]any{
		"inviteSecret": invite.Secret, "deviceId": m.DeviceID,
		"deviceIdentity": m.Identity, "memberCredential": m.Credential,
		"idempotencyKey": m.Idem,
	}, nil)
}

func (a *testAPI) admit(room roomFixture, n int) memberFixture {
	a.t.Helper()
	invite := a.issueInvite(room, 3600)
	m := newMemberFixture(n)
	status, body := a.redeem(invite, m)
	if status != http.StatusCreated {
		a.t.Fatalf("redeem invite: status=%d body=%s", status, body)
	}
	m.ID = decodeMap(a.t, body)["memberId"].(string)
	return m
}

func bearer(credential string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + credential}
}

func TestCreateRoomAuthenticationAndCORS(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(3)
	path := "/v2/rooms/" + room.RoomID

	status, body := a.request(http.MethodGet, path, nil, nil)
	if status != http.StatusUnauthorized || !strings.Contains(string(body), `"authentication_required"`) {
		t.Fatalf("missing auth: status=%d body=%s", status, body)
	}
	status, _ = a.request(http.MethodGet, path, nil, bearer("wrong-credential-xxxxxxxx"))
	if status != http.StatusUnauthorized {
		t.Fatalf("wrong auth status=%d", status)
	}
	status, body = a.request(http.MethodGet, path, nil, bearer(room.OwnerCredential))
	if status != http.StatusOK {
		t.Fatalf("owner member auth: status=%d body=%s", status, body)
	}
	info := decodeMap(t, body)
	if info["memberCount"].(float64) != 1 || info["maxMembers"].(float64) != 3 {
		t.Fatalf("unexpected room info: %v", info)
	}

	status, _ = a.request(http.MethodGet, path, nil, map[string]string{
		"Authorization": "Bearer " + room.OwnerCredential,
		"Origin":        "https://evil.example",
	})
	if status != http.StatusForbidden {
		t.Fatalf("disallowed origin status=%d", status)
	}
}

func TestInviteIssuanceWrongExpiredAndRevoked(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	a := newTestAPI(t, func() time.Time { return now })
	room := a.createRoom(5)

	status, _ := a.request(http.MethodPost, "/v2/rooms/"+room.RoomID+"/invites",
		map[string]any{"expiresInSeconds": 60}, map[string]string{"X-Owner-Capability": "wrong-capability-xxxxxxxx"})
	if status != http.StatusForbidden {
		t.Fatalf("wrong owner capability status=%d", status)
	}
	invite := a.issueInvite(room, 1)
	wrong := invite
	wrong.Secret = "wrong-invite-secret-xxxxxxxx"
	status, _ = a.redeem(wrong, newMemberFixture(1))
	if status != http.StatusForbidden {
		t.Fatalf("wrong invite status=%d", status)
	}
	now = now.Add(2 * time.Second)
	status, body := a.redeem(invite, newMemberFixture(1))
	if status != http.StatusGone || !strings.Contains(string(body), `"invite_expired"`) {
		t.Fatalf("expired invite: status=%d body=%s", status, body)
	}

	active := a.issueInvite(room, 60)
	status, _ = a.request(http.MethodDelete,
		"/v2/rooms/"+room.RoomID+"/invites/"+active.ID, nil,
		map[string]string{"X-Owner-Capability": room.OwnerCapability})
	if status != http.StatusNoContent {
		t.Fatalf("revoke status=%d", status)
	}
	status, body = a.redeem(active, newMemberFixture(2))
	if status != http.StatusGone || !strings.Contains(string(body), `"invite_revoked"`) {
		t.Fatalf("revoked invite: status=%d body=%s", status, body)
	}
}

func TestSuccessfulRedemptionReplayAndReconnect(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	invite := a.issueInvite(room, 3600)
	m := newMemberFixture(1)

	status, body := a.redeem(invite, m)
	if status != http.StatusCreated {
		t.Fatalf("first redemption: status=%d body=%s", status, body)
	}
	first := decodeMap(t, body)
	m.ID = first["memberId"].(string)
	if first["reconnected"].(bool) {
		t.Fatal("first redemption marked reconnected")
	}
	status, body = a.redeem(invite, m)
	if status != http.StatusOK {
		t.Fatalf("replay: status=%d body=%s", status, body)
	}
	replay := decodeMap(t, body)
	if replay["memberId"] != m.ID || !replay["reconnected"].(bool) {
		t.Fatalf("replay was not idempotent: %v", replay)
	}

	status, body = a.request(http.MethodGet, "/v2/rooms/"+room.RoomID, nil, bearer(m.Credential))
	if status != http.StatusOK || decodeMap(t, body)["memberCount"].(float64) != 2 {
		t.Fatalf("member reconnect: status=%d body=%s", status, body)
	}
	extra := a.issueInvite(room, 3600)
	status, body = a.redeem(extra, newMemberFixture(2))
	if status != http.StatusConflict || !strings.Contains(string(body), `"room_full"`) {
		t.Fatalf("capacity: status=%d body=%s", status, body)
	}
}

func TestConcurrentRedemptionAtLastSeat(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	invites := []inviteFixture{a.issueInvite(room, 3600), a.issueInvite(room, 3600)}
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for i := range invites {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			status, _ := a.redeem(invites[i], newMemberFixture(i+1))
			statuses <- status
		}(i)
	}
	wg.Wait()
	close(statuses)
	got := make([]int, 0, 2)
	for status := range statuses {
		got = append(got, status)
	}
	sort.Ints(got)
	want := []int{http.StatusCreated, http.StatusConflict}
	sort.Ints(want)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("statuses=%v want=%v", got, want)
	}
	status, body := a.request(http.MethodGet, "/v2/rooms/"+room.RoomID, nil, bearer(room.OwnerCredential))
	if status != http.StatusOK || decodeMap(t, body)["memberCount"].(float64) != 2 {
		t.Fatalf("member count after race: status=%d body=%s", status, body)
	}
}

func TestSessionIsolationAuthenticationAndCandidatesDoNotConsumeCapacity(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(3)
	alice := a.admit(room, 1)
	bob := a.admit(room, 2)
	sessionA := "session_A_123456"
	sessionB := "session_B_123456"
	pathA := "/v2/rooms/" + room.RoomID + "/sessions/" + sessionA + "/signals"

	status, _ := a.request(http.MethodPost, pathA, map[string]any{
		"kind": "offer", "fromDeviceId": room.OwnerDeviceID,
		"toDeviceId": alice.DeviceID, "envelope": "opaque-offer",
	}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated signal status=%d", status)
	}
	status, _ = a.request(http.MethodPost, pathA, map[string]any{
		"kind": "offer", "fromDeviceId": bob.DeviceID,
		"toDeviceId": alice.DeviceID, "envelope": "opaque-offer",
	}, bearer(room.OwnerCredential))
	if status != http.StatusForbidden {
		t.Fatalf("spoofed sender status=%d", status)
	}
	status, body := a.request(http.MethodPost, pathA, map[string]any{
		"kind": "offer", "fromDeviceId": room.OwnerDeviceID,
		"toDeviceId": alice.DeviceID, "envelope": "opaque-offer",
	}, bearer(room.OwnerCredential))
	if status != http.StatusCreated {
		t.Fatalf("offer: status=%d body=%s", status, body)
	}

	status, body = a.request(http.MethodGet, pathA, nil, bearer(bob.Credential))
	if status != http.StatusOK || len(decodeMap(t, body)["signals"].([]any)) != 0 {
		t.Fatalf("recipient isolation: status=%d body=%s", status, body)
	}
	pathB := "/v2/rooms/" + room.RoomID + "/sessions/" + sessionB + "/signals"
	status, body = a.request(http.MethodGet, pathB, nil, bearer(alice.Credential))
	if status != http.StatusOK || len(decodeMap(t, body)["signals"].([]any)) != 0 {
		t.Fatalf("session isolation: status=%d body=%s", status, body)
	}
	status, body = a.request(http.MethodGet, pathA, nil, bearer(alice.Credential))
	if status != http.StatusOK || len(decodeMap(t, body)["signals"].([]any)) != 1 {
		t.Fatalf("offer delivery: status=%d body=%s", status, body)
	}

	for i := 0; i < 8; i++ {
		status, body = a.request(http.MethodPost, pathA, map[string]any{
			"kind": "candidate", "fromDeviceId": room.OwnerDeviceID,
			"toDeviceId": alice.DeviceID, "envelope": fmt.Sprintf("opaque-candidate-%d", i),
		}, bearer(room.OwnerCredential))
		if status != http.StatusCreated {
			t.Fatalf("candidate %d: status=%d body=%s", i, status, body)
		}
	}
	status, body = a.request(http.MethodGet, "/v2/rooms/"+room.RoomID, nil, bearer(room.OwnerCredential))
	if status != http.StatusOK || decodeMap(t, body)["memberCount"].(float64) != 3 {
		t.Fatalf("signals changed capacity: status=%d body=%s", status, body)
	}
}

func TestNewInviteeReadsRoomCheckpointWhileUploaderOffline(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	otherRoom := a.createRoom(2)
	checkpointPath := "/v2/rooms/" + room.RoomID + "/mailbox/checkpoint"
	// This represents ciphertext the client can decrypt. The server must preserve
	// it exactly and must not depend on the uploader remaining connected.
	opaqueEnvelope := `encrypted:v1:nonce=abc123:ciphertext=AAECAwQFBgcICQ==`

	status, body := a.request(http.MethodPut, checkpointPath,
		map[string]any{"envelope": opaqueEnvelope}, bearer(room.OwnerCredential))
	if status != http.StatusNoContent {
		t.Fatalf("creator put room checkpoint: status=%d body=%s", status, body)
	}
	status, _ = a.request(http.MethodGet, checkpointPath, nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated room checkpoint read status=%d", status)
	}
	status, _ = a.request(http.MethodGet, checkpointPath, nil, bearer(otherRoom.OwnerCredential))
	if status != http.StatusUnauthorized {
		t.Fatalf("nonmember room checkpoint read status=%d", status)
	}

	// No further creator request is made: the newly admitted member bootstraps
	// solely from the durable room checkpoint.
	newcomer := a.admit(room, 1)
	status, body = a.request(http.MethodGet, checkpointPath, nil, bearer(newcomer.Credential))
	if status != http.StatusOK {
		t.Fatalf("new invitee get room checkpoint: status=%d body=%s", status, body)
	}
	checkpoint := decodeMap(t, body)
	if checkpoint["envelope"] != opaqueEnvelope {
		t.Fatalf("opaque checkpoint changed: got=%q want=%q", checkpoint["envelope"], opaqueEnvelope)
	}
	if checkpoint["uploaderDeviceId"] != room.OwnerDeviceID {
		t.Fatalf("uploader device=%v want=%s", checkpoint["uploaderDeviceId"], room.OwnerDeviceID)
	}
}

func TestDeviceListingIsAuthenticatedRoomScopedAndSecretFree(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(3)
	alice := a.admit(room, 1)
	bob := a.admit(room, 2)
	otherRoom := a.createRoom(2)
	path := "/v2/rooms/" + room.RoomID + "/devices"

	status, _ := a.request(http.MethodGet, path, nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated devices status=%d", status)
	}
	status, _ = a.request(http.MethodGet, path, nil, bearer(otherRoom.OwnerCredential))
	if status != http.StatusUnauthorized {
		t.Fatalf("cross-room devices status=%d", status)
	}
	status, body := a.request(http.MethodGet, path, nil, bearer(alice.Credential))
	if status != http.StatusOK {
		t.Fatalf("list devices: status=%d body=%s", status, body)
	}
	response := decodeMap(t, body)
	devices := response["devices"].([]any)
	if len(devices) != 3 {
		t.Fatalf("devices=%v", devices)
	}
	gotIDs := make([]string, 0, len(devices))
	for _, raw := range devices {
		device := raw.(map[string]any)
		gotIDs = append(gotIDs, device["deviceId"].(string))
	}
	sort.Strings(gotIDs)
	wantIDs := []string{room.OwnerDeviceID, alice.DeviceID, bob.DeviceID}
	sort.Strings(wantIDs)
	if fmt.Sprint(gotIDs) != fmt.Sprint(wantIDs) {
		t.Fatalf("device IDs=%v want=%v", gotIDs, wantIDs)
	}
	lowerBody := strings.ToLower(string(body))
	for _, forbidden := range []string{"credential", "secret", "hash", room.OwnerCredential, room.OwnerCapability, alice.Credential, bob.Credential} {
		if strings.Contains(lowerBody, strings.ToLower(forbidden)) {
			t.Fatalf("device listing exposed forbidden value %q: %s", forbidden, body)
		}
	}
}

func TestOpaqueMailboxCheckpointAndOperations(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	alice := a.admit(room, 1)
	checkpointPath := "/v2/rooms/" + room.RoomID + "/mailbox/" + alice.DeviceID + "/checkpoint"
	checkpoint := `v1:base64:AAECAwQFBgcICQ==`

	status, body := a.request(http.MethodPut, checkpointPath,
		map[string]any{"envelope": checkpoint}, bearer(alice.Credential))
	if status != http.StatusNoContent {
		t.Fatalf("put checkpoint: status=%d body=%s", status, body)
	}
	status, body = a.request(http.MethodGet, checkpointPath, nil, bearer(alice.Credential))
	if status != http.StatusOK || decodeMap(t, body)["envelope"] != checkpoint {
		t.Fatalf("get checkpoint: status=%d body=%s", status, body)
	}
	status, _ = a.request(http.MethodGet, checkpointPath, nil, bearer(room.OwnerCredential))
	if status != http.StatusForbidden {
		t.Fatalf("cross-device checkpoint read status=%d", status)
	}

	operationsPath := "/v2/rooms/" + room.RoomID + "/mailbox/" + alice.DeviceID + "/operations"
	operation := `ciphertext-without-server-interpretation`
	status, body = a.request(http.MethodPost, operationsPath,
		map[string]any{"envelope": operation}, bearer(room.OwnerCredential))
	if status != http.StatusCreated {
		t.Fatalf("post operation: status=%d body=%s", status, body)
	}
	status, body = a.request(http.MethodGet, operationsPath, nil, bearer(alice.Credential))
	if status != http.StatusOK {
		t.Fatalf("get operations: status=%d body=%s", status, body)
	}
	operations := decodeMap(t, body)["operations"].([]any)
	if len(operations) != 1 || operations[0].(map[string]any)["envelope"] != operation {
		t.Fatalf("opaque operation changed: %v", operations)
	}
}

func TestPayloadLimits(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	alice := a.admit(room, 1)

	signalPath := "/v2/rooms/" + room.RoomID + "/sessions/session_payload_1/signals"
	status, body := a.request(http.MethodPost, signalPath, map[string]any{
		"kind": "candidate", "fromDeviceId": room.OwnerDeviceID, "toDeviceId": alice.DeviceID,
		"envelope": strings.Repeat("x", maxSignalPayload+1),
	}, bearer(room.OwnerCredential))
	if status != http.StatusRequestEntityTooLarge || !strings.Contains(string(body), `"payload_too_large"`) {
		t.Fatalf("signal limit: status=%d body=%s", status, body)
	}

	checkpointPath := "/v2/rooms/" + room.RoomID + "/mailbox/checkpoint"
	status, body = a.request(http.MethodPut, checkpointPath,
		map[string]any{"envelope": strings.Repeat("x", maxCheckpointPayload+1)}, bearer(room.OwnerCredential))
	if status != http.StatusRequestEntityTooLarge || !strings.Contains(string(body), `"payload_too_large"`) {
		t.Fatalf("checkpoint limit: status=%d body=%s", status, body)
	}
}

func TestSecretsAreHashedAndNotExposedAfterProvisioning(t *testing.T) {
	a := newTestAPI(t, time.Now)
	room := a.createRoom(2)
	invite := a.issueInvite(room, 3600)
	m := newMemberFixture(1)
	status, body := a.redeem(invite, m)
	if status != http.StatusCreated {
		t.Fatalf("redeem: status=%d body=%s", status, body)
	}
	for _, secret := range []string{invite.Secret, m.Credential, m.Identity, m.Idem, room.OwnerCapability, room.OwnerCredential} {
		if bytes.Contains(body, []byte(secret)) {
			t.Fatalf("redemption response exposed secret %q", secret)
		}
	}
	status, roomBody := a.request(http.MethodGet, "/v2/rooms/"+room.RoomID, nil, bearer(m.Credential))
	if status != http.StatusOK {
		t.Fatalf("room info status=%d body=%s", status, roomBody)
	}
	for _, secret := range []string{invite.Secret, m.Credential, m.Identity, m.Idem, room.OwnerCapability, room.OwnerCredential} {
		if bytes.Contains(roomBody, []byte(secret)) {
			t.Fatalf("room response exposed secret %q", secret)
		}
		var count int
		for _, tableColumn := range [][2]string{
			{"v2_rooms", "owner_capability_hash"},
			{"v2_members", "credential_hash"},
			{"v2_members", "device_identity_hash"},
			{"v2_members", "idempotency_hash"},
			{"v2_invites", "secret_hash"},
		} {
			query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", tableColumn[0], tableColumn[1])
			if err := a.server.db.QueryRow(query, secret).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatalf("plaintext secret stored in %s.%s", tableColumn[0], tableColumn[1])
			}
		}
	}
}

func TestPrototypeAdminAndDevEndpointsRemoved(t *testing.T) {
	a := newTestAPI(t, time.Now)
	for _, path := range []string{"/admin", "/dev/reset-room/x", "/dev/set-capacity/x", "/register", "/get", "/room/x/offer"} {
		status, body := a.request(http.MethodGet, path, nil, nil)
		if status != http.StatusNotFound || !strings.Contains(string(body), `"not_found"`) {
			t.Fatalf("%s: status=%d body=%s", path, status, body)
		}
	}
}
