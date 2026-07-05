package factory

import (
	"fmt"
	"strings"
)

// AttributionFooter is the required source block for every X post (compliance).
func AttributionFooter(podcast, youtubeURL string) string {
	podcast = strings.TrimSpace(podcast)
	if podcast == "" {
		podcast = "Podcast"
	}
	youtubeURL = strings.TrimSpace(youtubeURL)
	if youtubeURL == "" {
		return fmt.Sprintf("Source: %s", podcast)
	}
	return fmt.Sprintf("Source: %s\n%s", podcast, youtubeURL)
}

func sourcePodcastFromPostText(postText string) string {
	var line string
	if idx := strings.LastIndex(postText, "\n\nSource:"); idx >= 0 {
		line = postText[idx+len("\n\nSource:"):]
	} else if strings.HasPrefix(postText, "Source:") {
		line = postText[len("Source:"):]
	} else {
		return ""
	}
	if nl := strings.Index(line, "\n"); nl >= 0 {
		line = line[:nl]
	}
	return strings.TrimSpace(line)
}

// EnsurePostTextAttribution appends podcast name + YouTube URL if missing.
// Replaces an existing trailing "Source:" block so attribution stays canonical.
func EnsurePostTextAttribution(postText, podcast, youtubeURL string) string {
	postText = strings.TrimSpace(postText)
	youtubeURL = strings.TrimSpace(youtubeURL)

	if youtubeURL != "" && strings.Contains(postText, youtubeURL) {
		return postText
	}

	if strings.TrimSpace(podcast) == "" {
		if existing := sourcePodcastFromPostText(postText); existing != "" {
			podcast = existing
		}
	}

	if idx := strings.LastIndex(postText, "\n\nSource:"); idx >= 0 {
		postText = strings.TrimSpace(postText[:idx])
	} else if strings.HasPrefix(postText, "Source:") {
		postText = ""
	}

	footer := AttributionFooter(podcast, youtubeURL)
	if postText == "" {
		return footer
	}
	return postText + "\n\n" + footer
}

func (s *Store) enrichCandidatePostText(c Candidate) (Candidate, error) {
	src, err := s.GetSource(c.SourceID)
	if err != nil {
		return c, err
	}
	c.PostText = EnsurePostTextAttribution(c.PostText, src.Podcast, src.YouTubeURL)
	return c, nil
}