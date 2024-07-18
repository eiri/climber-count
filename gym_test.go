package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Helper function to read all records from the SQLite database
func readAllActions(db *sql.DB) ([][]string, error) {
	rows, err := db.Query("SELECT timestamp, action FROM gym")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions [][]string
	for rows.Next() {
		var timestamp, action string
		if err := rows.Scan(&timestamp, &action); err != nil {
			return nil, err
		}
		actions = append(actions, []string{timestamp, action})
	}
	return actions, nil
}

func TestNewGym(t *testing.T) {
	dbPath := "test_gym.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatalf("expected non-nil Gym instance")
	}
}

func TestGymIn(t *testing.T) {
	dbPath := "test_gym.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions, err := readAllActions(g.db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 1 || actions[0][1] != "in" {
		t.Errorf("expected 'in' action but got %v", actions)
	}
}

func TestGymOut(t *testing.T) {
	dbPath := "test_gym.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions, err := readAllActions(g.db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 1 || actions[0][1] != "out" {
		t.Errorf("expected 'out' action but got %v", actions)
	}
}

func TestGymInAndOut(t *testing.T) {
	dbPath := "test_gym.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions, err := readAllActions(g.db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 2 || actions[0][1] != "in" || actions[1][1] != "out" {
		t.Errorf("expected 'in' and 'out' actions but got %v", actions)
	}
}

func TestOut_TimeDelta(t *testing.T) {
	dbPath := "test_gym_time_delta.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Perform an "in" action
	if err := g.In(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait a bit to ensure a noticeable time delta
	time.Sleep(2 * time.Second)

	// Perform an "out" action and capture the time delta
	timeDelta, err := g.Out()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if timeDelta != "2 seconds ago" {
		t.Errorf("expected time delta around 2s but got %s", timeDelta)
	}
}
