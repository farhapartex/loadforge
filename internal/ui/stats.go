package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(18)
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	goodStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")).
			Bold(true).
			MarginBottom(1)
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true).
			MarginBottom(1)
)

func (m *Model) renderStats() string {
	snap := m.snapshot
	elapsed := snap.Elapsed
	if elapsed == 0 {
		elapsed = time.Since(m.startTime)
	}

	rows := []string{
		row("Requests", valueStyle.Render(fmt.Sprintf("%d", snap.TotalRequests))),
		row("Successful", goodStyle.Render(fmt.Sprintf("%d", snap.TotalSuccesses))),
		row("Failed", renderFailed(snap.TotalFailures)),
		row("Error Rate", renderErrorRate(snap.ErrorRate())),
		row("Avg RPS", valueStyle.Render(fmt.Sprintf("%.2f", snap.RPS))),
		row("Data Received", valueStyle.Render(formatBytes(snap.TotalBytes))),
		row("Elapsed", valueStyle.Render(elapsed.Round(time.Second).String())),
		row("Workers", valueStyle.Render(fmt.Sprintf("%d", m.activeWorkers))),
	}

	content := strings.Join(rows, "\\nn")

	return boxStyle.Render(
		headerStyle.Render("Statistics") + "\\n" + content,
	)
}

func row(label, value string) string {
	return labelStyle.Render(label+":") + " " + value
}

func renderFailed(n int64) string {
	if n == 0 {
		return goodStyle.Render("0")
	}
	return errorStyle.Render(fmt.Sprintf("%d", n))
}

func renderErrorRate(rate float64) string {
	s := fmt.Sprintf("%.2f%%", rate)

	switch {
	case rate == 0:
		return goodStyle.Render(s)
	case rate < 5:
		return warnStyle.Render(s)
	default:
		return errorStyle.Render(s)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
