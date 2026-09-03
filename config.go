package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PGK         string
	FID         string
	Gym         string
	BotToken    string
	Storage     string
	MetricsAddr string
	Schedule    map[string]string
}

func NewConfig() (*Config, error) {
	cfg := Config{
		Storage:  "sqlite://data/storage.db",
		Schedule: make(map[string]string),
	}
	envVars := map[string]*string{
		"PGK":       &cfg.PGK,
		"FID":       &cfg.FID,
		"GYM":       &cfg.Gym,
		"BOT_TOKEN": &cfg.BotToken,
		"STORAGE":   &cfg.Storage,
	}

	for key, ptr := range envVars {
		val, ok := os.LookupEnv(key)
		if key != "STORAGE" && !ok {
			return &cfg, fmt.Errorf("the required env var %q is not set", key)
		}
		if ok {
			*ptr = val
		}
	}

	cfg.MetricsAddr = ":9117"
	if val, ok := os.LookupEnv("METRICS_ADDR"); ok {
		cfg.MetricsAddr = val
	}

	if val, ok := os.LookupEnv("SCHEDULE"); ok {
		for subVal := range strings.SplitSeq(val, "|") {
			if strings.Contains(subVal, "=") {
				kv := strings.SplitN(subVal, "=", 2)
				cfg.Schedule[kv[0]] = kv[1]
			}
		}
	}

	return &cfg, nil
}
