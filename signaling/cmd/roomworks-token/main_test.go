package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func writeSeed(t *testing.T, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "creator-signing.seed")
	seed := bytes.Repeat([]byte{0x23}, ed25519.SeedSize)
	if err := os.WriteFile(path, []byte(base64.RawURLEncoding.EncodeToString(seed)+"\n"), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadSeedRequiresPrivateRegularFile(t *testing.T) {
	path := writeSeed(t, 0o600)
	seed, err := loadSeed(path)
	if err != nil || len(seed) != ed25519.SeedSize {
		t.Fatalf("load seed: len=%d err=%v", len(seed), err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSeed(path); err == nil {
		t.Fatal("group/world-readable seed was accepted")
	}

	target := writeSeed(t, 0o600)
	symlink := filepath.Join(t.TempDir(), "seed-link")
	if err := os.Symlink(target, symlink); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSeed(symlink); err == nil {
		t.Fatal("symlinked seed was accepted")
	}
}

func TestInitKeyCreatesPrivateSeedAndRefusesReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys", "creator-signing.seed")
	if err := initKey([]string{"--seed-file", path}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("seed mode=%#o", info.Mode().Perm())
	}
	if _, err := loadSeed(path); err != nil {
		t.Fatal(err)
	}
	if err := initKey([]string{"--seed-file", path}); err == nil {
		t.Fatal("existing seed was replaced")
	}
}
