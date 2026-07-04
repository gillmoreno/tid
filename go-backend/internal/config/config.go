package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port          int
	PublicBaseURL string
	SiteName      string
	SiteTagline   string
	RepoRoot      string
	DatabasePath  string
}

func Load() Config {
	repoRoot := getEnv("TID_REPO_ROOT", findRepoRoot())
	dbPath := getEnv("DATABASE_PATH", filepath.Join(repoRoot, "data", "factory", "tid.db"))

	return Config{
		Port:          getEnvInt("GO_API_PORT", 8080),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8000"),
		SiteName:      getEnv("SITE_NAME", "The Idea Guy"),
		SiteTagline:   getEnv("SITE_TAGLINE", "Testing what's real in AI — one person, real loops, no hype."),
		RepoRoot:      repoRoot,
		DatabasePath:  dbPath,
	}
}

func findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go-backend")); err == nil {
			if _, err2 := os.Stat(filepath.Join(dir, "loops", "clip-to-post")); err2 == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}