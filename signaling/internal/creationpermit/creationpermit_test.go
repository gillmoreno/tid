package creationpermit

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func testSeed() []byte {
	return bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
}

func signClaims(t *testing.T, seed []byte, claims any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	signed := Prefix + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(ed25519.NewKeyFromSeed(seed), []byte(signed))
	return signed + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func requireKind(t *testing.T, err error, kind ErrorKind) {
	t.Helper()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || validationErr.Kind != kind {
		t.Fatalf("error=%v want kind=%s", err, kind)
	}
}

func TestMintAndVerify(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	seed := testSeed()
	publicKey, err := PublicKeyFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	token, minted, err := Mint(seed, 11, now, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(token, Prefix+".") || strings.Count(token, ".") != 2 {
		t.Fatalf("unexpected token format")
	}
	verified, err := Verify(token, publicKey, now)
	if err != nil {
		t.Fatal(err)
	}
	if verified != minted || verified.MaxMembers != 11 || verified.ExpiresAt != now.Add(24*time.Hour).Unix() {
		t.Fatalf("verified=%+v minted=%+v", verified, minted)
	}
	encoded := EncodePublicKey(publicKey)
	parsed, err := ParsePublicKey(encoded)
	if err != nil || !bytes.Equal(parsed, publicKey) {
		t.Fatalf("public key round trip: key=%x err=%v", parsed, err)
	}
}

func TestMintRejectsBounds(t *testing.T) {
	for _, capacity := range []int{1, 51} {
		if _, _, err := Mint(testSeed(), capacity, time.Now(), time.Hour); err == nil {
			t.Fatalf("capacity %d was accepted", capacity)
		}
	}
	for _, ttl := range []time.Duration{time.Minute, 8 * 24 * time.Hour} {
		if _, _, err := Mint(testSeed(), 3, time.Now(), ttl); err == nil {
			t.Fatalf("ttl %s was accepted", ttl)
		}
	}
}

func TestVerifyRejectsTamperingAndMalformedTokens(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	seed := testSeed()
	publicKey, _ := PublicKeyFromSeed(seed)
	token, _, _ := Mint(seed, 3, now, time.Hour)
	parts := strings.Split(token, ".")
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"v":1,"purpose":"create_room","jti":"AAAAAAAAAAAAAAAAAAAAAA","maxMembers":11,"iat":1,"exp":2}`))
	requireKind(t, func() error { _, err := Verify(strings.Join(parts, "."), publicKey, now); return err }(), Invalid)

	for _, malformed := range []string{"", "rwp1.only", "wrong.a.b", strings.Repeat("x", maxTokenLen+1)} {
		_, err := Verify(malformed, publicKey, now)
		requireKind(t, err, Invalid)
	}
	if _, err := ParsePublicKey("not-a-public-key"); err == nil {
		t.Fatal("invalid public key accepted")
	}
}

func TestVerifyTimeAndClaimValidation(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	seed := testSeed()
	publicKey, _ := PublicKeyFromSeed(seed)
	base := Claims{
		Version: 1, Purpose: Purpose, ID: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 18)),
		MaxMembers: 3, IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
	}

	expired := base
	expired.IssuedAt = now.Add(-time.Hour).Unix()
	expired.ExpiresAt = now.Unix()
	_, err := Verify(signClaims(t, seed, expired), publicKey, now)
	requireKind(t, err, Expired)

	future := base
	future.IssuedAt = now.Add(2 * time.Minute).Unix()
	future.ExpiresAt = now.Add(time.Hour).Unix()
	_, err = Verify(signClaims(t, seed, future), publicKey, now)
	requireKind(t, err, NotYetValid)

	unknownField := map[string]any{
		"v": 1, "purpose": Purpose, "jti": base.ID, "maxMembers": 3,
		"iat": now.Unix(), "exp": now.Add(time.Hour).Unix(), "admin": true,
	}
	_, err = Verify(signClaims(t, seed, unknownField), publicKey, now)
	requireKind(t, err, Invalid)
}
