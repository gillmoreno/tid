package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"signaling/internal/creationpermit"

	_ "modernc.org/sqlite"
)

const (
	maxJSONBody          = 300 << 10
	maxSignalPayload     = 64 << 10
	maxCheckpointPayload = 256 << 10
	maxOperationPayload  = 64 << 10
	maxMailboxOperations = 1000
	maxSignalsPerSession = 512
)

type serverOptions struct {
	AllowedOrigins   []string
	CreatorVerifyKey string
	Now              func() time.Time
}

type server struct {
	db               *sql.DB
	now              func() time.Time
	allowedOrigins   map[string]struct{}
	creatorVerifyKey ed25519.PublicKey
}

type member struct {
	ID       string
	RoomID   string
	DeviceID string
	IsOwner  bool
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newServer(dbPath string, opts serverOptions) (*server, error) {
	if dbPath == "" {
		return nil, errors.New("database path is required")
	}
	creatorVerifyKey, err := creationpermit.ParsePublicKey(opts.CreatorVerifyKey)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open signaling database: %w", err)
	}
	// Serializing writers makes invite redemption/capacity admission deterministic.
	// This process is intentionally a small single-node SQLite service.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure signaling database: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	origins := make(map[string]struct{}, len(opts.AllowedOrigins))
	for _, origin := range opts.AllowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return &server{
		db:               db,
		now:              now,
		allowedOrigins:   origins,
		creatorVerifyKey: creatorVerifyKey,
	}, nil
}

