package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          int
	PublicBaseURL string
	SiteName      string
	SiteTagline   string
}

func Load() Config {
	return Config{
		Port:          getEnvInt("GO_API_PORT", 8080),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8000"),
		SiteName:      getEnv("SITE_NAME", "The Idea Guy"),
		SiteTagline:   getEnv("SITE_TAGLINE", "Testing what's real in AI — one person, real loops, no hype."),
	}
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