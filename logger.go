package main

import (
	"log/slog"
	"os"

	qlog "github.com/reugn/go-quartz/logger"
)

func SetLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	// turn off quarz logger
	qlog.SetDefault(qlog.NewSimpleLogger(nil, qlog.LevelOff))
}
