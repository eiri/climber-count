package main

import (
	"database/sql"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Helper function to read all records from the SQLite database
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
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if st == nil {
		t.Fatalf("expected non-nil Storage instance")
	}
}

func TestStore(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
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
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
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
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
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
		t.Fatalf("expected last counter to be found")
	}

	expectedCounter := counter2

	if !reflect.DeepEqual(lastCounter, expectedCounter) {
		t.Errorf("expected last counter %v but got %v", expectedCounter, lastCounter)
	}
}

func TestLast_EmptyStorage(t *testing.T) {
	dbPath := "test_empty_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := st.Last()
	if ok {
		t.Fatalf("expected no last counter in empty storage")
	}
}

func TestSetCallback(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	callbackInvoked := false
	st.SetCallback(func(c Counter) bool {
		callbackInvoked = true
		return true // Indicate that the callback should be removed after being called
	})

	counter := Counter{
		Count:    1,
		Capacity: 100,
		LastUpdate: LastUpdate{
			Time: time.Now(),
		},
	}

	if err := st.Store(counter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !callbackInvoked {
		t.Fatalf("expected callback to be invoked")
	}

	if st.callback != nil {
		t.Fatalf("expected callback to be removed after being called")
	}
}

func TestCallbackRemovedAfterCalledOnce(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	callCount := 0
	st.SetCallback(func(c Counter) bool {
		callCount++
		return true // Indicate that the callback should be removed after being called
	})

	counter := Counter{
		Count:    1,
		Capacity: 100,
		LastUpdate: LastUpdate{
			Time: time.Now(),
		},
	}

	if err := st.Store(counter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected callback to be called once, but it was called %d times", callCount)
	}

	if st.callback != nil {
		t.Fatalf("expected callback to be removed after being called once")
	}

	// Store another counter to verify callback is not called again
	counter.Count = 2
	if err := st.Store(counter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected callback to not be called again, but it was called %d times", callCount)
	}
}

func TestRemoveCallback(t *testing.T) {
	dbPath := "test_storage.sqlite"
	os.Remove(dbPath)       // Ensure the file does not exist before testing
	defer os.Remove(dbPath) // Clean up after test

	st, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	callbackInvoked := false
	st.SetCallback(func(c Counter) bool {
		callbackInvoked = true
		return true
	})

	st.RemoveCallback()

	counter := Counter{
		Count:    1,
		Capacity: 100,
		LastUpdate: LastUpdate{
			Time: time.Now(),
		},
	}

	if err := st.Store(counter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callbackInvoked {
		t.Fatalf("expected callback to not be invoked after removal")
	}
}
