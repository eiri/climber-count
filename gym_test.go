package main

import (
	"os"
	"testing"
	"time"

	"github.com/cvilsmeier/sqinn-go/v2"
)

// Helper function to read all records from the SQLite database
func readAllActions(sq *sqinn.Sqinn) ([][]string, error) {
	rows, err := sq.QueryRows(
		"SELECT timestamp, action FROM gym",
		nil,
		[]byte{sqinn.ValString, sqinn.ValString},
	)
	if err != nil {
		return nil, err
	}

	var actions [][]string
	for _, row := range rows {
		actions = append(actions, []string{row[0].String, row[1].String})
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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

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
	defer g.Close()

	// Insert an "in" row with a known past timestamp directly, bypassing In()
	inTime := time.Now().Add(-2 * time.Second).Format(time.RFC3339)
	err = g.db.ExecParams(
		"INSERT INTO gym (timestamp, action) VALUES (?, ?)",
		1, 2,
		[]sqinn.Value{
			sqinn.StringValue(inTime),
			sqinn.StringValue("in"),
		},
	)
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
