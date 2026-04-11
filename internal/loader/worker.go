package loader

import (
	"context"
	"math/rand"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
)

type worker struct {
	id      int
	eng     *engine.Engine
	cfg     *config.Config
	metrics *Metrics
}

func (w *worker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			scenario := w.pickScenario()
			w.executeScenario(ctx, scenario)
		}
	}
}

func (w *worker) pickScenario() config.Scenario {
	scenarios := w.cfg.Scenarios

	if len(scenarios) == 1 {
		return scenarios[0]
	}

	totalWeight := 0
	for _, s := range scenarios {
		totalWeight += s.Weight
	}

	if totalWeight == 0 {
		return scenarios[rand.Intn(len(scenarios))]
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for _, s := range scenarios {
		cumulative += s.Weight
		if r < cumulative {
			return s
		}
	}

	return scenarios[len(scenarios)-1]
}

func (w *worker) executeScenario(ctx context.Context, scenario config.Scenario) {
	for _, step := range scenario.Steps {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result := w.eng.ExecuteStep(step)
		w.metrics.record(result)

		if step.Think != "" {
			thinkDuration, err := time.ParseDuration(step.Think)
			if err == nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(thinkDuration):
				}
			}
		}
	}
}
