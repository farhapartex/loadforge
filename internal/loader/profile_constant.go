package loader

import (
	"context"
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

// runConstant runs a fixed number of workers for the configured duration
func runConstant(ctx context.Context, cfg *config.Config, eng *engine.Engine, metrics *Metrics, onTick func(int)) {
	numWorkers := cfg.Load.Workers

	// If duration is set, cancel context after duration
	if cfg.Load.Duration != "" {
		duration, _ := time.ParseDuration(cfg.Load.Duration)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}

	onTick(numWorkers) // notify current worker count

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		w := &worker{
			id:      i,
			eng:     eng,
			cfg:     cfg,
			metrics: metrics,
		}
		go func() {
			defer wg.Done()
			w.run(ctx)
		}()
	}

	// Handle max_requests stopping condition
	if cfg.Load.MaxRequests > 0 {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					snap := metrics.Snapshot()
					if snap.TotalRequests >= int64(cfg.Load.MaxRequests) {
						return
					}
				}
			}
		}()
	}

	wg.Wait()
}
