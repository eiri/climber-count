package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Gym represents one gym's visit log.
type Gym struct {
	db  *sql.DB
	gym string
}

// NewGym creates a new Gym instance with the shared database.
func NewGym(db *sql.DB, gymName string) (*Gym, error) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS visit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gym TEXT NOT NULL,
			timestamp TEXT,
			action TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_visit_gym_timestamp ON visit(gym, timestamp, id);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("init visit schema: %w", err)
		}
	}

	return &Gym{db: db, gym: strings.ToUpper(gymName)}, nil
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
	var action, timestampStr string
	query := "SELECT action, timestamp FROM visit WHERE gym = ? ORDER BY timestamp DESC, id DESC LIMIT 1"
	err := g.db.QueryRow(query, g.gym).Scan(&action, &timestampStr)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, err
	}

	ts, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return "", time.Time{}, err
	}

	return action, ts, nil
}

func (g *Gym) writeAction(action string) error {
	timestamp := time.Now().Format(time.RFC3339)
	_, err := g.db.Exec("INSERT INTO visit (gym, timestamp, action) VALUES (?, ?, ?)", g.gym, timestamp, action)
	return err
}
