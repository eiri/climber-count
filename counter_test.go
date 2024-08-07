package main

import (
	"fmt"
	"testing"
	"time"
)

func TestLastUpdate_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "Valid time with minutes ago",
			input:    `"Lastupdated:&nbsp12minsago(10:42PM)"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 22, 42, 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with now",
			input:    `"Lastupdated:&nbspnow(10:52PM)"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 22, 52, 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Null input",
			input:    `null`,
			expected: time.Time{},
			hasError: false,
		},
		{
			name:     "Empty string input",
			input:    `""`,
			expected: time.Time{},
			hasError: false,
		},
		{
			name:     "Invalid time format",
			input:    `"Lastupdated:&nbsp12minsago(10:xxPM)"`,
			expected: time.Time{},
			hasError: true,
		},
		{
			name:     "Boundary time 12:00AM",
			input:    `"Lastupdated:&nbsp12minsago(12:00AM)"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Boundary time 11:59PM",
			input:    `"Lastupdated:&nbsp12minsago(11:59PM)"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 0, 0, time.Local),
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var lu LastUpdate
			err := lu.UnmarshalJSON([]byte(tc.input))

			if tc.hasError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got %v", err)
				}

				if !lu.Equal(tc.expected) {
					t.Errorf("expected %v but got %v", tc.expected, lu)
				}
			}
		})
	}
}

func TestCounter_String(t *testing.T) {
	tests := []struct {
		counter Counter
		want    string
	}{
		{Counter{Count: 1, LastUpdate: LastUpdate{time.Now().Add(-2000000 * time.Hour)}}, "a long time ago there've been one person on the wall"},
		{Counter{Count: 2, LastUpdate: LastUpdate{time.Now()}}, "a few seconds ago there've been 2 people on the wall"},
		{Counter{Count: 11, LastUpdate: LastUpdate{time.Now()}}, "a few seconds ago there've been 11 people on the wall"},
		{Counter{Count: 21, LastUpdate: LastUpdate{time.Now()}}, "a few seconds ago there've been 21 person on the wall"},
		{Counter{Count: 0, LastUpdate: LastUpdate{time.Now().Add(-2000000 * time.Hour)}}, "a long time ago there've been zero people on the wall"},
		{Counter{Count: 100, LastUpdate: LastUpdate{time.Now().Add(-3 * time.Minute)}}, "3 minutes ago there've been 100 people on the wall"},
		{Counter{Count: 101, LastUpdate: LastUpdate{time.Now().Add(-2 * time.Hour)}}, "2 hours ago there've been 101 person on the wall"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Count=%d", tt.counter.Count), func(t *testing.T) {
			if got := tt.counter.String(); got != tt.want {
				t.Errorf("Counter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCounters(t *testing.T) {
	counters := NewCounters()
	if counters == nil {
		t.Fatal("expected non-nil Counters")
	}
	if len(*counters) != 0 {
		t.Fatalf("expected empty Counters map, got size %d", len(*counters))
	}
}

func TestCounters_Counter(t *testing.T) {
	gymName := "GymA"
	expectedTime, _ := time.Parse("2006-01-02", "2024-05-30")
	expectedCounter := Counter{Count: 10, Capacity: 50, LastUpdate: LastUpdate{Time: expectedTime}}
	counters := &Counters{"GymA": expectedCounter}

	retrievedCounter := counters.Counter(gymName)
	if retrievedCounter != expectedCounter {
		t.Errorf("expected Counter %v, got %v", expectedCounter, retrievedCounter)
	}

	nonExistentGym := "GymB"
	retrievedCounter = counters.Counter(nonExistentGym)
	if (retrievedCounter != Counter{}) {
		t.Errorf("expected default Counter for non-existent gym, got %v", retrievedCounter)
	}
}
