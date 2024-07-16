package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PGK      string
	FID      string
	Gym      string
	BotToken string
	Storage  string
	Schedule map[string]string
}

func NewConfig() (*Config, error) {
	cfg := Config{
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
		*ptr = val
	}

	if val, ok := os.LookupEnv("SCHEDULE"); ok {
		for _, subVal := range strings.Split(val, "|") {
			if strings.Contains(subVal, "=") {
				kv := strings.SplitN(subVal, "=", 2)
				cfg.Schedule[kv[0]] = kv[1]
			}
		}
	}

	return &cfg, nil
}
