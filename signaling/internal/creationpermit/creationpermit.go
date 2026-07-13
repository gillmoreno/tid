package creationpermit

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	Prefix      = "rwp1"
	Purpose     = "create_room"
	MinMembers  = 2
	MaxMembers  = 50
	MinTTL      = 5 * time.Minute
	MaxTTL      = 7 * 24 * time.Hour
	maxTokenLen = 2048
)

type Claims struct {
	Version    int    `json:"v"`
	Purpose    string `json:"purpose"`
	ID         string `json:"jti"`
	MaxMembers int    `json:"maxMembers"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
}

type ErrorKind string

const (
	Invalid     ErrorKind = "invalid"
	Expired     ErrorKind = "expired"
	NotYetValid ErrorKind = "not_yet_valid"
)

type ValidationError struct {
	Kind ErrorKind
}

func (e *ValidationError) Error() string {
	return "creation permit is " + string(e.Kind)
}

func ParsePublicKey(encoded string) (ed25519.PublicKey, error) {
	value, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(value) != ed25519.PublicKeySize {
		return nil, errors.New("ROOMWORKS_CREATOR_VERIFY_KEY must be a base64url Ed25519 public key")
	}
	return ed25519.PublicKey(value), nil
}

func EncodePublicKey(key ed25519.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(key)
}

func GenerateSeed() ([]byte, ed25519.PublicKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	seed := append([]byte(nil), privateKey.Seed()...)
	return seed, publicKey, nil
}

func PublicKeyFromSeed(seed []byte) (ed25519.PublicKey, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("private seed must contain %d bytes", ed25519.SeedSize)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return append(ed25519.PublicKey(nil), publicKey...), nil
}

func Mint(seed []byte, maxMembers int, now time.Time, ttl time.Duration) (string, Claims, error) {
	if len(seed) != ed25519.SeedSize {
		return "", Claims{}, fmt.Errorf("private seed must contain %d bytes", ed25519.SeedSize)
	}
	if maxMembers < MinMembers || maxMembers > MaxMembers {
		return "", Claims{}, fmt.Errorf("capacity must be between %d and %d", MinMembers, MaxMembers)
	}
	if ttl < MinTTL || ttl > MaxTTL {
		return "", Claims{}, fmt.Errorf("ttl must be between %s and %s", MinTTL, MaxTTL)
	}
	tokenID := make([]byte, 18)
	if _, err := rand.Read(tokenID); err != nil {
		return "", Claims{}, err
	}
	now = now.UTC().Truncate(time.Second)
	claims := Claims{
		Version:    1,
		Purpose:    Purpose,
		ID:         base64.RawURLEncoding.EncodeToString(tokenID),
		MaxMembers: maxMembers,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", Claims{}, err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signed := Prefix + "." + encodedPayload
	signature := ed25519.Sign(ed25519.NewKeyFromSeed(seed), []byte(signed))
	return signed + "." + base64.RawURLEncoding.EncodeToString(signature), claims, nil
}

func Verify(token string, publicKey ed25519.PublicKey, now time.Time) (Claims, error) {
	if len(token) == 0 || len(token) > maxTokenLen || len(publicKey) != ed25519.PublicKeySize {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != Prefix {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(parts[2])
	if err != nil || len(signature) != ed25519.SignatureSize ||
		!ed25519.Verify(publicKey, []byte(parts[0]+"."+parts[1]), signature) {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	payload, err := base64.RawURLEncoding.Strict().DecodeString(parts[1])
	if err != nil {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	var claims Claims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	tokenID, err := base64.RawURLEncoding.Strict().DecodeString(claims.ID)
	if err != nil || len(tokenID) < 16 || len(tokenID) > 32 ||
		claims.Version != 1 || claims.Purpose != Purpose ||
		claims.MaxMembers < MinMembers || claims.MaxMembers > MaxMembers ||
		claims.IssuedAt <= 0 || claims.ExpiresAt <= claims.IssuedAt ||
		claims.ExpiresAt-claims.IssuedAt > int64(MaxTTL/time.Second) {
		return Claims{}, &ValidationError{Kind: Invalid}
	}
	nowUnix := now.Unix()
	if claims.IssuedAt > nowUnix+60 {
		return Claims{}, &ValidationError{Kind: NotYetValid}
	}
	if claims.ExpiresAt <= nowUnix {
		return Claims{}, &ValidationError{Kind: Expired}
	}
	return claims, nil
}
