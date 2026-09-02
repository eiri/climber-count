package main

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

// Gym represents the gym structure with a connection to the SQLite database.
type Gym struct {
	db *sql.DB
}

// NewGym creates a new Gym instance with the given SQLite database path.
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
	last, lastTs, err := g.lastAction()
	if err != nil {
		return err
	}

	if last == "in" {
		now := time.Now()
		sameDay := lastTs.Year() == now.Year() && lastTs.YearDay() == now.YearDay()
		if sameDay {
			return errors.New("cannot check in: already checked in without checking out")
		}
	}

	return g.writeAction("in")
}

// Out writes an "out" action and returns the visit start time.
func (g *Gym) Out() (time.Time, error) {
	last, lastTs, err := g.lastAction()
	if err != nil {
		return time.Time{}, err
	}
	if last != "in" {
		return time.Time{}, errors.New("cannot check out: no active check-in")
	}

	if err = g.writeAction("out"); err != nil {
		return time.Time{}, err
	}

	return lastTs, nil
}

func (g *Gym) lastAction() (string, time.Time, error) {
	_, action, ts, err := g.lastRow()
	return action, ts, err
}

func (g *Gym) lastRow() (int64, string, time.Time, error) {
	var id int64
	var action, timestampStr string
	err := g.db.QueryRow("SELECT id, action, timestamp FROM gym ORDER BY timestamp DESC, id DESC LIMIT 1").Scan(&id, &action, &timestampStr)
	if err == sql.ErrNoRows {
		return 0, "", time.Time{}, nil
	}
	if err != nil {
		return 0, "", time.Time{}, err
	}

	ts, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return 0, "", time.Time{}, err
	}

	return id, action, ts, nil
}

func (g *Gym) writeAction(action string) error {
	timestamp := time.Now().Format(time.RFC3339)
	_, err := g.db.Exec("INSERT INTO gym (timestamp, action) VALUES (?, ?)", timestamp, action)
	return err
}
