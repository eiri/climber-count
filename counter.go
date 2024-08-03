package main

import (
	"fmt"
	"regexp"
	"strconv"
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
