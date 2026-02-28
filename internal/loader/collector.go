package loader

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/farhapartex/loadforge/internal/engine"
)

type Stats struct {
	TotalRequests int64
	SuccessCount  int64
	FailureCount  int64
	TotalBytes    int64
	TotalDuration int64 // nanoseconds
	MinDuration   int64 // nanoseconds
	MaxDuration   int64 // nanoseconds

	mu      sync.Mutex
	Results []*engine.Result
}

type StatsSnapshot struct {
	TotalRequest int64
	SuccessCount int64
	FailureCount int64
	TotalBytes   int64
	AvgDuration  time.Duration
	MinDuration  time.Duration
	MaxDuration  time.Duration
	ErrorRate    float64
}

func (s *Stats) Record(r *engine.Result) {
	atomic.AddInt64(&s.TotalRequests, 1)
	atomic.AddInt64(&s.TotalBytes, r.BytesRead)
	atomic.AddInt64(&s.TotalDuration, int64(r.Duration))

	if r.Error != nil || r.StatusCode >= 400 {
		atomic.AddInt64(&s.FailureCount, 1)
	} else {
		atomic.AddInt64(&s.SuccessCount, 1)
	}

	durNs := int64(r.Duration)

	for {
		cur := atomic.LoadInt64(&s.MaxDuration)
		if cur >= durNs {
			break
		}

		if atomic.CompareAndSwapInt64(&s.MaxDuration, cur, durNs) {
			break
		}
	}

	s.mu.Lock()
	s.Results = append(s.Results, r)
	s.mu.Unlock()
}

func (s *Stats) Snapshot() StatsSnapshot {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessCount)
	failures := atomic.LoadInt64(&s.FailureCount)
	bytes := atomic.LoadInt64(&s.TotalBytes)
	totalDur := atomic.LoadInt64(&s.TotalDuration)
	minDur := atomic.LoadInt64(&s.MinDuration)
	maxDur := atomic.LoadInt64(&s.MaxDuration)

	var avgDur time.Duration
	var errorRate float64

	if total > 0 {
		avgDur = time.Duration(totalDur / total)
		errorRate = float64(failures) / float64(total) * 100
	}

	return StatsSnapshot{
		TotalRequest: total,
		SuccessCount: success,
		FailureCount: failures,
		TotalBytes:   bytes,
		AvgDuration:  avgDur,
		MinDuration:  time.Duration(minDur),
		MaxDuration:  time.Duration(maxDur),
		ErrorRate:    errorRate,
	}
}

func (s *Stats) Percentiles() (p50, p90, p95, p99 time.Duration) {
	s.mu.Lock()

	results := make([]*engine.Result, len(s.Results))
	copy(results, s.Results)
	s.mu.Unlock()

	if len(results) == 0 {
		return 0, 0, 0, 0
	}

	durations := make([]int64, 0, len(results))

	for _, r := range results {
		if r.Error == nil {
			durations = append(durations, int64(r.Duration))
		}
	}

	if len(durations) == 0 {
		return 0, 0, 0, 0
	}

	sortInt64(durations)

	p50 = time.Duration(percentile(durations, 50))
	p90 = time.Duration(percentile(durations, 90))
	p95 = time.Duration(percentile(durations, 95))
	p99 = time.Duration(percentile(durations, 99))

	return
}

func percentile(sorted []int64, p float64) int64 {
	if len(sorted) == 0 {
		return 0
	}

	idx := int(float64(len(sorted)-1) * p / 100)

	return sorted[idx]
}

// sortInt64 is a simple insertion sort
func sortInt64(data []int64) {
	for idx := 1; idx < len(data); idx++ {
		key := data[idx]
		jdx := idx - 1
		for jdx >= 0 && data[jdx] > key {
			data[jdx+1] = data[jdx]
			jdx--
		}

		data[jdx+1] = key
	}
}
