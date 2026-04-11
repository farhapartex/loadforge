package runner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/loader"
)

type Runner struct {
	mu      sync.Mutex
	active  atomic.Bool
	cancel  context.CancelFunc
	results *ResultStore
}

func New(historyFile string) *Runner {
	return &Runner{
		results: newResultStore(historyFile),
	}
}

func (r *Runner) Start(cfg *config.Config, specURL string, onDone func(status string)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.active.Load() {
		return fmt.Errorf("a test is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.active.Store(true)

	record := &RunRecord{
		ID:        newRunID(),
		SpecURL:   specURL,
		Profile:   cfg.Load.Profile,
		Workers:   cfg.Load.Workers,
		Duration:  cfg.Load.Duration,
		StartedAt: time.Now(),
		Status:    StatusRunning,
	}

	log.Printf("Run started  id=%s spec=%s workers=%d duration=%s",
		record.ID, specURL, cfg.Load.Workers, cfg.Load.Duration)

	go r.execute(ctx, cancel, cfg, record, onDone)

	return nil
}

func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Runner) IsActive() bool {
	return r.active.Load()
}

func (r *Runner) Results() *ResultStore {
	return r.results
}

func (r *Runner) execute(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, record *RunRecord, onDone func(string)) {
	defer func() {
		cancel() // stop logMetricsTicks and broadcastMetrics on natural completion
		r.active.Store(false)
		r.cancel = nil
	}()

	metricsCh := make(chan loader.MetricsSnapshot, 20)
	doneCh := make(chan struct{})

	onTick := func(activeWorkers int) {
		log.Printf("Workers active=%d", activeWorkers)
	}

	var result *loader.RunResult
	var runErr error

	go func() {
		result, runErr = loader.Run(ctx, cfg, onTick, metricsCh, doneCh)
	}()

	go r.logMetricsTicks(ctx, metricsCh)

	<-doneCh

	record.EndedAt = time.Now()

	if runErr != nil {
		record.Status = StatusFailed
		record.Error = runErr.Error()
		log.Printf("Run failed  id=%s error=%s", record.ID, runErr)
	} else if ctx.Err() != nil {
		record.Status = StatusStopped
		log.Printf("Run stopped id=%s elapsed=%s", record.ID, record.EndedAt.Sub(record.StartedAt).Round(time.Millisecond))
	} else {
		record.Status = StatusCompleted
		log.Printf("Run completed id=%s elapsed=%s", record.ID, record.EndedAt.Sub(record.StartedAt).Round(time.Millisecond))
	}

	if result != nil && result.Metrics != nil {
		snap := result.Metrics.Snapshot()
		record.Percentiles = ComputePercentiles(&snap) // computes p50-p99 and clears raw latencies
		record.Snapshot = &snap
		log.Printf("Results  requests=%d successes=%d failures=%d rps=%.2f",
			snap.TotalRequests, snap.TotalSuccesses, snap.TotalFailures, snap.RPS)
	}

	r.results.add(record)

	if onDone != nil {
		onDone(string(record.Status))
	}
}

func (r *Runner) logMetricsTicks(ctx context.Context, ch <-chan loader.MetricsSnapshot) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case snap := <-ch:
				log.Printf("Progress  requests=%d rps=%.2f errors=%d",
					snap.TotalRequests, snap.RPS, snap.TotalFailures)
				for msg, count := range snap.Errors {
					log.Printf("  error [%dx] %s", count, msg)
				}
			default:
			}
		}
	}
}

func newRunID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
