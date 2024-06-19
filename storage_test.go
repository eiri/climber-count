package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

// Helper function to read all records from a CSV file
func readAllRecords(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func TestNewStorage(t *testing.T) {
	filePath := "test_storage.csv"
	st := NewStorage(filePath)

	if st == nil {
		t.Fatalf("expected non-nil Storage instance")
	}

	if st.filePath != filePath {
		t.Errorf("expected file path %q but got %q", filePath, st.filePath)
	}
}

func TestStore(t *testing.T) {
	filePath := "test_storage.csv"
	os.Remove(filePath)       // Ensure the file does not exist before testing
	defer os.Remove(filePath) // Clean up after test

	st := NewStorage(filePath)

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

	records, err := readAllRecords(filePath)
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
	filePath := "test_storage.csv"
	os.Remove(filePath)       // Ensure the file does not exist before testing
	defer os.Remove(filePath) // Clean up after test

	st := NewStorage(filePath)

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
				Time: time.Date(2024, time.May, 30, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, counter := range counters {
		if err := st.Store(counter); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	records, err := readAllRecords(filePath)
	if err != nil {
		t.Fatalf("unexpected error reading records: %v", err)
	}

	expectedRecords := [][]string{
		{"1", "100", "2024-05-30T10:00:00Z"},
		{"2", "200", "2024-05-30T11:00:00Z"},
	}

	if !reflect.DeepEqual(records, expectedRecords) {
		t.Errorf("expected records %v but got %v", expectedRecords, records)
	}
}

func TestStore_Deduplication(t *testing.T) {
	filePath := "test_storage.csv"
	os.Remove(filePath)       // Ensure the file does not exist before testing
	defer os.Remove(filePath) // Clean up after test

	st := NewStorage(filePath)

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
			Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
		},
	}

	if err := st.Store(counter1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := st.Store(counter2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, err := readAllRecords(filePath)
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

func TestLast(t *testing.T) {
	filePath := "test_storage.csv"
	os.Remove(filePath)       // Ensure the file does not exist before testing
	defer os.Remove(filePath) // Clean up after test

	st := NewStorage(filePath)

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
			Time: time.Date(2024, time.May, 30, 11, 0, 0, 0, time.UTC),
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
	filePath := "test_empty_storage.csv"
	os.Remove(filePath)       // Ensure the file does not exist before testing
	defer os.Remove(filePath) // Clean up after test

	st := NewStorage(filePath)

	_, ok := st.Last()
	if ok {
		t.Fatalf("expected no last counter in empty storage")
	}
}

func TestStorage_Rotate_FileNotExist(t *testing.T) {
	storage := &Storage{filePath: "nonexistentfile.csv"}
	err := storage.Rotate()
	if err == nil || err.Error() != "failed to open file: open nonexistentfile.csv: no such file or directory" {
		t.Errorf("expected file does not exist error, got %v", err)
	}
}

func TestStorage_Rotate_Success(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file
	tempFile, err := os.CreateTemp(tempDir, "testfile*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()

	// Initialize storage
	storage := &Storage{filePath: tempFilePath}

	// Rotate the file
	err = storage.Rotate()
	if err != nil {
		t.Fatalf("unexpected error during rotate: %v", err)
	}

	// Check if the file was renamed correctly
	currentDate := time.Now().Format("20060102")
	expectedFilePath := fmt.Sprintf("%s-%s.csv", tempFilePath[:len(tempFilePath)-4], currentDate)
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Errorf("expected file to be renamed to %s, but it does not exist", expectedFilePath)
	}
}
