package main

import (
	"net/url"
	"path/filepath"
	"testing"
)

func TestOpenDB_SQLiteURL(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data", "storage.db")
	db, err := OpenDB("sqlite://" + dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenDB_UnsupportedScheme(t *testing.T) {
	_, err := OpenDB("postgres://localhost/db")
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestSQLitePath(t *testing.T) {
	cases := map[string]string{
		"sqlite://data/storage.db": "data/storage.db",
		"sqlite:///tmp/storage.db": "/tmp/storage.db",
	}

	for rawURL, want := range cases {
		u, err := url.Parse(rawURL)
		if err != nil {
			t.Fatalf("parse %q: %v", rawURL, err)
		}
		if got := sqlitePath(u); got != want {
			t.Errorf("%q: expected %q, got %q", rawURL, want, got)
		}
	}
}

func TestSQLiteDSN_PreservesQuery(t *testing.T) {
	u, err := url.Parse("sqlite://data/storage.db?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	want := "data/storage.db?_pragma=busy_timeout(5000)"
	if got := sqliteDSN(u); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
