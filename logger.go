package main

import (
	"log/slog"
	"os"
)

func SetLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
