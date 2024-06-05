package main

import (
	"fmt"
	"os"
)

type Config struct {
	PGK     string
	FID     string
	Gym     string
	Storage string
}

func NewConfig() (*Config, error) {
	cfg := Config{}
	envVars := map[string]*string{
		"PGK":     &cfg.PGK,
		"FID":     &cfg.FID,
		"GYM":     &cfg.Gym,
		"STORAGE": &cfg.Storage,
	}

	for key, ptr := range envVars {
		val, ok := os.LookupEnv(key)
		if key != "STORAGE" && !ok {
			return &cfg, fmt.Errorf("The required env var %q is not set", key)
		}
		*ptr = val
	}

	return &cfg, nil
}
