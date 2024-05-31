package main

import (
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
