package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"time"
)

type Storer interface {
	Store(counter Counter) error
}

type Storage struct {
	filePath string
}

func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

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
