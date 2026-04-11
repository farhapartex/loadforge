package loader

import (
	"context"
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

func runStep(ctx context.Context, cfg *config.Config, eng *engine.Engine, metrics *Metrics, onTick func(int)) {
	stepCfg := cfg.Load.Step
	stepDuration, _ := time.ParseDuration(stepCfg.StepDuration)
	totalDuration, _ := time.ParseDuration(cfg.Load.Duration)

	ctx, cancel := context.WithTimeout(ctx, totalDuration)
	defer cancel()

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		activeCount int
	)

	launchWorkers := func(n int) {
		mu.Lock()
		defer mu.Unlock()
		for i := 0; i < n; i++ {
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
				w.run(ctx)
			}()
		}
		onTick(activeCount)
	}

	if stepCfg.StartWorkers > 0 {
		launchWorkers(stepCfg.StartWorkers)
	}

	ticker := time.NewTicker(stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
			mu.Lock()
			current := activeCount
			mu.Unlock()

			if current >= stepCfg.MaxWorkers {
				continue
			}

			toAdd := stepCfg.StepSize
			if current+toAdd > stepCfg.MaxWorkers {
				toAdd = stepCfg.MaxWorkers - current
			}
			launchWorkers(toAdd)
		}
	}

done:
	wg.Wait()
}
