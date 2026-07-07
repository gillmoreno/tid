package factory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Idea struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Kind        string   `json:"kind"`
	Status      string   `json:"status"`
	Summary     string   `json:"summary"`
	Body        string   `json:"body"`
	XPost       string   `json:"x_post"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	PublishedAt *string  `json:"published_at,omitempty"`
}

func Slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = slugSanitizer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = fmt.Sprintf("idea-%d", time.Now().Unix())
	}
	return s
}

func NewIdeaID(slug string) string {
	return fmt.Sprintf("idea-%s", slug)
}

func (s *Store) SeedIdeas() error {
	count, err := s.ideaCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seed := Idea{
		Title: "AI Glossary 2026",
		Slug:  "ai-glossary-2026",
		Kind:  "glossary",
		Status: "idea",
		Summary: "Plain-language definitions of AI terms people actually use in 2026 — harness, agent, prompt engineering, memory, KV cache, pre-fill, and more. Colloquial explanations with examples for newcomers.",
		Tags:  []string{"glossary", "ai", "education"},
	}
	_, err = s.CreateIdea(seed)
	return err
}

func (s *Store) ideaCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ideas`).Scan(&n)
	return n, err
}

func encodeTags(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	raw, _ := json.Marshal(tags)
	return string(raw)
}

func decodeTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return []string{}
	}
	return tags
}

func scanIdea(row interface{ Scan(...any) error }) (Idea, error) {
	var idea Idea
	var tagsRaw string
	var published sql.NullString
	err := row.Scan(
		&idea.ID, &idea.Title, &idea.Slug, &idea.Kind, &idea.Status,
		&idea.Summary, &idea.Body, &idea.XPost, &tagsRaw,
		&idea.CreatedAt, &idea.UpdatedAt, &published,
	)
	if err != nil {
		return idea, err
	}
	idea.Tags = decodeTags(tagsRaw)
	if published.Valid {
		idea.PublishedAt = &published.String
	}
	return idea, nil
}

func (s *Store) ListIdeas() ([]Idea, error) {
	rows, err := s.db.Query(`
		SELECT id, title, slug, kind, status, summary, body, x_post, tags,
		       created_at, updated_at, published_at
		FROM ideas
		ORDER BY
			CASE status
				WHEN 'idea' THEN 0
				WHEN 'drafting' THEN 1
				WHEN 'ready' THEN 2
				WHEN 'published' THEN 3
				ELSE 4
			END,
			updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Idea
	for rows.Next() {
		idea, err := scanIdea(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, idea)
	}
	return out, rows.Err()
}

func (s *Store) GetIdea(id string) (Idea, error) {
	return scanIdea(s.db.QueryRow(`
		SELECT id, title, slug, kind, status, summary, body, x_post, tags,
		       created_at, updated_at, published_at
		FROM ideas WHERE id = ?`, id))
}

func (s *Store) CreateIdea(input Idea) (Idea, error) {
	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = Slugify(input.Title)
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = NewIdeaID(slug)
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		kind = "essay"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "idea"
	}

	_, err := s.db.Exec(`
		INSERT INTO ideas (id, title, slug, kind, status, summary, body, x_post, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, strings.TrimSpace(input.Title), slug, kind, status,
		strings.TrimSpace(input.Summary), strings.TrimSpace(input.Body),
		strings.TrimSpace(input.XPost), encodeTags(input.Tags),
	)
	if err != nil {
		return Idea{}, err
	}
	return s.GetIdea(id)
}

func (s *Store) UpdateIdea(id string, idea Idea) (Idea, error) {
	_, err := s.db.Exec(`
		UPDATE ideas SET
			title = ?,
			kind = ?,
			status = ?,
			summary = ?,
			body = ?,
			x_post = ?,
			tags = ?,
			updated_at = datetime('now')
		WHERE id = ?`,
		strings.TrimSpace(idea.Title),
		strings.TrimSpace(idea.Kind),
		strings.TrimSpace(idea.Status),
		idea.Summary,
		idea.Body,
		idea.XPost,
		encodeTags(idea.Tags),
		id,
	)
	if err != nil {
		return Idea{}, err
	}
	return s.GetIdea(id)
}

func (s *Store) DeleteIdea(id string) error {
	res, err := s.db.Exec(`DELETE FROM ideas WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("idea not found")
	}
	return nil
}