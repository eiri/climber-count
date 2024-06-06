package main

import (
	"fmt"
	"regexp"
	"time"
)

type Counters map[string]Counter

func NewCounters() *Counters {
	return new(Counters)
}

func (c *Counters) Counter(gym string) Counter {
	return (*c)[gym]
}

type Counter struct {
	Count      int        `json:"count"`
	Capacity   int        `json:"capacity"`
	LastUpdate LastUpdate `json:"lastUpdate"`
}

func (c Counter) String() string {
	msg := "One person on the wall"
	if c.Count > 1 {
		msg = fmt.Sprintf("%d people on the wall", c.Count)
	}
	return msg
}

var re = regexp.MustCompile(`\d{1,}:\d{2}[AP]M`)

type LastUpdate struct {
	time.Time
}

func (lu *LastUpdate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	match := re.Find(data)
	kitchen, err := time.Parse(time.Kitchen, string(match))
	if err != nil {
		return err
	}
	now := time.Now()
	lastUpdate := time.Date(now.Year(), now.Month(), now.Day(), kitchen.Hour(), kitchen.Minute(), 0, 0, time.Local)
	*lu = LastUpdate{lastUpdate}
	return nil
}
