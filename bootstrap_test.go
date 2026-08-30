package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestBootstrapVisitsActive(t *testing.T) {
	st := newStubStorer(t)
	if err := st.NewGym(); err != nil {
		t.Fatalf("NewGym: %v", err)
	}

	entry := time.Now().Add(-time.Hour).Format(time.RFC3339)
	_, err := st.gym.db.Exec("INSERT INTO gym (timestamp, action) VALUES (?, ?)", entry, "in")
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	m := NewMetrics(prometheus.NewRegistry())
	if err := bootstrapVisits(map[string]Storer{"TST": st}, m); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	want, err := time.Parse(time.RFC3339, entry)
	if err != nil {
		t.Fatalf("parse entry: %v", err)
	}
	if got := testutil.ToFloat64(m.entry.WithLabelValues("TST")); got != float64(want.Unix()) {
		t.Errorf("expected entry %d, got %v", want.Unix(), got)
	}
}

func TestBootstrapVisitsDone(t *testing.T) {
	st := newStubStorer(t)
	if err := st.NewGym(); err != nil {
		t.Fatalf("NewGym: %v", err)
	}

	entry := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	exit := time.Now().Add(-time.Hour).Format(time.RFC3339)
	_, err := st.gym.db.Exec("INSERT INTO gym (timestamp, action) VALUES (?, ?), (?, ?)", entry, "in", exit, "out")
	if err != nil {
		t.Fatalf("seed visit: %v", err)
	}

	m := NewMetrics(prometheus.NewRegistry())
	if err := bootstrapVisits(map[string]Storer{"TST": st}, m); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	if got := testutil.ToFloat64(m.entry.WithLabelValues("TST")); got != 0 {
		t.Errorf("expected entry reset, got %v", got)
	}
	if got := histCount(t, m.visitDuration.WithLabelValues("TST").(prometheus.Metric)); got != 1 {
		t.Errorf("expected duration count 1, got %d", got)
	}
}
