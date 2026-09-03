package main

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func readAllActions(db *sql.DB) ([][]string, error) {
	rows, err := db.Query("SELECT gym, timestamp, action FROM visit ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions [][]string
	for rows.Next() {
		var gym, timestamp, action string
		if err := rows.Scan(&gym, &timestamp, &action); err != nil {
			return nil, err
		}
		actions = append(actions, []string{gym, timestamp, action})
	}
	return actions, nil
}

func TestNewGym(t *testing.T) {
	g, err := NewGym(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatalf("expected non-nil Gym instance")
	}
	if g.gym != "TST" {
		t.Errorf("expected TST, got %q", g.gym)
	}
}

func TestGymIn(t *testing.T) {
	db := newTestDB(t)
	g, err := NewGym(db, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions, err := readAllActions(db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 1 || actions[0][0] != "TST" || actions[0][2] != "in" {
		t.Errorf("expected TST in action but got %v", actions)
	}
}

func TestGymOut(t *testing.T) {
	g, err := NewGym(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := g.Out(); err == nil {
		t.Fatal("expected error when calling Out without prior In, got nil")
	}
}

func TestGymInAndOut(t *testing.T) {
	db := newTestDB(t)
	g, err := NewGym(db, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on In: %v", err)
	}
	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on Out: %v", err)
	}

	actions, err := readAllActions(db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 2 || actions[0][2] != "in" || actions[1][2] != "out" {
		t.Errorf("expected 'in' and 'out' actions but got %v", actions)
	}
}

func TestGymIn_BlocksSecondInSameDay(t *testing.T) {
	g, err := NewGym(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on first In: %v", err)
	}
	if err := g.In(); err == nil {
		t.Fatal("expected error on second In same day without Out, got nil")
	}
}

func TestGymIn_AllowsAfterOut(t *testing.T) {
	db := newTestDB(t)
	g, err := NewGym(db, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on first In: %v", err)
	}
	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on Out: %v", err)
	}
	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on second In after Out: %v", err)
	}

	actions, err := readAllActions(db)
	if err != nil {
		t.Fatalf("unexpected error reading actions: %v", err)
	}

	if len(actions) != 3 || actions[0][2] != "in" || actions[1][2] != "out" || actions[2][2] != "in" {
		t.Errorf("expected in/out/in sequence but got %v", actions)
	}
}

func TestGymOut_BlocksDoubleOut(t *testing.T) {
	g, err := NewGym(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := g.In(); err != nil {
		t.Fatalf("unexpected error on In: %v", err)
	}
	if _, err := g.Out(); err != nil {
		t.Fatalf("unexpected error on first Out: %v", err)
	}
	if _, err := g.Out(); err == nil {
		t.Fatal("expected error on second Out without In, got nil")
	}
}

func TestOut_EntryTime(t *testing.T) {
	db := newTestDB(t)
	g, err := NewGym(db, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inTime := time.Now().Add(-2 * time.Second).Format(time.RFC3339)
	_, err = db.Exec("INSERT INTO visit (gym, timestamp, action) VALUES (?, ?, ?)", "TST", inTime, "in")
	if err != nil {
		t.Fatalf("unexpected error seeding 'in' row: %v", err)
	}

	entry, err := g.Out()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Format(time.RFC3339) != inTime {
		t.Errorf("expected %q but got %q", inTime, entry.Format(time.RFC3339))
	}
}

func TestGymIsolation(t *testing.T) {
	db := newTestDB(t)
	slb, err := NewGym(db, "SLB")
	if err != nil {
		t.Fatalf("SLB: %v", err)
	}
	sbg, err := NewGym(db, "SBG")
	if err != nil {
		t.Fatalf("SBG: %v", err)
	}

	if err := slb.In(); err != nil {
		t.Fatalf("SLB in: %v", err)
	}
	if _, err := sbg.Out(); err == nil {
		t.Fatal("expected SBG out to ignore SLB check-in")
	}
}
