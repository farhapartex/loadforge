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

func Run(
	ctx context.Context,
	cfg *config.Config,
	onTick func(workers int),
	metricsCh chan<- MetricsSnapshot,
	doneCh chan<- struct{},
	//onResult func(*Metrics),
) (*RunResult, error) {

	eng := engine.New(cfg)
	metrics := newMetrics()

	fmt.Printf("Profile  : %s\n", cfg.Load.Profile)
	if cfg.Load.Duration != "" {
		fmt.Printf("Duration : %s\n", cfg.Load.Duration)
	}
	if cfg.Load.MaxRequests > 0 {
		fmt.Printf("Max Reqs : %d\n", cfg.Load.MaxRequests)
	}
	fmt.Println()

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

	// if onResult != nil {
	// 	onResult(metrics)
	// }

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
				// Drop if UI is not reading fast enough
				return
			}
		}
	}
}
