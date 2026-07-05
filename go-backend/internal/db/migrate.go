package db

// migrateV2 drops hook/take from candidates and title from sources.
func (p *Provider) migrateV2() error {
	var hookCol int
	_ = p.DB.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('candidates') WHERE name = 'hook'`,
	).Scan(&hookCol)
	if hookCol == 0 {
		return nil
	}

	tx, err := p.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM scheduled_posts`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		CREATE TABLE candidates_new (
		  id TEXT PRIMARY KEY,
		  source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
		  rank INTEGER NOT NULL DEFAULT 0,
		  start_time TEXT NOT NULL,
		  end_time TEXT NOT NULL,
		  post_text TEXT NOT NULL DEFAULT '',
		  why_interesting TEXT NOT NULL DEFAULT '',
		  confidence REAL NOT NULL DEFAULT 0,
		  clip_path TEXT NOT NULL DEFAULT '',
		  status TEXT NOT NULL DEFAULT 'proposed',
		  created_at TEXT NOT NULL DEFAULT (datetime('now')),
		  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO candidates_new (
		  id, source_id, rank, start_time, end_time, post_text,
		  why_interesting, confidence, clip_path, status, created_at, updated_at
		)
		SELECT id, source_id, rank, start_time, end_time,
		       CASE
		         WHEN TRIM(post_text) != '' THEN post_text
		         WHEN TRIM(hook) != '' AND TRIM(take) != '' THEN hook || char(10) || char(10) || take
		         WHEN TRIM(hook) != '' THEN hook
		         ELSE take
		       END,
		       why_interesting, confidence, clip_path, status, created_at, updated_at
		FROM candidates`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE candidates`); err != nil {
		return err
	}
	if _, err := tx.Exec(`ALTER TABLE candidates_new RENAME TO candidates`); err != nil {
		return err
	}
	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_candidates_source ON candidates(source_id)`); err != nil {
		return err
	}

	var titleCol int
	_ = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('sources') WHERE name = 'title'`).Scan(&titleCol)
	if titleCol > 0 {
		if _, err := tx.Exec(`
			CREATE TABLE sources_new (
			  id TEXT PRIMARY KEY,
			  youtube_url TEXT NOT NULL,
			  podcast TEXT NOT NULL DEFAULT '',
			  status TEXT NOT NULL DEFAULT 'ingested',
			  error_message TEXT NOT NULL DEFAULT '',
			  created_at TEXT NOT NULL DEFAULT (datetime('now')),
			  analyzed_at TEXT
			)`); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO sources_new (id, youtube_url, podcast, status, error_message, created_at, analyzed_at)
			SELECT id, youtube_url, podcast, status, error_message, created_at, analyzed_at FROM sources`); err != nil {
			return err
		}
		if _, err := tx.Exec(`DROP TABLE sources`); err != nil {
			return err
		}
		if _, err := tx.Exec(`ALTER TABLE sources_new RENAME TO sources`); err != nil {
			return err
		}
	}

	return tx.Commit()
}