package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// OpenDB opens the configured storage database.
func OpenDB(rawURL string) (*sql.DB, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse storage url: %w", err)
	}
	if u.Scheme != "sqlite" {
		return nil, fmt.Errorf("unsupported storage scheme %q", u.Scheme)
	}

	dsn := sqliteDSN(u)
	if err := os.MkdirAll(filepath.Dir(sqlitePath(u)), 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return db, nil
}

func sqliteDSN(u *url.URL) string {
	path := sqlitePath(u)
	if u.RawQuery == "" {
		return path
	}
	return path + "?" + u.RawQuery
}

func sqlitePath(u *url.URL) string {
	if u.Host == "" {
		return u.Path
	}
	return strings.TrimLeft(u.Host+u.Path, "/")
}
