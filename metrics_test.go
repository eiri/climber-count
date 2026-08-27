package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	counter := Counter{
		Count: 7,
		LastUpdate: LastUpdate{
			Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
		},
	}

	m.ScrapeOK(time.Second)
	m.Store("TST", true)
	m.Counter("TST", counter)
	m.BotCommand("count", true)
	m.Visit("TST", "in")

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"scrape", testutil.ToFloat64(m.scrapes.WithLabelValues("ok")), 1},
		{"store", testutil.ToFloat64(m.stores.WithLabelValues("TST", "ok")), 1},
		{"people", testutil.ToFloat64(m.people.WithLabelValues("TST")), 7},
		{"last update", testutil.ToFloat64(m.lastUpdate.WithLabelValues("TST")), 1717063200},
		{"command", testutil.ToFloat64(m.commands.WithLabelValues("count", "ok")), 1},
		{"visit", testutil.ToFloat64(m.visits.WithLabelValues("TST", "in")), 1},
	}

	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s: expected %v, got %v", check.name, check.want, check.got)
		}
	}
}
