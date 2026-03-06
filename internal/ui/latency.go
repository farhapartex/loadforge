package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	barStyleGood = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	barStyleWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	barStyleBad  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	pctStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(5)
	msStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Width(10)
)

func (m *Model) renderLatency() string {
	latencies := m.snapshot.Latencies

	if len(latencies) == 0 {
		return boxStyle.Render(
			headerStyle.Render("Latency") + "\\n" +
				labelStyle.Render("Waiting for data..."),
		)
	}

	p50 := percentile(latencies, 50)
	p90 := percentile(latencies, 90)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)
	max := percentile(latencies, 100)

	rows := []string{
		latencyBar("p50", p50, max),
		latencyBar("p90", p90, max),
		latencyBar("p95", p95, max),
		latencyBar("p99", p99, max),
	}

	content := strings.Join(rows, "\\n")
	return boxStyle.Render(
		headerStyle.Render("Latency Percentiles") + "\\n" + content,
	)
}

func latencyBar(label string, d, max time.Duration) string {
	ms := float64(d.Milliseconds())
	maxMs := float64(max.Milliseconds())

	// Calculate bar width proportional to max
	barWidth := 0
	if maxMs > 0 {
		barWidth = int((ms / maxMs) * float64(maxBarWidth))
	}
	if barWidth < 1 && ms > 0 {
		barWidth = 1
	}

	bar := strings.Repeat("", barWidth)
	barStyled := colorBar(bar, d)

	msLabel := fmt.Sprintf("%dms", d.Milliseconds())

	return pctStyle.Render(label) +
		barStyled +
		strings.Repeat(" ", maxBarWidth-barWidth+1) +
		msStyle.Render(msLabel)
}

func colorBar(bar string, d time.Duration) string {
	switch {
	case d < 200*time.Millisecond:
		return barStyleGood.Render(bar)
	case d < 500*time.Millisecond:
		return barStyleWarn.Render(bar)
	default:
		return barStyleBad.Render(bar)
	}
}

func percentile(latencies []time.Duration, n int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	if n == 100 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(len(sorted)-1) * float64(n) / 100.0)
	return sorted[idx]
}
