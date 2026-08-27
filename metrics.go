package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	scrapes        *prometheus.CounterVec
	scrapeDuration prometheus.Histogram
	stores         *prometheus.CounterVec
	commands       *prometheus.CounterVec
	people         *prometheus.GaugeVec
	lastUpdate     *prometheus.GaugeVec
	visits         *prometheus.CounterVec
}

func NewMetrics(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		scrapes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "climber_count_scrape_total",
			Help: "Counter scrape attempts.",
		}, []string{"status"}),
		scrapeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "climber_count_scrape_duration_seconds",
			Help: "Counter scrape duration.",
		}),
		stores: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "climber_count_store_total",
			Help: "Counter store attempts.",
		}, []string{"gym", "status"}),
		commands: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "climber_count_bot_commands_total",
			Help: "Bot command attempts.",
		}, []string{"command", "status"}),
		people: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_people",
			Help: "Collected people count.",
		}, []string{"gym"}),
		lastUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "climber_count_last_update_timestamp_seconds",
			Help: "Collected counter last update time.",
		}, []string{"gym"}),
		visits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "climber_count_gym_visits_total",
			Help: "Gym visit events.",
		}, []string{"gym", "action"}),
	}

	reg.MustRegister(m.scrapes, m.scrapeDuration, m.stores, m.commands, m.people, m.lastUpdate, m.visits)
	return m
}

func (m *Metrics) ScrapeOK(d time.Duration) {
	m.scrape("ok", d)
}

func (m *Metrics) ScrapeErr(d time.Duration) {
	m.scrape("error", d)
}

func (m *Metrics) Store(gym string, ok bool) {
	if m == nil || m.stores == nil {
		return
	}

	m.stores.WithLabelValues(gym, status(ok)).Inc()
}

func (m *Metrics) Counter(gym string, counter Counter) {
	if m == nil || m.people == nil || m.lastUpdate == nil {
		return
	}

	m.people.WithLabelValues(gym).Set(float64(counter.Count))
	m.lastUpdate.WithLabelValues(gym).Set(float64(counter.LastUpdate.Unix()))
}

func (m *Metrics) BotCommand(command string, ok bool) {
	if m == nil || m.commands == nil {
		return
	}

	m.commands.WithLabelValues(command, status(ok)).Inc()
}

func (m *Metrics) Visit(gym, action string) {
	if m == nil || m.visits == nil {
		return
	}

	m.visits.WithLabelValues(gym, action).Inc()
}

func (m *Metrics) scrape(status string, d time.Duration) {
	if m == nil || m.scrapes == nil || m.scrapeDuration == nil {
		return
	}

	m.scrapes.WithLabelValues(status).Inc()
	m.scrapeDuration.Observe(d.Seconds())
}

func status(ok bool) string {
	if ok {
		return "ok"
	}
	return "error"
}
