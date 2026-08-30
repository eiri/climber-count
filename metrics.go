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
	lastUpdate     *prometheus.GaugeVec
	entry          *prometheus.GaugeVec
	visitDuration  *prometheus.HistogramVec
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
		lastUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_last_update_timestamp_seconds",
			Help: "Collected counter last update time.",
		}, []string{"gym"}),
		entry: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_gym_entry_timestamp_seconds",
			Help: "Current gym visit start time.",
		}, []string{"gym"}),
		visitDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "climber_gym_visit_duration_seconds",
			Help: "Completed gym visit duration.",
		}, []string{"gym"}),
	}

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.scrapeDuration,
		m.writes,
		m.people,
		m.lastUpdate,
		m.entry,
		m.visitDuration,
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
	if m == nil || m.people == nil || m.lastUpdate == nil {
		return
	}

	m.people.WithLabelValues(gym).Set(float64(counter.Count))
	m.lastUpdate.WithLabelValues(gym).Set(float64(counter.LastUpdate.Unix()))
}

func (m *Metrics) VisitIn(gym string, ts time.Time) {
	if m == nil || m.entry == nil {
		return
	}

	m.entry.WithLabelValues(gym).Set(float64(ts.Unix()))
}

func (m *Metrics) VisitOut(gym string, d time.Duration) {
	if m == nil || m.entry == nil || m.visitDuration == nil {
		return
	}

	m.entry.WithLabelValues(gym).Set(0)
	m.visitDuration.WithLabelValues(gym).Observe(d.Seconds())
}

func (m *Metrics) scrape(d time.Duration) {
	if m == nil || m.scrapeDuration == nil {
		return
	}

	m.scrapeDuration.Observe(d.Seconds())
}
