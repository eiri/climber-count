package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func TestMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	counter := Counter{
		Count:    7,
		Capacity: 70,
		LastUpdate: LastUpdate{
			Time: time.Date(2024, time.May, 30, 10, 0, 0, 0, time.UTC),
		},
	}

	m.ScrapeOK(time.Second)
	m.Write("counter_store", time.Second)
	m.Counter("TST", counter)

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"people", testutil.ToFloat64(m.people.WithLabelValues("TST")), 7},
		{"capacity", testutil.ToFloat64(m.capacity.WithLabelValues("TST")), 70},
		{"last update", testutil.ToFloat64(m.lastUpdate.WithLabelValues("TST")), 1717063200},
	}

	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s: expected %v, got %v", check.name, check.want, check.got)
		}
	}

	if got := histCount(t, m.writes.WithLabelValues("counter_store").(prometheus.Metric)); got != 1 {
		t.Errorf("expected write count 1, got %d", got)
	}
}

func histCount(t *testing.T, metric prometheus.Metric) uint64 {
	t.Helper()

	pb := &io_prometheus_client.Metric{}
	if err := metric.Write(pb); err != nil {
		t.Fatalf("write metric: %v", err)
	}
	return pb.GetHistogram().GetSampleCount()
}
