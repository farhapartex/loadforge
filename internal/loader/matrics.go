package loader

import (
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/engine"
)

type Metrics struct {
	mu sync.RWMutex

	TotalRequests  int64
	TotalSuccesses int64
	TotalFailures  int64
	TotalBytes     int64
	Latencies      []time.Duration
	StatusCodes    map[int]int64
	Errors         map[string]int64
	StartTime      time.Time
	EndTime        time.Time
}

func newMetrics() *Metrics {
	return &Metrics{
		StatusCodes: make(map[int]int64),
		Errors:      make(map[string]int64),
		StartTime:   time.Now(),
	}
}

func (m *Metrics) record(result *engine.Result) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.TotalBytes += result.BytesRead

	if result.Error != nil {
		m.TotalFailures++
		errMsg := result.Error.Error()
		m.Errors[errMsg]++
		return
	}

	m.TotalSuccesses++
	m.Latencies = append(m.Latencies, result.Duration)
	m.StatusCodes[result.StatusCode]++
}

func (m *Metrics) finish() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EndTime = time.Now()
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	latenciesCopy := make([]time.Duration, len(m.Latencies))
	copy(latenciesCopy, m.Latencies)

	statusCopy := make(map[int]int64, len(m.StatusCodes))
	for k, v := range m.StatusCodes {
		statusCopy[k] = v
	}

	errorsCopy := make(map[string]int64, len(m.Errors))
	for k, v := range m.Errors {
		errorsCopy[k] = v
	}

	elapsed := time.Since(m.StartTime)
	rps := 0.0
	if elapsed.Seconds() > 0 {
		rps = float64(m.TotalRequests) / elapsed.Seconds()
	}

	return MetricsSnapshot{
		TotalRequests:  m.TotalRequests,
		TotalSuccesses: m.TotalSuccesses,
		TotalFailures:  m.TotalFailures,
		TotalBytes:     m.TotalBytes,
		Latencies:      latenciesCopy,
		StatusCodes:    statusCopy,
		Errors:         errorsCopy,
		Elapsed:        elapsed,
		RPS:            rps,
	}
}

type MetricsSnapshot struct {
	TotalRequests  int64
	TotalSuccesses int64
	TotalFailures  int64
	TotalBytes     int64
	Latencies      []time.Duration
	StatusCodes    map[int]int64
	Errors         map[string]int64
	Elapsed        time.Duration
	RPS            float64
}

func (s *MetricsSnapshot) ErrorRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.TotalFailures) / float64(s.TotalRequests) * 100
}
