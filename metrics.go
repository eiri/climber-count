package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	scrapeDuration prometheus.Histogram
	writes         *prometheus.HistogramVec
	people         *prometheus.GaugeVec
	capacity       *prometheus.GaugeVec
	lastUpdate     *prometheus.GaugeVec
}

func NewMetrics(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		scrapeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "climber_count_scrape_duration_seconds",
			Help: "Counter scrape duration.",
		}),
		writes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "sqlite_write_duration_seconds",
			Help: "SQLite write duration.",
		}, []string{"operation"}),
		people: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_people",
			Help: "Collected people count.",
		}, []string{"gym"}),
		capacity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_max_capacity",
			Help: "Collected gym max capacity.",
		}, []string{"gym"}),
		lastUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_last_update_timestamp_seconds",
			Help: "Collected counter last update time.",
		}, []string{"gym"}),
	}

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.scrapeDuration,
		m.writes,
		m.people,
		m.capacity,
		m.lastUpdate,
	)
	return m
}

func (m *Metrics) ScrapeOK(d time.Duration) {
	m.scrape(d)
}

func (m *Metrics) ScrapeErr(d time.Duration) {
	m.scrape(d)
}

func (m *Metrics) Write(operation string, d time.Duration) {
	if m == nil || m.writes == nil {
		return
	}

	m.writes.WithLabelValues(operation).Observe(d.Seconds())
}

func (m *Metrics) Counter(gym string, counter Counter) {
	if m == nil || m.people == nil || m.capacity == nil || m.lastUpdate == nil {
		return
	}

	m.people.WithLabelValues(gym).Set(float64(counter.Count))
	m.capacity.WithLabelValues(gym).Set(float64(counter.Capacity))
	m.lastUpdate.WithLabelValues(gym).Set(float64(counter.LastUpdate.Unix()))
}

func (m *Metrics) scrape(d time.Duration) {
	if m == nil || m.scrapeDuration == nil {
		return
	}

	m.scrapeDuration.Observe(d.Seconds())
}
