package loader

import (
	"context"
	"fmt"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

const metricsInterval = 250 * time.Millisecond

type RunResult struct {
	Metrics *Metrics
}

// Run executes a load test from an in-memory config.
// It does not read from disk — callers are responsible for building cfg
// (via config.LoadFromFile for the CLI, or via the OpenAPI generator for the web layer).
func Run(
	ctx context.Context,
	cfg *config.Config,
	onTick func(workers int),
	metricsCh chan<- MetricsSnapshot,
	doneCh chan<- struct{},
) (*RunResult, error) {

	eng := engine.New(cfg)
	metrics := newMetrics()

	go broadcastMetrics(ctx, metrics, metricsCh)

	switch cfg.Load.Profile {
	case "constant":
		runConstant(ctx, cfg, eng, metrics, onTick)
	case "ramp":
		runRamp(ctx, cfg, eng, metrics, onTick)
	case "step":
		runStep(ctx, cfg, eng, metrics, onTick)
	case "spike":
		runSpike(ctx, cfg, eng, metrics, onTick)
	default:
		return nil, fmt.Errorf("unknown profile: %s", cfg.Load.Profile)
	}

	metrics.finish()

	if doneCh != nil {
		close(doneCh)
	}

	return &RunResult{Metrics: metrics}, nil
}

func broadcastMetrics(ctx context.Context, m *Metrics, ch chan<- MetricsSnapshot) {
	ticker := time.NewTicker(metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snap := m.Snapshot()
			select {
			case ch <- snap:
			default:
			}
		}
	}
}
