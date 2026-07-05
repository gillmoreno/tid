package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Provider struct {
	DB *sql.DB
}

func NewProvider(dbPath string) (*Provider, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir db dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	p := &Provider{DB: conn}
	if err := p.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return p, nil
}

func (p *Provider) migrate() error {
	if _, err := p.DB.Exec(SchemaSQL); err != nil {
		return err
	}
	return p.migrateV2()
}

func (p *Provider) Close() error {
	if p.DB == nil {
		return nil
	}
	return p.DB.Close()
}