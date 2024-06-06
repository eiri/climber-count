package main

import (
	"os"
	"testing"
)

func setEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		os.Setenv(key, value)
	}
}

func unsetEnvVars(envVars map[string]string) {
	for key := range envVars {
		os.Unsetenv(key)
	}
}

func TestNewConfig_AllVarsSet(t *testing.T) {
	envVars := map[string]string{
		"PGK":       "pgk_value",
		"FID":       "fid_value",
		"GYM":       "gym_value",
		"BOT_TOKEN": "bot_token_value",
		"STORAGE":   "storage_value",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.PGK != envVars["PGK"] {
		t.Errorf("expected PGK %q, got %q", envVars["PGK"], cfg.PGK)
	}
	if cfg.FID != envVars["FID"] {
		t.Errorf("expected FID %q, got %q", envVars["FID"], cfg.FID)
	}
	if cfg.Gym != envVars["GYM"] {
		t.Errorf("expected GYM %q, got %q", envVars["GYM"], cfg.Gym)
	}
	if cfg.BotToken != envVars["BOT_TOKEN"] {
		t.Errorf("expected BOT_TOKEN %q, got %q", envVars["BOT_TOKEN"], cfg.BotToken)
	}
	if cfg.Storage != envVars["STORAGE"] {
		t.Errorf("expected STORAGE %q, got %q", envVars["STORAGE"], cfg.Storage)
	}
}

func TestNewConfig_RequiredVarsNotSet(t *testing.T) {
	envVars := map[string]string{
		"STORAGE": "storage_value",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	expectedError := "The required env var"
	if err != nil && err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("expected error to start with %q, got %q", expectedError, err.Error())
	}
}

func TestNewConfig_OptionalVarNotSet(t *testing.T) {
	envVars := map[string]string{
		"PGK":       "pgk_value",
		"FID":       "fid_value",
		"GYM":       "gym_value",
		"BOT_TOKEN": "bot_token_value",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.PGK != envVars["PGK"] {
		t.Errorf("expected PGK %q, got %q", envVars["PGK"], cfg.PGK)
	}
	if cfg.FID != envVars["FID"] {
		t.Errorf("expected FID %q, got %q", envVars["FID"], cfg.FID)
	}
	if cfg.Gym != envVars["GYM"] {
		t.Errorf("expected GYM %q, got %q", envVars["GYM"], cfg.Gym)
	}
	if cfg.BotToken != envVars["BOT_TOKEN"] {
		t.Errorf("expected BOT_TOKEN %q, got %q", envVars["BOT_TOKEN"], cfg.BotToken)
	}
	if cfg.Storage != "" {
		t.Errorf("expected STORAGE to be empty, got %q", cfg.Storage)
	}
}
