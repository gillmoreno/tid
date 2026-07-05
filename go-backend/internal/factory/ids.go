package factory

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func NewSourceID(youtubeURL, podcast string) string {
	videoID := extractVideoID(youtubeURL)
	slug := slugSanitizer.ReplaceAllString(strings.ToLower(podcast), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = videoID[:min(8, len(videoID))]
	}
	return fmt.Sprintf("%s-%s", time.Now().Format("20060102"), slug)
}

func extractVideoID(url string) string {
	if idx := strings.Index(url, "v="); idx >= 0 {
		id := url[idx+2:]
		if amp := strings.Index(id, "&"); amp >= 0 {
			id = id[:amp]
		}
		return id
	}
	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "?")[0]
		}
	}
	return "unknown"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchYouTubeMetadata fetches title and channel name using yt-dlp (no download).
// Returns empty strings on failure.
func FetchYouTubeMetadata(youtubeURL string) (title, podcast string) {
	if youtubeURL == "" {
		return "", ""
	}
	cmd := exec.Command("yt-dlp",
		"--print", "%(title)s",
		"--print", "%(channel)s",
		"--no-download",
		"--quiet",
		"--no-warnings",
		youtubeURL,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", ""
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(parts) > 0 {
		title = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		podcast = strings.TrimSpace(parts[1])
	}
	return title, podcast
}