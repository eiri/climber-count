package main

import (
	"net"
	"net/url"
	"path/filepath"
	"strings"
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
	_, err := OpenDB("mysql://localhost/db")
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestOpenDB_InvalidURL(t *testing.T) {
	for _, rawURL := range []string{
		"postgres://user:secret@localhost:bad/db",
		"postgres://user:secret@localhost/db?sslmode=invalid",
	} {
		db, err := OpenDB(rawURL)
		if db != nil || err == nil {
			t.Fatalf("expected nil database and an error, got %v, %v", db, err)
		}
		if strings.Contains(err.Error(), "secret") {
			t.Fatal("error exposes database password")
		}
	}
}

func TestOpenDB_Unavailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := OpenDB("postgres://user:secret@" + addr + "/db?sslmode=disable&connect_timeout=1")
	if db != nil || err == nil {
		t.Fatalf("expected nil database and an error, got %v, %v", db, err)
	}
	if !strings.Contains(err.Error(), "ping storage") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unexpected connection error: %v", err)
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
