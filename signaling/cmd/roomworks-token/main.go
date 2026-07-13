package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"signaling/internal/creationpermit"
)

func main() {
	syscall.Umask(0o077)
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "roomworks-token:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("expected init, public-key, or mint command")
	}
	switch args[0] {
	case "init":
		return initKey(args[1:])
	case "public-key":
		return printPublicKey(args[1:])
	case "mint":
		return mint(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func defaultSeedPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "creator-signing.seed"
	}
	return filepath.Join(home, ".config", "roomworks", "creator-signing.seed")
}

func initKey(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	seedPath := flags.String("seed-file", defaultSeedPath(), "private signing seed path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("init accepts flags only")
	}
	if _, err := os.Lstat(*seedPath); err == nil {
		return fmt.Errorf("refusing to replace existing key %s", *seedPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	seed, publicKey, err := creationpermit.GenerateSeed()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*seedPath), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(*seedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	encoded := base64.RawURLEncoding.EncodeToString(seed) + "\n"
	if _, err = file.WriteString(encoded); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	fmt.Printf("Initialized creator signing key: %s\n", *seedPath)
	fmt.Printf("ROOMWORKS_CREATOR_VERIFY_KEY=%s\n", creationpermit.EncodePublicKey(publicKey))
	return nil
}

func printPublicKey(args []string) error {
	flags := flag.NewFlagSet("public-key", flag.ContinueOnError)
	seedPath := flags.String("seed-file", defaultSeedPath(), "private signing seed path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("public-key accepts flags only")
	}
	seed, err := loadSeed(*seedPath)
	if err != nil {
		return err
	}
	publicKey, err := creationpermit.PublicKeyFromSeed(seed)
	if err != nil {
		return err
	}
	fmt.Println(creationpermit.EncodePublicKey(publicKey))
	return nil
}

func mint(args []string) error {
	flags := flag.NewFlagSet("mint", flag.ContinueOnError)
	capacity := flags.Int("capacity", 0, "total unique members, including the creator")
	ttl := flags.Duration("ttl", 24*time.Hour, "permit lifetime")
	seedPath := flags.String("seed-file", defaultSeedPath(), "private signing seed path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("mint accepts flags only")
	}
	seed, err := loadSeed(*seedPath)
	if err != nil {
		return err
	}
	token, claims, err := creationpermit.Mint(seed, *capacity, time.Now(), *ttl)
	if err != nil {
		return err
	}
	copyCommand := exec.Command("pbcopy")
	copyCommand.Stdin = strings.NewReader(token)
	if err := copyCommand.Run(); err != nil {
		return fmt.Errorf("copy token to clipboard: %w", err)
	}
	fmt.Printf("Token copied: capacity=%d expires_at=%s single_use=true\n",
		claims.MaxMembers, time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339))
	return nil
}

func loadSeed(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("private key must be a regular file: %s", path)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("private key must not be accessible by group or others: %s", path)
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return nil, fmt.Errorf("private key must be owned by the current user: %s", path)
	}
	if info.Size() > 256 {
		return nil, errors.New("private key file is unexpectedly large")
	}
	encoded, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seed, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("private key file does not contain a base64url Ed25519 seed")
	}
	return seed, nil
}