func migrate(db *sql.DB) error {
	// V2 uses new table names so an existing prototype database is left intact.
	// No raw credentials, invite secrets, or content decryption keys are stored.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS signaling_schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS v2_rooms (
			room_id TEXT PRIMARY KEY,
			owner_capability_hash TEXT NOT NULL,
			max_members INTEGER NOT NULL CHECK(max_members BETWEEN 2 AND 50),
			created_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS v2_members (
			member_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			device_id TEXT NOT NULL,
			device_identity_hash TEXT NOT NULL,
			credential_hash TEXT NOT NULL,
			idempotency_hash TEXT NOT NULL,
			is_owner INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			last_seen_at INTEGER NOT NULL,
			UNIQUE(room_id, device_id),
			UNIQUE(room_id, device_identity_hash),
			UNIQUE(room_id, credential_hash),
			UNIQUE(room_id, idempotency_hash)
		);
		CREATE TABLE IF NOT EXISTS v2_invites (
			invite_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			secret_hash TEXT NOT NULL,
			expires_at INTEGER NOT NULL,
			revoked_at INTEGER,
			redeemed_member_id TEXT REFERENCES v2_members(member_id),
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_v2_invites_room ON v2_invites(room_id);
		CREATE TABLE IF NOT EXISTS v2_signals (
			signal_id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id TEXT NOT NULL REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			session_id TEXT NOT NULL,
			kind TEXT NOT NULL CHECK(kind IN ('offer', 'answer', 'candidate')),
			from_device_id TEXT NOT NULL,
			to_device_id TEXT NOT NULL,
			envelope TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_v2_signals_delivery
			ON v2_signals(room_id, session_id, to_device_id, signal_id);
		CREATE TABLE IF NOT EXISTS v2_checkpoints (
			room_id TEXT NOT NULL REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			device_id TEXT NOT NULL,
			envelope TEXT NOT NULL,
			updated_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			PRIMARY KEY(room_id, device_id)
		);
		CREATE TABLE IF NOT EXISTS v2_room_checkpoints (
			room_id TEXT PRIMARY KEY REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			uploader_device_id TEXT NOT NULL,
			envelope TEXT NOT NULL,
			updated_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS v2_operations (
			operation_id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id TEXT NOT NULL REFERENCES v2_rooms(room_id) ON DELETE CASCADE,
			from_device_id TEXT NOT NULL,
			to_device_id TEXT NOT NULL,
			envelope TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_v2_operations_delivery
			ON v2_operations(room_id, to_device_id, operation_id);
		CREATE TABLE IF NOT EXISTS v2_creator_permit_uses (
			permit_id_hash TEXT PRIMARY KEY,
			consumed_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			room_id TEXT NOT NULL UNIQUE
		);
		INSERT OR IGNORE INTO signaling_schema_migrations(version, applied_at)
			VALUES (2, unixepoch() * 1000);
		INSERT OR IGNORE INTO signaling_schema_migrations(version, applied_at)
			VALUES (3, unixepoch() * 1000);
		INSERT OR IGNORE INTO signaling_schema_migrations(version, applied_at)
			VALUES (4, unixepoch() * 1000);
	`)
	if err != nil {
		return fmt.Errorf("migrate signaling database: %w", err)
	}
	if err := migrateCreatorPermitUses(db); err != nil {
		return fmt.Errorf("migrate creator permit uses: %w", err)
	}
	return nil
}

func migrateCreatorPermitUses(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA foreign_key_list(v2_creator_permit_uses)`)
	if err != nil {
		return err
	}
	hasRoomForeignKey := rows.Next()
	if err := rows.Close(); err != nil {
		return err
	}
	if !hasRoomForeignKey {
		_, err = db.Exec(`INSERT OR IGNORE INTO signaling_schema_migrations(version, applied_at)
			VALUES (5, unixepoch() * 1000)`)
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`
		DROP TABLE IF EXISTS v2_creator_permit_uses_v5;
		CREATE TABLE v2_creator_permit_uses_v5 (
			permit_id_hash TEXT PRIMARY KEY,
			consumed_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			room_id TEXT NOT NULL UNIQUE
		);
		INSERT INTO v2_creator_permit_uses_v5(permit_id_hash, consumed_at, expires_at, room_id)
			SELECT permit_id_hash, consumed_at, expires_at, room_id FROM v2_creator_permit_uses;
		DROP TABLE v2_creator_permit_uses;
		ALTER TABLE v2_creator_permit_uses_v5 RENAME TO v2_creator_permit_uses;
		INSERT OR IGNORE INTO signaling_schema_migrations(version, applied_at)
			VALUES (5, unixepoch() * 1000);
	`); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *server) Close() error {
	return s.db.Close()
}

func (s *server) Handler() http.Handler {
	return s.withRequestLog(s.withCORS(http.HandlerFunc(s.route)))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (s *server) withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID, err := randomToken(12)
		if err != nil {
			requestID = "unavailable"
		}
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("request_id=%s method=%s route=%s status=%d duration_ms=%d",
			requestID, r.Method, routeTemplate(r.URL.Path), status, time.Since(started).Milliseconds())
	})
}

func routeTemplate(path string) string {
	parts := splitPath(path)
	switch {
	case len(parts) == 1 && parts[0] == "healthz":
		return "/healthz"
	case len(parts) == 2 && parts[0] == "v2" && parts[1] == "rooms":
		return "/v2/rooms"
	case len(parts) == 4 && parts[0] == "v2" && parts[1] == "invites" && parts[3] == "redeem":
		return "/v2/invites/:inviteId/redeem"
	case len(parts) == 3 && parts[0] == "v2" && parts[1] == "rooms":
		return "/v2/rooms/:roomId"
	case len(parts) == 4 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "devices":
		return "/v2/rooms/:roomId/devices"
	case len(parts) == 4 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "invites":
		return "/v2/rooms/:roomId/invites"
	case len(parts) == 5 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "invites":
		return "/v2/rooms/:roomId/invites/:inviteId"
	case len(parts) == 6 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "sessions" && parts[5] == "signals":
		return "/v2/rooms/:roomId/sessions/:sessionId/signals"
	case len(parts) == 5 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "mailbox" && parts[4] == "checkpoint":
		return "/v2/rooms/:roomId/mailbox/checkpoint"
	case len(parts) == 6 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "mailbox" && parts[5] == "checkpoint":
		return "/v2/rooms/:roomId/mailbox/:deviceId/checkpoint"
	case len(parts) == 6 && parts[0] == "v2" && parts[1] == "rooms" && parts[3] == "mailbox" && parts[5] == "operations":
		return "/v2/rooms/:roomId/mailbox/:deviceId/operations"
	default:
		return "unmatched"
	}
}

func (s *server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := s.allowedOrigins[origin]; !ok {
				writeError(w, http.StatusForbidden, "origin_not_allowed", "origin is not allowed")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Owner-Capability, X-Room-Creator-Permit")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) route(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
		return
	}
	if len(parts) == 1 && parts[0] == "healthz" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if len(parts) == 2 && parts[0] == "v2" && parts[1] == "rooms" && r.Method == http.MethodPost {
		s.createRoom(w, r)
		return
	}
	if len(parts) == 4 && parts[0] == "v2" && parts[1] == "invites" && parts[3] == "redeem" && r.Method == http.MethodPost {
		s.redeemInvite(w, r, parts[2])
		return
	}
	if len(parts) >= 3 && parts[0] == "v2" && parts[1] == "rooms" {
		roomID := parts[2]
		switch {
		case len(parts) == 3 && r.Method == http.MethodGet:
			s.getRoom(w, r, roomID)
		case len(parts) == 4 && parts[3] == "devices" && r.Method == http.MethodGet:
			s.getDevices(w, r, roomID)
		case len(parts) == 4 && parts[3] == "invites" && r.Method == http.MethodPost:
			s.issueInvite(w, r, roomID)
		case len(parts) == 5 && parts[3] == "invites" && r.Method == http.MethodDelete:
			s.revokeInvite(w, r, roomID, parts[4])
		case len(parts) == 6 && parts[3] == "sessions" && parts[5] == "signals" && r.Method == http.MethodPost:
			s.postSignal(w, r, roomID, parts[4])
		case len(parts) == 6 && parts[3] == "sessions" && parts[5] == "signals" && r.Method == http.MethodGet:
			s.getSignals(w, r, roomID, parts[4])
		case len(parts) == 5 && parts[3] == "mailbox" && parts[4] == "checkpoint" && r.Method == http.MethodPut:
			s.putRoomCheckpoint(w, r, roomID)
		case len(parts) == 5 && parts[3] == "mailbox" && parts[4] == "checkpoint" && r.Method == http.MethodGet:
			s.getRoomCheckpoint(w, r, roomID)
		case len(parts) == 6 && parts[3] == "mailbox" && parts[5] == "checkpoint" && r.Method == http.MethodPut:
			s.putCheckpoint(w, r, roomID, parts[4])
		case len(parts) == 6 && parts[3] == "mailbox" && parts[5] == "checkpoint" && r.Method == http.MethodGet:
			s.getCheckpoint(w, r, roomID, parts[4])
		case len(parts) == 6 && parts[3] == "mailbox" && parts[5] == "operations" && r.Method == http.MethodPost:
			s.postOperation(w, r, roomID, parts[4])
		case len(parts) == 6 && parts[3] == "mailbox" && parts[5] == "operations" && r.Method == http.MethodGet:
			s.getOperations(w, r, roomID, parts[4])
		default:
			writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
		}
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
}

func (s *server) createRoom(w http.ResponseWriter, r *http.Request) {
	providedPermit := r.Header.Get("X-Room-Creator-Permit")
	if providedPermit == "" {
		writeError(w, http.StatusUnauthorized, "creator_permit_required", "room creator permit is required")
		return
	}
	var input struct {
		MaxMembers int `json:"maxMembers"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if input.MaxMembers == 0 {
		input.MaxMembers = 2
	}
	if input.MaxMembers < 2 || input.MaxMembers > 50 {
		writeError(w, http.StatusBadRequest, "invalid_max_members", "maxMembers must be between 2 and 50")
		return
	}
	permit, err := creationpermit.Verify(providedPermit, s.creatorVerifyKey, s.now())
	if err != nil {
		writeCreationPermitError(w, err)
		return
	}
	if permit.MaxMembers != input.MaxMembers {
		writeError(w, http.StatusForbidden, "creator_permit_capacity_mismatch", "room capacity does not match the creator permit")
		return
	}
	roomID, ownerCapability, memberID, deviceID, memberCredential, err := randomValues()
	if err != nil {
		s.internalError(w, err)
		return
	}
	tx, err := s.db.Begin()
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer tx.Rollback()
	transactionNow := s.now()
	if permit.ExpiresAt <= transactionNow.Unix() {
		writeError(w, http.StatusForbidden, "creator_permit_expired", "room creator permit has expired")
		return
	}
	now := transactionNow.UnixMilli()
	if _, err = tx.Exec(`INSERT INTO v2_rooms(room_id, owner_capability_hash, max_members, created_at) VALUES(?, ?, ?, ?)`,
		roomID, hashSecret(ownerCapability), input.MaxMembers, now); err != nil {
		s.internalError(w, err)
		return
	}
	result, err := tx.Exec(`
		INSERT INTO v2_creator_permit_uses(permit_id_hash, consumed_at, expires_at, room_id)
		VALUES(?, ?, ?, ?) ON CONFLICT(permit_id_hash) DO NOTHING`,
		hashSecret(permit.ID), now, permit.ExpiresAt*1000, roomID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	consumed, err := result.RowsAffected()
	if err != nil {
		s.internalError(w, err)
		return
	}
	if consumed != 1 {
		writeError(w, http.StatusConflict, "creator_permit_used", "room creator permit has already been used")
		return
	}
	ownerIdentity := "owner:" + deviceID
	if _, err = tx.Exec(`
		INSERT INTO v2_members(member_id, room_id, device_id, device_identity_hash, credential_hash, idempotency_hash, is_owner, created_at, last_seen_at)
		VALUES(?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		memberID, roomID, deviceID, hashSecret(ownerIdentity), hashSecret(memberCredential), hashSecret("owner:"+memberID), now, now); err != nil {
		s.internalError(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"roomId":                roomID,
		"maxMembers":            input.MaxMembers,
		"ownerCapability":       ownerCapability,
		"ownerMemberId":         memberID,
		"ownerDeviceId":         deviceID,
		"ownerMemberCredential": memberCredential,
	})
}

func writeCreationPermitError(w http.ResponseWriter, err error) {
	var validationErr *creationpermit.ValidationError
	if !errors.As(err, &validationErr) {
		writeError(w, http.StatusForbidden, "invalid_creator_permit", "room creator permit is invalid")
		return
	}
	switch validationErr.Kind {
	case creationpermit.Expired:
		writeError(w, http.StatusForbidden, "creator_permit_expired", "room creator permit has expired")
	case creationpermit.NotYetValid:
		writeError(w, http.StatusForbidden, "creator_permit_not_yet_valid", "room creator permit is not valid yet")
	default:
		writeError(w, http.StatusForbidden, "invalid_creator_permit", "room creator permit is invalid")
	}
}

func randomValues() (string, string, string, string, string, error) {
	values := make([]string, 5)
	sizes := []int{18, 32, 18, 18, 32}
	for i := range values {
		value, err := randomToken(sizes[i])
		if err != nil {
			return "", "", "", "", "", err
		}
		values[i] = value
	}
	return values[0], values[1], values[2], values[3], values[4], nil
}

func (s *server) getRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	if _, ok := s.authenticateMember(w, r, roomID); !ok {
		return
	}
	var maxMembers, memberCount int
	var createdAt int64
	err := s.db.QueryRow(`
		SELECT r.max_members, r.created_at, COUNT(m.member_id)
		FROM v2_rooms r LEFT JOIN v2_members m ON m.room_id = r.room_id
		WHERE r.room_id = ? GROUP BY r.room_id`, roomID).Scan(&maxMembers, &createdAt, &memberCount)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "room_not_found", "room not found")
		return
	}
	if err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"roomId": roomID, "maxMembers": maxMembers, "memberCount": memberCount,
		"createdAt": time.UnixMilli(createdAt).UTC(),
	})
}

