package loader

import (
	"context"
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

func runSpike(ctx context.Context, cfg *config.Config, eng *engine.Engine, metrics *Metrics, onTick func(int)) {
	spikeCfg := cfg.Load.Spike
	totalDuration, _ := time.ParseDuration(cfg.Load.Duration)
	spikeDuration, _ := time.ParseDuration(spikeCfg.SpikeDuration)
	spikeEvery, _ := time.ParseDuration(spikeCfg.SpikeEvery)

	ctx, cancel := context.WithTimeout(ctx, totalDuration)
	defer cancel()

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		activeCount int
	)

	launchWorkers := func(n int) context.CancelFunc {
		workerCtx, workerCancel := context.WithCancel(ctx)
		mu.Lock()
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
				w.run(workerCtx)
			}()
		}
		mu.Unlock()
		onTick(activeCount)
		return workerCancel
	}

	launchWorkers(spikeCfg.BaseWorkers)

	spikeTimer := time.NewTicker(spikeEvery)
	defer spikeTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			goto done
		case <-spikeTimer.C:
			spikeCancel := launchWorkers(spikeCfg.SpikeWorkers)
			go func() {
				select {
				case <-ctx.Done():
				case <-time.After(spikeDuration):
					spikeCancel()
					mu.Lock()
					activeCount -= spikeCfg.SpikeWorkers
					if activeCount < spikeCfg.BaseWorkers {
						activeCount = spikeCfg.BaseWorkers
					}
					mu.Unlock()
					onTick(activeCount)
				}
			}()
		}
	}

done:
	wg.Wait()
}
