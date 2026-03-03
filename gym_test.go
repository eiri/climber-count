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
	os.Remove(dbPath)
	defer os.Remove(dbPath)

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
	os.Remove(dbPath)
	defer os.Remove(dbPath)

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
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Out without a prior In should fail
	if _, err := g.Out(); err == nil {
		t.Fatal("expected error when calling Out without prior In, got nil")
	}
}

func TestGymInAndOut(t *testing.T) {
	dbPath := "test_gym.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on In: %v", err)
	}

	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on Out: %v", err)
	}

	actions, err := readAllActions(g.db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 2 || actions[0][1] != "in" || actions[1][1] != "out" {
		t.Errorf("expected 'in' and 'out' actions but got %v", actions)
	}
}

func TestGymIn_BlocksSecondInSameDay(t *testing.T) {
	dbPath := "test_gym_double_in.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on first In: %v", err)
	}

	// Second In on same day (without Out) should be rejected
	if err := g.In(); err == nil {
		t.Fatal("expected error on second In same day without Out, got nil")
	}
}

func TestGymIn_AllowsAfterOut(t *testing.T) {
	dbPath := "test_gym_in_after_out.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on first In: %v", err)
	}

	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on Out: %v", err)
	}

	// In again after Out should be allowed
	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on second In after Out: %v", err)
	}

	actions, err := readAllActions(g.db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 3 || actions[0][1] != "in" || actions[1][1] != "out" || actions[2][1] != "in" {
		t.Errorf("expected in/out/in sequence but got %v", actions)
	}
}

func TestGymOut_BlocksDoubleOut(t *testing.T) {
	dbPath := "test_gym_double_out.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on In: %v", err)
	}

	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on first Out: %v", err)
	}

	// Second Out without a new In should fail
	if _, err := g.Out(); err == nil {
		t.Fatal("expected error on second Out without In, got nil")
	}
}

func TestOut_TimeDelta(t *testing.T) {
	dbPath := "test_gym_time_delta.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	g, err := NewGym(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Insert an "in" row with a known past timestamp directly, bypassing In()
	inTime := time.Now().Add(-2 * time.Second).Format(time.RFC3339)
	_, err = g.db.Exec("INSERT INTO gym (timestamp, action) VALUES (?, ?)", inTime, "in")
	if err != nil {
		t.Fatalf("unexpected error seeding 'in' row: %v", err)
	}

	timeDelta, err := g.Out()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if timeDelta != "2 seconds ago" {
		t.Errorf("expected '2 seconds ago' but got %s", timeDelta)
	}
}
