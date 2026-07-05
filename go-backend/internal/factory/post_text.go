package factory

import (
	"strings"
)

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

func (s *Store) enrichCandidatePostText(c Candidate) (Candidate, error) {
	src, err := s.GetSource(c.SourceID)
	if err != nil {
		return c, err
	}
	mentions, err := s.GetActiveMentions()
	if err != nil {
		return c, err
	}
	dict := ParseMentionDictionary(mentions.Content)
	c.PostText = EnsurePostTextAttribution(c.PostText, src.Podcast, dict)
	return c, nil
}