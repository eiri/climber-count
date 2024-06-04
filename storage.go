package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"time"
)

// Storer interface with a single Store method
type Storer interface {
	Store(counter Counter) error
	Last() (Counter, bool)
}

// Storage struct with the path to the storage file
type Storage struct {
	filePath string
}

// NewStorage creates a new Storage instance with the given file path
func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

// Store stores the given counter in the storage file as a CSV record
func (s *Storage) Store(counter Counter) error {
	// Check if the file exists
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the existing records
	records, err := csv.NewReader(file).ReadAll()
	if err != nil && err.Error() != "EOF" {
		return err
	}

	// Check for deduplication
	if len(records) > 0 {
		lastRecord := records[len(records)-1]
		lastUpdate, err := time.Parse(time.RFC3339, lastRecord[2])
		if err != nil {
			return err
		}
		if counter.LastUpdate.Equal(lastUpdate) {
			// Last update is the same, do nothing
			return nil
		}
	}

	record := []string{
		strconv.Itoa(counter.Count),
		strconv.Itoa(counter.Capacity),
		counter.LastUpdate.Format(time.RFC3339),
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	return writer.Write(record)
}

// Last returns the last stored Counter and a boolean indicating if it was successful
func (s *Storage) Last() (Counter, bool) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return Counter{}, false
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil && err.Error() != "EOF" {
		return Counter{}, false
	}

	if len(records) == 0 {
		return Counter{}, false
	}

	lastRecord := records[len(records)-1]
	count, _ := strconv.Atoi(lastRecord[0])
	capacity, _ := strconv.Atoi(lastRecord[1])
	lastUpdate, _ := time.Parse(time.RFC3339, lastRecord[2])

	return Counter{
		Count:    count,
		Capacity: capacity,
		LastUpdate: LastUpdate{
			Time: lastUpdate,
		},
	}, true
}
