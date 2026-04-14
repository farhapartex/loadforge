package runner

import (
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/loader"
)

func evaluateAssertions(assertions []config.Assertion, snap *loader.MetricsSnapshot, percentiles *LatencyPercentiles) ([]config.AssertionResult, bool) {
	if len(assertions) == 0 {
		return nil, true
	}

	metricValues := buildMetricValues(snap, percentiles)

	results := make([]config.AssertionResult, 0, len(assertions))
	allPassed := true

	for _, a := range assertions {
		if !a.Enabled {
			continue
		}

		actual, ok := metricValues[a.Metric]
		if !ok {
			continue
		}

		passed := applyOperator(actual, a.Operator, a.Value)
		if !passed {
			allPassed = false
		}

		results = append(results, config.AssertionResult{
			Metric:    a.Metric,
			Operator:  a.Operator,
			Threshold: a.Value,
			Actual:    actual,
			Passed:    passed,
		})
	}

	return results, allPassed
}

func buildMetricValues(snap *loader.MetricsSnapshot, p *LatencyPercentiles) map[string]float64 {
	values := map[string]float64{
		"rps":            snap.RPS,
		"total_requests": float64(snap.TotalRequests),
		"total_errors":   float64(snap.TotalFailures),
		"error_rate":     snap.ErrorRate(),
		"success_rate":   successRate(snap),
	}

	if p != nil {
		values["p50_latency"] = durationToMs(p.P50)
		values["p90_latency"] = durationToMs(p.P90)
		values["p95_latency"] = durationToMs(p.P95)
		values["p99_latency"] = durationToMs(p.P99)
		values["avg_latency"] = durationToMs(p.Avg)
		values["max_latency"] = durationToMs(p.Max)
	}

	return values
}

func durationToMs(d time.Duration) float64 {
	return float64(d.Milliseconds())
}

func successRate(snap *loader.MetricsSnapshot) float64 {
	if snap.TotalRequests == 0 {
		return 0
	}
	return float64(snap.TotalSuccesses) / float64(snap.TotalRequests) * 100
}

func applyOperator(actual float64, operator string, threshold float64) bool {
	switch operator {
	case "less_than":
		return actual < threshold
	case "less_than_or_equal":
		return actual <= threshold
	case "greater_than":
		return actual > threshold
	case "greater_than_or_equal":
		return actual >= threshold
	case "equal":
		return actual == threshold
	default:
		return false
	}
}
