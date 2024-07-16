package main

import (
	"os"
	"reflect"
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

	expectedError := "the required env var"
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

func TestNewConfig_NoEnvVar(t *testing.T) {
	envVars := map[string]string{
		"PGK":       "pgk_value",
		"FID":       "fid_value",
		"GYM":       "gym_value",
		"BOT_TOKEN": "bot_token_value",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	os.Unsetenv("SCHEDULE")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Schedule) != 0 {
		t.Errorf("expected empty Schedule, got %v", cfg.Schedule)
	}
}

func TestNewConfig_WithEnvVar(t *testing.T) {
	envVars := map[string]string{
		"PGK":       "pgk_value",
		"FID":       "fid_value",
		"GYM":       "gym_value",
		"BOT_TOKEN": "bot_token_value",
		"SCHEDULE":  "task1=0 */5 * * * MON-FRI|task2=* 12 * * * 2|task3=0 5 12,2 * * *",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	expectedSchedule := map[string]string{
		"task1": "0 */5 * * * MON-FRI",
		"task2": "* 12 * * * 2",
		"task3": "0 5 12,2 * * *",
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(cfg.Schedule, expectedSchedule) {
		t.Errorf("expected Schedule %v, got %v", expectedSchedule, cfg.Schedule)
	}
}

func TestNewConfig_WithMalformedEnvVar(t *testing.T) {
	envVars := map[string]string{
		"PGK":       "pgk_value",
		"FID":       "fid_value",
		"GYM":       "gym_value",
		"BOT_TOKEN": "bot_token_value",
		"SCHEDULE":  "task1=0 */5 * * * MON-FRI|task2|task3=0 5 12,2 * * *",
	}
	setEnvVars(envVars)
	defer unsetEnvVars(envVars)

	expectedSchedule := map[string]string{
		"task1": "0 */5 * * * MON-FRI",
		"task3": "0 5 12,2 * * *",
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(cfg.Schedule, expectedSchedule) {
		t.Errorf("expected Schedule %v, got %v", expectedSchedule, cfg.Schedule)
	}

	if _, exists := cfg.Schedule["task2"]; exists {
		t.Errorf("did not expect task2 in Schedule, got %v", cfg.Schedule)
	}
}
