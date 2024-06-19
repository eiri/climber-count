package main

import (
	"encoding/csv"
	"io"
	"log/slog"
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
	logger := slog.Default().With("component", "storage")
	// Check if the file exists
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Error("can't open file", "msg", err)
		return err
	}
	defer file.Close()
	logger.Info("opened file", "path", s.filePath)

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
			logger.Info("skipping duplicated counter", "counter", counter)
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

	logger.Info("storing record", "record", record)
	return writer.Write(record)
}

// Last returns the last stored Counter and a boolean indicating if it was successful
func (s *Storage) Last() (Counter, bool) {
	logger := slog.Default().With("component", "storage", "function", "last")
	file, err := os.Open(s.filePath)
	if err != nil {
		logger.Error("can't open file", "msg", err)
		return Counter{}, false
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil && err.Error() != "EOF" {
		logger.Error("can't read file", "msg", err)
		return Counter{}, false
	}

	if len(records) == 0 {
		logger.Info("no records in file")
		return Counter{}, false
	}

	lastRecord := records[len(records)-1]
	count, _ := strconv.Atoi(lastRecord[0])
	capacity, _ := strconv.Atoi(lastRecord[1])
	lastUpdate, _ := time.Parse(time.RFC3339, lastRecord[2])

	counter := Counter{
		Count:    count,
		Capacity: capacity,
		LastUpdate: LastUpdate{
			Time: lastUpdate,
		},
	}
	logger.Info("found last record", "counter", counter)
	return counter, true
}
