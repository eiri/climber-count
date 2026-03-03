package main

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cvilsmeier/sqinn-go/v2"
)

// Helper function to read all records from the SQLite database
func readAllRecords(sq *sqinn.Sqinn) ([][]string, error) {
	rows, err := sq.QueryRows(
		"SELECT count, capacity, last_update FROM count",
		nil,
		[]byte{sqinn.ValInt32, sqinn.ValInt32, sqinn.ValString},
	)
	if err != nil {
		return nil, err
	}

	var records [][]string
	for _, row := range rows {
		records = append(records, []string{
			strconv.Itoa(int(row[0].Int32)),
			strconv.Itoa(int(row[1].Int32)),
			row[2].String,
		})
	}
	return records, nil
}

func TestNewStorage(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

	if st == nil {
		t.Fatalf("expected non-nil Storage instance")
	}
}

func TestStore(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

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
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

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
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

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
		t.Fatalf("expected last counter to be found")
	}
	expectedCounter := counter2

	if !reflect.DeepEqual(lastCounter, expectedCounter) {
		t.Errorf("expected last counter %v but got %v", expectedCounter, lastCounter)
	}
}

func TestLast_EmptyStorage(t *testing.T) {
	dbPath := "test_empty_storage.sqlite"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

	_, ok := st.Last()
	if ok {
		t.Fatalf("expected no last counter in empty storage")
	}
}
