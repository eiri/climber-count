package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Storer persists counters and visit logs.
type Storer interface {
	Store(counter Counter) error
	Last() (Counter, bool)
	NewGym() error
	GetGym() *Gym
}

// Storage stores counters for one gym in a shared DB.
type Storage struct {
	db  *sql.DB
	gym string
	log *Gym
}

// NewStorage creates a storage wrapper for the given gym.
func NewStorage(db *sql.DB, gymName string) (*Storage, error) {
	if err := initStorage(db); err != nil {
		return nil, err
	}

	return &Storage{db: db, gym: strings.ToUpper(gymName)}, nil
}

func initStorage(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS counter (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gym TEXT NOT NULL,
			count INTEGER,
			capacity INTEGER,
			last_update TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_counter_gym_id ON counter(gym, id);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("init counter schema: %w", err)
		}
	}
	return nil
}

// NewGym initializes and stores the Gym instance using the Storage's DB.
func (s *Storage) NewGym() error {
	var err error
	s.log, err = NewGym(s.db, s.gym)
	return err
}

// GetGym returns the Gym instance associated with the Storage object.
func (s *Storage) GetGym() *Gym {
	return s.log
}

// Store stores the given counter in the storage table.
func (s *Storage) Store(counter Counter) error {
	logger := slog.Default().With("component", "storage", "gym", s.gym)

	var lastUpdate string
	query := "SELECT last_update FROM counter WHERE gym = ? ORDER BY id DESC LIMIT 1"
	err := s.db.QueryRow(query, s.gym).Scan(&lastUpdate)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if lastUpdate != "" {
		lastTime, err := time.Parse(time.RFC3339, lastUpdate)
		if err != nil {
			return err
		}
		if counter.LastUpdate.Equal(lastTime) {
			logger.Info("skipping duplicated counter", "counter", counter)
			return nil
		}
	}

	insertQuery := `
	INSERT INTO counter (gym, count, capacity, last_update)
	VALUES (?, ?, ?, ?)`
	_, err = s.db.Exec(insertQuery, s.gym, counter.Count, counter.Capacity, counter.LastUpdate.Format(time.RFC3339))
	if err != nil {
		return err
	}

	logger.Info("storing record", "counter", counter)
	return nil
}

// Last returns the last stored Counter.
func (s *Storage) Last() (Counter, bool) {
	logger := slog.Default().With("component", "storage", "function", "last", "gym", s.gym)

	var counter Counter
	query := "SELECT count, capacity, last_update FROM counter WHERE gym = ? ORDER BY id DESC LIMIT 1"
	row := s.db.QueryRow(query, s.gym)

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
