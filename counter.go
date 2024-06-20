package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
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
	lastUpdate := humanize.Time(c.LastUpdate.Time)
	peopleCount := strconv.Itoa(c.Count)
	if peopleCount == "0" {
		peopleCount = "zero"
	} else if peopleCount == "1" {
		peopleCount = "one"
	}
	peopleCounter := english.PluralWord(c.Count%10, "person", "people")

	if lastUpdate == "now" {
		return fmt.Sprintf("at the moment there's %s %s on the wall", peopleCount, peopleCounter)
	}
	return fmt.Sprintf("%s there've been %s %s on the wall", lastUpdate, peopleCount, peopleCounter)
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
