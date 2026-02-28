package loader

import (
	"context"
	"fmt"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

type RunResult struct {
	Metrics *Metrics
}

func Run(ctx context.Context, cfg *config.Config, onTick func(workers int), onResult func(*Metrics)) (*RunResult, error) {
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

	if onResult != nil {
		onResult(metrics)
	}

	return &RunResult{Metrics: metrics}, nil
}
