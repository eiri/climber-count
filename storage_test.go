package main

import (
	"database/sql"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	return db
}

func readAllRecords(db *sql.DB) ([][]string, error) {
	rows, err := db.Query("SELECT gym, count, capacity, last_update FROM counter ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records [][]string
	for rows.Next() {
		var gym, lastUpdate string
		var count, capacity int
		if err := rows.Scan(&gym, &count, &capacity, &lastUpdate); err != nil {
			return nil, err
		}
		records = append(records, []string{gym, strconv.Itoa(count), strconv.Itoa(capacity), lastUpdate})
	}
	return records, nil
}

func TestNewStorage(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil {
		t.Fatal("expected non-nil Storage instance")
	}
	if st.gym != "TST" {
		t.Errorf("expected gym TST, got %q", st.gym)
	}
}

func TestNewStorage_GymNameUppercased(t *testing.T) {
	cases := []struct {
		gym  string
		want string
	}{
		{"SLB", "SLB"},
		{"sbg", "SBG"},
		{"MiXeD", "MIXED"},
	}
	for _, tc := range cases {
		st, err := NewStorage(newTestDB(t), tc.gym)
		if err != nil {
			t.Fatalf("gym %q: unexpected error: %v", tc.gym, err)
		}
		if st.gym != tc.want {
			t.Errorf("gym %q: expected %q, got %q", tc.gym, tc.want, st.gym)
		}
	}
}

func TestStore(t *testing.T) {
	db := newTestDB(t)
	st, err := NewStorage(db, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counter := Counter{
		Count:    1,
		Capacity: 100,
		LastUpdate: LastUpdate{
			Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
		},
	}
	if err := st.Store(counter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, err := readAllRecords(db)
	if err != nil {
		t.Fatalf("unexpected error reading records: %v", err)
	}

	expectedRecords := [][]string{{"TST", "1", "100", "2024-05-30T10:00:00Z"}}
	if !reflect.DeepEqual(records, expectedRecords) {
		t.Errorf("expected records %v but got %v", expectedRecords, records)
	}
}

func TestStore_Append(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counters := []Counter{
		{Count: 1, Capacity: 100, LastUpdate: LastUpdate{Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC)}},
		{Count: 2, Capacity: 200, LastUpdate: LastUpdate{Time: time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)}},
	}
	for _, counter := range counters {
		if err := st.Store(counter); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	lastCounter, ok := st.Last()
	if !ok {
		t.Fatal("expected last counter to be found")
	}
	if !reflect.DeepEqual(lastCounter, counters[1]) {
		t.Errorf("expected last counter %v but got %v", counters[1], lastCounter)
	}
}

func TestLast(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counter1 := Counter{Count: 1, Capacity: 100, LastUpdate: LastUpdate{Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC)}}
	counter2 := Counter{Count: 2, Capacity: 200, LastUpdate: LastUpdate{Time: time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)}}

	if err := st.Store(counter1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := st.Store(counter2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lastCounter, ok := st.Last()
	if !ok {
		t.Fatal("expected last counter to be found")
	}
	if !reflect.DeepEqual(lastCounter, counter2) {
		t.Errorf("expected last counter %v but got %v", counter2, lastCounter)
	}
}

func TestLast_EmptyStorage(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := st.Last()
	if ok {
		t.Fatal("expected no last counter in empty storage")
	}
}

func TestGetGym_BeforeNewGym(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	if g := st.GetGym(); g != nil {
		t.Errorf("expected nil Gym before NewGym, got %v", g)
	}
}

func TestGetGym_AfterNewGym(t *testing.T) {
	st, err := NewStorage(newTestDB(t), "TST")
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	if err := st.NewGym(); err != nil {
		t.Fatalf("NewGym: %v", err)
	}
	if g := st.GetGym(); g == nil {
		t.Error("expected non-nil Gym after NewGym")
	}
}

func TestMultipleGyms(t *testing.T) {
	db := newTestDB(t)

	stSLB, err := NewStorage(db, "SLB")
	if err != nil {
		t.Fatalf("SLB storage: %v", err)
	}
	stSBG, err := NewStorage(db, "SBG")
	if err != nil {
		t.Fatalf("SBG storage: %v", err)
	}

	cSLB := Counter{Count: 10, Capacity: 50, LastUpdate: LastUpdate{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}}
	cSBG := Counter{Count: 99, Capacity: 200, LastUpdate: LastUpdate{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}}

	if err := stSLB.Store(cSLB); err != nil {
		t.Fatalf("store SLB: %v", err)
	}
	if err := stSBG.Store(cSBG); err != nil {
		t.Fatalf("store SBG: %v", err)
	}

	gotSLB, ok := stSLB.Last()
	if !ok || gotSLB.Count != 10 {
		t.Errorf("SLB: expected count 10, got %+v (ok=%v)", gotSLB, ok)
	}
	gotSBG, ok := stSBG.Last()
	if !ok || gotSBG.Count != 99 {
		t.Errorf("SBG: expected count 99, got %+v (ok=%v)", gotSBG, ok)
	}
}
