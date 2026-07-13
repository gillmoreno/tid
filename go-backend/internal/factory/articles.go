package factory

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ArticlePromptName is the fixed name under which the editable article analysis
// prompt is stored in prompt_templates (kept inactive so it never collides with
// the clip factory's single active prompt).
const ArticlePromptName = "article-analysis"

// ArticleSource is a written-article URL ingested into the Articles factory.
type ArticleSource struct {
	ID           string  `json:"id"`
	URL          string  `json:"url"`
	Publication  string  `json:"publication"`
	Title        string  `json:"title"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
	AnalyzedAt   *string `json:"analyzed_at,omitempty"`
}

// ArticleCandidate is a proposed X post derived from an article.
type ArticleCandidate struct {
	ID             string  `json:"id"`
	SourceID       string  `json:"source_id"`
	Rank           int     `json:"rank"`
	PostText       string  `json:"post_text"`
	WhyInteresting string  `json:"why_interesting"`
	Confidence     float64 `json:"confidence"`
	Status         string  `json:"status"`
	ScheduledAt    *string `json:"scheduled_at,omitempty"`
	PreparedAt     *string `json:"prepared_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ArticleAnalysisCandidate is the raw shape returned by analyze.sh.
type ArticleAnalysisCandidate struct {
	PostText       string  `json:"post_text"`
	WhyInteresting string  `json:"why_interesting"`
	Confidence     float64 `json:"confidence"`
}

type ArticleAnalysisResult struct {
	Title      string                     `json:"title"`
	Candidates []ArticleAnalysisCandidate `json:"candidates"`
}

// NewArticleSourceID builds a date-prefixed slug id from the publication (or host).
func NewArticleSourceID(articleURL, publication string) string {
	slug := slugSanitizer.ReplaceAllString(strings.ToLower(publication), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		if u, err := url.Parse(articleURL); err == nil {
			host := strings.TrimPrefix(u.Hostname(), "www.")
			slug = slugSanitizer.ReplaceAllString(strings.ToLower(host), "-")
			slug = strings.Trim(slug, "-")
		}
	}
	if slug == "" {
		slug = "article"
	}
	return fmt.Sprintf("%s-%s-%d", time.Now().Format("20060102"), slug, time.Now().Unix()%100000)
}

func (s *Store) CreateArticleSource(id, articleURL, publication string) (ArticleSource, error) {
	_, err := s.db.Exec(`
		INSERT INTO article_sources (id, url, publication, status)
		VALUES (?, ?, ?, 'ingested')`, id, articleURL, publication)
	if err != nil {
		return ArticleSource{}, err
	}
	return s.GetArticleSource(id)
}

func scanArticleSource(row interface{ Scan(...any) error }) (ArticleSource, error) {
	var src ArticleSource
	var analyzed sql.NullString
	err := row.Scan(
		&src.ID, &src.URL, &src.Publication, &src.Title, &src.Status,
		&src.ErrorMessage, &src.CreatedAt, &analyzed)
	if err != nil {
		return src, err
	}
	if analyzed.Valid {
		src.AnalyzedAt = &analyzed.String
	}
	return src, nil
}

func (s *Store) GetArticleSource(id string) (ArticleSource, error) {
	return scanArticleSource(s.db.QueryRow(`
		SELECT id, url, publication, title, status, error_message, created_at, analyzed_at
		FROM article_sources WHERE id = ?`, id))
}

func (s *Store) ListArticleSources() ([]ArticleSource, error) {
	rows, err := s.db.Query(`
		SELECT id, url, publication, title, status, error_message, created_at, analyzed_at
		FROM article_sources ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ArticleSource
	for rows.Next() {
		src, err := scanArticleSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, src)
	}
	return out, rows.Err()
}

func (s *Store) SetArticleSourceStatus(id, status, errMsg string) error {
	_, err := s.db.Exec(`UPDATE article_sources SET status = ?, error_message = ? WHERE id = ?`, status, errMsg, id)
	return err
}

func (s *Store) SetArticleSourceTitle(id, title string) error {
	if strings.TrimSpace(title) == "" {
		return nil
	}
	_, err := s.db.Exec(`UPDATE article_sources SET title = ? WHERE id = ?`, title, id)
	return err
}

func (s *Store) MarkArticleSourceAnalyzed(id string) error {
	_, err := s.db.Exec(`UPDATE article_sources SET status = 'analyzed', analyzed_at = datetime('now'), error_message = '' WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteArticleSource(id string) error {
	res, err := s.db.Exec(`DELETE FROM article_sources WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("article source not found")
	}
	return nil
}

func (s *Store) InsertArticleCandidates(sourceID string, items []ArticleAnalysisCandidate) ([]ArticleCandidate, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM article_candidates WHERE source_id = ?`, sourceID); err != nil {
		return nil, err
	}

	src, err := s.GetArticleSource(sourceID)
	if err != nil {
		return nil, err
	}
	mentions, err := s.GetActiveMentions()
	if err != nil {
		return nil, err
	}
	dict := ParseMentionDictionary(mentions.Content)

	var out []ArticleCandidate
	for i, item := range items {
		id := fmt.Sprintf("%s-c%02d", sourceID, i+1)
		postText := EnsureArticlePostAttribution(item.PostText, src.Publication, dict)
		_, err := tx.Exec(`
			INSERT INTO article_candidates (
				id, source_id, rank, post_text, why_interesting, confidence, status
			) VALUES (?, ?, ?, ?, ?, ?, 'proposed')`,
			id, sourceID, i+1, postText, item.WhyInteresting, item.Confidence)
		if err != nil {
			return nil, err
		}
		c, err := scanArticleCandidate(tx.QueryRow(articleCandidateSelect+` WHERE id = ?`, id))
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

const articleCandidateSelect = `
	SELECT id, source_id, rank, post_text, why_interesting, confidence,
	       status, scheduled_at, prepared_at, created_at, updated_at
	FROM article_candidates`

func scanArticleCandidate(row interface{ Scan(...any) error }) (ArticleCandidate, error) {
	var c ArticleCandidate
	var scheduled, prepared sql.NullString
	err := row.Scan(
		&c.ID, &c.SourceID, &c.Rank, &c.PostText, &c.WhyInteresting, &c.Confidence,
		&c.Status, &scheduled, &prepared, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	if scheduled.Valid {
		c.ScheduledAt = &scheduled.String
	}
	if prepared.Valid {
		c.PreparedAt = &prepared.String
	}
	return c, nil
}

func (s *Store) GetArticleCandidate(id string) (ArticleCandidate, error) {
	return scanArticleCandidate(s.db.QueryRow(articleCandidateSelect+` WHERE id = ?`, id))
}

func (s *Store) ListArticleCandidates(sourceID string) ([]ArticleCandidate, error) {
	query := articleCandidateSelect
	args := []any{}
	if sourceID != "" {
		query += ` WHERE source_id = ?`
		args = append(args, sourceID)
	}
	query += ` ORDER BY source_id DESC, rank ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ArticleCandidate
	for rows.Next() {
		c, err := scanArticleCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpdateArticleCandidate(id, postText, status string) (ArticleCandidate, error) {
	_, err := s.db.Exec(`
		UPDATE article_candidates SET
			post_text = COALESCE(NULLIF(?, ''), post_text),
			status = COALESCE(NULLIF(?, ''), status),
			updated_at = datetime('now')
		WHERE id = ?`, postText, status, id)
	if err != nil {
		return ArticleCandidate{}, err
	}
	return s.GetArticleCandidate(id)
}

func (s *Store) DeleteArticleCandidate(id string) error {
	res, err := s.db.Exec(`DELETE FROM article_candidates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("article candidate not found")
	}
	return nil
}

func (s *Store) ScheduleArticleCandidate(id string, at time.Time) (ArticleCandidate, error) {
	_, err := s.db.Exec(`
		UPDATE article_candidates SET
			scheduled_at = ?, status = 'scheduled', prepared_at = NULL, updated_at = datetime('now')
		WHERE id = ?`, at.UTC().Format(time.RFC3339), id)
	if err != nil {
		return ArticleCandidate{}, err
	}
	return s.GetArticleCandidate(id)
}

func (s *Store) MarkArticleCandidatePrepared(id string) error {
	_, err := s.db.Exec(`
		UPDATE article_candidates SET status = 'ready', prepared_at = datetime('now'), updated_at = datetime('now')
		WHERE id = ?`, id)
	return err
}

// ListScheduledArticleCandidates returns candidates that have a schedule set, soonest first.
func (s *Store) ListScheduledArticleCandidates() ([]ArticleCandidate, error) {
	rows, err := s.db.Query(articleCandidateSelect + `
		WHERE scheduled_at IS NOT NULL AND scheduled_at != ''
		ORDER BY scheduled_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ArticleCandidate
	for rows.Next() {
		c, err := scanArticleCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListDueArticleCandidates(now time.Time) ([]ArticleCandidate, error) {
	rows, err := s.db.Query(articleCandidateSelect+`
		WHERE status = 'scheduled' AND scheduled_at IS NOT NULL AND scheduled_at <= ?
		ORDER BY scheduled_at ASC`, now.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ArticleCandidate
	for rows.Next() {
		c, err := scanArticleCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
