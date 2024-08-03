package main

import (
	"database/sql"
	"time"

	"github.com/imbue11235/humanize"
	_ "modernc.org/sqlite"
)

// Gym represents the gym structure with a connection to the SQLite database.
type Gym struct {
	db *sql.DB
}

// New creates a new Gym instance with the given SQLite database path.
func NewGym(dbPath string) (*Gym, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS gym (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp TEXT,
        action TEXT
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return &Gym{db: db}, nil
}

// In writes an "in" action with the current timestamp to the database.
func (g *Gym) In() error {
	return g.writeAction("in")
}

// Out writes an "out" action with the current timestamp to the database and returns the time delta since the latest "in" action.
func (g *Gym) Out() (string, error) {
	var lastInTimestampStr string
	err := g.db.QueryRow("SELECT timestamp FROM gym WHERE action = 'in' ORDER BY timestamp DESC LIMIT 1").Scan(&lastInTimestampStr)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	var lastInTimestamp time.Time
	if lastInTimestampStr != "" {
		lastInTimestamp, err = time.Parse(time.RFC3339, lastInTimestampStr)
		if err != nil {
			return "", err
		}
	}

	err = g.writeAction("out")
	if err != nil {
		return "", err
	}

	return humanize.ExactTime(lastInTimestamp).FromNow(), nil
}

func (g *Gym) writeAction(action string) error {
	timestamp := time.Now().Format(time.RFC3339)
	_, err := g.db.Exec("INSERT INTO gym (timestamp, action) VALUES (?, ?)", timestamp, action)
	return err
}
