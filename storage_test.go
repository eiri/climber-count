package main

import (
	"database/sql"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func readAllRecords(db *sql.DB) ([][]string, error) {
	rows, err := db.Query("SELECT count, capacity, last_update FROM count")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records [][]string
	for rows.Next() {
		var count, capacity int
		var lastUpdate string
		if err := rows.Scan(&count, &capacity, &lastUpdate); err != nil {
			return nil, err
		}
		records = append(records, []string{strconv.Itoa(count), strconv.Itoa(capacity), lastUpdate})
	}
	return records, nil
}

func TestNewStorage(t *testing.T) {
	dir := t.TempDir()
	st, err := NewStorage(dir, "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil {
		t.Fatal("expected non-nil Storage instance")
	}
	want := filepath.Join(dir, "tst.db")
	if st.filePath != want {
		t.Errorf("expected filePath %q, got %q", want, st.filePath)
	}
}

func TestNewStorage_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "sub", "storage")
	st, err := NewStorage(dir, "TSB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil {
		t.Fatal("expected non-nil Storage instance")
	}
	want := filepath.Join(dir, "tsb.db")
	if st.filePath != want {
		t.Errorf("expected filePath %q, got %q", want, st.filePath)
	}
}

func TestNewStorage_GymNameLowercased(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		gym  string
		want string
	}{
		{"SLB", "slb.db"},
		{"sbg", "sbg.db"},
		{"MiXeD", "mixed.db"},
	}
	for _, tc := range cases {
		st, err := NewStorage(dir, tc.gym)
		if err != nil {
			t.Fatalf("gym %q: unexpected error: %v", tc.gym, err)
		}
		if got := filepath.Base(st.filePath); got != tc.want {
			t.Errorf("gym %q: expected file %q, got %q", tc.gym, tc.want, got)
		}
	}
}

func TestStore(t *testing.T) {
	st, err := NewStorage(t.TempDir(), "TST")
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

	records, err := readAllRecords(st.db)
	if err != nil {
		t.Fatalf("unexpected error reading records: %v", err)
	}

	expectedRecords := [][]string{
		{"1", "100", "2024-05-30T10:00:00Z"},
	}
	if !reflect.DeepEqual(records, expectedRecords) {
		t.Errorf("expected records %v but got %v", expectedRecords, records)
	}
}

func TestStore_Append(t *testing.T) {
	st, err := NewStorage(t.TempDir(), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counters := []Counter{
		{
			Count:    1,
			Capacity: 100,
			LastUpdate: LastUpdate{
				Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
			},
		},
		{
			Count:    2,
			Capacity: 200,
			LastUpdate: LastUpdate{
				Time: time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, counter := range counters {
		if err := st.Store(counter); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	records, err := readAllRecords(st.db)
	if err != nil {
		t.Fatalf("unexpected error reading records: %v", err)
	}

	expectedRecords := [][]string{
		{"1", "100", "2024-05-30T10:00:00Z"},
		{"2", "200", "2024-06-01T10:00:00Z"},
	}
	if !reflect.DeepEqual(records, expectedRecords) {
		t.Errorf("expected records %v but got %v", expectedRecords, records)
	}
}

func TestLast(t *testing.T) {
	st, err := NewStorage(t.TempDir(), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counter1 := Counter{
		Count:    1,
		Capacity: 100,
		LastUpdate: LastUpdate{
			Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
		},
	}
	counter2 := Counter{
		Count:    2,
		Capacity: 200,
		LastUpdate: LastUpdate{
			Time: time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC),
		},
	}

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
	st, err := NewStorage(t.TempDir(), "TST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := st.Last()
	if ok {
		t.Fatal("expected no last counter in empty storage")
	}
}

func TestGetGym_BeforeNewGym(t *testing.T) {
	st, err := NewStorage(t.TempDir(), "TST")
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	if g := st.GetGym(); g != nil {
		t.Errorf("expected nil Gym before NewGym, got %v", g)
	}
}

func TestGetGym_AfterNewGym(t *testing.T) {
	st, err := NewStorage(t.TempDir(), "TST")
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

// TestMultipleGyms verifies that two gyms stored under the same dir
// produce separate files and independent data.
func TestMultipleGyms(t *testing.T) {
	dir := t.TempDir()

	stSLB, err := NewStorage(dir, "SLB")
	if err != nil {
		t.Fatalf("SLB storage: %v", err)
	}
	stSBG, err := NewStorage(dir, "SBG")
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

	// Confirm two separate files exist.
	if stSLB.filePath == stSBG.filePath {
		t.Error("SLB and SBG share the same file path")
	}
}
