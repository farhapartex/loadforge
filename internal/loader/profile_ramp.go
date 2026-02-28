package loader

import (
	"context"
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

// runRamp gradually increases workers from StartWorkers to EndWorkers
func runRamp(ctx context.Context, cfg *config.Config, eng *engine.Engine, metrics *Metrics, onTick func(int)) {
	rampCfg := cfg.Load.RampUp
	rampDuration, _ := time.ParseDuration(rampCfg.Duration)
	totalDuration, _ := time.ParseDuration(cfg.Load.Duration)

	ctx, cancel := context.WithTimeout(ctx, totalDuration)
	defer cancel()

	startWorkers := rampCfg.StartWorkers
	endWorkers := rampCfg.EndWorkers
	workerDiff := endWorkers - startWorkers

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		activeCount int
		cancelFuncs []context.CancelFunc
	)

	launchWorkers := func(n int) {
		mu.Lock()
		defer mu.Unlock()
		for i := 0; i < n; i++ {
			workerCtx, workerCancel := context.WithCancel(ctx)
			cancelFuncs = append(cancelFuncs, workerCancel)
			w := &worker{
				id:      activeCount,
				eng:     eng,
				cfg:     cfg,
				metrics: metrics,
			}
			activeCount++
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.run(workerCtx)
			}()
		}
		onTick(activeCount)
	}

	// Launch starting batch
	if startWorkers > 0 {
		launchWorkers(startWorkers)
	}

	// Calculate how often to add a worker during ramp
	if workerDiff > 0 {
		interval := rampDuration / time.Duration(workerDiff)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		added := 0
		for added < workerDiff {
			select {
			case <-ctx.Done():
				goto done
			case <-ticker.C:
				launchWorkers(1)
				added++
			}
		}
	}

done:
	<-ctx.Done()
	wg.Wait()

	// Cancel all individual worker contexts on cleanup
	mu.Lock()
	for _, c := range cancelFuncs {
		c()
	}
	mu.Unlock()
}
