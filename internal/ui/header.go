package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	brandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")).
			Bold(true)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	profileBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("99")).
				Bold(true).
				Padding(0, 1)

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
)

func (m *Model) renderHeader() string {
	brand := brandStyle.Render("loadforge")
	profileBardge := profileBadgeStyle.Render(strings.ToUpper(m.profile))
	scenarioProfile := metaStyle.Render(fmt.Sprintf(" %s -> %s", m.scenarioName, m.baseUrl))
	divider := dividerStyle.Render(strings.Repeat("-", m.width-2))

	topLine := brand + " " + profileBardge + scenarioProfile

	progressLine := m.renderProgress()

	return topLine + "\\n" + progressLine + "\\n" + divider
}

func (m *Model) renderProgress() string {
	if m.duration == "" {
		elapsed := time.Since(m.startTime).Round(time.Second)
		return metaStyle.Render(fmt.Sprintf("  elapsed: %s  (no time limit)", elapsed))
	}

	totalDuration, err := time.ParseDuration(m.duration)
	if err != nil {
		return ""
	}

	elapsed := time.Since(m.startTime)
	if elapsed > totalDuration {
		elapsed = totalDuration
	}

	pct := elapsed.Seconds() / totalDuration.Seconds()
	if pct > 1 {
		pct = 1
	}

	barTotal := 40
	filled := int(pct * float64(barTotal))
	empty := barTotal - filled

	bar := progressStyle.Render(strings.Repeat("█", filled)) +
		dividerStyle.Render(strings.Repeat("░", empty))

	timeLabel := metaStyle.Render(fmt.Sprintf(
		"  [%s] %s / %s  %.0f%%",
		bar,
		elapsed.Round(time.Second),
		totalDuration,
		pct*100,
	))

	return timeLabel
}
