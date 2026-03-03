package main

import (
	"log/slog"
	"time"

	"github.com/cvilsmeier/sqinn-go/v2"
)

// Storer interface with a single Store method
type Storer interface {
	Store(counter Counter) error
	Last() (Counter, bool)
	NewGym() error
	GetGym() *Gym
}

// Storage struct with the path to the storage file
type Storage struct {
	filePath string
	db       *sqinn.Sqinn
	gym      *Gym
}

// NewStorage creates a new Storage instance with the given file path
func NewStorage(filePath string) (*Storage, error) {
	sq, err := sqinn.Launch(sqinn.Options{
		Db: filePath,
	})
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
	if err := sq.ExecSql(createTableQuery); err != nil {
		_ = sq.Close()
		return nil, err
	}

	return &Storage{db: sq, filePath: filePath}, nil
}

func (s *Storage) Close() {
	_ = s.db.Close()
}

// NewGym initializes and stores the Gym instance using the Storage's file path.
// It returns an error if the Gym instance cannot be created.
func (s *Storage) NewGym() error {
	var err error
	s.gym, err = NewGym(s.filePath)
	return err
}

// GetGym returns the Gym instance associated with the Storage object.
func (s *Storage) GetGym() *Gym {
	return s.gym
}

// Store stores the given counter in the storage table
func (s *Storage) Store(counter Counter) error {
	logger := slog.Default().With("component", "storage")

	// Check for deduplication — fetch last_update of most recent row
	rows, err := s.db.QueryRows(
		"SELECT last_update FROM count ORDER BY id DESC LIMIT 1",
		nil,
		[]byte{sqinn.ValString},
	)
	if err != nil {
		return err
	}

	if len(rows) > 0 {
		lastUpdate := rows[0][0].String
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
	}

	// Insert new record
	err = s.db.ExecParams(
		"INSERT INTO count (count, capacity, last_update) VALUES (?, ?, ?)",
		1, 3,
		[]sqinn.Value{
			sqinn.Int32Value(counter.Count),
			sqinn.Int32Value(counter.Capacity),
			sqinn.StringValue(counter.LastUpdate.Format(time.RFC3339)),
		},
	)
	if err != nil {
		return err
	}

	logger.Info("storing record", "counter", counter)
	return nil
}

// Last returns the last stored Counter and a boolean indicating if it was successful
func (s *Storage) Last() (Counter, bool) {
	logger := slog.Default().With("component", "storage", "function", "last")

	rows, err := s.db.QueryRows(
		"SELECT count, capacity, last_update FROM count ORDER BY id DESC LIMIT 1",
		nil,
		[]byte{sqinn.ValInt32, sqinn.ValInt32, sqinn.ValString},
	)
	if err != nil {
		logger.Error("can't read from table", "msg", err)
		return Counter{}, false
	}
	if len(rows) == 0 {
		logger.Info("no records in table")
		return Counter{}, false
	}

	row := rows[0]
	count := int(row[0].Int32)
	capacity := int(row[1].Int32)
	lastUpdateStr := row[2].String

	parsedTime, err := time.Parse(time.RFC3339, lastUpdateStr)
	if err != nil {
		logger.Error("invalid time format", "msg", err)
		return Counter{}, false
	}

	counter := Counter{
		Count:      count,
		Capacity:   capacity,
		LastUpdate: LastUpdate{Time: parsedTime},
	}
	logger.Info("found last record", "counter", counter)
	return counter, true
}
