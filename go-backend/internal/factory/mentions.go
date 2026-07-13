package factory

import (
	"encoding/json"
	"strings"
)

type MentionEntry struct {
	Name    string   `json:"name"`
	Handle  string   `json:"handle"`
	Aliases []string `json:"aliases,omitempty"`
	URL     string   `json:"url,omitempty"`
}

type MentionDictionary struct {
	People    []MentionEntry `json:"people"`
	Companies []MentionEntry `json:"companies"`
	Podcasts  []MentionEntry `json:"podcasts"`
	NewsFeeds []MentionEntry `json:"news_feeds"`
}

func ParseMentionDictionary(raw string) MentionDictionary {
	var dict MentionDictionary
	if strings.TrimSpace(raw) == "" {
		return dict
	}
	_ = json.Unmarshal([]byte(raw), &dict)
	return dict
}

func marshalDictionary(dict MentionDictionary) (string, error) {
	raw, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (d MentionDictionary) PromptText() string {
	if len(d.People) == 0 && len(d.Companies) == 0 && len(d.Podcasts) == 0 && len(d.NewsFeeds) == 0 {
		return "(empty — add handles in the Sources dictionary)"
	}
	raw, _ := json.MarshalIndent(d, "", "  ")
	return string(raw)
}

func (d MentionDictionary) ResolvePodcastHandle(podcast string) string {
	return d.resolveHandle(podcast, d.Podcasts)
}

func (d MentionDictionary) ResolvePublicationHandle(publication string) string {
	return d.resolveHandle(publication, d.NewsFeeds)
}

func (d MentionDictionary) PodcastOptions() []PodcastOption {
	return toSourceOptions(d.Podcasts)
}

func (d MentionDictionary) PublicationOptions() []PodcastOption {
	return toSourceOptions(d.NewsFeeds)
}

func toSourceOptions(entries []MentionEntry) []PodcastOption {
	out := make([]PodcastOption, 0, len(entries))
	for _, e := range entries {
		out = append(out, PodcastOption{
			Name:   e.Name,
			Handle: normalizeHandle(e.Handle),
		})
	}
	return out
}

func (d MentionDictionary) resolveHandle(name string, entries []MentionEntry) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)
	for _, e := range entries {
		if strings.EqualFold(e.Name, name) || strings.EqualFold(e.Handle, strings.TrimPrefix(name, "@")) {
			return normalizeHandle(e.Handle)
		}
		for _, alias := range e.Aliases {
			if strings.EqualFold(alias, name) || strings.ToLower(alias) == lower {
				return normalizeHandle(e.Handle)
			}
		}
	}
	return ""
}

func normalizeHandle(handle string) string {
	return strings.TrimPrefix(strings.TrimSpace(handle), "@")
}

func formatTag(handle string) string {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ""
	}
	return "@" + handle
}

// AttributionFooter tags the podcast account (not the YouTube video).
func AttributionFooter(podcast string, dict MentionDictionary) string {
	if handle := dict.ResolvePodcastHandle(podcast); handle != "" {
		return formatTag(handle)
	}
	podcast = strings.TrimSpace(podcast)
	if podcast == "" {
		podcast = "Podcast"
	}
	return "Source: " + podcast
}

func stripAttributionTail(postText string) string {
	postText = strings.TrimSpace(postText)
	for {
		changed := false
		if idx := strings.LastIndex(postText, "\n\nSource:"); idx >= 0 {
			postText = strings.TrimSpace(postText[:idx])
			changed = true
		} else if strings.HasPrefix(postText, "Source:") {
			postText = ""
			changed = true
		}
		lines := strings.Split(postText, "\n")
		if len(lines) > 0 {
			last := strings.TrimSpace(lines[len(lines)-1])
			lower := strings.ToLower(last)
			if strings.HasPrefix(last, "@") ||
				strings.Contains(lower, "youtube.com") ||
				strings.Contains(lower, "youtu.be") {
				postText = strings.TrimSpace(strings.Join(lines[:len(lines)-1], "\n"))
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return postText
}

// ArticleAttributionFooter tags the publication account (or falls back to its name).
func ArticleAttributionFooter(publication string, dict MentionDictionary) string {
	if handle := dict.ResolvePublicationHandle(publication); handle != "" {
		return formatTag(handle)
	}
	publication = strings.TrimSpace(publication)
	if publication == "" {
		publication = "Source"
	}
	return "Source: " + publication
}

// EnsureArticlePostAttribution appends the publication @ tag to article posts.
func EnsureArticlePostAttribution(postText, publication string, dict MentionDictionary) string {
	postText = stripAttributionTail(strings.TrimSpace(postText))
	footer := ArticleAttributionFooter(publication, dict)
	if postText == "" {
		return footer
	}
	if strings.Contains(postText, footer) {
		return postText
	}
	return postText + "\n\n" + footer
}

// EnsurePostTextAttribution appends the podcast @ tag (never a YouTube URL).
func EnsurePostTextAttribution(postText, podcast string, dict MentionDictionary) string {
	raw := strings.TrimSpace(postText)
	if strings.TrimSpace(podcast) == "" {
		if existing := sourcePodcastFromPostText(raw); existing != "" {
			podcast = existing
		}
	}
	postText = stripAttributionTail(raw)

	footer := AttributionFooter(podcast, dict)
	if postText == "" {
		return footer
	}
	if strings.Contains(postText, footer) {
		return postText
	}
	return postText + "\n\n" + footer
}
