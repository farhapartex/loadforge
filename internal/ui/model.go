package ui

import (
	"time"

	"github.com/farhapartex/loadforge/internal/loader"
)

const (
	tickInterval  = 250 * time.Millisecond // how often the UI refresh
	sparklineSize = 60
	maxBarWidth   = 30
)

type Model struct {
	scenarioName string
	baseUrl      string
	profile      string
	duration     string

	snapshot      loader.MetricsSnapshot
	rpsHistory    []float64
	activeWorkers int
	started       bool
	done          bool
	startTime     time.Time

	width  int
	height int

	showErrors bool

	metricsCh <-chan loader.MetricsSnapshot
	doneCh    <-chan struct{}
}

func NewModel(scenarioName, baseURL, profile, duration string, metricsCh <-chan loader.MetricsSnapshot, doneCh <-chan struct{}) Model {
	return Model{
		scenarioName: scenarioName,
		baseUrl:      baseURL,
		profile:      profile,
		duration:     duration,
		metricsCh:    metricsCh,
		doneCh:       doneCh,
		startTime:    time.Now(),
		rpsHistory:   make([]float64, 0, sparklineSize),
	}
}

func (m *Model) pushRPS(rps float64) {
	if len(m.rpsHistory) >= sparklineSize {
		m.rpsHistory = m.rpsHistory[1:]
	}

	m.rpsHistory = append(m.rpsHistory, rps)
}
