package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDBReopen(t *testing.T) {
	rawURL := testDBURL(t)
	db, err := OpenDB(rawURL)
	if err != nil {
		t.Fatal(err)
	}

	// Seed both tables before reopening to check schema initialization preserves data.
	for _, query := range []string{
		`INSERT INTO counter (gym, count, capacity, last_update) VALUES ('TST', 7, 20, '2024-05-30T10:00:00Z')`,
		`INSERT INTO visit (gym, timestamp, action) VALUES ('TST', '2024-05-30T10:00:00Z', 'in')`,
	} {
		if _, err := db.Exec(query); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				t.Error(closeErr)
			}
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = OpenDB(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	st, err := NewStorage(db, "TST")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := st.Last(); !ok || got.Count != 7 {
		t.Fatalf("counter after reopen: %+v, found=%v", got, ok)
	}
	if err := st.NewGym(); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetGym().Out(); err != nil {
		t.Fatalf("visit after reopen: %v", err)
	}

	counter := Counter{Count: 9, Capacity: 20, LastUpdate: LastUpdate{Time: time.Now()}}
	if err := st.Store(counter); err != nil {
		t.Fatal(err)
	}
	if got, ok := st.Last(); !ok || got.Count != 9 {
		t.Fatalf("counter after append: %+v, found=%v", got, ok)
	}
}

func TestStoreDuplicate(t *testing.T) {
	db := newTestDB(t)
	st, err := NewStorage(db, "TST")
	if err != nil {
		t.Fatal(err)
	}
	counter := Counter{Count: 7, Capacity: 20, LastUpdate: LastUpdate{Time: time.Now().Truncate(time.Second)}}
	for range 2 {
		if err := st.Store(counter); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := readAllRecords(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one counter, got %d", len(rows))
	}
}

func TestPostgresSchemes(t *testing.T) {
	if !strings.HasPrefix(os.Getenv("TEST_POSTGRES_URL"), "postgres") {
		t.Skip("TEST_POSTGRES_URL is not set")
	}
	for _, scheme := range []string{"postgres", "postgresql"} {
		t.Run(scheme, func(t *testing.T) {
			rawURL := testDBURL(t)
			rawURL = scheme + rawURL[strings.Index(rawURL, ":"):]
			db, err := OpenDB(rawURL)
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
