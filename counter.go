package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imbue11235/humanize"
)

type Counters map[string]Counter

func NewCounters() *Counters {
	return new(Counters)
}

func (c *Counters) Counter(gym string) Counter {
	counter, ok := (*c)[gym]
	if !ok {
		return Counter{}
	}
	return counter
}

type Counter struct {
	Count      int        `json:"count"`
	Capacity   int        `json:"capacity"`
	LastUpdate LastUpdate `json:"lastUpdate"`
}

func (c Counter) String() string {
	lastUpdate := humanize.Time(c.LastUpdate.Time).FromNow()
	peopleCount := strconv.Itoa(c.Count)
	peopleCounter := "people"
	switch peopleCount {
	case "0":
		peopleCount = "zero"
	case "1":
		peopleCount = "one"
		peopleCounter = "person"
	case "21", "31", "41", "51", "61", "71", "81", "91", "101":
		peopleCounter = "person"
	}

	return fmt.Sprintf("%s there've been %s %s on the wall", lastUpdate, peopleCount, peopleCounter)
}

type LastUpdate struct {
	time.Time
}

var reRelative = regexp.MustCompile(`(?i)(\d+)?\s?(day|hour|min|mins|days|hours)s?ago`)
var reNow = regexp.MustCompile(`(?i)now`)

func (lu *LastUpdate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "" || str == "null" {
		lu.Time = time.Time{}
		return nil
	}

	now := time.Now()
	if reNow.MatchString(str) {
		lu.Time = now.Truncate(time.Minute)
		return nil
	}

	matches := reRelative.FindStringSubmatch(str)
	if len(matches) >= 3 {
		n := 1
		if matches[1] != "" {
			var err error
			n, err = strconv.Atoi(matches[1])
			if err != nil {
				return err
			}
		}

		unit := strings.ToLower(matches[2])
		var d time.Duration
		switch unit {
		case "day", "days":
			d = time.Hour * 24 * time.Duration(n)
		case "hour", "hours":
			d = time.Hour * time.Duration(n)
		case "min", "mins":
			d = time.Minute * time.Duration(n)
		default:
			return fmt.Errorf("unrecognized time unit: %s", unit)
		}

		lu.Time = now.Add(-d).Truncate(time.Minute)
		return nil
	}

	return fmt.Errorf("invalid time format: %s", str)
}

// To support Equal comparison in tests
func (lu LastUpdate) Equal(t time.Time) bool {
	return lu.Time.Year() == t.Year() &&
		lu.Time.Month() == t.Month() &&
		lu.Time.Day() == t.Day() &&
		lu.Time.Hour() == t.Hour() &&
		lu.Time.Minute() == t.Minute()
}
