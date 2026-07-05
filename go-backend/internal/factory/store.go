package factory

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tid/go-backend/internal/db"
)

type Store struct {
	db *sql.DB
}

func NewStore(p *db.Provider) *Store {
	return &Store{db: p.DB}
}

func (s *Store) SeedDefaults(biasContent, promptContent string) error {
	if strings.TrimSpace(biasContent) == "" || strings.TrimSpace(promptContent) == "" {
		return fmt.Errorf("empty seed content")
	}
	var biasCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM bias_profiles`).Scan(&biasCount); err != nil {
		return err
	}
	if biasCount == 0 {
		_, err := s.db.Exec(`INSERT INTO bias_profiles (name, content, is_active) VALUES ('default', ?, 1)`, biasContent)
		if err != nil {
			return err
		}
	}
	var promptCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM prompt_templates`).Scan(&promptCount); err != nil {
		return err
	}
	if promptCount == 0 {
		_, err := s.db.Exec(`INSERT INTO prompt_templates (name, content, is_active) VALUES ('default', ?, 1)`, promptContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SeedMentions(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty mentions seed")
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM mention_dictionaries`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := s.db.Exec(`INSERT INTO mention_dictionaries (name, content, is_active) VALUES ('default', ?, 1)`, content)
		return err
	}
	return nil
}

func (s *Store) GetActiveMentions() (MentionDictionaryProfile, error) {
	var m MentionDictionaryProfile
	var active int
	err := s.db.QueryRow(`
		SELECT id, name, content, is_active, updated_at
		FROM mention_dictionaries WHERE is_active = 1
		ORDER BY updated_at DESC LIMIT 1`).Scan(&m.ID, &m.Name, &m.Content, &active, &m.UpdatedAt)
	if err != nil {
		return m, err
	}
	m.IsActive = active == 1
	return m, nil
}

func (s *Store) UpdateActiveMentions(content string) (MentionDictionaryProfile, error) {
	_, err := s.db.Exec(`UPDATE mention_dictionaries SET is_active = 0`)
	if err != nil {
		return MentionDictionaryProfile{}, err
	}
	_, err = s.db.Exec(`
		INSERT INTO mention_dictionaries (name, content, is_active, updated_at)
		VALUES ('default', ?, 1, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET content = excluded.content, is_active = 1, updated_at = datetime('now')`, content)
	if err != nil {
		return MentionDictionaryProfile{}, err
	}
	return s.GetActiveMentions()
}

func (s *Store) GetActiveBias() (BiasProfile, error) {
	var b BiasProfile
	var active int
	err := s.db.QueryRow(`
		SELECT id, name, content, is_active, updated_at
		FROM bias_profiles WHERE is_active = 1
		ORDER BY updated_at DESC LIMIT 1`).Scan(&b.ID, &b.Name, &b.Content, &active, &b.UpdatedAt)
	if err != nil {
		return b, err
	}
	b.IsActive = active == 1
	return b, nil
}

func (s *Store) UpdateActiveBias(content string) (BiasProfile, error) {
	_, err := s.db.Exec(`UPDATE bias_profiles SET is_active = 0`)
	if err != nil {
		return BiasProfile{}, err
	}
	_, err = s.db.Exec(`
		INSERT INTO bias_profiles (name, content, is_active, updated_at)
		VALUES ('default', ?, 1, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET content = excluded.content, is_active = 1, updated_at = datetime('now')`, content)
	if err != nil {
		return BiasProfile{}, err
	}
	return s.GetActiveBias()
}

func (s *Store) GetActivePrompt() (PromptTemplate, error) {
	var p PromptTemplate
	var active int
	err := s.db.QueryRow(`
		SELECT id, name, content, is_active, updated_at
		FROM prompt_templates WHERE is_active = 1
		ORDER BY updated_at DESC LIMIT 1`).Scan(&p.ID, &p.Name, &p.Content, &active, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	p.IsActive = active == 1
	return p, nil
}

func (s *Store) UpdateActivePrompt(content string) (PromptTemplate, error) {
	_, err := s.db.Exec(`UPDATE prompt_templates SET is_active = 0`)
	if err != nil {
		return PromptTemplate{}, err
	}
	_, err = s.db.Exec(`
		INSERT INTO prompt_templates (name, content, is_active, updated_at)
		VALUES ('default', ?, 1, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET content = excluded.content, is_active = 1, updated_at = datetime('now')`, content)
	if err != nil {
		return PromptTemplate{}, err
	}
	return s.GetActivePrompt()
}

func (s *Store) CreateSource(id, url, title, podcast string) (Source, error) {
	_, err := s.db.Exec(`
		INSERT INTO sources (id, youtube_url, title, podcast, status)
		VALUES (?, ?, ?, ?, 'ingested')`, id, url, title, podcast)
	if err != nil {
		return Source{}, err
	}
	return s.GetSource(id)
}

func (s *Store) GetSource(id string) (Source, error) {
	var src Source
	var analyzed sql.NullString
	err := s.db.QueryRow(`
		SELECT id, youtube_url, title, podcast, status, error_message, created_at, analyzed_at
		FROM sources WHERE id = ?`, id).Scan(
		&src.ID, &src.YouTubeURL, &src.Title, &src.Podcast, &src.Status, &src.ErrorMessage, &src.CreatedAt, &analyzed)
	if err != nil {
		return src, err
	}
	if analyzed.Valid {
		src.AnalyzedAt = &analyzed.String
	}
	return src, nil
}

func (s *Store) ListSources() ([]Source, error) {
	rows, err := s.db.Query(`
		SELECT id, youtube_url, title, podcast, status, error_message, created_at, analyzed_at
		FROM sources ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Source
	for rows.Next() {
		var src Source
		var analyzed sql.NullString
		if err := rows.Scan(&src.ID, &src.YouTubeURL, &src.Title, &src.Podcast, &src.Status, &src.ErrorMessage, &src.CreatedAt, &analyzed); err != nil {
			return nil, err
		}
		if analyzed.Valid {
			src.AnalyzedAt = &analyzed.String
		}
		out = append(out, src)
	}
	return out, rows.Err()
}

func (s *Store) SetSourceStatus(id, status, errMsg string) error {
	_, err := s.db.Exec(`UPDATE sources SET status = ?, error_message = ? WHERE id = ?`, status, errMsg, id)
	return err
}

func (s *Store) MarkSourceAnalyzed(id string) error {
	_, err := s.db.Exec(`UPDATE sources SET status = 'analyzed', analyzed_at = datetime('now'), error_message = '' WHERE id = ?`, id)
	return err
}

func (s *Store) InsertCandidates(sourceID string, items []AnalysisCandidate) ([]Candidate, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM candidates WHERE source_id = ?`, sourceID); err != nil {
		return nil, err
	}

	src, err := s.GetSource(sourceID)
	if err != nil {
		return nil, err
	}
	mentions, err := s.GetActiveMentions()
	if err != nil {
		return nil, err
	}
	dict := ParseMentionDictionary(mentions.Content)

	var out []Candidate
	for i, item := range items {
		id := fmt.Sprintf("%s-c%02d", sourceID, i+1)
		postText := EnsurePostTextAttribution(item.PostText, src.Podcast, dict)
		_, err := tx.Exec(`
			INSERT INTO candidates (
				id, source_id, rank, start_time, end_time, hook, take, post_text,
				why_interesting, confidence, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'proposed')`,
			id, sourceID, i+1, item.StartTime, item.EndTime, item.Hook, item.Take, postText,
			item.WhyInteresting, item.Confidence)
		if err != nil {
			return nil, err
		}
		c, err := scanCandidateTx(tx, id)
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

func scanCandidateTx(q sqlExecutor, id string) (Candidate, error) {
	var c Candidate
	var scheduled sql.NullString
	err := q.QueryRow(`
		SELECT c.id, c.source_id, c.rank, c.start_time, c.end_time, c.hook, c.take, c.post_text,
		       c.why_interesting, c.confidence, c.clip_path, c.status, c.created_at, c.updated_at,
		       s.scheduled_at
		FROM candidates c
		LEFT JOIN scheduled_posts s ON s.candidate_id = c.id
		WHERE c.id = ?`, id).Scan(
		&c.ID, &c.SourceID, &c.Rank, &c.StartTime, &c.EndTime, &c.Hook, &c.Take, &c.PostText,
		&c.WhyInteresting, &c.Confidence, &c.ClipPath, &c.Status, &c.CreatedAt, &c.UpdatedAt, &scheduled)
	if err != nil {
		return c, err
	}
	if scheduled.Valid {
		c.ScheduledAt = &scheduled.String
	}
	return c, nil
}

type sqlExecutor interface {
	QueryRow(query string, args ...any) *sql.Row
}

func (s *Store) GetCandidate(id string) (Candidate, error) {
	return scanCandidateTx(s.db, id)
}

func (s *Store) ListCandidates(sourceID string) ([]Candidate, error) {
	query := `
		SELECT c.id, c.source_id, c.rank, c.start_time, c.end_time, c.hook, c.take, c.post_text,
		       c.why_interesting, c.confidence, c.clip_path, c.status, c.created_at, c.updated_at,
		       s.scheduled_at
		FROM candidates c
		LEFT JOIN scheduled_posts s ON s.candidate_id = c.id`
	args := []any{}
	if sourceID != "" {
		query += ` WHERE c.source_id = ?`
		args = append(args, sourceID)
	}
	query += ` ORDER BY c.source_id DESC, c.rank ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Candidate
	for rows.Next() {
		var c Candidate
		var scheduled sql.NullString
		if err := rows.Scan(
			&c.ID, &c.SourceID, &c.Rank, &c.StartTime, &c.EndTime, &c.Hook, &c.Take, &c.PostText,
			&c.WhyInteresting, &c.Confidence, &c.ClipPath, &c.Status, &c.CreatedAt, &c.UpdatedAt, &scheduled); err != nil {
			return nil, err
		}
		if scheduled.Valid {
			c.ScheduledAt = &scheduled.String
		}
		enriched, err := s.enrichCandidatePostText(c)
		if err != nil {
			return nil, err
		}
		out = append(out, enriched)
	}
	return out, rows.Err()
}

func (s *Store) UpdateCandidate(id string, hook, take, postText, status string) (Candidate, error) {
	_, err := s.db.Exec(`
		UPDATE candidates SET
			hook = COALESCE(NULLIF(?, ''), hook),
			take = COALESCE(NULLIF(?, ''), take),
			post_text = COALESCE(NULLIF(?, ''), post_text),
			status = COALESCE(NULLIF(?, ''), status),
			updated_at = datetime('now')
		WHERE id = ?`, hook, take, postText, status, id)
	if err != nil {
		return Candidate{}, err
	}
	c, err := s.GetCandidate(id)
	if err != nil {
		return Candidate{}, err
	}
	return s.enrichCandidatePostText(c)
}

func (s *Store) GetCandidateEnriched(id string) (Candidate, error) {
	c, err := s.GetCandidate(id)
	if err != nil {
		return c, err
	}
	return s.enrichCandidatePostText(c)
}

func (s *Store) DeleteCandidate(id string) error {
	res, err := s.db.Exec(`DELETE FROM candidates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("candidate not found")
	}
	return nil
}

func (s *Store) SetCandidateClip(id, clipPath string) error {
	_, err := s.db.Exec(`
		UPDATE candidates SET clip_path = ?, status = 'clipped', updated_at = datetime('now')
		WHERE id = ?`, clipPath, id)
	return err
}

func (s *Store) ScheduleCandidate(candidateID string, at time.Time) (ScheduledPost, error) {
	id := fmt.Sprintf("sched-%s", candidateID)
	atStr := at.UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT INTO scheduled_posts (id, candidate_id, scheduled_at, status)
		VALUES (?, ?, ?, 'pending')
		ON CONFLICT(candidate_id) DO UPDATE SET scheduled_at = excluded.scheduled_at, status = 'pending', prepared_at = NULL`,
		id, candidateID, atStr)
	if err != nil {
		return ScheduledPost{}, err
	}
	_, err = s.db.Exec(`UPDATE candidates SET status = 'scheduled', updated_at = datetime('now') WHERE id = ?`, candidateID)
	if err != nil {
		return ScheduledPost{}, err
	}
	return s.GetScheduled(id)
}

func (s *Store) GetScheduled(id string) (ScheduledPost, error) {
	var sp ScheduledPost
	var prepared sql.NullString
	err := s.db.QueryRow(`
		SELECT id, candidate_id, scheduled_at, status, prepared_at, created_at
		FROM scheduled_posts WHERE id = ?`, id).Scan(
		&sp.ID, &sp.CandidateID, &sp.ScheduledAt, &sp.Status, &prepared, &sp.CreatedAt)
	if err != nil {
		return sp, err
	}
	if prepared.Valid {
		sp.PreparedAt = &prepared.String
	}
	c, err := s.GetCandidate(sp.CandidateID)
	if err == nil {
		sp.Candidate = &c
	}
	return sp, nil
}

func (s *Store) ListScheduled() ([]ScheduledPost, error) {
	rows, err := s.db.Query(`
		SELECT id, candidate_id, scheduled_at, status, prepared_at, created_at
		FROM scheduled_posts ORDER BY scheduled_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScheduledPost
	for rows.Next() {
		var sp ScheduledPost
		var prepared sql.NullString
		if err := rows.Scan(&sp.ID, &sp.CandidateID, &sp.ScheduledAt, &sp.Status, &prepared, &sp.CreatedAt); err != nil {
			return nil, err
		}
		if prepared.Valid {
			sp.PreparedAt = &prepared.String
		}
		c, err := s.GetCandidate(sp.CandidateID)
		if err == nil {
			sp.Candidate = &c
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *Store) ListDueScheduled(now time.Time) ([]ScheduledPost, error) {
	rows, err := s.db.Query(`
		SELECT id, candidate_id, scheduled_at, status, prepared_at, created_at
		FROM scheduled_posts
		WHERE status = 'pending' AND scheduled_at <= ?
		ORDER BY scheduled_at ASC`, now.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScheduledPost
	for rows.Next() {
		var sp ScheduledPost
		var prepared sql.NullString
		if err := rows.Scan(&sp.ID, &sp.CandidateID, &sp.ScheduledAt, &sp.Status, &prepared, &sp.CreatedAt); err != nil {
			return nil, err
		}
		if prepared.Valid {
			sp.PreparedAt = &prepared.String
		}
		c, err := s.GetCandidate(sp.CandidateID)
		if err == nil {
			sp.Candidate = &c
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *Store) MarkScheduledPrepared(id string) error {
	_, err := s.db.Exec(`
		UPDATE scheduled_posts SET status = 'ready', prepared_at = datetime('now') WHERE id = ?`, id)
	return err
}