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
			name:     "Valid time with one day ago",
			input:    `"Lastupdated:&nbsp1dayago"`,
			expected: time.Date(time.Now().Year(), time.Now().Add(-24*time.Hour).Month(), time.Now().Add(-24*time.Hour).Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with days ago",
			input:    `"Lastupdated:&nbsp2daysago"`,
			expected: time.Date(time.Now().Year(), time.Now().Add(-24*time.Hour).Month(), time.Now().Add(-48*time.Hour).Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with one hour ago",
			input:    `"Lastupdated:&nbsp1hourago"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Add(-1*time.Hour).Hour(), time.Now().Add(-1*time.Hour).Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with hours ago",
			input:    `"Lastupdated:&nbsp3hoursago"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Add(-3*time.Hour).Hour(), time.Now().Add(-3*time.Hour).Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with one minute ago",
			input:    `"Lastupdated:&nbsp1minago"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Add(-1*time.Minute).Hour(), time.Now().Add(-1*time.Minute).Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with minutes ago",
			input:    `"Lastupdated:&nbsp12minsago"`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Add(-12*time.Minute).Hour(), time.Now().Add(-12*time.Minute).Minute(), 0, 0, time.Local),
			hasError: false,
		},
		{
			name:     "Valid time with now",
			input:    `"Lastupdated:&nbspnow "`,
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.Local),
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
			input:    `"Lastupdated:&nbsp12xxago"`,
			expected: time.Time{},
			hasError: true,
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