func (s *server) getDevices(w http.ResponseWriter, r *http.Request, roomID string) {
	authenticated, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	rows, err := s.db.Query(`
		SELECT device_id, is_owner FROM v2_members
		WHERE room_id = ? ORDER BY created_at, device_id`, roomID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer rows.Close()
	devices := make([]map[string]any, 0)
	for rows.Next() {
		var deviceID string
		var isOwner bool
		if err := rows.Scan(&deviceID, &isOwner); err != nil {
			s.internalError(w, err)
			return
		}
		devices = append(devices, map[string]any{
			"deviceId": deviceID,
			"isOwner":  isOwner,
			"isSelf":   deviceID == authenticated.DeviceID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"roomId": roomID, "devices": devices})
}

func (s *server) issueInvite(w http.ResponseWriter, r *http.Request, roomID string) {
	if !s.authenticateOwner(w, r, roomID) {
		return
	}
	var input struct {
		ExpiresInSeconds int64 `json:"expiresInSeconds"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if input.ExpiresInSeconds == 0 {
		input.ExpiresInSeconds = int64((24 * time.Hour) / time.Second)
	}
	if input.ExpiresInSeconds < 1 || input.ExpiresInSeconds > int64((7*24*time.Hour)/time.Second) {
		writeError(w, http.StatusBadRequest, "invalid_expiry", "expiresInSeconds must be between 1 and 604800")
		return
	}
	inviteID, err := randomToken(18)
	if err != nil {
		s.internalError(w, err)
		return
	}
	secret, err := randomToken(32)
	if err != nil {
		s.internalError(w, err)
		return
	}
	now := s.now()
	expires := now.Add(time.Duration(input.ExpiresInSeconds) * time.Second)
	if _, err := s.db.Exec(`
		INSERT INTO v2_invites(invite_id, room_id, secret_hash, expires_at, created_at)
		VALUES(?, ?, ?, ?, ?)`, inviteID, roomID, hashSecret(secret), expires.UnixMilli(), now.UnixMilli()); err != nil {
		s.internalError(w, err)
		return
	}
	// The secret is returned exactly once. Listing/revocation never exposes it.
	writeJSON(w, http.StatusCreated, map[string]any{
		"inviteId": inviteID, "inviteSecret": secret, "expiresAt": expires.UTC(),
	})
}

func (s *server) revokeInvite(w http.ResponseWriter, r *http.Request, roomID, inviteID string) {
	if !s.authenticateOwner(w, r, roomID) {
		return
	}
	result, err := s.db.Exec(`
		UPDATE v2_invites SET revoked_at = COALESCE(revoked_at, ?)
		WHERE invite_id = ? AND room_id = ?`, s.now().UnixMilli(), inviteID, roomID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "invite_not_found", "invite not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) redeemInvite(w http.ResponseWriter, r *http.Request, inviteID string) {
	var input struct {
		InviteSecret     string `json:"inviteSecret"`
		DeviceID         string `json:"deviceId"`
		DeviceIdentity   string `json:"deviceIdentity"`
		MemberCredential string `json:"memberCredential"`
		IdempotencyKey   string `json:"idempotencyKey"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if len(input.InviteSecret) < 16 || len(input.DeviceIdentity) < 16 || len(input.MemberCredential) < 16 || len(input.IdempotencyKey) < 8 {
		writeError(w, http.StatusBadRequest, "invalid_credentials", "inviteSecret, deviceIdentity, memberCredential, and idempotencyKey are required")
		return
	}
	if !validID(input.DeviceID) {
		writeError(w, http.StatusBadRequest, "invalid_device_id", "deviceId must be 8-128 URL-safe characters")
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer tx.Rollback()
	var roomID, secretHash string
	var expiresAt int64
	var revokedAt sql.NullInt64
	var redeemedMemberID sql.NullString
	err = tx.QueryRow(`
		SELECT room_id, secret_hash, expires_at, revoked_at, redeemed_member_id
		FROM v2_invites WHERE invite_id = ?`, inviteID).
		Scan(&roomID, &secretHash, &expiresAt, &revokedAt, &redeemedMemberID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "invite_not_found", "invite not found")
		return
	}
	if err != nil {
		s.internalError(w, err)
		return
	}
	if !secretMatches(secretHash, input.InviteSecret) {
		writeError(w, http.StatusForbidden, "invalid_invite", "invite secret is invalid")
		return
	}
	if revokedAt.Valid {
		writeError(w, http.StatusGone, "invite_revoked", "invite has been revoked")
		return
	}
	if s.now().UnixMilli() >= expiresAt {
		writeError(w, http.StatusGone, "invite_expired", "invite has expired")
		return
	}

	identityHash := hashSecret(input.DeviceIdentity)
	credentialHash := hashSecret(input.MemberCredential)
	idempotencyHash := hashSecret(input.IdempotencyKey)
	existing, found, err := findMemberForRedemption(tx, roomID, identityHash, idempotencyHash)
	if err != nil {
		s.internalError(w, err)
		return
	}
	if redeemedMemberID.Valid {
		if !found || existing.ID != redeemedMemberID.String || existing.DeviceID != input.DeviceID ||
			!memberCredentialMatches(tx, existing.ID, credentialHash) {
			writeError(w, http.StatusConflict, "invite_already_redeemed", "invite has already been redeemed")
			return
		}
		_, err = tx.Exec(`UPDATE v2_members SET last_seen_at = ? WHERE member_id = ?`, s.now().UnixMilli(), existing.ID)
		if err != nil {
			s.internalError(w, err)
			return
		}
		if err = tx.Commit(); err != nil {
			s.internalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memberResponse(existing, true))
		return
	}

	if found {
		if existing.DeviceID != input.DeviceID || !memberCredentialMatches(tx, existing.ID, credentialHash) {
			writeError(w, http.StatusConflict, "device_identity_conflict", "device identity is already registered")
			return
		}
		if _, err = tx.Exec(`UPDATE v2_invites SET redeemed_member_id = ? WHERE invite_id = ? AND redeemed_member_id IS NULL`, existing.ID, inviteID); err != nil {
			s.internalError(w, err)
			return
		}
		if err = tx.Commit(); err != nil {
			s.internalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memberResponse(existing, true))
		return
	}

	var maxMembers, memberCount int
	if err = tx.QueryRow(`
		SELECT r.max_members, COUNT(m.member_id)
		FROM v2_rooms r LEFT JOIN v2_members m ON m.room_id = r.room_id
		WHERE r.room_id = ? GROUP BY r.room_id`, roomID).Scan(&maxMembers, &memberCount); err != nil {
		s.internalError(w, err)
		return
	}
	if memberCount >= maxMembers {
		writeError(w, http.StatusConflict, "room_full", "room has reached its member capacity")
		return
	}
	memberID, err := randomToken(18)
	if err != nil {
		s.internalError(w, err)
		return
	}
	now := s.now().UnixMilli()
	_, err = tx.Exec(`
		INSERT INTO v2_members(member_id, room_id, device_id, device_identity_hash, credential_hash, idempotency_hash, created_at, last_seen_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		memberID, roomID, input.DeviceID, identityHash, credentialHash, idempotencyHash, now, now)
	if err != nil {
		if isConstraintError(err) {
			writeError(w, http.StatusConflict, "device_identity_conflict", "device or idempotency identity is already registered")
			return
		}
		s.internalError(w, err)
		return
	}
	result, err := tx.Exec(`
		UPDATE v2_invites SET redeemed_member_id = ?
		WHERE invite_id = ? AND redeemed_member_id IS NULL`, memberID, inviteID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		writeError(w, http.StatusConflict, "invite_already_redeemed", "invite has already been redeemed")
		return
	}
	if err = tx.Commit(); err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, memberResponse(member{ID: memberID, RoomID: roomID, DeviceID: input.DeviceID}, false))
}

func findMemberForRedemption(tx *sql.Tx, roomID, identityHash, idempotencyHash string) (member, bool, error) {
	var m member
	err := tx.QueryRow(`
		SELECT member_id, room_id, device_id, is_owner
		FROM v2_members
		WHERE room_id = ? AND (device_identity_hash = ? OR idempotency_hash = ?)
		LIMIT 1`, roomID, identityHash, idempotencyHash).Scan(&m.ID, &m.RoomID, &m.DeviceID, &m.IsOwner)
	if err == sql.ErrNoRows {
		return member{}, false, nil
	}
	return m, err == nil, err
}

func memberCredentialMatches(tx *sql.Tx, memberID, credentialHash string) bool {
	var stored string
	if err := tx.QueryRow(`SELECT credential_hash FROM v2_members WHERE member_id = ?`, memberID).Scan(&stored); err != nil {
		return false
	}
	return constantStringEqual(stored, credentialHash)
}

func memberResponse(m member, reconnected bool) map[string]any {
	return map[string]any{
		"roomId": m.RoomID, "memberId": m.ID, "deviceId": m.DeviceID, "reconnected": reconnected,
	}
}

func (s *server) postSignal(w http.ResponseWriter, r *http.Request, roomID, sessionID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if !validID(sessionID) {
		writeError(w, http.StatusBadRequest, "invalid_session_id", "sessionId must be 8-128 URL-safe characters")
		return
	}
	var input struct {
		Kind             string `json:"kind"`
		FromDeviceID     string `json:"fromDeviceId"`
		ToDeviceID       string `json:"toDeviceId"`
		Envelope         string `json:"envelope"`
		ExpiresInSeconds int64  `json:"expiresInSeconds"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if input.FromDeviceID != m.DeviceID {
		writeError(w, http.StatusForbidden, "sender_mismatch", "fromDeviceId must match the authenticated device")
		return
	}
	if input.Kind != "offer" && input.Kind != "answer" && input.Kind != "candidate" {
		writeError(w, http.StatusBadRequest, "invalid_signal_kind", "kind must be offer, answer, or candidate")
		return
	}
	if len(input.Envelope) == 0 || len(input.Envelope) > maxSignalPayload {
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "signal envelope must be 1-65536 bytes")
		return
	}
	if !s.deviceExists(roomID, input.ToDeviceID) {
		writeError(w, http.StatusNotFound, "recipient_not_found", "recipient device is not a room member")
		return
	}
	ttl, ok := boundedTTL(w, input.ExpiresInSeconds, 10*time.Minute, time.Hour)
	if !ok {
		return
	}
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer tx.Rollback()
	_, _ = tx.Exec(`DELETE FROM v2_signals WHERE expires_at <= ?`, now.UnixMilli())
	var count int
	if err = tx.QueryRow(`
		SELECT COUNT(*) FROM v2_signals WHERE room_id = ? AND session_id = ?`, roomID, sessionID).Scan(&count); err != nil {
		s.internalError(w, err)
		return
	}
	if count >= maxSignalsPerSession {
		writeError(w, http.StatusTooManyRequests, "signal_quota_exceeded", "session signal quota exceeded")
		return
	}
	result, err := tx.Exec(`
		INSERT INTO v2_signals(room_id, session_id, kind, from_device_id, to_device_id, envelope, created_at, expires_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		roomID, sessionID, input.Kind, m.DeviceID, input.ToDeviceID, input.Envelope, now.UnixMilli(), now.Add(ttl).UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	id, _ := result.LastInsertId()
	if err = tx.Commit(); err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"signalId": id})
}

func (s *server) getSignals(w http.ResponseWriter, r *http.Request, roomID, sessionID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if !validID(sessionID) {
		writeError(w, http.StatusBadRequest, "invalid_session_id", "sessionId must be 8-128 URL-safe characters")
		return
	}
	after, err := parseAfter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_after", "after must be a non-negative integer")
		return
	}
	rows, err := s.db.Query(`
		SELECT signal_id, kind, from_device_id, to_device_id, envelope, created_at, expires_at
		FROM v2_signals
		WHERE room_id = ? AND session_id = ? AND to_device_id = ? AND signal_id > ? AND expires_at > ?
		ORDER BY signal_id LIMIT 200`, roomID, sessionID, m.DeviceID, after, s.now().UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, createdAt, expiresAt int64
		var kind, fromDevice, toDevice, envelope string
		if err := rows.Scan(&id, &kind, &fromDevice, &toDevice, &envelope, &createdAt, &expiresAt); err != nil {
			s.internalError(w, err)
			return
		}
		items = append(items, map[string]any{
			"signalId": id, "sessionId": sessionID, "kind": kind,
			"fromDeviceId": fromDevice, "toDeviceId": toDevice, "envelope": envelope,
			"createdAt": time.UnixMilli(createdAt).UTC(), "expiresAt": time.UnixMilli(expiresAt).UTC(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"signals": items})
}

func (s *server) putRoomCheckpoint(w http.ResponseWriter, r *http.Request, roomID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	var input struct {
		Envelope         string `json:"envelope"`
		ExpiresInSeconds int64  `json:"expiresInSeconds"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if len(input.Envelope) == 0 || len(input.Envelope) > maxCheckpointPayload {
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "checkpoint envelope must be 1-262144 bytes")
		return
	}
	ttl, ok := boundedTTL(w, input.ExpiresInSeconds, 7*24*time.Hour, 30*24*time.Hour)
	if !ok {
		return
	}
	now := s.now()
	_, err := s.db.Exec(`
		INSERT INTO v2_room_checkpoints(room_id, uploader_device_id, envelope, updated_at, expires_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(room_id) DO UPDATE SET
			uploader_device_id = excluded.uploader_device_id,
			envelope = excluded.envelope,
			updated_at = excluded.updated_at,
			expires_at = excluded.expires_at`,
		roomID, m.DeviceID, input.Envelope, now.UnixMilli(), now.Add(ttl).UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) getRoomCheckpoint(w http.ResponseWriter, r *http.Request, roomID string) {
	if _, ok := s.authenticateMember(w, r, roomID); !ok {
		return
	}
	var uploaderDeviceID, envelope string
	var updatedAt, expiresAt int64
	err := s.db.QueryRow(`
		SELECT uploader_device_id, envelope, updated_at, expires_at
		FROM v2_room_checkpoints
		WHERE room_id = ? AND expires_at > ?`,
		roomID, s.now().UnixMilli()).Scan(&uploaderDeviceID, &envelope, &updatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "checkpoint_not_found", "room checkpoint not found")
		return
	}
	if err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"roomId": roomID, "uploaderDeviceId": uploaderDeviceID, "envelope": envelope,
		"updatedAt": time.UnixMilli(updatedAt).UTC(), "expiresAt": time.UnixMilli(expiresAt).UTC(),
	})
}

// Per-device checkpoints remain available for compatibility. New clients should
// use the room-wide checkpoint endpoint so a newly admitted member can bootstrap.
func (s *server) putCheckpoint(w http.ResponseWriter, r *http.Request, roomID, deviceID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if m.DeviceID != deviceID {
		writeError(w, http.StatusForbidden, "mailbox_forbidden", "checkpoint access is limited to the authenticated device")
		return
	}
	var input struct {
		Envelope         string `json:"envelope"`
		ExpiresInSeconds int64  `json:"expiresInSeconds"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if len(input.Envelope) == 0 || len(input.Envelope) > maxCheckpointPayload {
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "checkpoint envelope must be 1-262144 bytes")
		return
	}
	ttl, ok := boundedTTL(w, input.ExpiresInSeconds, 7*24*time.Hour, 30*24*time.Hour)
	if !ok {
		return
	}
	now := s.now()
	_, err := s.db.Exec(`
		INSERT INTO v2_checkpoints(room_id, device_id, envelope, updated_at, expires_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(room_id, device_id) DO UPDATE SET
			envelope = excluded.envelope, updated_at = excluded.updated_at, expires_at = excluded.expires_at`,
		roomID, deviceID, input.Envelope, now.UnixMilli(), now.Add(ttl).UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) getCheckpoint(w http.ResponseWriter, r *http.Request, roomID, deviceID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if m.DeviceID != deviceID {
		writeError(w, http.StatusForbidden, "mailbox_forbidden", "checkpoint access is limited to the authenticated device")
		return
	}
	var envelope string
	var updatedAt, expiresAt int64
	err := s.db.QueryRow(`
		SELECT envelope, updated_at, expires_at FROM v2_checkpoints
		WHERE room_id = ? AND device_id = ? AND expires_at > ?`,
		roomID, deviceID, s.now().UnixMilli()).Scan(&envelope, &updatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "checkpoint_not_found", "checkpoint not found")
		return
	}
	if err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"deviceId": deviceID, "envelope": envelope,
		"updatedAt": time.UnixMilli(updatedAt).UTC(), "expiresAt": time.UnixMilli(expiresAt).UTC(),
	})
}

func (s *server) postOperation(w http.ResponseWriter, r *http.Request, roomID, recipientDeviceID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if !s.deviceExists(roomID, recipientDeviceID) {
		writeError(w, http.StatusNotFound, "recipient_not_found", "recipient device is not a room member")
		return
	}
	var input struct {
		Envelope         string `json:"envelope"`
		ExpiresInSeconds int64  `json:"expiresInSeconds"`
	}
	if err := decodeJSON(w, r, &input, maxJSONBody); err != nil {
		return
	}
	if len(input.Envelope) == 0 || len(input.Envelope) > maxOperationPayload {
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "operation envelope must be 1-65536 bytes")
		return
	}
	ttl, ok := boundedTTL(w, input.ExpiresInSeconds, 7*24*time.Hour, 30*24*time.Hour)
	if !ok {
		return
	}
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer tx.Rollback()
	_, _ = tx.Exec(`DELETE FROM v2_operations WHERE expires_at <= ?`, now.UnixMilli())
	var count int
	if err = tx.QueryRow(`
		SELECT COUNT(*) FROM v2_operations WHERE room_id = ? AND to_device_id = ?`,
		roomID, recipientDeviceID).Scan(&count); err != nil {
		s.internalError(w, err)
		return
	}
	if count >= maxMailboxOperations {
		writeError(w, http.StatusTooManyRequests, "mailbox_quota_exceeded", "mailbox operation quota exceeded")
		return
	}
	result, err := tx.Exec(`
		INSERT INTO v2_operations(room_id, from_device_id, to_device_id, envelope, created_at, expires_at)
		VALUES(?, ?, ?, ?, ?, ?)`,
		roomID, m.DeviceID, recipientDeviceID, input.Envelope, now.UnixMilli(), now.Add(ttl).UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	id, _ := result.LastInsertId()
	if err = tx.Commit(); err != nil {
		s.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"operationId": id})
}

func (s *server) getOperations(w http.ResponseWriter, r *http.Request, roomID, deviceID string) {
	m, ok := s.authenticateMember(w, r, roomID)
	if !ok {
		return
	}
	if m.DeviceID != deviceID {
		writeError(w, http.StatusForbidden, "mailbox_forbidden", "mailbox reads are limited to the authenticated device")
		return
	}
	after, err := parseAfter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_after", "after must be a non-negative integer")
		return
	}
	rows, err := s.db.Query(`
		SELECT operation_id, from_device_id, envelope, created_at, expires_at
		FROM v2_operations
		WHERE room_id = ? AND to_device_id = ? AND operation_id > ? AND expires_at > ?
		ORDER BY operation_id LIMIT 200`, roomID, deviceID, after, s.now().UnixMilli())
	if err != nil {
		s.internalError(w, err)
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, createdAt, expiresAt int64
		var fromDevice, envelope string
		if err := rows.Scan(&id, &fromDevice, &envelope, &createdAt, &expiresAt); err != nil {
			s.internalError(w, err)
			return
		}
		items = append(items, map[string]any{
			"operationId": id, "fromDeviceId": fromDevice, "toDeviceId": deviceID,
			"envelope": envelope, "createdAt": time.UnixMilli(createdAt).UTC(),
			"expiresAt": time.UnixMilli(expiresAt).UTC(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"operations": items})
}

func (s *server) authenticateMember(w http.ResponseWriter, r *http.Request, roomID string) (member, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") || len(strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))) < 16 {
		writeError(w, http.StatusUnauthorized, "authentication_required", "valid bearer member credential required")
		return member{}, false
	}
	credentialHash := hashSecret(strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")))
	var m member
	err := s.db.QueryRow(`
		SELECT member_id, room_id, device_id, is_owner FROM v2_members
		WHERE room_id = ? AND credential_hash = ?`, roomID, credentialHash).
		Scan(&m.ID, &m.RoomID, &m.DeviceID, &m.IsOwner)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "invalid_credential", "member credential is invalid for this room")
		return member{}, false
	}
	if err != nil {
		s.internalError(w, err)
		return member{}, false
	}
	_, _ = s.db.Exec(`UPDATE v2_members SET last_seen_at = ? WHERE member_id = ?`, s.now().UnixMilli(), m.ID)
	return m, true
}

func (s *server) authenticateOwner(w http.ResponseWriter, r *http.Request, roomID string) bool {
	capability := r.Header.Get("X-Owner-Capability")
	if len(capability) < 16 {
		writeError(w, http.StatusUnauthorized, "owner_authentication_required", "valid X-Owner-Capability required")
		return false
	}
	var stored string
	err := s.db.QueryRow(`SELECT owner_capability_hash FROM v2_rooms WHERE room_id = ?`, roomID).Scan(&stored)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "room_not_found", "room not found")
		return false
	}
	if err != nil {
		s.internalError(w, err)
		return false
	}
	if !secretMatches(stored, capability) {
		writeError(w, http.StatusForbidden, "invalid_owner_capability", "owner capability is invalid")
		return false
	}
	return true
}

func (s *server) deviceExists(roomID, deviceID string) bool {
	var exists int
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM v2_members WHERE room_id = ? AND device_id = ?)`,
		roomID, deviceID).Scan(&exists)
	return err == nil && exists == 1
}

func (s *server) internalError(w http.ResponseWriter, err error) {
	log.Printf("signaling internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		}
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain one JSON value")
		return errors.New("multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var response apiError
	response.Error.Code = code
	response.Error.Message = message
	writeJSON(w, status, response)
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func secretMatches(storedHash, candidate string) bool {
	return constantStringEqual(storedHash, hashSecret(candidate))
}

func constantStringEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func randomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func validID(value string) bool {
	if len(value) < 8 || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func splitPath(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func boundedTTL(w http.ResponseWriter, seconds int64, fallback, maximum time.Duration) (time.Duration, bool) {
	if seconds == 0 {
		return fallback, true
	}
	ttl := time.Duration(seconds) * time.Second
	if seconds < 1 || ttl > maximum {
		writeError(w, http.StatusBadRequest, "invalid_ttl", "expiresInSeconds is outside the allowed range")
		return 0, false
	}
	return ttl, true
}

func parseAfter(r *http.Request) (int64, error) {
	value := r.URL.Query().Get("after")
	if value == "" {
		return 0, nil
	}
	after, err := strconv.ParseInt(value, 10, 64)
	if err != nil || after < 0 {
		return 0, errors.New("invalid after")
	}
	return after, nil
}

func isConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "constraint")
}
