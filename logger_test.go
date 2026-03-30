package main

import "testing"

func TestSetLogger(t *testing.T) {
	// SetLogger must not panic and must leave slog in a usable state.
	// Called twice to confirm idempotence.
	SetLogger()
	SetLogger()
}
