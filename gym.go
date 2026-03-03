package main

import (
	"errors"
	"time"

	"github.com/cvilsmeier/sqinn-go/v2"
	"github.com/imbue11235/humanize"
)

// Gym represents the gym structure with a connection to the SQLite database.
type Gym struct {
	db *sqinn.Sqinn
}

// New creates a new Gym instance with the given SQLite database path.
func NewGym(dbPath string) (*Gym, error) {
	sq, err := sqinn.Launch(sqinn.Options{
		Db: dbPath,
	})
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
	if err := sq.ExecSql(createTableQuery); err != nil {
		_ = sq.Close()
		return nil, err
	}

	return &Gym{db: sq}, nil
}

func (g *Gym) Close() {
	_ = g.db.Close()
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

// Out writes an "out" action with the current timestamp to the database and returns the time delta since the latest "in" action.
func (g *Gym) Out() (string, error) {
	last, lastTs, err := g.lastAction()
	if err != nil {
		return "", err
	}
	if last != "in" {
		return "", errors.New("cannot check out: no active check-in")
	}

	if err = g.writeAction("out"); err != nil {
		return "", err
	}

	return humanize.ExactTime(lastTs).FromNow(), nil
}

func (g *Gym) lastAction() (string, time.Time, error) {
	rows, err := g.db.QueryRows(
		"SELECT action, timestamp FROM gym ORDER BY timestamp DESC, id DESC LIMIT 1",
		nil,
		[]byte{sqinn.ValString, sqinn.ValString},
	)
	if err != nil {
		return "", time.Time{}, err
	}
	if len(rows) == 0 {
		return "", time.Time{}, nil
	}

	action := rows[0][0].String
	timestampStr := rows[0][1].String

	ts, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return "", time.Time{}, err
	}

	return action, ts, nil
}

func (g *Gym) writeAction(action string) error {
	timestamp := time.Now().Format(time.RFC3339)
	return g.db.ExecParams(
		"INSERT INTO gym (timestamp, action) VALUES (?, ?)",
		1, 2,
		[]sqinn.Value{
			sqinn.StringValue(timestamp),
			sqinn.StringValue(action),
		},
	)
}
