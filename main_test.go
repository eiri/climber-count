package main

import (
	"io"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Suppress log output
	log.SetOutput(io.Discard)

	// Run the tests
	os.Exit(m.Run())
}
