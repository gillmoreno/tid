package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Site struct {
	Name        string `json:"name"`
	Tagline     string `json:"tagline"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Mission     string `json:"mission"`
}

type Section struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	ExternalURL string `json:"external_url,omitempty"`
}

type Item struct {
	Slug        string   `json:"slug"`
	SectionSlug string   `json:"section_slug"`
	Title       string   `json:"title"`
	Tagline     string   `json:"tagline"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
	ExternalURL string   `json:"external_url,omitempty"`
	Featured    bool     `json:"featured"`
}

type sectionDetail struct {
	Section
	Items []Item `json:"items"`
}

func (a *App) handleSite(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, Site{
		Name:        a.cfg.SiteName,
		Tagline:     a.cfg.SiteTagline,
		Domain:      "theideaguy.com",
		Description: "A personal lab for testing AI ideas — separating signal from hype, and showing what one builder can actually ship.",
		Mission:     "Cut through AI noise. Build small loops. Share what works.",
	})
}

func (a *App) handleSections(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, allSections())
}

func (a *App) handleSection(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(chi.URLParam(r, "slug"))
	section, ok := sectionBySlug(slug)
	if !ok {
		writeError(w, http.StatusNotFound, "section not found")
		return
	}
	writeJSON(w, sectionDetail{
		Section: section,
		Items:   itemsForSection(slug),
	})
}

func (a *App) handleItem(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(chi.URLParam(r, "slug"))
	item, ok := itemBySlug(slug)
	if !ok {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, item)
}

func allSections() []Section {
	return []Section{
		{
			Slug:        "ai-guild",
			Title:       "AI Guild of Dev",
			Subtitle:    "aigil.dev",
			Description: "A developer guild for people building with AI — community, tools, and experiments in the open.",
			Icon:        "guild",
			Order:       1,
			ExternalURL: "https://aigil.dev/",
		},
		{
			Slug:        "loops-with-taste",
			Title:       "Loops with Taste",
			Subtitle:    "Content automation, human-filtered",
			Description: "AI loops that turn raw signal into posts and shorts — but only after my taste filter. The hook is the output, not the pipeline.",
			Icon:        "loop",
			Order:       2,
		},
		{
			Slug:        "software-ideas",
			Title:       "Software Ideas",
			Subtitle:    "Things I'm building",
			Description: "Software experiments and products — small bets on what AI makes possible for a solo builder today.",
			Icon:        "code",
			Order:       3,
		},
	}
}

func sectionBySlug(slug string) (Section, bool) {
	for _, s := range allSections() {
		if s.Slug == slug {
			return s, true
		}
	}
	return Section{}, false
}

func allItems() []Item {
	return []Item{
		{
			Slug:        "aigil",
			SectionSlug: "ai-guild",
			Title:       "AIGil",
			Tagline:     "Developer guild for the AI era",
			Description: "Community and tooling for developers navigating AI — what's worth learning, what's worth building, and what's just noise.",
			Status:      "live",
			Tags:        []string{"community", "developers", "ai"},
			ExternalURL: "https://aigil.dev/",
			Featured:    true,
		},
		{
			Slug:        "clip-to-post",
			SectionSlug: "loops-with-taste",
			Title:       "Clip → Post",
			Tagline:     "YouTube signal, X output",
			Description: "Take a small slice of an AI YouTube clip that caught my attention, extract the insight, and rewrite it as tweets or X posts — filtered through my taste, not generic AI slop.",
			Status:      "testing",
			Tags:        []string{"x", "youtube", "content", "loop"},
			Featured:    true,
		},
		{
			Slug:        "grok-shorts",
			SectionSlug: "loops-with-taste",
			Title:       "Grok Imagine Shorts",
			Tagline:     "AI visuals, tight edits",
			Description: "A loop for creating short-form video using Grok Imagine — fast visual generation paired with tight editing for reels and shorts.",
			Status:      "testing",
			Tags:        []string{"video", "shorts", "grok", "imagine"},
			Featured:    true,
		},
		{
			Slug:        "roger",
			SectionSlug: "software-ideas",
			Title:       "Roger",
			Tagline:     "Restaurant reservations, rebuilt",
			Description: "Multi-restaurant reservation platform — Go API, React widget, the full stack. Proof that one person can ship production software with modern AI-assisted development.",
			Status:      "live",
			Tags:        []string{"go", "react", "saas", "hospitality"},
			Featured:    true,
		},
		{
			Slug:        "tid",
			SectionSlug: "software-ideas",
			Title:       "The Idea Guy",
			Tagline:     "This site",
			Description: "The meta-project — a personal lab site to document and test AI ideas in public. Go backend, React frontend, Roger-style deploy.",
			Status:      "building",
			Tags:        []string{"go", "react", "brand", "lab"},
			Featured:    true,
		},
	}
}

func itemsForSection(sectionSlug string) []Item {
	var items []Item
	for _, item := range allItems() {
		if item.SectionSlug == sectionSlug {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []Item{}
	}
	return items
}

func itemBySlug(slug string) (Item, bool) {
	for _, item := range allItems() {
		if item.Slug == slug {
			return item, true
		}
	}
	return Item{}, false
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}