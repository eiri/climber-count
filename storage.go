package main

import (
	"database/sql"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

// Storer interface with a single Store method
type Storer interface {
	Store(counter Counter) error
	Last() (Counter, bool)
}

// Storage struct with the path to the storage file
type Storage struct {
	filePath string
	db       *sql.DB
}

// NewStorage creates a new Storage instance with the given file path
func NewStorage(filePath string) (*Storage, error) {
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS count (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        count INTEGER,
        capacity INTEGER,
        last_update TEXT
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db, filePath: filePath}, nil
}

// Store stores the given counter in the storage table
func (s *Storage) Store(counter Counter) error {
	logger := slog.Default().With("component", "storage")

	// Check for deduplication
	var lastUpdate string
	query := "SELECT last_update FROM count ORDER BY id DESC LIMIT 1"
	err := s.db.QueryRow(query).Scan(&lastUpdate)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if lastUpdate != "" {
		lastTime, err := time.Parse(time.RFC3339, lastUpdate)
		if err != nil {
			return err
		}
		if counter.LastUpdate.Equal(lastTime) {
			// Last update is the same, do nothing
			logger.Info("skipping duplicated counter", "counter", counter)
			return nil
		}
	}

	// Insert new record
	insertQuery := `
    INSERT INTO count (count, capacity, last_update)
    VALUES (?, ?, ?)`
	_, err = s.db.Exec(insertQuery, counter.Count, counter.Capacity, counter.LastUpdate.Format(time.RFC3339))
	if err != nil {
		return err
	}

	logger.Info("storing record", "counter", counter)
	return nil
}

// Last returns the last stored Counter and a boolean indicating if it was successful
func (s *Storage) Last() (Counter, bool) {
	logger := slog.Default().With("component", "storage", "function", "last")
	var counter Counter

	query := "SELECT count, capacity, last_update FROM count ORDER BY id DESC LIMIT 1"
	row := s.db.QueryRow(query)
	var lastUpdate string
	if err := row.Scan(&counter.Count, &counter.Capacity, &lastUpdate); err != nil {
		if err == sql.ErrNoRows {
			logger.Info("no records in table")
			return Counter{}, false
		}
		logger.Error("can't read from table", "msg", err)
		return Counter{}, false
	}

	parsedTime, err := time.Parse(time.RFC3339, lastUpdate)
	if err != nil {
		logger.Error("invalid time format", "msg", err)
		return Counter{}, false
	}

	counter.LastUpdate = LastUpdate{Time: parsedTime}
	logger.Info("found last record", "counter", counter)
	return counter, true
}
